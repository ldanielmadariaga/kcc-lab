# Greenfield quality enforcement: how the checks are designed

This covers the experimental sandbox (`kcc-lab`), not upstream policy. It describes what is built
and what is still in review; the open questions are at the end.

## What this document is for

Four other documents already explain what to build and how to build it:

| Document | What it answers |
|---|---|
| `greenfield-coverage-strategy.md` | Which resources to implement, in what order, and to what standard. |
| `greenfield-tracker.md` | How to pick the next piece of work. |
| `greenfield-bulk-generation.md` | The procedure for generating one resource. |
| `greenfield-generator-mechanics.md` | Changes to the generator so that procedure has less manual work in it. |

This document explains the remaining half of the system. Once the resources exist, how do we stop
their quality from slipping, and how do we raise it over time? That is what most of the check work
actually is.

Throughout, "Step 1" means the first stage of generating a resource, which produces the API types
and the CRD but no controller, no test fixtures and no MockGCP support. Several of the decisions
below only make sense once you know that a Step 1 resource is deliberately incomplete.

Some of the mechanisms below are merged and some are still in review. Each one is marked, so that
this document does not describe things that are not there yet.

| Mechanism | PR | Status |
|---|---|---|
| Manifest scoping | #5 | merged |
| Dropped-field ratchet | #6 | merged |
| Unit tests for the checkers | #7 | merged |
| Reference detector, `missingrefs` as a ratchet, `refs_deferred` | #9 | merged |
| GCS reference taxonomy refinements | #10 | open |
| Additional checks derived from the skills | #11 | open |
| Step 1 scoping, enum prohibition, nested dropped fields | #12 | open |
| Judgement queue and `[refs]` suppression | #16 | open |
| Identity collection casing | #18 | open |

---

## The problem

The strategy is to generate a large number of resources at low quality on purpose, and then improve
them in passes across the whole corpus. That creates two opposite ways for the system to fail, and
most of the design is about staying between them.

The first is absorption. A newly generated resource has defects, those defects get written into an
exception file, and nobody notices. The file grows a little each time. The check still exists and
still passes, so everyone believes the standard is being enforced when it is not.

The second is retroactive failure. Someone writes a new check and it immediately fires on hundreds
of resources that were written long before the check existed. There are then only two ways to land
it: weaken the check until it stops complaining, or add every existing resource to an exception
list. Both of those are absorption arriving by a different route.

---

## The grandfathering rule

This is the policy everything else has to respect, so it is worth stating precisely. It has three
parts.

Resources that already existed keep whatever shortcomings they have. We are not going to fix
hundreds of resources as a precondition for improving new ones. Those shortcomings are recorded in
a baseline file and nobody is required to clean them up.

Those same resources are still not allowed to get worse. Recording a shortcoming is not the same as
inviting more of them, so if somebody edits an old resource and introduces a new violation of the
same kind, that fails.

The full standard applies only to resources we have identified as bulk-generated. Membership is
explicit: a Kind is listed in `tests/apichecks/testdata/greenfield_bulk.txt`, it joins that list
when it is generated, and it never leaves.

Together those produce three tiers:

| Resource | Defects it already has | A new defect | The full greenfield standard |
|---|---|---|---|
| Predates this work | Grandfathered in a baseline | Fails | Not applied |
| Predates this work, later edited | Still grandfathered | Fails | Not applied |
| Listed in the bulk manifest | Recorded, and expected to be fixed | Fails | Applied |

When you add a check, the rule decides where it goes. If the check expresses the new standard,
scope it to the manifest, in the way the `TestGreenfield*` checks do. If instead it expresses
something that was always meant to be true, and you only want to stop it getting worse, it can run
across the whole repository. In that case you have to seed its baseline with every violation that
already exists, so no existing resource is being asked to change.

Two checks currently run across the whole repository on that second basis: `TestMissingRefs` and
`TestIdentityCollectionCasing`. Both were seeded with the violations that were already there, so no
resource that predates them is failing today. They prevent regressions; they do not impose the new
standard.

---

## Two kinds of baseline file

This is the most important distinction in the check code, and you cannot see it from the filenames,
because both kinds write a file into `testdata/exceptions/`. The difference is which function the
test calls at the end.

| | `CompareGoldenFile` | `CompareRatchetFile` |
|---|---|---|
| A new violation appears | It is written into the file when you run with `WRITE_GOLDEN_OUTPUT=1` | The test fails, and it fails even when that variable is set |
| A violation is fixed | It disappears when the file is rewritten | It is pruned from the file, and the test logs that it did so |
| Which way the file can move | Either way | It can only shrink |
| What the file describes | The corpus as we inherited it | Work we are doing right now |

On master there are 17 golden files and 2 ratchets. The ratchets are `missingrefs.txt` and
`greenfield_dropped_fields.txt`. A third, `identity_collection_casing.txt`, is waiting in #18.

The golden files are not a failure of nerve. A golden file is the right choice when a class of
finding is large, already spread throughout the corpus, and not something anyone is working on at
the moment. It records the situation and makes any growth visible in code review, which is all you
need when nobody is touching that area.

A ratchet is the right choice when a new entry would mean somebody did the work incorrectly today.
That is the test to apply: if a finding could be produced by work that is currently in flight, it
needs a ratchet, and if it could not, a golden file is honest and cheaper.

`missingrefs.txt` was converted from a golden file into a ratchet for exactly that reason. Deciding
which fields are references is the judgement that bulk generation most often gets wrong, so a new
entry there means something needs fixing now.

---

## Scoping checks to the manifest

`tests/apichecks/testdata/greenfield_bulk.txt` lists the Kinds held to the new standard. It
currently contains one entry:

```
networkservices.cnrm.cloud.google.com/NetworkServicesLBTrafficExtension
```

A check that applies to every resource can only ever be as strict as the worst resource in the
repository, because anything stricter would fail on that one. Scoping to an explicit list lets a
check be strict from the day it is written, without touching anything that came before it.

The Kind on its own is enough to find every file belonging to a resource, because
`TestDirectResourceFileNaming` already requires files under `apis/` and `pkg/controller/direct/` to
be named with the lowercased Kind as a prefix.

There is a cost to this and it is worth being honest about it. The manifest is a second thing that
has to be kept in step with reality, and a resource missing from it is silently not checked.
`TestGreenfieldBulkManifestIsResolvable` limits the damage: every entry must resolve to a real CRD
and to real files on disk, so a stale entry produces a clear error rather than quietly checking
nothing.

---

## Every finding names the resource it came from

Findings are written with the resource attached, either as `crd=` or as `kind=` and `group=`:

```
[refs] crd=alloydbbackups.alloydb.cnrm.cloud.google.com version=v1alpha1: field ".spec.encryptionConfig.kmsKeyName" should be a reference
```

Without that, a baseline file is several hundred lines of text nobody can act on, and there is no
way to prune the entries for one resource when somebody fixes it. With it, answering "what does
this resource still owe?" is a single grep.

---

## The four states a reference finding can be in

Deciding which string fields are references is the hardest judgement in Step 1, and the one most
likely to be got wrong, so it gets four states rather than a simple yes or no. A resource is in
exactly one of them at any time, which is what stops the files contradicting each other.

| State | File | Written by | Kind of file |
|---|---|---|---|
| Not yet triaged (#16, open) | `apis/<service>/needs_judgement_call.txt` | The generator | A work queue, one per service |
| Triaged, deferred, with a reason | `refs_deferred.txt` | A human | Curated input, subtracted from findings |
| Triaged, still owed | `missingrefs.txt` | Computed | Ratchet |
| Structurally impossible today | `refs_not_representable.txt` | Computed | Golden |

Three things about this are not obvious, and each was learned by getting it wrong first.

`missingrefs.txt` cannot hold an explanation. The test recomputes it from the CRDs on every run, so
any note a human adds is erased the next time somebody runs with `WRITE_GOLDEN_OUTPUT=1`. That is
why `refs_deferred.txt` exists as a separate, hand-edited file in which a reason is mandatory. It
is subtracted from the findings before the ratchet is applied.

The queue exists so the mechanical generation pass can be merged at all (#16, open). A generated
resource has reference-shaped string fields by construction, and `missingrefs.txt` is a ratchet, so
without the queue the very first bulk-generation pull request would be blocked. While a resource
has open entries in the queue, its `[refs]` findings are suppressed.

Suppressing a resource must not look like fixing it, and this is the trap worth reading twice. A
suppressed resource produces no findings at all, so anything it already owed looks as though it has
been fixed, and the ratchet prunes those entries. When the resource is later triaged and leaves the
queue, those same entries come back, and now they count as new, so the check fails for work nobody
did. The fix is to carry a suppressed resource's existing baseline entries forward, so suppression
only ever stops findings being added.

`refs_not_representable.txt` is deliberately a golden file rather than a ratchet. Its entries are
not actionable by construction, because they describe fields KCC has no way to represent yet, so
the list is expected to grow as more resources arrive. Growth is visible in review and every entry
carries the reason it is there.

---

## When a proto annotation can be used as a check

Proto annotations look like ground truth, but they are not all usable in the same way. The question
that decides it is not how many fields carry the annotation. It is what you are entitled to
conclude when the annotation is absent.

| Annotation | Coverage | What absence means | Usable as a check? |
|---|---:|---|---|
| `google.api.field_behavior` | 40.7% | Treat the field as optional, which is what already happens | Yes, because doing nothing is the correct fallback |
| `google.api.resource_reference` | About 15%, and 0% across compute | Nothing at all; the field may still be a reference | No, because any fallback has to guess |

Matching `resource_reference` by field name was built and measured before being rejected. It
flagged 2,164 fields as probable references, against 78 for the description-based heuristics alone.
The cause is that a name such as `network` is annotated in one service and then appears on hundreds
of unrelated fields elsewhere. It was removed rather than tuned, because tuning would not have
addressed that.

The underlying obstacle is structural. A CRD does not record which proto field each of its fields
came from, so the only thing connecting the two at that layer is the leaf name, and the leaf name
is too weak to carry the connection. Using the annotation properly would mean mapping each CRD
field back to its proto field through the `+kcc:proto:field=` markers the generator emits, and then
looking up the fully-qualified path. That is a worthwhile change, but it is a different one, and it
would reuse nothing from the leaf-name approach.

So the rule is to use an annotation as the basis for a check when its absence has a safe default.
Otherwise it belongs in the generator, which can afford to be wrong in ways a check cannot.

---

## Choosing something to check against

Every check needs a source of truth to compare against. Which one you pick determines whether the
check can run in CI at all, and that matters more than whether it is the most authoritative source
available.

| Source | Authoritative? | Committed to the repo? | Notes |
|---|---|---|---|
| The proto descriptor (`.build/googleapis.pb`, the compiled API definitions) | Yes | No, it is gitignored | A check that needs it will skip silently in CI |
| Recorded traffic (`_http.log`, the HTTP exchanges captured by the test fixtures) | Nearly | Yes, 894 files and about 23 MB | Contains KCC's own requests as well as GCP's responses |
| The CRDs | No | Yes | Derived from the KRM types, so it cannot answer questions about GCP's own formats |

`TestIdentityCollectionCasing` (#18) is the worked example. It asks whether the collection segment
an identity builds matches what GCP actually uses. The proto pattern is the authoritative answer,
but a check reading the proto descriptor would skip in CI, where that file does not exist. So the
check uses recorded traffic instead, and its limitation is written into the test: it can only catch
a wrong casing when some other part of KCC gets that casing right. A resource wrong everywhere
would pass, and resources with no test fixtures are not covered at all. That is why its baseline
lists four entries where an audit against the proto finds nine.

Trading authority for a check that actually runs is usually the right call. What matters is that
the compromise is written down in the test rather than left for the next reader to work out.

---

## The checkers are code, so they are tested

Detection logic has bugs like any other code. What makes checker bugs different is that they are
invisible. A broken checker reports fewer findings, everything passes, and there is nothing to
investigate.

Two examples from this work.

The first was a parser that extracted the key from a baseline entry. It silently produced
`"foos.example.com:"`, with a trailing colon, which meant it never matched anything again. A unit
test found it. The corpus never would have, because the symptom was silence.

The second was the regular expression in `TestIdentityCollectionCasing`, which has to be anchored
on `parent.String()`. Without the anchor it matches the `/locations/` inside the *parent's*
`String()` method instead, so every resource appears to use `locations` as its collection. The
casing of `locations` is correct everywhere, so the check passes while verifying nothing. This was
confirmed by removing the anchor: the four findings dropped to zero and the test still passed.
`TestIdentityCollectionRegex` now asserts that the expression extracts the collection and not the
parent.

Because the characteristic failure of this code is silent weakening rather than noise, the tests
target the extraction and matching functions directly, and not only the end-to-end result.

---

## What is deliberately left unenforced

Field coverage (`TestGreenfieldBulkFieldCoverage`) only runs when `GREENFIELD_STRICT=1` is set,
which means it does not run in CI. Step 1 legitimately fails it, because coverage is measured
against test fixtures and Step 1 does not produce any. Making it a CI gate would force a choice
between blocking Step 1 and weakening the check, and neither is worth it while fixtures belong to a
later step.

`alpha-missingfields.txt` is expected to grow during Step 1 for the same reason. Its entries are
attributed by `crd=` and are removed once fixtures exist.

Enum markers are prohibited rather than required (#12, open). Writing
`+kubebuilder:validation:Enum` into a CRD freezes the list of accepted values, so GCP adding a
value would require a KCC release before anyone could use it. GCP already validates these values,
so the marker adds maintenance work without adding safety. Existing resources that have them are
grandfathered, and new ones may not add them.

---

## Open questions

Which golden files should become ratchets? The split is 17 to 2 today. `missingrefs.txt` was the
obvious conversion, because reference decisions are what bulk generation gets wrong. The other 16
have not been examined against the same question, which is whether work in flight can produce new
entries in them.

Does a flat manifest still work at scale? It has one entry today. At several hundred resources it
may be better to derive membership from properties the resource already has, such as being
`v1alpha1`, having a direct controller, and carrying the generated-types annotation, rather than
maintaining a list by hand.

External formats have a third source of truth that nothing checks. Each `<kind>_reference.go`
carries a doc comment of the form `Should be in the format "..."`. It is written independently of
the parser in the identity file and can disagree with it: `ManagedFolderRef` documents `locations/`
where its own parser requires `buckets/`. Checking the two against each other would be cheap to add
to the existing identity check.

Nothing makes the judgement queue drain. A resource could sit in `needs_judgement_call.txt`
indefinitely, and its `[refs]` findings would stay suppressed for as long as it did. Once bulk
generation is running at volume this probably needs either a limit on how long a resource can stay
queued, or a cap on how many can be queued at once.
