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
// scaffolding templates would otherwise guess these, and the guesses are wrong
// more often than they are right: the naive "lowercase name + s" collection
// segment disagrees with the declared pattern for 752 of 1417 annotated
// messages, and the assumed projects/locations parent holds for only about a
// third. See docs/ai/greenfield-generator-findings.md for the derivation.
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
	// ParentPath is the parent's literal collection segments, joined, e.g.
	// "projects/locations". ParentStyle collapses the uncommon shapes into
	// "other", so this is what tells a human triaging one what it actually is.
	ParentPath string
	// ParentStyle classifies ParentPath into a shape the templates can render.
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
	md.Collection, md.ParentPath = splitPattern(md.Pattern)
	if md.Collection == "" && md.Plural != "" {
		md.Collection = md.Plural
	}
	md.ParentStyle = classifyParent(md.ParentPath)
	return md
}

// splitPattern separates a resource name pattern into the resource's own
// collection segment and its parent's collection segments, joined back into a
// path.
//
// "projects/{project}/locations/{location}/foos/{foo}"
//
//	-> collection "foos", parent "projects/locations"
func splitPattern(pattern string) (collection string, parentPath string) {
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
		return "", ""
	}

	// A pattern may end in a literal with no trailing placeholder (e.g.
	// "projects/{project}/locations"); in that case there is no distinct
	// collection for the resource itself.
	last := segs[len(segs)-1]
	if !strings.HasPrefix(last, "{") {
		return "", strings.Join(literals, "/")
	}
	return literals[len(literals)-1], strings.Join(literals[:len(literals)-1], "/")
}

// ParentPair returns the collection segment and placeholder naming a resource's
// direct parent, or two empty strings when the pattern declares no parent.
//
// Per AIP-122 a parent is the whole path down to that segment, so the
// collection is what identifies the parent's type; the placeholder only hints
// at what to call the field, and hints badly, because placeholders run their
// words together ("{keyring}", "{datastore}") where the collection beside them
// keeps the API's own casing ("keyRings", "dataStores").
//
// Patterns end in one of three ways. Most end on the resource's own
// collection/{id}, where the parent is the pair before it. 274 of the 3160 in
// googleapis end in a literal naming a singleton, such as
// "accounts/{account}/programs/{program}/checkoutSettings", where the last
// pair is already the parent. A pattern ending in two placeholders in a row
// names no collection between them, so there is no parent to read.
func ParentPair(pattern string) (collection string, placeholder string) {
	segs := strings.Split(pattern, "/")

	type pair struct{ collection, placeholder string }
	var pairs []pair
	for i := 0; i+1 < len(segs); i++ {
		if !strings.HasPrefix(segs[i], "{") && strings.HasPrefix(segs[i+1], "{") {
			pairs = append(pairs, pair{segs[i], strings.Trim(segs[i+1], "{}")})
		}
	}
	// One pair is the resource's own, leaving nothing for a parent.
	if len(pairs) < 2 {
		return "", ""
	}

	if strings.HasPrefix(segs[len(segs)-1], "{") {
		if strings.HasPrefix(segs[len(segs)-2], "{") {
			return "", ""
		}
		parent := pairs[len(pairs)-2]
		return parent.collection, parent.placeholder
	}
	last := pairs[len(pairs)-1]
	return last.collection, last.placeholder
}

func classifyParent(parentPath string) ParentStyle {
	switch parentPath {
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
