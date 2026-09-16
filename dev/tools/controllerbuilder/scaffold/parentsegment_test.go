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
	"os"
	"path/filepath"
	"slices"
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
			// A collection between the location and the resource. Upstream
			// carries spec.location and spec.collection; the template emits
			// neither, so both belong in the entry.
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
			// Act
			got := parentSegmentJudgement(g.pattern, string(g.parentStyle))

			// Assert
			if !slices.Equal(paths(got), g.wantPaths) {
				t.Fatalf("entries = %v, want %v", paths(got), g.wantPaths)
			}
			if !slices.Equal(reasons(got), g.wantReasons) {
				t.Errorf("reasons = %v, want %v", reasons(got), g.wantReasons)
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

func reasons(items []JudgementItem) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Reason)
	}
	return out
}

func TestParentSegments(t *testing.T) {
	grid := []struct {
		name     string
		pattern  string
		want     [][2]string
		location string
	}{
		{
			name:     "projects and locations",
			pattern:  "projects/{project}/locations/{location}/clusters/{cluster}/topics/{topic}",
			want:     [][2]string{{"projects", "project"}, {"locations", "location"}, {"clusters", "cluster"}},
			location: "location",
		},
		{
			// The placeholder is "location" but the collection is "regions", and
			// upstream names the field after the collection.
			name:     "regions, not locations",
			pattern:  "projects/{project}/regions/{location}/jobs/{job}",
			want:     [][2]string{{"projects", "project"}, {"regions", "location"}},
			location: "region",
		},
		{
			name:    "organization root",
			pattern: "organizations/{organization}/muteConfigs/{mute_config}",
			want:    [][2]string{{"organizations", "organization"}},
		},
		{
			name:    "no parent",
			pattern: "foos/{foo}",
			want:    nil,
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			// Act
			got := parentSegments(g.pattern)

			// Assert
			if !slices.Equal(got, g.want) {
				t.Fatalf("parentSegments(%q) = %v, want %v", g.pattern, got, g.want)
			}
			var loc string
			for i, s := range got {
				if i == 0 {
					continue
				}
				if f, ok := locationFieldNames[s[0]]; ok {
					loc = f
				}
			}
			if loc != g.location {
				t.Errorf("location field = %q, want %q", loc, g.location)
			}
		})
	}
}

func TestSiblingResourceKeying(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	write := func(name, body string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}
	write("discoveryenginedatastore_types.go", "package v1alpha1\n\ntype DiscoveryEngineDataStoreSpec struct{}\n")
	write("discoveryenginecontrol_types.go", "package v1alpha1\n\ntype DiscoveryEngineControlSpec struct{}\n")
	// Not a _types.go file, so not a resource.
	write("helpers.go", "package v1alpha1\n\ntype NotAResourceSpec struct{}\n")

	// Act
	got := SiblingResources(dir, "discoveryengine")

	// Assert
	if got["datastore"] != "DiscoveryEngineDataStore" {
		t.Errorf(`siblings["datastore"] = %q, want DiscoveryEngineDataStore`, got["datastore"])
	}
	if got["control"] != "DiscoveryEngineControl" {
		t.Errorf(`siblings["control"] = %q, want DiscoveryEngineControl`, got["control"])
	}
	if _, ok := got["notaresource"]; ok {
		t.Error("a type outside a _types.go file must not count as a resource")
	}
	// A Kind that is only the service name leaves nothing to key on.
	if len(got) != 2 {
		t.Errorf("got %d siblings, want 2: %v", len(got), got)
	}
}
