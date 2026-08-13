# Greenfield generator mechanics: mechanising the manual half

**Status:** design, not yet implemented. **Scope:** the experimental sandbox (`kcc-lab`), not
upstream policy.

Companion to [`greenfield-bulk-generation.md`](greenfield-bulk-generation.md). That document is the
*procedure* an agent follows per resource, given the generator as it exists today. This one is about
changing the generator so the procedure has less to do.

Concretely, it targets the "What you must fix by hand" table in that doc. Four of its seven rows are
mechanical and should not be a human's job; one — reference fields — is genuine judgement and should
be tracked explicitly rather than left implicit.

---

## 1. Why the top-level file is manual

`types.generated.go` is produced correctly, including a complete struct for the top-level message.
`<kind>_types.go` is a near-empty scaffold stamped `Please EDIT it!`. That asymmetry looks arbitrary,
and it is worth understanding why it exists, because it explains what is cheap to fix.

Two separate pieces of code write Go files during generation, and they share no state.

**The type generator** (`pkg/codegen/typegenerator.go`) opens the proto message, walks every field,
resolves each to a Go type, and splits Spec from ObservedState by checking
`IsFieldBehavior(f, annotations.FieldBehavior_OUTPUT_ONLY)` — three call sites, at lines 130, 168 and
363. It has the full message descriptor.

**The scaffolder** (`scaffold/apis.go`) writes `<kind>_types.go`. It cannot see the proto message at
all. Its struct holds five strings:

```go
type APIScaffolder struct {
	BaseDir, GoPackage, Group, Version, PackageProtoTag string
}
```

and `buildAPIArgs` passes only names onward — `Kind`, `KindProtoTag`, `ProtoResource`,
`ProtoMessageName`, `ProtoMessageFullName`. No descriptor reaches that path, so the template could
not enumerate fields even if it wanted to.

The identity template shows the same limitation more visibly, because it has to guess at facts it
cannot look up:

```go
// template/apis/identity.go
return "projects/" + p.ProjectID + "/locations/" + p.Location          // assumes project+location
return i.parent.String() + "/{{.ProtoMessageName | ToLower}}s/" + i.id // plural = name + "s"
```

Both facts are stated in the proto's `google.api.resource` annotation, as `pattern` and `plural`.
Neither reaches the scaffolder. `Please EDIT it!` is the template admitting it guessed.

So: the field data exists, and the file that needs it is written by a pass that cannot see it. The
type generator emits the full top-level struct, `prunetypes` comments it out because nothing
references it, and a human copies fields from the commented block into a scaffold written without any
knowledge of them.

That transcription is where the pilot's defects came from — a `Location string` in the scaffold, a
2025 copyright, a missing `stability-level` label. None of them were judgement errors.

**Limit of what was checked:** the scaffolder demonstrably has no access to field data and the type
generator does. Whether anyone has previously tried joining the two and hit a problem not visible
from the code was not investigated.

## 2. How far proto annotations can be trusted

Measured across googleapis, excluding the ads services:

| | count | share |
|---|---:|---:|
| Total proto fields | 109,515 | |
| Carrying any `field_behavior` | 44,577 | 40.7% |
| `REQUIRED` | 20,782 | |
| `OPTIONAL` | 12,508 | |
| `OUTPUT_ONLY` | 10,242 | |

Compare `resource_reference` at roughly 15%, and 0% across compute.

The coverage numbers matter less than **what absence means** in each case:

| Annotation | Absent means | Fallback |
|---|---|---|
| `field_behavior` | assume optional | inert — that is already the behaviour |
| `resource_reference` | nothing can be concluded | must actively guess |

This is the whole reason the two are treated differently below. Proto annotations are canonical where
present, with existing parsing as fallback — but that rule buys far more for field behaviour, where
being wrong costs nothing, than for references, where the fallback *is* the problem. An earlier
attempt to match `resource_reference` by field name produced 2,164 findings against 78 for
description heuristics alone, and was abandoned.

## 3. Derivable versus judgement

| Element | Source | Status today |
|---|---|---|
| Field list, Go types, pointer rules | typegenerator | computed |
| Spec vs ObservedState split | `OUTPUT_ONLY` | computed |
| `+required` | `field_behavior` | **parsed, then unused** |
| Parent shape (project / location / org / folder) | `google.api.resource.pattern` | **available, unused** |
| Collection segment and plural | `google.api.resource.plural` | **available, unused** |
| Which strings are references | nowhere — ~15% annotated | **judgement** |
| Deliberate omissions (e.g. `Labels`) | nowhere | **judgement** |
| Field renames for KRM conventions | nowhere | **judgement** |

`FieldBehavior_REQUIRED` is parsed at `pkg/protoapi/overlay.go:281-282` and appears **nowhere** in
`pkg/codegen`. `IsFieldBehavior` is already generic over the behaviour, so emitting `// +required` is
a small change against machinery that exists.

Roughly: the Spec is ~90% mechanically derivable. The residue is reference decisions and deliberate
omissions — exactly the judgement Step 1 is supposed to concentrate on.

## 4. The design

Four phases. Phases 1–2 are independently useful; phase 4 depends on 3.

### Phase 1 — emit `+required` from `field_behavior`

`+required` is what produces the CRD's `required:` list. Without it, every field carrying
`json:",omitempty"` — which the generator emits for everything — is optional, so a generated resource
has an entirely permissive spec and nothing fails.

`+optional` is deliberately **not** emitted: `omitempty` already implies optional to controller-gen.
The gap is `+required` alone.

*Risk:* resources whose proto marks a field REQUIRED that KCC intentionally treats as optional. Any
change to an existing CRD's `required:` list is such a case and needs a decision before merging.

### Phase 2 — feed `google.api.resource` into the scaffolder

Thread `pattern` and `plural` into `APIArgs` so the identity template stops hardcoding
project+location and naive pluralisation. Prerequisite for phase 4 being correct on org-parented,
folder-parented and irregularly-pluralised resources.

### Phase 3 — the judgement queue

Add `apis/<service>/needs_judgement_call.txt`, **per service** so parallel generation of different
services never conflicts, and a compile script can glob `apis/*/needs_judgement_call.txt` to pick the
next resource to refine.

### Phase 4 — pre-populate the Spec

Emit the top-level Spec from the generated type: all fields minus identity fields, plus the parent ref
derived from the phase-2 pattern, with `+required` from phase 1. Fields the generator cannot decide
are still emitted as their raw proto-derived type **and** recorded in the queue, so the file compiles
and the open questions are explicit.

Queue entries cover only what nothing can derive: reference candidates, deliberate omissions, and KRM
renames. Required/optional divergence is *not* a queue entry — with phase 1 the annotation answers it
mechanically, and only deliberate contradiction of the proto needs a human.

**Gate phase 4 behind a flag** on `generate-types` (alongside the existing `--skip-scaffold-files`),
set per service in `apis/<service>/generate.sh`. Not the bulk manifest: the documented procedure
appends the Kind to `greenfield_bulk.txt` *after* generation runs, so a generator gated on it would
never fire for the resource being generated — and that file is test data consumed by checks, so
reading it from the generator inverts the dependency.

## 5. Where a decision gets recorded

`TestMissingRefs` already maintains three files, and they are not interchangeable. The distinction
that matters is **derived versus curated**:

| File | Kind | Can hold a reason? |
|---|---|---|
| `missingrefs.txt` | computed output, **ratchet** | No — recomputed each run |
| `refs_deferred.txt` | **curated input**, loaded and subtracted | Yes — its entire purpose |
| `refs_not_representable.txt` | computed output, golden | Reason set by the classifier |

`missingrefs.txt` cannot be a ledger. `TestMissingRefs` recomputes it from the CRDs and writes it
through `CompareRatchetFile` (`tests/apichecks/crds_test.go:192`), so anything a human writes into it
is erased on the next `WRITE_GOLDEN_OUTPUT=1` run.

`refs_deferred.txt` is the ledger. It is hand-edited, carries a stated reason per entry, and is
subtracted from the findings before the ratchet applies (`crds_test.go:178-186`). Its header already
describes the case exactly: correctly detected, not implementable now, target has no KCC resource or
Ref type yet.

Adding the queue gives four states, and a resource is in exactly one:

| State | File | Written by |
|---|---|---|
| not yet triaged | `apis/<service>/needs_judgement_call.txt` | generator |
| triaged: deferred, with reason | `refs_deferred.txt` | human |
| triaged: still owed, actionable | `missingrefs.txt` | computed ratchet |
| structurally impossible | `refs_not_representable.txt` | computed golden |

### Suppression is load-bearing, not bookkeeping

`missingrefs.txt` is a **ratchet**: new entries fail the check. A mechanically generated resource has
ref-shaped string fields, which produce exactly those new entries. Without suppression, every
mechanical-pass PR is blocked on day one.

So: while a resource has open entries in its service's `needs_judgement_call.txt`, its `[refs]`
findings are suppressed. Clearing the queue graduates the resource and the normal ratchet takes over.

Suppression is scoped to `[refs]` only. Every other greenfield check — proto annotations,
`observedGeneration`, copyright, labels, CRD conformance — still applies, so mechanical defects
surface immediately rather than at the judgement pass.

`loadDeferredRefs` + `refEntryKey` + `.Has()` (`crds_test.go:254-282`) is already a
load-key-subtract pipeline over this exact entry format. Suppression is a second call to the same
pattern, not new machinery.

## 6. Two caveats

**This changes shared tooling.** `dev/tools/controllerbuilder` is used by brownfield and TF-migration
work too. Everything the sandbox has done so far was additive or scoped by the bulk manifest; the
generator affects every resource anyone generates. Experiment here, but this belongs upstream on a
real discussion rather than as a permanent sandbox divergence. The phase-4 flag exists partly for
that reason: the new behaviour stays off by default and opt-in per service.

**"Merge the mechanical pass without worrying about correctness" has a scope.** It is safe in the
sandbox, where there are no users and CRD shapes are free to change. It is *not* safe at port-back: a
field shipped as a plain string that should have been a reference is a breaking CRD change once it
reaches a published beta, and the field and JSON names are baked into the schema. The judgement pass
is a prerequisite for upstreaming any resource generated this way, not indefinitely deferrable work.
