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
2b. **Seed reference hints** into the queue, once the CRDs exist.
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
| 2b. Reference hints | `scripts/queue-ref-hints` seeds the queue | Yes |
| 3. Judgement pass | Refs, omissions, ObservedState, KRM renames | Yes, by hand |
| 4. Mappers and deepcopy | `generate-mapper`, then `controller-gen` | Yes |
| 5. Register and verify | Manifest entry, baselines, checks | Yes |
| 6. Ready for identity and refs | Hands off to the next step | No — see [Gaps](#gaps) |

## Before you start

The generator changes are on master. Four are opt-in per service, so a service that has not turned
them on yet still gets the old behaviour:

| Capability | How you get it |
|---|---|
| Spec **and ObservedState** filled from the proto message | `--prepopulate-spec` |
| `+required` from `field_behavior` | `--emit-required-from-proto` |
| Plural acronyms cased as KRM wants (`relatedURIs`) | `--emit-plural-acronyms` |
| Report output-only fields the proto states only in prose | `--detect-output-only-in-comments` |
| Real collection segment and parent shape | always on |
| `Location` only where the parent shape calls for it | always on |
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

## Stage 2b — Seed the queue with reference hints

Run after the CRDs exist, because the detector reads them:

```bash
go run ./scripts/queue-ref-hints
```

This is what makes the queue usable as a work list. Without it the queue names only the references
the proto annotates with `google.api.resource_reference`, which measured **11 of 111** on the
239-resource run — `CloudBuildConnection` needs 15 references and the queue listed none. The seeder
applies the same rules as `TestMissingRefs`, plus name rules for the classes those rules cannot see,
and reaches **82 of 111**.

Entries say how confident they are, and the difference matters when you work through them:

| reason | what it means |
|---|---|
| `possible-reference` | the proto's own `resource_reference` annotation — a fact |
| `possible-reference-by-description` | the description names a resource path — strong |
| `possible-reference-by-name (TargetRef)` | the field name matches a known target — a hint |

Roughly a third of hints are wrong, and that is the intended trade for a list a person confirms.
Most of the wrong ones are fields whose exact name upstream made a reference in a *different*
resource, so they are worth a moment's thought rather than a reflex dismissal.

## Stage 3 — The judgement pass

Three decisions cannot be derived from anything, and this stage is where they get made:

1. **Which strings are really references.**
2. **Which fields to leave out deliberately.**
3. **Which fields need renaming for KRM conventions.**

Two things are deliberately not on that list. Required-versus-optional is answered by
`--emit-required-from-proto` from the annotation, and only a considered contradiction of the proto
needs a person. ObservedState used to be here and is not any more: `--prepopulate-spec` fills it
from the proto's `OUTPUT_ONLY` fields, including the `*Foo` versus `*FooObservedState` choice for
nested messages. What the proto states only in prose is reported by
`--detect-output-only-in-comments` and is a hand edit; see the gotchas.

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

One generator gap will stop the package compiling:

- **`google.protobuf.Value` and `ListValue` have no mappers anywhere.** There is no fix at this
  stage; dropping the field is the only option, which makes it a stage 3 omission decision.

The import gaps that used to sit here are fixed. `common.Status` and `apiextensionsv1.JSON` now
bring their own imports into the scaffolded file, and the generated file no longer keeps an import
for a type the scaffolder took.

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

Five places the pipeline stops short. These are stated here, not solved.

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
- **`Location` follows the parent shape now, and the nested case needs a decision.** The scaffold
  emits it only when the proto's pattern makes the parent project+location. A resource nested under
  another resource — a Lake, a DataStore, a KeyRing — gets none, because its parent already fixes
  one and upstream is split 8 to 7 on repeating it; the queue records that as
  `location-omitted-nested-parent` for you to settle. A resource whose proto carries no
  `google.api.resource` gets none either, recorded as `location-omitted-unknown-parent`.
- **Never delete a resource whose invocation passes `--skip-scaffold-files`.** The generator does
  not write its types file, so deleting one removes something nothing will restore. Two resources
  were lost this way in the 239-resource run before anyone noticed.
- **A Kind whose name matches a proto message in the same package collides.** `APIHubInstance`,
  `Document` and `BillingAccount` end up declared by both the scaffold and `types.generated.go`.
- **Regenerating a `v1alpha1` package can break a `v1beta1` one.** `backupdr/v1beta1` references
  `v1alpha1.Parent`, so a change on the alpha side stops the beta package compiling.
- **Two `generate-types` invocations writing one `types.generated.go`** overwrite each other; the
  second wins. `networksecurity` does this.
- **`map<string, Message>` generates now; the JSON-ish value types still do not.** The value
  struct is written like any other nested message, and the field renders as
  `map[string]TheValueType` with the value form rather than a pointer, which the corpus prefers 16
  to 7. Three value types are still declined: `google.protobuf.Value` and `google.protobuf.ListValue`
  are mutually recursive, so a map of one makes it reachable from the CRD and `controller-gen` then
  fails on the whole package rather than on the one field, and `google.protobuf.Struct` has a
  special-cased Go type with no map spelling. Together those are 100 of the 1002 map-of-message
  fields across the Google API protos; mapping all three to `apiextensionsv1.JSON`, which is what
  upstream writes by hand in `apis/firestore/v1alpha1` and `apis/aiplatform/v1alpha1/recursive_types.go`,
  would close them. A declined field never reaches the CRD; on a fresh scaffold it appears in the
  queue twice, as a `field ".spec.<name>" reason=unsupported-field-type` entry and as a per-service
  `# dropped:` comment. **Regenerating a resource whose `_types.go` already exists records only the
  second**, because the scaffolder skips a file that is already there and the field-level entry is
  written on that path. So a rerun looks quieter than a first run for the same defect.
- **When `generate-crds` panics without naming a package**, run `controller-gen` per service with
  `paths="./<svc>/v1alpha1"` from `apis/`. The tree-wide `paths="./..."` lets one unloadable package
  block every other, and a panic on an unresolvable type names nothing to attribute it to.
- **A bare `--resource Kind:Message` in a multi-service block picks the first match.** Where the
  `--service` flag lists several comma-separated services, an unqualified proto message resolves
  against whichever of them declares it first. `NotebookInstanceV2` took `notebooks.v1.Instance`
  that way instead of the v2 message and came out 39 CRD fields short, with nothing reporting a
  problem. Qualify the message with its full proto name whenever the block names more than one
  service. The generator now logs a warning on an ambiguous name; it is worth reading.
- **Editing a `generate.sh` puts flags inside the invocation, not after it.** Several services
  write theirs on one line with no `\` continuations, so a flag appended below becomes its own
  command. `tpu` and `edgecontainer` were corrupted this way: the original invocation ran first and
  wrote a three-field stub, and only then did the shell say `--prepopulate-spec: command not found`.
  `bash -n` does not help — an orphan flag line is valid syntax, just a command that does not
  exist — so check for the shape directly:

  ```sh
  awk 'prev !~ /\\$/ && /^[[:space:]]*--/ {print FILENAME":"FNR": "$0} {prev=$0}' apis/*/generate.sh
  ```

  Any output is a flag line whose predecessor did not end in a continuation. Silence means clean.
  Then check the resulting Spec has more than three fields, since the stub is already on disk by
  the time the script fails.
- **A comment above a Spec field becomes that field's CRD `description`.** `controller-gen` takes
  the whole contiguous `//` block, so a note meant for whoever reads the generator ends up in the
  published schema, where it reads as nonsense. The Location rationale leaked into eight CRDs this
  way before anyone noticed. Rationale belongs on the template in
  `dev/tools/controllerbuilder/template/apis/types.go`, not inside the backtick-quoted template
  string. `grep` your new wording in `config/crds/resources/` to check.
- **The scaffold templates hardcode `Copyright 2025`.** `template/apis/{doc,groupversion_info,identity,refs}.go`
  all carry that year, so every newly scaffolded file gets it, while CLAUDE.md asks for the current
  year on new files. Fix the header by hand, or fix the templates once.
- **`bin/controllerbuilder` is reused if present**, at any age, by every `apis/*/generate.sh`. If
  you are changing the generator itself, rebuild it or delete it — a stale binary fails silently and
  the symptom shows up in a service you never touched.
