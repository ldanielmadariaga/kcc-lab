# Bulk-generating greenfield resources: what we tried, and what it measured

*Everything here is reproducible from the sandbox repo
[ldanielmadariaga/kcc-lab](https://github.com/ldanielmadariaga/kcc-lab); the commands are named
throughout.*

> **There is a second, fuller version of this report.**
> [`greenfield-experiment-report.html`](greenfield-experiment-report.html) is the page published to
> claude.ai, and it has since moved ahead of this file: it is framed as a report rather than a
> proposal, and it carries an appendix on the prerequisites that made deterministic generation
> possible and on how the non-deterministic parts are flagged. Edit the HTML and republish it;
> treat this file as the plain-text summary.

## The proposal, up front

KCC manages 1,006 create-capable GCP resources' worth of API surface, and has direct controllers for
457 of them — **45.4%**. Reaching 80% means implementing about 348 more. At the rate a hand-written
resource takes, that is a multi-year queue, and the tail is made of resources nobody is asking for
loudly enough to prioritise individually.

We think most of that surface can be generated rather than written, and we have spent this
experiment finding out how much. The answer, measured rather than asserted: **a generator reproduces
94.2% of the CRD fields upstream wrote by hand, and 1.1% of the surface is produced nowhere and
mentioned nowhere.** The rest is accounted for — flagged for a human, or produced under a different
name or shape.

What we are asking for is a decision on whether to point this at the unimplemented resources, and
what quality bar to require when we do.

## Why we deleted upstream's work to test this

The obvious way to evaluate a generator is to run it on resources nobody has implemented and read
the output. That does not work: there is nothing to check the output against, and "looks plausible"
is not a measurement.

So we did the opposite. We took 231 resources the team had already implemented by hand, **deleted
our types files, regenerated them from the proto, and compared field by field against upstream's
CRDs.** Upstream's version is the known-good answer. Every place the generated CRD differs from it
is a specific, countable way that mechanical generation falls short — and because a real engineer
made each of those hand-written choices, the diff is a list of the judgements a generator cannot
make.

This is the part worth stealing regardless of what happens to the rest: **an existing implementation
is a test oracle, and deleting it is how you use it.** It also gives a second oracle for free. The
hand-written `_identity.go` and `_reference.go` files were left in place, so `go build ./apis/...`
fails wherever generated types no longer satisfy upstream's own controller code, naming missing
fields in seconds.

## Where it stands against the three things it had to prove

### 1. Coverage comparable to upstream — met

231 of 231 in-scope resources generate both a types file and a published CRD. Against baseline
`c1df0b9326`:

```
  1. implemented                      13195   (94.3%)   the same field at the same path
  2. discrepancy                        445   (3.2%)    we produce it, but not as upstream has it
        flagged for a second pass         281   (76%)
        nothing says so                    88
  3. missing                            357   (2.6%)    we produce nothing at all
        a gap to close                    295
        we model it differently on purpose 62
        flagged for a second pass          39   (6%)
                                     ------
  fields in KCC master CRDs           13997
```

295 fields is the gap that needs generating, 2.1% of the surface. The 368 discrepancies are fields
we do produce in a shape upstream does not have; they need detecting or moving. Reproduce with
`hack/tools/greenfield/silence_report.py`.

### 2. Everything needing judgement is flagged — met, and machine-checked

Anything the generator could not justify from the proto gets a `+kcc:guess` marker in the types file
**and** an entry in `apis/<service>/needs_judgement_call.txt`. Currently 299 markers, and
`check_guess_entries.py` enforces that **none** of them lacks an entry.

That check earns its keep. The invariant broke three separate times during this work and the checker
caught all three, in places code review would not have looked: a marker written on every generator
invocation while the queue entry was written on only some; a merge that silently dropped every
comment line of the existing queue file; and markers emitted into commented-out blocks describing
code the generator did not actually produce. A rule nothing checks is a rule that regresses.

### 3. Generation as mechanical as possible — mostly

The generator now derives from the proto what used to be hand-written: `+required` from
`field_behavior`, the resource's parent shape and location from `google.api.resource`, output-only
placement, and reference candidates from four independent signals. What remains genuinely manual is
recorded in the judgement queue rather than left to be discovered.

## The finding that changed how we work

Early on the instinct was to fix each coverage gap directly — see a missing field, write a rule that
produces it. That does not scale and it does not transfer: you cannot enumerate the ways a thousand
APIs differ, and a rule learned from one service usually misfires on another.

The rule we settled on is **detection over prescription**: it is more valuable to reliably *notice*
that a field needs a human than to guess what the human would say. A flagged field is a fine
outcome. A field nobody was told about is not.

That reframing gives a test for whether a new rule is worth having. Does it *derive* its answer from
something the API supplies, or does it *remember* something a person once noticed?

| signal | source | derives or remembers |
|---|---|---|
| `google.api.resource_reference` | the proto states it | it is a fact |
| description resource-name templates | the field's own docs | derives |
| sibling resource in the same service | the service's own resource list | derives |
| `refs.NameRules` | a list of known spellings | remembers |

The derived three work on a service nobody has looked at. The remembered one only ever finds
references somebody has already seen, which is why a growing `NameRules` list is a signal that one of
the other three is missing something, not a sign of progress.

**The sibling rule is the clearest example.** If a service declares a resource called `DataStore`,
then a string field named `dataStore` is probably a reference to it. That needs no vocabulary at all,
and it gets *stronger as more resources are generated*, because the service's own resource list is
what it reads. Measured against what upstream actually did: 77% precision. Nobody maintains it.

**And the payoff from derived signals can be very large for very little.** The output-only detector
tested whether a proto comment opened with `"Output only."`. Compute writes `[Output Only]` instead.
That one unrecognised spelling covers **1,605 fields in `compute.proto` alone**, and every one of
`ComputeInterconnect`'s misplaced status fields — `googleIPAddress`, `circuitInfos`,
`expectedOutages` — turned out to be nothing more exotic than that. Two strings in a list.

## Measurement discipline, because this metric is easy to cheat

A coverage report like this improves if you simply stop emitting a field and flag it instead. So the
report prints `implemented` beside the gap, and any change that moves the gap without holding
`implemented` steady is treated as suspect. One recent change took the not-generated count from 232 to
123 **and** `implemented` from 10,176 to 10,232, which is what a real improvement looks like.

Two further habits worth keeping:

**Measure a rule's precision before adding it, and write the number next to it.** Several plausible
rules were measured and rejected on the evidence — a rule for fields ending `Certificate` scored 0
useful hits against 6 false ones, because `pemCertificate` is a blob, not a reference.

**Report differences you cannot prove in their own column.** 34 fields are references upstream
renamed so thoroughly that no name match bridges them, so they are counted as missed. That
overstates our gap rather than flattering it, which is the right direction for a number we are using
to make a decision.

## What is built and reusable

* `dev/tools/controllerbuilder` — the generator, extended with proto-derived required fields, parent
  identity, output-only placement and reference detection
* `dev/tasks/greenfield-regenerate` — end-to-end regeneration of the corpus, including the queue
  re-seeding step that is easy to forget and silently discards findings when skipped
* `hack/tools/greenfield/silence_report.py` — the field-by-field comparison against upstream
* `hack/tools/greenfield/check_guess_entries.py` — the marker-implies-entry invariant
* `hack/tools/greenfield/sibling_precision.py` — scores a detector against what upstream did
* `docs/ai/greenfield-*.md` — the runbook, the strategy, the measurements, and the approaches that
  were tried and abandoned with the reasons

## Honest limitations

**The corpus is upstream's choices, not a random sample.** These 231 are resources the team judged
worth implementing. The ~550 unimplemented ones may be systematically harder — less documented, odder
shapes, or unimplemented precisely because someone looked and found a problem. We should expect the
94.2% to be an optimistic ceiling, not a forecast.

**22 packages do not compile, by design.** Hand-written `_identity.go` and `_reference.go` files were
left at upstream's version while the types beneath them were regenerated. That is the oracle working,
not damage. About 16 are a measurement floor we chose not to chase.

**First-pass output is not production quality.** References are generated as plain strings and
flagged, not resolved into typed refs. There are no test fixtures and no MockGCP coverage for
generated resources. The strategy is deliberately generate-first, retrofit in corpus-wide passes —
which is only safe because the sandbox has no users, and would need a different bar upstream.

**Two thirds of the remaining gap is nested fields.** The detectors that generalise read a field's
description or its name against the service's resources, and both weaken with depth because there is
less context around a deeply nested field.

## What we would like decided

1. **Is the 1.1% gap acceptable to start generating against?** If not, what would be — and measured how?
2. **What quality bar do generated resources need to merge upstream?** Specifically whether typed
   references, fixtures and MockGCP must land in the same change or can follow in passes.
3. **Is the judgement queue the right hand-off?** It currently assumes a human reviews each resource
   before it graduates. That is the throughput limit on the whole idea, and worth designing
   deliberately rather than inheriting.

## Where to read more

* [greenfield-state-of-play.md](greenfield-state-of-play.md) — current numbers and what is parked
* [greenfield-coverage-strategy.md](greenfield-coverage-strategy.md) — sequencing and the acceptance bar
* [greenfield-step1-workflow.md](greenfield-step1-workflow.md) — the per-resource runbook
* [greenfield-detection-gaps.md](greenfield-detection-gaps.md) — what the remaining gap is made of
* [greenfield-generator-findings.md](greenfield-generator-findings.md) — approaches tried and abandoned
