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
./generate-proto.sh

# An illustration of every opt-in generator flag on one new resource. There is
# no generate-mapper call: go.mod has no Go client for Chronicle, so a mapper
# would not compile.
${CONTROLLERBUILDER} generate-types \
  --service google.cloud.chronicle.v1 \
  --api-version chronicle.cnrm.cloud.google.com/v1alpha1 \
  --include-skipped-output \
  --prepopulate-spec \
  --emit-required-from-proto \
  --emit-plural-acronyms \
  --emit-message-maps \
  --place-server-set-fields \
  --detect-output-only-in-comments \
  --emit-parent-refs \
  --emit-sibling-refs \
  --detect-empty-observedstate \
  --emit-reference-hints \
  --resource ChronicleWatchlist:Watchlist

cd ${REPO_ROOT}
dev/tasks/generate-crds
