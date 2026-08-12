// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lint

import (
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

// protoReferenceFields holds proto field names that carry
// (google.api.resource_reference) somewhere in googleapis.
//
// The annotation names the exact target type and is upstream ground truth, so
// where it exists it should beat reading a field's prose description.
//
// NOT CURRENTLY WIRED INTO TestMissingRefs. Matching by field name was measured
// and is unusable: it produced 2,164 findings against 78 for the description
// heuristics alone. A CRD does not record which proto field it came from, so a
// name like "network" - annotated in one service - matches hundreds of
// unrelated CRD fields elsewhere.
//
// To use the annotation safely, map each CRD field to its originating proto
// field via the +kcc:proto:field= markers the generator already emits into
// _types.go, then look up the fully-qualified proto path rather than the leaf
// name. The data and generator here are the groundwork for that; the mapping is
// its own change.
//
// Regenerate with: dev/tasks/generate-proto-reference-fields
var protoReferenceFields = loadProtoReferenceFields("testdata/proto_reference_fields.txt")

func loadProtoReferenceFields(path string) sets.String {
	out := sets.NewString()
	data, err := os.ReadFile(path)
	if err != nil {
		// Absent list degrades to description-only detection rather than failing
		// every CRD test; the file is regenerated from a googleapis checkout.
		klog.Warningf("proto reference field list %q not loaded (%v); falling back to description heuristics only", path, err)
		return out
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out.Insert(line)
	}
	return out
}

// Referenced so the loader and data file stay live while the proto-path mapping
// is built; see the note on protoReferenceFields.
var _ = protoReferenceFields
