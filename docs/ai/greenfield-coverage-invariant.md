# The coverage invariant: generated, explained, silent

A generator that cannot produce a field is not the problem. A generator that
cannot produce a field *and does not say so* is, because the resource then looks
finished: its types compile, its CRD installs, its checks pass, and it is missing
things nobody was told about.

So the target is not "generate everything". It is:

> **Every field the baseline CRD has is either generated, or named in the
> judgement queue with a reason.**

Not everything needs to be automatic. A field a human must decide is a perfectly
good outcome. A field nobody knows about is not.

## The three states

**generated** — the field is in our CRD, matching the baseline.

**explained** — it is not, and `apis/<service>/needs_judgement_call.txt` names it
in a field-level entry, with a reason saying why.

**silent** — it is not, and nothing says so. This is the number to drive to zero.

Run the report with:

```sh
python3 hack/tools/greenfield/silence_report.py \
    --resources docs/ai/experiments/data-greenfield/inscope.tsv \
    --only docs/ai/experiments/data-greenfield/measured_scorable.txt \
    --ref c1df0b9326 --verbose-dir /tmp/silence-cache
```

`--verbose-dir` caches `crd-mcp-server score --verbose` output. Scoring 200
resources takes minutes, so keep it while iterating — but **delete it after
regenerating**, or you will measure the previous tree and believe it.

## Three rules that decide whether the number means anything

Each of these produced a confidently wrong answer before it was fixed. They are
worth knowing because any new analysis over the same data will hit them again.

**Count roots, not paths.** When a parent is missing, the score reports its whole
subtree missing as well. 298 missing observedState paths were 156 actual defects;
1264 missing reference paths were roughly 280 actual references. Counting paths
inflates everything several-fold and makes the reference problem look four times
worse than it is.

**Pair names across the reference rename.** The queue names a field as *we*
generated it — `.spec.pipelineJob`. The baseline names it as *upstream* has it —
`.spec.pipelineJobRef.external`. They never match literally, so a naive
comparison reports zero explained references when the real figure is not zero.
Strip a `Ref`/`Refs` suffix and any `.external` / `.name` / `.namespace` /
`.kind` tail before comparing.

The same trap applies to any entry describing a field that should move. An
`output-only-in-comment-only` entry names `.status.observedState.createTime`,
where the field belongs, not `.spec.createTime`, where it currently sits —
otherwise it explains nothing it is meant to explain.

**Ignore the blanket entry.** `untriaged-bulk-generation` is one entry per
resource naming no field. It exists to suppress `[refs]` findings while a
resource is unreviewed, and it covers everything trivially. Counting it would
report 100% explained on day one. Only field-level entries count.

## Reading a result

Two numbers move together and must be read together:

* **silent falling** is progress.
* **generated falling** is not, whatever happened to silent. A change that
  explains fields by no longer emitting them has made the CRD worse and the
  report better. Check both.

The report prints generated counts per section for exactly this reason.

## What the queue reasons mean

| reason | what happened |
|---|---|
| `untriaged-bulk-generation` | blanket, per resource; not an explanation |
| `possible-reference` | a plain string that `google.api.resource_reference` says is a reference |
| `location-omitted-nested-parent` | parent is not project+location, so no `spec.location` was emitted |
| `location-omitted-unknown-parent` | the proto declares no `google.api.resource` pattern |
| `unsupported-field-type` | the type was declined; the field is a `// TODO:` and never reaches the CRD |
| `observedstate-identity-field-omitted` | OUTPUT_ONLY, but skipped because KCC carries it in `status.externalRef` |
| `output-only-in-comment-only` | the proto comment says output only, but no `field_behavior` annotation, so it landed in the Spec |

## Two things the measurement cannot tell you

**It penalises correctness.** The baseline is upstream's hand-written CRD,
including its mistakes. `AIPlatformModel` and `VertexAITrainingPipeline` each
show ~25 "missing" fields that are the arms of upstream's hand-crippled
`google.protobuf.Value` union — which the generator now correctly emits as
`x-kubernetes-preserve-unknown-fields`, and which fixed KCC sending Vertex AI a
double-encoded `trainingTaskInputs`. Any change that corrects something upstream
got wrong reads here as a regression.

**Some "missing" fields are renamed, not absent.** Sixteen are case-only
divergences from an incomplete acronym list: we emit `bootDiskMIB`,
`mysqlProfile`, `targetRpoMinutes` where upstream has `bootDiskMiB`,
`mySQLProfile`, `targetRPOMinutes`. The data is there under another name, which
is a different problem from not generating it, and wants a different fix.

## Related

* [greenfield-step1-workflow.md](greenfield-step1-workflow.md) — the runbook that
  produces the resources this measures.
* [experiments/PROGRESS.md](experiments/PROGRESS.md) — the tracked numbers over time.
* [experiments/greenfield-population.md](experiments/greenfield-population.md) —
  how the greenfield set is defined, and why Terraform-backed resources are excluded.
