# Greenfield generator: measured progress

One table per measurement, so the trend is visible without re-reading each write-up. Every row comes
from the same experiment — delete resources upstream implemented by hand, regenerate them, score the
CRDs against `c1df0b9326` with `crd-mcp-server score`.

Add a row rather than editing one. A number that moved for a methodology reason is still a number
that moved, and the reason belongs beside it.

## Headline

Measured on the 93 resources scorable in every run to date.

| run | spec | spec excl. refs | required | observedState |
|---|---|---|---|---|
| pilot, 3 resources | — | — | — | 0% by construction |
| 100-resource run | — | — | — | — |
| 239-resource run | 75.2% | not measured | 82.3% | 92.8% |
| after map + JSON well-known types | **80.1%** | **98.2%** | **85.5%** | **93.9%** |

The 239-run figures are quoted here over the 93-resource intersection, not the 98 the original
write-up used, so the columns are comparable. Over its own 98 that run read 73.7 / 82.8 / 91.4.

## What each change was worth

| change | effect |
|---|---|
| ObservedState prepopulation | 0% → ~91%; the single largest move so far |
| output-only comment detector | folded into the above |
| plural acronyms (`--emit-plural-acronyms`) | off by default; renames 83 fields across 23 packages |
| Location from the proto's parent shape | `spec.location` misses 13 → 4 resources; absent refs 47 → 35 |
| `map<string, Message>` | 40 previously-dropped fields across 18 services |
| `Value`/`ListValue`/`Struct` → `apiextensionsv1.JSON` | removes a wire-format bug; **lowers** the score, see below |
| reference-hint seeder | zero effect on any rate — it writes queue entries, not fields |

## Where the remaining gap is

Of 604 missing spec paths in the latest run, **576 are references and 28 are everything else.**

Each unreproduced reference costs four or five missing paths (`fooRef`, `.external`, `.name`,
`.namespace`, sometimes `.kind`), so 576 paths is roughly 130 actual references. The metric amplifies
reference debt fourfold, which is why the raw spec rate reads so much worse than the generator is.

References themselves: 237 in the baseline, **93 reproduced**, 109 left as plain strings, 35 absent.

Non-reference remainder, after discounting references whose names do not end in `Ref` and so escape
the classifier: roughly **15 fields across five resources**. observedState has **70** missing, none
of them references, 43 of which sit in four resources (`DataLabelingDataset`, `RunWorkerPool`,
`DataLabelingEvaluationJob`, `DataLabelingInstruction`).

**Reference support is the next thing worth building.** Nothing else in the measured set is close.

## Two things the metric cannot tell you

**It penalises correctness.** `AIPlatformModel` and `VertexAITrainingPipeline` each gained 20 missing
fields when `google.protobuf.Value` started mapping to `apiextensionsv1.JSON`. The 20 are the
`boolValue` / `nullValue` / `numberValue` / `stringValue` / `structValue` arms of upstream's
hand-written union, whose recursive arm is commented out because it destabilises the CRD. Replacing
it fixed KCC sending Vertex AI a double-encoded `trainingTaskInputs`. The score went down because the
output got better. Any change that corrects something upstream got wrong will read as a regression.

**A stale tree scores as a good one.** A service that fails CRD generation keeps its previous CRD,
which is often identical to the baseline, and scores a false 100%. The measured set therefore counts
a resource only when its `_types.go` actually differs from baseline. `compute` scored this way in the
239-run.

## Traps that have each cost a full run

* **`--prepopulate-spec` does not rewrite an existing types file.** Regenerating in place leaves the
  old Spec and measures a generator version that is no longer current. Delete the resource first;
  that step is the experiment, not tidying.
* **`generate.sh` calls `generate-crds`, which runs `controller-gen` tree-wide.** One unloadable
  package panics the CRD step for every service. Use `SKIP_GENERATE_CRDS=1`, then `controller-gen`
  per service.
* **`WRITE_GOLDEN_OUTPUT=1` absorbs unrelated drift**, reliably `IAPSettings.diff`. Read the diff.
* **Validate the analyser against a known answer before trusting a new number.** Three figures in the
  239-run drafts were measurement bugs. The current analyser reproduces the published
  73.7 / 82.8 / 91.4 from the preserved score file before being pointed at anything new.

## Sources

| run | write-up | raw data |
|---|---|---|
| pilot, 3 | [pilot-3-replication.md](pilot-3-replication.md) | — |
| 100 | [pilot-100-replication.md](pilot-100-replication.md) | — |
| 239 | [replication-239.md](replication-239.md) | [data-239/](data-239/) |
| 239 re-run | [replication-239.md](replication-239.md#re-run-after-the-map-and-json-well-known-type-work) | [data-239-rerun/](data-239-rerun/) |
