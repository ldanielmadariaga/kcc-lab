# Greenfield generator: measured progress

One table per measurement, so the trend is visible without re-reading each write-up. Every row comes
from the same experiment — delete resources upstream implemented by hand, regenerate them, score the
CRDs against `c1df0b9326` with `crd-mcp-server score`.

Add a row rather than editing one. A number that moved for a methodology reason is still a number
that moved, and the reason belongs beside it.

## Headline

Measured on the 93 resources scorable in every run to date.

| run | spec | spec excl. refs | required | observedState |
|---|---|---|---|---|
| pilot, 3 resources | — | — | — | 0% by construction |
| 100-resource run | — | — | — | — |
| 239-resource run | 75.2% | not measured | 82.3% | 92.8% |
| after map + JSON well-known types | **80.1%** | **98.2%** | **85.5%** | **93.9%** |

The 239-run figures are quoted here over the 93-resource intersection, not the 98 the original
write-up used, so the columns are comparable. Over its own 98 that run read 73.7 / 82.8 / 91.4.

## Compile conformance

The measure closest to "would this resource actually work". Every resource upstream implemented has
hand-written `_identity.go` / `_reference.go` files, left at upstream's version while the types
underneath were regenerated; if the package builds, the generated types satisfy the code a person
wrote against them.

| run | packages failing | distinct fields the hand-written code needs and we lack |
|---|---|---|
| after the map + JSON well-known types run | 44 | 29 |

Almost all 29 are a parent segment or a parent reference — `Spec.OrganizationRef` 14,
`Spec.Location` 14, `Spec.EntryGroupRef` 6, `Spec.DatabaseRef` 6. Reproduce with
`go build ./apis/... 2>&1 | grep -c '^#'`.

It covers 18% of what the CRD comparison finds and only 44 of 189 resources, because it sees only
fields the controller code dereferences. Track both.

## Population, not sample

The rows above track a fixed 93-resource intersection so the trend is honest. Separately, the whole
greenfield population has now been defined and measured once — see
[greenfield-population.md](greenfield-population.md).

| | count |
|---|---|
| greenfield resources upstream | **300** |
| regenerable by the Step 1 pipeline | 231 |
| actually regenerated | 225 |
| scorable (CRD not stale) | **189** |

Measured over those 189: spec **78.5%** (**95.7%** excluding missing references), required 78.1%,
observedState 88.5%. References: 484 in baseline, 168 reproduced, 203 plain strings, 113 absent.

Do not read 78.5% against the 80.1% above as a regression. It is a larger and differently-defined
population — every v1alpha1 resource whose `SupportedControllers` is `[Direct]` or which has no
controller registered yet, rather than every resource that happened to have a `generate.sh` line.

**The old 239 set contained 12 Terraform-backed resources** — `APIKeysKey`, seven `Compute` kinds,
two KMS kinds, `BigQueryReservationCapacityCommitment`, `VertexAITensorboard`. Their KRM shape was
set by the Terraform provider, so scoring generator output against them asks it to reproduce a TF
artifact, and they are excluded now.

## Accounting: is every field KCC master has either produced or flagged?

Past ~80% generated, the useful target changed. A field needing a human is a fine outcome; a field
nobody was told about is not. See
[greenfield-coverage-invariant.md](../greenfield-coverage-invariant.md).

Over the same 189 greenfield resources, against KCC baseline `c1df0b9326`, counting **roots** rather
than paths:

| | fields |
|---|---|
| in KCC master's CRDs | 9,310 |
| we produce | 8,727 (93.7%) |
| **we miss** | **583 (6.3%)** |

Two axes, not one, and they are independent: the class says *why* a field differs, the columns say
*whether anyone was told*. A field can be `absent` and flagged at once — that is the queue working.
An earlier version of this table put both on one axis and read as a contradiction.

`renamed` and `intentionally-different` are excluded from the target: the first is a casing table,
the second is the `google.protobuf.Value` arms we map to `apiextensionsv1.JSON` on purpose, a
decision deliberately deferred. Both are still reported.

| class | we miss | flagged by field | by section | unflagged |
|---|---|---|---|---|
| `reference-shape` | 314 | 180 | 0 | 134 |
| `moved` | 102 | 36 | 41 | 25 |
| `absent` | 85 | 40 | 1 | 44 |
| **subtotal, the target** | **501** | **256** | **42** | **203** |
| `renamed` (accepted) | 22 | 0 | 0 | 22 |
| `intentionally-different` (accepted) | 60 | 0 | 0 | 60 |

How the target got from 450 to 92:

| | unflagged |
|---|---|
| first published | 450 |
| de-duplicate repeated fields and reference children | 412 |
| run `scripts/queue-hints` — a tool that already existed and had never been run | 276 |
| pair the suffix upstream drops when it adds `Ref` (`kmsKeyName` → `kmsKeyRef`) | 256 |
| gate the seeder per Kind, so it stops queueing hand-written upstream resources | 259 |
| flag `empty-observedstate` on the 36 resources that generated no status at all | 217 |
| name every omitted parent segment (`spec.collection`, `spec.tenant`, `spec.lake`) | 203 |
| separate renamed references from undetected ones | 109 |
| walk map values in `queue-hints`, which never entered a map | 97 |
| two measured name rules, `secret` and `project` | **92** |

"We produce" held at 8,727 throughout. That check matters: a change that flags fields by no longer
producing them improves the report and damages the CRD.

What is left, and why, is in
[greenfield-detection-gaps.md](../greenfield-detection-gaps.md).

## What each change was worth

| change | effect |
|---|---|
| ObservedState prepopulation | 0% → ~91%; the single largest move so far |
| output-only comment detector | folded into the above |
| plural acronyms (`--emit-plural-acronyms`) | off by default; renames 83 fields across 23 packages |
| Location from the proto's parent shape | `spec.location` misses 13 → 4 resources; absent refs 47 → 35 |
| `map<string, Message>` | 40 previously-dropped fields across 18 services |
| `Value`/`ListValue`/`Struct` → `apiextensionsv1.JSON` | removes a wire-format bug; **lowers** the score, see below |
| reference-hint seeder | zero effect on any rate — it writes queue entries, not fields |
| running that seeder against the corpus | reference-shape unflagged 298 → 147 |
| `empty-observedstate` flag | `moved` unflagged 66 → 25 |

## Where the remaining gap is

Of 604 missing spec paths in the latest run, **576 are references and 28 are everything else.**

Each unreproduced reference costs four or five missing paths (`fooRef`, `.external`, `.name`,
`.namespace`, sometimes `.kind`), so 576 paths is roughly 130 actual references. The metric amplifies
reference debt fourfold, which is why the raw spec rate reads so much worse than the generator is.

References themselves: 237 in the baseline, **93 reproduced**, 109 left as plain strings, 35 absent.

Non-reference remainder, after discounting references whose names do not end in `Ref` and so escape
the classifier: roughly **15 fields across five resources**. observedState has **70** missing, none
of them references, 43 of which sit in four resources (`DataLabelingDataset`, `RunWorkerPool`,
`DataLabelingEvaluationJob`, `DataLabelingInstruction`).

**Reference support is the next thing worth building.** Nothing else in the measured set is close.

## Two things the metric cannot tell you

**It penalises correctness.** `AIPlatformModel` and `VertexAITrainingPipeline` each gained 20 missing
fields when `google.protobuf.Value` started mapping to `apiextensionsv1.JSON`. The 20 are the
`boolValue` / `nullValue` / `numberValue` / `stringValue` / `structValue` arms of upstream's
hand-written union, whose recursive arm is commented out because it destabilises the CRD. Replacing
it fixed KCC sending Vertex AI a double-encoded `trainingTaskInputs`. The score went down because the
output got better. Any change that corrects something upstream got wrong will read as a regression.

**A stale tree scores as a good one.** A service that fails CRD generation keeps its previous CRD,
which is often identical to the baseline, and scores a false 100%. The measured set therefore counts
a resource only when its `_types.go` actually differs from baseline. `compute` scored this way in the
239-run.

## Traps that have each cost a full run

* **`--prepopulate-spec` does not rewrite an existing types file.** Regenerating in place leaves the
  old Spec and measures a generator version that is no longer current. Delete the resource first;
  that step is the experiment, not tidying.
* **`generate.sh` calls `generate-crds`, which runs `controller-gen` tree-wide.** One unloadable
  package panics the CRD step for every service. Use `SKIP_GENERATE_CRDS=1`, then `controller-gen`
  per service.
* **`WRITE_GOLDEN_OUTPUT=1` absorbs unrelated drift**, reliably `IAPSettings.diff`. Read the diff.
* **Validate the analyser against a known answer before trusting a new number.** Three figures in the
  239-run drafts were measurement bugs. The current analyser reproduces the published
  73.7 / 82.8 / 91.4 from the preserved score file before being pointed at anything new.

## Sources

| run | write-up | raw data |
|---|---|---|
| pilot, 3 | [pilot-3-replication.md](pilot-3-replication.md) | — |
| 100 | [pilot-100-replication.md](pilot-100-replication.md) | — |
| 239 | [replication-239.md](replication-239.md) | [data-239/](data-239/) |
| 239 re-run | [replication-239.md](replication-239.md#re-run-after-the-map-and-json-well-known-type-work) | [data-239-rerun/](data-239-rerun/) |
