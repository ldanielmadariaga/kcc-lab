# The greenfield tracker: how to build it and how to keep it honest

**Status:** draft, iterating. **Scope:** the experimental sandbox (`kcc-lab`), not upstream policy.

The tracker is the authoritative list of resources we have committed to implementing, and the record
of where each one is. It is what agents read to choose work and what humans read to see progress.

Strategy and rationale live in `greenfield-coverage-strategy.md`. This document covers only the
tracker's construction and upkeep.

**File:** `docs/ai/greenfield-work-queue.md` (not yet created — see §6).

---

## 1. The one rule: generate once, then freeze

The tracker is a **committed snapshot**, hand- and agent-edited from then on. It is not a view
regenerated on demand.

This is deliberate. A continuously-regenerated list has no stable membership, so progress against it
is meaningless — the denominator moves under you and "60% done" can go *down* because upstream added
protos. Freezing gives a **cut line**: a fixed set we committed to, against which `N of M` is a real
number.

Regeneration comes back later as a **drift check** (§5) that *reports* what changed and proposes
additions. It never rewrites the table.

## 2. Data sources

| Input | Path | Provides |
|---|---|---|
| GCP resource universe | googleapis protos at a pinned SHA | every `google.api.resource` type |
| KCC's current surface | `config/crds/resources/*.yaml` | what already exists |
| In-flight upstream work | `hack/tools/greenfield/RESOURCE_STATUS.md` | kinds the team owns; exclude |
| Scope policy | `hack/tools/greenfield/coverage_skip.json` | patterns to exclude, with reasons |
| Extraction + matching | `hack/tools/greenfield/calculate_coverage.py` | `get_gcp_resources`, `get_kcc_resources`, `get_inflight_resources`, `match_resources` |

Two known traps in that tooling:

- **`get_gcp_resources()` skips any path containing `third_party`.** The vendored checkout lives at
  `.build/third_party/googleapis`, so pointing the function at it returns **zero resources**,
  silently. Either fix the filter to be relative to the scan root, or pass a path without
  `third_party` in it. A count of 0 is the symptom.
- **Proto→CRD matching is heuristic.** There are ~608 CRDs on disk but only ~457 match a proto
  resource. Some of that is legitimate (Terraform-only resources with no proto), some is match
  failure. Treat unmatched as "unknown", not as "missing", and spot-check before adding to the table.

## 3. Building the tracker

Order matters; step 2 is the one with the most leverage.

1. **Pin the googleapis SHA.** Record it in the table header. Different revisions give materially
   different universes — the vendored copy yields ~1,506 resource types where the 2026-07-31 snapshot
   recorded 1,724. A tracker that does not say which revision it came from cannot be reconciled later.

2. **Apply the scope policy.** `coverage_skip.json` currently excludes 57 resources and **nothing for
   non-GCP-infrastructure APIs**. Roughly 23% of the universe sits in services like `googleads` (177
   resource types), `searchads360` (59), `merchantapi` (43), `analyticsadmin` (37), `admanager` (26)
   and `chat` (12). Whether those belong in KCC is a **charter decision requiring human sign-off** —
   propose patterns with reasons, get agreement, then apply. Do not infer it with a heuristic; a
   substring match on `tasks` will happily exclude `cloudtasks`, which is real GCP infrastructure.

3. **Filter to manageable resources** — those with a `Create` or `Upsert` RPC. Read-only, transient
   and singleton resources are not KCC resources.

4. **Subtract what exists and what is claimed.** Remove kinds with a CRD already on disk, and kinds
   that `get_inflight_resources()` reports as `OPEN` or `PLANNED`.

5. **Sort by implementation ease** — full-lifecycle leaf resources (parent is Project, Folder, Org or
   Location) first, then next-layer. This is what makes the cut line meaningful: batches are drawn
   top-down, so the easiest work is also the first work.

6. **Group into service batches.** A batch is one service, because generation is service-scoped: a
   batch appends `--resource` lines to a single `apis/<service>/generate.sh`. One service = one batch
   = one PR.

7. **Emit, review, commit.** Then it is frozen.

## 4. Table format

```markdown
<!-- googleapis SHA: <sha>   frozen: <date>   scope policy: coverage_skip.json@<sha> -->

| Kind | Service | Proto type | Parent | Batch | Phase | PR | Notes |
|------|---------|-----------|--------|-------|-------|----|-------|
| NetAppBackupVault | netapp | netapp.googleapis.com/BackupVault | Location | netapp-1 | – | | |
```

**Phase** is the core column. A resource is not binary done/not-done. Starting vocabulary, reusing
`RESOURCE_STATUS.md`'s so both trackers read together:

| Phase | Meaning |
|---|---|
| – | Not started |
| 1 Skeleton | Types, CRD, IdentityV2 |
| 2 Brain | Controller, mappers, registered and reconciling |
| 3 Proof | MockGCP + e2e fixtures |

**Phase definitions are provisional.** They get settled alongside `greenfield-bulk-generation.md`,
since the generation mechanics decide what the real checkpoints are. Retrofit passes (refs, fuzzers,
field completeness) will likely become a parallel column rather than more phases — resolve later.

## 5. Keeping it honest

**Update the Phase column in the same PR that changes the phase.** A tracker updated separately from
the work is a tracker that lies.

**The filesystem is truth; the table is intent.** Before generating any kind, check whether its CRD
already exists in `config/crds/resources/` and skip if so. This is the guard that survives a stale
table — and tables do go stale: `RESOURCE_STATUS.md` currently carries **100 `OPEN`/`PLANNED` rows
whose CRDs already exist on disk**. That is the failure mode to design against, and it is why the
cheap filesystem check matters more than any amount of table discipline.

**Drift check, run periodically.** Re-derive from a newer googleapis SHA and report only:
- resources that appeared upstream since the freeze (candidate additions),
- rows whose Phase disagrees with what is on disk (the staleness above),
- rows now claimed upstream in `RESOURCE_STATUS.md` (candidate removals).

It emits a report for a human to act on. It does not edit the table.

**Re-freezing.** When enough has drifted, cut a new frozen snapshot deliberately, record the new SHA,
and note the delta from the previous freeze. Never silently.

**Concurrency (not needed yet).** Generation currently runs locally and serially, so claim protocols
are unnecessary. If parallel invocations ever happen, claim at *service* granularity — two agents in
one service both edit the same `generate.sh` and conflict by construction.

## 6. Current status

**The tracker file does not exist yet.** It is blocked on the §3 step 2 scope decision, which
determines its membership and could change its size by roughly a third. Building it before that
decision would mean freezing the wrong list.
