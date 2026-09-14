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

import "testing"

// The map SiblingResources builds: lowercased Kind suffix -> Kind.
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

func TestSingularLeavesNonPluralsAlone(t *testing.T) {
	// A word ending in s that is not a plural must not be mangled into a
	// spurious match: "status" is on nearly every resource in the tree.
	//
	// Not exhaustive, and not meant to be. An acronym like "https" does become
	// "http", which the ss/us/is guards do not catch. It costs nothing here
	// because no service declares a Kind ending in Http, and a rule that only
	// decides whether to ask a person a question does not need to be right about
	// English, only about GCP resource names.
	for _, s := range []string{"status", "access", "analysis", "as"} {
		if got := singular(s); got != s {
			t.Errorf("singular(%q) = %q, want it left alone", s, got)
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
