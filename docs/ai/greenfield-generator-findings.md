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
annotation lands in every schema position that type occupies — redis `PscConfig` / `PscConnection`
are the clearest case. CRD structural validation covers the status subresource, so a GCP response
missing such a field makes KCC write a status the API server rejects, and reconciliation fails at
runtime.

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
