# State of play

Where the greenfield generation experiment stands, and what is left. Refine freely; this is meant to
be the page you read first, with the detail in the docs it links.

## What the experiment is

We delete resources upstream implemented by hand, regenerate them from the proto, and compare.
Upstream's version is a known-good answer produced by a person, so it shows where mechanical
generation falls short. The point is to prove bulk generation is sound **before** pointing it at the
~550 greenfield resources that have no implementation at all.

Three things had to hold first, and this page reports on each:

1. generation as mechanical as possible,
2. everything needing a judgement call flagged,
3. coverage roughly equal to upstream.

## 1. Coverage — met

All 231 in-scope resources generate both a types file and a CRD. Measured against baseline
`c1df0b9326`:

```
  1. implemented                      10232   (94.2%)   same field, same path
  2. discrepancy                        368   (3.4%)    we produce it, but not as upstream has it
        flagged for a second pass         280   (76%)
        nothing says so                    88
  3. missing                            257   (2.4%)    we produce nothing at all
        a gap to close                    195
        we model it differently on purpose 62
        flagged for a second pass          39   (15%)
```

The split is on what we produced, not on whether we mentioned it. Those turn out to be close to
independent: three quarters of the discrepancies carry a judgement-queue entry, against one in seven
of the absences. Leading with "flagged" hid that, and made a 306-field "missed" bucket read as
absence when 88 of it was a field we do emit in a shape upstream does not have.

The 62 we model differently on purpose are the `google.protobuf.Value` union arms. We map `Value`
whole to `apiextensionsv1.JSON`, so the individual arms cannot exist as fields. Counting them as
absences would overstate the gap.

195 is the number to drive down. It is what we produce nowhere, less what we decline to produce
deliberately.

Run it with `hack/tools/greenfield/silence_report.py`; see
[greenfield-coverage-invariant.md](greenfield-coverage-invariant.md) for what each state means, and
[greenfield-detection-gaps.md](greenfield-detection-gaps.md) for how the mislabelling arose.

**Do not compare totals across corpus sizes.** An earlier measurement covered 189 resources rather
than 231, because 42 generated nothing at all; every absolute number moved with the denominator.

## 2. Flagging — met, and enforced

295 `+kcc:guess` markers, **every one with a judgement-queue entry**, checked by
`hack/tools/greenfield/check_guess_entries.py` as the last step of `dev/tasks/greenfield-regenerate`.

The rule is: anything the generator marks as a guess needs a human, so it belongs in the queue —
typed references included, since a target inferred from a plural noun is not something a generator
can be sure of.

Getting to zero found two bugs nothing else could see:

* the parent generation recorded emitted segments and suppressed their queue entries, so
  `BigtableCluster` carried `Instance *string` where upstream has `spec.instanceRef` and nothing said
  so — the compiler proved 32 fields missing while the queue named 2;
* `writeJudgementQueue` used `os.WriteFile`, which truncates. The queue is per service but
  `generate.sh` calls `generate-types` once per proto version — `dialogflow` four times — so every
  call after the first discarded the previous ones' entries. That had been true from the start.

## 3. Mechanical generation — mostly

23 packages fail to compile. **The tree does not build by design** — hand-written `_identity.go` and
`_reference.go` files are left at upstream's version while the types under them are regenerated,
which makes them the strictest oracle available. See
[greenfield-detection-gaps.md](greenfield-detection-gaps.md#the-tree-does-not-build-and-that-is-the-measurement).

| | packages |
|---|---|
| fail **only** on fields the generator cannot produce | **16** |
| have a repairable cause | 7 |

The 16 are the measurement, naming 26 distinct fields the hand-written controller code needs. Do not
"fix" them.

### The 7 repairable, and why they are parked

| packages | cause |
|---|---|
| `notebooks`, `contentwarehouse` | one invocation spans two proto versions and both declare `ContainerImage` / `Document`. Needs a naming scheme that disambiguates a message across versions — a brand-new resource never spans versions, so this blocks nothing for greenfield |
| `networksecurity` | a malformed generated `zz_generated.deepcopy.go`; a syntax error, so something emitted invalid Go |
| `backupdr`, `clouddeploy` | `v1alpha1.Parent` undefined, and a `DeliveryPipelineRef` whose ref type does not exist anywhere |
| `compute`, `discoveryengine` | residue after the pointer-ness and stale-file fixes |

### A caveat on the compile count

`inscope.tsv` wipes 231 resources, but the `generate.sh` scripts produce **409**. The other 178 keep
their upstream types files and skip regeneration entirely — `TypeFileExists` short-circuits, which is
what made `DiscoveryEngineSearchEngine` look like a broken GVK check when it had simply never been
regenerated.

That is correct behaviour: those resources are outside the experiment and wiping them would destroy
upstream work. But it means **the compile count includes resources we are not testing**, and a stale
one can fail a package that also holds corpus resources. Read the count with that in mind.

## The 123, by cause

| cause | n |
|---|---|
| absent, nested spec | 32 |
| reference nothing detects, nested spec | 26 |
| absent, nested status | 19 |
| absent, top-level spec | 18 |
| absent, top-level status | 15 |
| reference nothing detects, top-level spec | 13 |

Two thirds are nested, which is the shape of the remaining work: the detectors that generalise read a
field's own description or its name against the service's resources, and both get weaker the deeper a
field sits, because there is less context around it. No single Kind dominates — the largest is
`NetworkSecurityAuthzPolicy` at 10, and the tail is long.

References are 39 of the 123 and remain the largest coherent class. Their leaf names are long-tailed:
38 distinct names across the reference misses, 24 of which appear exactly once, so a lookup table
would have to grow one entry per field. That is the case for detectors that derive rather than
remember. See the runbook's four-signal table.

## Two measures, and which to trust

**Count undefined fields, not failing packages.** A package fails whole on a single type mismatch, so
the package count moves far slower than real progress — one regeneration took distinct fields 29 → 20
while packages went 44 → 43.

**The compiler covers 18% of what the CRD comparison finds** (106 of 583 at the time it was measured),
across 44 of 189 resources, because it only sees fields the controller code dereferences. It is a fast
first check, not a replacement. And it exists only because upstream implemented these resources: for
the ~550 with no implementation there is no `_identity.go` to compile against, which is why the CRD
comparison is the method that transfers.

## What I would do next

1. **Generate a batch of genuinely greenfield resources.** The prerequisites are met, and nothing in
   the remaining packages blocks it.
2. The 123, starting with the nested references — 26 of them, and the largest coherent group.
3. The 30 misplaced fields, which need placement rules rather than detection.
4. The parked packages, if they start costing time.

## Related

* [greenfield-experiment-report.md](greenfield-experiment-report.md) — the write-up for leads: what
  the experiment proves, what it does not, and the decisions we want made
* [greenfield-coverage-strategy.md](greenfield-coverage-strategy.md) — the acceptance bar and sequencing
* [greenfield-coverage-invariant.md](greenfield-coverage-invariant.md) — what the three states mean
* [greenfield-detection-gaps.md](greenfield-detection-gaps.md) — what the missed fields are made of
* [greenfield-step1-workflow.md](greenfield-step1-workflow.md) — the runbook, including the flag table
* [greenfield-reference-generation.md](greenfield-reference-generation.md) — deferred follow-up
