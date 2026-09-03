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
| `2026-09-02-231-resources.txt` | 231 | 10232 (94.2%) | 368 | 257 | **Superseded.** The corpus was under-scoped: 44 greenfield kinds with an upstream baseline were absent from `inscope.tsv`. |
| `2026-09-03-275-resources.txt` | 275 | 10966 (91.4%) | 369 | 665 | First run on the derived corpus (`build_inscope.py`). The gap to close goes 195 → 603. |

## What the rescope changed

The 44 added resources brought 1,143 baseline fields and about 734 implemented ones, so they score
near 64% where the original set scored 94%. Nearly all of the cost is `absent`, which went 114 → 481:
fields we generate nowhere, rather than fields we generate in the wrong shape. `discrepancy` barely
moved, 368 → 369.

407 of the newly silent fields come from the added resources, concentrated in a handful of large,
mature ones: `MemorystoreInstance` (33), `RedisCluster` (28), `WorkstationConfig` (20),
`WorkstationCluster` (18), `BigQueryConnectionConnection` (17), `AlloyDBInstance` (16),
`BackupDRBackupVault` (16).

**94.2% was a property of the corpus, not of the generator.** The remaining work is roughly three
times what the earlier number implied, and it is generation work rather than detection work.

**Do not compare totals across rows with different corpus sizes.** Every absolute number moves with
the denominator; only the percentages are even loosely comparable, and not those either when the
added resources differ in character from the ones already counted.
