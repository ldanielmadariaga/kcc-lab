# Greenfield generator: what was measured, and what was abandoned

Evidence behind the four phases in
[`greenfield-generator-mechanics.md`](greenfield-generator-mechanics.md). Nothing here is needed to
understand what the generator does; it is here because several of these results overturned the
assumption the design started with, and because each one is easy to repeat by accident.

**Scope:** the experimental sandbox (`kcc-lab`). Measurements are point-in-time against googleapis
and the tree as it stood when each phase landed.

## How far proto annotations reach

Measured across googleapis, excluding the ads services:

| | count | share |
|---|---:|---:|
| Total proto fields | 109,515 | |
| Carrying any `field_behavior` | 44,577 | 40.7% |
| `REQUIRED` | 20,782 | |
| `OPTIONAL` | 12,508 | |
| `OUTPUT_ONLY` | 10,242 | |

`resource_reference` sits at roughly 15%, and 0% across compute.

The coverage figures matter less than what *absence* implies. For `field_behavior`, absence means
"assume optional", which is already the behaviour — so a missing annotation costs nothing. For
`resource_reference`, absence licenses no conclusion at all, so the fallback is an active guess.
That asymmetry, not the percentages, is why the design derives field behaviour mechanically and
treats references as judgement.

## Phase 1: what `+required` changes

Measured over three full sweeps of all 131 `apis/*/generate.sh` — identical procedure, one variable,
the flag forced on for every service. That is not the intended use, since the flag is opt-in per
service, but running it everywhere is what surfaces the failure modes:

| sweep | `+required` marker lines | `required:` under `spec:` | `required:` under `status:` |
|---|---:|---:|---:|
| flag off — the control | 1,321 | 10,233 | **15** |
| flag on, one shared struct | 2,020 | 10,488 | **35** |
| flag on, split on REQUIRED | 2,020 | 10,488 | **21** |

The flag adds 699 markers and 255 `required` lists under `spec:`, across 62 CRD files. (Marker
counts here are strict `// +required` lines; the 1,341 in the prior-art table below counts every
line containing the string.)

Every spec-side addition is nested, and a nested `required` is conditional. In JSON Schema a
`required` list inside an object applies only when that object is present, so `httpHeaders[].name`
means "if you supply a header it must have a name", not "every object must have a header". The case
to watch for is a `required` at the *top level* of `spec`, where it means "always".

That cannot arise here. `+required` is emitted by `WriteField` into `types.generated.go`, which
holds only nested types; the top-level `<Kind>Spec` lives in the hand-written `<kind>_types.go`,
which `generate-types` never overwrites. Verified across the whole tree: no generated `<Kind>Spec`
struct belongs to a CRD of the same service. Three apparent hits — `BigQueryRoutineSpec`,
`BigQueryTableSpec`, `ServiceSpec` — are datacatalog's own nested proto types, colliding by name
with unrelated Kinds in other groups.

The 20 status entries are the one case nesting does not cover: a type reused across contexts with
different requirements. Nested `required` expresses "optional parent, required child" correctly;
what it cannot express is "required when a user supplies this, not guaranteed when GCP returns it".
Nested message types are generated once by `WriteMessage` and shared between spec and observed
state, so a marker taken from a field's own annotation lands in every schema position that type
occupies. CRD structural validation covers the status subresource, so a GCP response missing such a
field makes KCC write a status the API server rejects, and reconciliation fails at runtime.

Splitting the struct removes 14 of those 20. Spec and marker counts are identical with and without
it, so the only difference is how the structs are arranged:

| removed by the split | lists |
|---|---:|
| `VertexAITuningJob` (v1alpha1) | 7 |
| `DiscoveryEngineSession` (v1alpha1) | 5 |
| `RedisCluster` (v1beta1) | 1 |
| `RedisCluster` (v1alpha1) | 1 |

`RedisCluster` v1beta1 is a published beta CRD, and its entry is `required: [network]` from
`PSCConfig` being shared between the spec and `DiscoveryEndpoint`'s observed state. That is the case
that started this.

### Prior art: required in status

The gate contains the behaviour rather than endorsing it. Required-in-status runs against what KCC
already does:

| | count |
|---|---:|
| `required` lists under `spec:` across 615 shipped CRDs | 10,233 |
| `required` lists under `status:` | 15 |
| `+required` markers in `apis/` | 1,341 |
| ...written inside an ObservedState struct | **2** |
| ObservedState structs in the tree | 971 |

The two deliberate ones are `CloudBuildWorkerPoolObservedState.WorkerConfig` and
`Task_ExecutionSpecObservedState.ServiceAccount`. A third grep hit, `DefaultSnatStatus`, is a proto
message name — `google.container.v1.DefaultSnatStatus` — not a KCC status type, so a naive search
returns 3.

`docs/develop-resources/api-conventions/validations.md` Rule 1 defines `+required` and every example
it gives is a Spec. It says nothing about status, so there is no written rule to cite — only 2
markers in 1,341.

The other 13 arrived the same way this phase would produce them, only by hand. `NodeTaint`
(`apis/container/v1beta1/containercluster_types.go:1266`) carries `+required` on `effect`, `key` and
`value`, and is used by both the spec-side node config and `NodePoolNodeConfigObservedState`
(`apis/container/v1beta1/containernodepool_types.go:424`), with no `NodeTaintObservedState` between
them. One struct, two schema positions, one set of markers — shipped in v1beta1.

The deliberate case is already a latent bug. `CloudBuildWorkerPoolObservedState_FromProto`
(`pkg/controller/direct/cloudbuild/workerpool_mappings.go:26`) returns a non-nil struct and fills
`WorkerConfig` only when `GetPrivatePoolV1Config()` is non-nil. The CRD declares
`required: [workerConfig]` under `status.observedState` and declares a status subresource, so a
WorkerPool returned without that config makes KCC write a status its own API server rejects. Nothing
about the `+required` line says so. Pre-existing, and not fixed by anything here.

Incomplete responses are common.
`pkg/controller/direct/videostitcher/videostitchercdnkey_controller.go:279` — "private keys and
token keys are write-only fields and not returned by GCP".
`pkg/controller/direct/documentai/documentaiprocessor_controller.go:209` — "the
SetDefaultProcessorVersion API does not return the updated Processor, so we need to read it again".
Status is also populated from KCC's own state rather than from the response:
`status.ExternalRef = direct.LazyPtr(a.id.String())`, across many direct controllers. A `required`
in status therefore asserts a guarantee about a third party's response body that KCC cannot enforce,
and whose violation lands on KCC's reconcile loop rather than on a user's apply.

### The root cause, and what it took to fix

`needsObservedState` (`dev/tools/controllerbuilder/pkg/codegen/typegenerator.go`) decides whether a
message gets its own `XObservedState` struct, and returned true only when the message recursively
contains an `OUTPUT_ONLY` field. Its comment states the premise: *"If the regular Go struct and the
ObservedState version are identical, we fall back to using the regular Go struct to reduce
redundancy."*

Emitting `+required` is what makes them stop being identical, so the premise no longer holds once
the flag is on. `PSCConfig` has no output-only field, deduplicates to one struct, and that one
struct carries the marker into both schema positions — even though `WriteObservedStateMessage`
already passes `WriteOptions{}` so variants never emit it.

The obvious repair is a second trigger: a message carrying a REQUIRED field also needs its own
ObservedState struct. On its own it breaks `generate-crds` in two opposite directions, because a
split needs two decisions taken in different places — how a nested field is *named*, from membership
in `observedStateMessages`, and whether the struct is *written*, from a filesystem scan, since the
generator never overwrites a hand-written type. When they disagree, a field points at a type nothing
defines:

- **The trigger alone dangles a reference.** aiplatform's `FunctionDeclaration` has a hand-written
  plain type claiming its `+kcc:proto` tag, so `WriteOutputMessages` skips writing the struct while
  the parent still points at `*FunctionDeclarationObservedState`, and controller-gen reports
  `unknown type FunctionDeclarationObservedState`.
- **Pruning the set to compensate breaks the other direction.** Six dataplex types —
  `DataQualityDimensionResult` and friends — have hand-written `XObservedState` structs, so
  `*XObservedState` does resolve and must stay. Pruning strips the suffix and points them at plain
  types that do not exist.

What made the two indistinguishable was the annotation parser. `findTypeDeclarationWithProtoTag`
asks `GetProtoMessageFromAnnotation` (`pkg/codegen/common.go`) whether a human already wrote
something for a message; that helper accepts all four annotation kinds and returns **only the
message name, discarding which one matched**. Both classes are large: 1,175 hand-written types carry
`+kcc:proto`, 453 carry `+kcc:observedstate:proto`.

**The fix is to scan for hand-written types first, and split only messages the generator fully
owns.** A kind-aware parser variant reports which annotation matched; one pass over a package's
non-generated files records every hand-written type name and every proto message claimed under any
annotation; the REQUIRED trigger fires only where that scan finds nothing.

Excluding those messages costs nothing. A message with a hand-written counterpart is skipped by
`WriteVisitedMessages` and never gets a generated struct at all — so it never receives a generated
marker and could not have leaked. The rule separates all three observed cases:

| message | hand-written? | outcome |
|---|---|---|
| redis `PSCConfig` | no — only in `types.generated.go` | split |
| aiplatform `FunctionDeclaration` | yes, plain `+kcc:proto` | excluded |
| dataplex `DataQualityDimensionResult` | yes, `+kcc:observedstate:proto` | excluded |

It also scopes the change to new resources without a second flag. Greenfield resources have no
hand-written types yet, so they are the set that gets split; existing resources are left alone as a
consequence of the rule rather than by opting out.

### What the fix does not close

Two limits survive, both reachable only by opting a service in.

**Six status entries remain, and they are a different defect.** In `ColabRuntime`,
`DialogflowConversationDataset`, `DiscoveryEngineIdentityMappingStore` and `SaasServiceMgmtRelease`
the generator now emits the correct `XObservedState`, and a **hand-written** ObservedState struct
still names the plain type. `generate-types` never rewrites hand-written files, so it cannot reach
this. The same structs show authors already use the variant wherever one existed when they wrote the
file:

```go
// colabruntime_types.go
EncryptionSpec     *EncryptionSpecObservedState   // correct — already split on OUTPUT_ONLY
IdleShutdownConfig *NotebookIdleShutdownConfig    // leaks — no variant existed then
```

The remedy is one field type per occurrence. Verified on `DiscoveryEngineIdentityMappingStore`:
changing `*CmekConfig` to `*CmekConfigObservedState` clears both of its entries, after which the
mapper regenerates itself and `generate.sh`, `generate-crds` and `go build ./apis/...` all pass.

**Splitting can orphan a plain type that hand-written code outside `apis/` still uses.** The split
propagates to children, since a parent must reference `ChildObservedState` in the observed context.
Where a message chain is reachable only through the observed-state tree, the plain versions lose
their last referent and `prunetypes` comments them out. aiplatform is the one case in 131:
`FunctionCall` and `FunctionResponse` get commented out while
`pkg/controller/direct/aiplatform/model_mapping.go` still names `krm.FunctionCall`, and the package
stops compiling.

`prunetypes` only considers references within `apis/<service>/<version>/`, so hand-written mappers
are invisible to it. That is the underlying gap, it is not addressed here, and it is filed upstream
as [#12465](https://github.com/GoogleCloudPlatform/k8s-config-connector/issues/12465). A type swap
is not the remedy either, since those mappers are bidirectional and an ObservedState type in a
`ToProto` is meaningless. aiplatform would need its mappers reconciled before opting in. Nothing
else in the tree is affected, and neither limit is reachable with the flag off.

### The assumption this overturned

Phase 1 was written **ungated**, on the assumption that it was the low-risk phase and could apply
globally. The status entries are what forced the flag, and they are invisible if you only look at
the field being annotated — the damage is done by where a type is *reused*, not by the field the
annotation sits on.

An earlier draft also argued that emitting `+required` is *good* because it moves the failure from a
GCP round-trip to `kubectl apply`. That is backwards — KCC will never know an API's rules as well as
the team that owns it, and `required` is backwards-incompatible — and inverting it produced the
"defer behaviour decisions to the API" principle. What the measurement establishes is only that
these additions agree with what the proto declares, which is a statement about consistency, not
desirability.

## The baseline is a control sweep, not the checked-in tree

Regenerating is not a no-op, so measuring a generator change against the committed CRDs attributes
pre-existing drift to the change. Six `apis/` files change when their own `generate.sh` is re-run
with no flag and no local edit — `bigtable`, `datastream`, `edgecontainer`, `orgpolicy`, `tpu`,
`vertexai` — and two of them break:

- **orgpolicy** — `generate.sh` generates only `OrgPolicyCustomConstraint`, while
  `v1beta1/types.generated.go` also carries `// resource: OrgPolicyPolicy:Policy` and defines
  `AlternatePolicySpec`, `PolicySpec_PolicyRule_StringValues` and `Expr`. Commit `71af28e15d`
  ("promote OrgPolicyPolicy to v1beta1") added those 50 lines without adding the resource to the
  script. Re-running deletes them, and `generate-crds` fails with `unknown type
  PolicySpec_PolicyRule_StringValues`. Reproduced on upstream `master` with a controllerbuilder
  built from the same commit, and filed as
  [#12463](https://github.com/GoogleCloudPlatform/k8s-config-connector/issues/12463).
- **tpu** — regenerating emits `pb "cloud.google.com/go/tpu/apiv2/tpupb"` into
  `pkg/controller/direct/tpu/mapper.generated.go`; that module is not in `go.mod`, so `go build
  ./...` fails. Filed as
  [#12464](https://github.com/GoogleCloudPlatform/k8s-config-connector/issues/12464).

`datastream` picks up newer proto field descriptions. None of this touches `required` counts: the
control sweep measures 10,233 / 15, matching the committed tree. That agreement is a coincidence,
though, not something the method guarantees. A generated file is only as current as the last person
who ran the script, and nothing checks that re-running it is a no-op.

Two smaller traps in the same family. `SKIP_GENERATE_CRDS` must be scoped per invocation rather than
exported, or the final `generate-crds` skips itself and exits 0 while appearing to pass. And after a
`SKIP_GENERATE_CRDS=1` run `go build ./apis/...` fails on stale deepcopy, because deepcopy is
regenerated by `generate-crds`; that is an artifact of the skip, not a defect.

## Phase 2: how wrong the guess is

Measured over the 1417 messages carrying `google.api.resource`, using `protoapi.GetResourceMetadata`
— the production path — rather than a separate reimplementation:

| Bucket | Count | Share |
|---|---:|---:|
| correct | 562 | 39.7% |
| wrong: casing only | 554 | 39.1% |
| wrong: pluralisation | 198 | 14.0% |
| not comparable (pattern declares no collection) | 103 | 7.3% |
| **wrong overall** | **752** | **53.1%** |

The two failure modes are worth keeping apart, because they are different arguments:

| Mode | Proto message | Template emits | API uses |
|---|---|---|---|
| casing | `LbTrafficExtension` | `lbtrafficextensions` | `lbTrafficExtensions` |
| casing | `ChannelGroup` | `channelgroups` | `channelGroups` |
| pluralisation | `Batch` | `batchs` | `batches` |
| pluralisation | `Property` | `propertys` | `properties` |

The `LbTrafficExtension` row is the pilot, so this was already being fixed by hand.

### The number was first reported as 60.1%

That came from a throwaway probe that normalised every `{placeholder}` to `*` and took the
second-to-last segment, so patterns ending in a literal — `projects/{project}/locations` — scored as
wrong when there is simply nothing to compare. Those are the 103 "not comparable" rows. Re-running
through the production path gives 53.1%.

The casing half of this has already shipped into nine `_identity.go` files; that is its own finding,
in [`identity-collection-casing.md`](identity-collection-casing.md).

## Rejected: matching `resource_reference` by field name

An attempt to close the reference-annotation gap heuristically, by matching `resource_reference`
targets against field names, produced **2,164** findings against **78** for description heuristics
alone. Abandoned. This is why references are treated as judgement rather than derivation: the
fallback is not just incomplete, it produces far more noise than signal.

## Rejected: annotation-driven queue

The queue was designed to take field-level entries from `google.api.resource_reference`.
Implementing it ruled that design out. `LbTrafficExtension` — the pilot — carries no
`resource_reference` on any field, including `forwarding_rules`, which is exactly the field that has
to become a ref:

```
google.cloud.networkservices.v1.LbTrafficExtension
  name                  resource_reference: -
  forwarding_rules      resource_reference: -   <-- must become a ref
  extension_chains      resource_reference: -
```

An annotation-only queue would have been empty for this resource, so no file would have been
written, no suppression would have happened, and it would have gone straight into the ratchet and
failed — the thing the queue exists to prevent. Hence a resource-level entry emitted
unconditionally, with field-level entries as a bonus rather than the mechanism.

## Verified: baseline entries must carry forward

A queued resource contributes no `[refs]` findings, so anything it already owed reads as *removed*
and gets pruned from the ratchet — then reappears as a new violation the moment it graduates.

Checked directly against `AlloyDBBackup`: queueing it without the carry-forward reported its two
existing entries as fixed; with the carry-forward they stay. Queueing therefore only ever stops
findings being *added*.

## Limit of what was checked

The scaffolder has no access to field data and the type generator does. Whether anyone has
previously tried joining the two and hit a problem not visible from the code was not investigated.
