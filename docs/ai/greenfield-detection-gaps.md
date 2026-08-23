# What the unflagged fields are actually made of

A companion to [greenfield-coverage-invariant.md](greenfield-coverage-invariant.md), which defines
the measurement. This one records what was found underneath it: the causes behind the fields we
miss without telling anyone, what each cause is worth, and what was done about it.

## The point is detection, not prescription

Worth stating before the numbers, because the numbers invite the opposite reading. Driving the
unflagged count to zero by *fixing* every gap is not the goal and is not achievable — there will
always be an outlier the generator cannot classify. The goal is that the pipeline **notices** when
it is guessing, and hands that to a human. A gap somebody was told about is a finished outcome.

Two consequences for how changes here are judged:

* A detector that names 40 fields is worth more than a fix that silently corrects 40, because the
  detector keeps working on the resources nobody has looked at yet.
* **Anywhere the generator makes a call it cannot justify from the proto, it should leave a trace in
  the generated artifact, not only in the queue.** The queue is a work list somebody clears; the
  types file is what a reader opens when the outlier finally shows up. This applies to any field
  placed or shaped by a rule that is right *in general* — server-set placement today, and reference
  detection when it starts emitting refs rather than naming candidates.

Baseline `c1df0b9326`, 189 greenfield resources, target classes `reference-shape` + `moved` +
`absent`. **92 fields are undetected today**, of which 18 are in 8 Kinds that have no queue entry at
all — their types file already existed when `--prepopulate-spec` ran, so nothing was written and
`queue-hints` skips them by design. No detector missed those; nothing ran. Fixing them means getting
the resource through the pipeline, which is a different job from improving detection, and the report
now says so under the table.

| | unflagged | what changed |
|---|---|---|
| first published | 450 | — |
| de-duplicate repeated fields and reference children | 412 | measurement fix, no product change |
| run `scripts/queue-hints` | 276 | a tool that already existed, never run here |
| pair the suffix upstream drops when it adds `Ref` | 256 | measurement fix |
| gate the seeder per Kind | 259 | costs 3, avoids pruning a ratchet |
| flag `empty-observedstate` | 217 | 36 resources, one fact each |
| name every omitted parent segment | 203 | 18 lines merged from a scratch regeneration |
| separate renamed refs from undetected ones | 109 | measurement fix; see below |
| walk map values in `queue-hints` | 97 | a detector that never entered a map |
| two measured name rules | **92** | `secret` and `project`, 4-for-4 each |

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

### references: 26 undetected, not 147

The single number was three states. Queue entries name a field as *we* generated it; the baseline
names it as *upstream* renamed it. Stripping `Ref` and a trailing noun pairs `kmsKeyName` with
`kmsKeyRef`, but nothing pairs `cryptoKeyName` with `kmsKeyRef`, or `vpc` with `networkRef` — and
both of those are flagged already while reading as misses. Of the 303 reference fields we miss:

| | count |
|---|---|
| flagged, paired by name | 191 |
| queue named something at that parent, cannot tell if it is this field | 86 |
| **nothing named anywhere near it** | **26** |

Positional pairing was tried for the middle bucket and rejected. Matching on the parent alone
credits `billingAccountRef` to `provisionedResourcesParent`; restricting to an unambiguous
one-missing-one-extra parent still mispairs `clusterRef` with `topics`, about 3 wrong in 15. A lossy
guess inside the metric is worse than an honest column. The real fix is at the source: a queue entry
recording the reference *target type* could be resolved by target rather than by name — see
[greenfield-reference-generation.md](greenfield-reference-generation.md).

Two findings got the count from 147 to 26.

**A reference we emit under the wrong name is a rename, not a gap.** `ConnectorsConnection`'s CRD
already has `spec.configVariables[].secretValue` as a reference object — external, name, namespace —
where upstream has `secretValueRef`. Eleven of its twelve "undetected references" were that. Still a
real CRD difference, since a user writing `secretValueRef` gets nothing, but filing it as undetected
sends someone off to build detection for a field we already reference.

**`queue-hints` never entered a map.** `visitProps` walked `Properties` and `Items` and stopped, but
a map's value schema lives in `AdditionalProperties`. Every one of `HypercomputeClusterCluster`'s
seven undetected references was inside `networkResources` or `storageResources`, both
`map<string, Message>`, plus two of `CloudDeployTarget`'s. Twenty new hints from four lines.

That second one is the shape of fix worth preferring: it finds references in services nobody has
looked at, rather than encoding the ones already seen. See the runbook section on extending the
detectors for why `NameRules` is the last resort and not the first.

### moved, 25

Almost all server-set fields the protos failed to annotate: `etag` ×6, `state` ×2, `createTime`,
`lastUpdateTime`, `id`, `startTime`, `endTime`. Ten of the 25 are `SQLAdminBackup` alone, whose
ObservedState has exactly one field, so the empty check did not fire.

`--place-server-set-fields` (opt-in, default off) places these into ObservedState rather than
leaving them in the Spec, and says so in three places: the field carries a `PLACEMENT GUESSED`
comment in the generated type, the CRD description inherits it, and a `server-set-field-placed`
entry goes into the queue.

**Reading the names beat counting them.** An earlier draft took the allowlist straight from the
Spec-appearance counts in
[greenfield-generator-findings.md](greenfield-generator-findings.md), which put `state` and `status`
in the "allow, but queue it" tier on 7 and 2 appearances. Opening those Specs says otherwise:

* `ConfigDeliveryFleetPackage.spec.state` — "Optional. **The desired state** of the fleet package."
* `DLPConnection.spec.state` — "**Required.** The connection's state in its lifecycle."
* `NetworkManagementVPCFlowLogsConfig.spec.state` — "Optional… **Default value is ENABLED**", an
  enable/disable toggle.
* `DiscoveryEngineConversation` and `Session` — a `+kubebuilder:validation:Enum` the user picks from.
* `DLPDiscoveryConfig.spec.status` — "**Required.** A status for this configuration."
* `AccessContextManagerServicePerimeter.spec.status` — in the GCP API `status` names the *enforced*
  perimeter config as opposed to the dry-run `spec`. User-authored configuration whose name
  collides with the CRD's own `status`.

Both are desired state, not observed state, and moving either would take a settable field away.
Excluded. It costs 4 fields out of 44 in this corpus.

`etag` stayed in, on 2 appearances — `AlloyDBCluster` and `ContainerAttachedCluster`, both genuine
optimistic-concurrency inputs — against 27 greenfield resources carrying it status-side and none
spec-side. That is exactly the kind of call the queue entry and the code comment exist for.

The guard is that the message carries **no** `field_behavior` on any field. Relaxing it to "this
field is unannotated" recovers 9 more, almost all `etag` in protos that do annotate other fields,
which is where silence is most likely deliberate. Not taken.

### absent, 41

Characterized, so the residue is known rather than assumed:

| cause | n | example |
|---|---|---|
| no mechanical explanation yet | 13 | `AnalyticsAccount.spec.redirectURI`, `CloudSecurityFramework.spec.labels` |
| parent/identity segment | 12 | `DataprocJob.spec.parent`, `DiscoveryEngineControl.spec.location` |
| deep nested field | 9 | `CloudSecurityComplianceCloudControl.spec.parameterSpec[].subParameters[]` |
| inside a map value | 3 | `HypercomputeClusterCluster.spec.storageResources.KEY.filestore.filestore` |
| `status.observedState.name` | 3 | `ParameterManagerParameter`, `VertexAISchedule`, `DialogflowKnowledgeBase` |
| `resourceID` on an embedded resource | 1 | `VertexAITrainingPipeline.spec.modelToUpload.resourceID` |

Three of these are not detector gaps:

* The **12 parent segments** are the residue `parentSegmentJudgement` could not reach. Eight belong
  to Kinds with no queue entry at all, and `DataprocJob` and `FirestoreDocument` declare no
  `google.api.resource`, so there is no pattern to walk. Their `spec.parent` and `spec.collection`
  are not derivable from the proto by any means.
* The **9 deep-nested** fields are one resource,
  `CloudSecurityComplianceCloudControl.spec.parameterSpec[].subParameters[]` — a self-recursive
  message the generator truncates. A generator limit that currently leaves no trace anywhere, which
  is the thing worth fixing rather than the fields themselves.
* The **3 `observedState.name`** were a real silent drop and are fixed. `PrepopulateSpec` skips the
  identity field with `identityFields`, and `PrepopulateObservedState` files
  `observedstate-identity-field-omitted` for it — but only reaches a field that is `OUTPUT_ONLY`,
  since that is what puts it in `OutputFields`. Where the proto does not mark `name` output-only, it
  was dropped on one side, never seen on the other, and recorded nowhere. `ParameterManagerParameter`
  is the clearest case: its ObservedState has `createTime` and no `name`. Now emits
  `identity-field-omitted`.

That leaves **13 with no mechanical explanation**, which is the honest floor of this pass. They need
reading one at a time rather than forcing into a bucket.

### The parent-segment detector

Done. The types template emits `projectRef` and `resourceID` always, and `location` only when the
parent is exactly `projects/locations`; every other variable segment of the resource pattern was
dropped in silence. `ParentStyle` could not see this — it collapses everything past
project+location into `other`, which is enough to decide what the template renders but not enough
to say what was left out. `protoapi.ParentVariables` now walks the pattern's placeholders directly,
and `AddTypeFile` emits `parent-segment-omitted` for each one the Spec does not carry.

`DiscoveryEngineEngine` is the worked example: pattern
`projects/{project}/locations/{location}/collections/{collection}/engines/{engine}`, upstream
carries `spec.location` and `spec.collection`, we emitted neither and named only location.

Two limits worth knowing. Entries use the **placeholder's** name, not the collection segment
singularised. They disagree once in the corpus — `FirestoreField`'s pattern is
`collectionGroups/{collection}` while upstream calls the field `collectionGroup` — and
singularising is the worse trade, since it fixes that one and mangles `policies` and `addresses`.
And a resource with no `google.api.resource` at all is beyond reach: `DataprocJob`'s `spec.parent`
and `FirestoreDocument`'s `spec.collection` are not derivable from the proto, so they keep only the
`location-omitted-unknown-parent` entry.

**Applying it without a regeneration.** A generator change normally shows up only after a full
regeneration, which would overwrite the tree everything here is measured against. `--output-api`
avoids that: the queue path derives from it, so every service can be generated into a scratch tree
(1.4s each) and only the parent-related lines merged back, gated per Kind exactly like the seeder.
No types file is touched.

That merged 18 lines and took the target from 217 to **203**. Most of the credit landed in
`reference-shape` rather than `absent`, which is right — a parent segment is usually a reference to
the parent resource. `GKEBackupBackup` needs `spec.backupPlanRef`, `ManagedKafkaTopic` needs
`spec.clusterRef`, `WorkflowsExecution` needs `spec.workflowRef`, and none of the three was
mentioned anywhere before.

Three things that run did **not** reach, so the detector is worth more than 18 lines:

* Only 114 of the 189 measured kinds regenerated. Twelve invocations failed with `proto: not found`
  — `hypercomputecluster`, `workloadmanager`, `networksecurity` and others need an overlay or a
  newer descriptor set that the per-service `generate.sh` supplies and this re-invocation did not.
* Every invocation also reported a failure at the *prune* step (`no packages found`), because a bare
  scratch tree has no Go packages to prune. Harmless — the queue is written before pruning — but
  pass `--prune-unused-types=false` next time to keep the log readable.
* The DiscoveryEngine group, which is where this detector was aimed, is gated out. Those kinds have
  no resource-level queue entry to gate on, and queueing them would prune their ratchet. Their
  `spec.collection` stays unflagged until that is resolved on its own terms.

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
