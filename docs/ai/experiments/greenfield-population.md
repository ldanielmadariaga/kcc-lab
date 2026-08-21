# Measuring every greenfield resource

The 239-resource experiment was scoped by "has a `--resource` line in a v1alpha1 `generate.sh`",
which is a proxy for greenfield, not a definition of it. This run defines the population properly
and measures all of it.

## Defining greenfield

A resource is greenfield when it was written for the direct controller from the start, rather than
inherited from Terraform or DCL. Two signals, and neither is sufficient alone:

* **`pkg/controller/resourceconfig/static_config.go`** — a resource whose `SupportedControllers` is
  exactly `[Direct]` has never had a TF or DCL implementation. One that lists `Terraform` or `DCL`,
  including `[Direct, Terraform]`, is a migration and its KRM shape was inherited.
* **The CRD's own versions** — greenfield work happens at `v1alpha1`.

`static_config.go` alone is not enough: a CRD with no controller yet does not appear in it at all,
and those are greenfield by construction. Parsed from `upstream/master`, it holds 531 rows:

| `SupportedControllers` | count |
|---|---|
| `Direct` | 227 |
| `Terraform` | 171 |
| `Direct + Terraform` | 64 |
| `DCL` | 51 |
| `DCL + Direct` | 14 |
| IAM special cases | 4 |

Crossed with the 613 CRDs in `config/crds/resources`:

| class | has v1alpha1 | no v1alpha1 |
|---|---|---|
| direct-only | **218** | 8 |
| absent from static_config | **82** | 0 |
| terraform-backed | 112 | 123 |
| dcl-backed | 0 | 65 |

**The greenfield population is 300**: 218 direct-only plus 82 with no controller registered yet, all
carrying a v1alpha1 version. 245 are v1alpha1-only; 55 also have a v1beta1.

## What the old measurement was actually measuring

Reclassifying the 239: **225 greenfield, 12 Terraform-backed, 2 unmatched.**

The twelve are worth naming, because they are where the confusing results came from — `APIKeysKey`,
`BigQueryReservationCapacityCommitment`, `VertexAITensorboard`, two KMS resources, and seven
`Compute` ones. These are resources whose KRM shape was set by the Terraform provider years ago, so
scoring generator output against them asks the generator to reproduce a Terraform artifact. The
clearest case is the map-modelled-as-slice pattern: the TF SDK's `TypeMap` holds only primitives, so
a proto `map<string, Message>` had to become a `TypeSet` with the key promoted to an attribute, and
KCC inherited that. `artifactregistry`'s `[]CleanupPolicy` is exactly this, and it is why the mapper
generator needed a guard for a KRM side that is not a map.

Contamination was only 5% by count, but concentrated in the resources that swung hardest.

## Coverage: what can actually be measured

Of the 300, **231 are regenerable** — they have a v1alpha1 `--resource` line that scaffolds a types
file. The other 69 break down as:

| | count | |
|---|---|---|
| generated, but only as a v1beta1 invocation | 37 | promoted; measurable, but against a different version |
| hand-written in a service that otherwise generates | 19 | e.g. most of `apigee` |
| service has no `generate.sh` | 8 | e.g. `bigqueryanalyticshub`, `binaryauthorization` |
| generated with `--skip-scaffold-files` | 5 | the generator never writes their types file |

Of the 231, **225 actually regenerated**, and **189 are scorable** — the rest sit in services whose
CRD generation fails, so their CRD is stale and would score a false 100%.

## Result, over all 189

| bucket | rate | excluding missing references |
|---|---|---|
| spec | 78.5% | **95.7%** |
| required | 78.1% | — |
| status.observedState | 88.5% | 88.5% (no references involved) |

References: **484 in the baseline, 168 reproduced, 203 left as plain strings, 113 absent.**

For contrast, the 90 resources of the earlier comparable set score 80.5% / 98.3% on the same tree.
The wider population scores a little lower, which is what you would expect: the smaller set is the
one that has had attention.

**References remain the whole story.** 1264 of 1466 missing spec paths are reference paths — 86%.
Since each unreproduced reference costs four or five paths, that is roughly 280 actual references
against 202 non-reference field gaps across 189 resources.

## What is left, once references are set aside

**202 non-reference spec fields.** Concentrated: `CloudSecurityComplianceCloudControl` (40),
`VertexAITrainingPipeline` (32), `AIPlatformModel` (25), `DatastreamConnectionProfile` (15),
`NetworkManagementConnectivityTest` (13) are 125 of them. Two of those are the JSON well-known type
effect rather than a gap — `VertexAITrainingPipeline` and `AIPlatformModel` "lose" the
`boolValue`/`nullValue`/`numberValue`/`stringValue`/`structValue` arms of upstream's hand-written
`Value` union, which the generator now correctly emits as
`x-kubernetes-preserve-unknown-fields`. Discounting those, the real remainder is around 145.

`spec.location` accounts for 27, where the proto's `google.api.resource` pattern is not
project+location but upstream has a location anyway.

**298 observedState fields, across 61 of 189 resources.** More concentrated still:
`ComputeFutureReservation` (59), `ComputeInterconnect` (31),
`NetworkConnectivityServiceConnectionPolicy` (22), `ComputeNetworkAttachment` (17) and
`DataLabelingDataset` (17) are 146 of the 298. A handful of resources with large observed-state
surfaces the OUTPUT_ONLY detector is not reading, rather than a broad weakness — 128 of the 189
resources have no observedState gap at all.

## Reading this next to the earlier runs

The headline is not comparable to the 239-run's 80.1%, and should not be quoted as a regression: it
is a different and larger population, defined by controller type rather than by whether a
`generate.sh` line happened to exist. The comparable figure is in
[PROGRESS.md](PROGRESS.md), which tracks the fixed 93-resource intersection across runs.

What this run adds is scope and a clean denominator: 189 greenfield resources measured with no
Terraform-shaped resources in the set, and a stated account of the 111 greenfield resources that
cannot be measured this way and why.

Inputs preserved in [data-greenfield/](data-greenfield/).
