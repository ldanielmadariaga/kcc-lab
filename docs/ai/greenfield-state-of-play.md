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

All 231 in-scope resources now generate both a types file and a CRD. That is the number to lead
with: the previous measurement covered 189, because 42 produced nothing at all, and a resource that
generates nothing cannot be scored.

Measured over 231 resources against baseline `c1df0b9326`:

```
  1. implemented                      10176   (93.9%)
  2. flagged                            298   (2.7%)
        named in needs_judgement_call.txt               298
        ...and also in the types file                   22
  3. missed                             368   (3.4%)
        truly missed                         232   (2.1%)
        emitted, renamed or reshaped         102   (0.9%)
        reference, name unpairable            34   (0.3%)
```

232 is the gap. The other 136 are not absences: 102 we emit under a different name or shape, and 34
are references upstream renamed rather than suffixed. `DatastreamPrivateConnection`'s `vpc` is
upstream's `networkRef`, and no name match bridges those, so they are counted as missed, which
overstates rather than flatters.

**Do not compare these totals to the 189-resource ones.** The denominator grew by 22%, so every
absolute number grew with it. By cause, `reference-shape` misses went 84 → 96 over those 42 extra
resources, which is close to flat per resource, while `absent` went 34 → 125 — nearly all of that
jump is in the newly generating resources, and it is the thing to look at next. A same-set
comparison is not recoverable: the earlier run's `--verbose-dir` output was not kept.

Run it with `hack/tools/greenfield/silence_report.py`; see
[greenfield-coverage-invariant.md](greenfield-coverage-invariant.md) for what each state means.

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

## The 148, by cause

| cause | n |
|---|---|
| nested reference, nothing detects it | 36 |
| top-level reference, nothing detects it | 27 |
| output-only per upstream, proto never said so | 16 |
| resource never went through the pipeline | 13 |
| no mechanical explanation | 12 |
| deep nested / recursive message | 9 |
| parent segment, proto has no `google.api.resource` | 7 |
| inside a map value | 2 |

References are 63 of 148 and remain the largest single class. The 12 with no mechanical explanation
are the honest floor of the current pass; nine more turned out to be fields upstream invented with no
proto counterpart, which no proto-driven generator could ever produce — see
[greenfield-detection-gaps.md](greenfield-detection-gaps.md#the-ceiling-fields-no-generator-could-produce).

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
   the remaining 7 packages blocks it.
2. The 148, starting with references — 63 of them, and the largest coherent group.
3. The parked packages, if they start costing time.

## Related

* [greenfield-coverage-strategy.md](greenfield-coverage-strategy.md) — the acceptance bar and sequencing
* [greenfield-coverage-invariant.md](greenfield-coverage-invariant.md) — what the three states mean
* [greenfield-detection-gaps.md](greenfield-detection-gaps.md) — what the missed fields are made of
* [greenfield-step1-workflow.md](greenfield-step1-workflow.md) — the runbook, including the flag table
* [greenfield-reference-generation.md](greenfield-reference-generation.md) — deferred follow-up
