# Greenfield coverage strategy: conformance-first bulk implementation

**Status:** draft, iterating. **Scope:** the experimental sandbox (`kcc-lab`), not upstream policy.

This document is the acceptance bar and sequencing rule for closing KCC's greenfield resource
coverage gap. Agentic invocations working on bulk resource implementation should treat it as
normative and cite it in generated issues and PRs.

---

## 1. The target, and its denominator

Coverage is measured against **manageable** GCP resources — those with a `Create` or `Upsert` RPC.
Read-only, transient and singleton resources are excluded; most should not be CRDs at all.

From `hack/tools/greenfield/gap_analysis.txt` (snapshot 2026-07-31, googleapis `73aa1b6`):

| Metric | Value |
|---|---:|
| Total GCP resources (processed) | 1,724 |
| Implemented in KCC | 457 |
| Missing, manageable | 549 |
| Missing, full-lifecycle | 349 |
| **Manageable coverage** | **45.43%** |
| Total API coverage | 26.51% |

The manageable denominator is `457 + 549 = 1,006`. **80% of that is 805, so the gap is +348
resources** — which is almost exactly the 349 full-lifecycle missing set.

Measured against total API surface instead, 80% would be 1,379 resources, requiring ~573
implementations with no create RPC. That target is not meaningful and is explicitly not what this
document pursues.

## 2. Three constraints that determine the sequencing

### 2.1 Review throughput is the bottleneck, not generation

`hack/tools/greenfield/RESOURCE_STATUS.md` tracks 232 resources:

| MERGED | OPEN | PLANNED | CLOSED | RELEASED |
|---:|---:|---:|---:|---:|
| 38 | **153** | 38 | 3 | **2** |

171 sit at Phase 1. Generating more resources adds inventory to a queue that is not draining.
Any plan that increases generation rate without reducing per-resource review cost makes the
constraint worse.

### 2.2 The exception baselines absorb the debt they exist to catch

`tests/apichecks/testdata/exceptions/` holds 20 baseline files totalling ~8,900 lines. All but
`missingrefs.txt` and `refs_not_representable.txt` are compared with `test.CompareGoldenFile`,
which **rewrites the baseline whenever `WRITE_GOLDEN_OUTPUT` is set**. The standard
regenerate-goldens workflow therefore promotes new violations into the accepted-exceptions list,
and CI goes green on the next run.

Root-caused in
[issue #12344](https://github.com/GoogleCloudPlatform/k8s-config-connector/issues/12344): in
PR #11964, eighteen batch-generated Vertex AI resources shipped reference-shaped fields as plain
strings. Human review caught all of them; CI caught none.

**What this does and does not imply.** It is a real defect, and it matters at *retrofit* time: a
check that silently absorbs findings cannot tell you whether a retrofit pass actually finished. It
does **not** mean the baselines must be locked before generation, for two reasons:

- **Attribution already exists.** Every entry carries `crd=<name>` or `file=<path>`, so separating a
  newly generated resource's debt from the legacy mass is a `grep`, not archaeology.
- **Two of these files are the intended output of bulk generation.** `missingfields.txt` and
  `alpha-missingfields.txt` are 7,612 of the ~8,900 lines (**85%**) and record fields not exercised
  by test fixtures. Crude first-pass resources add to them by construction. Ratcheting them up front
  would mandate 100% field-level test coverage per resource — the opposite of the strategy.

So the enforcement mechanism is real and needed, but it belongs **after** each retrofit pass, not
before generation. See §3.

### 2.3 The coverage metric counts CRD files, not working resources

`calculate_coverage.py` determines "implemented" by listing YAML in `config/crds/resources`. A
resource counts the moment its CRD merges, whether or not a controller exists.

So a strategy of shipping CRDs quickly would move the coverage number to 80% while shipping
resources that never reconcile. The metric currently rewards the failure mode, and must be fixed
before it is used to steer the work.

## 3. Strategy: generate first, then retrofit in passes

The approach is a pincer — bulk-implement broadly, and separately raise the floor with tests,
checkers and skills until everything complies. The jaws are **not simultaneous**, and the quality
jaw closes *second*.

**This is an experimental repo with no users.** Nothing shipped here can break anyone. So pass 1
ships the crudest implementation that compiles and reconciles: references as plain `string`, no e2e
fixtures, no MockGCP, minimal field coverage, `v1alpha1` only. Quality arrives in later mechanical
retrofit passes, each one corpus-wide.

Sequencing: **generate → retrofit by category → ratchet each category once it is clean.**

### The role of ratchets

A ratchet is a **latch on retrofit progress, not a gate on generation**. Fix a category across the
whole corpus, *then* convert its check to a ratchet so it cannot regress. Applying them before
generation would block the very imperfection the strategy depends on.

The one argument for pre-empting — that a wrong reference is an API-shape error, expensive to change
after publication — is an *upstream* constraint. It does not apply in a repo with no published
users. Refs are retrofitted like everything else, before any port-back upstream.

### Retrofit passes, in order

Each pass is mechanical and corpus-wide, and ends by latching its check:

1. **Refs** — `string` → `*Ref`, using the three-way classifier from `9b675df44c`; latch
   `missingrefs.txt`.
2. **Identity / reference conformance** — the `kcc-identity-reference` skill.
3. **Fuzzers** — `generate-fuzzer`, round-trip coverage.
4. **MockGCP + e2e fixtures**.
5. **Field completeness** — last and largest; latch `missingfields.txt` only after.

### Scope exclusion

Do **not** work any resource that appears as `OPEN` or `PLANNED` in `RESOURCE_STATUS.md`. Those are
owned by the team on the upstream repo. All work here targets the unclaimed remainder.

Sizing that remainder: of the 191 `OPEN`/`PLANNED` rows, **100 already have a CRD on disk** — their
Phase 1 PR merged and a later phase is still open — so they are already inside the 457 implemented,
not in the 549 missing. Only ~91 of the claimed set is genuinely still missing, leaving roughly
**458 unclaimed missing** against the 348 needed for 80%. Comfortable headroom rather than a knife's
edge.

A further reduction is available before any code is written: `coverage_skip.json` currently excludes
57 resources and **nothing for non-GCP-infrastructure APIs**. Around 23% of the resource universe
sits in services like `googleads` (177 resource types), `searchads360` (59), `merchantapi` (43) and
`analyticsadmin` (37). Excluding them removes those resources from both the gap and the denominator
and plausibly cuts the work to 80% by roughly a third. This is a charter decision, not a heuristic —
it needs explicit sign-off before being applied.

---

## 4. Reuse inventory

Extend these rather than building new equivalents.

| Asset | Path | Use |
|---|---|---|
| Coverage calculator | `hack/tools/greenfield/calculate_coverage.py` | Extend with tiers |
| In-flight parser | `get_inflight_resources()` in same file | Already parses `RESOURCE_STATUS.md`; reuse for the exclusion rule |
| Skip policy schema | `hack/tools/greenfield/coverage_skip.json` | Pattern + reason shape to copy |
| Conformance checks | `tests/apichecks/*_test.go` | The de-facto spec; formalize, don't reinvent |
| Ref types | `apis/refs/v1beta1/` | KMS, PubSub, compute, Secret Manager, BigQuery refs already exist |
| Per-phase skills | `.claude/skills/kcc-direct-*` | The layered delivery arms |
| Chore pattern | `.agents/*.md` | Where new agentic workflows go |

---

## 5. Phase 0 — Measure honestly, build the tracker

1. **Add conformance tiers to `calculate_coverage.py`**, derived from what is on disk:

   | Tier | Definition |
   |---|---|
   | `T0` CRD-only | CRD exists; no direct controller registered |
   | `T1` Controller | Controller registered under `pkg/controller/direct/<service>/` |
   | `T2` Verified | MockGCP service + e2e fixture present |
   | `T3` Released | Appears in a release |

   Report coverage per tier. **The 80% target is defined against T1 or better.** T0 never counts —
   otherwise the metric is satisfied by shipping CRDs that never reconcile (§2.3).

2. **Settle the scope policy** (§3) and apply it to `coverage_skip.json`. Do this before freezing the
   tracker: it decides what is even on the list, and it is the single largest reduction available.

3. **Build and freeze the tracker.** Specified in `greenfield-tracker.md` — a committed snapshot
   sorted by implementation ease, pinned to a googleapis SHA, excluding in-flight kinds. Frozen
   membership is what makes "N of M done" a real number.

4. **Snapshot** into `gap_analysis.txt` so progress is attributable to batches.

## 6. Phase 1 — Latching mechanism

Built early, applied late. The mechanism must exist before the first retrofit pass completes; it is
not applied to generation.

1. **Add `CompareRatchetFile` to `pkg/test/utils.go`.** Semantics: new entries fail **even when
   `WRITE_GOLDEN_OUTPUT` is set** and are never written; entries that disappear are pruned, so a
   baseline can only shrink. Implementation exists at commit `9b675df44c`.

2. **Latch the four already-empty baselines now** — `spec_dislike_etag.txt`, `observed_state.txt`,
   `no_refs_in_status.txt`, `comments.txt`. Zero entries today, so converting them costs nothing,
   blocks nothing, and makes four checks permanent guarantees for free. This is the only ratcheting
   that should precede generation.

3. **Latch everything else only as its retrofit pass completes**, in the §3 order. A category is
   latched when its baseline reaches the level we intend to hold, not before.

4. **Ref rules for the five known categories** — KMS key, Pub/Sub topic, VPC network, Secret
   Manager, BigQuery dataset. Roughly 32 actionable finds; ref types already exist under
   `apis/refs/v1beta1/`, so no new API design is required. Part of retrofit pass 1.

## 7. Phase 2 — Layered vs one-go, decided by measurement

Both delivery shapes are plausible and the sandbox exists to settle it with data.

Select **two matched service batches** (~8–12 resources each) from the tracker, matched on
resource count, proto field count, and whether MockGCP already exists for the service.

- **Arm A — layered.** Three passes across the whole batch: types+CRD for all, then controllers for
  all, then realgcp+mockgcp for all. Uses the existing per-phase skills unchanged.
- **Arm B — one-go.** Each resource driven to T2 in a single pass, one PR per resource.

Record per resource: agent turns, wall-clock to green CI, CI runs required, new baseline entries
introduced, defects found by human spot-check, and rework commits. Publish to
`docs/ai/experiments/layered-vs-onego.md`.

Run this **after** a pilot batch has proven the generation path end to end. Comparing two delivery
shapes is only meaningful once one of them is known to work.

Hypothesis to test explicitly: layered amortizes service-level context (protos, client library, mock
scaffolding) and should win on large services; one-go should win on small ones. If that holds, the
output is a **size threshold**, not a global preference.

## 8. Phase 3 — Bulk execution

1. **Batch by service, not by resource.** Amortizes proto, client-library and mock learning, and
   matches the tooling: `controllerbuilder generate-types` takes a repeatable `--resource` flag, so a
   whole service's types and mappers generate in one command. PR #11964 (18 Vertex AI resources at
   once) shows the shape works.
2. **Add `.agents/greenfield-bulk-batch.md`**, modelled on
   `.agents/greenfield-direct-new-resource-types.md` but batch-scoped: read the tracker, take the top
   unstarted batch, refuse anything `get_inflight_resources()` reports, and skip any kind whose CRD
   already exists on disk. The existing chore's "no more than 5 outstanding open issues" throttle is
   a deliberate rate limiter — set an equivalent tuned to sandbox capacity rather than copying it.
3. **Retrofit loop.** Work the §3 passes corpus-wide, latching each as it completes. Track total
   exception line count as the headline quality metric alongside tier coverage.

Detailed generation mechanics live in `greenfield-bulk-generation.md`; the tracker that drives batch
selection is specified in `greenfield-tracker.md`.

---

## 9. Verification

1. **Phase 0** — tiered coverage reported, and the tracker shares zero kinds with the
   `OPEN`/`PLANNED` rows of `RESOURCE_STATUS.md`. Assert the empty intersection in a test.
2. **Latching mechanism** — the decisive check: introduce a deliberate violation, run
   `go test ./tests/apichecks/...` **with `WRITE_GOLDEN_OUTPUT=1`**, and confirm the test still fails
   and the baseline file is unmodified. This is exactly the scenario the golden-file version got
   wrong. Then fix the violation and confirm the entry is pruned.
3. **Suite green** — `presubmit-gatekeeper` on the sandbox PR. The sandbox baseline is 238 success /
   1 skipped / 0 failures, so any delta is attributable.
4. **Generation** — a batch reaches the pass-1 bar: builds, `go vet` clean, CRDs generate,
   controllers register. Record which exception baselines grew and by how much, attributed via
   `grep crd=<kind>` — that measurement is what orders the retrofit passes.
5. **Retrofit** — a category is done when its baseline is at the intended level *and* its check has
   been converted to a ratchet, verified per item 2.
6. **Port-back readiness** — retrofit work is independent of generation and is the most valuable
   piece to land upstream. Keep it on its own topic branch so
   `git rebase --onto upstream/master` yields a clean upstream PR. See `SANDBOX.md`.
