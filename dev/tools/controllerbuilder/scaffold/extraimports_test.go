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
	"strings"
	"testing"
)

// Both bodies are scanned because a special-cased type can land in either:
// securitycentermanagement puts an apiextensionsv1.JSON in the Spec, transcoder a
// common.Status in the ObservedState. Each was a separate compile failure.
func TestExtraImportsFor(t *testing.T) {
	tests := []struct {
		name       string
		spec       string
		observed   string
		wantSubstr string
		wantNone   bool
	}{
		{name: "apiextensionsv1 in spec", spec: "\tConfig apiextensionsv1.JSON `json:\"config\"`",
			wantSubstr: `apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"`},
		{name: "common in observed state", observed: "\tError *common.Status `json:\"error\"`",
			wantSubstr: `common "github.com/GoogleCloudPlatform/k8s-config-connector/apis/common"`},
		{name: "nothing special", spec: "\tName *string `json:\"name\"`", wantNone: true},
	}
	for _, tt := range tests {
		got := ExtraImportsFor(tt.spec, tt.observed)
		if tt.wantNone {
			if len(got) != 0 {
				t.Errorf("%s: want no imports, got %v", tt.name, got)
			}
			continue
		}
		if len(got) != 1 || !strings.Contains(got[0], tt.wantSubstr) {
			t.Errorf("%s: got %v, want one line containing %q", tt.name, got, tt.wantSubstr)
		}
		// The alias must be emitted: without it the qualifier does not resolve and
		// goimports strips the import as unused.
		if !strings.Contains(got[0], " \"") {
			t.Errorf("%s: import line %q has no alias", tt.name, got[0])
		}
	}
}
