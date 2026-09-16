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
//	Discovery  no field_behavior anywhere, like compute's protos
//	Annotated  one field marked OUTPUT_ONLY, the rest bare
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
	// Only "description" is annotated. No allowlisted field carries an
	// annotation itself, so the guard has to come from the message, not the
	// field.
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

	tests := []struct {
		name  string
		msg   protoreflect.MessageDescriptor
		field string
		opts  WriteOptions
		want  bool
	}{
		{"allowlisted, nothing annotated", discovery, "creation_timestamp", on, true},
		{"allowlisted, nothing annotated", discovery, "self_link", on, true},
		{"etag is in the list", discovery, "etag", on, true},

		// These three names were considered and rejected. state and status are
		// desired state upstream, not observed state; type appears in 36 Specs.
		{"state stays in the Spec", discovery, "state", on, false},
		{"status stays in the Spec", discovery, "status", on, false},
		{"type stays in the Spec", discovery, "type", on, false},
		// identityFields in the scaffold package handles name instead.
		{"name is left to the identity policy", discovery, "name", on, false},
		{"an ordinary field is untouched", discovery, "description", on, false},

		// This is the guard. One annotation anywhere on the message means the
		// author placed the other fields on purpose.
		{"annotated message, allowlisted field", annotated, "creation_timestamp", on, false},
		{"annotated message, etag", annotated, "etag", on, false},

		{"off by default", discovery, "creation_timestamp", WriteOptions{}, false},
	}
	for _, g := range tests {
		t.Run(g.name+"/"+g.field, func(t *testing.T) {
			// Arrange
			f := g.msg.Fields().ByName(protoreflect.Name(g.field))
			if f == nil {
				t.Fatalf("no field %q on %s", g.field, g.msg.Name())
			}

			// Act
			got := IsServerSetField(f, g.msg, g.opts)

			// Assert
			if got != g.want {
				t.Errorf("IsServerSetField(%s.%s) = %v, want %v", g.msg.Name(), g.field, got, g.want)
			}
		})
	}
}

// TestIsServerSetFieldOnlyAppliesToTheRootMessage pins where the rule stops:
// the same field name answers true on the resource's own message and false on
// a nested one, where the user often sets it.
func TestIsServerSetFieldOnlyAppliesToTheRootMessage(t *testing.T) {
	// Arrange
	fd := serverSetTestFile(t)
	discovery := fd.Messages().ByName("Discovery")
	nested := fd.Messages().ByName("Annotated")
	g := &TypeGenerator{
		writeOptions:   WriteOptions{PlaceServerSetFields: true},
		rootMessageFQN: string(discovery.FullName()),
	}

	// Act
	root := g.isServerSet(discovery.Fields().ByName("creation_timestamp"), discovery)
	child := g.isServerSet(nested.Fields().ByName("creation_timestamp"), nested)

	// Assert
	if !root {
		t.Error("the root message's creationTimestamp is not server-set, want server-set")
	}
	if child {
		t.Error("a nested message's creationTimestamp is server-set, want not server-set")
	}
}
