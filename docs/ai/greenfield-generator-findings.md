# Greenfield generator: what was measured, and what was abandoned

Evidence behind the four phases in
[`greenfield-generator-mechanics.md`](greenfield-generator-mechanics.md). Nothing here is needed to
understand what the generator does; it is here because several of these results overturned the
assumption the design started with, and each is a trap worth knowing about before repeating it.

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

Regenerating all 116 services with the marker enabled:

| | |
|---|---:|
| New `+required` markers | 638 |
| CRDs changed | 47 (45 alpha, `redisclusters` beta, `runjobs` stable) |
| `required:` added under `spec:` | 223 |
| `required:` added under `status:` | **18** |

**All 241 additions are nested, and nested `required` is conditional by construction.** In JSON
Schema a `required` list inside an object applies only when that object is present, so
`httpHeaders[].name` means "if you supply a header it must have a name", not "every object must have
a header". The dangerous class is a `required` at the *top level* of `spec`, where it means
"always".

That class is structurally empty. `+required` is emitted by `WriteField` into `types.generated.go`,
which holds only nested types; the top-level `<Kind>Spec` lives in the hand-written
`<kind>_types.go`, which `generate-types` never overwrites. Verified across the whole tree: no
generated `<Kind>Spec` struct belongs to a CRD of the same service. Three apparent hits —
`BigQueryRoutineSpec`, `BigQueryTableSpec`, `ServiceSpec` — are datacatalog's own nested proto
types, colliding by name with unrelated Kinds in other groups.

**The 18 status entries are the reason for the gate**, and they are the one case nesting does not
cover: a type reused across contexts with different requirements. Nested `required` expresses
"optional parent, required child" correctly; what it cannot express is "required when a user
supplies this, not guaranteed when GCP returns it". Nested message types are generated once by
`WriteMessage` and shared between spec and observed state, so a marker taken from a field's own
annotation lands in every schema position that type occupies. CRD structural validation covers the
status subresource, so a GCP response missing such a field makes KCC write a status the API server
rejects, and reconciliation fails at runtime.

Reproduced independently on a current tree. Baseline, with the flag off, is 10,233 `required` lists
under `spec:` and 15 under `status:` across 615 CRDs. Turning the flag on for three services alone:

| services regenerated with the flag | under `spec:` | under `status:` |
|---|---:|---:|
| none (baseline) | 10,233 | 15 |
| redis | 10,235 | **17** |
| dataplex + aiplatform + redis | 10,283 | **24** |

The redis additions are `RedisCluster` `required: [network]`, in **v1beta1** as well as v1alpha1 — a
published beta CRD, from `PSCConfig` being shared between the spec and `DiscoveryEndpoint`'s
observed state.

### Prior art: required in status

The gate is containment, not endorsement. Required-in-status runs against what KCC already does:

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

The other 13 arrived exactly the way this phase would industrialise. `NodeTaint`
(`apis/container/v1beta1/containercluster_types.go:1266`) carries `+required` on `effect`, `key` and
`value`, and is used by both the spec-side node config and `NodePoolNodeConfigObservedState`
(`apis/container/v1beta1/containernodepool_types.go:424`), with no `NodeTaintObservedState` between
them. One struct, two schema positions, one set of markers — shipped in v1beta1.

**The deliberate case is already a latent bug.** `CloudBuildWorkerPoolObservedState_FromProto`
(`pkg/controller/direct/cloudbuild/workerpool_mappings.go:26`) returns a non-nil struct and fills
`WorkerConfig` only when `GetPrivatePoolV1Config()` is non-nil. The CRD declares
`required: [workerConfig]` under `status.observedState` and declares a status subresource, so a
WorkerPool returned without that config makes KCC write a status its own API server rejects. Nothing
about the `+required` line says so. Pre-existing, and not fixed by anything here.

**Incomplete responses are normal, not exceptional.**
`pkg/controller/direct/videostitcher/videostitchercdnkey_controller.go:279` — "private keys and
token keys are write-only fields and not returned by GCP".
`pkg/controller/direct/documentai/documentaiprocessor_controller.go:209` — "the
SetDefaultProcessorVersion API does not return the updated Processor, so we need to read it again".
Status is also populated from KCC's own state rather than from the response:
`status.ExternalRef = direct.LazyPtr(a.id.String())`, across many direct controllers. A `required`
in status therefore asserts a guarantee about a third party's response body that KCC cannot enforce,
and whose violation lands on KCC's reconcile loop rather than on a user's apply.

### The root cause is known; the fix is still open

`needsObservedState` (`dev/tools/controllerbuilder/pkg/codegen/typegenerator.go`) decides whether a
message gets its own `XObservedState` struct, and returns true only when the message recursively
contains an `OUTPUT_ONLY` field. Its comment states the premise: *"If the regular Go struct and the
ObservedState version are identical, we fall back to using the regular Go struct to reduce
redundancy."*

Emitting `+required` is exactly what makes them stop being identical, so the premise no longer holds
once the flag is on. `PSCConfig` has no output-only field, deduplicates to one struct, and that one
struct carries the marker into both schema positions — even though `WriteObservedStateMessage`
already passes `WriteOptions{}` so variants never emit it.

The obvious repair is to add a second trigger: a message carrying a REQUIRED field also needs its
own ObservedState struct. **That was tried and reverted.** It works for the case it targets — redis
goes 17 → 15 with every spec-side addition intact, at a cost of one extra generated struct — but it
breaks `generate-crds` in two opposite ways:

- **Adding the trigger alone dangles a reference.** Membership in `observedStateMessages` is what
  makes a nested field render as `*XObservedState`, but `WriteOutputMessages` separately *skips*
  generating that struct when a hand-written type already claims the message's `+kcc:proto` tag.
  aiplatform's `FunctionDeclaration` is that case: the struct is written only as a comment while the
  parent still points at it, and controller-gen reports
  `unknown type FunctionDeclarationObservedState`.
- **Pruning the set to compensate breaks the other direction.** Six dataplex types —
  `DataQualityDimensionResult` and friends — have *hand-written* `XObservedState` structs, so
  `*XObservedState` resolves and must stay. With a pruning pass their references lose the suffix and
  point at plain types that do not exist.

The invariant to respect: **a message belongs in `observedStateMessages` only when something will
define `XObservedState` — whether the generator writes it, or a hand-written type already does.**
Those two cases have to be told apart, and neither the name lookup nor the proto-tag lookup does it
alone: dataplex's hand-written `DataQualityDimensionResultObservedState` is found by *both*, because
it carries `+kcc:observedstate:proto` for the same message.

Until that is settled the flag is the containment, and no service should opt in without checking its
own status schemas.

### The assumption this overturned

Phase 1 was written **ungated**, on the assumption that it was the low-risk phase and could apply
globally. The 18 status entries are what forced the flag, and they are invisible if you only look at
the field being annotated.

An earlier draft also argued that emitting `+required` is *good* because it moves the failure from a
GCP round-trip to `kubectl apply`. That is backwards — KCC will never know an API's rules as well as
the team that owns it, and `required` is backwards-incompatible — and inverting it produced the
"defer behaviour decisions to the API" principle. What the measurement establishes is only that
these additions agree with what the proto declares, which is a statement about consistency, not
desirability.

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
alone. Abandoned. This is the concrete reason references are treated as judgement rather than
derivation: the fallback is not merely incomplete, it is noisy enough to be worse than nothing.

## Rejected: annotation-driven queue

The queue was designed to take field-level entries from `google.api.resource_reference`.
Implementing it killed that design. `LbTrafficExtension` — the pilot — carries no
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
failed — the precise thing the queue exists to prevent. Hence a resource-level entry emitted
unconditionally, with field-level entries as a bonus rather than the mechanism.

## Verified: baseline entries must carry forward

A queued resource contributes no `[refs]` findings, so anything it already owed reads as *removed*
and gets pruned from the ratchet — then reappears as a new violation the moment it graduates.

Checked directly against `AlloyDBBackup`: queueing it without the carry-forward reported its two
existing entries as fixed; with the carry-forward they stay. Queueing therefore only ever stops
findings being *added*.

## Limit of what was checked

The scaffolder demonstrably has no access to field data and the type generator does. Whether anyone
has previously tried joining the two and hit a problem not visible from the code was not
investigated.
