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

package protoapi

import (
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// ParentStyle names the shape of a resource's parent, as declared by the
// google.api.resource pattern.
type ParentStyle string

const (
	// ParentProjectLocation is "projects/*/locations/*" - the most common shape,
	// and the one the identity template assumes unconditionally today.
	ParentProjectLocation ParentStyle = "project_location"
	// ParentProject is "projects/*".
	ParentProject ParentStyle = "project"
	// ParentOrganization is "organizations/*".
	ParentOrganization ParentStyle = "organization"
	// ParentFolder is "folders/*".
	ParentFolder ParentStyle = "folder"
	// ParentOther is anything else: deeper nesting, or a parent we do not model.
	// Callers should treat this as "a human must decide" rather than guessing.
	ParentOther ParentStyle = "other"
	// ParentUnknown means the proto declared no pattern for us to read.
	ParentUnknown ParentStyle = "unknown"
)

// ResourceMetadata holds what google.api.resource states about a resource. The
// scaffolding templates would otherwise guess these, and measurement says the
// guesses are wrong more often than not: the naive "lowercase name + s"
// collection segment disagrees with the declared pattern for 852 of 1417
// annotated messages, and the assumed projects/locations parent holds for only
// about a third.
type ResourceMetadata struct {
	// Pattern is the first declared resource name pattern, e.g.
	// "projects/{project}/locations/{location}/lbTrafficExtensions/{extension}".
	Pattern string
	// Plural is the declared plural, when set. Only about a quarter of annotated
	// messages set it, so Collection is usually the more reliable source.
	Plural string
	// Collection is the segment naming this resource's collection, taken from the
	// pattern, e.g. "lbTrafficExtensions". This is the value that belongs in a
	// resource name, preserving the API's own casing.
	Collection string
	// ParentSegments are the literal collection segments of the parent, e.g.
	// ["projects", "locations"].
	ParentSegments []string
	// ParentStyle classifies ParentSegments into a shape the templates can render.
	ParentStyle ParentStyle
}

// GetResourceMetadata reads google.api.resource from a message. It returns nil
// when the message carries no usable pattern, which is the common case: only
// 1417 of ~24900 messages are annotated resources.
func GetResourceMetadata(msg protoreflect.MessageDescriptor) *ResourceMetadata {
	if msg == nil {
		return nil
	}
	v := proto.GetExtension(msg.Options(), annotations.E_Resource)
	rd, _ := v.(*annotations.ResourceDescriptor)
	if rd == nil {
		return nil
	}
	patterns := rd.GetPattern()
	if len(patterns) == 0 {
		return nil
	}

	md := &ResourceMetadata{
		Pattern: patterns[0],
		Plural:  rd.GetPlural(),
	}
	md.Collection, md.ParentSegments = splitPattern(md.Pattern)
	if md.Collection == "" && md.Plural != "" {
		md.Collection = md.Plural
	}
	md.ParentStyle = classifyParent(md.ParentSegments)
	return md
}

// splitPattern separates a resource name pattern into the resource's own
// collection segment and the literal collection segments of its parent.
//
// "projects/{project}/locations/{location}/foos/{foo}"
//
//	-> collection "foos", parent ["projects", "locations"]
func splitPattern(pattern string) (collection string, parentSegments []string) {
	segs := strings.Split(pattern, "/")

	// Literal segments alternate with {placeholders}. Anything that is not a
	// placeholder is a collection name.
	var literals []string
	for _, s := range segs {
		if s == "" {
			continue
		}
		if strings.HasPrefix(s, "{") {
			continue
		}
		literals = append(literals, s)
	}
	if len(literals) == 0 {
		return "", nil
	}

	// A pattern may end in a literal with no trailing placeholder (e.g.
	// "projects/{project}/locations"); in that case there is no distinct
	// collection for the resource itself.
	last := segs[len(segs)-1]
	if !strings.HasPrefix(last, "{") {
		return "", literals
	}
	return literals[len(literals)-1], literals[:len(literals)-1]
}

func classifyParent(segments []string) ParentStyle {
	switch strings.Join(segments, "/") {
	case "":
		return ParentUnknown
	case "projects/locations":
		return ParentProjectLocation
	case "projects":
		return ParentProject
	case "organizations":
		return ParentOrganization
	case "folders":
		return ParentFolder
	default:
		return ParentOther
	}
}
