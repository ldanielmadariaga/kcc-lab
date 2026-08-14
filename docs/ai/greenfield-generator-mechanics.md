# Greenfield generator mechanics: mechanising the manual half

**Status:** all four phases implemented. **Scope:** the experimental sandbox (`kcc-lab`), not
upstream policy.

| Phase | Change | PR | Gated? |
|---|---|---|---|
| 1 | `+required` from `field_behavior` | #14 | `--emit-required-from-proto` |
| 2 | `google.api.resource` into the scaffolder | #15 | no — only writes new files |
| 3 | judgement queue + `[refs]` suppression | #16 | n/a |
| 4 | pre-populate the Spec | #17 | `--prepopulate-spec` |

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

Every phase that changes generated output is opt-in per service. That was not the original intent —
phase 1 looked safe enough to apply globally — but measuring it showed otherwise (below), and the same
reasoning applies to anything else that reshapes an existing CRD.

### Phase 1 — emit `+required` from `field_behavior`

`+required` is what produces the CRD's `required:` list. Without it, every field carrying
`json:",omitempty"` — which the generator emits for everything — is optional, so a generated resource
has an entirely permissive spec and nothing fails.

`+optional` is deliberately **not** emitted: `omitempty` already implies optional to controller-gen.
The gap is `+required` alone.

**Gated behind `--emit-required-from-proto`, default off.** This was written ungated first, on the
assumption that it was the low-risk phase, and then measured. Regenerating all 116 services:

| | |
|---|---:|
| New `+required` markers | 638 |
| CRDs changed | 47 (45 alpha, `redisclusters` beta, `runjobs` stable) |
| `required:` added under `spec:` | 223 |
| `required:` added under `status:` | **18** |

**Consistency with the proto is not the same as desirability.** An earlier draft argued that
tightening the CRD is good because it moves the failure from a GCP round-trip to `kubectl apply`.
That is wrong, and the opposite is closer to the truth: KCC will never know an API's rules as well as
the team that owns it, so deferring to the API is the safer default. `required` is also a
backwards-incompatible change, so absent certainty, lax presence validation is preferable. What the
measurement establishes is only that these additions agree with what the proto declares.

**All 241 additions are nested, and nested `required` is conditional by construction.** In JSON
Schema, a `required` list inside an object applies only when that object is present, so
`httpHeaders[].name` means "if you supply a header it must have a name" — not "every object must have
a header". The dangerous class is a `required` at the *top level* of `spec`, where it means "always".

That class is structurally empty here. `+required` is emitted by `WriteField` into
`types.generated.go`, which holds only nested types; the top-level `<Kind>Spec` lives in the
hand-written `<kind>_types.go`, which `generate-types` never overwrites. Verified across the whole
tree: no generated `<Kind>Spec` struct belongs to a CRD of the same service. (Three apparent hits —
`BigQueryRoutineSpec`, `BigQueryTableSpec`, `ServiceSpec` — are datacatalog's own nested proto types
colliding by name with unrelated Kinds in other groups.)

**The real risk is a type reused across contexts with different requirements**, and the 18 status
entries are exactly that. Nested `required` handles "optional parent, required child" correctly; what
it cannot express is "required when a user supplies this, not guaranteed when GCP returns it".

The status entries are the reason for the gate. Nested message types are generated **once** by
`WriteMessage` and shared between the spec and the observed state, so a marker derived from a field's
own annotation lands in every schema position that type occupies — redis `PscConfig` /
`PscConnection` are the clearest case. CRD structural validation covers the status subresource, so if
GCP returns an object missing such a field, KCC writes a status the API server rejects and
reconciliation fails at runtime. That is worse than an apply-time tightening, and it is invisible if
you only look at the field being annotated.

Suppressing it in `WriteObservedStateMessage` does not fix it, because the shared struct is written by
`WriteMessage`. It is done anyway, on the principle that an observed-state struct describes what GCP
returned and should never constrain it.

*Remaining risk once opted in:* a resource whose proto marks a field REQUIRED that KCC intentionally
treats as optional. That is a real divergence and needs a decision, but it is now confined to services
that opted in.

### Phase 2 — feed `google.api.resource` into the scaffolder

Thread `pattern` and `plural` into `APIArgs` so the identity template stops hardcoding
project+location and naive pluralisation.

Measured over the 1417 messages carrying `google.api.resource`, using
`protoapi.GetResourceMetadata` — the production path — rather than a separate reimplementation:

| Bucket | Count | Share |
|---|---:|---:|
| correct | 562 | 39.7% |
| wrong: casing only | 554 | 39.1% |
| wrong: pluralisation | 198 | 14.0% |
| not comparable (pattern declares no collection) | 103 | 7.3% |
| **wrong overall** | **752** | **53.1%** |

*(An earlier draft said 60.1%. That came from a throwaway probe that normalised every `{placeholder}`
to `*` and took the second-to-last segment, so patterns ending in a literal — `projects/{project}/locations`
— were scored as wrong when there is simply nothing to compare. Those are the 103 above.)*

The two failure modes are worth keeping apart, because they are different arguments:

| Mode | Proto message | Template emits | API uses |
|---|---|---|---|
| casing | `LbTrafficExtension` | `lbtrafficextensions` | `lbTrafficExtensions` |
| casing | `ChannelGroup` | `channelgroups` | `channelGroups` |
| pluralisation | `Batch` | `batchs` | `batches` |
| pluralisation | `Property` | `propertys` | `properties` |

The `LbTrafficExtension` row is the pilot, so this was already being fixed by hand.

**On casing specifically: this is not KRM naming.** Two different namespaces, and only the second is
in scope here:

| | Example | Set by | Covered by |
|---|---|---|---|
| KRM field name | `spec.forwardingRules` | `GetJSONForKRM` | `TestCRDsAcronyms` |
| GCP collection segment | `projects/p/locations/l/lbTrafficExtensions/x` | identity template | nothing |

KRM field names must not follow proto casing, and nothing here changes them — `GetJSONForKRM` is
untouched and no CRD field name moves. The collection segment is part of a **GCP resource name**,
written into `status.externalRef` and into request URLs, so it has to match GCP byte-for-byte.

### The naive casing has already shipped, and it is observably wrong

Auditing the 90 `_identity.go` files that can be matched to a declared pattern finds **9 mismatches,
every one of them casing**:

`apphubdiscoveredservice`, `apphubdiscoveredworkload`, `bigquerydatapolicy`,
`clouddmsconversionworkspace`, `dataprocnodegroup`, `discoveryenginedatastore`,
`managedkafkaconsumergroup`, `netappbackupvault`, `storagemanagedfolder`.

Checked against an oracle independent of the proto — recorded GCP traffic under
`pkg/test/resourcefixture/testdata/` — camelCase is what GCP actually receives: `backupVaults` 27
files to 0, `nodeGroups` 6 to 0, `conversionWorkspaces` 5 to 0.

`StorageManagedFolder` is the clearest case. GCP's own response body, in `_http.log` after the `OK`
marker — a **response**, not one of KCC's requests, which is the distinction that matters when
citing this file as evidence:

```json
{
  "createTime": "2024-04-01T12:34:56.123456Z",
  "metageneration": "1",
  "name": "projects/_/buckets/bucket-${uniqueId}/managedFolders/managedfolder-${uniqueId}/",
  "updateTime": "2024-04-01T12:34:56.123456Z"
}
```

The normative source agrees and does not depend on a recording at all:
`+kcc:spec:proto=google.storage.control.v2.ManagedFolder`, whose `google.api.resource` pattern
declares `managedFolders`.

Against that, three sources inside KCC disagree with GCP and with each other:

| Source | Format |
|---|---|
| GCP (response above, and the proto pattern) | `projects/_/buckets/{b}/managedFolders/{f}/` |
| `ParseManagedFolderExternal` (`_identity.go:107`) | requires `buckets` and `managedfolders`, exactly 6 tokens |
| `ManagedFolderRef` doc comment (`_reference.go:34`) | `projects/{p}/locations/{l}/managedfolders/{id}` |

The doc comment says `locations` where the parser demands `buckets`; both disagree with GCP on
casing. GCP's trailing slash makes `strings.Split` yield seven tokens, so a name copied from GCP
fails the length check before casing is even reached.

The path is user-reachable: `NormalizedExternal` (`_reference.go:53`) parses a user-supplied
`external:` value and returns the error. KCC's own round-trip still works, because the writer at
`_identity.go:34` and the parser at line 107 share the same wrong constant — the system is internally
consistent and disagrees only with the outside world, which is why nothing catches it.

**No existing test covers this.** `TestCRDsAcronyms` checks acronym casing in CRD *field* names,
`shortname_pluralization.txt` checks CRD `shortNames`, and `naming_violations.txt` checks *file*
naming. `naming_test.go:67` mentions `_identity.go` only as a filename suffix. Nothing inspects a
collection segment or compares it to the proto pattern — which is why these nine shipped unnoticed.

**Parent handling is deliberately narrow.** Only `projects/*` and `projects/*/locations/*` are
specialised, because those are the shapes the scaffolded spec supports; an org- or folder-parented
resource has no matching spec field, so generating code for it would not compile. Everything else
keeps the projects/locations shape and gains a TODO naming the real pattern. The collection fix
applies regardless of parent shape.

**No opt-in flag needed**, unlike phase 1: every scaffold path is guarded by an "already exists,
skipping" check, so the scaffolder only ever writes new files and cannot change an existing resource.

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

**The queue must always mark the resource, whatever the detector found.** This was designed
annotation-driven — field entries from `google.api.resource_reference` — and implementing it killed
that design. `LbTrafficExtension` carries no `resource_reference` on any field, including
`forwarding_rules`, which is exactly the field that has to become a ref:

```
google.cloud.networkservices.v1.LbTrafficExtension
  name                  resource_reference: -
  forwarding_rules      resource_reference: -   <-- must become a ref
  extension_chains      resource_reference: -
```

An annotation-only queue would have been empty, so no file would have been written, no suppression
would have happened, and the resource would have gone straight into the ratchet and failed — the
precise thing the queue exists to prevent. So a resource-level entry is emitted unconditionally, and
field-level entries are a bonus rather than the mechanism.

**Emit undecided fields, do not omit them.** A field emitted with a listed open question is visible;
an omitted one is invisible to every other check, because a field absent from the CRD cannot be
reported as missing from it.

**ObservedState is out of scope.** Output fields reached through nested messages need the generated
`<Proto>ObservedState` variants rather than the plain structs, and choosing per field is its own
problem.

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

### Suppression must not look like a fix

A queued resource contributes no findings, so anything it *already owed* reads as removed, gets
pruned from the ratchet, and then reappears as a **new** violation the moment it graduates — failing
the check for work nobody did.

The fix is to carry those baseline entries forward, so queueing only ever stops findings being
*added*. Verified against `AlloyDBBackup`: queueing it without the carry-forward reported its two
entries as fixed; with it, they stay.

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
