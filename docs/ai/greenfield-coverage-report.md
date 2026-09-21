# How much of a CRD the generator now writes

*The generator changes reported here are merged in the kcc-lab sandbox repository. Every one sits
behind a flag that is off by default, so no existing resource changes until a service opts in.*

## Summary

KCC has direct controllers for 457 of the 1,006 create-capable GCP resources, or 45.4%. Reaching
80% means about 348 more, and nobody is going to write 348 CRDs by hand.

Before this work, `generate-types` handed whoever picked up a new resource three fields:
`projectRef`, `location` and `resourceID`. A person then read the proto and decided, field by
field, what belonged in the Spec, what belonged in `status.observedState`, what was required, what
should point at another resource, and what to call it. The generator now writes 91.4% of the fields
a person wrote, and it records the decisions it could not make in one file per service that both a
reviewer and the API checks read.

Coverage and flagging need different work. Coverage is how much of the API a new resource carries
on the day it is generated, and flagging is whether anyone learns about the rest. A field absent
from a CRD cannot be reported as missing from it by any check we have, so an unflagged gap stays
invisible for as long as the resource exists.

## How coverage is measured

There is no way to score a generator on a resource nobody has implemented, because nothing exists
to compare its output against. So we ran it the other way round: delete 275 resources the team
implemented by hand, regenerate them from their protos, and compare field by field. The
hand-written CRD is the known-good answer, and every place the generated one differs names a
decision a person made and the generator did not.

Each field in each baseline CRD lands in one of three states, so the columns sum and nothing hides
in a residue.

| | fields | share | |
|---|---|---|---|
| produced as the baseline has it | 10,966 | 91.4% | same field, same path |
| produced differently | 369 | 3.1% | in the other section, under another name, or as a string where the baseline has a reference |
| not produced | 665 | 5.5% | nothing at that path at all |

A resource generated today therefore starts with most of its API already written. The last two rows
need different fixes: a field produced in the wrong shape needs a rule that notices it, and a field
produced nowhere needs a rule that writes it.

## Where the coverage came from

### The Spec comes from the proto

`--prepopulate-spec` replaces the three-field stub with the proto's own fields. Everything the proto
does not mark output-only goes to the Spec, everything it does goes to `status.observedState`.
Almost all of the 91.4% is this one change, and the rest of the work is about the fields it puts in
the wrong section, spells the wrong way, or cannot type.

### Fields land in the right section

`field_behavior: OUTPUT_ONLY` decides placement, which works until a proto carries no annotations
at all. Protos derived from a discovery document carry none, and Compute states it in prose
instead. Three flags handle the protos that do not annotate:

- `--place-server-set-fields` moves a short allowlist of server-computed names, `createTime`,
  `etag`, `selfLink` and a few more, into ObservedState when the message annotates nothing, and
  queues each move as a guess.
- `--detect-output-only-in-comments` reports a field whose comment says "Output only" with no
  annotation behind it. Its first version tested for `Output only.` and missed `[Output Only]`, the
  spelling Compute uses, which covers 1,605 fields in `compute.proto` alone.
- `--detect-empty-observedstate` names a resource whose ObservedState came out empty. The corpus
  has 36 of them, `ComputeInterconnect` among them, where the baseline carries 19 observed fields.
  Nothing else notices this, because the struct is written either way and every later check sees a
  resource that looks finished.

### Fields the generator used to drop

`--emit-message-maps` generates `map<string, Message>` fields, which the generator dropped with a
`// TODO:` in the source. `--emit-plural-acronyms` writes `relatedURIs` where the
generator wrote `relatedUris`, as KRM conventions want, and `generate-mapper` takes the same flag
so the mapper agrees with the types. A field the baseline spells differently counts against us even
though we emit it, so naming is coverage rather than tidying.

### Parents, roots and locations

The old scaffolder assumed every resource lived at `projects/{project}/locations/{location}`, and
wrote `projectRef` and a required `location` whatever the resource's name looked like. The
generator now reads `google.api.resource`. The root becomes `projectRef`, `organizationRef` or
`folderRef` as the pattern says, or a reference looked up by name for a root such as Analytics'
`properties/{property}`. The direct parent becomes a reference when one type matches, and a queue
entry naming the parent path when none does or several do. A `location` field is written only when
the resource's name contains one.

The generator used to read the parent out of the message name rather than the pattern. The two
disagree for 752 of the 1,417 messages that carry the annotation.

## What the generator flags

Everything the generator infers rather than reads goes to
`apis/<service>/needs_judgement_call.txt`, one line per decision, and anything marked `+kcc:guess`
in the types file has a matching entry. On the corpus that is 295 markers and 295 entries, checked
by a script rather than by review. The invariant broke three times during this work and the check
caught each one, including markers written into commented-out blocks describing code the generator
never produced.

We measure detection by counting the fields that differ from the baseline with nothing saying so.
On a 189-resource corpus that count went from 450 to 92.

| | unflagged | |
|---|---|---|
| first measured | 450 | |
| de-duplicate repeated fields and reference children | 412 | measurement fix |
| apply the reference rules to every field | 276 | the rules existed and had never been run here |
| pair the suffix the baseline drops when it adds `Ref` | 256 | measurement fix |
| gate the queue per Kind rather than per service | 259 | costs 3, and stops a ratchet being pruned |
| flag an empty ObservedState | 217 | 36 resources, one line each |
| name every parent segment left out | 203 | |
| separate renamed references from undetected ones | 109 | measurement fix |
| apply the reference rules inside map values | 97 | |
| two name rules, each measured before it was added | 92 | `secret` and `project`, four hits and no misses each |

Three of those corrected the measurement rather than the output. They stay in the table because the
number they corrected had already been published.

## References, the largest category

`google.api.resource_reference` states the target and is rare. One run needed 111 reference fields;
with only that annotation to go on, the queue named 11 of them. `--emit-reference-hints` applies
three more rules to every Spec field at any depth: a resource-name template in the field's
description, looser prose such as "the resource name of", and a short list of well-known names
such as `network` and `kmsKeyName`. Together they name 82 of the 111. On another run, the
generator wrote 263 references as plain strings and flagged 246 of them.

None of these rules turns a field into a reference. Each files an entry naming the likely target,
because a wrong reference is harder to catch in review than an absent one.

We decide whether to add a rule by asking whether it reads something the API supplies or remembers
something a person noticed. The sibling rule reads the service's own resource list, so a string
field called `dataStore` in a service that declares a `DataStore` resource is queued as a reference
to it. It scores 77% against what the baseline did, needs no list anyone maintains, and gets
stronger as more of a service is generated. We rejected a rule matching field names alone: it
produced 2,164 findings where the description rules produce 78, because a name like `network`
recurs across unrelated services.

`refs.Classify` decides both what `TestMissingRefs` reports and what the generator files, from one
copy, so the check and the queue cannot drift apart. While a Kind has queue entries,
`TestMissingRefs` suppresses its findings and carries its existing `missingrefs.txt` entries
forward, so a half-finished resource can merge without the ratchet reading the queue as a fix.

## Limits

The corpus is the set of resources the team chose to implement, not a sample of what is left. When
44 resources that had been missing from it were added, the same generator scored near 64% on those
against 94% on the original set. Read 91.4% as a ceiling.

First-pass output is not production quality. References are flagged rather than resolved, and a
generated resource has no fixtures and no MockGCP coverage. That is deliberate in a sandbox with no
users, and upstream would want a different bar.

These figures were measured on the experiment branch, against baseline `c1df0b9326`, with that
branch's generator. The changes have since landed as about twenty pull requests, each checked with
its flags off against a master build, but nobody has rerun the corpus against master.

## What we would like decided

1. Is 5.5% of fields not produced, with three quarters of the differences flagged, a reasonable
   place to start generating resources nobody has implemented?
2. What has to land with a generated resource: typed references, fixtures and MockGCP in the same
   change, or in later passes?
3. The queue assumes a person reviews each resource before it graduates, which makes that review
   the limit on throughput rather than generation. It is worth designing rather than inheriting.
