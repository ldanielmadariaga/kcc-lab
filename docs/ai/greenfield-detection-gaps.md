# What the unflagged fields are actually made of

A companion to [greenfield-coverage-invariant.md](greenfield-coverage-invariant.md), which defines
the measurement. This one records what was found underneath it: the causes behind the fields we
miss without telling anyone, what each cause is worth, and what was done about it.

Baseline `c1df0b9326`, 189 greenfield resources, target classes `reference-shape` + `moved` +
`absent`.

| | unflagged | what changed |
|---|---|---|
| first published | 450 | — |
| de-duplicate repeated fields and reference children | 412 | measurement fix, no product change |
| run `scripts/queue-hints` | 276 | a tool that already existed, never run here |
| pair the suffix upstream drops when it adds `Ref` | 256 | measurement fix |
| gate the seeder per Kind | 259 | costs 3, avoids pruning a ratchet |
| flag `empty-observedstate` | **217** | 36 resources, one fact each |

Two of those are measurement fixes rather than coverage improvements. They are in the table because
the number they corrected was published, and quietly restating it would be worse than showing the
correction.

## The biggest hole was a tool nobody ran

`scripts/queue-hints` — then called `queue-ref-hints` — already existed, already walked nested paths
and array items, already applied the shared rules in `tests/apichecks/refs`, and already pruned
subtrees that were reference-shaped. The runbook lists running it as stage 2b. **The 189-resource
bulk run skipped it.**

Running it moved `reference-shape` unflagged from 298 to 147. No new detection logic was written.
Anything that generates in bulk must run it, which is why the runbook now says so in bold.

### It needed a per-Kind gate first

The tool skipped a *service* with no queue file, reasoning that the file's presence marks a service
as bulk-generated. But a service is not uniformly bulk-generated. `compute` holds freshly generated
kinds alongside `ComputeInstance`, which upstream has maintained by hand for years. The
service-level gate let 30 such resources through — `ComputeInstance`, `ComputeInstanceTemplate`,
`CloudBuildTrigger`, `TPUNode`, `DatastreamStream` among them.

Queueing one of those does more than add a note. A queued resource contributes no `[refs]` findings,
so its existing `missingrefs.txt` entries read as *fixed* and get pruned, then return as fresh
violations the moment it graduates. The gate is now per Kind.

That costs 3 fields, because **15 of the 189 measured resources have no resource-level queue entry
to gate on**: BeyondCorpClientConnectorService, CloudBatchResourceAllowance,
ConfigDeliveryResourceBundle, DataformFolder, DialogflowConversationDataset, DialogflowGenerator,
DialogflowKnowledgeBase, DialogflowSecuritySettings, and seven DiscoveryEngine kinds. They are the
resources whose types file already existed when `--prepopulate-spec` ran, so `AddTypeFile` skipped
and no queue was written — the known no-op-on-existing-file gotcha, surfacing somewhere new. Between
them they owe 18 unflagged fields.

## 36 resources generate an empty ObservedState

`ComputeInterconnectObservedState` and `ComputeNetworkAttachmentObservedState` are literally `{}`,
where upstream's CRDs carry 19 and 9 observed fields. The cause is systemic rather than per-field:
those protos carry no `google.api.field_behavior` anywhere, so the generator has nothing to identify
an output field by and everything lands in the Spec.

This is the one signal in the corpus that requires no inference — the generated CRD either has
properties under `status.observedState` or it does not. It is reported per resource, naming the
section, because "your status is empty" is what a human needs and it reads better than nineteen
separate lines saying the same thing. `moved` unflagged fell from 66 to 25.

## Where the remaining 217 sit

### reference-shape, 147

The head of the distribution is gone. What is left is a flat tail of service-specific target types —
`dataStoreRef` 7, `organizationRef` 7, `secretRef` 5, then a long list at 4 and below including
`cloudControlRef`, `lakeRef` and `entryTypeRefs`. No generic name rule reaches those.

**Adding more `refs.NameRules` was measured and rejected for now.** Precision across the 189
resources, counting every field we generate with that leaf name against the ones upstream actually
turned into a reference:

| leaf | fields we generate | upstream made a Ref | precision |
|---|---|---|---|
| `project` | 4 | 4 | 100% |
| `dataStore` | 3 | 3 | 100% |
| `cluster` | 3 | 3 | 100% |
| `topic` | 2 | 2 | 100% |
| `subnetwork` | 6 | 4 | 67% |
| `network` (existing rule) | 17 | 10 | 59% |
| `service` | 19 | 8 | 42% |
| `source` | 10 | 2 | 20% |
| `version` | 17 | 2 | 12% |
| `target` | 4 | 0 | 0% |
| `database` | 4 | 0 | 0% |

The perfect-precision rules would recover about 12 fields. That is not the reason to hold off — this
is: `NameRules` also feeds `TestMissingRefs`, which gates the **whole tree**, where `project`,
`cluster` and `topic` appear far more often and in contexts that are not references. Measuring
precision on 189 resources and shipping a rule that acts on 450 is the mistake that produced the
rejected 2,164-finding heuristic. The existing comment on `MatchName` says tightening the check is a
separate decision that has to come with implementations or reviewed `refs_deferred.txt` entries,
and `TestMissingRefs` cannot even be run in this tree today (see below). Left for someone who can
run that check.

### moved, 25

Almost all server-set fields the protos failed to annotate: `etag` ×6, `state` ×2, `createTime`,
`lastUpdateTime`, `id`, `startTime`, `endTime`. Ten of the 25 are `SQLAdminBackup` alone, whose
ObservedState has exactly one field, so the empty check did not fire.

The fix here is not another flag. `--place-server-set-fields` *produces* these into ObservedState,
which is strictly better — a field in the right place beats a field flagged. The safe allowlist was
measured previously in
[greenfield-generator-findings.md](greenfield-generator-findings.md): `createTime`, `updateTime`,
`deleteTime`, `creationTimestamp`, `uid`, `selfLink`, `selfLinkWithID` and `id` never appear in a
resource-level Spec anywhere in the tree; `kind`, `status`, `etag`, `name` and `state` appear a
handful of times and should be queued as well as moved; `type` appears 36 times and is excluded.

### absent, 45

| cause | count | note |
|---|---|---|
| parent segments never named | ~15 | `spec.location`, `spec.collection`, `spec.collectionGroup`, `spec.tenant`, `spec.parent`, `spec.region` |
| map-value interiors | 6 | `HypercomputeClusterCluster.spec.networkResources.KEY.…` |
| `status.observedState.name` | 3 | `observedstate-identity-field-omitted` should have fired |
| `resourceID` on an embedded resource | 2 | a nested message that itself carries `google.api.resource` |
| no mechanical explanation | ~19 | including a self-recursive message the generator truncates |

The parent-segment group is the next mechanical win. `AddTypeFile` in
`dev/tools/controllerbuilder/scaffold/apis.go` already parses the `google.api.resource` pattern and
already prints it in the `location-omitted-*` detail — it just only ever names `.spec.location`.
Walking every variable segment of the pattern and naming each one we did not produce covers the
group. It needs proto access, so it belongs in the generator rather than in `queue-hints`, and it
only takes effect on a regeneration.

## This tree does not build

44 packages fail to compile, mostly `_identity.go` files written against types the regeneration
changed — `firestorefield_identity.go` reads `obj.Spec.CollectionGroup` where the regenerated Spec
has no such field, and `certificatemanager` compares a `*string` to `""`. One was a straight
regression the run introduced and has been fixed: `pkg/controller/direct/tpu/mapper.generated.go`
had been rewritten to import `cloud.google.com/go/tpu/apiv2/tpupb`, which does not exist.

The consequence for this work is that `go test ./tests/apichecks/...` cannot run the root package,
which is where `TestMissingRefs` lives. `tests/apichecks/refs` and `tests/apichecks/greenfield` both
pass. Any change to `refs.NameRules` is unverifiable until the tree builds.

## Related

* [greenfield-coverage-invariant.md](greenfield-coverage-invariant.md) — the measurement itself
* [greenfield-step1-workflow.md](greenfield-step1-workflow.md) — the runbook, including stage 2b
* [greenfield-generator-findings.md](greenfield-generator-findings.md) — earlier measurements,
  including the two reference heuristics that were tried and rejected
