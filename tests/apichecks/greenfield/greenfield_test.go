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

package greenfield

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
	return path
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []Resource
		wantErr string
	}{
		{
			name:    "entries, comments and blank lines",
			content: "# a comment\n\nfoo.cnrm.cloud.google.com/FooBar\n\n# another\nbaz.cnrm.cloud.google.com/BazQux\n",
			want: []Resource{
				{Group: "foo.cnrm.cloud.google.com", Kind: "FooBar"},
				{Group: "baz.cnrm.cloud.google.com", Kind: "BazQux"},
			},
		},
		{
			name:    "empty file",
			content: "# only comments\n",
			want:    nil,
		},
		{
			name:    "malformed line is rejected",
			content: "NotAGroupSlashKind\n",
			wantErr: `expected "<group>/<Kind>"`,
		},
		{
			name:    "empty kind is rejected",
			content: "foo.cnrm.cloud.google.com/\n",
			wantErr: `expected "<group>/<Kind>"`,
		},
		{
			// A duplicate would silently do nothing, but it signals a mistake in a
			// file that is appended to by hand.
			name:    "duplicate entry is rejected",
			content: "foo.cnrm.cloud.google.com/FooBar\nfoo.cnrm.cloud.google.com/FooBar\n",
			wantErr: "duplicate entry",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := writeTemp(t, "manifest.txt", tc.content)
			got, err := Load(path)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tc.want, got.Resources(), cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Resources() diff (-want +got):\n%s", diff)
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join(t.TempDir(), "nope.txt")); err == nil {
		t.Error("expected an error for a missing manifest, got nil")
	}
}

func TestIsBulk(t *testing.T) {
	path := writeTemp(t, "manifest.txt", "foo.cnrm.cloud.google.com/FooBar\n")
	m, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	in := schema.GroupKind{Group: "foo.cnrm.cloud.google.com", Kind: "FooBar"}
	if !m.IsBulk(in) {
		t.Errorf("IsBulk(%v) = false, want true", in)
	}
	// Same Kind, different group: scope must not leak across groups.
	out := schema.GroupKind{Group: "other.cnrm.cloud.google.com", Kind: "FooBar"}
	if m.IsBulk(out) {
		t.Errorf("IsBulk(%v) = true, want false", out)
	}
}

func TestResourceService(t *testing.T) {
	r := Resource{Group: "networkservices.cnrm.cloud.google.com", Kind: "NetworkServicesLBTrafficExtension"}
	if got, want := r.Service(), "networkservices"; got != want {
		t.Errorf("Service() = %q, want %q", got, want)
	}
}

// TestDroppedFields is the important one: Spec and ObservedState map the same
// proto message, so each reports the other's fields as MISSING. A field is only
// genuinely dropped when it is missing from both.
func TestDroppedFields(t *testing.T) {
	const mapper = `
func FooSpec_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Foo) *krm.FooSpec {
	// MISSING: Name
	// MISSING: Labels
	out.Description = direct.LazyPtr(in.GetDescription())
	out.BarRefs = something()
	return out
}
func FooObservedState_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Foo) *krm.FooObservedState {
	// MISSING: Name
	// MISSING: Labels
	// MISSING: Description
	// MISSING: Bar
	out.CreateTime = direct.StringTimestamp_FromProto(mapCtx, in.GetCreateTime())
	return out
}
`

	path := writeTemp(t, "mapper.generated.go", mapper)
	got, err := DroppedFields(path, "Foo")
	if err != nil {
		t.Fatalf("DroppedFields: %v", err)
	}

	// Name and Labels are missing from both -> genuinely dropped.
	// Description and Bar are mapped in Spec, so their appearance in the
	// ObservedState mapper is cross-talk, not a drop.
	want := []string{"Labels", "Name"}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("DroppedFields diff (-want +got):\n%s", diff)
	}
}

func TestDroppedFieldsNoObservedState(t *testing.T) {
	// A resource with no ObservedState mapper falls back to the Spec list alone.
	const mapper = `
func FooSpec_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Foo) *krm.FooSpec {
	// MISSING: Name
	out.Description = direct.LazyPtr(in.GetDescription())
	return out
}
`
	path := writeTemp(t, "mapper.generated.go", mapper)
	got, err := DroppedFields(path, "Foo")
	if err != nil {
		t.Fatalf("DroppedFields: %v", err)
	}
	if diff := cmp.Diff([]string{"Name"}, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("DroppedFields diff (-want +got):\n%s", diff)
	}
}

func TestDroppedFieldsOtherKindsIgnored(t *testing.T) {
	// Nested and unrelated types in the same mapper file must not be attributed
	// to this Kind.
	const mapper = `
func FooSpec_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Foo) *krm.FooSpec {
	out.Description = direct.LazyPtr(in.GetDescription())
	return out
}
func BarSpec_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Bar) *krm.BarSpec {
	// MISSING: SomethingElse
	return out
}
func Nested_Type_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Nested) *krm.Nested {
	// MISSING: AlsoNotMine
	return out
}
`
	path := writeTemp(t, "mapper.generated.go", mapper)
	got, err := DroppedFields(path, "Foo")
	if err != nil {
		t.Fatalf("DroppedFields: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("DroppedFields = %v, want none", got)
	}
}

func TestDroppedFieldsMissingMapperFile(t *testing.T) {
	// A service with no generated mapper yet is not an error.
	got, err := DroppedFields(filepath.Join(t.TempDir(), "nope.go"), "Foo")
	if err != nil {
		t.Fatalf("expected no error for a missing mapper file, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("DroppedFields = %v, want none", got)
	}
}

func TestFilesFor(t *testing.T) {
	root := t.TempDir()
	apiDir := filepath.Join(root, "apis", "networkservices", "v1alpha1")
	ctlDir := filepath.Join(root, "pkg", "controller", "direct", "networkservices")
	for _, d := range []string{apiDir, ctlDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}

	r := Resource{Group: "networkservices.cnrm.cloud.google.com", Kind: "NetworkServicesLBTrafficExtension"}
	prefix := strings.ToLower(r.Kind)

	// Files that belong to the resource, plus decoys that must not be picked up.
	want := []string{
		filepath.Join(apiDir, prefix+"_types.go"),
		filepath.Join(apiDir, prefix+"_identity.go"),
		filepath.Join(ctlDir, prefix+"_controller.go"),
	}
	decoys := []string{
		filepath.Join(apiDir, "types.generated.go"),                       // shared, deliberately excluded
		filepath.Join(ctlDir, "mapper.generated.go"),                      // shared, deliberately excluded
		filepath.Join(apiDir, "networkserviceslbrouteextension_types.go"), // a different resource
	}
	for _, p := range append(append([]string{}, want...), decoys...) {
		if err := os.WriteFile(p, []byte("package v1alpha1\n"), 0644); err != nil {
			t.Fatalf("writing %s: %v", p, err)
		}
	}

	m := &Manifest{}
	got, err := m.FilesFor(root, r)
	if err != nil {
		t.Fatalf("FilesFor: %v", err)
	}

	if diff := cmp.Diff(sorted(want), got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("FilesFor diff (-want +got):\n%s", diff)
	}
}

func sorted(in []string) []string {
	out := append([]string{}, in...)
	for i := range out {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func TestCheckGoSource(t *testing.T) {
	const header = "// Copyright 2026 Google LLC\n\npackage v1alpha1\n\n"

	tests := []struct {
		name        string
		src         string
		wantProblem string // substring; empty means the source must be clean
	}{
		{
			name: "clean",
			src: header + `type FooFields struct {
	Description *string ` + "`json:\"description,omitempty\"`" + `
	Labels map[string]string ` + "`json:\"labels,omitempty\"`" + `
	Items []string ` + "`json:\"items,omitempty\"`" + `
}`,
		},
		{
			name:        "wrong copyright year",
			src:         "// Copyright 2025 Google LLC\n\npackage v1alpha1\n",
			wantProblem: "Copyright 2026",
		},
		{
			name:        "no copyright at all",
			src:         "package v1alpha1\n",
			wantProblem: "Copyright 2026",
		},
		{
			name: "scalar must be a pointer",
			src: header + `type FooFields struct {
	Description string ` + "`json:\"description,omitempty\"`" + `
}`,
			wantProblem: "scalar primitives must be pointers",
		},
		{
			name: "bool scalar must be a pointer",
			src: header + `type FooFields struct {
	Enabled bool ` + "`json:\"enabled,omitempty\"`" + `
}`,
			wantProblem: "scalar primitives must be pointers",
		},
		{
			name: "slice must not be a pointer",
			src: header + `type FooFields struct {
	Items *[]string ` + "`json:\"items,omitempty\"`" + `
}`,
			wantProblem: "slices must not be pointers",
		},
		{
			name: "map must not be a pointer",
			src: header + `type FooFields struct {
	Labels *map[string]string ` + "`json:\"labels,omitempty\"`" + `
}`,
			wantProblem: "maps must not be pointers",
		},
		{
			name:        "NormalizeWithFallback is rejected",
			src:         header + "func f() { refs.NormalizeWithFallback(ctx, reader, ns) }\n",
			wantProblem: "must use refs.Normalize",
		},
		{
			// Embedded fields have no name; they must not panic or be flagged.
			name: "embedded field is ignored",
			src: header + `type FooFields struct {
	*parent.ProjectAndLocationRef ` + "`json:\",inline\"`" + `
}`,
		},
		{
			// Unexported fields are not part of the API surface.
			name: "unexported field is ignored",
			src: header + `type FooFields struct {
	internal string
}`,
		},
		{
			// Pointer to a struct is the normal case for optional messages.
			name: "pointer to struct is fine",
			src: header + `type FooFields struct {
	Nested *NestedType ` + "`json:\"nested,omitempty\"`" + `
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			problems, err := checkGoSource("test.go", []byte(tc.src))
			if err != nil {
				t.Fatalf("checkGoSource: %v", err)
			}
			if tc.wantProblem == "" {
				if len(problems) != 0 {
					t.Errorf("expected no problems, got %v", problems)
				}
				return
			}
			found := false
			for _, p := range problems {
				if strings.Contains(p, tc.wantProblem) {
					found = true
				}
			}
			if !found {
				t.Errorf("expected a problem containing %q, got %v", tc.wantProblem, problems)
			}
		})
	}
}

func TestParseBaselineEntry(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantCRD   string
		wantField string
		wantOK    bool
	}{
		{
			name:      "missing_field entry",
			line:      `[missing_field] crd=foos.example.com version=v1alpha1: field ".spec.bar" is not set in unstructured objects`,
			wantCRD:   "foos.example.com",
			wantField: ".spec.bar",
			wantOK:    true,
		},
		{
			name:      "dropped entry with reason",
			line:      `[dropped] crd=foos.example.com: field "Labels" reason=intentional`,
			wantCRD:   "foos.example.com",
			wantField: "Labels",
			wantOK:    true,
		},
		{
			name:      "nested field path",
			line:      `[missing_field] crd=foos.example.com version=v1alpha1: field ".spec.a[].b.c" is not set`,
			wantCRD:   "foos.example.com",
			wantField: ".spec.a[].b.c",
			wantOK:    true,
		},
		{
			name:   "no crd token",
			line:   `[missing_field] version=v1alpha1: field ".spec.bar" is not set`,
			wantOK: false,
		},
		{
			name:   "no field token",
			line:   `[missing_field] crd=foos.example.com version=v1alpha1: something else`,
			wantOK: false,
		},
		{
			name:   "unterminated field quote",
			line:   `[missing_field] crd=foos.example.com version=v1alpha1: field ".spec.bar`,
			wantOK: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			crd, field, ok := ParseBaselineEntry(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !tc.wantOK {
				return
			}
			if crd != tc.wantCRD {
				t.Errorf("crd = %q, want %q", crd, tc.wantCRD)
			}
			if field != tc.wantField {
				t.Errorf("field = %q, want %q", field, tc.wantField)
			}
		})
	}
}

func TestBaselineLines(t *testing.T) {
	path := writeTemp(t, "baseline.txt", "# comment\n\nentry one\n  entry two  \n\n# trailing comment\n")
	got, err := BaselineLines(path)
	if err != nil {
		t.Fatalf("BaselineLines: %v", err)
	}
	want := []string{"entry one", "entry two"}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("BaselineLines diff (-want +got):\n%s", diff)
	}
}

func TestBaselineLinesMissingFile(t *testing.T) {
	got, err := BaselineLines(filepath.Join(t.TempDir(), "nope.txt"))
	if err != nil {
		t.Fatalf("expected no error for a missing file, got %v", err)
	}
	if len(got) != 0 {
		t.Errorf("BaselineLines = %v, want none", got)
	}
}

func TestFindCRD(t *testing.T) {
	crds := []apiextensions.CustomResourceDefinition{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "foos.example.com"},
			Spec: apiextensions.CustomResourceDefinitionSpec{
				Group: "example.com",
				Names: apiextensions.CustomResourceDefinitionNames{Kind: "Foo"},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "bars.other.com"},
			Spec: apiextensions.CustomResourceDefinitionSpec{
				Group: "other.com",
				Names: apiextensions.CustomResourceDefinitionNames{Kind: "Bar"},
			},
		},
	}

	got, ok := FindCRD(crds, Resource{Group: "example.com", Kind: "Foo"})
	if !ok {
		t.Fatal("FindCRD did not find an existing CRD")
	}
	if got.Name != "foos.example.com" {
		t.Errorf("CRD name = %q, want %q", got.Name, "foos.example.com")
	}

	// Right Kind, wrong group: must not match.
	if _, ok := FindCRD(crds, Resource{Group: "example.com", Kind: "Bar"}); ok {
		t.Error("FindCRD matched on Kind alone, ignoring the group")
	}
	if _, ok := FindCRD(crds, Resource{Group: "example.com", Kind: "Nope"}); ok {
		t.Error("FindCRD matched a nonexistent Kind")
	}
}

func TestCheckGoSourceProtoAnnotation(t *testing.T) {
	const header = "// Copyright 2026 Google LLC\n\npackage v1alpha1\n\n"

	tests := []struct {
		name        string
		src         string
		wantProblem string
	}{
		{
			name: "spec with annotation is clean",
			src: header + `// FooSpec defines the desired state.
// +kcc:spec:proto=google.cloud.foo.v1.Foo
type FooSpec struct {
	Description *string ` + "`json:\"description,omitempty\"`" + `
}`,
		},
		{
			name: "spec without annotation is flagged",
			src: header + `// FooSpec defines the desired state.
type FooSpec struct {
	Description *string ` + "`json:\"description,omitempty\"`" + `
}`,
			wantProblem: "missing the `// +kcc:spec:proto=",
		},
		{
			name: "observedstate with annotation is clean",
			src: header + `// FooObservedState is observed state.
// +kcc:observedstate:proto=google.cloud.foo.v1.Foo
type FooObservedState struct {
	CreateTime *string ` + "`json:\"createTime,omitempty\"`" + `
}`,
		},
		{
			name: "observedstate without annotation is flagged",
			src: header + `// FooObservedState is observed state.
type FooObservedState struct {
	CreateTime *string ` + "`json:\"createTime,omitempty\"`" + `
}`,
			wantProblem: "missing the `// +kcc:observedstate:proto=",
		},
		{
			// Structs that are neither Spec nor ObservedState carry a different
			// marker and must not be flagged by this rule.
			name: "nested type is not required to have the spec annotation",
			src: header + `type ExtensionChain struct {
	Name *string ` + "`json:\"name,omitempty\"`" + `
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			problems, err := checkGoSource("test.go", []byte(tc.src))
			if err != nil {
				t.Fatalf("checkGoSource: %v", err)
			}
			assertProblem(t, problems, tc.wantProblem)
		})
	}
}

func TestCheckGoSourceObservedGeneration(t *testing.T) {
	const header = "// Copyright 2026 Google LLC\n\npackage v1alpha1\n\n"

	tests := []struct {
		name        string
		field       string
		wantProblem string
	}{
		{name: "correct type", field: "ObservedGeneration *int64 `json:\"observedGeneration,omitempty\"`"},
		{name: "non-pointer", field: "ObservedGeneration int64 `json:\"observedGeneration,omitempty\"`", wantProblem: "must be exactly *int64"},
		{name: "wrong width", field: "ObservedGeneration *int32 `json:\"observedGeneration,omitempty\"`", wantProblem: "must be exactly *int64"},
		{name: "string", field: "ObservedGeneration *string `json:\"observedGeneration,omitempty\"`", wantProblem: "must be exactly *int64"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			src := header + "type FooStatus struct {\n\t" + tc.field + "\n}"
			problems, err := checkGoSource("test.go", []byte(src))
			if err != nil {
				t.Fatalf("checkGoSource: %v", err)
			}
			assertProblem(t, problems, tc.wantProblem)
		})
	}
}

func TestCheckGoSourceEnumMarkerProhibited(t *testing.T) {
	const header = "// Copyright 2026 Google LLC\n\npackage v1alpha1\n\n"

	tests := []struct {
		name        string
		src         string
		wantProblem string
	}{
		{
			// The marker hardcodes a value list the GCP API already enforces, and
			// couples KCC releases to GCP enum additions.
			name: "enum marker is prohibited",
			src: header + `type FooFields struct {
	// +kubebuilder:validation:Enum=A;B
	Scheme *string ` + "`json:\"scheme,omitempty\"`" + `
}`,
			wantProblem: "do not hardcode enum values",
		},
		{
			name: "prohibited even when the type is already *string",
			src: header + `type FooFields struct {
	// +kubebuilder:validation:Enum=INTERNAL_MANAGED;EXTERNAL_MANAGED
	LoadBalancingScheme *string ` + "`json:\"loadBalancingScheme,omitempty\"`" + `
}`,
			wantProblem: "do not hardcode enum values",
		},
		{
			name: "plain *string with no marker is clean",
			src: header + `type FooFields struct {
	Scheme *string ` + "`json:\"scheme,omitempty\"`" + `
}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			problems, err := checkGoSource("test.go", []byte(tc.src))
			if err != nil {
				t.Fatalf("checkGoSource: %v", err)
			}
			assertProblem(t, problems, tc.wantProblem)
		})
	}
}

func TestNestedDroppedFields(t *testing.T) {
	const mapper = `
func FooSpec_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Foo) *krm.FooSpec {
	// MISSING: KindLevelField
	return out
}
func FooObservedState_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Foo) *krm.FooObservedState {
	// MISSING: KindLevelField
	return out
}
func ExtensionChain_Extension_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.X) *krm.X {
	// MISSING: Service
	return out
}
func ExtensionChain_Extension_v1alpha1_ToProto(mapCtx *direct.MapContext, in *krm.X) *pb.X {
	// MISSING: Service
	return out
}
func SharedThing_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Y) *krm.Y {
	// MISSING: Other
	return out
}
`

	path := writeTemp(t, "mapper.generated.go", mapper)
	got, err := NestedDroppedFields(path, []string{"Foo"})
	if err != nil {
		t.Fatalf("NestedDroppedFields: %v", err)
	}

	want := []NestedDrop{
		{Type: "ExtensionChain_Extension", Field: "Service"},
		{Type: "SharedThing", Field: "Other"},
	}
	if diff := cmp.Diff(want, got, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("NestedDroppedFields diff (-want +got):\n%s", diff)
	}
}

func TestNestedDroppedFieldsExcludesKindLevel(t *testing.T) {
	// Kind-level types are DroppedFields' job; reporting them here too would
	// duplicate every entry under a second key.
	const mapper = `
func BarSpec_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.Bar) *krm.BarSpec {
	// MISSING: Mine
	return out
}
func BarObservedState_v1alpha1_ToProto(mapCtx *direct.MapContext, in *krm.BarObservedState) *pb.Bar {
	// MISSING: Mine
	return out
}
`
	path := writeTemp(t, "mapper.generated.go", mapper)
	got, err := NestedDroppedFields(path, []string{"Bar"})
	if err != nil {
		t.Fatalf("NestedDroppedFields: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no nested drops for Kind-level types, got %v", got)
	}
}

func TestNestedDroppedFieldsDeduplicates(t *testing.T) {
	// The same type appears in both FromProto and ToProto; the drop is one fact,
	// not two.
	const mapper = `
func Shared_v1alpha1_FromProto(mapCtx *direct.MapContext, in *pb.S) *krm.S {
	// MISSING: Field
	return out
}
func Shared_v1alpha1_ToProto(mapCtx *direct.MapContext, in *krm.S) *pb.S {
	// MISSING: Field
	return out
}
`
	path := writeTemp(t, "mapper.generated.go", mapper)
	got, err := NestedDroppedFields(path, nil)
	if err != nil {
		t.Fatalf("NestedDroppedFields: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("expected 1 deduplicated drop, got %v", got)
	}
}

func TestCheckShellFile(t *testing.T) {
	t.Run("correct year", func(t *testing.T) {
		p := writeTemp(t, "generate.sh", "#!/bin/bash\n# Copyright 2026 Google LLC\n")
		problems, err := CheckShellFile(p)
		if err != nil {
			t.Fatalf("CheckShellFile: %v", err)
		}
		if len(problems) != 0 {
			t.Errorf("expected no problems, got %v", problems)
		}
	})

	t.Run("wrong year", func(t *testing.T) {
		p := writeTemp(t, "generate.sh", "#!/bin/bash\n# Copyright 2025 Google LLC\n")
		problems, err := CheckShellFile(p)
		if err != nil {
			t.Fatalf("CheckShellFile: %v", err)
		}
		if len(problems) == 0 {
			t.Error("expected a problem for the 2025 header, got none")
		}
	})

	t.Run("missing file is not an error", func(t *testing.T) {
		problems, err := CheckShellFile(filepath.Join(t.TempDir(), "nope.sh"))
		if err != nil {
			t.Fatalf("expected no error for a missing file, got %v", err)
		}
		if len(problems) != 0 {
			t.Errorf("expected no problems, got %v", problems)
		}
	})
}

// assertProblem checks that problems contains want (or is clean when want is "").
func assertProblem(t *testing.T, problems []string, want string) {
	t.Helper()
	if want == "" {
		for _, p := range problems {
			// The copyright rule is satisfied by the shared header in these
			// fixtures; anything else is an unexpected finding.
			t.Errorf("expected no problems, got %q", p)
		}
		return
	}
	for _, p := range problems {
		if strings.Contains(p, want) {
			return
		}
	}
	t.Errorf("expected a problem containing %q, got %v", want, problems)
}
