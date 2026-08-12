# Experiment: can the checks reproduce PR #11964's human review?

**Result: partially, and the misses are more interesting than the hits.**

PR #11964 (18 batch-generated Vertex AI resources) was reviewed by hand and closed unmerged. Human
review left 13 comments on `vertexaibatchpredictionjobs`; CI flagged nothing. Since the resource
never landed, regenerating it is a genuine reproduction rather than an audit of already-fixed code.

Method: add `VertexAIBatchPredictionJob` to `apis/vertexai/generate.sh`, generate, and assemble the
Spec **exactly as generated** — no manual reference conversion. A careful assembly would test the
implementer, not the checks.

---

## Scorecard

| # | Review comment | Field | Outcome |
|---|---|---|---|
| 1 | "missing flextstart.NodeRecycling" | `flexStart.nodeRecycling` | **N/A** — field absent from the pinned googleapis; cannot reproduce |
| 2 | "should be a list of GCS Refs" | `gcsSource.uris` | **MISSED** |
| 3 | "should we have a detector for URI?" | *(proposal)* | still open — see below |
| 4 | `google.protobuf.Struct` "weird one" | `nearestNeighborSearchConfig` | **CAUGHT — as a build failure** |
| 5 | `ListValue` "weird one" | `outputIndices` | **CAUGHT — as a build failure** |
| 6 | BQ URI, predicted false positive | `bigquerySource.inputURI` | **MISSED** (so also no false positive) |
| 7 | "probably needs to be a ref?" | `gcsSource` | **MISSED** |
| 8 | "another ref" | `notificationChannels` | **MISSED** |
| 9 | "another ref?" | `analysisInstanceSchemaURI` | **MISSED** |
| 10 | "what's going on here?" | `objectiveConfigs` (`skipped_kcc:proto=`) | **N/A** — marker no longer emitted |
| 11 | "ref?" | `outputURIPrefix` | **MISSED** |
| 12 | "ref" | `gcsSource` (dataset) | **MISSED** |
| 13 | "descriptions devolved" | `modelMonitoringStatsAnomalies` | **N/A** — same marker, gone |

Caught: 2. Missed: 6. Not reproducible: 3. Still open: 1.

**Found that review did not:**

| Finding | Check |
|---|---|
| `Location string` must be `*string` (generator scaffold defect) | `TestGreenfieldBulkTypesConformance` |
| `.spec.model` should be a reference | `TestMissingRefs` |
| `.spec.serviceAccount` should be a reference | `TestMissingRefs` |
| 3 proto fields silently dropped | `TestGreenfieldDroppedFields` |

## What the misses have in common

Every miss is a **URI-shaped Cloud Storage or BigQuery field**. The detector on this branch matches
path templates (`projects/`, `locations/{`, …) plus an `erviceAccount` suffix, and nothing else:

- `gcsSource.uris` — "Google Cloud Storage URI(-s) to the input file(s)" — no path template.
- `bigquerySource.inputURI` — "`bq://projectId.bqDatasetId.bqTableId`" — `projectId`, not `projects/`.

This is exactly the gap
[issue #12344](https://github.com/GoogleCloudPlatform/k8s-config-connector/issues/12344) documents,
reproduced independently. Comment #3 — "should we have a detector that flags CRDs with the term
URI?" — is still the right question, and still unanswered on this branch.

A stronger classifier exists on the `refs-classifier-standalone` branch (`isPatternField`,
`notRepresentableReason`, Cloud Storage bucket detection). It is **not** on master. Porting it is the
single highest-value follow-up, and this experiment is the argument for it.

## Corrections to prior claims

**#12344 says the generator "got two of them right — `model` and `serviceAccount` shipped as proper
refs".** It did not. Generated naively, both are plain `*string`. The PR's final state presumably
reflected review fixes, not generator output. Our `TestMissingRefs` flagged both — so the checker now
does mechanically what a reviewer did by hand.

**Struct/ListValue is not a subtle CRD nit.** Review called them "weird ones". They are a hard build
break: the type generator emits `Value`/`ListValue` structs and the mapper generator emits calls to
`Value_v1alpha1_FromProto`, but **no service anywhere defines that function**. Any resource with a
`google.protobuf.Value` field fails to compile. This is a systemic generator gap, not a per-resource
judgement call.

## Generator gaps found

Three, all blocking, all needing hand intervention:

1. **`google.protobuf.Value` / `ListValue` have no mappers anywhere.** Reached transitively here via
   `ExplanationSpec` → `Examples` / `ExplanationParameters`. Dropping the field is the only Step 1
   workaround.
2. **`[]common.Status` emits a mapper call without its import.** The singular `*common.Status` is
   handled by `direct.Status_FromProto`; the repeated form emits `common.Status_v1alpha1_FromProto`
   and never adds the `common` import.
3. **The scaffold emits `Location string`, not `*string`** — violating the pointer rule the project
   enforces elsewhere. Every generated resource starts with this defect.

## The golden-file defect, demonstrated

`missingrefs.txt` uses `CompareGoldenFile`. Running the suite with `WRITE_GOLDEN_OUTPUT=1` **absorbed
both new ref violations**, and the test then passed:

```
$ go test ./tests/apichecks/ -run TestMissingRefs        # FAIL: 2 new violations
$ WRITE_GOLDEN_OUTPUT=1 go test ./tests/apichecks/ -run TestMissingRefs
$ go test ./tests/apichecks/ -run TestMissingRefs        # ok
```

The findings were real, correct, and silently promoted to accepted exceptions by the standard
regenerate-goldens workflow. `greenfield_dropped_fields.txt` uses `CompareRatchetFile` and refused
the equivalent write. Converting `missingrefs.txt` to the ratchet is the obvious next step.

## Bug found in our own tooling

`CompareRatchetFile`'s prune wrote the file **without a trailing newline**. Appending an entry with
`cat >>` then concatenated it onto the last existing line, silently corrupting both. These baselines
are appended to by hand and by agents, so the newline is load-bearing. Fixed, with a regression test.

## State of the resource

Ships at the Step 1 bar with known gaps, all recorded rather than hidden:

- `ModelParameters`, `ExplanationSpec`, `PartialFailures` — in `greenfield_dropped_fields.txt` with
  reasons pointing at the generator gaps above.
- `.spec.serviceAccount` → `ServiceAccountRef` implemented, matching `VertexAICustomJob`.
- `.spec.model` — **cannot** be fixed in Step 1: no `VertexAIModelRef` exists, and creating
  `_reference.go` / `_identity.go` is a separate step. Recorded in `missingrefs.txt`.

That last point is a real boundary this experiment surfaced: **the "implement the reference" rule is
unsatisfiable in Step 1 when the target ref type does not exist yet.** Either Step 1 grows to include
creating external-only refs, or the rule needs an explicit escape for this case.
