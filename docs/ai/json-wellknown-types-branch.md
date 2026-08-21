# JSON well-known types in the generator

`google.protobuf.Value`, `ListValue` and `Struct` now map to `apiextensionsv1.JSON`, and a map whose
value is a message generates instead of being dropped. The work sits on a standalone branch off
`upstream/master`, because it is a fix KCC needs regardless of anything in the greenfield sandbox.

**Branch** `generator-json-wellknown-types`, worktree `/home/luks/kcc-wkt`, based on
`0948781ecb`. It carries none of the sandbox work: no Location change, no judgement queue, no
`generate.sh` repairs.

It is **two commits, narrow first**, so the scope decision stays mechanical:

* `3155caea82` — the JSON well-known types, plus the two latent bugs that change alone trips.
  Complete and reviewable on its own; 31 files, five services.
* `91a2c9ab1f` — `map<string, Message>` support and the thirteen further services it brings back;
  35 files.

`git reset --hard HEAD~1` gives you the narrow PR. Keeping both gives the wide one. Nothing needs
unpicking by hand either way.

## Why this was needed

`google.protobuf.Value` is a union, and one of its arms is a `ListValue` whose values are `Value`s
again. Generating Go structs for that pair produces a mutual recursion, and `controller-gen` cannot
build a terminating OpenAPI schema for it — so it fails on the whole package, not on the one field.

KCC has been working around this by hand for a while, in two different ways:

* `apis/aiplatform/v1alpha1/recursive_types.go` declares both types by hand with the recursive field
  commented out and a note saying "ListValue is temporarily disabled due to CRD instability". The
  resulting type cannot represent a list at all.
* `apis/firestore/v1alpha1` types Firestore's `Document.fields`, a `map<string, Value>`, as
  `map[string]apiextensionsv1.JSON`.

The second of those is the right answer, and it is what the generator now does everywhere. The
generator already treated `google.protobuf.Struct` this way; `Value` and `ListValue` were simply
missing from the same list.

## The change

Four hunks of actual logic. Everything else in the diff is regenerated output or tests.

| file | what |
|---|---|
| `pkg/codegen/common.go` | `Value` and `ListValue` join `Struct` in `protoMessagesNotMappedToGoStruct`, mapping to `apiextensionsv1.JSON` |
| `pkg/codegen/typegenerator.go` | a map value whose message has a special-cased Go type takes that type, rather than a struct name it does not have |
| `pkg/codegen/mappergenerator.go` | two cases in the `krm{From,To}ProtoFunctionName` switches, which `klog.Fatalf` on anything unlisted |
| `pkg/controller/direct/maputils.go` | `Value_{From,To}Proto` and `ListValue_{From,To}Proto`, mirroring the existing `Struct_*` pair |

The converters round-trip through `AsInterface`/`NewValue`, the same JSON-compatible Go types
`json.Marshal` already produces, so they compose the way `Struct`'s do. A `Value` holding NaN or
+Inf has no JSON form and `NewValue` rejects it; that surfaces as a `mapCtx` error rather than
silent data loss.

## Three latent bugs it exposed

None of these were introduced by the change. Each was already wrong and simply had nothing reaching
it, and each blocked compilation once something did.

**The pruner orphaned imports.** `prunetypes` comments out unreachable types *after* the file is
written, so the import decision is made while the type is still live. An unused import is a compile
error in Go, and nothing catches it later — `goimports` runs over `pkg/controller/direct`, never
over `apis/`. Fixed by re-parsing the pruned file and dropping imports nothing references outside a
comment. Parsing rather than string-matching is what makes it correct: the commented-out type is
still in the file and still spells the qualifier.

**Pointer bridging was hardcoded to one function name.** `GoTypeForField` deliberately rewrites
`*apiextensionsv1.JSON` to `apiextensionsv1.JSON`, but the converters take and return a pointer, so
call sites have to bridge. That bridging was spelled as a comparison against the literal string
`"direct.Struct_ToProto"`, which covered neither the other proto messages mapping to the same Go
type nor any field inside a `oneof`. Now keyed off the mapped Go type, and applied in the oneof
branch too.

**Map-of-message punted to a helper nobody wrote.** For any map shape it did not recognise, the
mapper generator emitted a call to `<FieldName>_FromProto` and expected a human to supply it. In
practice nobody did, so the generated mapper did not compile. It now writes the loop itself, for
both message values and values with a special-cased Go type.

### Three more, in the map-of-message path

These only appear in the second commit, and only surfaced when more than one service was built. The
first pass had been verified against `batch` alone, which happened to be the shape that worked.

* **The KRM side is not always a map.** `artifactregistry`'s `Repository.cleanup_policies` is a
  proto map that KCC models by hand as a `[]CleanupPolicy` slice, so emitting a map loop for it does
  not compile. The loop is now guarded on the KRM field actually being a `map[string]...`, which
  routes it back to `CleanupPolicies_FromProto` — a hand-written map/slice bridge that already
  exists in `pkg/controller/direct/artifactregistry/mapper.go`.

  The cause is worth knowing, because it is not a one-off. It comes from Terraform: the SDK's
  `TypeMap` holds only primitive values, so a proto `map<string, Message>` has to become a
  `TypeSet` of objects with the key promoted to a required attribute. The vendored schema at
  `third_party/.../resource_artifact_registry_repository.go` still describes the field as
  "Map keys are policy IDs supplied by users" while declaring `Type: schema.TypeSet`. KCC's
  TF-backed CRDs inherited that shape, and 19 proto-map field names are currently modelled as KRM
  slices this way. Every TF-to-direct migration of a resource with a proto map will hit the same
  collision, and it is silent until a mapper fails to compile.
* **Generated converters carry a version suffix.** The loop called `CommonUsageStats_FromProto`
  where the generated function is `CommonUsageStats_v1alpha1_FromProto`. The non-map path already
  appends `versionSpecifier`; the map path did not. The special-cased converters in the `direct`
  package take no suffix, so the two cases differ.
* **A pruned value type comes back.** `datacatalog`'s `CommonUsageStats` was commented out as
  unreachable, because nothing referenced it while the map field was being dropped. Once the field
  generates, the type is referenced and the pruner leaves it live. That one turned out to be correct
  behaviour rather than a bug, but it looks alarming in the diff and is worth knowing about.

## A real bug in what KCC sends Vertex AI

`TrainingPipeline.training_task_inputs` is a `google.protobuf.Value`. Its JSON encoding is just the
value. KCC was sending the *protobuf internal representation* instead — visible in the recorded
traffic as a literal `"fields"` key wrapping `"listValue"` and `"stringValue"` arms:

```json
"trainingTaskInputs": {"fields": {"workerPoolSpecs": {"listValue": {"values": [...]}}}}
```

and now sends

```json
"trainingTaskInputs": {"workerPoolSpecs": [{"machineSpec": {...}, "replicaCount": "1"}]}
```

The mock echoes whatever it is given, so no test caught it. Two fixtures encoded the old shape and
have been rewritten; `dev/ci/presubmits/tests-e2e-fixtures-suite aiplatform` passes.

Note the old fixtures used a `listValue:` key that the CRD never declared, because that arm was the
one commented out in `recursive_types.go`. It survived only because it sat inside a `structValue`,
which was already free-form JSON.

## What it deletes

The point of the change is that hand-written workarounds go away.

* `recursive_types.go` loses both stand-in types: 38 lines out, 12 in.
* `pkg/controller/direct/aiplatform/model_mapping.go` loses its local `Value_{From,To}Proto`, which
  had no list arm and so could not round-trip a list: 114 lines out, 20 in.

## The scope fork, still open

Two separable changes ended up on one branch, and they should probably not land together.

**Narrow** — the JSON well-known types only. `map<string, Value>` works; `map<string, SomeMessage>`
stays declined as it is today. Touches about five services, needs no mapper-loop code, and is small
enough to review in one sitting.

**Wide** — what is on the branch now. Also enables `map<string, Message>` generally, which brings
back **40 previously-dropped fields across 18 services**: aiplatform, artifactregistry, batch,
billingbudgets, ces, cloudbuild, compute, contactcenterinsights, container, datacatalog, dataflow,
datalineage, firestore, gkehub, hypercomputecluster, managedkafka, networkconnectivity, vertexai.
Real fields — GKE maintenance exclusions, Compute preserved-state disks, Artifact Registry cleanup
policies, Data Catalog tag fields.

The recommendation is to send the narrow one first. The wide one changes generated output in a lot
of services at once, which is a harder review, and it depends on the mapper-loop code that the
narrow one does not need.

## Blast radius, as measured

Checked against `upstream/master` before making any change, because `Value` is an ambiguous name.
`contentwarehouse` and `firestore` declare their *own* `Value` messages, which are untouched.

Packages carrying `google.protobuf.Value` or `ListValue`: aiplatform, billingbudgets, ces,
datalineage, vertexai.

**No served v1beta1 CRD schema changes.** `billingbudgets`'s copies are purely self-referential, so
nothing reaches them. `vertexai/v1beta1`'s `Dataset.metadata` is not reachable from the served CRD
either — no `nullValue` appears anywhere in it. The only two CRDs whose schema actually changes are
`aiplatformmodels` and `vertexaitrainingpipelines`, both v1alpha1.

Those two do change in a user-visible way: a hand-crippled union with `boolValue`/`nullValue`/
`numberValue`/`stringValue`/`structValue` arms becomes `x-kubernetes-preserve-unknown-fields: true`.
Old YAML still validates, but it is *interpreted* differently, so existing objects change meaning.
That is worth calling out explicitly in the PR description rather than leaving to be discovered.

## Verification done

Run against **each commit separately**, not just the final tree, because the point of the split is
that either one stands alone.

* `go build ./...` clean on both.
* `go test ./pkg/controller/direct/` — round-trip tests for each arm of the `Value` union, including
  the two nested ones that make the recursion, plus nil and empty-list cases.
* `go test ./pkg/codegen/` — `TestGoTypeForFieldMaps` pins the type string across six map shapes;
  `TestWriteMessageMaps` pins rendered output, which is what distinguishes a field typed wrongly
  from one that is not there at all. Maps had no coverage at all before, including the two forms
  that already worked.
* `go test ./tests/apichecks/...` passes on both, with `alpha-missingfields.txt` absorbed
  separately in each — it changes in both, so absorbing once at the end would put half the churn in
  the wrong commit.
* `dev/ci/presubmits/tests-e2e-fixtures-suite aiplatform` passes on both.
* **The split was tested for real**: `git reset --hard HEAD~1` onto the narrow commit still builds
  and passes `go test ./pkg/codegen/ ./pkg/controller/direct/`. That is the property the two-commit
  shape exists to provide, so it was exercised rather than assumed.

`dev/tools/controllerbuilder/pkg/commands/iteratetypes` fails `go vet` with a `fmt.Errorf` verb
mismatch. That is pre-existing on `upstream/master` in a file this branch does not touch.

## Notes for whoever picks this up

The `Value`/`ListValue` mapping is the whole fix; everything else exists because enabling it made
three dormant code paths run for the first time. If the wide version is dropped, the mapper-loop
code in `mappergenerator.go` can go with it, but the pruner fix and the pointer-bridging fix are
needed either way — the narrow version still produces `map[string]apiextensionsv1.JSON`, which trips
both.

Do not try to keep the generated `Value` struct alongside the JSON mapping. The two representations
disagree about what a `Value` field means on the wire, which is exactly the Vertex AI bug above.
