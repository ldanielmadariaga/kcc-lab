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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/codegen"
	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/protoapi"
)

// sharedRefsPackage holds the reference types services use in common, relative
// to the repo root.
const sharedRefsPackage = "apis/refs/v1beta1"

// parentRef renders the Spec field naming a resource's direct parent, together
// with its queue entry. Both are empty where the pattern names no parent below
// project and location, which is most resources.
//
// The field is written only where a reference type for the parent already
// exists. Where none does, or where several match, the entry says so and no
// field is written: a wrong reference is harder for a reviewer to catch than an
// absent one, because it compiles and reads as deliberate.
//
// Measured against the 46 hand-written kinds with such a parent, upstream
// models it in 45, names it as the collection segment predicts in 40, and has
// a reference type for it inside the service in 39.
func (a *APIScaffolder) parentRef(pattern string) (field string, item *JudgementItem) {
	collection, placeholder := protoapi.ParentPair(pattern)
	switch collection {
	case "", "projects", "locations", "regions", "zones", "global",
		"organizations", "folders":
		// projectRef and location already carry these shapes.
		return "", nil
	}

	// The collection segment names the field, not the placeholder beside it. See
	// ParentPair. Reading the collection is right 32 times out of 46, against 28
	// for the placeholder.
	name := codegen.Singular(collection)
	path := parentPath(pattern, collection, placeholder)

	candidates := parentRefTypes(a.repoRoot(), filepath.Join(a.BaseDir, a.GoPackage), name)
	if len(candidates) != 1 {
		detail := fmt.Sprintf("the parent is %s, and no %sRef type exists to point at; "+
			"add the parent resource first, then model this as a reference", path, exportedName(name))
		if len(candidates) > 1 {
			detail = fmt.Sprintf("the parent is %s, and several reference types match it (%s); "+
				"pick one and add the field by hand", path, strings.Join(sortedNames(candidates), ", "))
		}
		return "", &JudgementItem{
			Reason: "parent-ref-not-modelled",
			Detail: detail,
		}
	}

	typeName := sortedNames(candidates)[0]
	goType := typeName
	if qualifier := candidates[typeName]; qualifier != "" {
		goType = qualifier + "." + typeName
	}

	field = fmt.Sprintf("\t// A reference to the %s this resource belongs to.\n"+
		"\t// +kcc:guess\n"+
		"\t%sRef *%s `json:%q`", path, exportedName(name), goType, name+"Ref,omitempty")
	return field, &JudgementItem{
		FieldPath: ".spec." + name + "Ref",
		Reason:    "parent-ref-guessed",
		Detail: fmt.Sprintf("emitted as a reference to %s, read from the %s segment of %s; "+
			"the field name comes from the pattern rather than the proto, so confirm both the name and the target",
			typeName, collection, pattern),
	}
}

// parentRefTypes returns the reference types that could name a parent whose
// collection singularises to name, as a map of type name to import qualifier.
//
// Services spell a reference type either as a bare noun, ClusterRef in alloydb,
// or with the Kind's own prefix, as in DNSManagedZoneRef, so the match is on
// the suffix. The service's own package wins over the shared one, and
// unexported types are skipped: apis/kms/v1beta1 declares kmsCryptoKeyRef
// beside the real KMSCryptoKeyRef, which would make a single match look
// ambiguous.
func parentRefTypes(repoRoot, serviceDir, name string) map[string]string {
	want := strings.ToLower(name) + "ref"
	re := regexp.MustCompile(`(?m)^type ([A-Z]\w*Ref) struct`)

	scan := func(dir, qualifier string) map[string]string {
		found := map[string]string{}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return found
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
				continue
			}
			body, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			for _, m := range re.FindAllStringSubmatch(string(body), -1) {
				if strings.HasSuffix(strings.ToLower(m[1]), want) {
					found[m[1]] = qualifier
				}
			}
		}
		return found
	}

	if found := scan(serviceDir, ""); len(found) > 0 {
		return found
	}
	return scan(filepath.Join(repoRoot, sharedRefsPackage), "refsv1beta1")
}

// repoRoot is the tree BaseDir sits in, so the shared refs package can be
// found. BaseDir is the output API directory, normally <repo>/apis.
func (a *APIScaffolder) repoRoot() string {
	if filepath.Base(a.BaseDir) == "apis" {
		return filepath.Dir(a.BaseDir)
	}
	return a.BaseDir
}

// parentPath is the pattern truncated at the parent, which is the whole path a
// reference to it carries in its External field, per AIP-122.
func parentPath(pattern, collection, placeholder string) string {
	segs := strings.Split(pattern, "/")
	for i := 0; i+1 < len(segs); i++ {
		if segs[i] == collection && segs[i+1] == "{"+placeholder+"}" {
			return strings.Join(segs[:i+2], "/")
		}
	}
	return pattern
}

func sortedNames(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func exportedName(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
