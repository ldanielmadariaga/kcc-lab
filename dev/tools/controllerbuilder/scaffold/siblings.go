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
	"regexp"
	"strings"
)

// kindDeclaration matches the Spec struct each Kind's types file declares, which
// is how a package states which Kinds it holds.
var kindDeclaration = regexp.MustCompile(`(?m)^type (\w+)Spec struct`)

// SiblingResources maps a lowercased field-name candidate to the Kind of a
// resource the target package declares.
//
// It reads two sources, because neither is complete on its own. The
// invocation's own resource list covers only its slice of a service generated
// by several generate-types calls: discoveryengine runs twice, dialogflow four
// times. The package scan covers the whole service, but finds nothing after a
// wipe-based regeneration, which deletes every _types.go before the generator
// runs. The map is the union of the two.
//
// The key strips the service prefix, so DiscoveryEngineDataStore is keyed
// "datastore" and a field called dataStore matches it.
func SiblingResources(dir, service string, alsoKnown ...string) map[string]string {
	out := map[string]string{}
	add := func(kind string) {
		trimmed := strings.TrimPrefix(strings.ToLower(kind), strings.ToLower(service))
		if trimmed != "" && trimmed != strings.ToLower(kind) {
			out[trimmed] = kind
		}
	}
	for _, kind := range alsoKnown {
		add(kind)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_types.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		for _, m := range kindDeclaration.FindAllStringSubmatch(string(body), -1) {
			add(m[1])
		}
	}
	return out
}
