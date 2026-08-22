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

package protoapi

import (
	"testing"
)

func TestSplitPattern(t *testing.T) {
	grid := []struct {
		name           string
		pattern        string
		wantCollection string
		wantParent     string
	}{
		{
			name:           "project and location",
			pattern:        "projects/{project}/locations/{location}/foos/{foo}",
			wantCollection: "foos",
			wantParent:     "projects/locations",
		},
		{
			// The casing here is the whole point: the template's ToLower would
			// produce "lbtrafficextensions", which is not a valid resource name.
			name:           "camelCase collection is preserved",
			pattern:        "projects/{project}/locations/{location}/lbTrafficExtensions/{lb_traffic_extension}",
			wantCollection: "lbTrafficExtensions",
			wantParent:     "projects/locations",
		},
		{
			name:           "irregular english plural",
			pattern:        "projects/{project}/policies/{policy}",
			wantCollection: "policies",
			wantParent:     "projects",
		},
		{
			name:           "organization parent",
			pattern:        "organizations/{organization}/bars/{bar}",
			wantCollection: "bars",
			wantParent:     "organizations",
		},
		{
			name:           "deeply nested parent",
			pattern:        "projects/{project}/locations/{location}/clusters/{cluster}/nodePools/{node_pool}",
			wantCollection: "nodePools",
			wantParent:     "projects/locations/clusters",
		},
		{
			// Some patterns name a collection with no trailing id.
			name:           "pattern ending in a literal has no collection",
			pattern:        "projects/{project}/locations",
			wantCollection: "",
			wantParent:     "projects/locations",
		},
		{
			name:           "singleton at root",
			pattern:        "foos/{foo}",
			wantCollection: "foos",
			wantParent:     "",
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			collection, parent := splitPattern(g.pattern)
			if collection != g.wantCollection {
				t.Errorf("collection = %q, want %q", collection, g.wantCollection)
			}
			if parent != g.wantParent {
				t.Errorf("parent = %q, want %q", parent, g.wantParent)
			}
		})
	}
}

func TestClassifyParent(t *testing.T) {
	grid := []struct {
		parentPath string
		want       ParentStyle
	}{
		{"projects/locations", ParentProjectLocation},
		{"projects", ParentProject},
		{"organizations", ParentOrganization},
		{"folders", ParentFolder},
		{"projects/locations/clusters", ParentOther},
		{"properties", ParentOther},
		{"", ParentUnknown},
	}

	for _, g := range grid {
		t.Run(g.parentPath, func(t *testing.T) {
			if got := classifyParent(g.parentPath); got != g.want {
				t.Errorf("classifyParent(%q) = %q, want %q", g.parentPath, got, g.want)
			}
		})
	}
}

func TestParentVariables(t *testing.T) {
	grid := []struct {
		name    string
		pattern string
		want    []string
	}{
		{
			name:    "project and location",
			pattern: "projects/{project}/locations/{location}/foos/{foo}",
			want:    []string{"project", "location"},
		},
		{
			// The case this was written for. ParentStyle collapses this to
			// "other", which tells the template what to render but not that
			// spec.collection is missing.
			name:    "collection between location and the resource",
			pattern: "projects/{project}/locations/{location}/collections/{collection}/dataStores/{data_store}",
			want:    []string{"project", "location", "collection"},
		},
		{
			name:    "organization parent",
			pattern: "organizations/{organization}/bars/{bar}",
			want:    []string{"organization"},
		},
		{
			name:    "no parent at all",
			pattern: "foos/{foo}",
			want:    nil,
		},
		{
			// A collection pattern names no resource, so nothing is the id and
			// every placeholder belongs to the parent.
			name:    "pattern ending in a literal",
			pattern: "projects/{project}/locations",
			want:    []string{"project"},
		},
		{
			name:    "no placeholders",
			pattern: "projects/locations",
			want:    nil,
		},
	}
	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			got := ParentVariables(g.pattern)
			if len(got) != len(g.want) {
				t.Fatalf("ParentVariables(%q) = %v, want %v", g.pattern, got, g.want)
			}
			for i := range got {
				if got[i] != g.want[i] {
					t.Errorf("ParentVariables(%q) = %v, want %v", g.pattern, got, g.want)
					break
				}
			}
		})
	}
}
