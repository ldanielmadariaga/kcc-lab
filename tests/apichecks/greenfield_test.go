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

// Conformance checks for resources produced by the experimental bulk greenfield
// workflow. Scope comes from testdata/greenfield_bulk.txt; everything not listed
// there is out of scope, because it predates this bar.
//
// SANDBOX-ONLY. These checks do not exist upstream.

package lint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/crd/crdloader"
	"github.com/GoogleCloudPlatform/k8s-config-connector/tests/apichecks/greenfield"
)

const (
	bulkManifestPath = "testdata/greenfield_bulk.txt"
	repoRoot         = "../.."
)

func loadBulkManifest(t *testing.T) *greenfield.Manifest {
	t.Helper()
	m, err := greenfield.Load(bulkManifestPath)
	if err != nil {
		t.Fatalf("loading bulk manifest: %v", err)
	}
	return m
}

// TestGreenfieldBulkManifestIsResolvable checks that every entry in the manifest
// corresponds to a real CRD and at least one on-disk file. A typo here would
// silently narrow the scope of every other check in this file, so it is checked
// first and explicitly.
func TestGreenfieldBulkManifestIsResolvable(t *testing.T) {
	t.Parallel()

	m := loadBulkManifest(t)
	crds, err := crdloader.LoadAllCRDs()
	if err != nil {
		t.Fatalf("loading CRDs: %v", err)
	}

	for _, r := range m.Resources() {
		if _, ok := greenfield.FindCRD(crds, r); !ok {
			t.Errorf("FAIL: %s/%s is in %s but no matching CRD was found. Check the group and Kind spelling.",
				r.Group, r.Kind, bulkManifestPath)
			continue
		}
		files, err := m.FilesFor(repoRoot, r)
		if err != nil {
			t.Errorf("FAIL: resolving files for %s: %v", r.Kind, err)
			continue
		}
		if len(files) == 0 {
			t.Errorf("FAIL: %s/%s is in %s but no per-resource files were found. Expected files named %s_*.go.",
				r.Group, r.Kind, bulkManifestPath, strings.ToLower(r.Kind))
		}
	}
}

// TestGreenfieldBulkTypesConformance checks the hand-edited per-resource Go
// files against the bulk generation bar.
//
// These are all syntax-level facts, so the files are parsed with go/parser
// rather than run through a full analysis pass.
func TestGreenfieldBulkTypesConformance(t *testing.T) {
	t.Parallel()

	m := loadBulkManifest(t)

	for _, r := range m.Resources() {
		files, err := m.FilesFor(repoRoot, r)
		if err != nil {
			t.Fatalf("resolving files for %s: %v", r.Kind, err)
		}
		for _, path := range files {
			for _, problem := range checkGoFile(t, path) {
				t.Errorf("FAIL: %s: %s", path, problem)
			}
		}
	}
}

// checkGoFile returns the conformance problems in a single hand-edited resource
// file. Returns nil when the file is clean.
func checkGoFile(t *testing.T, path string) []string {
	t.Helper()

	src, err := os.ReadFile(path)
	if err != nil {
		return []string{fmt.Sprintf("could not read: %v", err)}
	}

	var problems []string

	// Copyright header. The generator emits 2025; new files must say 2026.
	if !strings.Contains(string(src), "Copyright 2026 Google LLC") {
		problems = append(problems,
			"missing `// Copyright 2026 Google LLC` header (the generator emits 2025 - fix it by hand)")
	}

	// refs.NormalizeWithFallback is not permitted for greenfield resources.
	if strings.Contains(string(src), "NormalizeWithFallback") {
		problems = append(problems,
			"uses refs.NormalizeWithFallback; greenfield resources must use refs.Normalize")
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return append(problems, fmt.Sprintf("could not parse: %v", err))
	}

	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			return true
		}
		for _, field := range st.Fields.List {
			if len(field.Names) == 0 {
				continue // embedded, e.g. *parent.ProjectAndLocationRef
			}
			if !field.Names[0].IsExported() {
				continue
			}
			if problem := checkFieldType(ts.Name.Name, field); problem != "" {
				problems = append(problems, problem)
			}
		}
		return true
	})

	return problems
}

// TestGreenfieldBulkCRDConformance checks the generated CRDs for the bulk
// resources. Greenfield resources are v1alpha1 only.
func TestGreenfieldBulkCRDConformance(t *testing.T) {
	t.Parallel()

	m := loadBulkManifest(t)
	crds, err := crdloader.LoadAllCRDs()
	if err != nil {
		t.Fatalf("loading CRDs: %v", err)
	}

	for _, r := range m.Resources() {
		crd, ok := greenfield.FindCRD(crds, r)
		if !ok {
			continue // reported by TestGreenfieldBulkManifestIsResolvable
		}
		for _, version := range crd.Spec.Versions {
			if version.Name != "v1alpha1" {
				t.Errorf("FAIL: crd=%s has version %q; greenfield resources must be v1alpha1 only",
					crd.Name, version.Name)
			}
		}
	}
}

// TestGreenfieldBulkFieldCoverage fails when a bulk resource still has fields
// recorded in alpha-missingfields.txt, i.e. fields no test fixture exercises.
//
// This is a local-only nudge, not a CI gate: Step 1 of the bulk workflow ships
// types and CRDs with no fixtures, so every new resource legitimately fails this
// until a later phase adds them. Run it deliberately with:
//
//	GREENFIELD_STRICT=1 go test ./tests/apichecks/... -run TestGreenfieldBulkFieldCoverage
//
// Opt-in rather than skip-when-CI-detected: if the environment check were ever
// wrong, opt-in simply does not run, instead of breaking CI.
//
// Fields that genuinely cannot be covered belong in
// testdata/exceptions/greenfield_fields_accepted.txt with a reason.
func TestGreenfieldBulkFieldCoverage(t *testing.T) {
	t.Parallel()

	if os.Getenv("GREENFIELD_STRICT") == "" {
		t.Skip("set GREENFIELD_STRICT=1 to run; this is a local-only check, not a CI gate")
	}

	m := loadBulkManifest(t)
	crds, err := crdloader.LoadAllCRDs()
	if err != nil {
		t.Fatalf("loading CRDs: %v", err)
	}

	missing, err := readBaselineLines("testdata/exceptions/alpha-missingfields.txt")
	if err != nil {
		t.Fatalf("reading alpha-missingfields.txt: %v", err)
	}
	accepted, err := readBaselineLines("testdata/exceptions/greenfield_fields_accepted.txt")
	if err != nil {
		t.Fatalf("reading greenfield_fields_accepted.txt: %v", err)
	}

	// Index accepted fields by "crd=<name>|<field path>".
	acceptedFields := make(map[string]bool)
	for _, line := range accepted {
		crdName, fieldPath, ok := parseBaselineEntry(line)
		if !ok {
			t.Errorf("FAIL: unparseable entry in greenfield_fields_accepted.txt: %q", line)
			continue
		}
		if !strings.Contains(line, "reason=") {
			t.Errorf("FAIL: entry in greenfield_fields_accepted.txt has no reason=: %q", line)
			continue
		}
		acceptedFields[crdName+"|"+fieldPath] = true
	}

	for _, r := range m.Resources() {
		crd, ok := greenfield.FindCRD(crds, r)
		if !ok {
			continue
		}
		var uncovered []string
		for _, line := range missing {
			crdName, fieldPath, ok := parseBaselineEntry(line)
			if !ok || crdName != crd.Name {
				continue
			}
			if acceptedFields[crdName+"|"+fieldPath] {
				continue
			}
			uncovered = append(uncovered, fieldPath)
		}
		if len(uncovered) > 0 {
			t.Errorf("FAIL: crd=%s has %d field(s) not exercised by any test fixture:\n  %s\n\n"+
				"Add fixture coverage, or record genuinely uncoverable fields in "+
				"testdata/exceptions/greenfield_fields_accepted.txt with a reason=.",
				crd.Name, len(uncovered), strings.Join(uncovered, "\n  "))
		}
	}
}

// readBaselineLines returns the non-empty, non-comment lines of an exceptions file.
// A missing file is treated as empty, so a not-yet-created accepted-list is fine.
func readBaselineLines(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	return out, nil
}

// parseBaselineEntry extracts the CRD name and field path from an apichecks
// exceptions line, which look like:
//
//	[missing_field] crd=foo.example.com version=v1alpha1: field ".spec.bar" is not set ...
func parseBaselineEntry(line string) (crdName string, fieldPath string, ok bool) {
	_, rest, found := strings.Cut(line, "crd=")
	if !found {
		return "", "", false
	}
	crdName, rest, found = strings.Cut(rest, " ")
	if !found {
		return "", "", false
	}
	_, rest, found = strings.Cut(rest, `field "`)
	if !found {
		return "", "", false
	}
	fieldPath, _, found = strings.Cut(rest, `"`)
	if !found {
		return "", "", false
	}
	return crdName, fieldPath, true
}

// scalarKinds are the Go primitives that must be represented as pointers so that
// "unset" is distinguishable from "zero".
var scalarKinds = map[string]bool{
	"string": true, "bool": true, "int": true, "int32": true,
	"int64": true, "float32": true, "float64": true, "byte": true,
}

// checkFieldType enforces the pointer rules:
//   - scalar primitives must be pointers
//   - slices and maps must NOT be pointers
func checkFieldType(structName string, field *ast.Field) string {
	name := field.Names[0].Name

	switch t := field.Type.(type) {
	case *ast.Ident:
		if scalarKinds[t.Name] {
			return fmt.Sprintf("%s.%s is %s; scalar primitives must be pointers (*%s)",
				structName, name, t.Name, t.Name)
		}
	case *ast.StarExpr:
		// Pointer to a slice or map is wrong; pointer to anything else is fine.
		switch t.X.(type) {
		case *ast.ArrayType:
			return fmt.Sprintf("%s.%s is a pointer to a slice; slices must not be pointers", structName, name)
		case *ast.MapType:
			return fmt.Sprintf("%s.%s is a pointer to a map; maps must not be pointers", structName, name)
		}
	}
	return ""
}
