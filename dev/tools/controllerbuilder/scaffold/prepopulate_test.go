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
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/codegen"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

func strPtr(s string) *string { return &s }
func i32Ptr(i int32) *int32   { return &i }

func fieldType(t descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto_Type {
	return &t
}

func behaviorOptions(bs ...annotations.FieldBehavior) *descriptorpb.FieldOptions {
	o := &descriptorpb.FieldOptions{}
	proto.SetExtension(o, annotations.E_FieldBehavior, bs)
	return o
}

// testMessage builds a message with a representative mix: an identity field, an
// output-only field, a required field and a plain optional one.
func testMessage(t *testing.T) protoreflect.MessageDescriptor {
	t.Helper()
	fdp := &descriptorpb.FileDescriptorProto{
		Name:    strPtr("test.proto"),
		Package: strPtr("google.cloud.test.v1"),
		MessageType: []*descriptorpb.DescriptorProto{
			{
				Name: strPtr("Widget"),
				Field: []*descriptorpb.FieldDescriptorProto{
					{
						Name:    strPtr("name"),
						Number:  i32Ptr(1),
						Type:    fieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
						Options: behaviorOptions(annotations.FieldBehavior_IDENTIFIER),
					},
					{
						Name:    strPtr("create_time"),
						Number:  i32Ptr(2),
						Type:    fieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
						Options: behaviorOptions(annotations.FieldBehavior_OUTPUT_ONLY),
					},
					{
						Name:    strPtr("display_name"),
						Number:  i32Ptr(3),
						Type:    fieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
						Options: behaviorOptions(annotations.FieldBehavior_REQUIRED),
					},
					{
						Name:   strPtr("description"),
						Number: i32Ptr(4),
						Type:   fieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
					},
				},
			},
		},
	}
	fd, err := protodesc.NewFile(fdp, nil)
	if err != nil {
		t.Fatalf("building file descriptor: %v", err)
	}
	return fd.Messages().ByName("Widget")
}

func TestPrepopulateSpec(t *testing.T) {
	msg := testMessage(t)

	got, err := PrepopulateSpec(msg, codegen.WriteOptions{EmitRequired: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The rendered body has to be valid Go inside a struct.
	src := "package p\ntype S struct {\n" + got.SpecFields + "}\n"
	if _, err := parser.ParseFile(token.NewFileSet(), "s.go", src, parser.AllErrors); err != nil {
		t.Fatalf("rendered spec is not valid Go: %v\n\n%s", err, src)
	}

	for _, want := range []string{
		"DisplayName *string",
		"Description *string",
		"+kcc:proto:field=google.cloud.test.v1.Widget.display_name",
		"// +required",
	} {
		if !strings.Contains(got.SpecFields, want) {
			t.Errorf("missing %q in:\n%s", want, got.SpecFields)
		}
	}

	// name is the resource's own identity, expressed as metadata.name /
	// spec.resourceID, and create_time belongs in ObservedState.
	//
	// Match on the proto annotation, not the Go field name: "DisplayName *string"
	// contains "Name *string", so a substring check on the latter passes for the
	// wrong reason.
	for _, notWant := range []string{
		"+kcc:proto:field=google.cloud.test.v1.Widget.name\n",
		"+kcc:proto:field=google.cloud.test.v1.Widget.create_time",
	} {
		if strings.Contains(got.SpecFields, notWant) {
			t.Errorf("unexpected %q in:\n%s", notWant, got.SpecFields)
		}
	}
}

func TestPrepopulateSpecAlwaysQueuesTheResource(t *testing.T) {
	msg := testMessage(t)

	got, err := PrepopulateSpec(msg, codegen.WriteOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No field here carries google.api.resource_reference, which is the common
	// case: the pilot resource has none either, including on the field that has
	// to become a ref. Suppression must not depend on detection, so a
	// resource-level entry is always emitted.
	if len(got.Judgement) == 0 {
		t.Fatal("expected at least a resource-level judgement entry")
	}
	if got.Judgement[0].Reason != "untriaged-bulk-generation" {
		t.Errorf("first entry reason = %q, want untriaged-bulk-generation", got.Judgement[0].Reason)
	}
	if got.Judgement[0].FieldPath != "" {
		t.Errorf("resource-level entry should have no field path, got %q", got.Judgement[0].FieldPath)
	}
}

func TestPrepopulateSpecRequiresAMessage(t *testing.T) {
	if _, err := PrepopulateSpec(nil, codegen.WriteOptions{}); err == nil {
		t.Fatal("expected an error for a nil message")
	}
}

func TestFormatJudgementEntries(t *testing.T) {
	items := []JudgementItem{
		{Reason: "untriaged-bulk-generation", Detail: "generated mechanically"},
		{FieldPath: ".spec.network", Reason: "possible-reference", Detail: "target=compute.googleapis.com/Network"},
	}

	got := FormatJudgementEntries("ExampleWidget", "example.cnrm.cloud.google.com", items)
	lines := strings.Split(strings.TrimSpace(got), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want 2:\n%s", len(lines), got)
	}

	// Both forms must satisfy what the apichecks loader requires: kind=, group=
	// and reason= on every line.
	for _, l := range lines {
		for _, need := range []string{"kind=ExampleWidget", "group=example.cnrm.cloud.google.com", "reason="} {
			if !strings.Contains(l, need) {
				t.Errorf("line missing %q: %s", need, l)
			}
		}
		if head, _, _ := strings.Cut(l, ":"); strings.Contains(head, "reason=") {
			t.Errorf("reason must follow the colon so the head parses cleanly: %s", l)
		}
	}
	if !strings.Contains(lines[0], ": resource reason=") {
		t.Errorf("resource-level entry has the wrong shape: %s", lines[0])
	}
	if !strings.Contains(lines[1], `: field ".spec.network" reason=`) {
		t.Errorf("field-level entry has the wrong shape: %s", lines[1])
	}
}

// commentedMessage builds a message whose fields carry leading comments, which
// is what DetectOutputOnlyInComments reads. SourceCodeInfo paths are
// [4=message_type, msgIndex, 2=field, fieldIndex].
func commentedMessage(t *testing.T, comments ...string) protoreflect.MessageDescriptor {
	t.Helper()
	var fields []*descriptorpb.FieldDescriptorProto
	var locs []*descriptorpb.SourceCodeInfo_Location
	for i, c := range comments {
		fields = append(fields, &descriptorpb.FieldDescriptorProto{
			Name:   strPtr(fmt.Sprintf("field_%d", i)),
			Number: i32Ptr(int32(i + 1)),
			Type:   fieldType(descriptorpb.FieldDescriptorProto_TYPE_STRING),
		})
		locs = append(locs, &descriptorpb.SourceCodeInfo_Location{
			Path:            []int32{4, 0, 2, int32(i)},
			Span:            []int32{int32(i), 0, 1},
			LeadingComments: strPtr(" " + c + "\n"),
		})
	}
	fdp := &descriptorpb.FileDescriptorProto{
		Name:           strPtr("commented.proto"),
		Package:        strPtr("google.cloud.test.v1"),
		MessageType:    []*descriptorpb.DescriptorProto{{Name: strPtr("Widget"), Field: fields}},
		SourceCodeInfo: &descriptorpb.SourceCodeInfo{Location: locs},
	}
	fd, err := protodesc.NewFile(fdp, nil)
	if err != nil {
		t.Fatalf("building file descriptor: %v", err)
	}
	return fd.Messages().ByName("Widget")
}

func TestDetectOutputOnlyInComments(t *testing.T) {
	msg := commentedMessage(t,
		"Output only. Set by the server.",                   // field_0, the long-standing spelling
		"[Output Only] IP address on the Google side.",      // field_1, how Compute writes it
		"The display name of the widget.",                   // field_2, no signal
		"Set by the user. Output only in some other sense.", // field_3, marker not at the front
	)
	got := DetectOutputOnlyInComments(msg)

	var paths []string
	for _, c := range got {
		paths = append(paths, c.FieldPath)
	}
	want := []string{".spec.field0", ".spec.field1"}
	if len(paths) != len(want) {
		t.Fatalf("got %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("got %v, want %v", paths, want)
		}
	}
}

// A field the proto already annotates needs no prose detection. Reporting it
// would queue a field the generator has already placed correctly.
func TestDetectOutputOnlySkipsAnnotatedFields(t *testing.T) {
	msg := commentedMessage(t, "[Output Only] Set by the server.")
	if got := DetectOutputOnlyInComments(msg); len(got) != 1 {
		t.Fatalf("unannotated field should be reported, got %d", len(got))
	}
	if got := DetectOutputOnlyInComments(testMessage(t)); len(got) != 0 {
		t.Errorf("annotated fields should not be reported, got %v", got)
	}
}
