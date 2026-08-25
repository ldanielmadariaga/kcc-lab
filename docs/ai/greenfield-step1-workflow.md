# KCC-LAB: Experimental Greenfield Step 1: the workflow for types and CRDs

This is the runbook. It gives the commands to take a GCP resource from nothing to a complete
`<kind>_types.go` and CRD, using the experimental generator flags, and says what each command
produces and where it lands.

Its neighbours answer different questions. `greenfield-generator-mechanics.md` is why the generator
works this way; `greenfield-generator-findings.md` is the evidence behind those choices;
`greenfield-coverage-strategy.md` is which resources to do and in what order. This one is what to
run.

Scope is types, mappers and the CRD. Identity files, reference files, controllers, MockGCP and
test fixtures are later steps and are not produced here. This covers the experimental sandbox
(`kcc-lab`), not upstream policy.

## The pipeline

1. **Generate** the types and CRD mechanically, filling the Spec from the proto.
2. **Inventory** what the generator could not decide, from the queue it writes.
3. **Judgement pass** where an agent or a human resolves those, and clears the queue.
4. **Mappers and deepcopy**, regenerated against the settled types.
5. **Register and verify** the resource against the conformance checks.
6. **Ready for identity and refs** — the handoff to the next step.

Stage 2 is the part the pipeline gained, and it is what makes a mechanical first pass mergeable at
all. A generated resource has reference-shaped string fields by construction, and `missingrefs.txt`
is a ratchet that fails on any new entry, so without a queue to suppress those findings the first
bulk-generation PR could not land.

| Stage | What it does | Runnable today? |
|---|---|---|
| 1. Generate | Types, CRD, Spec filled from the proto | Yes |
| 2. Inventory | Reads `needs_judgement_call.txt` | Yes |
| 3. Judgement pass | Refs, omissions, ObservedState, KRM renames | Yes, by hand |
| 4. Mappers and deepcopy | `generate-mapper`, then `controller-gen` | Yes |
| 5. Register and verify | Manifest entry, baselines, checks | Yes |
| 6. Ready for identity and refs | Hands off to the next step | No — see [Gaps](#gaps) |

## Before you start

All four generator changes are on master. Two are opt-in per service, so a service that has not
turned them on yet still gets the old behaviour:

| Capability | How you get it |
|---|---|
| Spec filled from the proto message | `--prepopulate-spec` |
| `+required` from `field_behavior` | `--emit-required-from-proto` |
| Real collection segment and parent shape | always on |
| `[refs]` suppression while a resource is queued | always on, reads the queue |

The two flags are opt-in because turning them on for a resource people already use can change its
CRD schema — nested types are shared between spec and status, so a `+required` marker can reach
places you did not intend. For a new greenfield resource neither risk applies, so turn both on.

Two things to check before picking a resource, both of which waste a day if you skip them:

- **Skip any kind whose CRD already exists** in `config/crds/resources/`. This is the reliable
  duplicate-work guard, because the tracker goes stale.
- **Skip anything `RESOURCE_STATUS.md` lists as `OPEN` or `PLANNED`.** The team owns those upstream.

You also need `.build/` populated, which is gitignored and around 2.9 GB. In a worktree, symlink it
from a full checkout rather than letting `generate-proto.sh` rebuild every descriptor:
`ln -sfn /path/to/main/checkout/.build .build`

---

## Stage 1 — Generate

Add the resource to its service's `generate.sh`. 131 services already have one; you are appending a
line to the `v1alpha1` block, and adding the two flags if the service has not opted in yet. Both
flags are per-invocation, so enabling them here enables them for every resource in that block.

```bash
# apis/networkservices/generate.sh, in the --- v1alpha1 --- block
${CONTROLLERBUILDER} generate-types \
    --service google.cloud.networkservices.v1 \
    --api-version "networkservices.cnrm.cloud.google.com/v1alpha1" \
    --prepopulate-spec \
    --emit-required-from-proto \
    --resource NetworkServicesLBRouteExtension:LbRouteExtension \
    --resource NetworkServicesLBTrafficExtension:LbTrafficExtension
```

`--resource` takes `Kind:ProtoMessage` and is repeatable, so a whole batch for one service generates
in a single command. Kind naming follows the service's existing convention — note
`LBTrafficExtension`, not `LbTrafficExtension`, matching its sibling. If the service only has a
`v1beta1` block, add a separate `v1alpha1` one; greenfield resources are always `v1alpha1`.

Then run it:

```bash
./apis/networkservices/generate.sh
```

**Outputs**

| Path | Contents |
|---|---|
| `apis/<service>/<version>/<kind>_types.go` | Spec filled from the proto, with `+required` markers and `+kcc:proto:field=` annotations |
| `apis/<service>/<version>/types.generated.go` | Every proto message as a complete Go struct |
| `apis/<service>/needs_judgement_call.txt` | What the generator could not decide (stage 2) |
| `pkg/controller/direct/<service>/mapper.generated.go` | Proto↔KRM conversion, with `// MISSING:` for what it could not map (stage 4) |
| `apis/<service>/<version>/zz_generated.deepcopy.go` | `DeepCopy` for every type, from `controller-gen` (stage 4) |
| `config/crds/resources/*.yaml` | The CRD |

One behaviour surprises everyone the first time. `prunetypes` comments the generated struct out as
an `unreachable type` because nothing references it yet. As soon as your Spec references it, the
next run un-comments it automatically. **Never hand-edit `types.generated.go`** — your edits are
regenerated away, and the file is not where the fix belongs.

## Stage 2 — Inventory what needs judgement

The generator writes what it could not decide to `apis/<service>/needs_judgement_call.txt`. The file
is per service rather than global so that generating two services in parallel never produces a
conflicting diff in the same file.

```
kind=NetworkServicesLBTrafficExtension group=networkservices.cnrm.cloud.google.com: resource reason=untriaged-bulk-generation (spec was generated mechanically; confirm refs, omissions and KRM names)
kind=NetworkServicesLBTrafficExtension group=networkservices.cnrm.cloud.google.com: field ".spec.forwardingRules" reason=possible-reference (target=compute.googleapis.com/ForwardingRule)
```

Two kinds of entry. The **resource-level** one is always emitted, whether or not anything was
detected, and it is the entry that actually drives suppression. The **field-level** ones come from
`google.api.resource_reference` and are a bonus on top.

That distinction matters more than it looks. Measured on the pilot, `LbTrafficExtension` carries no
`resource_reference` on any field, including `forwarding_rules`, which is precisely the field that
has to become a ref. A queue built only from annotations would have been empty there, nothing would
have been suppressed, and the resource would have gone straight into the ratchet and failed.

While a resource has entries here, its `[refs]` findings are suppressed and it will not fail
`TestMissingRefs`. That is the only thing suppressed — every other check applies normally.

## Stage 3 — The judgement pass

Four decisions cannot be derived from anything, and this stage is where they get made:

1. **Which strings are really references.**
2. **Which fields to leave out deliberately.**
3. **Which fields need renaming for KRM conventions.**
4. **What belongs in ObservedState**, which the scaffolder leaves as an empty struct.

Required-versus-optional is deliberately not on that list: `--emit-required-from-proto` answers it
from the annotation, and only a considered contradiction of the proto needs a person.

References are the one that matters, because the mistake is expensive to undo — the field name is
baked into the CRD schema. Check `google.api.resource_reference` first, since it names the target
type exactly and is authoritative where present. It covers only about 15% of string fields overall
and none at all in compute, so where it is absent use the field name plus a corroborating
description.

Both the Go field and the JSON name change:

```go
// before, as generated
ForwardingRules []string `json:"forwardingRules,omitempty"`

// after
// +kcc:proto:field=google.cloud.networkservices.v1.LbTrafficExtension.forwarding_rules
ForwardingRuleRefs []*computev1beta1.ForwardingRuleRef `json:"forwardingRuleRefs,omitempty"`
```

The cost stops at `_types.go`. Once the type is right the mapper generator handles the ref by
itself, with no hand-editing.

Do not add entries to `missingrefs.txt` to make a finding go away — implement the reference, or
defer it explicitly in `refs_deferred.txt` with a reason.

ObservedState is the decision people skip, because an empty struct compiles and looks finished.
Twenty-three alpha resources currently ship with one. Only the resource-level
`<Kind>ObservedState` in `<kind>_types.go` needs filling; the nested `<Proto>ObservedState` structs
in `types.generated.go` are already generated for you.

Start from the `OUTPUT_ONLY` fields, which cover about 84% of what ends up there in practice. Two
things to get right:

- **Message-typed fields take the ObservedState variant.** If `types.generated.go` defines
  `FooObservedState`, the field is `*FooObservedState`, not `*Foo`. Nine fields in the tree get
  this wrong today, which is how the `CmekConfig` bug got in.
- **The proto often forgets to mark server-computed fields.** `create_time`, `update_time`,
  `delete_time`, `uid`, `etag`, `self_link` and `id` belong in ObservedState whether or not they
  carry the annotation. Where several identifier-shaped fields turn up together — compute resources
  routinely have `id`, `selfLink` and `selfLinkWithID` — deciding which one is the resource's
  identity is the real call, and it is the same one Step 2 has to make.

Some services give you nothing to work from. `apis/compute/v1alpha1/types.generated.go` contains a
single `Output only.` in the entire file, and at least ten other services contain none, because
their protos are derived from a discovery document that has no `field_behavior`. There, every field
is a judgement call.

**Clearing the queue entries graduates the resource.** Suppression stops, and anything it still owes
lands in `missingrefs.txt` as a normal finding. A resource is in exactly one state at a time, which
is what stops these files contradicting each other.

## Stage 4 — Mappers and deepcopy

Mappers follow automatically once the types are right, which is why this comes after the judgement
pass rather than before it. It is the same `generate.sh`, re-run; the stage exists because the
mapper that matters is the one generated against the settled Spec.

Most services only invoke `generate-mapper` for `v1beta1`, so a greenfield resource needs a
`v1alpha1` invocation added. `apis/aiplatform/generate.sh` is one of the few that already has one.

```bash
# apis/networkservices/generate.sh, after the generate-types block
${CONTROLLERBUILDER} generate-mapper \
    --service google.cloud.networkservices.v1 \
    --api-version "networkservices.cnrm.cloud.google.com/v1alpha1"
```

`zz_generated.deepcopy.go` needs no command of its own — `controller-gen` writes it inside
`dev/tasks/generate-crds`, which every `generate.sh` already calls at the end.

**Outputs**

| Path | Contents |
|---|---|
| `pkg/controller/direct/<service>/mapper.generated.go` | `FromProto`/`ToProto` for the Spec and ObservedState, plus `// MISSING:` for every field it could not map |
| `apis/<service>/<version>/zz_generated.deepcopy.go` | `DeepCopy` and `DeepCopyObject` for every type |

Skipping this stage does not just postpone work, it disables a check. `TestGreenfieldDroppedFields`
finds dropped fields by reading those `// MISSING:` markers, and
`greenfield.DroppedFields` returns `nil, nil` when the mapper file does not exist
(`tests/apichecks/greenfield/greenfield.go:177`). A resource with no mapper therefore *passes* the
dropped-fields ratchet while verifying nothing at all.

The same check is what makes stage 3's ObservedState work visible. A field counts as dropped only
when it is `// MISSING:` in **both** the Spec and the ObservedState mapper, because the two map the
same proto message and each reports the other's fields. So an unfilled ObservedState correctly
reports every output-only field as dropped, and a filled one clears them.

Two generator gaps will stop the package compiling, and both are worth recognising quickly:

- **`google.protobuf.Value` and `ListValue` have no mappers anywhere.** There is no fix at this
  stage; dropping the field is the only option, which makes it a stage 3 omission decision.
- **`[]common.Status` emits a mapper call without adding the `common` import.** Add
  `github.com/GoogleCloudPlatform/k8s-config-connector/apis/common` by hand.

Hand-written mapping is needed only where the generator emits `// MISSING:`. Everything else,
references included, it handles by itself.

## Stage 5 — Register and verify

Add the Kind to the bulk manifest. This is what puts the resource in scope for the greenfield
conformance checks; resources not listed are not checked, because they predate the bar.

```bash
echo "networkservices.cnrm.cloud.google.com/NetworkServicesLBTrafficExtension" \
  >> tests/apichecks/testdata/greenfield_bulk.txt
```

The Kind alone is enough to locate every file for the resource, because
`TestDirectResourceFileNaming` already requires files under `apis/` and `pkg/controller/direct/` to
be prefixed with the lowercased Kind.

Then regenerate the baselines and confirm a clean re-run:

```bash
WRITE_GOLDEN_OUTPUT=1 go test ./tests/apichecks/...
go test ./tests/apichecks/...
```

**Definition of done**

- `go build ./apis/...` is clean.
- Re-running `generate.sh` produces no further diff.
- No `unreachable type <YourProto>` remains in `types.generated.go`.
- The CRD spec contains every proto field, and `OUTPUT_ONLY` fields appear under
  `status.observedState`.
- `go test ./tests/apichecks/...` passes.

Expect `alpha-missingfields.txt` to grow, and leave it. It records fields no test fixture exercises,
and Step 1 has no fixtures by design. Entries are attributed by `crd=` and are removed once fixtures
arrive in a later step. On the pilot it grew by 17 lines.

One check is deliberately not a CI gate, for the same reason:

```bash
GREENFIELD_STRICT=1 go test ./tests/apichecks/ -run TestGreenfieldBulkFieldCoverage
```

It lists every field of your resource that no fixture covers, which is the worklist for the fixture
step. A field that genuinely cannot be covered goes in `greenfield_fields_accepted.txt` with a
mandatory reason. That file is for "cannot be covered", not "not done yet".

## Stage 6 — Ready for identity and refs

This stage has no output and no way to run it today. Nothing lists which resources have types and a
CRD but no identity file, and identity and reference files cannot be generated without also
scaffolding a controller. See [Gaps](#gaps) and [What comes next](#what-comes-next).

---

## Every output, and where it lives

| Artifact | Path | Written by | Read by | Kind |
|---|---|---|---|---|
| Resource types | `apis/<service>/<version>/<kind>_types.go` | `generate-types` | everything | generated, then hand-edited |
| All proto types | `apis/<service>/<version>/types.generated.go` | `generate-types` | the CRD generator | generated, never hand-edit |
| Mappers | `pkg/controller/direct/<service>/mapper.generated.go` | `generate-mapper` | `TestGreenfieldDroppedFields` | generated, then hand-edited |
| Deepcopy | `apis/<service>/<version>/zz_generated.deepcopy.go` | `controller-gen` | the compiler | generated, never hand-edit |
| CRD | `config/crds/resources/*.yaml` | `generate-crds` | the CRD checks | generated |
| Judgement queue | `apis/<service>/needs_judgement_call.txt` | `--prepopulate-spec` | `TestMissingRefs` | work queue |
| Bulk manifest | `tests/apichecks/testdata/greenfield_bulk.txt` | you | all `TestGreenfield*` | hand-maintained |
| Owed references | `testdata/exceptions/missingrefs.txt` | recomputed each run | `TestMissingRefs` | **ratchet** |
| Dropped fields | `testdata/exceptions/greenfield_dropped_fields.txt` | recomputed each run | `TestGreenfieldDroppedFields` | **ratchet** |
| Deprecated refs | `testdata/exceptions/deprecated_refs_v1beta1.txt` | recomputed each run | `TestGreenfieldNoNewDeprecatedRefs` | **ratchet** |
| Deferred references | `testdata/exceptions/refs_deferred.txt` | you, with a reason | `TestMissingRefs` | hand-maintained input |
| Unrepresentable refs | `testdata/exceptions/refs_not_representable.txt` | recomputed each run | `TestMissingRefs` | golden |
| Identity collection casing | `testdata/exceptions/identity_collection_casing.txt` | recomputed each run | `TestIdentityCollectionCasing` | **ratchet** |
| Uncovered alpha fields | `testdata/exceptions/alpha-missingfields.txt` | recomputed each run | `TestCRDFieldPresenceInTestsForAlpha` | golden |
| Accepted coverage gaps | `testdata/exceptions/greenfield_fields_accepted.txt` | you, with a reason | `TestGreenfieldBulkFieldCoverage` | hand-maintained input |

The **ratchet versus golden** distinction is the one most likely to trip you up, because it is
invisible from the filenames and both kinds live in `testdata/exceptions/`. A golden absorbs new
violations when you run with `WRITE_GOLDEN_OUTPUT=1`. A ratchet refuses them, and refuses them *even
with the flag set* — it can only shrink. Master currently has 17 goldens and 4 ratchets.

If a run fails and rerunning with `WRITE_GOLDEN_OUTPUT=1` does not fix it, you have hit a ratchet,
and the answer is to fix the finding rather than to record it.

## Gaps

Six places the pipeline stops short. These are stated here, not solved.

1. **No "ready for identity and refs" list.** Stage 5 produces nothing that tells you which
   resources are complete enough to hand on, so stage 6 has no input. It is a disk scan: manifest
   kinds that have `_types.go` and a CRD but no `_identity.go`.
2. **Identity and refs cannot be produced on their own.** They come from `generate-controller`,
   which also scaffolds and registers a full controller. Getting to stage 6 needs either a
   scaffold-only flag or a separate subcommand.
3. **No reverse manifest check.** Every conformance check runs manifest → files, and
   `TestGreenfieldBulkManifestIsResolvable` only validates that listed entries resolve. A resource
   that gets generated but never added to `greenfield_bulk.txt` is silently unchecked — which is
   exactly the absorption failure the strategy exists to prevent.
4. **No aggregate view of the queue.** The files are per-service by design, so a run across many
   services has no single report of what is outstanding to drive the judgement pass from.
5. **Nothing forces the queue to drain.** A resource can sit queued indefinitely with its `[refs]`
   findings suppressed the whole time, which makes suppression permanent in practice.
6. **ObservedState is not prepopulated.** The scaffolder emits an empty struct even though roughly
   84% of what belongs in it is derivable from `OUTPUT_ONLY`, and the generator already computes the
   plain-versus-`ObservedState` type choice that filling it needs. Designed but not built; see
   phase 5 in [the mechanics doc](greenfield-generator-mechanics.md).

## What comes next

Types, mappers and CRDs got generator support first, then a queue for what only a person can decide,
then checks scoped to the resources produced that way. Controllers, identity and reference files,
MockGCP and fixtures are each intended to get the same treatment, in that order. None of it is
designed yet, and the first prerequisite is gap 2 above.

## Gotchas

- **Worktrees have no `.build/`.** Symlink it from a full checkout, or `generate-proto.sh` rebuilds
  every proto descriptor.
- **`WRITE_GOLDEN_OUTPUT=1` picks up unrelated drift.** The pilot run also rewrote
  `multi_version_crd_diff/IAPSettings.diff`, which had nothing to do with the change. Read
  `git diff` and revert anything not attributable to your resource.
- **A resource may look missing when it is not.** Proto→CRD matching is case-sensitive: proto
  `LbRouteExtension` against kind `NetworkServicesLBRouteExtension` does not match on a naive
  comparison. Grep `apis/` and `config/crds/resources/` before generating.
- **`SupportsIAM` warns** for a types-only resource, saying it is `not recognized as a direct kind`.
  That is expected until a controller exists.
- **The scaffold emits `Location` as a plain `string`.** KCC types are pointers throughout; fix it
  to `*string` by hand in `<kind>_types.go`.
- **`bin/controllerbuilder` is reused if present**, at any age, by every `apis/*/generate.sh`. If
  you are changing the generator itself, rebuild it or delete it — a stale binary fails silently and
  the
  symptom shows up in a service you never touched.
