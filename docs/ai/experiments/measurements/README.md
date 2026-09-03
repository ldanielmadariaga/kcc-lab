# Measurement history

One file per scored run, so a number quoted in a doc or a PR can be traced back to the run that
produced it. Runs are not comparable across corpus changes, and the filename carries the corpus size
for exactly that reason.

Regenerate a run with:

```bash
python3 hack/tools/greenfield/silence_report.py \
  --resources docs/ai/experiments/data-greenfield/inscope.tsv \
  --ref c1df0b9326 --binary ./bin/crd-mcp-server \
  --verbose-dir /tmp/gf_score --list-silent
```

| run | corpus | implemented | discrepancy | missing | notes |
|---|---|---|---|---|---|
| `2026-09-02-231-resources.txt` | 231 | 10232 (94.2%) | 368 | 257 | The corpus was later found to be under-scoped: 39 greenfield kinds with an upstream baseline were absent from `inscope.tsv`. Numbers in `greenfield-experiment-report.*` and PRs #21/#22 come from this run. |

**Do not compare totals across rows with different corpus sizes.** Every absolute number moves with
the denominator; only the percentages are even loosely comparable, and not those either when the
added resources differ in character from the ones already counted.
