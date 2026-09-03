# The coverage invariant: implemented, discrepancy, missing

**Every field present in a k8s-config-connector master CRD is either produced by
our pipeline, or flagged in the judgement queue.** Nothing silently absent.

This is a CRD-compatibility target, not a proto-fidelity one. If master's CRD has
`spec.pipelineJobRef.external` and we produce `spec.pipelineJob`, a user's YAML
breaks — it does not help that both come from the same proto field. Renamed,
moved and reference-shaped fields are all fields KCC has and we do not.

A generator that cannot produce a field is not the problem. A generator that
cannot produce a field *and does not say so* is, because the resource then looks
finished: its types compile, its CRD installs, its checks pass, and it is missing
things nobody was told about. A field a human must decide is a fine outcome. A
field nobody knows about is not.

Measured over the 189 greenfield resources against baseline `c1df0b9326`:

| | count |
|---|---|
| fields in KCC master's CRDs | 9,310 |
| we produce | 8,727 (93.7%) |
| we miss | 583 (6.3%) |

## Two questions, not one

The report answers two questions about the fields we miss, and they are
**independent**. Reading them as one axis is what made an earlier version of this
output contradict itself: it said "104 genuinely appear nowhere in our output" and
then showed 37 of those 104 flagged, which reads as nonsense.

It is not nonsense. The two questions are:

* **why does the field differ?** — the five classes, the rows of the table
* **was anybody told about it?** — the three columns

A field can be `absent` **and** flagged at the same time. That is the queue doing
its job: we did not produce the field, and the queue says so, so a human will see
it. Nothing about "absent" implies "unreported".

## The three states

Every baseline field lands in exactly one, and the split is on **what we produced**, not on whether
we mentioned it.

**implemented** — the same field at the same path. Nothing to do.

**discrepancy** — we produce the field, but not as upstream has it. In the other section, under a
different name, or as a plain string where upstream has a reference object. A user's YAML still
breaks, so this is a real difference, but the work is detection or placement rather than generation.

**missing** — we produce nothing at all. This is where generation work goes, less the part we
decline to produce on purpose (see below).

Leading on "flagged" instead, as this report did until recently, buried that distinction. The two
axes are close to independent: three quarters of discrepancies carry a queue entry against one in
seven of absences, so a combined figure describes neither population.

## Whether we said anything, as a second axis

Flagging is recorded against both states rather than being a state of its own.

**field-flagged** — a queue entry names this exact field, with a reason.

**section-flagged** — a resource-level entry names the *section* the field belongs to. Today that is
`empty-observedstate`, written when a resource's `status.observedState` was generated with no fields
at all. It is specific enough to act on. The blanket `untriaged-bulk-generation` entry names neither
field nor section and deliberately counts for nothing.

**leaf-flagged** — a shared nested message's finding, written as a `#` comment because it belongs to
no single Kind, matched back by leaf name. Looser than path matching, and the docstring says so.

**silent** — nothing says anything. A silent discrepancy and a silent absence are both bad and need
different fixes, which is why they are counted apart.

## The target, and what is excluded from it

Four of the six classes are the target. The other two are differences we accept, reported below the
subtotal and left out of it:

* `renamed` (40) is a casing table — `bootDiskMIB` against upstream's
  `bootDiskMiB`. A fix, not a judgement call.
* `intentionally-different` (62) is the `google.protobuf.Value` union arms, which
  we map to `apiextensionsv1.JSON` on purpose. Whether to keep doing that is an
  open question and is **deliberately deferred**; it is recorded rather than
  counted as a miss.

Current state, with the produced axis alongside the class:

| why it differs | total | we emit it | we emit nothing | flagged |
|---|---|---|---|---|
| `reference-shape` | 82 | 1 | 81 | 9 |
| `moved` | 64 | 64 | 0 | 34 |
| `absent` | 114 | 0 | 114 | 30 |
| `reference-not-detected` | 263 | 263 | 0 | 246 |
| **subtotal, the target** | **523** | **328** | **195** | **319** |
| `renamed` | 40 | 40 | 0 | 0 |
| `intentionally-different` | 62 | 0 | 62 | 0 |

The `we emit it` column is what makes the classes legible as different problems. `moved` and
`reference-not-detected` are entirely fields we produce, so no amount of generation work touches
them. `absent` is entirely fields we do not, so no amount of detection work does.

Three of the classes imply the answer rather than needing it measured: `classify` returns `moved`
only when the same path is in the other section, `renamed` only when a matching name sits at the
same parent, and `reference-not-detected` only when a plain field sits at the de-suffixed stem. Each
is proof we emitted something. Name-pairing alone disagreed with the classifier on 29 `renamed`
fields and put them in `missing` while the class column said we emit them, so the classifier wins
for those three.

A fourth column, "unsure", used to sit here: 34 references where the queue named a field at the same
parent but upstream renamed it beyond pairing. They are counted as silent, which overstates the gap
rather than flattering it. See
[greenfield-detection-gaps.md](greenfield-detection-gaps.md).

## Why each field differs

Every field we miss is classed by *why*, to route the fix. All five are real gaps;
they need different work.

| class | what it means | where the fix lives |
|---|---|---|
| `reference-not-detected` | **we emit the field as a plain string**, KCC has a `Ref` object | detection; `scripts/queue-hints` |
| `reference-shape` | KCC has a `Ref` and we have **no field at that stem at all** | generation, then detection |
| `moved` | we emit it in Spec, KCC has it in `status.observedState` | placement rules |
| `absent` | it appears nowhere in our output | generation |

The first two are easy to confuse and want different work. `reference-not-detected` is the field
being *there* and unrecognised, which a detector fixes. `reference-shape` is the field being absent
in any form, which generation has to fix before detection has anything to look at. They were one
class until the classifier was taught to check `extras` for a plain field at the de-suffixed stem,
and the merged number made the detection gap look four times larger than it is: of 263 plain-string
references, 246 are already flagged.
| `intentionally-different` | we model it deliberately otherwise, e.g. `google.protobuf.Value` as `apiextensionsv1.JSON` | a decision to record — still a CRD difference |
| `renamed` | same field, different name: `bootDiskMIB` vs `bootDiskMiB` | the acronym list |

A reference is recognised by a `Ref`/`Refs` segment, and also by a missing
`.external` child under a repeated field. Upstream does not always add the suffix:
`producerAcceptLists[]` and `relatedProjects[]` are references with plain names.
`.external` is the discriminator because KCC puts it on references and nothing
else in a CRD has it, whereas a plain repeated message can have a `name` field of
its own.

`moved` and `renamed` are detected from the score's own *extra* list — a field we
put somewhere else shows up as missing in one place and extra in another. That
catches a section swap (Spec versus observedState) and a same-parent case
difference. It does **not** catch a field moved to a different nesting level, so
`absent` is somewhat overstated: comparing `+kcc:proto:field` annotations across
the two trees suggests the true never-produced count is nearer 79 than 85.

That annotation comparison is a useful second opinion — it says the generator
emits 16,295 of the 16,438 proto fields KCC declares, so generation itself is
healthy and the gap is mostly shape, name and placement. It is **not** the
measurement. Using it to shrink the gap explains away fields users genuinely
cannot set.

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

## Four rules that decide whether the number means anything

Each of these produced a confidently wrong answer before it was fixed. They are
worth knowing because any new analysis over the same data will hit them again.

**Count roots, not paths.** When a parent is missing, the score reports its whole
subtree missing as well. 298 missing observedState paths were 156 actual defects;
1264 missing reference paths were roughly 280 actual references. Counting paths
inflates everything several-fold and makes the reference problem look four times
worse than it is.

**Count a repeated field once.** The score reports `foo` and `foo[]` as two
separate missing paths; they are one field. 28 of the first published 533 were
this, and the parent check did not catch them because it splits on `.`, so the
`[]` sibling never matched.

**Count a reference site once, suffixed or not.** A missing `fooRef` hides its own
`.external` / `.name` / `.namespace` / `.kind` children, so the roots rule
collapses it to one entry. Upstream does not always add the suffix, and without
the same collapsing an unsuffixed reference such as `producerAcceptLists[]` costs
four where a suffixed one costs one.

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
report 100% explained on day one, so it counts for nothing.

A resource-level entry that names a *section* is different, and does count — in
its own column. `empty-observedstate` says which part of the resource is missing
and is actionable on its own. The distinction is whether an entry tells a human
where to look.

## Reading a result

Two numbers move together and must be read together:

* **unflagged falling** is progress.
* **produced falling** is not, whatever happened to unflagged.

**A field produced into the wrong place beats a field dropped and flagged.**
`createTime` lands in Spec rather than ObservedState today — wrong struct, but
present in the CRD and readable. Deleting it and filing a queue entry would
improve every number in this report and take a working field away. The report
cannot see that difference, which is why the rule is written here and why the
tool prints "we produce" beside the gap.

## What the queue reasons mean

| reason | what happened |
|---|---|
| `untriaged-bulk-generation` | blanket, per resource; does not count as a flag |
| `possible-reference` | a plain string that `google.api.resource_reference` says is a reference |
| `location-omitted-nested-parent` | parent is not project+location, so no `spec.location` was emitted |
| `parent-segment-omitted` | another part of the resource name — `collection`, `tenant`, `database` — that the Spec does not carry |
| `location-omitted-unknown-parent` | the proto declares no `google.api.resource` pattern |
| `unsupported-field-type` | the type was declined; the field is a `// TODO:` and never reaches the CRD |
| `possible-reference-by-description` | the description spells out a resource-name path — strong |
| `possible-reference-by-name` | the field name matches a known target — a hint, roughly a third wrong |
| `empty-observedstate` | nothing was generated into `status.observedState`; counts as a section flag |
| `server-set-field-placed` | put in ObservedState by name because the proto carries no `field_behavior`; the field also says so in the generated type |
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
