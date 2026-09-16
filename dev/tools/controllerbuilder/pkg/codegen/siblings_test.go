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

package codegen

import (
	"maps"
	"os"
	"path/filepath"
	"testing"
)

// The map SiblingKinds builds: lowercased Kind suffix -> Kind.
var testSiblings = map[string]string{
	"datastore":  "DiscoveryEngineDataStore",
	"subnetwork": "ComputeSubnetwork",
	"policy":     "ComputePolicy",
	"address":    "ComputeAddress",
	"database":   "FirestoreDatabase",
}

func TestSiblingResourceByName(t *testing.T) {
	for _, tc := range []struct {
		name string
		want string
	}{
		// Exact, which is the bulk of them.
		{"dataStore", "DiscoveryEngineDataStore"},
		{"database", "FirestoreDatabase"},
		// Plural. Six of the fifteen matches on this corpus are these, so
		// dropping the rule halves the reach.
		{"subnetworks", "ComputeSubnetwork"},
		{"policies", "ComputePolicy"},
		{"addresses", "ComputeAddress"},
		// Not a match at all.
		{"displayName", ""},
		{"etag", ""},
		// Suffix matches are deliberately not taken: endswith measured 68%
		// against 75% for exact, and pulls in fields like pipelineJob.
		{"defaultDataStore", ""},
	} {
		got, ok := SiblingResourceByName(tc.name, testSiblings)
		if tc.want == "" {
			if ok {
				t.Errorf("%s: matched %q, want no match", tc.name, got)
			}
			continue
		}
		if !ok || got != tc.want {
			t.Errorf("%s: got %q %v, want %q true", tc.name, got, ok, tc.want)
		}
	}
}

func TestSingular(t *testing.T) {
	// One case per branch of the switch, so deleting a branch names itself
	// rather than surfacing as a missed match somewhere downstream.
	for _, tc := range []struct {
		branch string
		in     string
		want   string
	}{
		{"ies, longer than four", "policies", "policy"},
		{"ies, too short for that branch, falls to trailing s", "ies", "ie"},
		{"sses", "addresses", "address"},
		{"ches", "batches", "batch"},
		{"shes", "dishes", "dish"},
		{"xes", "indexes", "index"},
		{"ss is not a plural", "access", "access"},
		{"us is not a plural", "status", "status"},
		{"is is not a plural", "analysis", "analysis"},
		{"trailing s", "subnetworks", "subnetwork"},
		{"trailing s, too short", "as", "as"},
		{"no plural to strip", "policy", "policy"},

		// Two known-wrong answers, pinned so a change to them is deliberate.
		// Neither costs anything today, because no Kind in the tree ends in
		// Http, Cache or Size, and a rule that only decides whether to ask a
		// person a question need not be right about English.
		//
		// The ss/us/is guards do not catch an acronym, so "https" loses its s.
		// The ches/shes/xes/zes branch assumes the stem ends in a consonant,
		// which holds for batches and boxes but not for a stem already ending
		// in e: caches and sizes lose that e as well.
		{"acronym, mangled", "https", "http"},
		{"stem ends in e, mangled", "caches", "cach"},
		{"stem ends in e, mangled", "sizes", "siz"},
	} {
		if got := Singular(tc.in); got != tc.want {
			t.Errorf("%s: Singular(%q) = %q, want %q", tc.branch, tc.in, got, tc.want)
		}
	}
}

func TestScanSiblingGuesses(t *testing.T) {
	body := `
type Control struct {
	// The data store this control belongs to.
	// +kcc:guess=possible-reference target=DiscoveryEngineDataStore
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Control.data_store
	DataStore *string ` + "`json:\"dataStore,omitempty\"`" + `

	// +kcc:proto:field=google.cloud.discoveryengine.v1.Control.display_name
	DisplayName *string ` + "`json:\"displayName,omitempty\"`" + `
}
`
	got := scanSiblingGuesses("google.cloud.discoveryengine.v1.Control", body)
	if len(got) != 1 {
		t.Fatalf("got %d guesses, want 1: %+v", len(got), got)
	}
	if got[0].Field != "dataStore" || got[0].Target != "DiscoveryEngineDataStore" {
		t.Errorf("got %+v", got[0])
	}
}

// TestSiblingKinds pins the map the sibling rule matches field names against.
// SiblingResource looks up a field's name in it, so a key that is wrong either
// marks the wrong field as a probable reference or misses one, and the
// generated type carries that guess into review.
func TestSiblingKinds(t *testing.T) {
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
			got := SiblingKinds(filepath.Join(dir, g.readSubdir), "discoveryengine", g.alsoKnown...)

			// Assert
			if !maps.Equal(got, g.want) {
				t.Errorf("SiblingKinds() = %v, want %v", got, g.want)
			}
		})
	}
}
