# Closing the gap between a generated CRD and a written one

*Status: merged in the kcc-lab sandbox repository, every flag off by default. Figures are measured
against the baseline, the CRDs the team wrote by hand, on a corpus built by deleting those
resources and regenerating them from their protos.*

## Why this work exists

A generated CRD used to differ from a written one in four ways: fields in the wrong section, fields
that should be references and were plain strings, parents named by a template rather than by the
API, and fields that were simply absent. The first three are visible to a reviewer. The fourth is
not, and that is the dangerous one: a field absent from a CRD cannot be reported as missing from it
by any check we have.

So the work went in two directions at once. Read more of what the proto states, so fewer decisions
are left to a person. Then record every remaining decision, so the ones that are left are countable.
Across 275 resources, generation now reproduces 91.4% of baseline fields in place, and three
quarters of the differences carry an entry telling someone to look.

## Fields in the wrong section

A field the API only ever returns belongs in `status.observedState`. The generator reads
`field_behavior: OUTPUT_ONLY` for this, which works until a proto carries no annotations at all, as
discovery-derived ones do, or states it only in prose.

Three changes cover that ground. A short allowlist of server-computed names, `createTime`, `etag`,
`selfLink` and a few more, moves to `observedState` when a message carries no `field_behavior`
anywhere, and each move is queued as a guess. A comment saying "Output only" with no annotation
behind it is reported rather than acted on. And a resource whose `observedState` comes out empty is
named in the queue, which is otherwise invisible: the struct is written either way, and every later
check sees a resource that looks finished.

The detector for the prose case is the clearest illustration of where the value sits. It tested for
`"Output only."` and missed `[Output Only]`, the spelling Compute uses, which covers 1,605 fields in
`compute.proto` alone.

## Fields that should be references

`google.api.resource_reference` is exact but rare. Taking it alone, the queue named 11 of the 111
reference fields one run needed. The generator now applies three further rules to every Spec field,
at any depth: a resource-name template in the field's description, looser prose such as "the
resource name of", and a short list of well-known names such as `network` and `kmsKeyName`. Together
they name 82 of 111. On another run, 263 references were generated as plain strings, and 246 of them
were flagged.

None of these rules turns a field into a reference. They file an entry naming the likely target,
because a wrong reference is harder to catch in review than an absent one.

One rule deserves singling out. If a service declares a resource called `DataStore`, a string field
named `dataStore` is probably a reference to it. That reads the service's own resource list, needs
no vocabulary anyone maintains, scores 77% against the baseline, and gets stronger as more resources
in the service are generated.

The same rules run in the API checks. `refs.Classify` decides both what `TestMissingRefs` reports
and what the generator files, from one copy, so the check and the queue cannot drift apart.

## Parents, roots and locations

The old scaffold assumed every resource lived at `projects/{project}/locations/{location}`. It wrote
`projectRef` and a required `location` whatever the resource's name looked like.

The generator now reads `google.api.resource` instead. The root becomes `projectRef`,
`organizationRef` or `folderRef` as the pattern says, or a looked-up reference for a root such as
Analytics' `properties/{property}`. The direct parent becomes a reference when exactly one type
matches, and a queue entry naming the parent path when none does or several do. A `location` is
written only when the name contains one, and each one is queued, because a reference to the parent
may carry it instead.

Guesses here are sometimes wrong in useful ways. Chronicle's `instances/{instance}` matched the
shared `SQLInstanceRef` on a name suffix. The field is marked `+kcc:guess` and queued, which is
exactly the hand-off the queue exists for.

## Fields that are simply absent

Everything above still leaves a gap, and the point of the queue is to make it a list.

| | fields | share |
|---|---|---|
| produced as the baseline has it | 10,966 | 91.4% |
| produced differently | 369 | 3.1% |
| not produced | 665 | 5.5% |

281 of the 369 differences carry an entry. 62 of the 665 absences are shapes we model differently on
purpose, leaving 603 to close. Anything marked `+kcc:guess` has a matching entry, 295 of each on
this corpus, machine-checked; the invariant broke three times during the work and the check caught
each one. While a Kind has entries, `TestMissingRefs` suppresses its findings and carries its
existing ones forward, so a half-finished resource can merge without hiding work.

## What this does not settle

The corpus is what the team chose to implement, not a sample of what is left. Adding 44 more
resources to it dropped the same generator from about 94% to near 64% on the new ones, so the
headline is a property of the corpus as much as of the code.

First-pass output is not production quality: references are flagged rather than resolved, and there
are no fixtures or MockGCP coverage. The open questions for the team are whether a 5.5% absence rate
is a reasonable place to start, what must land in the same change as a generated resource, and how
the per-resource review the queue assumes is going to scale, since that review, not generation, is
now the limit.
