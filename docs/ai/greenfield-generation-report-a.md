# Generating CRDs that say what they could not decide

*Status: the generator changes described here are merged in the kcc-lab sandbox repository. Every
flag is off by default, so no existing resource changes until a service opts in.*

## Summary

KCC has direct controllers for 457 of 1,006 create-capable GCP resources, 45.4%. Reaching 80% means
about 348 more, which is more than a hand-written queue can absorb. The work reported here makes
`generate-types` produce most of a new resource's API from its proto, and record what it could not
decide in a per-service judgement queue.

Two numbers frame the result, both measured against the baseline: the CRDs the team wrote by hand.
Across a 275-resource corpus, generation reproduces 91.4% of baseline fields in the same place, and
of what it does not reproduce, three quarters of the differences are already flagged for a person to
look at. The remaining gap is a list, not a mystery.

## What the generator now derives

Each item below was previously a judgement someone made while reading the proto. Each is now taken
from something the API states, behind its own opt-in flag.

| From the proto | What the generator does |
|---|---|
| `field_behavior: REQUIRED` | writes `+required`, and splits a shared nested type so status never inherits it |
| `field_behavior: OUTPUT_ONLY` | fills `status.observedState` rather than the Spec |
| `google.api.resource` pattern | names the root (`projectRef`, `organizationRef`, `folderRef`), the direct parent, and a `location` only when the name has one |
| `google.api.resource_reference` | files the target type for a field that should be a reference |
| map and acronym shapes | generates `map<string, Message>` fields, and cases plural acronyms as KRM wants |

The effect is that the same proto produces the same API every time, and the part left to a person
shrinks to the decisions the proto genuinely does not settle.

## What it flags

The rest of the work is detection. A generated field that might be wrong is worth far less than a
generated field that says so, because a field absent from a CRD cannot be reported as missing from
it by any later check.

Everything the generator infers rather than reads lands in
`apis/<service>/needs_judgement_call.txt`, one file per service, one line per decision: a reference
suggested by a field's description or name, a parent whose reference type does not exist yet, a
field it could not type, a resource whose `observedState` came out empty, a location a parent
reference could carry instead.

Three properties make the queue usable rather than decorative:

- Anything marked `+kcc:guess` in the generated types has a matching queue entry. On the corpus that
  is 295 markers and 295 entries, machine-checked. The invariant broke three times during this work
  and the check caught each one.
- The API checks read the queue. While a Kind has entries, `TestMissingRefs` suppresses its
  reference findings and carries its existing ones forward, so a half-finished resource can merge
  without either hiding work or resetting the ratchet. Clearing the entries graduates the resource.
- The rules behind the queue are the same code the checks run. `refs.Classify` decides what
  `TestMissingRefs` reports and what the generator files, so the two cannot drift apart.

## Evidence

Measured against the baseline CRDs, on the corpus built by deleting hand-written resources and
regenerating them:

| | fields | share |
|---|---|---|
| produced as the baseline has it | 10,966 | 91.4% |
| produced differently | 369 | 3.1% |
| not produced | 665 | 5.5% |

Of the 369 differences, 281 carry a queue entry. Of the 665 absences, 62 are shapes we model
differently on purpose, leaving 603 as the gap to close.

References are the category that moved most. Taking only `google.api.resource_reference`, the queue
named 11 of the 111 reference fields one run needed. Applying the description and name rules as
well, it names 82 of 111. On another run, 263 references were generated as plain strings and 246 of
them were flagged.

Two detector results are worth keeping for the pattern they show. A resource named `DataStore` in a
service makes a string field called `dataStore` a likely reference to it, which scores 77% against
what the baseline did and needs no vocabulary to maintain. And the output-only detector missed
`[Output Only]`, the spelling Compute uses, which covers 1,605 fields in `compute.proto` alone.
Reading what the API supplies beats remembering what someone noticed.

## Limits

The corpus is the set of resources the team chose to implement, so it is not a random sample of
what remains. When 44 more resources were added to it, the same generator scored near 64% on them
against 94% on the original set, which says the headline number is a property of the corpus as much
as of the generator.

First-pass output is not production quality. References are generated as plain strings and flagged
rather than resolved, and there are no fixtures or MockGCP coverage. Some parent guesses are wrong
in instructive ways: a Chronicle instance matched the shared `SQLInstanceRef` on a name suffix, and
the queue entry is what a reviewer needs to reject it.

## What we would like decided

1. Is a 5.5% absence rate, with three quarters of the differences flagged, enough to start
   generating against the unimplemented resources?
2. What must land with a generated resource upstream: typed references, fixtures and MockGCP in the
   same change, or in later passes?
3. The queue assumes a person reviews each resource before it graduates. That review is the
   throughput limit on the whole approach, and it is worth designing rather than inheriting.
