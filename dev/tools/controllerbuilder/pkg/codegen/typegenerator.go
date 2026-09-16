// Copyright 2024 Google LLC
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
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	codegenannotations "github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/annotations"
	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/protoapi"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
)

type TypeGenerator struct {
	generatorBase
	api                   *protoapi.Proto
	goPackage             string
	visitedMessages       []protoreflect.MessageDescriptor
	outputMessages        []*OutputMessageDetails
	observedStateMessages sets.String
	// unsupportedFields are fields the generator could not type, collected as
	// they are written so the judgement queue can report them.
	unsupportedFields []UnsupportedField

	// siblingGuesses are fields of a nested message that the sibling rule
	// flagged. PrepopulateSpec records the resource's own fields; this generator
	// writes the nested messages, so their markers are read back out of the
	// rendered body.
	siblingGuesses []SiblingGuess

	generatedFileAnnotation *codegenannotations.FileAnnotation
	includeSkippedOutput    bool
	writeOptions            WriteOptions

	// handWritten is what the target package already defines by hand, scanned once when we visit
	// the first message. See handWrittenTypes for why the decision to split a message and the
	// decision to write it both have to read from here.
	handWritten *handWrittenTypes

	// reservedTypeNames are Go type names the scaffolder will declare later in
	// this run, so the generator must not emit them into types.generated.go.
	//
	// findTypeDeclaration already skips a name the package declares by hand, but
	// it cannot see a file that does not exist yet: a wipe-based regeneration
	// removes <kind>_types.go, the generator runs, and the scaffolder writes the
	// Kind's struct afterwards. Where the Kind's name equals the Go name of its
	// own proto message, such as billing's BillingAccount or apihub's
	// APIHubInstance, that is a redeclaration.
	reservedTypeNames map[string]bool

	// rootMessageFQN is the full name of the message VisitProto is visiting.
	// isServerSet uses it to tell that message from the nested ones below it.
	rootMessageFQN string
}

type OutputMessageDetails struct {
	Message      protoreflect.MessageDescriptor
	OutputFields []protoreflect.FieldDescriptor
}

// WriteOptions turns on KRM markers derived from proto annotations.
//
// These are off by default and opted into one service at a time, because switching one on can
// change the CRD of a resource people already use. A nested message type is generated once and
// shared between the spec and the observed state, so a marker taken from one field's annotation
// appears everywhere that type is used, status included. Status is the case to worry about: the
// API server validates it like any other part of the object, so a tightened status schema can
// start rejecting a status that KCC itself wrote.
//
// New resources turn these on from the start, which gets the shape right before anyone depends
// on it.
type WriteOptions struct {
	// EmitRequired writes "// +required" for fields the proto marks REQUIRED.
	EmitRequired bool
	// EmitPluralAcronyms cases a plural acronym as KRM conventions want, so
	// related_uris becomes RelatedURIs rather than RelatedUris. See AcronymCasing
	// for why this is opt-in.
	EmitPluralAcronyms bool
	// EmitMessageMaps generates map<string, Message> fields as a map of the
	// value's Go type. Without it they are left out, with a "// TODO:" marker
	// in their place.
	EmitMessageMaps bool
	// PlaceServerSetFields puts a small allowlist of server-computed fields in
	// ObservedState when no field of the message carries field_behavior. See
	// IsServerSetField for the allowlist and the guard.
	PlaceServerSetFields bool
	// Siblings maps a lowercased Kind suffix to a Kind this service declares, so
	// a string field naming one can be marked as a probable reference. See
	// SiblingResource.
	//
	// Unlike the flags above, this one cannot change a CRD: it only adds a
	// +kcc:guess comment, which controller-gen strips before publishing. A shared
	// nested message is no trouble either, since the map holds one service's
	// Kinds and a nested message is shared only within a service.
	Siblings map[string]string
}

func NewTypeGenerator(goPackage string, outputBaseDir string, api *protoapi.Proto) *TypeGenerator {
	g := &TypeGenerator{
		goPackage:             goPackage,
		api:                   api,
		observedStateMessages: sets.NewString(),
	}
	g.generatorBase.init(outputBaseDir)
	return g
}

// WithGeneratedFileAnnotation sets the generated file annotation
func (g *TypeGenerator) WithGeneratedFileAnnotation(generatedFileAnnotation *codegenannotations.FileAnnotation) *TypeGenerator {
	g.generatedFileAnnotation = generatedFileAnnotation
	return g
}

// WithIncludeSkippedOutput sets whether to output skipped types as commented-out code
func (g *TypeGenerator) WithIncludeSkippedOutput(includeSkippedOutput bool) *TypeGenerator {
	g.includeSkippedOutput = includeSkippedOutput
	return g
}

// WithReservedTypeNames names the Go types the scaffolder will write later, so
// the generator leaves them alone.
func (g *TypeGenerator) WithReservedTypeNames(names ...string) *TypeGenerator {
	if g.reservedTypeNames == nil {
		g.reservedTypeNames = map[string]bool{}
	}
	for _, n := range names {
		g.reservedTypeNames[n] = true
		g.reservedTypeNames[n+"ObservedState"] = true
	}
	return g
}

// WithWriteOptions selects which proto-derived markers to emit.
func (g *TypeGenerator) WithWriteOptions(opts WriteOptions) *TypeGenerator {
	g.writeOptions = opts
	return g
}

func (g *TypeGenerator) VisitProto(resourceProtoFullName string) error {

	descriptor, err := g.api.Files().FindDescriptorByName(protoreflect.FullName(resourceProtoFullName))
	if err != nil {
		return fmt.Errorf("failed to find the proto message %s: %w", resourceProtoFullName, err)
	}
	messageDescriptor, ok := descriptor.(protoreflect.MessageDescriptor)
	if !ok {
		return fmt.Errorf("unexpected descriptor type: %T", descriptor)
	}

	if err := g.visitMessage(messageDescriptor); err != nil {
		return err
	}

	return nil
}

// typesOutputDir is the package directory types.generated.go is written to.
func (g *TypeGenerator) typesOutputDir() string {
	return g.getOutputFile(generatedFileKey{
		GoPackage: g.goPackage,
		FileName:  "types.generated.go",
	}).OutputDir()
}

func (g *TypeGenerator) visitMessage(message protoreflect.MessageDescriptor) error {
	//klog.Infof("found message %q", messageDescriptor.FullName())

	g.rootMessageFQN = string(message.FullName())
	g.visitedMessages = append(g.visitedMessages, message)

	msgs, err := FindDependenciesForMessage(message, nil) // TODO: explicitly set ignored fields when generating Go types
	if err != nil {
		return err
	}
	g.visitedMessages = append(g.visitedMessages, msgs...)

	outputDeps := make(map[string]*OutputMessageDetails)
	g.identifyOutputs(message, make(map[string]string), outputDeps, false)

	if g.handWritten == nil {
		handWritten, err := g.scanHandWrittenTypes(g.typesOutputDir())
		if err != nil {
			return err
		}
		g.handWritten = handWritten
	}

	needsObservedStateCache := make(map[string]bool)
	for fqn, details := range outputDeps {
		if g.needsObservedState(details.Message, needsObservedStateCache) {
			g.outputMessages = append(g.outputMessages, details)
			g.observedStateMessages.Insert(fqn)
		}
	}

	return nil
}

// ObservedStateMessages returns the proto messages that have their own
// ObservedState struct. A field of one of these message types must use
// <Proto>ObservedState rather than the plain type.
func (g *TypeGenerator) ObservedStateMessages() sets.String {
	return g.observedStateMessages
}

// OutputFieldsFor returns the output-only fields that identifyOutputs found for
// one message during the visit, and false when the message has none. The
// scaffolder fills the resource-level ObservedState struct from this rather
// than walking the proto again, so the rule that a field reached through an
// OUTPUT_ONLY parent is also output-only lives in one place.
func (g *TypeGenerator) OutputFieldsFor(fqn string) (*OutputMessageDetails, bool) {
	for _, details := range g.outputMessages {
		if string(details.Message.FullName()) == fqn {
			return details, true
		}
	}
	return nil, false
}

// isServerSet reports what IsServerSetField says for a field of the resource's
// own message, and false for a field of any other message. identifyOutputs
// recurses into nested messages, where a field called "id" or "kind" is often
// the user's to set.
func (g *TypeGenerator) isServerSet(field protoreflect.FieldDescriptor, msg protoreflect.MessageDescriptor) bool {
	if string(msg.FullName()) != g.rootMessageFQN {
		return false
	}
	return IsServerSetField(field, msg, g.writeOptions)
}

// needsObservedState determines if a message requires a separate ObservedState struct.
// If the regular Go struct and the ObservedState version are identical, we fall back
// to using the regular Go struct to reduce redundancy.
func (g *TypeGenerator) needsObservedState(msg protoreflect.MessageDescriptor, seen map[string]bool) bool {
	fqn := string(msg.FullName())
	if val, ok := seen[fqn]; ok {
		return val
	}
	seen[fqn] = false // Assume false for recursion

	// The loop below decides whether this message needs an ObservedState struct of its own instead
	// of sharing one struct between the spec and the observed state. An OUTPUT_ONLY field has
	// always forced that split, and a REQUIRED field has to force it too: WriteMessage marks such
	// a field "// +required" and WriteObservedStateMessage deliberately does not, so sharing the
	// struct would put required: into the status schema. The API server validates status against
	// what GCP returned, not against anything the user wrote.
	//
	// We split only when the generator owns the message outright. If a hand-written type claims
	// it, the generator writes no struct for it, so no marker is emitted and nothing can reach
	// status; splitting would only name a struct that collides with the hand-written one or that
	// nothing defines. We also split only when the flag is on, so services that have not opted in
	// keep byte-identical output.
	splitOnRequired := g.writeOptions.EmitRequired &&
		g.handWritten.ownedByGenerator(fqn, GoNameForProtoMessage(msg), goNameForOutputProtoMessage(msg))

	for i := 0; i < msg.Fields().Len(); i++ {
		f := msg.Fields().Get(i)
		if IsFieldBehavior(f, annotations.FieldBehavior_OUTPUT_ONLY) {
			seen[fqn] = true
			return true
		}
		// A server-set field needs the ObservedState struct as much as an
		// OUTPUT_ONLY field does. Without this, identifyOutputs collects it for a
		// struct nothing writes.
		if g.isServerSet(f, msg) {
			seen[fqn] = true
			return true
		}
		if splitOnRequired && IsFieldBehavior(f, annotations.FieldBehavior_REQUIRED) {
			seen[fqn] = true
			return true
		}
		if f.Kind() == protoreflect.MessageKind && !f.IsMap() {
			if _, ok := protoMessagesNotMappedToGoStruct[string(f.Message().FullName())]; ok {
				continue
			}
			if g.needsObservedState(f.Message(), seen) {
				seen[fqn] = true
				return true
			}
		}
	}
	return false
}

// identifyOutputs recursively identifies all messages in the proto tree that contain any output-only content.
// A message contains output-only content if:
// 1. It has a field explicitly marked as OUTPUT_ONLY.
// 2. It has a nested message field that itself contains output-only fields.
// 3. It is reached via a parent field marked as OUTPUT_ONLY.
func (g *TypeGenerator) identifyOutputs(msg protoreflect.MessageDescriptor, seen map[string]string, outputDeps map[string]*OutputMessageDetails, forceAll bool) bool {
	fqn := string(msg.FullName())
	seenKey := fqn
	if forceAll {
		seenKey += "|forced"
	}
	if state, ok := seen[seenKey]; ok {
		return state == "has_outputs"
	}
	seen[seenKey] = "visiting"

	details := &OutputMessageDetails{Message: msg}
	hasOutputs := false

	for i := 0; i < msg.Fields().Len(); i++ {
		f := msg.Fields().Get(i)
		isOut := IsFieldBehavior(f, annotations.FieldBehavior_OUTPUT_ONLY) || g.isServerSet(f, msg)

		if isPrimitive(f) {
			// Primitive fields are only included if explicitly marked OUTPUT_ONLY.
			if isOut || forceAll {
				details.OutputFields = append(details.OutputFields, f)
				hasOutputs = true
			}
		} else {
			// Message fields are included if they are OUTPUT_ONLY OR if the target message has outputs.
			childHasOutputs := g.identifyOutputs(f.Message(), seen, outputDeps, forceAll || isOut)
			if isOut || childHasOutputs || forceAll {
				details.OutputFields = append(details.OutputFields, f)
				hasOutputs = true
			}
		}
	}

	if hasOutputs {
		outputDeps[fqn] = details
		seen[seenKey] = "has_outputs"
		return true
	}
	seen[seenKey] = "no_outputs"
	return false
}

func writeCopyright(w io.Writer, year int) {
	s := `// Copyright {{.Year}} Google LLC
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

`
	s = strings.ReplaceAll(s, "{{.Year}}", strconv.Itoa(year))
	if _, err := w.Write([]byte(s)); err != nil {
		klog.Fatalf("writing copyright: %v", err)
	}
}

func (g *TypeGenerator) WriteVisitedMessages() error {
	for _, msg := range deduplicateAndSort(g.visitedMessages) {
		if msg.IsMapEntry() {
			continue
		}

		k := generatedFileKey{
			GoPackage: g.goPackage,
			FileName:  "types.generated.go",
		}
		out := g.getOutputFile(k)

		for i := 0; i < msg.Fields().Len(); i++ {
			field := msg.Fields().Get(i)
			if field.Message() != nil {
				name := field.Message().FullName()
				if name == "google.cloud.connectors.v1.Secret" {
					out.addImport("secretmanagerv1beta1", "github.com/GoogleCloudPlatform/k8s-config-connector/apis/secretmanager/v1beta1")
				}
				if name == "google.rpc.Status" {
					out.addImport("common", "github.com/GoogleCloudPlatform/k8s-config-connector/apis/common")
					break
				}
			}
		}

		out.goPackage = lastGoComponent(g.goPackage)

		out.fileAnnotation = g.generatedFileAnnotation

		goTypeName := GoNameForProtoMessage(msg)
		if g.reservedTypeNames[goTypeName] {
			klog.V(1).Infof("go type %q is a Kind the scaffolder will declare, won't generate", goTypeName)
			continue
		}
		skipGenerated := true
		goType, err := g.findTypeDeclaration(goTypeName, out.OutputDir(), skipGenerated)
		if err != nil {
			return fmt.Errorf("looking up go type: %w", err)
		}
		if goType != nil {
			klog.V(1).Infof("found existing non-generated go type %q, won't generate", goTypeName)
			if g.includeSkippedOutput {
				WriteMessageAsComment(&out.body, msg, fmt.Sprintf("found existing non-generated go type %q, skipping", goTypeName), g.writeOptions)
			}
			continue
		}

		goType, err = g.findTypeDeclarationWithProtoTag(string(msg.FullName()), out.OutputDir(), skipGenerated)
		if err != nil {
			return fmt.Errorf("looking up go type by proto tag: %w", err)
		}
		if goType != nil {
			klog.V(1).Infof("found existing non-generated go type with proto tag %q, won't generate", msg.FullName())
			if g.includeSkippedOutput {
				WriteMessageAsComment(&out.body, msg, fmt.Sprintf("found existing non-generated go type with proto tag %q, skipping", msg.FullName()), g.writeOptions)
			}
			continue
		}

		// The message renders into its own buffer so we can collect the markers
		// WriteField leaves behind: one for a field it could not type, and one
		// for a field the sibling rule flagged. WriteField takes an io.Writer and
		// has nowhere to collect them, so they are read back out of the finished
		// text. Otherwise an untypeable field is absent from the CRD and the only
		// trace is a "// TODO:" comment in the generated source.
		var rendered bytes.Buffer
		WriteMessage(&rendered, msg, g.writeOptions)
		g.unsupportedFields = append(g.unsupportedFields, scanUnsupported(string(msg.FullName()), rendered.String())...)
		g.siblingGuesses = append(g.siblingGuesses, scanSiblingGuesses(string(msg.FullName()), rendered.String())...)
		out.body.Write(rendered.Bytes())
	}
	return errors.Join(g.errors...)
}

func (g *TypeGenerator) WriteOutputMessages() error {
	for _, msgDetails := range deduplicateAndSortOutputMessages(g.outputMessages) {
		msg := msgDetails.Message
		if msg.IsMapEntry() {
			continue
		}

		k := generatedFileKey{
			GoPackage: g.goPackage,
			FileName:  "types.generated.go",
		}
		out := g.getOutputFile(k)

		for _, field := range msgDetails.OutputFields {
			if field.Message() != nil {
				name := field.Message().FullName()
				if name == "google.cloud.connectors.v1.Secret" {
					out.addImport("secretmanagerv1beta1", "github.com/GoogleCloudPlatform/k8s-config-connector/apis/secretmanager/v1beta1")
				}
				if name == "google.rpc.Status" {
					out.addImport("common", "github.com/GoogleCloudPlatform/k8s-config-connector/apis/common")
					break
				}
			}
		}

		out.goPackage = lastGoComponent(g.goPackage)

		out.fileAnnotation = g.generatedFileAnnotation

		goTypeName := goNameForOutputProtoMessage(msg)
		if g.reservedTypeNames[goTypeName] {
			klog.V(1).Infof("go type %q is a Kind the scaffolder will declare, won't generate", goTypeName)
			continue
		}
		skipGenerated := true
		goType, err := g.findTypeDeclaration(goTypeName, out.OutputDir(), skipGenerated)
		if err != nil {
			return fmt.Errorf("looking up go type: %w", err)
		}
		if goType != nil {
			klog.V(1).Infof("found existing non-generated go type %q, won't generate", goTypeName)
			if g.includeSkippedOutput {
				WriteObservedStateMessageAsComment(&out.body, msgDetails, fmt.Sprintf("found existing non-generated go type %q, skipping", goTypeName), g.observedStateMessages, g.writeOptions)
			}
			continue
		}

		goType, err = g.findTypeDeclarationWithProtoTag(string(msg.FullName()), out.OutputDir(), skipGenerated)
		if err != nil {
			return fmt.Errorf("looking up go type by proto tag: %w", err)
		}
		if goType != nil {
			klog.V(1).Infof("found existing non-generated go type with proto tag %q, won't generate", msg.FullName())
			if g.includeSkippedOutput {
				WriteObservedStateMessageAsComment(&out.body, msgDetails, fmt.Sprintf("found existing non-generated go type with proto tag %q, skipping", msg.FullName()), g.observedStateMessages, g.writeOptions)
			}
			continue
		}

		WriteObservedStateMessage(&out.body, msgDetails, g.observedStateMessages, g.writeOptions)
	}
	return errors.Join(g.errors...)
}

func WriteMessageAsComment(out io.Writer, msg protoreflect.MessageDescriptor, reason string, opts WriteOptions) {
	var b bytes.Buffer
	// Clear the map for this block: it dumps what the generator would have written
	// for a type the package already declares by hand, so nothing in it reaches
	// the CRD and there is no guess for anyone to review. With the map left in
	// place it produced 63 of the 82 markers in the tree, and every one of them
	// broke "a marker always has a queue entry", because the collector scans only
	// the messages that are emitted.
	opts.Siblings = nil
	WriteMessage(&b, msg, opts)
	fmt.Fprintf(out, "\n/* %s\n", reason)
	fmt.Fprintf(out, "%s", strings.ReplaceAll(b.String(), "*/", "* /"))
	fmt.Fprintf(out, "*/\n")
}

func WriteObservedStateMessageAsComment(out io.Writer, msgDetails *OutputMessageDetails, reason string, observedStateMessages sets.String, opts WriteOptions) {
	var b bytes.Buffer
	WriteObservedStateMessage(&b, msgDetails, observedStateMessages, opts)
	fmt.Fprintf(out, "\n/* %s\n", reason)
	fmt.Fprintf(out, "%s", strings.ReplaceAll(b.String(), "*/", "* /"))
	fmt.Fprintf(out, "*/\n")
}

func WriteMessage(out io.Writer, msg protoreflect.MessageDescriptor, opts WriteOptions) {
	goType := GoNameForProtoMessage(msg)

	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "// %s=%s\n", KCCProtoMessageAnnotationMisc, msg.FullName())
	fmt.Fprintf(out, "type %s struct {\n", goType)
	for i := 0; i < msg.Fields().Len(); i++ {
		field := msg.Fields().Get(i)
		if !IsFieldBehavior(field, annotations.FieldBehavior_OUTPUT_ONLY) {
			// Only write non-output fields.
			WriteField(out, field, msg, i, false, opts, "")
		}
	}
	fmt.Fprintf(out, "}\n")
}

func WriteObservedStateMessage(out io.Writer, msgDetails *OutputMessageDetails, observedStateMessages sets.String, opts WriteOptions) {
	msg := msgDetails.Message
	goType := goNameForOutputProtoMessage(msg)

	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "// %s=%s\n", KCCProtoMessageAnnotationObservedState, msg.FullName())
	fmt.Fprintf(out, "type %s struct {\n", goType)
	// Clear PlaceServerSetFields to keep placement notes out of the nested
	// structs in types.generated.go; only the resource's own ObservedState,
	// which the scaffolder writes, carries one. Every other flag
	// carries over, or a nested observed field is named and typed differently
	// same proto field in the spec struct beside it.
	nestedOpts := opts
	nestedOpts.PlaceServerSetFields = false
	WriteObservedStateFields(out, msgDetails, observedStateMessages, nil, nestedOpts)
	fmt.Fprintf(out, "}\n")
}

// ObservedStateFieldNote describes what WriteObservedStateFields did with one
// output-only field, so the caller can file a queue entry for a field that did
// not reach the struct.
//
// Rendered holds the field's output, so a caller can pass it to
// UnsupportedFieldMarker to find a field WriteField could not type.
type ObservedStateFieldNote struct {
	// JSONName is the field as KRM spells it, e.g. "createTime".
	JSONName string
	// Skipped is true when the caller's skip map excluded the field outright.
	Skipped bool
	// Rendered is the field's output, non-empty only when it was not skipped.
	Rendered string
}

// WriteObservedStateFields writes the body of an observed-state struct: one
// field per output-only field, with no enclosing type declaration. It returns
// a note for each field.
//
// The scaffolder calls this too, for the resource-level <Kind>ObservedState, so
// the choice between a plain type and its ObservedState variant lives in one
// place.
//
// skip names proto fields to leave out, or nil to write all of them.
func WriteObservedStateFields(out io.Writer, msgDetails *OutputMessageDetails, observedStateMessages sets.String, skip map[string]bool, opts WriteOptions) []ObservedStateFieldNote {
	msg := msgDetails.Message
	emitted := 0
	var notes []ObservedStateFieldNote

	// Never emit +required from here. An observed-state struct describes what
	// GCP returned, and the API server validates status, so requiring a field
	// GCP is free to omit would make it reject a status KCC itself wrote.
	//
	// Clear that one flag rather than passing a blank WriteOptions. A blank
	// struct switches off every other flag too, and every flag added later:
	// the observed field comes out relatedUris beside the spec's relatedURIs,
	// and a message-valued map the spec types becomes a "// TODO:" marker.
	observedOpts := opts
	observedOpts.EmitRequired = false

	for _, field := range msgDetails.OutputFields {
		if skip[string(field.Name())] {
			notes = append(notes, ObservedStateFieldNote{
				JSONName: GetJSONForKRM(field, observedOpts),
				Skipped:  true,
			})
			continue
		}
		isMessage := field.Kind() == protoreflect.MessageKind && !field.IsMap()
		useObservedState := false
		if isMessage {
			if observedStateMessages.Has(string(field.Message().FullName())) {
				useObservedState = true
			}
		}
		// Each field renders into its own buffer so its note can carry the output.
		// A field WriteField cannot type becomes a "// TODO:" marker and never
		// reaches the CRD.
		var field_ bytes.Buffer
		WriteField(&field_, field, msg, emitted, useObservedState, observedOpts, placementNote(field, msg, opts))
		out.Write(field_.Bytes())
		emitted++
		notes = append(notes, ObservedStateFieldNote{
			JSONName: GetJSONForKRM(field, observedOpts),
			Rendered: field_.String(),
		})
	}
	return notes
}

// placementNote returns a +kcc:guess marker for a field IsServerSetField places
// by name, and "" for any other field.
//
// The marker goes in the generated type as well as the judgement queue. The
// allowlist will be wrong for some outlier, and the person who meets it is
// reading the type, not the queue.
func placementNote(field protoreflect.FieldDescriptor, msg protoreflect.MessageDescriptor, opts WriteOptions) string {
	if !IsServerSetField(field, msg, opts) {
		return ""
	}
	// controller-gen strips +kcc: markers from the CRD description, so a
	// reviewer reading the type sees the guess while kubectl explain still shows
	// only the proto's own comment.
	return "+kcc:guess=placement reason=no-field-behavior-on-message"
}

func GoTypeForField(field protoreflect.FieldDescriptor, isTransitiveOutput bool, opts WriteOptions) (string, error) {
	if field.IsMap() {
		entryMsg := field.Message()
		keyField := entryMsg.Fields().ByName("key")
		valueField := entryMsg.Fields().ByName("value")
		if keyField.Kind() != protoreflect.StringKind {
			// A CRD keys additionalProperties by string, so no other key type
			// can be expressed.
			return "", fmt.Errorf("unsupported map type with key %v and value %v", keyField.Kind(), valueField.Kind())
		}
		switch valueField.Kind() {
		case protoreflect.StringKind:
			return "map[string]string", nil
		case protoreflect.Int64Kind:
			return "map[string]int64", nil
		case protoreflect.MessageKind:
			// Off by default: generating these adds fields to the CRD of a
			// resource people already use.
			if !opts.EmitMessageMaps {
				return "", fmt.Errorf("unsupported map type with key %v and value %v", keyField.Kind(), valueField.Kind())
			}
			// FindDependenciesForField already follows the map entry to the
			// value's message, so its struct is generated like any other nested
			// message. A message with a special-cased Go type uses that type
			// instead, so a google.protobuf.Struct value becomes
			// apiextensionsv1.JSON.
			valueName := string(valueField.Message().FullName())
			if goType, ok := protoMessagesNotMappedToGoStruct[valueName]; ok {
				return "map[string]" + goType, nil
			}
			return "map[string]" + GoNameForProtoMessage(valueField.Message()), nil
		default:
			return "", fmt.Errorf("unsupported map type with key %v and value %v", keyField.Kind(), valueField.Kind())
		}
	}

	var goType string
	switch field.Kind() {
	case protoreflect.MessageKind:
		if isTransitiveOutput {
			goType = goNameForOutputProtoMessage(field.Message())
		} else {
			goType = GoNameForProtoMessage(field.Message())
		}
	case protoreflect.EnumKind:
		goType = "string"
	default:
		goType = goTypeForProtoKind(field.Kind())
	}

	if field.Cardinality() == protoreflect.Repeated {
		goType = "[]" + goType
	} else {
		goType = "*" + goType
	}

	// Special case for proto "bytes" type
	if goType == "*[]byte" {
		goType = "[]byte"
	}
	// Special case for proto "google.protobuf.Struct" type
	if goType == "*apiextensionsv1.JSON" {
		goType = "apiextensionsv1.JSON"
	}

	return goType, nil
}

func WriteField(out io.Writer, field protoreflect.FieldDescriptor, msg protoreflect.MessageDescriptor, fieldIndex int, isTransitiveOutput bool, opts WriteOptions, note string) {
	sourceLocations := msg.ParentFile().SourceLocations().ByDescriptor(field)

	jsonName := GetJSONForKRM(field, opts)
	GoFieldName := goFieldNameOpts(field, opts)

	// The caller's own note wins: it knows something more specific than a name
	// match, such as why a field was placed in ObservedState.
	if note == "" {
		if target, ok := SiblingResource(field, opts); ok {
			note = SiblingGuessMarker + target
		}
	}

	goType, err := GoTypeForField(field, isTransitiveOutput, opts)
	if err != nil {
		// Name the field. Without it the marker says only "unsupported map type"
		// and neither a reader nor the judgement queue can tell which field went
		// missing from the CRD.
		fmt.Fprintf(out, "\n\t// TODO: %s: %v\n\n", jsonName, err)
		return
	}

	// Blank line between fields for readability
	if fieldIndex != 0 {
		fmt.Fprintf(out, "\n")
	}

	if sourceLocations.LeadingComments != "" {
		comment := strings.TrimSpace(sourceLocations.LeadingComments)
		for _, line := range strings.Split(comment, "\n") {
			if strings.TrimSpace(line) == "" {
				fmt.Fprintf(out, "\t//\n")
			} else {
				fmt.Fprintf(out, "\t// %s\n", line)
			}
		}
	}
	// note records a choice the generated type does not otherwise show, such as
	// a field placed in ObservedState by name. It follows the proto's own
	// comment, which describes the field rather than what the generator did.
	for _, line := range strings.Split(strings.TrimSpace(note), "\n") {
		if line != "" {
			fmt.Fprintf(out, "\t// %s\n", line)
		}
	}

	fmt.Fprintf(out, "\t// %s=%s\n", KCCProtoFieldAnnotation, field.FullName())

	// +required is what produces the CRD's required: list. Without it every field
	// is optional, because we emit json:",omitempty" on all of them and that is
	// what controller-gen reads. We do not emit +optional: omitempty already
	// implies it, so the marker would be redundant.
	//
	// OUTPUT_ONLY takes precedence over REQUIRED. An output field is never supplied
	// by the user, so requiring it would make the CRD reject valid objects. We do
	// not expect one field to carry both, but field_behavior can be repeated, so it
	// is worth handling.
	if opts.EmitRequired &&
		IsFieldBehavior(field, annotations.FieldBehavior_REQUIRED) &&
		!IsFieldBehavior(field, annotations.FieldBehavior_OUTPUT_ONLY) {
		fmt.Fprintf(out, "\t// +required\n")
	}

	fmt.Fprintf(out, "\t%s %s `json:\"%s,omitempty\"`\n",
		GoFieldName,
		goType,
		jsonName,
	)
}

func deduplicateAndSort(messages []protoreflect.MessageDescriptor) []protoreflect.MessageDescriptor {
	m := make(map[string]protoreflect.MessageDescriptor)
	for _, msg := range messages {
		key := string(msg.FullName())
		m[key] = msg
	}
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	messages = []protoreflect.MessageDescriptor{}
	for _, key := range keys {
		messages = append(messages, m[key])
	}
	return messages
}

func deduplicateAndSortOutputMessages(messages []*OutputMessageDetails) []*OutputMessageDetails {
	m := make(map[string]*OutputMessageDetails)
	for _, msg := range messages {
		key := string(msg.Message.FullName())
		m[key] = msg
	}
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	messages = []*OutputMessageDetails{}
	for _, key := range keys {
		messages = append(messages, m[key])
	}
	return messages
}

// AsSnakeCase returns the given string converted to lowercase snake_case. If the input is already snake_case, no
// change is made. Any transitions in the input from lowercase to uppercase are interpreted as camelCase-style word
// transitions, and are replaced with an underscore.
func AsSnakeCase(s string) string {
	res := regexp.MustCompile("(.)([A-Z][a-z]+)").ReplaceAllString(s, "${1}_${2}")
	return strings.ToLower(regexp.MustCompile("([a-z0-9])([A-Z])").ReplaceAllString(res, "${1}_${2}"))
}

func GoNameForProtoMessage(msg protoreflect.MessageDescriptor) string {
	fullName := string(msg.FullName())

	// Some special-case values that are not obvious how to map in KRM
	if goType, ok := protoMessagesNotMappedToGoStruct[fullName]; ok {
		return goType
	}

	fullName = strings.TrimPrefix(fullName, string(msg.ParentFile().FullName()))
	fullName = strings.TrimPrefix(fullName, ".")
	// Ensure acronyms in type names are also handled.
	parts := strings.Split(fullName, ".")
	for i, part := range parts {
		partInSnakeCase := AsSnakeCase(part)
		tokens := strings.Split(partInSnakeCase, "_")
		for j, token := range tokens {
			if IsAcronym(token) {
				token = strings.ToUpper(token)
			} else {
				token = strings.Title(token)
			}
			tokens[j] = token
		}
		parts[i] = strings.Join(tokens, "")
	}
	return strings.Join(parts, "_")
}

func goNameForOutputProtoMessage(msg protoreflect.MessageDescriptor) string {
	fullName := string(msg.FullName())
	if _, ok := protoMessagesNotMappedToGoStruct[fullName]; ok {
		return GoNameForProtoMessage(msg)
	}
	return GoNameForProtoMessage(msg) + "ObservedState"
}

func goTypeForProtoKind(kind protoreflect.Kind) string {
	goType := ""
	switch kind {
	case protoreflect.StringKind:
		goType = "string"

	case protoreflect.Int32Kind:
		goType = "int32"

	case protoreflect.Int64Kind:
		goType = "int64"

	case protoreflect.Uint32Kind:
		goType = "uint32"

	case protoreflect.Uint64Kind:
		goType = "uint64"

	case protoreflect.Fixed64Kind:
		goType = "uint64"

	case protoreflect.BoolKind:
		goType = "bool"

	case protoreflect.DoubleKind:
		goType = "float64"

	case protoreflect.FloatKind:
		goType = "float32"

	case protoreflect.BytesKind:
		goType = "[]byte"

	default:
		klog.Fatalf("unhandled kind %q", kind)
	}

	return goType
}

// GetJSONForKRM returns the KRM JSON name for the field, cased according to
// opts.
//
// Pass the options the field was written with. A judgement-queue path built
// with any other options can name a field the generated type does not have:
// under EmitPluralAcronyms the struct says relatedURIs, but blank options give
// relatedUris.
func GetJSONForKRM(protoField protoreflect.FieldDescriptor, opts WriteOptions) string {
	tokens := strings.Split(string(protoField.Name()), "_")
	for i, token := range tokens {
		if i == 0 {
			// Do not capitalize first token
			continue
		}
		if cased, ok := AcronymCasing(token, opts.EmitPluralAcronyms); ok {
			token = cased
		} else {
			token = strings.Title(token)
		}
		tokens[i] = token
	}
	return strings.Join(tokens, "")
}

// goFieldName returns the KRM go name for the field,
// honoring KRM conventions
func goFieldName(protoField protoreflect.FieldDescriptor) string {
	return goFieldNameOpts(protoField, WriteOptions{})
}

func goFieldNameOpts(protoField protoreflect.FieldDescriptor, opts WriteOptions) string {
	tokens := strings.Split(string(protoField.Name()), "_")
	for i, token := range tokens {
		if cased, ok := AcronymCasing(token, opts.EmitPluralAcronyms); ok {
			token = cased
		} else {
			token = strings.Title(token)
		}
		tokens[i] = token
	}
	return strings.Join(tokens, "")
}

// FindDependenciesForMessage recursively explores the dependent proto messages of the given message.
func FindDependenciesForMessage(message protoreflect.MessageDescriptor, ignoredFields sets.String) ([]protoreflect.MessageDescriptor, error) {
	msgs := make(map[string]protoreflect.MessageDescriptor)
	for i := 0; i < message.Fields().Len(); i++ {
		field := message.Fields().Get(i)
		FindDependenciesForField(field, msgs, ignoredFields)
	}

	RemoveNotMappedToGoStruct(msgs)

	res := []protoreflect.MessageDescriptor{}
	for _, msg := range msgs {
		res = append(res, msg)
	}
	return res, nil
}

// FindDependenciesForField recursively explores the dependent proto messages of the given field.
func FindDependenciesForField(field protoreflect.FieldDescriptor, deps map[string]protoreflect.MessageDescriptor, ignoredFields sets.String) {
	if ignoredFields.Has(string(field.FullName())) {
		return
	}

	if field.Message() != nil { // no need to find dependencies for proto messages that are not mapped to KRM Go struct
		if _, ok := protoMessagesNotMappedToGoStruct[string(field.Message().FullName())]; ok {
			return
		}
	}

	if field.IsMap() {
		mapEntry := field.Message()
		if keyField := mapEntry.Fields().ByName("key"); keyField != nil {
			FindDependenciesForField(keyField, deps, ignoredFields)
		}
		if valueField := mapEntry.Fields().ByName("value"); valueField != nil {
			FindDependenciesForField(valueField, deps, ignoredFields)
		}
	} else {
		switch field.Kind() {
		case protoreflect.MessageKind:
			msg := field.Message()
			fqn := string(msg.FullName())
			if _, ok := deps[fqn]; !ok {
				deps[fqn] = msg
				for i := 0; i < msg.Fields().Len(); i++ {
					field := msg.Fields().Get(i)
					FindDependenciesForField(field, deps, ignoredFields)
				}
			}
		case protoreflect.EnumKind:
			// deps[string(field.Enum().FullName())] = true  // Skip enum because enum is mapped to Go string in code generation
		}
	}
}

func RemoveNotMappedToGoStruct(msgs map[string]protoreflect.MessageDescriptor) {
	for msg := range protoMessagesNotMappedToGoStruct {
		delete(msgs, msg)
	}
}

func isPrimitive(field protoreflect.FieldDescriptor) bool {
	if field.Kind() != protoreflect.MessageKind || field.IsMap() {
		return true
	}
	if field.Message() != nil {
		if _, ok := protoMessagesNotMappedToGoStruct[string(field.Message().FullName())]; ok {
			return true
		}
	}
	return false
}

func IsFieldBehavior(field protoreflect.FieldDescriptor, fieldBehavior annotations.FieldBehavior) bool {
	d := field.Options()
	fieldBehaviors := proto.GetExtension(d, annotations.E_FieldBehavior).([]annotations.FieldBehavior)
	for _, f := range fieldBehaviors {
		if f == fieldBehavior {
			return true
		}
	}
	return false
}

// serverSetFieldNames are fields GCP computes that some protos do not mark
// OUTPUT_ONLY.
//
// A field's name is normally a poor guide to where it belongs, so the list is
// narrow. Every name here appears zero times in a resource-level Spec across
// the baseline tree, with one exception:
//
// etag appears twice, on AlloyDBCluster and ContainerAttachedCluster, where it
// is an optimistic-concurrency input that "can be sent on update and delete
// requests". Against that, 27 greenfield resources carry it status-side and
// none spec-side, so the list keeps it and the queue asks a human.
//
// Three names stay out of the list, because how upstream uses them differs
// from what their counts suggest:
//
//	state   7 upstream Specs, and they are desired state, not observed state.
//	        ConfigDeliveryFleetPackage calls it "the desired state of the fleet
//	        package"; DLPConnection marks it Required; VPCFlowLogsConfig is an
//	        enable/disable toggle defaulting to ENABLED. Moving it would take a
//	        settable field away.
//	status  2 upstream Specs, both required input. DLPDiscoveryConfig marks it
//	        Required. AccessContextManagerServicePerimeter differs: in the GCP
//	        API "status" names the enforced perimeter config as opposed to the
//	        dry-run "spec", so it is user-authored configuration whose name
//	        happens to collide with the CRD's own status.
//	type    36 upstream Specs.
//
// "name" is not in the list because identityFields in the scaffold package
// already handles it.
var serverSetFieldNames = map[string]bool{
	"createTime":        true,
	"updateTime":        true,
	"deleteTime":        true,
	"creationTimestamp": true,
	"uid":               true,
	"selfLink":          true,
	"selfLinkWithID":    true,
	"id":                true,
	"kind":              true,
	"etag":              true,
}

// IsServerSetField reports whether a field should go to ObservedState even
// though the proto does not say so.
//
// The guard is that msg carries no google.api.field_behavior on any field.
// Where one field carries it, the proto's author placed the others on purpose.
// Compute messages carry none, because they come from a discovery document, so
// without this rule their ObservedState is empty while creationTimestamp and
// selfLink sit in the Spec for a user to set and GCP to overwrite.
//
// A guard that only required this field to be unannotated would recover nine
// more fields, almost all of them etag in protos that do annotate other
// fields. Those protos are the most likely to have left etag unannotated on
// purpose.
//
// Call it only on the resource's own message. TypeGenerator does that
// through isServerSet.
func IsServerSetField(field protoreflect.FieldDescriptor, msg protoreflect.MessageDescriptor, opts WriteOptions) bool {
	if !opts.PlaceServerSetFields || msg == nil {
		return false
	}
	if hasAnyFieldBehavior(msg) {
		return false
	}
	return serverSetFieldNames[GetJSONForKRM(field, opts)]
}

// hasAnyFieldBehavior reports whether any field of msg carries a
// google.api.field_behavior annotation.
func hasAnyFieldBehavior(msg protoreflect.MessageDescriptor) bool {
	for i := 0; i < msg.Fields().Len(); i++ {
		d := msg.Fields().Get(i).Options()
		if len(proto.GetExtension(d, annotations.E_FieldBehavior).([]annotations.FieldBehavior)) > 0 {
			return true
		}
	}
	return false
}

// UnsupportedField is a proto field the generator could not produce a Go type
// for. The field is omitted from the generated struct, so it never reaches the
// CRD.
type UnsupportedField struct {
	// Message is the fully-qualified proto message that owns the field.
	Message string
	// Field is the KRM JSON name the field would have had.
	Field string
	// Reason is the generator's own explanation.
	Reason string
}

// UnsupportedFields returns everything the generator could not type during this
// run, for the judgement queue.
func (g *TypeGenerator) UnsupportedFields() []UnsupportedField {
	return g.unsupportedFields
}

// SiblingGuesses returns the nested-message fields the sibling rule flagged
// during the visit, so the caller can record them in the judgement queue.
// PrepopulateSpec cannot: it sees only the resource's own fields.
func (g *TypeGenerator) SiblingGuesses() []SiblingGuess {
	return g.siblingGuesses
}

// scanUnsupported returns every unsupported-field marker in a rendered
// message body.
func scanUnsupported(msgName, body string) []UnsupportedField {
	var out []UnsupportedField
	for _, line := range strings.Split(body, "\n") {
		if field, reason, ok := UnsupportedFieldMarker(line); ok {
			out = append(out, UnsupportedField{Message: msgName, Field: field, Reason: reason})
		}
	}
	return out
}

// UnsupportedFieldMarker returns the field name and reason from the first
// "// TODO: <field>: <reason>" marker in rendered. WriteField writes that
// marker in place of a field it cannot type, and the field never reaches the
// CRD. A marker without a field name returns an empty field.
//
// The 239-resource run had 15 such markers in scaffolded type files and 37 more
// in types.generated.go. Between them they lost 124 CRD field paths, and
// neither the judgement queue nor any report listed them.
func UnsupportedFieldMarker(rendered string) (field, reason string, ok bool) {
	for _, line := range strings.Split(rendered, "\n") {
		after, found := strings.CutPrefix(strings.TrimSpace(line), "// TODO: ")
		if !found {
			continue
		}
		if field, reason, named := strings.Cut(after, ": "); named {
			return field, reason, true
		}
		return "", after, true
	}
	return "", "", false
}
