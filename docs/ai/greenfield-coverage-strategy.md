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

This matters more at scale. One resource adding `+1` line to an exceptions file is conspicuous in
review. Eighteen resources across 36,788 added lines is not.

**Bulk generation against self-absorbing baselines converts review debt into archaeology.**

### 2.3 The coverage metric counts CRD files, not working resources

`calculate_coverage.py` determines "implemented" by listing YAML in `config/crds/resources`. A
resource counts the moment its CRD merges, whether or not a controller exists.

So a strategy of shipping CRDs quickly would move the coverage number to 80% while shipping
resources that never reconcile. The metric currently rewards the failure mode, and must be fixed
before it is used to steer the work.

## 3. Strategy: close the quality jaw first

The approach is a pincer — bulk-implement broadly, and separately raise the floor with tests,
checkers and skills until everything complies. The two jaws are **not simultaneous**. Ratchets are
what make a merely-adequate implementation safe, because they bound what can get worse. They land
first.

Sequencing: **conformance spec + ratchets → measure honestly → bulk generate in batches.**

### Scope exclusion

Do **not** work any resource that appears as `OPEN` or `PLANNED` in `RESOURCE_STATUS.md`. Those are
owned by the team on the upstream repo. All work here targets the unclaimed remainder.

Preliminary arithmetic: 549 missing manageable − ~191 claimed ≈ **358 unclaimed**, against the 348
needed for 80%. Phase 0 confirms this exactly.

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

## 5. Phase 0 — Measure honestly, build the work queue

1. **Add conformance tiers to `calculate_coverage.py`**, derived from what is on disk:

   | Tier | Definition |
   |---|---|
   | `T0` CRD-only | CRD exists; no direct controller registered |
   | `T1` Controller | Controller registered under `pkg/controller/direct/<service>/` |
   | `T2` Verified | MockGCP service + e2e fixture present |
   | `T3` Released | Appears in a release |

   Report coverage per tier. **The 80% target is defined against T1 or better.** T0 never counts.

2. **Emit `hack/tools/greenfield/work_queue.json`** — missing-manageable minus in-flight, with
   kind, service, proto path, parent shape, target tier and batch id.

3. **Snapshot** into `gap_analysis.txt` so progress is attributable to batches.

## 6. Phase 1 — Conformance spec and ratchets

Lands **before** any bulk generation.

1. **Write `docs/ai/resource-conformance.md`** — the normative per-tier definition of a compliant
   resource. One row per rule: what it requires, which check enforces it, which baseline records
   exceptions, and whether it is required at T1 or T2. Rows derive from the 20 existing checks.

2. **Add `CompareRatchetFile` to `pkg/test/utils.go`.** Semantics: new entries fail **even when
   `WRITE_GOLDEN_OUTPUT` is set** and are never written; entries that disappear are pruned. The
   baseline can only shrink.

3. **Convert baselines to ratchets, cheapest first.**
   - *Free today* — `spec_dislike_etag.txt`, `observed_state.txt`, `no_refs_in_status.txt`,
     `comments.txt` are already empty. Converting them costs nothing and makes them permanent
     guarantees.
   - *Small* — `shortnames`, `recursivetypes`, `outputonlyspecfields`, `sensitive`,
     `crds_have_parent_refs`, `printercolumns`, `naming_violations`.
   - *Deferred* — `alpha-missingfields` (4,206 lines), `missingfields` (3,406), `acronyms` (780).
     These need a per-resource attribution scheme first, so a new resource's entries are separable
     from the legacy mass.

4. **Add ref rules for the five known categories** — KMS key, Pub/Sub topic, VPC network, Secret
   Manager, BigQuery dataset. Roughly 32 actionable finds; ref types already exist under
   `apis/refs/v1beta1/`, so no new API design is required.

## 7. Phase 2 — Layered vs one-go, decided by measurement

Both delivery shapes are plausible and the sandbox exists to settle it with data.

Select **two matched service batches** (~8–12 resources each) from the work queue, matched on
resource count, proto field count, and whether MockGCP already exists for the service.

- **Arm A — layered.** Three passes across the whole batch: types+CRD for all, then controllers for
  all, then realgcp+mockgcp for all. Uses the existing per-phase skills unchanged.
- **Arm B — one-go.** Each resource driven to T2 in a single pass, one PR per resource.

Record per resource: agent turns, wall-clock to green CI, CI runs required, ratchet violations
attempted, defects found by human spot-check, and rework commits. Publish to
`docs/ai/experiments/layered-vs-onego.md`.

Hypothesis to test explicitly: layered amortizes service-level context (protos, client library, mock
scaffolding) and should win on large services; one-go should win on small ones. If that holds, the
output is a **size threshold**, not a global preference.

## 8. Phase 3 — Bulk execution against frozen bars

1. **Batch by service, not by resource.** Amortizes proto, client-library and mock learning.
   PR #11964 (18 Vertex AI resources at once) demonstrates both that the shape works and what it
   costs without ratchets.
2. **Add `.agents/greenfield-bulk-batch.md`**, modelled on
   `.agents/greenfield-direct-new-resource-types.md` but batch-scoped: read `work_queue.json`, claim
   a batch, refuse anything `get_inflight_resources()` reports, and cite `resource-conformance.md`
   as the acceptance bar. The existing chore's "no more than 5 outstanding open issues" throttle is a
   deliberate rate limiter — set an equivalent tuned to sandbox capacity rather than copying it.
3. **Retrofit loop.** Ratcheted baselines only shrink. Track total exception line count as the
   headline quality metric alongside tier coverage.

---

## 9. Verification

1. **Phase 0** — tiered coverage reported, and `work_queue.json` shares zero kinds with the
   `OPEN`/`PLANNED` rows of `RESOURCE_STATUS.md`. Assert the empty intersection in a test.
2. **Phase 1 ratchets** — the decisive check: introduce a deliberate violation, run
   `go test ./tests/apichecks/...` **with `WRITE_GOLDEN_OUTPUT=1`**, and confirm the test still fails
   and the baseline file is unmodified. This is exactly the scenario the golden-file version got
   wrong. Then fix the violation and confirm the entry is pruned.
3. **Suite green** — `presubmit-gatekeeper` on the sandbox PR. The sandbox baseline is 238 success /
   1 skipped / 0 failures, so any delta is attributable.
4. **Phase 2** — both arms reach T2, experiment table populated, recommendation stated as a threshold
   rule with supporting numbers.
5. **Port-back readiness** — the Phase 1 ratchet work is independent of bulk generation and is the
   most valuable piece to land upstream. Keep it on its own topic branch so
   `git rebase --onto upstream/master` yields a clean upstream PR. See `SANDBOX.md`.
