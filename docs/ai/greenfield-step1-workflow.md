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
| 2b. Queue hints | `scripts/queue-hints` seeds the queue | Yes |
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

### The queue covers status as well as spec

It used to describe spec fields only. Every `JudgementItem` was appended inside `PrepopulateSpec`,
so a field missing from `status.observedState` was reported nowhere — measured across the greenfield
set, that was 156 defects against zero entries. These reasons now appear too:

| reason | what happened |
|---|---|
| `unsupported-field-type` | the type was declined; the field is a `// TODO:` and never reaches the CRD. Emitted for spec and observedState alike |
| `observedstate-identity-field-omitted` | the proto marks it OUTPUT_ONLY, but KCC carries the resource name in `status.externalRef`. A decision, not a bug — but it was an invisible one |
| `output-only-in-comment-only` | the proto comment says output only while carrying no `field_behavior` annotation, so the field was generated into the Spec. The entry names `.status.observedState.<field>`, where it belongs, not where it currently sits |
| `possible-reference-by-sibling` | the field's name matches a resource this service declares, so it may want to be a reference. See the sibling rule below |
| `parent-segment-matches-sibling` | a segment of the resource's own name is also a resource this service manages — the strongest reference hint the parent walk produces |

**Some entries are comments, and that is deliberate.** A finding against a *shared nested message*
names no single Kind, and every non-comment line in this file suppresses `[refs]` for the Kind it
names, so inventing one would quietly switch off a real check. Those are written as `#` lines
instead:

```
# possible-reference-by-sibling: google.cloud.compute.v1.NetworkInterface.subnetwork target=ComputeSubnetwork
# dropped: google.cloud.bigquery.migration.v2alpha.TranslationTaskDetails.specialTokenMap reason=unsupported map type with key string and value enum
```

Read them; they are findings, not decoration. `silence_report.py` parses the sibling ones by leaf
name, which is looser than the path matching it uses for real entries.

**A resource is not finished when its fields are generated. It is finished when nothing about it is
silent.** A field a human must decide is a fine outcome; a field nobody was told about is not. See
[greenfield-coverage-invariant.md](greenfield-coverage-invariant.md) for what that means and how to
measure it — `hack/tools/greenfield/silence_report.py` compares every field of every baseline CRD
against ours and reports implemented / flagged / missed.

Read the `missed` breakdown rather than the total. Only its first line, **truly missed**, is a field
we produce nowhere; the lines under it are fields we do produce, in the wrong section or as a plain
string where upstream has a reference. Those need detection or placement work, not generation, and
treating them as one number sends people to write generators for fields the types file already
carries.

## Stage 2b — Seed the queue with hints

Run after the CRDs exist, because the detectors read them:

```bash
go run ./scripts/queue-hints
```

**Do not skip this.** The 189-resource bulk run did, and it cost 151 reference fields that nobody
was told about — the largest single hole in that run's coverage, closed by running a tool that
already existed.

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
| `possible-reference-by-sibling` | the name matches a resource the service declares — a hint, 77% right |

It also flags a resource whose generated `status.observedState` came out with **no fields at all**,
as `resource reason=empty-observedstate`. That is a fact rather than a guess — the CRD either has
properties there or it does not — and it caught 36 of the 189 resources in the bulk run, including
`ComputeInterconnect`, whose upstream CRD has 19 observed fields against our zero. The usual cause
is a proto carrying no `google.api.field_behavior` anywhere, which leaves the generator nothing to
identify an output field by, so everything lands in the Spec.

Roughly a third of reference hints are wrong, and that is the intended trade for a list a person
confirms.
Most of the wrong ones are fields whose exact name upstream made a reference in a *different*
resource, so they are worth a moment's thought rather than a reflex dismissal.

### How the reference detectors work, and how to extend them

Four detectors feed the queue. They differ in what they can reach, and the difference decides where
to put effort.

| detector | source | reaches | derivable? |
|---|---|---|---|
| `google.api.resource_reference` | the proto — a statement, not a guess; names the exact target | `possible-reference` | it is a fact, not a rule |
| `refs.Classify` | the field's **description** | `possible-reference-by-description`, and gates `TestMissingRefs` | yes |
| **sibling resource in the same service** | the field's **name**, against the service's own Kinds | `possible-reference-by-sibling`, generator | yes |
| `refs.MatchName` / `refs.NameRules` | the field's **name**, against a list of known spellings | `possible-reference-by-name`, seeder only | no — learned |

That last column is the one that matters. **`refs.NameRules` only ever finds a reference somebody
has already seen.** It is a lookup table of spellings; a service using a spelling nobody has met gets
nothing. The other three work on a service no one has looked at, because each reads something the
service itself supplies: an annotation, a description, or its own list of resources.

So: **a growing `NameRules` list is a signal that one of the other three is missing something, not a
sign of progress.** Check the descriptions first, every time.

The three rules that exist each came from a measurement:

* `hasSuffix("SecretVersion")` — every confirmed instance in the corpus ended that way, and the
  suffix is long enough not to collide with anything else.
* `eq("network", "vpc", "vpcName")` — admitted as a **whole leaf only**, never a substring, because
  `network` is the exact name the rejected 2,164-finding heuristic was built on. `networkConfig` and
  `networkPolicy` are not networks.
* `eq("kmsKey", "cmekKeyName", "encryptionKey", "kmsKeyName")` — four observed spellings of one
  thing. `EventarcChannel` calls it `cryptoKeyName`, a fifth, and it is missed today. That one line
  is the limitation in miniature.

#### The sibling rule

If a service declares a resource called `DataStore`, then a string field called `dataStore` is
probably a reference to it. That is the whole rule. It lives in the generator not the seeder,
in `codegen.SiblingResource`, so it can write a `+kcc:guess=possible-reference target=…` marker onto
the field as well as a queue entry. The seeder is a post-generation pass over CRDs and can only do
the second.

As built, it marks 25 fields across 18 distinct names, at **77% precision** over the ones upstream
actually modelled: 10 references, 3 kept plain, 5 upstream never modelled and so excluded. That
matches the 75% predicted from the CRDs before any of it was written, which is the reassuring part.
Loosening the exact match to `endswith` buys seven more at 68%, and starts pulling in
`spec.pipelineJob` and `localSsds[].interface`, so the exact form is what is implemented. The false
positives are honest and worth knowing: DataLabeling's `annotationSpecSet` and `instruction` both
match sibling Kinds and upstream keeps them plain.

Measure it with `hack/tools/greenfield/sibling_precision.py`, which scores every marker against the
baseline CRDs and excludes fields upstream had no opinion on — a rule cannot be wrong about a field
nobody modelled.

It reaches three places, and each was needed separately. Built for the top-level loop alone, the
rule fired on nothing at all:

* nested message fields, in `WriteField` via `WriteOptions.Siblings`. 16 of the 25, in shapes like
  `spec.selector.targets[].targetRef`.
* top-level spec fields, in `PrepopulateSpec`. 9 of the 25.
* synthesised parent segments, in `AddTypeFile`, 7 of which now name a sibling. `FirestoreDocument`'s
  `database` and `DiscoveryEngineDataStoreTargetSite`'s `dataStore` come from the resource pattern
  rather than from any proto field, so the name has to be matched directly with
  `SiblingResourceByName`.

One thing the nested path must not do is mark the `/* found existing non-generated go type …
skipping */` dumps. Those show what the generator would have written for a type the package already
declares by hand; nothing in them reaches the CRD, and the collector only scans messages that are
really emitted. Left on, they produced 63 of 82 markers and every one broke the marker-implies-entry
rule. `WriteMessageAsComment` clears `Siblings` for exactly that reason, and both the checker and the
precision tool skip `/* */` regions.

Plurals are matched too, by a deliberately narrow singulariser: `subnetworks` finds `ComputeSubnetwork`,
`dnsAuthorizations` finds `CertificateManagerDNSAuthorization`. It gives up on anything
irregular, and refuses to touch a word ending `ss`, `us` or `is` so that `status` and `access` are
left alone.

The list grows itself, which is the property that makes this worth having. The sibling set is built by
`scaffold.SiblingResources`, from the Kinds the service package declares plus the Kinds of the run in
progress. Nobody maintains it. Every resource added to a service makes the rule stronger for every
other resource in that service, which is exactly the property you want when generating a service's
resources in bulk. Contrast `NameRules`, which only improves when a person notices something and
writes it down.

Two things to know when reading the output. Nested matches are written into
`needs_judgement_call.txt` as `#` comments rather than entries, because a nested message is shared by
every resource that references it and there is no single Kind to attribute it to. Every
non-comment line in that file suppresses `[refs]` for the Kind it names, so a made-up Kind would
quietly switch off a real check. The marker itself is a comment, so `controller-gen` strips it before
the CRD is published; it cannot affect the schema, which is why this one is on for every service
rather than opt-in like `--emit-required-from-proto`.

If you meet a reference the detectors missed, the order to try is: fix `Classify` if the
description gives it away, check whether the target is a resource the service declares and the
sibling rule simply has not been generated yet, and only then add the spelling to `NameRules`.

#### Adding a rule

1. **Read the descriptions of the fields you are trying to catch.** If they carry a resource-name
   template, or say something like "Secret version reference", fix `Classify` instead — it will
   catch the same fields everywhere, including services nobody has generated yet.
2. **Measure precision before adding**, over fields upstream actually modelled: how many fields we
   generate carry that leaf name, against how many upstream turned into a reference. Exclude fields
   upstream has no opinion on — a rule cannot be wrong about a field nobody modelled — and exclude
   what an existing rule already covers, or you will measure that rule instead of yours.
3. **One rule at a time**, with `go run ./scripts/queue-hints -dry-run` before and after, so the
   added hint count is attributable.
4. **Record the number next to the rule.**

Measurements already done, so nobody repeats them:

| candidate | upstream made a Ref | upstream kept it plain | verdict |
|---|---|---|---|
| word `secret`, excluding the `SecretVersion` suffix | 4 | 0 | worth adding |
| `project` | 4 | 0 | worth adding |
| `dataStore`, `cluster`, `topic` | 3, 3, 2 | 0 | worth adding |
| `subnetwork` | 4 | 2 | marginal |
| `service` | 8 | 11 | no |
| `source`, `version` | 2, 2 | 8, 15 | no |
| ends `Certificate` | 0 | 6 | **no** — `pemCertificate`, `caCertificate` are contents, not refs |
| ends `PrivateKey` | 0 | 3 | **no** |
| `password` / `pass` | 0 | 0 | no evidence either way |

The last three are the trap worth naming. `ConnectorsConnection` models `clientCertificate` and
`clientPrivateKey` as references, and a rule learned from it fires on
`CertificateManagerTrustConfig.spec.trustStores[].intermediateCAs[].pemCertificate`, which is a PEM
blob. Past knowledge generalised badly.

**What `NameRules` cannot do**: find a reference in a service nobody has looked at. Coverage here is
not coverage generally, and the number to watch is the undetected count from
`silence_report.py`, not the size of this list.

### The generator flags

Every service's `generate.sh` carries the same five:

| flag | what it does |
|---|---|
| `--prepopulate-spec` | fill the Spec from the proto instead of a three-field stub |
| `--emit-required-from-proto` | `// +required` for fields the proto marks REQUIRED |
| `--emit-plural-acronyms` | `relatedURIs` rather than `relatedUris` |
| `--detect-output-only-in-comments` | report fields whose comment says output-only with no annotation, in either spelling (see below) |
| `--place-server-set-fields` | put `createTime`, `uid`, `selfLink`, `etag` and friends in ObservedState when the proto carries no `field_behavior` at all |

#### Two spellings of "output only"

A proto that documents a field as server-set in prose, without a
`google.api.field_behavior` annotation, gets that field generated into the Spec. The detector reads
the comment, and there are **two conventions**:

| spelling | who writes it |
|---|---|
| `Output only. …` | most Google APIs |
| `[Output Only] …` | Compute |

Only the first was recognised for a long time, so every Compute resource lost the signal entirely.
`compute.proto` alone carries 1,605 fields with the bracketed form, and all ten of
`ComputeInterconnect`'s misplaced observed-state fields — `googleIPAddress`, `circuitInfos`,
`expectedOutages`, `availableFeatures` and the rest — were nothing more exotic than that.

Both are matched as a **prefix**, not anywhere in the comment. That is the convention in practice:
of the Compute fields carrying the marker, 1,600 open with it and 5 mention it mid-sentence. Those
five are left alone, because the anchored test is the one whose false-positive rate was measured —
across 4,673 fields in hand-written Spec structs in the baseline tree, neither spelling appears once.

If you meet a third convention, add it to `outputOnlyPrefixes` in `scaffold/prepopulate.go` and
re-run the false-positive count over the baseline before trusting it.

#### A flag on no script does nothing, and says nothing

`--place-server-set-fields` was built, tested and then wired into nothing for weeks. The other four
were on all 94 scripts and it was on zero, which is why server-set fields kept showing up as missed
and `+kcc:guess=placement` appeared nowhere.

The same thing happens per service. **37 of 131 `generate.sh` scripts pass no `--prepopulate-spec`
at all**, so their resources scaffold as three-field stubs and every upstream field scores as
absent. `speech` was the one with in-scope resources, and its three Kinds alone accounted for 43 of
the 232 unflagged misses at the time — the single largest cause, and a one-line fix.
`SpeechRecognizer` went from 3 fields to 30 the moment the flag was added.

Check with:

```
for f in apis/*/generate.sh; do grep -q prepopulate-spec "$f" || basename "$(dirname "$f")"; done
```

Add the flag to the **v1alpha1** invocation only, which is the house pattern. A service's v1beta1
call deliberately runs without it: `--prepopulate-spec` scaffolds a types file that does not exist,
so putting it on the beta invocation would invent beta resources upstream does not have.

If you add a generator flag, add it to the scripts in the same change.

### Regenerating the whole corpus

```bash
dev/tasks/greenfield-regenerate
```

Use the script, not the steps by hand. It wipes the types files, runs each `generate.sh` with
`SKIP_GENERATE_CRDS=1`, rebuilds CRDs per service, post-processes and publishes them, **and re-seeds
the judgement queue**.

That last step is why the script exists. **Regeneration rewrites every `needs_judgement_call.txt`
from scratch and silently discards everything `queue-hints` wrote.** It has now cost two runs: the
original bulk generation never ran the seeder, and the run after it wiped 367 hints — `flagged` fell
from 321 to 77 and `missed` jumped from 262 to 472 before anyone worked out the cause was procedural
rather than real.

CRDs are skipped during generation for a separate reason. `generate.sh` calls
`dev/tasks/generate-crds`, which runs `controller-gen` over `./...` and fails wholesale the moment one
package will not load. In this tree 43 do not, by design, so it can never succeed; the script rebuilds
them per service instead.

### Compile the service before scoring it

```bash
go build ./apis/<service>/...
```

Do this first, every time. Where a resource has hand-written `_identity.go` or `_reference.go` files
— which every resource upstream has already implemented does — those files are the strictest oracle
available, and the compiler names a missing field in seconds where a scored run takes minutes and
reports a path.

It is what would have surfaced the parent-identity gap on day one. Across the corpus it names 29
distinct fields, and almost all of them are a parent segment or a parent reference:
`Spec.OrganizationRef`, `Spec.Location`, `Spec.EntryGroupRef`, `Spec.DatabaseRef`, `Spec.Tenant`.

**Count the distinct undefined fields, not the failing packages.** A package fails whole on a single
type mismatch, so the package count moves far slower than real progress: one regeneration took the
distinct fields from 29 to 20 while the package count went 44 to 43.

```bash
go build ./apis/... 2>&1 | grep -oP '\.(Spec|ObservedState)\.\w+ undefined' | sort -u | wc -l
```

Two things it will not tell you, so it supplements the score rather than replacing it. It only sees
fields the controller code *dereferences*, which is 18% of what the CRD comparison finds — a missing
`spec.labels` or a field placed in status compiles perfectly. And it needs a hand-written companion
to exist at all, so it says nothing about a genuinely new resource, which is the case this whole
process is for.

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

### The measurement tools

Separate from the artifacts above, because these produce numbers rather than files the build reads.
All live in `hack/tools/greenfield/`.

| tool | answers |
|---|---|
| `silence_report.py` | how much of upstream's CRD surface we reproduce, and how the rest is accounted for. The one to run before and after any generator change |
| `check_guess_entries.py` | does every `+kcc:guess` marker have a judgement-queue entry? Must stay at 0. Run it after any change to what the generator emits |
| `sibling_precision.py` | scores the sibling rule against what upstream actually did, excluding fields upstream never modelled |
| `calculate_coverage.py` | the wider resource-level coverage number, not this per-field one |

`dev/tasks/greenfield-regenerate` runs the whole corpus and finishes with `check_guess_entries.py`,
so a broken invariant fails the run rather than waiting to be noticed.

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
- **A finding written as a comment in the queue used to be discarded by the next invocation.** The
  merge in `writeJudgementQueue` skipped every `#` line when reading the existing file, so only the
  invocation that happened to run last kept its comment-form findings — and comment form is exactly
  what the shared-nested-message findings use, because a non-comment line suppresses `[refs]` for
  the Kind it names. `compute` kept 6 of its 56 sibling matches that way and `backupdr` none of its
  2. Fixed by preserving everything but the header, which is rewritten. If you add a finding that
  cannot be attributed to a single Kind, write it as a comment and check it survives a service with
  two `generate-types` calls.
- **A marker and its queue entry must share a gate.** `WriteField` writes `+kcc:guess` on every
  invocation, but the queue block for the sibling rule was written only under `--prepopulate-spec`.
  `backupdr` scaffolds `v1alpha1` with that flag and regenerates `v1beta1` without it, so two real
  references landed in `types.generated.go` with a marker no queue file mentioned.
  `check_guess_entries.py` is what caught it; run it after any change to what the generator emits.
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
- **An existing `_types.go` is never rewritten, and that is invisible.** `--prepopulate-spec` only
  writes a types file that does not already exist, and `AddTypeFile` skips one that is there — which
  also means the resource's judgement entries are never written. Regenerating in place therefore
  keeps the old Spec *and* the old queue, so a service can be regenerated with a new generator and
  show none of it. `TPUVirtualMachine` sat as a three-field stub through an entire measurement
  round with `--prepopulate-spec` present in its `generate.sh` the whole time; deleting the file
  first took it from 0 annotated fields to 24. **Delete the types file before regenerating** when
  you want the new behaviour, and check the resulting Spec has more than three fields.
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
