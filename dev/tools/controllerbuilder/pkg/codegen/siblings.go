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
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// SiblingGuessMarker is written above a field whose name matches a resource the
// same service declares. It is a comment marker, so controller-gen strips it
// before the CRD is published and it cannot affect the schema.
const SiblingGuessMarker = "+kcc:guess=possible-reference target="

// SiblingResource reports the resource this service declares whose name matches
// the field's, if any.
//
// siblings maps a lowercased Kind suffix to the Kind: DiscoveryEngineDataStore
// is keyed "datastore", so a field called dataStore matches it. The rule keeps
// no list of known names, so it works on a service nobody has looked at.
// refs.NameRules cannot do that, because it only recognises spellings someone
// has already written down.
//
// The leaf match is exact, after singularising. We measured a looser endswith
// match at 68% precision, against 75% for this one, and it wrongly matched
// spec.pipelineJob and localSsds[].interface. Two fields match a sibling Kind
// but stay plain upstream, both in DataLabeling: annotationSpecSet and
// instruction.
func SiblingResource(field protoreflect.FieldDescriptor, siblings map[string]string) (string, bool) {
	if len(siblings) == 0 || field.Kind() != protoreflect.StringKind {
		return "", false
	}
	return SiblingResourceByName(GetJSONForKRM(field), siblings)
}

// SiblingResourceByName matches a name the generator synthesised, where
// SiblingResource takes a proto field. Parent segments arrive this way:
// DiscoveryEngineDataStoreTargetSite's spec.dataStore comes from the resource
// pattern, not from any field of the message.
func SiblingResourceByName(name string, siblings map[string]string) (string, bool) {
	leaf := strings.ToLower(name)
	if target, ok := siblings[leaf]; ok {
		return target, true
	}
	// A repeated field is named for what it holds: ComputeNetworkAttachment's
	// subnetworks are each a ComputeSubnetwork. Six of the fifteen matches in the
	// corpus are plural. Without this the rule finds half as many.
	if s := singular(leaf); s != leaf {
		if target, ok := siblings[s]; ok {
			return target, true
		}
	}
	return "", false
}

// singular strips a regular English plural and gives up on anything else. It
// stays narrow on purpose. A match here puts a question in front of a reviewer,
// and no GCP resource name uses an irregular plural.
func singular(s string) string {
	switch {
	case strings.HasSuffix(s, "ies") && len(s) > 4:
		// policies -> policy
		return s[:len(s)-3] + "y"
	case strings.HasSuffix(s, "sses"), strings.HasSuffix(s, "shes"),
		strings.HasSuffix(s, "ches"), strings.HasSuffix(s, "xes"),
		strings.HasSuffix(s, "zes"):
		// addresses -> address
		return s[:len(s)-2]
	case strings.HasSuffix(s, "ss"), strings.HasSuffix(s, "us"),
		strings.HasSuffix(s, "is"):
		// access, status, analysis: not plurals at all.
		return s
	case strings.HasSuffix(s, "s") && len(s) > 2:
		// subnetworks -> subnetwork
		return s[:len(s)-1]
	}
	return s
}

// SiblingGuess is one field flagged by the sibling rule while writing a nested
// message, for the judgement queue.
type SiblingGuess struct {
	// Message is the proto message the field belongs to, e.g.
	// google.cloud.discoveryengine.v1.Control.
	Message string
	// Field is the KRM json name.
	Field string
	// Target is the sibling Kind the name matched.
	Target string
}

// scanSiblingGuesses recovers the sibling markers WriteMessage left in a
// rendered body. Scanned out of the output rather than returned from
// WriteField, which writes to an io.Writer and has nowhere to accumulate;
// scanUnsupported recovers the "// TODO:" markers the same way.
func scanSiblingGuesses(msgName, body string) []SiblingGuess {
	var out []SiblingGuess
	pending := ""
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			if _, target, found := strings.Cut(trimmed, SiblingGuessMarker); found {
				pending = target
			}
			continue
		}
		if pending == "" {
			continue
		}
		if _, after, found := strings.Cut(line, "`json:\""); found {
			name, _, _ := strings.Cut(after, ",")
			out = append(out, SiblingGuess{Message: msgName, Field: name, Target: pending})
			pending = ""
		}
	}
	return out
}
