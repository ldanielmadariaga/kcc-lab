# Replication experiment: 100 resources

The pilot measured three resources and got 100% on everything after a judgement pass. This run took
the 100 most recently added upstream CRDs to see whether that held at scale. It did not, and the ways
it failed are the useful part.

## What was measured

95 of the latest 100 upstream CRDs are generatable: all `v1alpha1`, across 56 services, each already
carrying a `--resource Kind:ProtoMessage` line. Three lack that line and two have no CRD on disk.

All 56 services regenerated with **zero generation errors**, and 93 of 95 produced a `_types.go`.
Compilation is where it broke, and packages that failed to compile had to be restored to their
originals before `controller-gen` would run at all, which is why the scored set is 40 rather than 95.

Baseline `ae2abcdb9b`, the commit still holding every hand-written original. Scored with
`crd-mcp-server score`.

## Result, 40 resources

| bucket | matched | missing | extra | mismatch | rate |
|---|---|---|---|---|---|
| spec | 1201 | 404 | 176 | 9 | 74.4% |
| required | 162 | 29 | 111 | 0 | 84.8% |
| status.observedState | 440 | 44 | 206 | 10 | 89.1% |

Per resource: 19 of 40 have a spec with no missing and no mismatched path at all, 26 of 40 for
required, 29 of 40 for observedState.

**ObservedState at 89.1% is the headline**, because before this session it was 0% by construction —
the scaffolder emitted an empty struct and a human filled it. It is now produced mechanically in the
first pass.

### The spec number needs unpacking

The 404 missing spec paths are not 404 separate problems:

| | paths | share | resources |
|---|---|---|---|
| reference fields | 313 | 77.5% | 20 |
| map-of-message the generator cannot type | 83 | 20.5% | **1** |
| everything else | 8 | 2.0% | 2 |

References are Step 1 doing what it is designed to do: it emits a plain string and defers the
decision to the judgement pass. Each ref costs four paths in this metric (the object plus `external`,
`name` and `namespace`), so 313 paths is roughly 78 reference fields.

The map case is one resource. `HypercomputeClusterCluster.spec.computeResources` is a
`map<string, Message>`, and `GoTypeForField` supports only `map[string]string` and
`map[string]int64`, so the field is dropped and takes 83 nested paths with it.

Restating the same measurement with those two accounted for:

| | rate |
|---|---|
| spec as measured | 74.4% |
| spec excluding references, which Step 1 defers by design | 92.3% |
| spec also excluding the one unsupported map type | **98.6%** |

## Why 55 resources could not be scored

Every failure is in the scaffold and pruning layers. None is in the ObservedState or acronym work,
both of which behaved correctly everywhere they ran.

- **`Location` emitted as a non-pointer `string`** (14 errors). Hand-written `_identity.go` files
  call `common.ValueOf(obj.Spec.Location)` and need `*string`. In `sqladminbackup_types.go` it is
  worse: the proto already has a `location` field, so the scaffold's addition is a redeclaration.
- **Types pruned while still referenced** (14). `contentwarehouse` references `Document_Style`,
  `Document_Page` and others that `--prune-unused-types` removed from the same file that refers to
  them.
- **Generated type name collides with the Kind** (5). `APIHubInstance` and `Document` end up declared
  by both the scaffold and `types.generated.go`.
- **Parent shape not reaching the Spec template.** `cloudsecuritycompliance` has a hand-written
  identity expecting `Spec.OrganizationRef` while the scaffold emitted `ProjectRef`. Phase 2 fixed
  the identity template; the Spec template still assumes project or project/location.
- **`apiextensionsv1.JSON` emitted without its import** (`ces`). The same class of bug as the
  `common.Status` import fixed this session, so that fix should be generalised rather than
  special-cased.
- **Two `generate-types` invocations writing one `types.generated.go`** (`networksecurity`), the
  second overwriting the first.

## A rule the runbook is missing

`apigee` and `artifactregistry` pass `--skip-scaffold-files`, so the generator never writes
`<kind>_types.go` for them. Deleting that file removes something nothing will regenerate. Any
bulk-regeneration procedure has to skip resources whose invocation sets that flag.
