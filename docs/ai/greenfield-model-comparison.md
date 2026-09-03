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

Effort and breadth run in opposite directions. Gemini's one change does roughly what PRs #15 and #17
did here: populate the root Spec and ObservedState from the proto, and infer the parent from
`(google.api.resource).pattern`.

The effort ratio is real, but it is not a like-for-like measure of autonomy, and reading it as
"reached the same design in one change" would be wrong. Gemini's starting design was steered by
someone carrying results from this experiment, and several of the conclusions that look like
agreement between the two were handed over rather than found. That is set out below under what was
independently found. Neither number in the table is a measure of what either model would do
unassisted.

## Like for like: one harness, one baseline, one denominator

The first version of this document compared Gemini's self-reported `exact_matches` against our
`implemented`. Those are different definitions computed by different code, so the comparison was not
worth much. Gemini has since published its generated CRDs in
[PR 12737](https://github.com/GoogleCloudPlatform/k8s-config-connector/pull/12737), which lets our
harness score both sides.

Everything below is controlled: the same scorer, the same baseline (`25aedf2f10ef`, Gemini's), the
same kinds, and a denominator that is a property of upstream rather than of either generator. The
judgement queue is switched off with `--no-queue`, so neither side is credited with our flagging.

**The denominator is `matched + missing + mismatch`** — the count of fields in upstream's CRD. Our
own report uses a different one, which collapses every child of a missing parent into a single
defect. That is the right unit for a work list and useless for comparing generators: a side that
misses one fifty-field subtree would score one defect where a side missing twenty scattered leaves
scores twenty. On this corpus that choice reverses the ordering, so the comparison uses the field
count and `compare_generators.py` refuses to report if the two sides' denominators disagree by more
than 2%.

| | kinds | Claude | Gemini |
|---|---|---|---|
| **A. all versions** | 255 | **78.3%** | 69.6% |
| **B. v1alpha1 only** | 217 | **79.4%** | 70.4% |

Both arms have exactly matching denominators. Both put the gap at about nine points.

Test B exists because Gemini's corpus is 40% v1beta1 against our 15%, and beta resources are the
mature hand-curated ones a generator does worst against, so Test A looked structurally unkind to it.
Restricting to alpha moves each side by about half a point, so that hypothesis was wrong and the
composition difference is not what separates them.

12 of our resources are excluded from both arms. Their types regenerated but `controller-gen` then
failed for the service, leaving the previously published CRD in place, so scoring them would have
credited us with upstream's own output. Removing them costs about a third of a point.

### The shape behind the aggregate

Nine points sounds modest against spot checks that showed us reproducing 269 of a resource's 333
fields where Gemini managed 64. Both are true, and the win/loss record explains why.

| | kinds | fields | Claude | Gemini |
|---|---|---|---|---|
| identical outcome | 141 | 4,959 | 73.8% | 73.8% |
| we produce more | 76 | 6,442 | **83.7%** | 67.7% |
| Gemini produces more | **0** | — | — | — |

**We do not lose a single kind.** An earlier version of this document said we lose on small flat
resources; that came from `GKEHubFleet` at 28/33 against 33/33, which turned out to be one of the 12
resources whose CRD had gone stale. With those excluded there are no losses at all.

The gap is concentrated rather than broad: the five largest wins account for 54% of the entire field
advantage, and the top twenty for 81%. On 141 of 217 kinds the two generators produce the same
outcome.

The mechanism is depth. Our generator walks the whole proto message tree and writes every nested
message into `types.generated.go`. Gemini's single change populates the root Spec from the root
message and does not recurse as far, so on a deeply nested resource the top-level fields appear and
the subtrees under them do not.

### Why neither side reaches 100%

On the 141 tied kinds both sit at exactly 73.8%, which means they are missing the same fields. Those
are the domain decisions no proto-driven generator makes: a raw string that upstream models as a
`*Ref`, a credential that becomes a `SecretRef`, a parent that is another resource rather than a
project. Both experiments reached that conclusion independently, and the tie rate is the measurement
of it.

So depth is why one generator beats the other, and the human-judgement layer is why neither gets
close to the baseline.

## What only one side reports

Coverage is not the only axis. A generator can also record what it could not decide, and that
changes what a reviewer has to discover for themselves.

Same 217 kinds, same harness, same baseline, with our judgement queue switched off for Gemini's
output because the queue is keyed by Kind and would otherwise credit its fields with our flagging.

| | reproduced exactly | diverges, flagged for a human | diverges, nobody is told |
|---|---|---|---|
| **Claude** | 79.4% | **9.4%** | **11.2%** |
| **Gemini** | 70.4% | 0.0% | **29.6%** |

Denominators agree to 0.01%.

**A reviewer is left unaware of 11.2% of the surface with our approach against 29.6% with Gemini's**,
a factor of about 2.6. We reproduce nine points more, and of the divergence that remains we hand
nearly half to a person: a `+kcc:guess` marker in the generated Go and an entry in the service's
judgement queue naming the field, with a machine check that no marker exists without one.

Two things this does not say. **Gemini's zero is a property of what it built, not a failure.** It set
out to generate, and its design proposes flagging ambiguous strings for agent review without building
the artifact. Flagging is a separable capability that neither model was asked for, and one of the two
experiments built it. And **our 9.4% is a self-assessment**: it says we told someone, not that what
we told them was right. The sibling rule's measured 77% precision is the only evidence on that
question, and it covers one detector out of four.

## Both experiments mis-scoped their corpus, in opposite directions

This is the least flattering section for both sides and the most useful.

### Gemini included resources its own design doc excludes

Section 5.1 of
`bulk-deterministic-crd-generation.md` states the taxonomy correctly: exclude "Dual-Mode & Legacy: 78
Brownfield resources (dual Direct + TF/DCL) and 226 Legacy resources". Its benchmark then evaluated
396 kinds of which **126 are dual-mode or legacy** by `static_config.go`. `AlloyDBCluster`,
`ArtifactRegistryRepository` and `BigtableTable` all carry Terraform or DCL controllers alongside
Direct, and all three appear in its five-service deep dive. The taxonomy is right in the design and
was not applied in the measurement. Those resources are years of hand-curation, so including them
depresses its parity for a reason that has nothing to do with the generator.

### We omitted 44 resources that belonged

Our corpus was hand-maintained and had drifted. Applying
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

### Breadth

122 services and 396 kinds against 275 resources, from one change. It measured the whole
repository; we measured a curated subset.

### Failure diagnosis

Its build failures are grouped by root cause with named services and a
quantified remediation estimate. Ours says 22 packages do not compile and 16 are a floor we chose not
to chase, which is less actionable.

### Presentation

The parity-tier distribution — Tier 1 at 100%, Tier 2 at 80–99.9%, and so on, with
kind counts — conveys the shape of the result better than our single percentage. 74.6% of its kinds
reach ≥80% parity, which is a more useful sentence for a reader deciding whether to fund the work.

### A machine-readable dataset

`benchmark-results.json` is per-kind and reusable. Ours lives in a
text report; this comparison had to parse theirs and re-derive ours.

### One modelling insight we lack

Its section 7.4 notes that a single KRM `SecretRef` replaces two
or three proto credential strings, so simple property-name matching undercounts coverage. Our pairing
does not model those atomic substitutions.

### It documented its own corrections

`learnings-and-corrections.md` records, in its own words, four
proposals it got wrong and had to be talked out of. That is the reason the error in the first version
of this comparison was catchable at all: had the doc presented those conclusions as findings, the
claim that both models converged independently would have gone unchallenged. Writing down what you
proposed before someone corrected you is a good habit and this side has no equivalent artifact.

## Where this side is ahead

### A flagging mechanism, not just an intention

Both experiments conclude that references and
secrets are the irreducible judgement calls. Gemini's design says to "flag remaining ambiguous
strings for AI agent / developer review", but there is no artifact and no enforcement. Here there is
a per-service judgement queue, a `+kcc:guess` marker in the generated Go, and a machine check that
every marker has a queue entry. It currently reports 301 markers and **zero** without an entry, and
it caught three separate regressions that review would not have. This is the clearest capability
difference in the comparison, though it is a difference in how far the work went rather than proof
about the models.

### Detection measured rather than asserted

Rules are scored against what upstream actually did and
rejected on the evidence — a `Certificate`-suffix rule scored 0 useful hits against 6 false ones and
was dropped. The sibling-resource rule runs at 77% precision, measured. Gemini proposes "multi-signal
heuristics" without precision numbers.

### A specific, large detection finding

The output-only detector tested for a comment opening
`Output only.`; Compute writes `[Output Only]`, which is 1,605 fields in `compute.proto` alone.

**Two of Gemini's Phase 1 roadmap items are already done here**, including the `apiextensionsv1`
auto-import it estimates would unblock 7 services.

## What was found independently, and what was handed over

An earlier version of this document called the architectural agreement between the two experiments
"the strongest signal in the comparison". That was wrong, and the evidence against it is in Gemini's
own learnings doc, whose stated purpose is to record "critical course corrections **provided during**
the design and evaluation sessions". Sections 1 to 4 are each structured *Initial Flawed Proposal*
followed by *The Correction*:

| Gemini's section | what it proposed | what it was corrected to |
|---|---|---|
| §1 enums | Go string types with `+kubebuilder:validation:Enum` markers | unvalidated `*string`, no markers |
| §2 overlays | an "elaborate YAML overlay DSL, custom AST patch engines" | KCC's native Go override model |
| §3 annotations | assumed `google.api.resource_reference` was widely available | present in under 10% of APIs |
| §4 credentials | credential gaps as a category of their own | a subtype of resource reference |

The corrections came from the person running the experiment, carrying results from this one. So the
architectural agreement was **transmitted, not independent**. This document previously said "Gemini
explicitly considered and rejected a YAML overlay DSL as over-engineering", which inverts what
happened: it proposed the overlay DSL and was told not to build it. The same applies to the
`resource_reference` sparsity finding, which reads as a striking match between the two experiments
and was in fact passed on to save time.

### What is still genuine convergence

Its sections 5 and 7, the validation-planning failure and the evaluation-suite failures, are its own.
Nobody handed it those, and they land on the same traps this side hit unprompted:

- **A global `controller-gen` abort making a harness compare a baseline against itself.** Gemini's
  §7.5 calls it the "99.66% fidelity illusion". This side runs `SKIP_GENERATE_CRDS=1` and then
  per-service `controller-gen` for exactly that reason.
- **Reporting accuracy numbers before building the harness that justifies them.** Its §5 names this
  as a planning failure. The counterpart here is arithmetic that double-counted 16 fields.
- **A wipe destroying hand-written files it was never meant to touch.**
- **Conflating "has a Direct controller" with "is greenfield"**, which both experiments got wrong in
  opposite directions.

Two harnesses, built independently, hitting the same four traps is a real finding. It says the traps
are properties of the problem rather than of either model, and it is the part of this section that
survives.

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

Measured properly, our output reproduces about nine points more of upstream's CRD fields than
Gemini's: 79.4% against 70.4% on v1alpha1 resources, with identical denominators. That is a real
gap, and it is larger than the five points the first version of this document reported from Gemini's
own harness.

It is also not a like-for-like measure of the two models. Ours had roughly twenty pull requests of
generator work behind it against Gemini's one, and the mechanism behind most of the gap is a single
structural difference: our generator recurses through the whole proto message tree while Gemini's
change populates the root Spec. That is a scope difference in what was built, not evidence about
what either model could build.

What this comparison cannot tell you is which model is better at the task, and the reasons are
worth being explicit about.

- Gemini's architectural conclusions were largely handed to it, by someone carrying results from this
  experiment. Its design and this one agreeing is therefore not evidence about either model.
- Both sides have now been scored by one harness on one baseline, so the nine-point gap is real as
  measured. What it measures is one generator against another at a moment in time, not either
  model's ceiling.
- The effort differs by roughly twenty PRs, and the flagging mechanism and measured detection rules
  that this side has are products of that time rather than of the model.
- Both corpora were mis-scoped, in opposite directions, and both were corrected only after the fact.

**What would answer it**: the same corpus, the same harness, the same number of iterations, the same
amount of human steering, and ideally a third party running both. Short of that, the defensible claim
is narrow — both models produced designs a KCC engineer would recognise as sound, and nothing here
argues for choosing one over the other.

## Related

* [greenfield-experiment-report.md](greenfield-experiment-report.md) — the Claude-side write-up
* [greenfield-state-of-play.md](greenfield-state-of-play.md) — current numbers
* [experiments/measurements/](experiments/measurements/) — both scored runs and what the rescope changed
