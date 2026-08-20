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
	"sort"
	"strings"

	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/codegen"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/apimachinery/pkg/util/sets"
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
	// ObservedStateFields is the same thing for the resource-level
	// <Kind>ObservedState struct. Empty when the proto marks nothing OUTPUT_ONLY,
	// which leaves the scaffolded struct empty as before.
	ObservedStateFields string
	// ExtraImports are import paths the rendered fields need beyond the three the
	// template always writes.
	ExtraImports []string
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
// ObservedState is filled separately, by PrepopulateObservedState, because it
// needs data only the type generator has.
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

		codegen.WriteField(&buf, field, msg, emitted, false, opts)
		emitted++

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

// PrepopulateObservedState renders the body of the resource-level
// <Kind>ObservedState struct, and reports any import the rendered fields need.
//
// This is mechanical, not a judgement call. Measured on the pilot: for
// NetworkSecurityURLList and TranscoderJob the proto alone gives the complete and
// correct answer, and producing it by hand consisted of copying what the generator
// had already worked out.
//
// details comes from the type generator's identifyOutputs, so the transitive rule
// -- a field is output-only if reached through an OUTPUT_ONLY parent -- is applied
// once, in one place.
func PrepopulateObservedState(details *codegen.OutputMessageDetails, observedStateMessages sets.String) (fields string, extraImports []string) {
	if details == nil {
		return "", nil
	}

	var buf bytes.Buffer
	// identityFields is skipped here for the same reason as in the Spec: "name" is
	// the resource's own resource name, which KCC carries in status.externalRef
	// rather than as an observed field, even where the proto marks it OUTPUT_ONLY.
	codegen.WriteObservedStateFields(&buf, details, observedStateMessages, identityFields)
	fields = buf.String()

	return fields, ExtraImportsFor(fields)
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

// DetectOutputOnlyInComments finds spec fields whose proto comment opens with
// "Output only." but whose field_behavior does not say OUTPUT_ONLY.
//
// It reports rather than acts. Applying the inference directly would move 90
// fields across 13 services, 29 of them in v1beta1, where relocating a field
// from spec to status breaks a schema people already depend on. The reported
// fields are moved by hand, in <kind>_types.go, once someone has agreed.
//
// The signal itself is trustworthy: across 3780 fields in hand-written Spec
// structs, not one carries "Output only." in its comment, so there are no
// measured false positives. What is missing is the review, not the accuracy.
func DetectOutputOnlyInComments(msg protoreflect.MessageDescriptor) []OutputOnlyCandidate {
	if msg == nil {
		return nil
	}
	var out []OutputOnlyCandidate
	for i := 0; i < msg.Fields().Len(); i++ {
		field := msg.Fields().Get(i)
		if codegen.IsFieldBehavior(field, annotations.FieldBehavior_OUTPUT_ONLY) {
			continue
		}
		if identityFields[string(field.Name())] {
			continue
		}
		comment := strings.TrimSpace(msg.ParentFile().SourceLocations().ByDescriptor(field).LeadingComments)
		if !strings.HasPrefix(comment, "Output only.") {
			continue
		}
		out = append(out, OutputOnlyCandidate{
			FieldPath: ".spec." + codegen.GetJSONForKRM(field),
			Comment:   strings.Join(strings.Fields(comment), " "),
		})
	}
	return out
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

// ExtraImportsFor reports the imports a rendered field body needs beyond the
// three the types template always writes.
//
// A handful of proto types map to Go types from other packages -- google.rpc.Status
// to common.Status, google.protobuf.Struct to apiextensionsv1.JSON. The template
// imports none of them, so anything the rendered Spec or ObservedState references
// has to be declared or the scaffolded file does not compile. Both bodies are
// scanned, because either can contain such a field: securitycentermanagement puts
// an apiextensionsv1.JSON in the Spec, transcoder a common.Status in the
// ObservedState.
func ExtraImportsFor(bodies ...string) []string {
	var out []string
	for qualifier, importPath := range codegen.QualifierImports {
		for _, body := range bodies {
			if strings.Contains(body, qualifier+".") {
				// Emit the alias, always. The path's last segment is often not the
				// qualifier the field uses -- apiextensions-apiserver/.../v1 provides
				// package "v1", while the field says apiextensionsv1.JSON -- and
				// goimports then removes the import as unused rather than fixing it.
				out = append(out, fmt.Sprintf("%s %q", qualifier, importPath))
				break
			}
		}
	}
	sort.Strings(out)
	return out
}
