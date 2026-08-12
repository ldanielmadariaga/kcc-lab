// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package greenfield scopes conformance checks to the resources produced by the
// experimental bulk greenfield workflow.
//
// SANDBOX-ONLY. This package does not exist upstream.
package greenfield

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// fileSuffixes are the per-resource file suffixes enforced by
// TestDirectResourceFileNaming. Keep in sync with tests/apichecks/naming_test.go.
var fileSuffixes = []string{
	"_types.go",
	"_types_test.go",
	"_identity.go",
	"_identity_test.go",
	"_reference.go",
	"_reference_test.go",
	"_mapper.go",
	"_mapper_test.go",
	"_fuzzer.go",
	"_fuzzer_test.go",
	"_controller.go",
	"_controller_test.go",
}

// Resource is one entry in the bulk manifest.
type Resource struct {
	Group string
	Kind  string
}

// GroupKind returns the schema.GroupKind for this resource.
func (r Resource) GroupKind() schema.GroupKind {
	return schema.GroupKind{Group: r.Group, Kind: r.Kind}
}

// Service returns the KCC service short name, derived from the group.
// For "networkservices.cnrm.cloud.google.com" this is "networkservices".
func (r Resource) Service() string {
	return strings.SplitN(r.Group, ".", 2)[0]
}

// Manifest is the set of resources produced by the bulk workflow.
type Manifest struct {
	resources []Resource
	byGK      map[schema.GroupKind]bool
}

// Load reads the bulk manifest. manifestPath is usually
// "testdata/greenfield_bulk.txt" relative to the test's working directory.
func Load(manifestPath string) (*Manifest, error) {
	f, err := os.Open(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("opening bulk manifest %q: %w", manifestPath, err)
	}
	defer f.Close()

	m := &Manifest{byGK: make(map[schema.GroupKind]bool)}

	scanner := bufio.NewScanner(f)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		group, kind, ok := strings.Cut(line, "/")
		if !ok || group == "" || kind == "" {
			return nil, fmt.Errorf("%s:%d: expected \"<group>/<Kind>\", got %q", manifestPath, lineNo, line)
		}
		r := Resource{Group: group, Kind: kind}
		if m.byGK[r.GroupKind()] {
			return nil, fmt.Errorf("%s:%d: duplicate entry %q", manifestPath, lineNo, line)
		}
		m.resources = append(m.resources, r)
		m.byGK[r.GroupKind()] = true
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading bulk manifest %q: %w", manifestPath, err)
	}
	return m, nil
}

// Resources returns the manifest entries, in file order.
func (m *Manifest) Resources() []Resource {
	return m.resources
}

// IsBulk reports whether the GroupKind was produced by the bulk workflow, and
// is therefore in scope for the greenfield conformance checks.
func (m *Manifest) IsBulk(gk schema.GroupKind) bool {
	return m.byGK[gk]
}

// FilesFor returns the existing per-resource Go files for r, under the given
// repo root.
//
// This relies on the naming convention enforced by TestDirectResourceFileNaming:
// every file under apis/ and pkg/controller/direct/ is prefixed with the
// lowercased Kind. Shared generated files (types.generated.go,
// mapper.generated.go) are deliberately excluded: they mix bulk-generated types
// with pre-existing ones, so file-level checks cannot attribute findings.
//
// Only paths that exist are returned, so a resource that has not reached a
// later phase yet simply yields fewer files.
func (m *Manifest) FilesFor(repoRoot string, r Resource) ([]string, error) {
	prefix := strings.ToLower(r.Kind)

	dirs, err := filepath.Glob(filepath.Join(repoRoot, "apis", r.Service(), "v*"))
	if err != nil {
		return nil, fmt.Errorf("globbing api dirs for %s: %w", r.Kind, err)
	}
	dirs = append(dirs, filepath.Join(repoRoot, "pkg", "controller", "direct", r.Service()))

	var found []string
	for _, dir := range dirs {
		for _, suffix := range fileSuffixes {
			path := filepath.Join(dir, prefix+suffix)
			if _, err := os.Stat(path); err == nil {
				found = append(found, path)
			}
		}
	}
	sort.Strings(found)
	return found, nil
}

// missingRe matches the generator's marker for a proto field with no KRM
// representation, e.g. "\t// MISSING: Labels".
var missingRe = regexp.MustCompile(`^\s*//\s*MISSING:\s*(\S+)\s*$`)

// funcRe matches a generated mapper function, capturing the KRM type name and
// the direction, e.g. "func FooSpec_v1alpha1_FromProto(" -> "FooSpec", "FromProto".
var funcRe = regexp.MustCompile(`^func ([A-Za-z0-9_]+)_v\d+(?:alpha|beta)?\d*_(FromProto|ToProto)\(`)

// DroppedFields returns the proto fields that have no representation in r's KRM
// types, keyed by field name.
//
// The generator emits "// MISSING: <Field>" while walking proto fields, whenever
// the KRM struct has neither <Field> nor <Field>Ref
// (dev/tools/controllerbuilder/pkg/codegen/mappergenerator.go).
//
// Spec and ObservedState map the *same* proto message, so each reports the
// other's fields: a field living in Spec shows up as MISSING in the
// ObservedState mapper and vice versa. A field is only genuinely dropped when it
// is MISSING in both. Resources with no ObservedState mapper use the Spec list
// alone.
func DroppedFields(mapperPath string, kind string) ([]string, error) {
	data, err := os.ReadFile(mapperPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %q: %w", mapperPath, err)
	}

	// krmType -> set of MISSING field names (union of FromProto and ToProto).
	byType := map[string]map[string]bool{}
	currentType := ""

	for _, line := range strings.Split(string(data), "\n") {
		if m := funcRe.FindStringSubmatch(line); m != nil {
			currentType = m[1]
			continue
		}
		if currentType == "" {
			continue
		}
		if m := missingRe.FindStringSubmatch(line); m != nil {
			if byType[currentType] == nil {
				byType[currentType] = map[string]bool{}
			}
			byType[currentType][m[1]] = true
		}
	}

	spec, hasSpec := byType[kind+"Spec"]
	observed, hasObserved := byType[kind+"ObservedState"]
	if !hasSpec {
		return nil, nil
	}

	var dropped []string
	for field := range spec {
		if hasObserved && !observed[field] {
			continue // present in ObservedState; not dropped
		}
		dropped = append(dropped, field)
	}
	sort.Strings(dropped)
	return dropped, nil
}

// MapperPath returns the generated mapper file for r's service.
func MapperPath(repoRoot string, r Resource) string {
	return filepath.Join(repoRoot, "pkg", "controller", "direct", r.Service(), "mapper.generated.go")
}

// FindCRD returns the CRD matching r from crds, and whether it was found.
//
// The CRD's metadata.name is what the apichecks exception baselines key on
// (crd=...), so it must come from the CRD itself rather than being guessed:
// the plural is not always Kind+"s".
func FindCRD(crds []apiextensions.CustomResourceDefinition, r Resource) (*apiextensions.CustomResourceDefinition, bool) {
	for i := range crds {
		crd := &crds[i]
		if crd.Spec.Group == r.Group && crd.Spec.Names.Kind == r.Kind {
			return crd, true
		}
	}
	return nil, false
}
