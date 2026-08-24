# Replication experiment: pilot 3

Delete three resources that were implemented by hand upstream, regenerate them through the Step 1
runbook, and score the result against the originals. This is the first measurement of the generator
against a known-good answer.

## Setup

kcc-lab branched from upstream at `2c91d37bde` (2026-08-11) and the 101 upstream commits since added
no CRDs, so the 15 most recently added upstream resources are present locally and identical to
upstream. The reference copies come from our own git history.

| Kind | Service | Types | CRD | proto annotates `OUTPUT_ONLY`? |
|---|---|---|---|---|
| `NetworkSecurityURLList` | networksecurity | 109 lines | 8 KB | yes |
| `TranscoderJob` | transcoder | 183 lines | 92 KB | yes |
| `GrafeasNote` | grafeas | 250 lines | 52 KB | **no** |

Scored with `crd-mcp-server score`, added for this experiment. It reuses `gitShow`, `parseCRD`,
`walk`/`flatten` and `getVersionSchemas` from `compare.go`; `compareEquivalence` itself answers a
pass/fail backward-compatibility question and is not reusable here.

## Results

`ee4da88822` is the CONTROL commit: both experimental flags enabled, nothing deleted.

| | spec M1 | spec M2 | required | observedState M1 | observedState M2 | refs |
|---|---|---|---|---|---|---|
| NetworkSecurityURLList | 100.0% | 100.0% | 100.0% | 0.0% | 100.0% | 1/1 |
| TranscoderJob | 98.4% | 100.0% | 100.0% | 0.0% | 100.0% | 2/2 |
| GrafeasNote | 97.4% | 100.0% | 100.0% | 0.0% | 100.0% | 1/1 |

M1 is mechanical, with no human input at all. M2 follows a judgement pass working from the proto and
the runbook only.

At M2 nothing is missing anywhere. The one remaining difference is `spec.labels` on TranscoderJob,
which the judgement pass kept and upstream omits — a divergence, not a miss. All four ratchets are
unchanged and `alpha-missingfields.txt` is the only golden that grew, which is expected, since Step 1
ships no fixtures.

## What the generator already gets right

Required markers came back at 100% on all three, 30 of 30 on TranscoderJob.
`--emit-required-from-proto` reproduced the hand-written required set exactly, and the control run
confirmed the split is working: 16 `required:` blocks were added across two CRDs and every one landed
under `spec`, none under `status`.

Spec field coverage is 97-100% before any human touches it.

## The four gaps, measured

**ObservedState is 0% mechanically, on every resource.** This is phase 5 showing up exactly as
designed rather than as a surprise. Filling it took only the proto: running `generate-types` into a
throwaway `--output-api` directory, where no hand-written type claims the message, makes the
generator emit the proto-derived `<Proto>ObservedState`. That is phase 5 performed by hand, and it
produced the correct answer for all three.

**"Output only." in the proto comment is a signal the generator ignores.** GrafeasNote's `Kind`,
`CreateTime` and `UpdateTime` landed in `spec` even though the proto documents all three as output.
`PrepopulateSpec` does skip `OUTPUT_ONLY`, but grafeas states it in prose and carries no
`google.api.field_behavior` annotation — the generator emitted no `NoteObservedState` at all, which
confirms it. The comment text is machine-readable and would have caught all three here, where the
name allowlist in the phase 5 design catches only two. Worth adding as a third input.

**The generator misses plural acronyms; the checker catches them.** `codegen.IsAcronym`
(`pkg/codegen/common.go:106`) does a plain `EqualFold` against `codegen.Acronyms`, so `Uris` and
`Ids` do not match and the generator wrote `RelatedUris` and `KbArticleIds`. `TestCRDsAcronyms`
(`tests/apichecks/crds_test.go:645`) uses **the same list**, strips a trailing `s` and retries, so it
correctly demands `RelatedURIs` and `KbArticleIDs`. Singular acronyms already work — the same struct
has `SupportURL` from `support_url`. One function, and `acronyms.txt` holds 781 recorded violations
corpus-wide to size the win against.

**A reference inside a nested generated type costs more than the runbook implies.** TranscoderJob's
`config.pubsubDestination.topic` is a Pub/Sub topic with no `resource_reference` annotation. Fixing
it meant hand-writing the whole `PubsubDestination` struct into `<kind>_types.go`, because a field
inside a generated nested type cannot be changed in place. The same applied to both grafeas acronym
fixes. The runbook's "the cost stops at `_types.go`" is true but understates it.

## Smaller findings

- **The scaffold's `Location` is wrong for a project-scoped resource.** GrafeasNote's identity is
  `projects/{project}/notes/{note}`, and the scaffolded `spec.location` broke the existing fixtures.
  Derivable from the identity file, which the generator already writes correctly.
- **An empty proto message produces an invalid CRD.** `grafeas.v1.SecretNote` has no fields, so the
  generated object has no properties and `TestCRDObjectTypes` rejects it. The Step 1 answer is to
  omit the field.
- **`Error *common.Status` needs a `common` import the generator does not add**, as the greenfield
  skill records.
- **Queue suppression transiently prunes `refs_not_representable.txt`.** While TranscoderJob was
  queued, its five `not_representable` entries disappeared from that golden and only returned on
  graduation. `missingrefs.txt` carries entries forward; this golden does not. A
  `WRITE_GOLDEN_OUTPUT=1` run mid-queue would drop them from the record.
- **Deleting several resources at once breaks `generate-crds` until all are regenerated**, because
  `controller-gen` runs over the whole tree.
- **`findTypeDeclarationWithProtoTag` matches a dangling comment.** Removing a struct but leaving its
  `+kcc:observedstate:proto=` comment made the generator believe a hand-written type still claimed
  the message.

## Caveat on the blind protocol

The hand-written `pkg/controller/direct/grafeas/mapper.go` survived the regeneration, and its compile
errors named GrafeasNote's correct acronym casing and ObservedState fields before the judgement pass
ran. Those GrafeasNote decisions are assisted rather than blind. URLList and TranscoderJob were not
leaked. The leak is mild, since `acronyms.txt` encodes the casing rule mechanically and the
ObservedState answer came from the proto, but it is a flaw in the setup: a real greenfield resource
has no hand-written mapper to check against.
