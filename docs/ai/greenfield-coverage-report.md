# Bulk-generating greenfield resources

*What a generator can produce of a KCC resource, measured against the resources the team wrote by
hand. The generator changes reported here are merged in the kcc-lab sandbox repository, each behind
a flag that is off by default.*

## The question

KCC has direct controllers for 457 of 1,006 create-capable GCP resources, or 45.4%. Reaching 80%
would mean implementing about 348 more. At the rate a hand-written resource takes, that is a
multi-year queue, and the tail is made of resources nobody requests loudly enough to prioritise
individually.

The open question was how much of that surface a generator can produce rather than a person. This
report answers it with two numbers. The generator reproduces 91.5% of the fields a KCC engineer
wrote by hand, at the same path, and it produces nothing at all for 0.3%. Nearly everything in
between is a field it does emit, in the wrong section, under another name, or as a plain string
where the hand-written version has a reference.

What follows is the method, the measurements, and what they leave open.

## Results

The starting point was not close. Before `--prepopulate-spec`, `generate-types` scaffolded three
fields into the Spec, `projectRef`, `location` and `resourceID`, and left ObservedState empty.
Running `generate.sh` on a new resource gave you a stub, and a person read the proto and wrote
everything else, including every field the API only ever returns.

Measured now, over 275 resources the team had implemented by hand:

| of 13,566 baseline fields | fields | share |
|---|---|---|
| reproduced at the same path | 12,414 | 91.5% |
| diverge, flagged for a person | 422 | 3.1% |
| diverge, nobody is told | 730 | 5.4% |

The middle row is the one that does not show up in a coverage number. No generator decides
everything on its own, so what matters is what happens to the rest. Ours writes a note in the code
and adds the field to a per-service list for a person to rule on.

### The number that says what is left to do

Of the 730 nobody is told about, **37 are fields we produce nowhere at all**, 0.3% of the surface.
The rest we do produce, at another path, under another name, or in another shape. The largest
single group, 177 fields, names proto fields our pinned descriptor does not declare, which nothing
could have produced.

Finding the fields is close to solved. What a deterministic generator cannot decide on its own is
what to do with a field once it has found it: whether it should point at another resource rather
than hold a plain string, whether it belongs in `status` rather than `spec`, how an acronym in its
name should be capitalised. Those are judgement calls, and the generator hands most of them to a
person rather than guessing silently.

## Why we deleted upstream's work to test this

The obvious way to evaluate a generator is to run it on resources nobody has implemented and read
the output. That does not work: there is nothing to check the output against, and "looks plausible"
is not a measurement.

So we did the opposite. We took resources the team had already implemented by hand, deleted our
types files, regenerated them from the proto, and compared field by field against upstream's CRDs.
Upstream's version is the known-good answer. Every place the generated CRD differs from it is a
specific, countable way that mechanical generation falls short. A real engineer made each of those
hand-written choices, so the diff is a list of the judgements a generator cannot make.

The method comes with a second oracle for free. The hand-written `_identity.go` and `_reference.go`
files were left in place while the types beneath them were regenerated, so `go build ./apis/...`
fails wherever the generated types no longer satisfy upstream's own controller code, naming missing
fields in seconds. The general form is worth stealing: an existing implementation is a test oracle,
and deleting it is how you use it.

## Coverage

All 275 in-scope resources generate both a types file and a published CRD. Every field in every
baseline CRD lands in exactly one of three states, so the columns sum and no arithmetic can drift.

### What counts as one field

The score is a list of paths, and a path is not a field. Three reductions are defensible, they
answer different questions, so all three are published. The numerator is the same in every row;
only what counts as one missing thing changes.

| unit | reproduced | of | share |
|---|---|---|---|
| every missing path, raw scorer output | 12,414 | 15,083 | 82.3% |
| **references counted once**, the headline | 12,414 | 13,566 | **91.5%** |
| distinct defects, a missing subtree counted once | 12,414 | 13,207 | 94.0% |

KCC encodes a reference as an object with `external`, `name`, `namespace` and sometimes `kind`, so
one site a user fills in one way expands into five paths; nearly half of every missing path in the
corpus is a child of one. A repeated field is printed twice, as `foo` and `foo[]`. Both are
artefacts of the output rather than facts about the API, which is why the raw count is too big.

Collapsing a missing subtree to one entry goes too far the other way. It is the right unit for a
work list, since we do not generate `ComputeFutureReservation`'s `shareSettings.projectMap` at all
and fixing that one thing resolves twelve paths, but the collapse ratio runs from 1x on a flat
resource to 32x on `NetworkSecurityAuthzPolicy`, where 321 missing paths reduce to ten lines.
Somebody told "ten defects" has no way to know how much of upstream's API sits behind them.

So the headline counts a missing subtree in full, because a reviewer really is blind to all of it,
and a missing reference once, because it is one decision and one fix.

### The work list

The same divergence, counted as defects rather than fields. 793 things to fix is actionable in a way
1,152 fields is not.

| state | fields | share |
|---|---|---|
| implemented, the same field at the same path | 12,414 | 94.0% |
| discrepancy, we produce it but not as upstream has it | 456 | 3.5% |
| &nbsp;&nbsp;flagged for a second pass | 325 | 71% |
| &nbsp;&nbsp;nothing says so | 131 | 29% |
| missing, we produce nothing at all | 337 | 2.6% |
| &nbsp;&nbsp;a gap to close | 193 | |
| &nbsp;&nbsp;we model it differently on purpose | 144 | |
| &nbsp;&nbsp;flagged for a second pass | 47 | 14% |

The split is on what we produced, not on whether we mentioned it. A discrepancy is a field we do
emit, in the other section or under a different name or as a plain string where upstream has a
reference object. It breaks a user's YAML just as surely, but it needs detecting or moving rather
than generating, and a report that files it under "missed" sends people to write generators for
fields the types file already carries.

### What the silent fields actually are

"Silent" means the field differs from upstream and nothing in our output says so. It does not mean
the field is missing, and for the most part it is not. Reading each one back to the proto field
behind it:

| why it is silent | fields | |
|---|---|---|
| `not-in-our-proto` | 177 | the baseline names a proto field our pinned descriptor does not declare |
| `reference-shape` | 109 | upstream has a reference and we have no field at that stem |
| `renamed` | 77 | same field, different name |
| `map-shape` | 67 | a proto map; upstream renders it as a list or as named keys |
| `intentionally-different` | 60 | `protobuf.Value` arms, mapped whole to JSON |
| `snake-case` | 59 | upstream's CRD uses underscores where we use camelCase |
| `emitted-elsewhere` | 55 | we carry that proto field, at another CRD path |
| `moved` | 54 | we emit it, in the other section |
| `reference-not-detected` | 25 | we emit the field and never spotted it is a reference |
| the tail, four classes | 47 | `not-generated?` 35, `emitted-elsewhere?` 6, `crd-stale` 4, `not-generated` 2 |

Only the two `not-generated` rows are fields we produce nowhere: 2 confirmed against the proto, and
35 that rest on a leaf name with no proto field to confirm them. **37 fields, 0.3% of the surface.**
A row ending in `?` rests on that weaker test, which is what once made `crd-stale` report 46 where
the true count was 4.

Getting down to 37 took three corrections, each of which had been inflating it: `absent` was a
residual bucket holding fields we emit at another path, the stale-CRD class rested on a leaf-name
match that mostly caught fields present elsewhere in our own CRD, and 177 fields name proto fields
that do not exist in the descriptor we compile against.

That last one is worth stating plainly because it is not ours to fix. `apis/git.versions` pins a
googleapis commit, and the baseline pins the same one, but the baseline's checked-in types files are
older than that pin, generated against a newer googleapis and never regenerated. Running upstream's
own `generate.sh` today would delete those fields. `APIHubAPI` shows it cleanly: its `Api` message
emits everything through field 15 and nothing from 16 on, because fields 16 through 20 were added to
googleapis after the pin.

The 456 we produce differently are the real work, and they are detection rather than generation: a
reference we emit as a plain string, a field placed in Spec where upstream has it in ObservedState,
an acronym cased the other way.

## Flagging what needs judgement

Anything the generator cannot justify from the proto gets a `+kcc:guess` marker in the types file
and an entry in that service's judgement queue. There are 299 markers today, and a checker enforces
that none of them lacks an entry.

That covers what the generator knows it guessed. Measured against upstream instead, 793 fields
diverge in some way and 372 of them carry an entry, and the rate splits sharply by population:

| | total | flagged | |
|---|---|---|---|
| discrepancy, we produce it in the wrong shape or place | 456 | 325 | 71% |
| missing, we produce nothing | 337 | 47 | 14% |

A field we emit in the wrong shape is usually something the generator knew it was unsure about; a
field we emit nowhere is usually something nothing looked for. Quoting the two together as one rate
describes neither.

The 421 nobody is told about are not one problem. Four classes carry almost all of them:

| class | we emit it | flagged | |
|---|---|---|---|
| `reference-not-detected` | 264 | 239 | the detector is doing well here |
| `moved` | 80 | 37 | server-set fields the proto never annotated |
| `not-in-our-proto` | 57 | 58 | the baseline names fields our pinned descriptor lacks |
| `renamed` | 53 | 0 | casing only, and a fix rather than a judgement call |

`renamed` is the largest unflagged group and the cheapest: our acronym casing writes `bootDiskMIB`
where the baseline writes `bootDiskMiB`. Nothing about that needs a person's opinion.

Two limits worth stating. Of the 372 flagged, only 28 also carry a marker in the types file, so a
reader opening the type mostly sees a plain string with nothing to suggest it is unfinished. And 29
of the unflagged are references the baseline renamed rather than suffixed, `DatastreamPrivateConnection`'s
`vpc` against `networkRef` among them, which no name match bridges. Those are counted as unflagged,
which overstates the gap rather than flattering it.

The invariant broke three separate times during this work and the checker caught all three, in
places code review would not have looked: a marker written on every generator invocation while the
queue entry was written on only some; a merge that silently dropped every comment line of the
existing queue file; and markers emitted into commented-out blocks describing code the generator did
not produce. A rule nothing checks is a rule that regresses.

## Detection over prescription

Early on the instinct was to fix each coverage gap directly: see a missing field, write a rule that
produces it. That does not scale and it does not transfer, because you cannot enumerate the ways a
thousand APIs differ, and a rule learned from one service usually misfires on another.

The rule we settled on is detection over prescription. It is more valuable to reliably notice that a
field needs a human than to guess what the human would say. A flagged field is a fine outcome; a
field nobody was told about is not.

That reframing gives a test for whether a new rule is worth having. Does it derive its answer from
something the API supplies, or does it remember something a person once noticed?

| signal | source | verdict |
|---|---|---|
| `google.api.resource_reference` | the proto states it | fact |
| resource-name templates in a description | the field's own docs | derives |
| a sibling resource in the same service | the service's own resource list | derives |
| `refs.NameRules` | a list of known spellings | remembers |

The derived three work on a service nobody has looked at. The remembered one only ever finds
references somebody has already seen, which is why a growing `NameRules` list is a signal that one of
the other three is missing something, not a sign of progress.

**The sibling rule.** If a service declares a resource called `DataStore`, then a string field named
`dataStore` is probably a reference to it. It needs no vocabulary at all, and it gets stronger as
more resources are generated, because the service's own resource list is what it reads. Measured
against what upstream did, it runs at 77% precision, and nobody maintains it.

**A spelling the detector did not know.** The output-only detector tested whether a proto comment
opened with `Output only.` Compute writes `[Output Only]` instead, and that one unrecognised
spelling covers 1,605 fields in `compute.proto` alone. Every one of `ComputeInterconnect`'s
misplaced status fields, `googleIPAddress` and `circuitInfos` and `expectedOutages`, turned out to be
nothing more exotic than that. Two strings in a list, for a signal that reaches services nobody has
looked at.

## How non-deterministic behaviour is flagged

Some decisions cannot be derived from a proto. Whether a plural noun in a resource pattern names a
real KCC Kind, whether a field the proto never annotated is server-set, whether a string is a
reference: each of those needs a person. The generator's job on them is not to guess better, but to
make sure the guess is visible to whoever reviews it.

A field the generator cannot vouch for is still emitted, with the open question recorded separately.
Omitting it instead would hide it from every other check in the system, because a field absent from
the CRD cannot be reported as missing from it. A wrong field is a bug someone finds; a missing field
is a bug nobody finds.

### A marker and a queue entry, for every guess

Anything the generator cannot justify from the proto produces both a marker in the generated Go and
an entry in that service's `needs_judgement_call.txt`. Neither alone is enough. The queue is a work
list somebody clears; the types file is what a reader actually opens.

| marker | what the generator could not justify | now |
|---|---|---|
| `parent-location` | a location segment read off the resource pattern: is the resource regional, and is this the name for it? | 193 |
| `placement` | a field put in ObservedState by name, because the proto carries no `field_behavior` anywhere on the message | 41 |
| `parent-segment` | a name segment emitted as a plain string; upstream may want a reference | 26 |
| `possible-reference` | a field whose name matches a resource this service declares | 26 |
| `parent-ref` | a typed reference whose target was assumed from a collection segment | 13 |

### Even a typed reference is queued

An earlier version suppressed the queue entry whenever the generator emitted something concrete, on
the reasoning that a typed ref needs no review. That traded detection for an unflagged guess:
`BigtableCluster` got `Instance *string` where upstream has `spec.instanceRef`, and nothing said so.
The compiler was proving 32 fields missing while the queue named 2.

"Very sure" is not a state a generator can be in about a target it inferred from a plural noun.
Anything marked as a guess belongs in the queue, typed references included.

### When a finding belongs to no Kind

A nested message is shared by every resource that references it, so a finding against one cannot be
attributed to a single Kind. Every non-comment line in the queue suppresses `[refs]` findings for the
Kind it names, so inventing an owner would quietly switch off a real check. Those findings are
written as comments instead, and the tooling reads them back:

```
# possible-reference-by-sibling: google.cloud.compute.v1.NetworkInterface.subnetwork target=ComputeSubnetwork
# dropped: …TranslationTaskDetails.specialTokenMap reason=unsupported map type with key string and value enum
```

### The queue is also the gate

While a resource has entries, its `[refs]` findings are suppressed, so a half-generated resource does
not trip a ratchet that can only shrink. Clearing the queue is what graduates it. That also makes the
queue the throughput limit on the whole idea, which is a question for the team rather than an
implementation detail.

## Limitations

Read 91.5% as a ceiling, not a forecast. The corpus is upstream's choices, not a random sample. These
275 are resources the team judged worth implementing, and the unimplemented ones may be
systematically harder: less documented, odder shapes, or unimplemented precisely because someone
looked and found a problem. Adding 44 resources that had been missing from the corpus moved the
headline by nearly three points, which is direct evidence for that caution.

22 packages do not compile, and that is by design, for the reason given under the method above. About
16 of them are a measurement floor we chose not to chase.

First-pass output is not production quality. References are generated as plain strings and flagged
rather than resolved into typed refs, and there are no test fixtures or MockGCP coverage for
generated resources. The strategy is deliberately generate-first and retrofit in corpus-wide passes,
which is only safe because the sandbox has no users. Upstream would need a different bar.

---

*Every figure here comes from one run, `2026-09-04-275-resources-proto-first.txt`: 275 resources
against baseline `c1df0b9326`, measured by `hack/tools/greenfield/silence_report.py`. Earlier runs
are indexed in `docs/ai/experiments/measurements/README.md`, which marks the superseded ones; the
run before this one moved 80 fields between classes without changing a total, and the run before
that added 1,448 reproduced fields. Quote the index rather than a figure found in another doc.*

*The measurement was taken on the experiment branch. The changes have since landed as about twenty
pull requests, each checked with its flags off against a master build, and nobody has rerun the
corpus against master.*
