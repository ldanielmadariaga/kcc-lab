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
	"strings"
	"testing"
)

// TestParentRef pins the rule that decides whether a resource gets a parent
// reference field or only a queue entry. A field emitted here lands in a CRD,
// so the cases that matter are the ones where the generator must decline: no
// reference type to point at, or several that match equally well. The queue
// entry is what TestMissingRefs reads, so an emitted field without one would be
// an unflagged guess.
func TestParentRef(t *testing.T) {
	grid := []struct {
		name       string
		pattern    string
		refSource  string
		wantField  string
		wantReason string
	}{
		{
			name:       "a reference type in the service package is used",
			pattern:    "projects/{project}/locations/{location}/clusters/{cluster}/nodePools/{node_pool}",
			refSource:  "package v1alpha1\n\ntype ClusterRef struct{}\n",
			wantField:  "ClusterRef *ClusterRef `json:\"clusterRef,omitempty\"`",
			wantReason: "parent-ref-guessed",
		},
		{
			// PrivateCACertificate: the type carries the Kind's prefix while
			// upstream still calls the field caPoolRef.
			name:       "a kind-prefixed reference type still matches",
			pattern:    "projects/{project}/locations/{location}/caPools/{ca_pool}/certificates/{certificate}",
			refSource:  "package v1alpha1\n\ntype PrivateCACAPoolRef struct{}\n",
			wantField:  "CaPoolRef *PrivateCACAPoolRef `json:\"caPoolRef,omitempty\"`",
			wantReason: "parent-ref-guessed",
		},
		{
			// apis/kms/v1beta1 declares both, and counting the unexported one
			// would make this look ambiguous and emit nothing.
			name:       "an unexported type beside the real one is ignored",
			pattern:    "projects/{project}/locations/{location}/cryptoKeys/{crypto_key}/cryptoKeyVersions/{version}",
			refSource:  "package v1alpha1\n\ntype KMSCryptoKeyRef struct{}\n\ntype kmsCryptoKeyRef struct{}\n",
			wantField:  "CryptoKeyRef *KMSCryptoKeyRef `json:\"cryptoKeyRef,omitempty\"`",
			wantReason: "parent-ref-guessed",
		},
		{
			// FirestoreIndex: no CollectionGroupRef exists anywhere.
			name:       "no reference type means no field",
			pattern:    "projects/{project}/databases/{database}/collectionGroups/{collection_group}/indexes/{index}",
			refSource:  "package v1alpha1\n",
			wantReason: "parent-ref-not-modelled",
		},
		{
			// DiscoveryEngineServingConfig: both really exist upstream.
			name:       "several matching types mean no field",
			pattern:    "projects/{project}/locations/{location}/engines/{engine}/servingConfigs/{config}",
			refSource:  "package v1alpha1\n\ntype DiscoveryEngineEngineRef struct{}\n\ntype DiscoveryEngineSearchEngineRef struct{}\n",
			wantReason: "parent-ref-not-modelled",
		},
		{
			name:      "a project and location parent is already carried",
			pattern:   "projects/{project}/locations/{location}/foos/{foo}",
			refSource: "package v1alpha1\n",
		},
		{
			name:      "a resource with no pattern has no parent to name",
			pattern:   "",
			refSource: "package v1alpha1\n",
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			// Arrange
			dir := t.TempDir()
			scaffolder := &APIScaffolder{BaseDir: dir, GoPackage: "svc/v1alpha1"}
			pkgDir := filepath.Join(dir, scaffolder.GoPackage)
			if err := os.MkdirAll(pkgDir, 0o755); err != nil {
				t.Fatalf("creating the service package: %v", err)
			}
			if err := os.WriteFile(filepath.Join(pkgDir, "refs.go"), []byte(g.refSource), 0o644); err != nil {
				t.Fatalf("writing the reference types: %v", err)
			}

			// Act
			field, item := scaffolder.parentRef(g.pattern)

			// Assert
			if g.wantField == "" && field != "" {
				t.Errorf("emitted a field where none was wanted:\n%s", field)
			}
			if g.wantField != "" && !strings.Contains(field, g.wantField) {
				t.Errorf("field = %q, want it to contain %q", field, g.wantField)
			}
			if g.wantField != "" && !strings.Contains(field, "+kcc:guess") {
				t.Errorf("an emitted field must carry the guess marker, got:\n%s", field)
			}
			switch {
			case g.wantReason == "" && item != nil:
				t.Errorf("queued %q where nothing was wanted", item.Reason)
			case g.wantReason != "" && item == nil:
				t.Errorf("queued nothing, want reason %q", g.wantReason)
			case g.wantReason != "" && item.Reason != g.wantReason:
				t.Errorf("reason = %q, want %q", item.Reason, g.wantReason)
			}
		})
	}
}
