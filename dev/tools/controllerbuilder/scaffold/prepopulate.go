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
	"bytes"
	"fmt"
	"strings"

	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/codegen"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// identityFields are proto fields that the KRM object expresses through its own
// identity rather than as spec fields. "name" is the resource's own resource
// name, which KCC models as metadata.name / spec.resourceID plus the parent refs.
var identityFields = map[string]bool{
	"name": true,
}

// JudgementItem is one thing the generator could not decide, destined for the
// service's needs_judgement_call.txt.
type JudgementItem struct {
	// FieldPath is the KRM path, e.g. ".spec.forwardingRules".
	FieldPath string
	// Reason is a short slug explaining what needs deciding.
	Reason string
	// Detail is optional extra context, e.g. the referenced proto type.
	Detail string
}

// PrepopulateResult is the rendered spec body plus what still needs a human.
type PrepopulateResult struct {
	// SpecFields is Go source for the body of the Spec struct: one field per
	// proto field, already indented, ready to paste between the braces.
	SpecFields string
	// Judgement lists fields the generator emitted mechanically but cannot
	// vouch for.
	Judgement []JudgementItem
}

// PrepopulateSpec renders the top-level Spec fields for a resource message.
//
// Fields it cannot decide about are still emitted, as their raw proto-derived
// type, with the open question recorded separately. Emitting a field with a
// listed open question beats omitting it, because an omission is invisible to
// every other check: a field absent from the CRD cannot be reported as missing
// from it.
//
// ObservedState is not pre-populated. Output fields reached through nested
// messages need the generated <Proto>ObservedState variants rather than the
// plain structs, and picking the right one per field is its own problem.
func PrepopulateSpec(msg protoreflect.MessageDescriptor, opts codegen.WriteOptions) (*PrepopulateResult, error) {
	if msg == nil {
		return nil, fmt.Errorf("no message descriptor")
	}

	out := &PrepopulateResult{}
	var buf bytes.Buffer

	emitted := 0
	for i := 0; i < msg.Fields().Len(); i++ {
		field := msg.Fields().Get(i)

		// Output-only fields belong in ObservedState, which the type generator
		// writes separately.
		if codegen.IsFieldBehavior(field, annotations.FieldBehavior_OUTPUT_ONLY) {
			continue
		}
		if identityFields[string(field.Name())] {
			continue
		}

		// We render each field on its own so we can inspect its output before
		// appending it. When the generator cannot type a field it writes a
		// "// TODO:" comment and moves on, and the field never reaches the CRD.
		// That is a silent drop unless somebody records it.
		var field_ bytes.Buffer
		codegen.WriteField(&field_, field, msg, emitted, false, opts)
		buf.Write(field_.Bytes())
		emitted++

		if reason, ok := unsupportedFieldReason(field_.String()); ok {
			out.Judgement = append(out.Judgement, JudgementItem{
				FieldPath: ".spec." + codegen.GetJSONForKRM(field),
				Reason:    "unsupported-field-type",
				Detail:    reason,
			})
		}
		if item, ok := judgementFor(field); ok {
			out.Judgement = append(out.Judgement, item)
		}
	}

	out.SpecFields = buf.String()

	// Always mark the resource untriaged, whatever we did or did not detect.
	//
	// The resource-level entry is what actually drives suppression, so it must not
	// depend on the detector finding anything. Measured on the pilot resource:
	// LbTrafficExtension carries no google.api.resource_reference on any field,
	// including forwarding_rules, which is precisely the field that has to become
	// a ref. A queue built only from annotations would have been empty, so no file
	// would have been written, nothing would have been suppressed, and the resource
	// would have gone straight into the missingrefs ratchet and failed. Preventing
	// exactly that is why the queue exists.
	out.Judgement = append([]JudgementItem{{
		Reason: "untriaged-bulk-generation",
		Detail: "spec was generated mechanically; confirm refs, omissions and KRM names",
	}}, out.Judgement...)

	return out, nil
}

// judgementFor reports whether a field needs a human decision that the generator
// cannot make.
//
// Only google.api.resource_reference is consulted. Where it is present it is
// authoritative, and it names the exact target type.
//
// Guessing from field names was built and measured before being rejected: across
// the CRDs it flagged 2164 fields as probable references, against 78 for the
// description-based heuristics, because a name like "network" is annotated in one
// service and then recurs on hundreds of unrelated fields elsewhere. So this
// deliberately under-reports rather than filling the queue with noise. Fields
// that look like references but carry no annotation are still caught later by
// TestMissingRefs, once the resource graduates from the queue.
func judgementFor(field protoreflect.FieldDescriptor) (JudgementItem, bool) {
	if field.Options() == nil {
		return JudgementItem{}, false
	}
	v := proto.GetExtension(field.Options(), annotations.E_ResourceReference)
	rr, _ := v.(*annotations.ResourceReference)
	if rr == nil {
		return JudgementItem{}, false
	}
	target := rr.GetType()
	if target == "" {
		target = rr.GetChildType()
	}
	if target == "" {
		return JudgementItem{}, false
	}

	return JudgementItem{
		FieldPath: ".spec." + codegen.GetJSONForKRM(field),
		Reason:    "possible-reference",
		Detail:    "target=" + target,
	}, true
}

// FormatJudgementEntries renders queue lines for one resource, in the format
// apis/<service>/needs_judgement_call.txt expects.
func FormatJudgementEntries(kind, group string, items []JudgementItem) string {
	var sb strings.Builder
	for _, it := range items {
		// A resource-level item has no field path.
		subject := "resource"
		if it.FieldPath != "" {
			subject = fmt.Sprintf("field %q", it.FieldPath)
		}
		sb.WriteString(fmt.Sprintf("kind=%s group=%s: %s reason=%s",
			kind, group, subject, it.Reason))
		if it.Detail != "" {
			sb.WriteString(" (" + it.Detail + ")")
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// OutputOnlyCandidate is a field the proto documents as output-only in prose
// while carrying no google.api.field_behavior annotation to say so.
type OutputOnlyCandidate struct {
	// FieldPath is the KRM path the field was emitted at, e.g. ".spec.createTime".
	FieldPath string
	// Comment is the proto's leading comment, so a reviewer can decide without
	// opening the proto.
	Comment string
}

// outputOnlyPrefixes are the ways a proto says "GCP sets this" in prose rather
// than in a google.api.field_behavior annotation.
//
// There are two spellings, because two families of API write it differently.
// Most protos open the comment "Output only."; Compute opens it "[Output
// Only]". For a long time only the first was recognised, so every Compute
// resource lost the signal entirely: 1,605 fields in compute.proto alone, and
// all ten of ComputeInterconnect's misplaced observed-state fields
// (googleIPAddress, circuitInfos, expectedOutages and the rest).
//
// Both are matched as a prefix rather than anywhere in the comment. That is the
// convention in practice: of the Compute fields carrying the marker, 1,600 open
// with it and 5 mention it mid-sentence. Those 5 are left, because an anchored
// test is the one whose false-positive rate was measured.
var outputOnlyPrefixes = []string{"Output only.", "[Output Only]"}

// DetectOutputOnlyInComments finds spec fields whose proto comment says the
// field is output-only while its field_behavior does not.
//
// It reports rather than acts. If the generator acted on the inference, it
// would move 90 fields across 13 services, 29 of them in v1beta1, and a field
// moved from spec to status breaks a schema people already depend on. Someone
// moves the reported fields by hand, in <kind>_types.go, once it is agreed.
//
// The signal itself is trustworthy: across 4,673 fields in hand-written Spec
// structs in the baseline tree, not one carries either spelling in its comment,
// so there are no measured false positives for either. What is missing is the
// review, not the accuracy.
func DetectOutputOnlyInComments(msg protoreflect.MessageDescriptor) []OutputOnlyCandidate {
	if msg == nil {
		return nil
	}
	var out []OutputOnlyCandidate
	for i := 0; i < msg.Fields().Len(); i++ {
		field := msg.Fields().Get(i)
		if codegen.IsFieldBehavior(field, annotations.FieldBehavior_OUTPUT_ONLY) || identityFields[string(field.Name())] {
			continue
		}
		comment, ok := outputOnlyComment(field)
		if !ok {
			continue
		}
		out = append(out, OutputOnlyCandidate{
			FieldPath: ".spec." + codegen.GetJSONForKRM(field),
			Comment:   comment,
		})
	}
	return out
}

// outputOnlyComment returns a field's leading comment, collapsed onto one line,
// when the comment opens with one of outputOnlyPrefixes.
func outputOnlyComment(field protoreflect.FieldDescriptor) (string, bool) {
	loc := field.ParentFile().SourceLocations().ByDescriptor(field)
	comment := strings.TrimSpace(loc.LeadingComments)
	for _, prefix := range outputOnlyPrefixes {
		if strings.HasPrefix(comment, prefix) {
			return strings.Join(strings.Fields(comment), " "), true
		}
	}
	return "", false
}

// FormatOutputOnlyCandidates renders detector output for the report file.
func FormatOutputOnlyCandidates(kind, group string, items []OutputOnlyCandidate) string {
	var sb strings.Builder
	for _, it := range items {
		sb.WriteString(fmt.Sprintf("kind=%s group=%s: field %q comment=%q\n",
			kind, group, it.FieldPath, it.Comment))
	}
	return sb.String()
}

// unsupportedFieldReason reports the generator's own explanation when it could
// not produce a Go type for a field.
//
// WriteField emits "// TODO: <err>" in place of the field and carries on, so
// the field is absent from the CRD with nothing but a comment in generated
// source to say why. The 239-resource run had 15 such markers in scaffolded
// type files and 37 more in types.generated.go. Between them they lost 124
// CRD field paths, and neither the judgement queue nor any report listed
// them.
func unsupportedFieldReason(rendered string) (string, bool) {
	for _, line := range strings.Split(rendered, "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "// TODO: "); ok {
			// WriteField prefixes the field name; FieldPath already carries it.
			if _, reason, found := strings.Cut(after, ": "); found {
				return reason, true
			}
			return after, true
		}
	}
	return "", false
}
