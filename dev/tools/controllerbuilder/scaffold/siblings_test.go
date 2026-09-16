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
	"maps"
	"os"
	"path/filepath"
	"testing"
)

// TestSiblingResources pins the map the sibling rule matches field names
// against. codegen.SiblingResource looks up a field's name in it, so a key that
// is wrong either marks the wrong field as a probable reference or misses one,
// and the generated type carries that guess into review.
func TestSiblingResources(t *testing.T) {
	grid := []struct {
		name       string
		readSubdir string
		files      map[string]string
		alsoKnown  []string
		want       map[string]string
	}{
		{
			name:       "kinds the package declares, keyed without the service prefix",
			readSubdir: ".",
			files: map[string]string{
				"discoveryenginedatastore_types.go": "package v1alpha1\n\ntype DiscoveryEngineDataStoreSpec struct{}\n",
				"discoveryengineengine_types.go":    "package v1alpha1\n\ntype DiscoveryEngineEngineSpec struct{}\n",
			},
			want: map[string]string{
				"datastore": "DiscoveryEngineDataStore",
				"engine":    "DiscoveryEngineEngine",
			},
		},
		{
			// This run's Kinds and the package's differ, so both sources are read.
			name:       "this run's kinds are added to the package's",
			readSubdir: ".",
			files: map[string]string{
				"discoveryenginedatastore_types.go": "package v1alpha1\n\ntype DiscoveryEngineDataStoreSpec struct{}\n",
			},
			alsoKnown: []string{"DiscoveryEngineControl"},
			want: map[string]string{
				"datastore": "DiscoveryEngineDataStore",
				"control":   "DiscoveryEngineControl",
			},
		},
		{
			// A Kind from another service would match field names across service
			// boundaries.
			name:       "a kind outside the service is left out",
			readSubdir: ".",
			files: map[string]string{
				"storagebucket_types.go": "package v1alpha1\n\ntype StorageBucketSpec struct{}\n",
			},
			want: map[string]string{},
		},
		{
			name:       "only types files are read",
			readSubdir: ".",
			files: map[string]string{
				"discoveryenginecontrol_reference.go": "package v1alpha1\n\ntype DiscoveryEngineControlSpec struct{}\n",
			},
			want: map[string]string{},
		},
		{
			// A wipe-based regeneration leaves no _types.go to scan, so this
			// run's list is all there is.
			name:       "no package to scan leaves this run's kinds",
			readSubdir: "absent",
			alsoKnown:  []string{"DiscoveryEngineControl"},
			want:       map[string]string{"control": "DiscoveryEngineControl"},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			// Arrange
			dir := t.TempDir()
			for name, body := range g.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
					t.Fatalf("writing %s: %v", name, err)
				}
			}

			// Act
			got := SiblingResources(filepath.Join(dir, g.readSubdir), "discoveryengine", g.alsoKnown...)

			// Assert
			if !maps.Equal(got, g.want) {
				t.Errorf("SiblingResources() = %v, want %v", got, g.want)
			}
		})
	}
}
