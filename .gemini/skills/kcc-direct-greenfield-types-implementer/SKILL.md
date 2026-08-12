---
name: kcc-direct-greenfield-types-implementer
description: Automate the initial scaffolding of a KCC "direct" resource, including CRD types and generation scripts. Use this when starting a new "direct" implementation for a GCP resource.
---

# KCC Direct Greenfield Types Implementer

This skill guides the initial scaffolding of *new* (greenfield) KCC "direct" resources, ensuring standardized CRD generation and adherence to project-wide validation patterns.

## Scope: Step 1 is types + CRD only  [SANDBOX DECISION]

This skill covers **types and the generated CRD, and nothing else**. `<kind>_identity.go` and
`<kind>_reference.go` are **Step 2** and must not be produced here.

Upstream groups types, identity and reference into one step. The sandbox splits them deliberately:
identity and reference generation is mechanical and hard to get wrong, so with decent tests it
barely needs review. Deciding the **types and CRD** is where the judgement lives - which fields
exist, which are references, what is dropped and why - so it gets a phase to itself, and that is
where review and automated checking concentrate.

## Provenance: what is enforced, and by what authority  [SANDBOX]

Some checks referenced here exist only in this sandbox and are **not** team-vetted. Do not infer
project policy from them:

| Sandbox-only (ours) | Upstream (team-vetted) |
|---|---|
| `greenfield_bulk.txt`, `greenfield_dropped_fields.txt`, `refs_deferred.txt` | `docs/develop-resources/api-conventions/resource-reference.md` |
| `refs_not_representable.txt`, `isPatternField`, `notRepresentableReason` | `TestMissingRefs` exclusions (`.zone`, `.location`, `.machineType`, `.acceleratorType`) |
| `deprecated_refs_v1beta1.txt`, the enum prohibition below | Merged implementations in `apis/` |

For when a field should be a reference, see `docs/ai/refs-decision-guide.md`.

## What the generator does and does not do

- **Automatic:** `types.generated.go` (every proto message as a Go struct, correct pointer and
  collection types, `+kcc:proto:field=` annotations), `mapper.generated.go`,
  `zz_generated.deepcopy.go`, and the CRD.
- **Manual:** composing the top-level Spec and ObservedState in `<kind>_types.go`. The scaffold ships
  with `ProjectRef`, `Location` and `ResourceID` only; you move fields in from the
  `/* unreachable type ... */` block. Referencing a type makes `prunetypes` un-comment it on the next
  run. Never hand-edit `types.generated.go`.
- Mappers follow automatically once the types are right. Hand-written mapping is needed only where
  the mapper emits `// MISSING:`.

**Known generator gaps** (measured, not theoretical):

1. `google.protobuf.Value` / `ListValue` have **no mappers anywhere** - any resource reaching them
   fails to compile. Dropping the field is the only Step 1 workaround.
2. `[]common.Status` (repeated `google.rpc.Status`) emits a mapper call without adding the `common`
   import.
3. The scaffold emits `Location` as a non-pointer `string`, violating the pointer rule.

## Prerequisites
You **must** also apply the standards from the base skill: `.gemini/skills/kcc-direct-base-types-implementer/SKILL.md`.

## Inputs
- `service`: The Google API service name (e.g., `google.cloud.aiplatform.v1`).
- `resource`: The mapping of KCC Kind to GCP Resource (e.g., `VertexAIExampleStore:ExampleStore`).
- `api_version`: The KCC API version (default: `v1alpha1`).

## Workflow

### 1. Add to generate.sh
Locate `apis/<service_short>/generate.sh`. If it doesn't exist, create it following the standard KCC template:
```bash
#!/bin/bash
# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit
set -o nounset
set -o pipefail

REPO_ROOT="$(git rev-parse --show-toplevel)"

CONTROLLERBUILDER="${CONTROLLERBUILDER:-}"
if [[ -z "${CONTROLLERBUILDER}" ]]; then
  if [[ -x "${REPO_ROOT}/bin/controllerbuilder" ]]; then
    CONTROLLERBUILDER="${REPO_ROOT}/bin/controllerbuilder"
  else
    CONTROLLERBUILDER="go run ${REPO_ROOT}/dev/tools/controllerbuilder"
  fi
fi
source "${REPO_ROOT}/dev/tools/goimports.sh"
cd ${REPO_ROOT}/dev/tools/controllerbuilder
# Note: generate-proto.sh reuses cached .build/googleapis-<SHA>.pb files by default.
# Pass --force (or FORCE_GENERATE_PROTOS=1) to force re-compiling proto descriptors when testing proto edits:
./generate-proto.sh

${CONTROLLERBUILDER} generate-types \
  --service <service> \
  --api-version <group>.cnrm.cloud.google.com/<api_version> \
  --resource <resource>
```

### 2. Generate Types
Set executable permissions and run the `generate.sh` script:
```bash
chmod +x apis/<service_short>/generate.sh
./apis/<service_short>/generate.sh
```

### 3. Validate and Enhance Output
Apply the baseline validations from `kcc-direct-base-types-implementer`, plus these greenfield-specific rules:

- **Stability Level**: Add `// +kubebuilder:metadata:labels="cnrm.cloud.google.com/stability-level=alpha"`.
- **Field Validation**: Manually add or verify kubebuilder tags:
  - Use `// +kubebuilder:validation:Required` for fields that are mandatory in the GCP API.
  - Use `// +kubebuilder:validation:Optional` for all other fields.
- **Enums**:
  - Use `*string` for the Go type of proto enum fields (do NOT use custom wrapped string types).
  - **Do NOT add `// +kubebuilder:validation:Enum=...`.** [SANDBOX DECISION - diverges from upstream]
    Hardcoding the permitted values duplicates validation the GCP API already performs, and couples
    KCC releases to GCP enum additions: every new value upstream would need a KCC release before a
    user could set it. A plain `*string` accepts new values with no code change.
    Enforced by `TestGreenfieldBulkTypesConformance`. Existing resources are grandfathered.

### 4. Reference Fields (do not skip)
Fields that point at another GCP resource **must** be implemented as KCC reference fields
(`Ref` suffix, e.g. `refsv1beta1.ProjectRef`, `pubsubv1beta1.PubSubTopicRef`), per
`.gemini/skills/kcc-direct-base-types-implementer/SKILL.md`.

**You MUST NOT add entries to `tests/apichecks/testdata/exceptions/missingrefs.txt`.**
That file is a shrink-only ratchet: new entries fail CI and cannot be regenerated away,
including under `WRITE_GOLDEN_OUTPUT=1`. If the check flags a field, implement the
reference - do not suppress it.

**Primary signal - check the proto, not the description.** GCP protos annotate reference
fields with `(google.api.resource_reference)`, which names the exact target type:

```proto
optional string service_account = 16 [
  (google.api.field_behavior) = REQUIRED,
  (google.api.resource_reference) = { type: "iam.googleapis.com/ServiceAccount" }
];
```

Inspect the source `.proto` for this annotation on every string field before deciding it is
a plain string. It is authoritative where present.

**Secondary signals** (the annotation is not always present):
- Description contains a path template: `projects/`, `locations/{`, `folders/{`, `organizations/{`
- Value is a URI: `gs://` (Cloud Storage), `bq://` (BigQuery), or a field named `*Uri` / `*Url`
- KMS key names, service accounts, networks/subnetworks

Not every URI is a reference - some point at API-internal representations rather than a
addressable KCC resource. Judge per field; when it is a real GCP resource, make it a `Ref`.

### 5. Journaling
Append any quirks about the proto-to-struct mapping (e.g., field name collisions) to `.claude/journals/<service>.md` using the format described in the `kcc-agentic-journaler` skill.
