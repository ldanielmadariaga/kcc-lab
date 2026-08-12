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
	"go/ast"
	"go/parser"
	"go/token"
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

// scalarKinds are the Go primitives that must be represented as pointers so that
// "unset" is distinguishable from "zero".
var scalarKinds = map[string]bool{
	"string": true, "bool": true, "int": true, "int32": true,
	"int64": true, "float32": true, "float64": true, "byte": true,
}

// CheckGoFile returns the conformance problems in a single hand-edited resource
// file, or nil when it is clean.
//
// Everything checked here is a syntax-level fact, so the file is parsed with
// go/parser; no type information is required.
func CheckGoFile(path string) ([]string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}
	return checkGoSource(path, src)
}

// checkGoSource is CheckGoFile against in-memory source, so the rules can be
// tested without touching the filesystem.
func checkGoSource(path string, src []byte) ([]string, error) {
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
		return append(problems, fmt.Sprintf("could not parse: %v", err)), nil
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

		// Spec and ObservedState structs must carry the proto annotation that
		// drives mapper generation. Missing it does not merely look untidy: the
		// mapper generator uses it, so the failure surfaces in a later phase as
		// rework rather than here as a diff.
		if problem := checkProtoAnnotation(ts, f); problem != "" {
			problems = append(problems, problem)
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
			if problem := checkObservedGeneration(ts.Name.Name, field); problem != "" {
				problems = append(problems, problem)
			}
			if problem := checkEnumField(ts.Name.Name, field); problem != "" {
				problems = append(problems, problem)
			}
		}
		return true
	})

	return problems, nil
}

// checkProtoAnnotation requires the +kcc:*:proto= marker on Spec and
// ObservedState structs. The mapper generator reads it; without it, mapper
// generation for the resource silently produces nothing usable.
func checkProtoAnnotation(ts *ast.TypeSpec, f *ast.File) string {
	name := ts.Name.Name

	var want string
	switch {
	case strings.HasSuffix(name, "ObservedState"):
		want = "+kcc:observedstate:proto="
	case strings.HasSuffix(name, "Spec"):
		want = "+kcc:spec:proto="
	default:
		return ""
	}

	// The marker sits in the doc comment of the enclosing GenDecl, which the
	// TypeSpec itself does not carry, so scan the declaration that holds it.
	for _, decl := range f.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			if spec != ast.Spec(ts) {
				continue
			}
			for _, cg := range []*ast.CommentGroup{gd.Doc, ts.Doc} {
				if cg == nil {
					continue
				}
				if strings.Contains(cg.Text(), strings.TrimPrefix(want, "+")) {
					return ""
				}
			}
		}
	}
	return fmt.Sprintf("%s is missing the `// %sgoogle...` annotation; mapper generation depends on it", name, want)
}

// checkObservedGeneration requires status.observedGeneration to be exactly
// *int64, per the base types skill.
func checkObservedGeneration(structName string, field *ast.Field) string {
	if field.Names[0].Name != "ObservedGeneration" {
		return ""
	}
	star, ok := field.Type.(*ast.StarExpr)
	if ok {
		if id, ok := star.X.(*ast.Ident); ok && id.Name == "int64" {
			return ""
		}
	}
	return fmt.Sprintf("%s.ObservedGeneration must be exactly *int64", structName)
}

// checkEnumField requires proto enum fields to be represented as *string rather
// than a custom wrapped string type, per the greenfield types skill.
//
// Detection is by the kubebuilder Enum validation marker: a field carrying
// +kubebuilder:validation:Enum is an enum by definition, whatever its Go type.
// This deliberately does not try to infer enum-ness from the type name, which
// would flag ordinary named string types.
func checkEnumField(structName string, field *ast.Field) string {
	if field.Doc == nil || !strings.Contains(field.Doc.Text(), "kubebuilder:validation:Enum") {
		return ""
	}
	if star, ok := field.Type.(*ast.StarExpr); ok {
		if id, ok := star.X.(*ast.Ident); ok && id.Name == "string" {
			return ""
		}
	}
	return fmt.Sprintf("%s.%s is an enum (has kubebuilder:validation:Enum) and must be *string, not a custom wrapped type",
		structName, field.Names[0].Name)
}

// CheckShellFile applies the copyright rule to generated shell scripts. The
// reviewgen skill requires the 2026 header on new .go AND .sh files; only .go
// was covered before.
func CheckShellFile(path string) ([]string, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}
	if !strings.Contains(string(src), "Copyright 2026 Google LLC") {
		return []string{"missing `# Copyright 2026 Google LLC` header"}, nil
	}
	return nil, nil
}

// GenerateScriptFor returns the service's generate.sh path, which is shared by
// every resource in the service rather than being per-resource.
func GenerateScriptFor(repoRoot string, r Resource) string {
	return filepath.Join(repoRoot, "apis", r.Service(), "generate.sh")
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

// BaselineLines returns the non-empty, non-comment lines of an exceptions file.
// A missing file is treated as empty, so a not-yet-created list is not an error.
func BaselineLines(path string) ([]string, error) {
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

// ParseBaselineEntry extracts the CRD name and field path from an apichecks
// exceptions line, which look like:
//
//	[missing_field] crd=foo.example.com version=v1alpha1: field ".spec.bar" is not set ...
func ParseBaselineEntry(line string) (crdName string, fieldPath string, ok bool) {
	_, rest, found := strings.Cut(line, "crd=")
	if !found {
		return "", "", false
	}
	crdName, rest, found = strings.Cut(rest, " ")
	if !found {
		return "", "", false
	}
	// Entry shapes differ: "crd=<name> version=v1alpha1: field ..." puts a space
	// after the name, while "crd=<name>: field ..." does not. Trim the separator
	// so both parse to the same CRD name.
	crdName = strings.TrimSuffix(crdName, ":")
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
