# Bulk greenfield generation: Step 1 (types + CRD)

**Scope:** the experimental sandbox (`kcc-lab`). Verified end-to-end on
`NetworkServicesLBTrafficExtension`.

Step 1 produces `<kind>_types.go` and the CRD. It does **not** produce `<kind>_identity.go`,
`<kind>_reference.go`, controllers, mappers, MockGCP or fixtures — those are separate steps.

---

## The one thing to understand first

`generate-types` does **not** write your Spec. It does two things:

1. Writes the complete struct — every proto field, correct Go types — into
   `apis/<service>/<version>/types.generated.go`, **wrapped in `/* unreachable type ... */`**
   because nothing references it yet.
2. Writes a near-empty scaffold `<kind>_types.go` with only `ProjectRef`, `Location` and
   `ResourceID`, and prints `Please EDIT it!`

**Your job is the assembly**: move the fields from the unreachable block into the scaffold's Spec and
ObservedState. When the Spec references a generated type, `prunetypes` un-comments it on the next run
automatically — the `unreachable` marker disappears on its own. Do not hand-edit
`types.generated.go`.

This is the per-resource manual cost. Everything else is mechanical.

## Procedure

**1. Add the resource to the service's `generate.sh`.** 129 services already have one. Append one
line to the `v1alpha1` block:

```bash
${CONTROLLERBUILDER} generate-types \
    --service google.cloud.networkservices.v1 \
    --api-version "networkservices.cnrm.cloud.google.com/v1alpha1" \
    --resource NetworkServicesLBRouteExtension:LbRouteExtension \
    --resource NetworkServicesLBTrafficExtension:LbTrafficExtension \   # <- added
```

`--resource` is repeatable, so a whole batch of resources for one service generates in one command.
Kind naming follows the service's existing convention — note `LBTrafficExtension`, not
`LbTrafficExtension`, matching the sibling.

If the service only has a `v1beta1` block, add a separate `v1alpha1` block; greenfield resources are
always `v1alpha1`.

**2. Run it.** `./apis/<service>/generate.sh`

**3. Assemble the Spec.** Copy fields from the `/* unreachable type <Proto> */` block in
`types.generated.go` into `<kind>_types.go`, and the `<Proto>ObservedState` fields into the
ObservedState struct. Follow an existing sibling resource in the same service if there is one — it is
the most reliable template.

**4. Re-run `generate.sh`**, then `go build ./apis/...`.

**5. Add the Kind to the bulk manifest.** Append `<group>/<Kind>` to
`tests/apichecks/testdata/greenfield_bulk.txt`. This is what puts the resource in scope for the
greenfield conformance checks — resources not listed are not checked, because they predate this bar.

The Kind alone is enough to find every file for the resource: `TestDirectResourceFileNaming` already
requires each file under `apis/` and `pkg/controller/direct/` to be prefixed with the lowercased
Kind.

**6. Regenerate baselines.** `WRITE_GOLDEN_OUTPUT=1 go test ./tests/apichecks/...`, then re-run
without the flag to confirm clean.

## What the generator gets right on its own

Verified on the pilot — no hand-fixing needed for any of these:

- `v1alpha1`, correct file path
- Scalars as pointers (`*string`); collections **not** pointers (`map[string]string`, `[]string`)
- Enums as `*string` (not custom wrapped types)
- `status.observedGeneration` as `*int64`
- `+kcc:spec:proto=`, `+kcc:observedstate:proto=`, `+kcc:proto:field=` annotations
- `OUTPUT_ONLY` proto fields split into a separate `<Proto>ObservedState` type
- CRD labels `managed-by-kcc=true` and `system=true`
- Every proto field present in `types.generated.go`

## What you must fix by hand

| Fix | Detail |
|---|---|
| **Assembly** | The main cost — see above. |
| **Copyright year** | Generator emits **2025**. Must be **2026**. |
| **Parent** | Scaffold emits `ProjectRef` + `Location` separately; existing resources use `*parent.ProjectAndLocationRef` inline. |
| **Reference fields** | Generator emits plain strings. See below. |
| **`+required` markers** | Generator emits none. Add for proto `REQUIRED` fields. |
| **`+kubebuilder:validation:Enum`** | Generator emits none, though the Go type is correct. |
| **`google.protobuf.Struct`** | Generator emits `apiextensionsv1.JSON`; existing resources use `*apiextensionsv1.JSON`. |

### Reference fields are the one that matters

The generator does **not** convert reference-shaped fields. On the pilot it emitted:

```go
ForwardingRules []string `json:"forwardingRules,omitempty"`
```

The correct form, matching the hand-reviewed sibling:

```go
// +kcc:proto:field=google.cloud.networkservices.v1.LbTrafficExtension.forwarding_rules
ForwardingRuleRefs []*computev1beta1.ForwardingRuleRef `json:"forwardingRuleRefs,omitempty"`
```

Note the field **and JSON name change** (`forwardingRules` → `forwardingRuleRefs`). Getting this
wrong is the expensive mistake, because it is baked into the CRD schema.

The cost is confined to `_types.go`. Once the type is right, the **mapper generator handles the ref
by itself** — on the pilot it emitted, with no hand-editing:

```go
out.ForwardingRuleRefs = append(out.ForwardingRuleRefs, &krmcomputev1beta1.ForwardingRuleRef{External: v[i]})
```

Check the proto for `(google.api.resource_reference)` first — it names the target type exactly and is
authoritative where present, but covers only ~15% of string fields overall and 0% in compute. Where
it is absent, use field name plus a corroborating description. **Do not add entries to
`missingrefs.txt`** — implement the reference instead.

Creating a *new* `<kind>_reference.go` / `<kind>_identity.go` for the resource itself is a separate
step and is not part of Step 1.

## Definition of done

- `go build ./apis/...` clean.
- Re-running `generate.sh` produces no further diff.
- No `unreachable type <YourProto>` remains in `types.generated.go`.
- CRD spec contains every proto field; `OUTPUT_ONLY` fields appear under `status.observedState`.
- `go test ./tests/apichecks/...` passes.

Expect **`alpha-missingfields.txt` to grow** — it records fields not exercised by test fixtures, and
Step 1 has no fixtures. On the pilot it grew by 17 lines. That is correct at this stage; the entries
are attributed by `crd=` and are removed when fixtures arrive in a later step.

## Checking your work

Resources in the bulk manifest are held to the generation bar by
`tests/apichecks/greenfield_test.go`. Three of its checks run in normal CI: the manifest resolves to
real CRDs and files, the per-resource Go files conform (2026 header, pointer rules, no
`refs.NormalizeWithFallback`), and the CRD is `v1alpha1` only.

### Dropped fields

A proto field with no KRM representation at all is invisible to every other check: it is not in the
CRD, so `missingfields.txt` cannot see it. It can be commented out, or simply never written, and
nothing notices.

`TestGreenfieldDroppedFields` catches these. It reads the `// MISSING: <Field>` markers the mapper
generator emits while walking proto fields, so the source of truth is the proto itself. A field is
only counted when it is missing from **both** the Spec and ObservedState mappers — those map the same
proto message, so each otherwise reports the other's fields.

Its baseline, `testdata/exceptions/greenfield_dropped_fields.txt`, is a **ratchet**: new drops fail
even under `WRITE_GOLDEN_OUTPUT=1` and are never written automatically; fixed drops are pruned. The
list can only shrink. Every entry needs a `reason=` — either implement the field, or say why it is
intentionally absent.

### Field coverage

The fourth check is **local-only and opt-in**, because Step 1 legitimately fails it until fixtures
exist:

```bash
GREENFIELD_STRICT=1 go test ./tests/apichecks/ -run TestGreenfieldBulkFieldCoverage
```

It lists every field of your resource that no fixture exercises — the worklist for the fixture step.
A field that genuinely cannot be covered goes in
`tests/apichecks/testdata/exceptions/greenfield_fields_accepted.txt` with a mandatory `reason=`.
That file is for "cannot be covered", not "not done yet"; entries without a reason fail the check.

## Gotchas

- **Worktrees have no `.build/`.** It is gitignored (~2.9 GB). Symlink it from a full checkout, or
  `generate-proto.sh` rebuilds every proto descriptor:
  `ln -sfn /path/to/main/checkout/.build .build`
- **`WRITE_GOLDEN_OUTPUT=1` picks up unrelated drift.** The pilot run also rewrote
  `multi_version_crd_diff/IAPSettings.diff`, which had nothing to do with the change. Check
  `git diff` and revert anything not attributable to your resource.
- **A resource may look missing when it is not.** Proto→CRD matching is case-sensitive: proto
  `LbRouteExtension` vs KCC kind `NetworkServicesLBRouteExtension` does not match on a naive
  comparison. Grep `apis/` and `config/crds/resources/` before generating.
- **`SupportsIAM` warns** for a types-only resource (`not recognized as a direct kind`). Expected
  until a controller exists.
