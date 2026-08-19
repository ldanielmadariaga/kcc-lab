# Identity collection segments GCP has never seen

**Status:** found by a proto-based audit (9 resources); a check ratcheting **4** of them is in PR
#18, open. None of the nine is fixed — changing a segment changes a published `status.externalRef`,
so they are recorded rather than corrected.

**Scope:** the experimental sandbox (`kcc-lab`). The defect itself is in code on upstream master.

## What is wrong

`template/apis/identity.go` builds a resource name by lowercasing the proto message name and
appending `s`:

```go
return i.parent.String() + "/{{.ProtoMessageName | ToLower}}s/" + i.id
```

GCP collection segments are camelCase. Auditing the 90 `_identity.go` files that can be matched to a
declared `google.api.resource` pattern finds **9 mismatches, every one of them casing**:

`apphubdiscoveredservice`, `apphubdiscoveredworkload`, `bigquerydatapolicy`,
`clouddmsconversionworkspace`, `dataprocnodegroup`, `discoveryenginedatastore`,
`managedkafkaconsumergroup`, `netappbackupvault`, `storagemanagedfolder`.

The segment goes into a **GCP resource name** — `status.externalRef`, and the format
`ParseXExternal` accepts in `spec…Ref.external` — so it has to match the API byte for byte.

**This is not KRM naming.** Two different namespaces, and only the second is affected:

| | Example | Set by | Covered by |
|---|---|---|---|
| KRM field name | `spec.forwardingRules` | `GetJSONForKRM` | `TestCRDsAcronyms` |
| GCP collection segment | `projects/p/locations/l/lbTrafficExtensions/x` | identity template | `TestIdentityCollectionCasing` (PR #18) |

## The evidence

Checked against an oracle independent of the proto — recorded GCP traffic under
`pkg/test/resourcefixture/testdata/` — camelCase is what GCP actually exchanges: `backupVaults` 27
files to 0, `nodeGroups` 6 to 0, `conversionWorkspaces` 5 to 0.

`StorageManagedFolder` is the clearest case. GCP's own response body, in `_http.log` after the `OK`
marker — a **response**, not one of KCC's requests, which is the distinction that matters when
citing this file as evidence:

```json
{
  "createTime": "2024-04-01T12:34:56.123456Z",
  "metageneration": "1",
  "name": "projects/_/buckets/bucket-${uniqueId}/managedFolders/managedfolder-${uniqueId}/",
  "updateTime": "2024-04-01T12:34:56.123456Z"
}
```

The normative source agrees and does not depend on a recording at all:
`+kcc:spec:proto=google.storage.control.v2.ManagedFolder`, whose `google.api.resource` pattern
declares `managedFolders`.

Against that, three sources inside KCC disagree with GCP and with each other:

| Source | Format |
|---|---|
| GCP (response above, and the proto pattern) | `projects/_/buckets/{b}/managedFolders/{f}/` |
| `ParseManagedFolderExternal` (`_identity.go:107`) | requires `buckets` and `managedfolders`, exactly 6 tokens |
| `ManagedFolderRef` doc comment (`_reference.go:34`) | `projects/{p}/locations/{l}/managedfolders/{id}` |

The doc comment says `locations` where the parser demands `buckets`; both disagree with GCP on
casing. GCP's trailing slash makes `strings.Split` yield seven tokens, so a name copied from GCP
fails the length check before casing is even reached.

## Why nothing caught it

Reconciliation still works: the API calls are built elsewhere and get the casing right. What is
wrong is the string users are handed and the one the parser accepts. The path is user-reachable
through `NormalizedExternal` (`_reference.go:53`), which parses a user-supplied `external:` value
and returns the error.

The writer (`_identity.go:34`) and the parser (`:107`) share the same wrong constant, so KCC
round-trips its own value and disagrees only with the outside world. Nothing compared any pair:
`TestCRDsAcronyms` checks acronym casing in CRD *field* names, `shortname_pluralization.txt` checks
CRD `shortNames`, and `naming_violations.txt` checks *file* naming. `naming_test.go:67` mentions
`_identity.go` only as a filename suffix.

## What the check covers, and what it does not

PR #18 adds `TestIdentityCollectionCasing`, which extracts each Identity's collection segment and
compares it against segments seen in recorded traffic, ratcheting through
`testdata/exceptions/identity_collection_casing.txt`.

It reads recorded traffic rather than the proto because `.build/googleapis.pb` is gitignored and
absent in CI, so a proto-based check would silently skip there. `_http.log` is committed — 894
files, scanned in about 0.3s.

The cost of that choice is stated in the test: the logs hold KCC's own requests as well as GCP's
responses, so a wrong casing is caught only when some other part of KCC gets it right. A resource
wrong *everywhere* slips through, and resources without fixtures are not covered at all. That is why
the baseline lists **4** — `netappbackupvault`, `discoveryenginedatastore`, `storagemanagedfolder`,
`dataprocnodegroup` — where the proto-based audit finds 9.

`TestIdentityCollectionRegex` guards the anchor on `parent.String()`. Without it the regex matches
the `/locations/` inside the Parent's own `String()` first, so every resource appears to use
`locations` — whose casing is right everywhere — and the check passes while verifying nothing.
Removing the anchor turns 9 real findings into 46 bogus ones in one direction and silence in the
other, which is why the guard is a test rather than a comment.

## What remains

- **Five of the nine are unratcheted**, because they have no recorded fixtures. Closing that gap
  needs either fixtures or a proto-based check that can run without `.build/googleapis.pb`.
- **None of the nine is corrected.** Fixing a segment changes a published `status.externalRef`, so
  it needs a decision about whether the old lowercase form must still parse — a parser that accepts
  both and a writer that emits only camelCase is the compatible shape.
- **New resources are covered from the generator side** by phase 2 of
  [`greenfield-generator-mechanics.md`](../greenfield-generator-mechanics.md), which reads the
  declared `google.api.resource` pattern instead of guessing. That fixes the template, not these
  nine files: the scaffolder is guarded by an "already exists, skipping" check and never rewrites an
  existing resource.
