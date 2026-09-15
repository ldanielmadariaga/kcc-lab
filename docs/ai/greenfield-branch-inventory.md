# Branch and PR inventory for the greenfield experiment

The sandbox repo accumulated 25 local branches across the experiment, most of them
superseded, several near-duplicates of each other. This is what each one is, so nobody has
to re-derive it from commit subjects. Update it when you open or close a PR.

Repo: `ldanielmadariaga/kcc-lab` (remote `lab`). Worktree: `/home/luks/kcc-lab-exp`.
All three worktrees — `/home/luks/k8s-config-connector`, `/home/luks/kcc-lab`,
`/home/luks/kcc-lab-exp` — share **one** repository, so this is a single branch namespace.

## The current stack

Three PRs, deliberately layered. Read them in this order.

| PR | branch | what it is | files |
|---|---|---|---|
| [#21](https://github.com/ldanielmadariaga/kcc-lab/pull/21) | `greenfield-generator-and-tooling` | the generator, the measurement tools, the docs | 80 |
| [#26](https://github.com/ldanielmadariaga/kcc-lab/pull/26) | `greenfield-generate-sh-flags` | the `generate.sh` flags that switch #21 on | 98 |
| [#22](https://github.com/ldanielmadariaga/kcc-lab/pull/22) | `greenfield-healthy100` | the regenerated corpus — evidence, not a merge candidate | 1,452 |

#26 is based on #21, not on master. #22 contains everything.

**None of the three will be green, for three different reasons**, and the reasons matter:

* #21 hits a pre-existing `go vet` failure in
  `dev/tools/controllerbuilder/pkg/commands/iteratetypes` — an `fmt.Errorf` verb mismatch
  we never touched, present on `lab/master` too.
* #21 and #26 will fail `validate-generated-files`, because the generator changed and the
  corpus it regenerates is in #22.
* #22 fails because **22 packages do not compile by design**. Upstream's hand-written
  `_identity.go` files are left in place as an oracle; repairing them would destroy the
  measurement. Do not "fix" this.

### What is in #22 and what could still come out

The corpus branch is 1,452 files. By category:

| n | category | reviewable? |
|---|---|---|
| 245 | `<kind>_types.go` | generated |
| 242 | CRDs under `config/crds/` | generated |
| 102 | `types.generated.go` | generated |
| 93 | `needs_judgement_call.txt` | generated |
| 64 | mappers | generated |
| 47 | `_identity.go` | scaffolder output |
| 27 + 23 | `doc.go`, `groupversion_info.go` | scaffolder output |
| — | everything hand-written | now in #21 and #26 |

The hand-written half has been extracted. What remains is generated or scaffolded, so
further splitting means grouping generated output by service, not separating code from
data.

## Waiting on a decision: `generator-json-wellknown-types`

**This is the only branch that targets the real upstream repo**,
`GoogleCloudPlatform/k8s-config-connector`, not the sandbox. It tracks `upstream/master`
and is 2 commits / 63 files. Opening it is a public contribution that notifies KCC
maintainers, which is why it has not been pushed.

It also cannot sensibly go to the sandbox: it is based on upstream, so the diff against
`lab/master` would be meaningless.

### What it actually does

Two related generator gaps, both of which silently removed fields from CRDs.

**`google.protobuf.Value` and `ListValue` now map to `apiextensionsv1.JSON`**, joining
`Struct`. A CRD cannot express them any other way: `Value.list_value` is a `ListValue`
whose values are `Value`s, and `controller-gen` cannot build a terminating schema for that
recursion — generating the structs makes any package that reaches one fail to produce CRDs
*at all*. This is why `apis/aiplatform/v1alpha1/recursive_types.go` declares both by hand
with the recursive field commented out, and why Firestore's `Document.fields`, a
`map<string, Value>`, is hand-written as `map[string]apiextensionsv1.JSON`. The branch
makes the generator do what those hand-written files already do.

**`map<string, Message>` generates**, where before `GoTypeForField` declined every map
whose value was not a string or an int64. A declined field becomes a `// TODO:` comment and
never reaches the CRD, so this was a silent drop. The value struct is written like any
other nested message, and the field renders as `map[string]TheValueType` in the value form
rather than a pointer, which the corpus prefers 16 to 7.

Three supporting fixes travel with it:

* the **mapper generator** writes the map loop itself, instead of falling through to a
  hand-written `<Field>_FromProto` helper that in practice nobody wrote, leaving the
  generated mapper not compiling;
* `krmIsUnpointeredJSON` replaces a comparison against the literal string
  `"direct.Struct_FromProto"`, which silently covered only one of the proto messages
  mapping to that Go type, and neither of them inside a oneof;
* **`prunetypes` drops now-unused imports.** Commenting a type out can orphan an import,
  Go rejects an unused import outright, and nothing catches it later because `goimports`
  runs over `pkg/controller/direct` and never over `apis`. It parses rather than
  string-matches, because the commented-out type still spells the qualifier.

Plus real tests: maps had no coverage at all, including the two forms that already worked
— which matters more than usual here, since a regression removes fields from the CRD
without failing anything.

The same work landed on `greenfield-healthy100` in `cdce683be3`, with a write-up in
`8cdf535316`. This branch is the cleaned-up version prepared for upstream.

## Superseded, contained, or not ours

Nothing below needs a PR. Kept so the reasoning is not re-derived.

| branch | why not |
|---|---|
| `pr19-check`, `greenfield-replication-exp`, `repro-11964` | fully contained in #22 |
| `prepopulate-spec` | `lab/master` already has this via merged #17; the local branch is an **older** pre-merge variant, so a PR would propose reverting improvements. Push is rejected non-fast-forward, which is the clue |
| `ci-baseline` | already went up as PR #1 and was closed as a throwaway probe |
| `pr-11964` | 110 files of *someone else's* PR, checked out locally for the reproduction work in #8/#9. Not ours to propose |
| `refs-ratchet` | fully contained in [#23](https://github.com/ldanielmadariaga/kcc-lab/pull/23) |
| `refs-classifier-standalone` | the classifier without the ratchet; differs from #23 by `pkg/test/utils.go` and one line of `crds_test.go` |
| `refs-detector-hardening` | the same work as #23, differing only in a 37-line `SKILL.md` |
| `greenfield-observedstate` | the earlier 100-resource run, superseded by #22 at 231. Opened as [#25](https://github.com/ldanielmadariaga/kcc-lab/pull/25) for visibility |
| merged branches | `controllerbuilder-rebuild-if-stale`, `identity-collection-casing`, `judgement-queue`, `resource-annotation-scaffold`, `required-from-field-behavior`, `pilot-lbtrafficextension`, `greenfield-coverage-strategy`, `greenfield-checker-tests`, `greenfield-dropped-fields`, `greenfield-enforcement`, `sandbox-master` |

**[#19](https://github.com/ldanielmadariaga/kcc-lab/pull/19) is still open** and is an
ancestor of #21/#22. It is the Step 1 runbook, which #21 substantially rewrites — worth
closing in favour of #21.

## Two junk files, now gitignored

Both were committed by a `git add -A` and both are build output:

* `apis/config/crd/` — 331 files. `dev/tasks/greenfield-regenerate` runs `controller-gen`
  from inside `apis/` with `output:crd:artifacts:config=config/crd/`, so raw CRDs land
  there before being post-processed and published to `config/crds/resources/`.
* `dev/tools/crd-mcp-server/crd-mcp-server` — a 12MB ELF. The tooling invokes
  `./bin/crd-mcp-server`; nothing reads the copy beside the source.

## A hazard worth knowing before you touch any of this

The Bash working directory **reverts on its own** from `/home/luks/kcc-lab-exp` to the
primary `/home/luks/kcc-lab`, and not only at session start. It happened three times in one
session. Relative paths then resolve against a different branch's content and a
`bin/controllerbuilder` built from different source.

It cost a killed 40-minute regeneration — a `grep` over the wrong tree returned zero
markers and read as a generator regression — and a confident, wrong claim that a binary had
been swapped when it was simply another tree's build.

Run `pwd` before trusting a measurement that says a change had no effect, prefer
`git -C /home/luks/kcc-lab-exp`, and compare binaries by `md5sum` rather than size or
mtime; two trees' builds can share a size.
