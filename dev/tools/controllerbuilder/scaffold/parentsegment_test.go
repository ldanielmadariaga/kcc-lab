// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scaffold

import (
	"testing"

	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/protoapi"
)

func TestParentSegmentJudgement(t *testing.T) {
	grid := []struct {
		name        string
		pattern     string
		parentStyle protoapi.ParentStyle
		wantPaths   []string
		wantReasons []string
	}{
		{
			// The template emits projectRef and location itself, so there is
			// nothing left to say.
			name:        "projects and locations is fully covered",
			pattern:     "projects/{project}/locations/{location}/foos/{foo}",
			parentStyle: protoapi.ParentProjectLocation,
			wantPaths:   nil,
		},
		{
			// DiscoveryEngineDataStore. Upstream carries both spec.location and
			// spec.collection; we emitted neither and named only location.
			name:        "nested collection is named",
			pattern:     "projects/{project}/locations/{location}/collections/{collection}/dataStores/{data_store}",
			parentStyle: protoapi.ParentOther,
			wantPaths:   []string{".spec.location", ".spec.collection"},
			wantReasons: []string{"location-omitted-nested-parent", "parent-segment-omitted"},
		},
		{
			name:        "snake_case placeholder becomes a JSON name",
			pattern:     "projects/{project}/databases/{database}/collectionGroups/{collection_group}/fields/{field}",
			parentStyle: protoapi.ParentOther,
			wantPaths:   []string{".spec.database", ".spec.collectionGroup"},
			wantReasons: []string{"parent-segment-omitted", "parent-segment-omitted"},
		},
		{
			// projectRef covers {project} whatever the style.
			name:        "project alone needs nothing",
			pattern:     "projects/{project}/foos/{foo}",
			parentStyle: protoapi.ParentProject,
			wantPaths:   nil,
		},
		{
			name:        "no pattern falls back to the unknown-parent entry",
			pattern:     "",
			parentStyle: protoapi.ParentUnknown,
			wantPaths:   []string{".spec.location"},
			wantReasons: []string{"location-omitted-unknown-parent"},
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			got := parentSegmentJudgement(g.pattern, string(g.parentStyle))
			if len(got) != len(g.wantPaths) {
				t.Fatalf("got %d entries %v, want %d %v", len(got), paths(got), len(g.wantPaths), g.wantPaths)
			}
			for i, it := range got {
				if it.FieldPath != g.wantPaths[i] {
					t.Errorf("entry %d path = %q, want %q", i, it.FieldPath, g.wantPaths[i])
				}
				if it.Reason != g.wantReasons[i] {
					t.Errorf("entry %d reason = %q, want %q", i, it.Reason, g.wantReasons[i])
				}
			}
		})
	}
}

func paths(items []JudgementItem) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.FieldPath)
	}
	return out
}
