# Two attempts at bulk CRD generation: Claude and Gemini

**Claude wrote this, about work Claude did, comparing it to work Gemini did.** That is a conflict of
interest and no amount of care removes it. What can be done is to anchor every comparative claim to
resources both experiments measured, cite the Gemini material by section so a reader can check it
without trusting this summary, and record our own defects at the same resolution as theirs. Where a
number flatters this side for a methodological reason, the reason is in the same sentence.

Sources: Gemini's write-up is
[PR 12681](https://github.com/GoogleCloudPlatform/k8s-config-connector/pull/12681) — a design
proposal, a benchmark, a learnings log, and an 18,639-line JSON dataset — with its generator change
in [PR 12682](https://github.com/GoogleCloudPlatform/k8s-config-connector/pull/12682). This side is
the sandbox repo `ldanielmadariaga/kcc-lab`, PRs #21, #22 and #26.

## The two experiments are not the same size

| | Claude | Gemini |
|---|---|---|
| generator work | ~20 PRs over three weeks | 1 PR, 6 files, +179/-4 |
| corpus | 275 greenfield resources | 396 kinds across 122 services |
| method | delete upstream's types, regenerate, compare field by field | per-service sandbox: wipe, regenerate, isolated `controller-gen`, compare, revert |
| headline | 91.4% implemented | 45.9% repo-wide; 82.6% on the build-passing subset |

Effort and breadth run in opposite directions, and that is the single most important thing to hold
onto. Gemini's one change does roughly what PRs #15 and #17 did here: populate the root Spec and
ObservedState from the proto, and infer the parent from `(google.api.resource).pattern`. Reaching
that in a single change is a strong result, and comparing it to twenty PRs of accumulated fixing
without saying so would be dishonest.

## Like for like: the 267 kinds both measured

267 of this side's 275 resources also appear in Gemini's 396, so the comparison does not have to rest
on two headline numbers from different corpora.

The denominators do not match exactly. On the shared kinds Gemini counts 14,041 baseline properties
against this side's 11,734, because our `roots()` collapses a missing parent's children into one
defect rather than counting each. So the percentages below are indicative, not exact. They are
closest, within 7%, on the subset where both generated successfully, which is also the subset that
matters most.

**Where both generated successfully — 197 kinds:**

| | surface | matched | rate |
|---|---|---|---|
| Gemini | 8,985 | 7,769 | **86.5%** |
| Claude | 8,389 | 7,675 | **91.5%** |

Five points apart, on denominators within 7% of each other. That is the fairest single comparison
available, and it is closer than either headline suggests.

**Where Gemini's build failed — 70 kinds:**

| | surface | matched | rate |
|---|---|---|---|
| Gemini | 5,056 | 0 | **0%**, scored as build-failed |
| Claude | 3,345 | 3,048 | **91.1%** |

This is where the twenty PRs show up. On resources Gemini could not generate at all, generation here
runs at its ordinary rate. Gemini's benchmark scores those as zero, which is the honest choice, and
it is what drags 86.5% down to 55.3% across the shared set.

Crucially, its own learnings doc diagnoses why. Of its 23 failing services, 7 fail on a missing
`apiextensionsv1` import and 8 on unresolved slice types, and it estimates fixing both raises the
build pass rate from 81.1% to 93.4%. The first of those is a bug this side had already fixed in
`generatorbase.go`. **Gemini's build failures partly measure bugs we had closed, not a difference in
what the model could do.**

## Both experiments mis-scoped their corpus, in opposite directions

This is the least flattering section for both sides and the most useful.

**Gemini included resources its own design doc excludes.** Section 5.1 of
`bulk-deterministic-crd-generation.md` states the taxonomy correctly: exclude "Dual-Mode & Legacy: 78
Brownfield resources (dual Direct + TF/DCL) and 226 Legacy resources". Its benchmark then evaluated
396 kinds of which **126 are dual-mode or legacy** by `static_config.go`. `AlloyDBCluster`,
`ArtifactRegistryRepository` and `BigtableTable` all carry Terraform or DCL controllers alongside
Direct, and all three appear in its five-service deep dive. The taxonomy is right in the design and
was not applied in the measurement. Those resources are years of hand-curation, so including them
depresses its parity for a reason that has nothing to do with the generator.

**We omitted 44 resources that belonged.** Our corpus was hand-maintained and had drifted. Applying
the criteria properly — no Terraform or DCL controller, declared in a `generate.sh`, not
`--skip-scaffold-files`, has a baseline CRD — gives 275, not 231. The 44 additions score near 64%
where the original set scored 94%, and correcting it moved our headline from 94.2% to **91.4%**.
Nothing about the generator changed between those two runs.

So the corrected 91.4% is the number this comparison uses, and the earlier 94.2% that appeared in our
report and PRs was an artefact of an incomplete corpus. Both runs are kept in
[experiments/measurements/](experiments/measurements/).

Notably, Gemini's doc is what prompted the audit. Its "Class B: In-Flight / Unregistered Greenfield"
category is the thing our first corrected filter got wrong, and reading it is what caught a rule that
would have dropped 96 resources.

## Where Gemini is ahead

**Breadth.** 122 services and 396 kinds against 275 resources, from one change. It measured the whole
repository; we measured a curated subset.

**Failure diagnosis.** Its build failures are grouped by root cause with named services and a
quantified remediation estimate. Ours says 22 packages do not compile and 16 are a floor we chose not
to chase, which is less actionable.

**Presentation.** The parity-tier distribution — Tier 1 at 100%, Tier 2 at 80–99.9%, and so on, with
kind counts — conveys the shape of the result better than our single percentage. 74.6% of its kinds
reach ≥80% parity, which is a more useful sentence for a reader deciding whether to fund the work.

**A machine-readable dataset.** `benchmark-results.json` is per-kind and reusable. Ours lives in a
text report; this comparison had to parse theirs and re-derive ours.

**One modelling insight we lack.** Its section 7.4 notes that a single KRM `SecretRef` replaces two
or three proto credential strings, so simple property-name matching undercounts coverage. Our pairing
does not model those atomic substitutions.

## Where this side is ahead

**A flagging mechanism, not just an intention.** Both experiments conclude that references and
secrets are the irreducible judgement calls. Gemini's design says to "flag remaining ambiguous
strings for AI agent / developer review", but there is no artifact and no enforcement. Here there is
a per-service judgement queue, a `+kcc:guess` marker in the generated Go, and a machine check that
every marker has a queue entry. It currently reports 301 markers and **zero** without an entry, and
it caught three separate regressions that review would not have. This is the clearest capability
difference in the comparison, though it is a difference in how far the work went rather than proof
about the models.

**Detection measured rather than asserted.** Rules are scored against what upstream actually did and
rejected on the evidence — a `Certificate`-suffix rule scored 0 useful hits against 6 false ones and
was dropped. The sibling-resource rule runs at 77% precision, measured. Gemini proposes "multi-signal
heuristics" without precision numbers.

**A specific, large detection finding.** The output-only detector tested for a comment opening
`Output only.`; Compute writes `[Output Only]`, which is 1,605 fields in `compute.proto` alone.

**Two of Gemini's Phase 1 roadmap items are already done here**, including the `apiextensionsv1`
auto-import it estimates would unblock 7 services.

## What both found independently

This is the strongest signal in the comparison, and it points at the design rather than at either
model.

- **The same architecture.** Deterministic generation into `types.generated.go`, hand or agent
  overrides in `<kind>_types.go`, and the generator skipping what the override file already declares.
  Gemini explicitly considered and rejected a YAML overlay DSL as over-engineering; no overlay
  mechanism was built here either.
- **The same irreducible core.** References and secrets are what a generator cannot decide.
- **The same finding about proto annotations.** Gemini measured `google.api.resource_reference` as
  present in under 10% of APIs; we measured 11 of 111 references on a 239-resource run. Both
  concluded it is authoritative where present and useless as a primary signal.
- **The same measurement traps**, listed below.

## Shortcomings, ours alongside theirs

Paired deliberately. Most of Gemini's logged failures have a direct counterpart here.

| trap | Gemini | Claude |
|---|---|---|
| global `controller-gen` aborts, comparison reads stale files | its "99.66% fidelity illusion", §7.5: one failing package aborted generation and the harness compared baseline against baseline | same failure; `dev/tasks/greenfield-regenerate` runs `SKIP_GENERATE_CRDS=1` then per-service `controller-gen` for exactly this reason |
| numbers reported before the harness existed | §5, "failed to plan ahead on how to test and validate results before producing/reporting accuracy numbers" | arithmetic double-counted 16 fields (148 vs 132); a detection rule was measured on one population and implemented against another, firing on nothing; a healthy 40-minute regeneration was killed on a misread |
| over-aggressive wipe destroys hand-written files | §7.2, `rm -f apis/{svc}/v*/*types*` deleted custom non-proto types | two resources lost to `--skip-scaffold-files` in the 239-resource run before anyone noticed |
| conflating Direct controllers with greenfield | §7.1, cost it a drop from ~90% to 53.33% on a five-service run | the same class of error, in the other direction: 44 greenfield resources omitted, worth 2.8 points |
| upstream proto drift versus a static baseline | §7.3, new proto fields counted as discrepancies | handled, but only after the report double-counted and had to be rebuilt to count buckets rather than derive them |

Two further defects on this side with no Gemini counterpart: 331 files of intermediate
`controller-gen` output and a 12MB binary were committed by a careless `git add -A`, and the
published report carried 94.2% for several days on an under-scoped corpus.

## Verdict

**On the evidence available the two models look close on this task, and the visible difference tracks
how much work each was given rather than capability.** Where both generated successfully, 86.5%
against 91.5% is a five-point gap after roughly twenty times the generator effort on this side.
Gemini reached the same architecture, the same irreducible core, and the same measurement traps
independently and in one change.

The larger headline gap is almost entirely build failures, and its own doc attributes those to two
named bugs it estimates would take the pass rate from 81.1% to 93.4%. One of them was already fixed
here.

Where a real difference does show is in what was built on top of generation: a flagging mechanism
with an enforced invariant, and detection rules with measured precision. That is depth of work, and
this experiment had twenty times more of it.

**What would actually answer the capability question**, and this does not: the same corpus, the same
harness, the same number of iterations, and ideally a third party running both. Until then the
defensible claim is that both models produced sound designs that agree with each other, and neither
result argues for choosing one over the other on this task.

## Related

* [greenfield-experiment-report.md](greenfield-experiment-report.md) — the Claude-side write-up
* [greenfield-state-of-play.md](greenfield-state-of-play.md) — current numbers
* [experiments/measurements/](experiments/measurements/) — both scored runs and what the rescope changed
