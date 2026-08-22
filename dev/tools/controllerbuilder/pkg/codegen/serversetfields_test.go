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

import (
	"testing"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

// serverSetTestFile builds two messages:
//
//	Discovery  -- no field_behavior anywhere, the compute shape
//	Annotated  -- one field marked OUTPUT_ONLY, the rest bare
//
// Both carry the same field names, so a test can isolate the guard from the
// allowlist.
func serverSetTestFile(t *testing.T) protoreflect.FileDescriptor {
	t.Helper()

	outputOnly := &descriptorpb.FieldOptions{}
	proto.SetExtension(outputOnly, annotations.E_FieldBehavior,
		[]annotations.FieldBehavior{annotations.FieldBehavior_OUTPUT_ONLY})

	names := []string{"creation_timestamp", "self_link", "etag", "state", "status", "type", "name", "description"}
	bare := func() []*descriptorpb.FieldDescriptorProto {
		var out []*descriptorpb.FieldDescriptorProto
		for i, n := range names {
			out = append(out, &descriptorpb.FieldDescriptorProto{
				Name:   protoPtr(n),
				Number: protoPtr(int32(i + 1)),
				Type:   typeDescriptor(descriptorpb.FieldDescriptorProto_TYPE_STRING),
			})
		}
		return out
	}

	annotated := bare()
	// "description" is the annotated one, so no allowlisted field is itself
	// annotated -- the guard has to come from the message, not the field.
	annotated[len(annotated)-1].Options = outputOnly

	fdp := &descriptorpb.FileDescriptorProto{
		Name:    protoPtr("serverset.proto"),
		Package: protoPtr("google.cloud.test.v1"),
		MessageType: []*descriptorpb.DescriptorProto{
			{Name: protoPtr("Discovery"), Field: bare()},
			{Name: protoPtr("Annotated"), Field: annotated},
		},
	}
	fd, err := protodesc.NewFile(fdp, nil)
	if err != nil {
		t.Fatalf("building file descriptor: %v", err)
	}
	return fd
}

func TestIsServerSetField(t *testing.T) {
	fd := serverSetTestFile(t)
	discovery := fd.Messages().ByName("Discovery")
	annotated := fd.Messages().ByName("Annotated")
	on := WriteOptions{PlaceServerSetFields: true}

	grid := []struct {
		name  string
		msg   protoreflect.MessageDescriptor
		field string
		opts  WriteOptions
		want  bool
	}{
		{"allowlisted, nothing annotated", discovery, "creation_timestamp", on, true},
		{"allowlisted, nothing annotated", discovery, "self_link", on, true},
		{"etag is in the list", discovery, "etag", on, true},

		// The three names that were considered and rejected. state and status are
		// desired state upstream, not observed state; type appears in 36 Specs.
		{"state stays in the Spec", discovery, "state", on, false},
		{"status stays in the Spec", discovery, "status", on, false},
		{"type stays in the Spec", discovery, "type", on, false},
		// Handled by identityFields in the scaffold package instead.
		{"name is left to the identity policy", discovery, "name", on, false},
		{"an ordinary field is untouched", discovery, "description", on, false},

		// The guard: one annotation anywhere on the message means the author made
		// a decision, so their silence about the rest is respected.
		{"annotated message, allowlisted field", annotated, "creation_timestamp", on, false},
		{"annotated message, etag", annotated, "etag", on, false},

		{"off by default", discovery, "creation_timestamp", WriteOptions{}, false},
	}
	for _, g := range grid {
		t.Run(g.name+"/"+g.field, func(t *testing.T) {
			f := g.msg.Fields().ByName(protoreflect.Name(g.field))
			if f == nil {
				t.Fatalf("no field %q on %s", g.field, g.msg.Name())
			}
			if got := IsServerSetField(f, g.msg, g.opts); got != g.want {
				t.Errorf("IsServerSetField(%s.%s) = %v, want %v", g.msg.Name(), g.field, got, g.want)
			}
		})
	}
}

// A nested message's "id" or "kind" is often genuine user input, so the rule is
// restricted to the resource's own message. identifyOutputs recurses, which is
// what makes the restriction necessary rather than decorative.
func TestIsServerSetFieldOnlyAppliesToTheRootMessage(t *testing.T) {
	fd := serverSetTestFile(t)
	discovery := fd.Messages().ByName("Discovery")
	nested := fd.Messages().ByName("Annotated")

	g := &TypeGenerator{
		writeOptions:   WriteOptions{PlaceServerSetFields: true},
		rootMessageFQN: string(discovery.FullName()),
	}

	rootField := discovery.Fields().ByName("creation_timestamp")
	if !g.isServerSet(rootField, discovery) {
		t.Error("expected the root message's creationTimestamp to be server-set")
	}
	// Same field name, different message: not the resource's own.
	nestedField := nested.Fields().ByName("creation_timestamp")
	if g.isServerSet(nestedField, nested) {
		t.Error("a nested message's creationTimestamp must not be treated as server-set")
	}
}
