# Bulk greenfield generation: mechanics

**Status: PLACEHOLDER — do not follow yet.** **Scope:** the experimental sandbox (`kcc-lab`).

This document will carry the executable instructions for generating resources in bulk: the exact
command sequence, what the tooling does and does not do for you, what needs hand-fixing, and the
definition of done for a first-pass resource.

It is deliberately unwritten. See §3.

---

## 1. What it will cover

- **Ground rules for pass 1** — explicit permission to ship crudely, and the explicit non-goals:
  references as plain `string`, no e2e fixtures, no MockGCP, minimal field coverage, `v1alpha1` only.
- **The command sequence** — per-service generation, end to end.
- **Minimum viable resource** — the artifact list that constitutes done for pass 1.
- **Batch workflow** — picking a batch from the tracker through to a PR.
- **What breaks** — the failure modes, and how to recover from each.

## 2. What is known so far

Preliminary findings from reading the tooling. **Unverified by execution** — treat as orientation,
not instructions.

The batch unit is the per-service `generate.sh`. 129 services already have one under
`apis/<service>/`, and they follow a common shape:

```bash
cd ${REPO_ROOT}/dev/tools/controllerbuilder
./generate-proto.sh                      # note: lives here, not under apis/
${CONTROLLERBUILDER} generate-types \
  --service google.cloud.<service>.v1 \
  --api-version <service>.cnrm.cloud.google.com/v1alpha1 \
  --resource <Kind>:<ProtoMessage>       # repeatable
${CONTROLLERBUILDER} generate-mapper --service ... --api-version ...
cd ${REPO_ROOT} && dev/tasks/generate-crds
```

Three things worth knowing before writing this up properly:

- **`generate-types` takes a repeatable `--resource` flag**, so a whole service's types and mappers
  generate in one command. This is the batch lever.
- **`generate-controller` and `generate-direct-reconciler` take a single `--resource`**, so
  controllers are a per-resource loop inside the batch.
- **There are two configuration styles.** Most services pass `--resource` flags inline; a few
  (e.g. `apis/accesscontextmanager/`) use a `generatetypes.yaml` config file listing `kind`/
  `protoName` pairs. Which is better for bulk work is an open question, and picking one is part of
  writing this document.

## 3. Why this is a placeholder

Everything in §2 was derived from reading `--help` output and example scripts, not from running
anything. That approach has already produced two errors in a single sitting: `generate-proto.sh` was
initially recorded as living under `apis/<service>/` when it is actually in
`dev/tools/controllerbuilder/` and requires a `cd` first — meaning the documented first command would
have failed — and the second configuration style was missed entirely.

A mechanics document written from flag inspection is a hypothesis. If it is wrong, every agent
following it inherits the error, and it will be wrong precisely where it matters, because those are
the parts you only discover by running them.

**This document will be written from a pilot batch**: one small service taken end to end, recording
the real command path, the real failures, and the real recovery steps. It then describes a path known
to work rather than one believed to work.

Until then, `greenfield-coverage-strategy.md` holds the direction and `greenfield-tracker.md` holds
the work-selection process.
