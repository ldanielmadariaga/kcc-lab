# Replication experiment: 239 resources

Take resources upstream implemented by hand, delete them, regenerate them through the Step 1
pipeline, and score the result against the originals. This is the third and largest such run; the
pilot covered 3 resources and the previous run 100.

Read it alongside [pilot-3-replication.md](pilot-3-replication.md) and
[pilot-100-replication.md](pilot-100-replication.md), which it supersedes in scale but not in
conclusions.

## What was measured

409 kinds carry a `--resource` line in a `generate.sh`. **239 are in scope.** The other 170 are
excluded for reasons unrelated to generator quality: 163 are not `v1alpha1` invocations, 6 sit in
services passing `--skip-scaffold-files` so the generator never writes their types file, and 1 has
no types file at all.

92 services regenerated, 90 without error. 127 resources came back genuinely regenerated. **98 were
scored**, after excluding `compute` and `networksecurity`, whose per-service CRD generation failed
and left stale or truncated files that would have produced meaningless comparisons.

## Result

| bucket | rate | resources with no missing or mismatched field |
|---|---|---|
| spec | 73.7% | 37/98 |
| required | 82.8% | 59/98 |
| status.observedState | 91.4% | 71/98 |

**ObservedState is the headline.** Before this session it was 0% by construction: the scaffolder
emitted an empty struct and a human filled it in. It is now produced mechanically in the first pass.

### What these numbers are, and are not

They are measured **after** this session's generation changes — ObservedState prepopulation, the
output-only comment detector and the plural-acronym flag all landed before the regeneration at
`fabd291f09`, which is why ObservedState reads 91.4% rather than 0%.

They are measured **before any reference work**, and no reference has been implemented. The queue
seeder and the unsupported-type entries described below change what the queue *reports*; neither
emits a single extra CRD field. The spec rate is unchanged by them and stays at 73.7% until someone
actually implements references.

### What the spec number is made of

73.7% on its own misleads. Of 859 missing field paths:

| cause | paths | share |
|---|---|---|
| reference fields, deferred to the judgement pass by design | 624 | 72.6% |
| `map<string, Message>` the generator cannot type | 124 | 14.4% |
| other | 105 | 12.2% |
| plural acronyms that slipped past the flag | 6 | 0.7% |

Excluding references it is 90.8%; excluding the map limitation too, 95.2%.

### The "other" bucket, unpicked

It is three unrelated things, and only one is a generator defect.

**63 fields are genuinely dropped** — the real finding. Whole objects vanish with their subtrees:
`BigQueryMigrationMigrationWorkflow.spec.tasks`, and `HypercomputeClusterCluster`'s
`spec.computeResources` and `spec.networkResources`. `AnalyticsAccount.spec.redirectURI` is a plain
string that is simply absent. These are silent and uncharacterised, and they are the part worth
chasing.

**25 are reference children misfiled by my classifier**, not a separate cause. Paths like
`NetworkManagementConnectivityTest.spec.destination.sqlInstance.external` are the `external`, `name`
and `namespace` children a reference expands into; the classifier missed them because the parent is
named `sqlInstance` rather than `sqlInstanceRef`. The reference gap is correspondingly larger than
152 fields and this bucket smaller.

**14 are a bug in the measurement harness, not the generator.** The Location normaliser recorded
each resource's original form by looking for a `Location` field in its `_types.go`. Upstream often
does not declare one: `DataplexLakeSpec` inlines `parent.ProjectAndLocationRef`, which contributes
both `projectRef` and `location` to the CRD without a `Location` field existing. The normaliser read
that as "no location", deleted the scaffold's, and the comparison then reported it missing. Verified
on `DataplexLake`, whose upstream CRD has exactly `location` and `projectRef` — the same shape the
scaffold emits, so these would have matched untouched. The spec rate is understated by this amount.

## What is missing, and whether anyone can see it

This is the table that matters, and it is less flattering than the coverage figure. Counting
distinct fields rather than paths, 387 are missing:

| category | in the queue today | seeder would add | new TODO entry | silent | total |
|---|---|---|---|---|---|
| reference | 5 | 47 | 0 | 100 | 152 |
| map-of-message | 0 | 0 | 124 | 0 | 124 |
| other | 0 | 0 | 0 | 105 | 105 |
| plural acronym | 0 | 0 | 0 | 6 | 6 |
| **total** | **5** | **47** | **124** | **211** | **387** |

**45% visible, 55% silent** — even with both queue improvements applied. By CRD field path rather
than distinct field the figures are 859 missing and 331 visible (39%); references expand into four
paths each (the object plus `external`, `name` and `namespace`), which is why the two counts differ.

Two things follow. `map<string, Message>` is the one unambiguous win: 124 fields, every one silent
before, every one now carrying a queue entry. And **"other" at 105 fields is both entirely silent
and entirely uncharacterised** — comparable in size to the whole reference gap the seeder closes,
and the largest thing we still cannot explain.

Note that the 54% reference recall quoted elsewhere is against the 111 references whose original
field could be paired, not against all 152 missing reference fields.

## Defect taxonomy

Six classes blocked resources from compiling after regeneration. All are in the scaffold and pruning
layers; none is in the ObservedState or acronym work, which behaved correctly everywhere it ran.

1. **`Location` emitted unconditionally, and as a non-pointer `string`.** 174 of 239 resources
   disagree with the original. Where the proto already has a `location` field it is a straight
   redeclaration. The generator already computes `ParentStyle` from `google.api.resource`, so
   emitting it only for location-scoped resources is derivable: right for 145 of 167.
2. **Pruning removes types the same file still references.** `contentwarehouse` refers to
   `Document_Style` and `Document_Page` after `--prune-unused-types` deleted them.
3. **Kind names collide with generated type names.** `APIHubInstance`, `BillingAccount` and
   `Document` end up declared twice.
4. **Parent shape reaches `_identity.go` but not the Spec template.** `cloudsecuritycompliance`
   expects `Spec.OrganizationRef`; the scaffold emits `ProjectRef`.
5. **`apiextensionsv1.JSON` emitted without its import.** Same class as the `common.Status` bug
   fixed this session, so the fix generalises.
6. **Two `generate-types` invocations writing one `types.generated.go`**, the second overwriting the
   first (`networksecurity`).

## Reference detection

The judgement queue is meant to list what a human still has to decide. Measured against the 111
confirmed references, it named **11 — 10%**. `CloudBuildConnection` needs 15 references and the
queue listed none of them.

The cause is not a bug: `judgementFor` consults only `google.api.resource_reference`, which is
present on a minority of fields. A post-generation seeder reusing the rules from `TestMissingRefs`
reaches **60 correct against 39 wrong — 61% precision, 54% recall.** For a list a person confirms
rather than an automatic rename, that trade is reasonable; a wrong hint costs a moment's reading.

What it still misses, by target type:

| target | confirmed | missed |
|---|---|---|
| `SecretManagerSecretVersionRef` | 12 | 12 |
| `SecretRef` | 13 | 11 |
| `ComputeNetworkRef` | 10 | 8 |
| `KMSCryptoKeyRef` | 6 | 2 |
| `IAMServiceAccountRef` | 8 | 1 |
| `ServiceDirectoryServiceRef` | 7 | 1 |

Secrets are half the remaining gap. `ComputeNetworkRef` is the second largest, and also the riskiest
to chase: "network" is the exact name the earlier heuristic was rejected over.

### One idea that measured badly

Flagging fields whose name contains URI or URL looked promising and is wrong on this corpus. **None
of the 111 confirmed references is URI-named.** There are 9 URI-named string fields and upstream
made references of none — they are container images, GCS object paths, an OAuth redirect. The
existing policy of treating them as not-representable is correct. Caveat: 9 is a small sample,
enough to say "do not invert this", not enough to say "never".

## Lessons for the evaluation framework

These cost more time than the generator defects did.

- **Generate CRDs per package.** `dev/tasks/generate-crds` runs `controller-gen` with
  `paths="./..."`, so one unloadable package blocks all 92 — and when it panics on an
  unresolvable type it names no package at all, leaving nothing to attribute the failure to.
  Per-service invocation succeeded for 90 of 92 and identifies failures by construction.
- **Commit the expensive artifact the moment it exists.** An iterative cleanup script destroyed a
  full regeneration that had not been committed, costing an hour.
- **Never derive a filename by guessing.** Resolving a Kind's CRD as `lowercase(kind)+"s"` silently
  dropped 11.5% of kinds: `*Policy` → `policies`, `Batch` → `batches`, `Process` → `processes`.
- **Report at the granularity the generator writes at.** `types.generated.go` is per-service, so an
  error in it cannot be attributed to one resource, and per-resource isolation cannot work.
- **Separate "wrong output" from "differs from pre-existing consumers".** Most compile failures were
  the second: regenerating `_types.go` while keeping the hand-written `_identity.go` leaves the new
  Spec having to satisfy contracts written against the old one. `generate-types` cannot regenerate
  identity files — only `generate-controller` writes those — so no deletion strategy fixes this, and
  a genuinely new resource would never hit it.

## Methodology

Recorded so the figures can be regenerated rather than believed. Three of the numbers in this
document's drafts turned out to be measurement bugs, each caught only because a sample looked wrong:
a plural-guessing bug that hid 9 resources, a path-normalisation bug that reported 0% precision when
the answer was 61%, and a reference-matching bug that undercounted queue visibility nearly fourfold.

**Baseline `c1df0b9326`.** It predates every experiment commit, so all resources there are
hand-written upstream work. `ae2abcdb9b` is *not* usable: it sits after the pilot re-run, so three
services are generator output there. Verified on `GrafeasNote`, whose CRD has no `location` at the
baseline but does at `ae2abcdb9b`.

**Comparison** by `crd-mcp-server score`, which flattens both CRDs to `path -> type` maps and
reports matched, missing, extra and mismatched per bucket. It reuses `walk`/`flatten` from
`compare.go`, so it sees schemas exactly as the existing equivalence check does.

**Resource resolution** from each CRD's own `kind:` field, never from the filename.

**The measured set** counts a resource only if its `_types.go` actually differs from baseline; a
restored resource would otherwise score a false 100%.

**Reference pairing**: the score tool reports each baseline reference path alongside the plain field
we generated in its place, so pairs are read from the comparison rather than reconstructed. Target
types resolve by matching the JSON tag in upstream's Go source; deriving the Go field name from the
path instead left 15 rows unresolved and mis-stated several per-type counts.

Inputs are preserved in [data-239/](data-239/): the in-scope list, the measured set, raw scores, the
confirmed pairs, the seeder's proposals, the exclusions, and the Location divergences. The reference
answer key is at `hack/tools/greenfield/reference_testset.tsv`.
