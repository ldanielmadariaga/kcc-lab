# Greenfield quality enforcement: how the checks are designed

**Scope:** the experimental sandbox (`kcc-lab`), not upstream policy. **Status:** describes what is
built and what is in review; open questions at the end.

The other documents cover *what* to build and *how* to build it:

| Document | Answers |
|---|---|
| `greenfield-coverage-strategy.md` | which resources, in what order, to what bar |
| `greenfield-tracker.md` | how to pick the next piece of work |
| `greenfield-bulk-generation.md` | the per-resource Step 1 procedure |
| `greenfield-generator-mechanics.md` | changes to the generator so the procedure has less to do |

This one covers the other half: **how quality is held and improved once the resources exist.**

Some of what follows is merged and some is still in review, so each mechanism is marked:

| Mechanism | PR | Status |
|---|---|---|
| Manifest scoping | #5 | merged |
| Dropped-field ratchet | #6 | merged |
| Unit tests for the checkers | #7 | merged |
| Reference detector, `missingrefs` as ratchet, `refs_deferred` | #9 | merged |
| GCS ref taxonomy refinements | #10 | open |
| Further skill checks | #11 | open |
| Step 1 scoping, enum prohibition, nested dropped fields | #12 | open |
| Judgement queue and `[refs]` suppression | #16 | open |
| Identity collection casing | #18 | open |

Where a section describes something unmerged, it says so.

---

## The problem this solves

The strategy is deliberately to generate many resources at low quality and improve them in
corpus-wide passes. That creates two opposite failure modes, and the whole design is the tension
between them:

1. **Absorption.** A new resource's defects get written into an exception list and nobody notices.
   The list grows, the bar erodes, and "we have a check for that" becomes false.
2. **Retroactive failure.** A new check fires on 450+ resources that predate it, so the only way to
   land it is to weaken it or to grandfather everything — which is absorption wearing a different hat.

Every decision below is a choice about where to sit between those two.

---

## 1. Two baseline semantics

The single most important distinction in this codebase, and it is invisible from the filenames —
both write a file under `testdata/exceptions/`.

| | `CompareGoldenFile` | `CompareRatchetFile` |
|---|---|---|
| New violation appears | absorbed on `WRITE_GOLDEN_OUTPUT=1` | **fails, even with the flag set** |
| Violation fixed | removed on rewrite | pruned, and logged |
| Direction | any | shrinks only |
| Describes | the corpus as inherited | work we are doing now |

Current split on master: **17 goldens, 2 ratchets** — `missingrefs.txt` and
`greenfield_dropped_fields.txt`. A third, `identity_collection_casing.txt`, is pending in #18.

The goldens are not a failure of nerve. A golden is the right tool for a finding class that is large,
pre-existing, and not what anyone is working on this quarter — it records the state and makes growth
visible in review. A ratchet is for a class where new entries mean someone did the work wrong *today*.

**The rule:** if a finding could be produced by work in flight, it belongs in a ratchet. Otherwise a
golden is honest and cheaper.

`missingrefs.txt` was converted from golden to ratchet, because reference decisions are exactly the
thing bulk generation gets wrong.

---

## 2. Scoping: the manifest

`tests/apichecks/testdata/greenfield_bulk.txt` lists the Kinds held to the new bar:

```
networkservices.cnrm.cloud.google.com/NetworkServicesLBTrafficExtension
```

A check that applies to everything can only be as strict as the *worst* existing resource. Scoping to
an explicit manifest lets a new check be strict from day one without touching anything that predates
it — the resource opts in when it is generated, and never opts out.

The Kind alone is enough to locate every file, because `TestDirectResourceFileNaming` already
requires files under `apis/` and `pkg/controller/direct/` to be prefixed with the lowercased Kind.

**Cost, stated plainly:** the manifest is a second thing to keep in sync, and a resource missing from
it is silently unchecked. `TestGreenfieldBulkManifestIsResolvable` exists for exactly that — every
manifest entry must resolve to a real CRD and real files, so the failure mode is a loud error rather
than silent non-coverage.

---

## 3. Attribution

Every finding carries the resource it belongs to (`crd=`, or `kind=`/`group=`):

```
[refs] crd=alloydbbackups.alloydb.cnrm.cloud.google.com version=v1alpha1: field ".spec.encryptionConfig.kmsKeyName" should be a reference
```

Without it, a baseline is a wall of text nobody can act on, and pruning per resource is impossible.
With it, "what does this resource still owe?" is a grep.

---

## 4. The state machine for reference findings

References are the hardest judgement in Step 1 and the one most likely to be wrong, so they get four
states rather than a boolean. **A resource is in exactly one**, which is what keeps the files from
needing reconciliation:

| State | File | Written by | Kind |
|---|---|---|---|
| not yet triaged *(pending, #16)* | `apis/<service>/needs_judgement_call.txt` | generator | queue, per service |
| triaged: deferred, with reason | `refs_deferred.txt` | human | curated input, subtracted |
| triaged: still owed | `missingrefs.txt` | computed | ratchet |
| structurally impossible | `refs_not_representable.txt` | computed | golden |

Three things about this are non-obvious and were each learned the hard way.

**`missingrefs.txt` cannot hold a reason.** It is recomputed from the CRDs every run, so anything a
human writes into it is erased. `refs_deferred.txt` is the ledger — hand-edited, reason mandatory,
subtracted from the findings before the ratchet applies.

**The queue exists to make the mechanical pass mergeable** *(#16, open).* A generated resource has ref-shaped
string fields by construction; against a ratchet, that means the first bulk PR cannot land. While a
resource has open queue entries, its `[refs]` findings are suppressed.

**Suppression must not look like a fix.** A queued resource contributes no findings, so anything it
already owed reads as *removed*, gets pruned from the ratchet, and reappears as a **new** violation
the moment it graduates — failing the check for work nobody did. Baseline entries for suppressed
resources are carried forward, so queueing only ever stops findings being *added*.

`refs_not_representable.txt` is a golden rather than a ratchet on purpose: those entries are
un-actionable by construction, so the list is *expected* to grow as resources are added. Growth is
reviewable in the diff, and each entry states its reason.

---

## 5. Detection: when to trust an annotation

Proto annotations look like ground truth and are not uniformly usable. The deciding question is not
coverage, it is **what absence means**:

| Signal | Coverage | Absence means | Usable as a check? |
|---|---:|---|---|
| `google.api.field_behavior` | 40.7% | assume optional — already the behaviour | yes, fallback is inert |
| `google.api.resource_reference` | ~15%, 0% in compute | nothing can be concluded | no, fallback must guess |

Matching `resource_reference` by field name was implemented and measured: **2,164 findings against
78** for description heuristics alone. A name like `network` is annotated in one service and appears
in hundreds of unrelated fields elsewhere. It was removed rather than tuned.

The blocker is structural: a CRD does not record which proto field it came from, so the leaf name is
the only bridge at that layer and it is too weak. Using the annotation properly means mapping CRD
field → proto field via the `+kcc:proto:field=` markers the generator emits, then looking up the
fully-qualified path. That is a different change, and it would not reuse a leaf-name list.

**The rule:** use an annotation as a check when its absence has a safe default. Otherwise it is a
generator input, not a checker input.

---

## 6. Choosing an oracle

What a check compares against determines whether it can run at all:

| Oracle | Normative? | Committed? | Notes |
|---|---|---|---|
| proto descriptor (`.build/googleapis.pb`) | yes | **no** — gitignored | a check needing it silently skips in CI |
| recorded traffic (`_http.log`) | close | yes, 894 files / 23 MB | contains KCC's requests as well as GCP's responses |
| CRDs | no | yes | derived from KRM; cannot answer GCP-format questions |

Worked example — `TestIdentityCollectionCasing` *(#18, open)*. The question is whether the collection segment an
identity builds matches what GCP uses. The proto pattern is the normative answer, but a proto-based
check would skip in CI, so it uses recorded traffic instead, with the limitation written into the
test: it catches a wrong casing only when *some other part of KCC* gets it right. A resource wrong
everywhere would slip through, and resources without fixtures are not covered — which is why its
baseline lists 4 where a proto-based audit finds 9.

Trading normative strength for a check that actually runs is usually right, but the compromise
belongs in a comment, not in the reader's head.

---

## 7. The checkers are code, so they get tests

Detection logic has bugs like anything else, and a checker's bugs are invisible: it reports fewer
findings and everything looks fine.

Two real cases from this work:

- An entry-key parser silently produced `"foos.example.com:"` with a trailing colon, breaking every
  subsequent match. Found by a unit test, not by the corpus.
- `TestIdentityCollectionCasing`'s regex, if not anchored on `parent.String()`, matches the
  `/locations/` inside the *Parent's* `String()`. Every resource then appears to use `locations`,
  whose casing is correct everywhere — so the check **passes while verifying nothing**. Guarded by
  `TestIdentityCollectionRegex`, which asserts the regex extracts the collection and not the parent.

**Silent weakening is the characteristic failure of this kind of code**, so tests target the
extraction and matching primitives, not just the end-to-end result.

---

## 8. What is deliberately not enforced

- **Field coverage** (`TestGreenfieldBulkFieldCoverage`) is local-only, behind `GREENFIELD_STRICT=1`.
  Step 1 legitimately fails it, because coverage is measured against test fixtures and Step 1 has
  none. Making it a CI gate would mean either blocking Step 1 or weakening the check.
- **`alpha-missingfields.txt` is expected to grow** during Step 1, for the same reason. Entries are
  attributed by `crd=` and removed when fixtures arrive.
- **Enum markers are prohibited, not required** *(#12, open).* `+kubebuilder:validation:Enum` bakes a value list
  into a CRD, so GCP adding a value needs a KCC release. GCP already validates. Existing resources
  are grandfathered; new ones may not add them.

---

## Open questions

- **Ratchet coverage.** 17 goldens to 3 ratchets. Which of the 17 describe classes that bulk work can
  regress, and should therefore be converted? `missingrefs.txt` was the obvious one; the rest are
  unexamined.
- **The manifest as a bottleneck.** One entry today. At 400 resources, is a flat file still the right
  scoping mechanism, or should membership be derived — for example from the `v1alpha1` + direct
  controller + generated-annotation combination?
- **A third source of truth for external formats.** `<kind>_reference.go` carries a
  `Should be in the format "..."` doc comment that is independent of the parser and can disagree with
  it — `ManagedFolderRef` documents `locations/` where the parser requires `buckets/`. Nothing checks
  it. Cheap to add to the identity check as a second assertion.
- **Enforcing the queue drains.** Nothing currently prevents a resource sitting in
  `needs_judgement_call.txt` forever, which would make suppression permanent. A staleness bound, or a
  cap on queued resources, may be needed once bulk generation is running.
