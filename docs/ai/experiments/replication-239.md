# Replication experiment: 239 resources

Take resources upstream implemented by hand, delete them, regenerate them through the Step 1
pipeline, and score the result against the originals. This is the third and largest such run; the
pilot covered 3 resources and the previous run 100.

Read it alongside [pilot-3-replication.md](pilot-3-replication.md) and
[pilot-100-replication.md](pilot-100-replication.md), which it supersedes in scale but not in
conclusions.

## What was measured

409 kinds carry a `--resource` line in a `generate.sh`. **239 are in scope.** The other 170 are
excluded for reasons unrelated to generator quality: 163 are not `v1alpha1` invocations, 6 sit in
services passing `--skip-scaffold-files` so the generator never writes their types file, and 1 has
no types file at all.

92 services regenerated, 90 without error. 127 resources came back genuinely regenerated. **98 were
scored**, after excluding `compute` and `networksecurity`, whose per-service CRD generation failed
and left stale or truncated files that would have produced meaningless comparisons.

## Result

| bucket | rate | resources with no missing or mismatched field |
|---|---|---|
| spec | 73.7% | 37/98 |
| required | 82.8% | 59/98 |
| status.observedState | 91.4% | 71/98 |

**ObservedState is the headline.** Before this session it was 0% by construction: the scaffolder
emitted an empty struct and a human filled it in. It is now produced mechanically in the first pass.

### What these numbers are, and are not

They are measured **after** this session's generation changes — ObservedState prepopulation, the
output-only comment detector and the plural-acronym flag all landed before the regeneration at
`fabd291f09`, which is why ObservedState reads 91.4% rather than 0%.

They are measured **before any reference work**, and no reference has been implemented. The queue
seeder and the unsupported-type entries described below change what the queue *reports*; neither
emits a single extra CRD field. The spec rate is unchanged by them and stays at 73.7% until someone
actually implements references.

### What the spec number is made of

73.7% on its own misleads. Of 859 missing field paths:

| cause | paths | share |
|---|---|---|
| reference fields, deferred to the judgement pass by design | 624 | 72.6% |
| `map<string, Message>` the generator cannot type | 124 | 14.4% |
| other | 105 | 12.2% |
| plural acronyms that slipped past the flag | 6 | 0.7% |

Excluding references it is 90.8%; excluding the map limitation too, 95.2%.

### The "other" bucket, unpicked

It is three unrelated things, and only one is a generator defect.

**63 fields are genuinely dropped** — the real finding. Whole objects vanish with their subtrees:
`BigQueryMigrationMigrationWorkflow.spec.tasks`, and `HypercomputeClusterCluster`'s
`spec.computeResources` and `spec.networkResources`. `AnalyticsAccount.spec.redirectURI` is a plain
string that is simply absent. These are silent and uncharacterised, and they are the part worth
chasing.

**25 are reference children misfiled by my classifier**, not a separate cause. Paths like
`NetworkManagementConnectivityTest.spec.destination.sqlInstance.external` are the `external`, `name`
and `namespace` children a reference expands into; the classifier missed them because the parent is
named `sqlInstance` rather than `sqlInstanceRef`. The reference gap is correspondingly larger than
152 fields and this bucket smaller.

**14 are a bug in the measurement harness, not the generator.** The Location normaliser recorded
each resource's original form by looking for a `Location` field in its `_types.go`. Upstream often
does not declare one: `DataplexLakeSpec` inlines `parent.ProjectAndLocationRef`, which contributes
both `projectRef` and `location` to the CRD without a `Location` field existing. The normaliser read
that as "no location", deleted the scaffold's, and the comparison then reported it missing. Verified
on `DataplexLake`, whose upstream CRD has exactly `location` and `projectRef` — the same shape the
scaffold emits, so these would have matched untouched. The spec rate is understated by this amount.

## What is missing, and whether anyone can see it

This is the table that matters, and it is less flattering than the coverage figure. Counting
distinct fields rather than paths, 387 are missing:

| category | in the queue today | seeder would add | new TODO entry | silent | total |
|---|---|---|---|---|---|
| reference | 5 | 47 | 0 | 100 | 152 |
| map-of-message | 0 | 0 | 124 | 0 | 124 |
| other | 0 | 0 | 0 | 105 | 105 |
| plural acronym | 0 | 0 | 0 | 6 | 6 |
| **total** | **5** | **47** | **124** | **211** | **387** |

**45% visible, 55% silent** — even with both queue improvements applied. By CRD field path rather
than distinct field the figures are 859 missing and 331 visible (39%); references expand into four
paths each (the object plus `external`, `name` and `namespace`), which is why the two counts differ.

Two things follow. `map<string, Message>` is the one unambiguous win: 124 fields, every one silent
before, every one now carrying a queue entry. And **"other" at 105 fields is both entirely silent
and entirely uncharacterised** — comparable in size to the whole reference gap the seeder closes,
and the largest thing we still cannot explain.

Note that the 54% reference recall quoted elsewhere is against the 111 references whose original
field could be paired, not against all 152 missing reference fields.

## Defect taxonomy

Six classes blocked resources from compiling after regeneration. All are in the scaffold and pruning
layers; none is in the ObservedState or acronym work, which behaved correctly everywhere it ran.

1. **`Location` emitted unconditionally, and as a non-pointer `string`.** 174 of 239 resources
   disagree with the original. Where the proto already has a `location` field it is a straight
   redeclaration. The generator already computes `ParentStyle` from `google.api.resource`, so
   emitting it only for location-scoped resources is derivable: right for 145 of 167.
2. **Pruning removes types the same file still references.** `contentwarehouse` refers to
   `Document_Style` and `Document_Page` after `--prune-unused-types` deleted them.
3. **Kind names collide with generated type names.** `APIHubInstance`, `BillingAccount` and
   `Document` end up declared twice.
4. **Parent shape reaches `_identity.go` but not the Spec template.** `cloudsecuritycompliance`
   expects `Spec.OrganizationRef`; the scaffold emits `ProjectRef`.
5. **`apiextensionsv1.JSON` emitted without its import.** Same class as the `common.Status` bug
   fixed this session, so the fix generalises.
6. **Two `generate-types` invocations writing one `types.generated.go`**, the second overwriting the
   first (`networksecurity`).

## Reference detection

The judgement queue is meant to list what a human still has to decide. Measured against the 111
confirmed references, it named **11 — 10%**. `CloudBuildConnection` needs 15 references and the
queue listed none of them.

The cause is not a bug: `judgementFor` consults only `google.api.resource_reference`, which is
present on a minority of fields. A post-generation seeder reusing the rules from `TestMissingRefs`
reaches **60 correct against 39 wrong — 61% precision, 54% recall.** For a list a person confirms
rather than an automatic rename, that trade is reasonable; a wrong hint costs a moment's reading.

What it still misses, by target type:

| target | confirmed | missed |
|---|---|---|
| `SecretManagerSecretVersionRef` | 12 | 12 |
| `SecretRef` | 13 | 11 |
| `ComputeNetworkRef` | 10 | 8 |
| `KMSCryptoKeyRef` | 6 | 2 |
| `IAMServiceAccountRef` | 8 | 1 |
| `ServiceDirectoryServiceRef` | 7 | 1 |

Secrets are half the remaining gap. `ComputeNetworkRef` is the second largest, and also the riskiest
to chase: "network" is the exact name the earlier heuristic was rejected over.

### One idea that measured badly

Flagging fields whose name contains URI or URL looked promising and is wrong on this corpus. **None
of the 111 confirmed references is URI-named.** There are 9 URI-named string fields and upstream
made references of none — they are container images, GCS object paths, an OAuth redirect. The
existing policy of treating them as not-representable is correct. Caveat: 9 is a small sample,
enough to say "do not invert this", not enough to say "never".

## Map support, and how it was verified

`map<string, Message>` was one of the dropped-field categories. It now generates, and this is the
evidence that the output is right rather than merely present.

**Against upstream, byte for byte.** Two resources were regenerated and their map declarations
compared with the hand-written upstream versions at `c1df0b9326`:

| resource | generated | upstream | |
|---|---|---|---|
| `BigQueryMigrationMigrationWorkflow` | `Tasks map[string]MigrationTask` | identical | match |
| `NetworkServicesWasmPlugin` | `Versions map[string]WasmPlugin_VersionDetails` | identical | match |
| `HypercomputeClusterCluster` | `map[string]NetworkResource` | `map[string]*NetworkResource` | differs |

The one difference is the value type's pointer-ness. The corpus settles it: across all `v1alpha1`
and `v1beta1` types, 16 map declarations use the value form and 7 use a pointer, so upstream's
`hypercomputecluster` is the minority spelling rather than a convention being violated. A map value's
pointer-ness also does not change the OpenAPI schema, so no CRD differs either way. Recorded, not
"fixed".

Worth noting about the second row: `networkserviceswasmplugin_types.go` had **not** in fact been
regenerated when this was first checked, and its old-format `// TODO: unsupported map type` was
mistaken for a match. It was regenerated for real and then compared. The tell was the TODO text,
which lacks the field name that `WriteField` now emits — old-format markers are a reliable sign that
a file predates the change you are trying to measure.

**Unit tests**, in `dev/tools/controllerbuilder/pkg/codegen/typegenerator_test.go`. Maps had no
coverage at all before this, including the two forms that already worked, which matters more than
usual here: when `GoTypeForField` declines a type, `WriteField` swallows the error, leaves a
`// TODO:` comment and carries on, so a regression removes fields from the CRD without failing
anything.

- `TestGoTypeForFieldMaps` covers the type string: `map<string,string>`, `map<string,int64>`,
  `map<string,Message>`, a non-string key, and a `google.protobuf.Value` value.
- `TestWriteMessage` covers the rendered output, which is what distinguishes a field that is typed
  wrongly from one that is not there at all. It now carries a supported map, a map of message, and a
  declined one, pinning the `// TODO:` line.

Reverting the map branch fails all five type cases; reverting only the map-of-message arm fails that
one. Confirmed by doing it.

**Sizing the remainder.** Across the Google API protos there are 1002 `map<string, Message>` fields.
The ones still declined are all JSON-shaped: `google.protobuf.Value` (88), `google.protobuf.Struct`
(8) and `google.protobuf.ListValue` (4). So roughly 900 of 1002 generate now, and the whole
remainder is one follow-up — mapping those three to `apiextensionsv1.JSON`, which is what upstream
writes by hand in `apis/firestore/v1alpha1` and `apis/aiplatform/v1alpha1/recursive_types.go`.

**A hazard found while writing the tests.** The map branch happily generated `map[string]Value`.
`Value` and `ListValue` are mutually recursive — `Value.list_value` is a `ListValue`,
`ListValue.values` is a `[]Value` — and `controller-gen` cannot build a terminating schema for that,
so it fails on the whole package rather than on the one field. This is the `DataLineageProcess`
failure, and upstream hit it too: `recursive_types.go` declares both types with the `ListValue`
field commented out "due to CRD instability". The generator now declines these as map values, which
drops the field visibly into the queue instead of producing a tree that cannot generate CRDs.

## Re-run, after the map and JSON well-known type work

The same experiment, run again once three more generation changes had landed: the Location fix,
`map<string, Message>`, and `Value`/`ListValue`/`Struct` mapped to `apiextensionsv1.JSON`. The
reference-hint seeder and the judgement queue also landed in between, but those change only what the
queue reports and cannot move a CRD field, so they were expected to show nothing here.

**Like for like, on the same 98 resources scored the first time.** This is the only comparison that
controls for the measured set shifting underneath the run.

| bucket | before | after | clean before | clean after |
|---|---|---|---|---|
| spec | 73.7% | **75.0%** | 37/98 | 39/98 |
| required | 82.8% | **83.0%** | 59/98 | 60/98 |
| status.observedState | 91.4% | **91.4%** | 71/98 | 71/98 |

Five of those 98 sit in services whose CRDs are stale this run, so their rows are carried over rather
than re-measured. Dropping them gives 93 fully comparable resources, where spec moves 75.2% to
**76.6%** and the other two buckets behave the same way.

**The movement is three resources, and all three are maps.**

| resource | spec fields missing, before → after |
|---|---|
| `BigQueryMigrationMigrationWorkflow` | 36 → 0 |
| `NetworkServicesWasmPlugin` | 8 → 0 |
| `DataLineageProcess` | 1 → 0 |

That is the whole of the +1.3. It is a small number for a change that recovered 40 fields across 18
services, and the reason is that the measured 98 is a sample: only three of them had a
`map<string, Message>` field at all. The benefit is real but lands mostly outside the sample, which
is worth remembering before reading this rate as a measure of overall generator quality.

**The Location fix moved nothing measurable.** No resource in the comparable set changed its spec
count because of it. That is not evidence it is wrong, only that the resources it affects are not in
this sample.

### The metric penalises the JSON well-known type fix

Two resources went the other way, `AIPlatformModel` and `VertexAITrainingPipeline`, each gaining 20
missing spec fields. Every one of them is a `Value` union arm:

```
spec.explanationSpec.metadata.inputs.KEY.inputBaselines[].boolValue   (boolean)
spec.explanationSpec.metadata.inputs.KEY.inputBaselines[].nullValue   (string)
spec.explanationSpec.metadata.inputs.KEY.inputBaselines[].numberValue (number)
spec.explanationSpec.metadata.inputs.KEY.inputBaselines[].stringValue (string)
spec.explanationSpec.metadata.inputs.KEY.inputBaselines[].structValue (json)
```

The baseline is upstream's hand-written `Value` struct, whose `list_value` arm is commented out
because it destabilises the CRD. Scoring against it rewards reproducing that shape. The change
replaces it with `x-kubernetes-preserve-unknown-fields`, which is what the field actually is, and is
what fixed KCC sending Vertex AI a double-encoded `trainingTaskInputs`. So the score drops because
the output got better.

Neither resource is in the comparable 98, so the headline is unaffected — but the effect is
structural, not a one-off. **Any change that corrects a shape upstream got wrong will read as a
regression here.** The score answers "how close are we to what upstream wrote", which is the right
question for replication and the wrong one for correctness.

### Exclusions moved, in both directions

| service | before | after | why |
|---|---|---|---|
| `ces` | excluded | **scores** | its exclusion was `undefined: apiextensionsv1`, the orphaned-import bug the `prunetypes` fix addresses |
| `compute` | excluded | **scores** | per-service CRD generation succeeds now; previously it failed and left CRDs identical to baseline, scoring a false 100% |
| `notebooks` | scored | **excluded** | regression from this session's own fix, below |
| `networksecurity` | excluded | excluded | `zz_generated.deepcopy.go` is still truncated |
| `backupdr` | excluded | excluded | `v1beta1` references `v1alpha1.Parent` |

`compute` returning is why the full measured set now reads differently from the 98: its resources
previously scored a false 100% by not being regenerated at all, and now score honestly.

**The `notebooks` regression is self-inflicted and worth recording.** Qualifying
`NotebookInstanceV2` to `notebooks.v2.Instance` recovered the 39 fields it had been losing to the v1
proto — but both versions now generate into one package, and message names common to both collide:
`ContainerImage` and `VMImage` are each declared twice, so the package does not compile. Fixing the
proto ambiguity created a Go name ambiguity. Two resources drop out of the measured set as a result.

### Method

Identical to the first run, and checked rather than assumed: the analyser reproduces the published
73.7 / 82.8 / 91.4 and 37 / 59 / 71 exactly from the preserved `data-239/score239.txt` before being
pointed at the new scores. Given that three figures in this document's drafts turned out to be
measurement bugs, pinning the tooling against a known answer first is cheap insurance.

One operational note that cost a full false start: `apis/<svc>/generate.sh` calls
`dev/tasks/generate-crds`, which runs `controller-gen` tree-wide, so one unloadable package panics
the CRD step for **every** service. All 92 failed identically until the run was split into
`SKIP_GENERATE_CRDS=1` for types and mappers, then `controller-gen` per service. That is this
document's own recorded lesson from the first run, and it still had to be learned twice.

Inputs preserved in [data-239-rerun/](data-239-rerun/).

## Lessons for the evaluation framework

These cost more time than the generator defects did.

- **Generate CRDs per package.** `dev/tasks/generate-crds` runs `controller-gen` with
  `paths="./..."`, so one unloadable package blocks all 92 — and when it panics on an
  unresolvable type it names no package at all, leaving nothing to attribute the failure to.
  Per-service invocation succeeded for 90 of 92 and identifies failures by construction.
- **Commit the expensive artifact the moment it exists.** An iterative cleanup script destroyed a
  full regeneration that had not been committed, costing an hour.
- **Never derive a filename by guessing.** Resolving a Kind's CRD as `lowercase(kind)+"s"` silently
  dropped 11.5% of kinds: `*Policy` → `policies`, `Batch` → `batches`, `Process` → `processes`.
- **Report at the granularity the generator writes at.** `types.generated.go` is per-service, so an
  error in it cannot be attributed to one resource, and per-resource isolation cannot work.
- **Separate "wrong output" from "differs from pre-existing consumers".** Most compile failures were
  the second: regenerating `_types.go` while keeping the hand-written `_identity.go` leaves the new
  Spec having to satisfy contracts written against the old one. `generate-types` cannot regenerate
  identity files — only `generate-controller` writes those — so no deletion strategy fixes this, and
  a genuinely new resource would never hit it.

## Methodology

Recorded so the figures can be regenerated rather than believed. Three of the numbers in this
document's drafts turned out to be measurement bugs, each caught only because a sample looked wrong:
a plural-guessing bug that hid 9 resources, a path-normalisation bug that reported 0% precision when
the answer was 61%, and a reference-matching bug that undercounted queue visibility nearly fourfold.

**Baseline `c1df0b9326`.** It predates every experiment commit, so all resources there are
hand-written upstream work. `ae2abcdb9b` is *not* usable: it sits after the pilot re-run, so three
services are generator output there. Verified on `GrafeasNote`, whose CRD has no `location` at the
baseline but does at `ae2abcdb9b`.

**Comparison** by `crd-mcp-server score`, which flattens both CRDs to `path -> type` maps and
reports matched, missing, extra and mismatched per bucket. It reuses `walk`/`flatten` from
`compare.go`, so it sees schemas exactly as the existing equivalence check does.

**Resource resolution** from each CRD's own `kind:` field, never from the filename.

**Map value pointer-ness** counted on `upstream/master`, not the working tree, which the experiment
has modified:

```sh
git grep -hoE "map\[string\]\*?[A-Z][A-Za-z0-9_]*" upstream/master -- 'apis/**/*.go'
```

Halve the totals: every declaration appears exactly twice, once in the types file and once in
`zz_generated.deepcopy.go`. That gives 16 value to 7 pointer. Restricting the glob to types files
instead gives 16 to 4, because the deepcopy copies are not evenly distributed across the two forms —
so quote the halved figure, not a filtered one.

**The measured set** counts a resource only if its `_types.go` actually differs from baseline; a
restored resource would otherwise score a false 100%.

**Reference pairing**: the score tool reports each baseline reference path alongside the plain field
we generated in its place, so pairs are read from the comparison rather than reconstructed. Target
types resolve by matching the JSON tag in upstream's Go source; deriving the Go field name from the
path instead left 15 rows unresolved and mis-stated several per-type counts.

Inputs are preserved in [data-239/](data-239/): the in-scope list, the measured set, raw scores, the
confirmed pairs, the seeder's proposals, the exclusions, and the Location divergences. The reference
answer key is at `hack/tools/greenfield/reference_testset.tsv`.
