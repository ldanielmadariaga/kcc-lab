# Deterministic CRD Generation with a Judgement Queue

## 1. Overview

About 550 GCP resources have no Config Connector implementation. Writing each CRD by hand does not
scale, so this design lets `generate-types` produce most of a new resource's API mechanically and
write down, per service, every decision it could not make. A person then works through that list
instead of reading the whole proto.

Two pieces make this work. The generator now reads more of what the proto states, such as
`field_behavior` annotations and `google.api.resource` patterns, so less is left to whoever finishes
the resource, and what is left comes out the same every time. And it records everything it guessed
or left out in `apis/<service>/needs_judgement_call.txt`, the judgement queue, which the API checks
read as well. Generation becomes the first step and judgement the second, and neither has to wait
for the other.

Every change is opt-in, behind a flag or behind `--prepopulate-spec`, apart from fixes to output that
would not otherwise compile. With the flags off, output is byte-identical to what the generator
produced before, so no existing resource changes.

## 2. Problem Statement

Before this work, `generate-types` scaffolded a new resource as a stub: `projectRef`, `location` and
`resourceID` in the Spec, and nothing in its ObservedState. A person copied fields from the proto by
hand, decided which ones belonged in status, which were references, and which were required, and
fixed names to KRM conventions. That work is slow, and it leaves no record of the decisions made.

Three problems block doing this in bulk:

1. The output depended on the person. The scaffolded Spec did not use the proto at all, and the
   generated nested types ignored `REQUIRED` annotations, plural acronyms and message-valued maps.
   The resource's name pattern played no part in its parent fields.
2. Nothing recorded what the generator could not decide. A field it could not type was dropped with
   only a `// TODO:` in generated source, and a field that should have been a reference looked like
   any other string. An omission is invisible to every later check, because a field absent from a
   CRD cannot be reported as missing from it.
3. The checks punished half-finished resources. `TestMissingRefs` ratchets `missingrefs.txt`, so a
   mechanically generated resource with reference-shaped strings could not merge until every
   reference was modelled.

## 3. Goals

- The same proto and flags always produce the same types.
- Generated coverage close to what a person writes by hand.
- Every guess and every omission recorded in the judgement queue.
- No change for any existing resource unless its service opts in.
- One copy of each rule, shared by the generator and the API checks.

## 4. Non-Goals

- Generating controllers, mappers beyond field naming, fixtures or MockGCP.
- Turning fields into references on a heuristic. The generator proposes references; a person decides.
- Changing resources that already exist. The flags are meant for new resources.
- Changing upstream's conventions. This work lives in the kcc-lab sandbox repository.

## 5. Proposed Design

### 5.1 Principle

The generator emits what the proto states and records what it does not. Where the proto is
explicit, such as a `REQUIRED` annotation or a resource pattern, the generator acts on it. Where it
has to infer something, such as a reference from a description or a parent type from a name, it
still emits its best output, marks it `+kcc:guess` where the output is a guess, and files a queue
entry. Emitting a field with an open question beats omitting it, because an omission cannot be
found later.

### 5.2 The judgement queue

Each service has one file, `apis/<service>/needs_judgement_call.txt`. A line looks like:

```
kind=NetworkServicesLBTrafficExtension group=networkservices.cnrm.cloud.google.com: field ".spec.location" reason=location-or-parent-ref (parent.ProjectAndLocationRef in apis/common/parent, inlined, could replace projectRef and location; ...)
```

- Entries are keyed by Kind and group, because the queue is written before the CRD exists.
- `generate-types` merges into the file rather than replacing it. `generate.sh` often calls it more
  than once per service, and replacing the file dropped earlier calls' entries.
- Findings in a nested message, which several Kinds may share, are written as `#` comment lines,
  since there is no single Kind to name.
- `TestJudgementQueueIsWellFormed` checks the format of every queue in the tree.
- While a Kind has any entry, `TestMissingRefs` skips its `[refs]` findings, and carries its existing
  `missingrefs.txt` entries forward so queueing never looks like a fix. Clearing a Kind's entries
  graduates it, and the ratchet applies as usual.

The workflow is two steps. First, generate the resource with the flags on, which gives compiling
types and a queue. Second, a person works through the queue: accept or rewrite each guess, add
references, move fields, and delete the entry.

### 5.3 What the generator now reads from the proto

| Flag | What it does | Queue reasons it files |
|---|---|---|
| `--prepopulate-spec` | Fills the Spec from the proto's non-output fields and ObservedState from its output-only fields, instead of the three-field stub. Names the root of the resource (`projectRef`, `organizationRef`, `folderRef`, or a looked-up reference), and writes `location` only when the name has one. | `untriaged-bulk-generation`, `possible-reference`, `unsupported-field-type`, `observedstate-identity-field-omitted`, `root-ref-guessed`, `root-ref-not-modelled`, `location-or-parent-ref`, `location-parent-unknown` |
| `--emit-required-from-proto` | Writes `+required` for `REQUIRED` fields, and splits a nested type into Spec and ObservedState variants so status never becomes required. | |
| `--emit-plural-acronyms` | Cases plural acronyms as KRM wants, so `related_uris` becomes `relatedURIs`. Also a `generate-mapper` flag. | |
| `--emit-message-maps` | Generates `map<string, Message>` fields instead of dropping them. Also a `generate-mapper` flag. | |
| `--place-server-set-fields` | Moves a short allowlist of server-computed fields (`createTime`, `etag`, `selfLink`, ...) into ObservedState when the message carries no `field_behavior` at all, as discovery-based protos do. | `server-set-field-placed` |
| `--detect-output-only-in-comments` | Reports fields whose comment says "Output only" but whose proto has no annotation. | `output-only-in-comment-only` |
| `--emit-parent-refs` | Adds one Spec field referencing the direct parent, when exactly one reference type matches it. | `parent-ref-guessed`, `parent-ref-not-modelled` |
| `--emit-sibling-refs` | Marks a string field whose name matches a Kind in the same service. | `possible-reference-by-sibling` |
| `--detect-empty-observedstate` | Records a resource whose ObservedState came out empty, usually because its proto marks no field `OUTPUT_ONLY`. | `empty-observedstate` |
| `--emit-reference-hints` | Checks every Spec field, at any depth, against the reference rules below. | `possible-reference-by-description`, `-by-description-loose`, `-by-name` |

Four smaller fixes keep the output consistent. Observed-state structs are written with the caller's
options, so a field is spelled the same in Spec and status. Queue paths use the same options as the
emitted names. A field the generator already writes as a reference type gets no reference hint. And
a new types file leaves out `<Kind>GVK` when another file in the package already declares it, as
some hand-written reference files do, so the package still compiles.

### 5.4 Reference rules

References are the hardest decision and the largest category in the queue. The rules live in
`dev/tools/controllerbuilder/pkg/refs`, and two callers use them:

- `TestMissingRefs` applies `Classify`, the strict rule: a resource-name template in the
  description, a service-account name, or a Cloud Storage bucket.
- `--emit-reference-hints` applies `Classify` too, then two looser rules that only ever hint:
  `MatchDescriptionLoose`, for prose such as "the resource name of", and `MatchName`, a short list
  of names such as `network` and `kmsKeyName`.

The loose rules stay out of `Classify` because `missingrefs.txt` is a ratchet: a hint there would
block merges until someone modelled the reference.

The main module resolves the generator module from the working tree through a `replace` directive
in `go.mod`, as it already does for `mockgcp`, so `tests/apichecks` can import `pkg/refs`.

### 5.5 Parent, root and location

The resource pattern decides the Spec's parent fields:

- The root: `projectRef` for `projects/`, `organizationRef` and `folderRef` for those roots, and for
  any other root, such as `properties/{property}`, a reference type looked up by name, or a queue
  entry when none matches.
- The direct parent, with `--emit-parent-refs`: a reference when exactly one type matches, otherwise
  a queue entry naming the parent path.
- The location: a required `location` whenever the name has a `locations`, `regions` or `zones`
  segment, and none otherwise. Each location written also files `location-or-parent-ref`, because a
  reference could carry it: `parent.ProjectAndLocationRef` for a project/location parent, or a
  reference to the parent for a nested resource.

## 6. The Experiment

The design was tested by deleting resources that upstream wrote by hand, regenerating them from the
proto, and comparing the result with upstream's version, which serves as the known-good answer. The
figures below come from the experiment branch's `greenfield-state-of-play.md` and
`greenfield-coverage-invariant.md`. They were measured with the generator on PR #21, before the
work was split into the PRs above, so they are evidence for the approach rather than a measurement
of master.

- Over a 275-resource corpus, 91.4% of upstream's fields were produced as upstream has them, 3.1%
  were produced differently, and 5.5% not at all.
- A smaller corpus had scored 94.2%. Nothing in the generator changed; the corpus did. Totals from
  different corpora cannot be compared.
- All 295 `+kcc:guess` markers had a matching queue entry.
- With only `google.api.resource_reference` to go on, the queue named 11 of the 111 reference
  fields a 239-resource run needed. Applying the reference rules as well, then in the second-pass
  `scripts/queue-hints` and now through `--emit-reference-hints`, it named 82 of 111.
- Of 263 references the generator left as plain strings on a 189-resource run, 246 were flagged.

Two bugs came out of making every guess reach the queue: the queue file was truncated by each
`generate-types` call after the first, and parent fields were emitted without their entries.

## 7. Implementation Plan

The work landed as a series of small PRs on the kcc-lab repository. Each generator change is either
off by default or only changes output that would otherwise be wrong, and each was checked against a
master build with its flags off:

- Placement and naming: #28 (prepopulation), #31 and #44 (plural acronyms), #32 (server-set
  fields), #46 (message maps), #54 and #55 (consistent options).
- References and parents: #51 (parent refs), #52 (siblings), #56 and #57 (shared rules and
  reference hints), #58 (root refs), #60 (no hint on generated references), #61 (`replace`
  directive), and #63 (shared refs lookup), which is still in review.
- The rest: #53 (empty ObservedState), #59 (duplicate GVK), #62 (location).

No service's `generate.sh` turns these flags on yet. The next step is an illustrative PR that
generates one new resource, ChronicleWatchlist, with every flag on.

## 8. Alternatives Considered

- A second pass over generated CRDs (`scripts/queue-hints`). It worked, but it was easy to forget,
  and a regeneration silently discarded its output. Moving the rules into the generator removed it.
- Detecting references by field name alone. It produced 2,164 findings against 78 for the
  description rules, because a name like `network` recurs across unrelated services.
- A field for every segment of the resource pattern. The corpus did not support it: per AIP-122 a
  reference to the direct parent carries the whole path.
- Moving an existing `<Kind>GVK` declaration into the new types file. It would mean the scaffolder
  editing hand-written reference files, so the scaffolder skips the declaration instead.
- Naming the location field `region` or `zone` after the pattern. `location` is the canonical name,
  and a manifest reader can tell it may hold a region or a zone.

## 9. Testing Strategy

- Unit tests for each rule and each flag, with table cases for every branch. Several were checked by
  breaking the code and confirming the test fails.
- For every PR, the generator run with the flags off on real protos and diffed against a binary
  built from master.
- `TestJudgementQueueIsWellFormed` for the queue format, and `TestMissingRefs` with its ratchet and
  golden files for the reference rules.
- Scaffolded packages compiled in a scratch directory; only the DeepCopy methods that controller-gen
  adds later are missing.

## 10. Known Gaps

- A nested resource still gets `projectRef` and `location` beside its parent reference. The cleaner
  model, where the parent reference carries the whole path as `AlloyDBInstance` does, needs
  `template/apis/identity.go` changed too, since it assumes a project/location parent.
- Shared reference types are matched by name suffix, which can pick a type from an unrelated
  service: Chronicle's `instances/{instance}` matches `SQLInstanceRef`. The guess is flagged, but
  still wrong.
- Some reference files written ahead of their target hard-code a `schema.GroupVersionKind` literal.
  A new types file relies on that declaration, stale or not.
- Some references no rule detects. `LbTrafficExtension`'s `forwardingRules` is one: upstream made
  it a reference, but its description names neither a template nor a resource name.
