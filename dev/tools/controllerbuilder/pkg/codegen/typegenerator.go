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
	unsupportedFields       []UnsupportedField
	generatedFileAnnotation *codegenannotations.FileAnnotation
	includeSkippedOutput    bool
	writeOptions            WriteOptions

	// handWritten is what the target package already defines by hand, scanned once when we visit
	// the first message. See handWrittenTypes for why the decision to split a message and the
	// decision to write it both have to read from here.
	handWritten *handWrittenTypes
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

// ObservedStateMessages is the set of proto messages that got an ObservedState
// struct of their own. A field whose message is in here must be referenced as
// <Proto>ObservedState rather than the plain type.
func (g *TypeGenerator) ObservedStateMessages() sets.String {
	return g.observedStateMessages
}

// OutputFieldsFor returns the output-only fields of one message, as computed by
// identifyOutputs during the visit. The scaffolder uses this to fill the
// resource-level ObservedState struct rather than re-walking the proto: the rule
// is transitive, since a field is output-only if it is reached through an
// OUTPUT_ONLY parent, and a second implementation would drift from this one.
//
// Reports false when the message has no output-only content, which is the signal
// to leave the scaffolded struct empty.
func (g *TypeGenerator) OutputFieldsFor(fqn string) (*OutputMessageDetails, bool) {
	for _, details := range g.outputMessages {
		if string(details.Message.FullName()) == fqn {
			return details, true
		}
	}
	return nil, false
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
		isOut := IsFieldBehavior(f, annotations.FieldBehavior_OUTPUT_ONLY)

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

		// Rendered to a scratch buffer first so the markers WriteField leaves for
		// fields it could not type can be collected. Otherwise the field is absent
		// from the CRD and the only trace is a comment nobody reads.
		var body bytes.Buffer
		WriteMessage(&body, msg, g.writeOptions)
		g.unsupportedFields = append(g.unsupportedFields, scanUnsupported(string(msg.FullName()), body.String())...)
		out.body.Write(body.Bytes())
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
		skipGenerated := true
		goType, err := g.findTypeDeclaration(goTypeName, out.OutputDir(), skipGenerated)
		if err != nil {
			return fmt.Errorf("looking up go type: %w", err)
		}
		if goType != nil {
			klog.V(1).Infof("found existing non-generated go type %q, won't generate", goTypeName)
			if g.includeSkippedOutput {
				WriteObservedStateMessageAsComment(&out.body, msgDetails, fmt.Sprintf("found existing non-generated go type %q, skipping", goTypeName), g.observedStateMessages)
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
				WriteObservedStateMessageAsComment(&out.body, msgDetails, fmt.Sprintf("found existing non-generated go type with proto tag %q, skipping", msg.FullName()), g.observedStateMessages)
			}
			continue
		}

		WriteObservedStateMessage(&out.body, msgDetails, g.observedStateMessages)
	}
	return errors.Join(g.errors...)
}

func WriteMessageAsComment(out io.Writer, msg protoreflect.MessageDescriptor, reason string, opts WriteOptions) {
	var b bytes.Buffer
	WriteMessage(&b, msg, opts)
	fmt.Fprintf(out, "\n/* %s\n", reason)
	fmt.Fprintf(out, "%s", strings.ReplaceAll(b.String(), "*/", "* /"))
	fmt.Fprintf(out, "*/\n")
}

func WriteObservedStateMessageAsComment(out io.Writer, msgDetails *OutputMessageDetails, reason string, observedStateMessages sets.String) {
	var b bytes.Buffer
	WriteObservedStateMessage(&b, msgDetails, observedStateMessages)
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
			WriteField(out, field, msg, i, false, opts)
		}
	}
	fmt.Fprintf(out, "}\n")
}

func WriteObservedStateMessage(out io.Writer, msgDetails *OutputMessageDetails, observedStateMessages sets.String) {
	msg := msgDetails.Message
	goType := goNameForOutputProtoMessage(msg)

	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "// %s=%s\n", KCCProtoMessageAnnotationObservedState, msg.FullName())
	fmt.Fprintf(out, "type %s struct {\n", goType)
	WriteObservedStateFields(out, msgDetails, observedStateMessages, nil)
	fmt.Fprintf(out, "}\n")
}

// WriteObservedStateFields writes the body of an observed-state struct: one field
// per output-only field, with no enclosing type declaration.
//
// The scaffolder calls this too, to fill the resource-level <Kind>ObservedState in
// the hand-written types file. Both callers go through here so the plain-versus-
// ObservedState choice below has exactly one implementation; a second copy in the
// scaffolder would drift from this one the first time either changed.
//
// skip names proto fields to leave out, or nil to write all of them.
// ObservedStateFieldNote records an output-only field that did not reach the
// struct cleanly, so the caller can say so rather than dropping it in silence.
//
// Rendered carries the field's own output for the caller to inspect. The reason
// a type was declined is spelled in the "// TODO:" comment WriteField leaves
// behind, and parsing that belongs with the queue rather than here.
type ObservedStateFieldNote struct {
	// JSONName is the field as KRM spells it, e.g. "createTime".
	JSONName string
	// Skipped is true when the caller's skip map excluded the field outright.
	Skipped bool
	// Rendered is the field's output, non-empty only when it was not skipped.
	Rendered string
}

func WriteObservedStateFields(out io.Writer, msgDetails *OutputMessageDetails, observedStateMessages sets.String, skip map[string]bool) []ObservedStateFieldNote {
	msg := msgDetails.Message
	emitted := 0
	var notes []ObservedStateFieldNote
	for _, field := range msgDetails.OutputFields {
		if skip[string(field.Name())] {
			notes = append(notes, ObservedStateFieldNote{
				JSONName: GetJSONForKRM(field),
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
		// Rendered separately so the field's own output can be inspected before it
		// is appended, the same way the Spec does it. A field whose type the
		// generator declines becomes a "// TODO:" comment and never reaches the
		// CRD, which is a silent drop unless somebody records it.
		var field_ bytes.Buffer
		// Never emit +required from here. An observed-state struct describes what GCP
		// returned, and the API server validates status, so requiring a field GCP is
		// free to omit would make it reject a status KCC itself wrote.
		WriteField(&field_, field, msg, emitted, useObservedState, WriteOptions{})
		out.Write(field_.Bytes())
		emitted++
		notes = append(notes, ObservedStateFieldNote{
			JSONName: GetJSONForKRM(field),
			Rendered: field_.String(),
		})
	}
	return notes
}

func GoTypeForField(field protoreflect.FieldDescriptor, isTransitiveOutput bool) (string, error) {
	if field.IsMap() {
		entryMsg := field.Message()
		keyField := entryMsg.Fields().ByName("key")
		valueField := entryMsg.Fields().ByName("value")
		if keyField.Kind() != protoreflect.StringKind {
			// A CRD keys additionalProperties by string; nothing else is expressible.
			return "", fmt.Errorf("unsupported map type with key %v and value %v", keyField.Kind(), valueField.Kind())
		}
		switch valueField.Kind() {
		case protoreflect.StringKind:
			return "map[string]string", nil
		case protoreflect.Int64Kind:
			return "map[string]int64", nil
		case protoreflect.MessageKind:
			// The value struct is generated like any other nested message:
			// FindDependenciesForField already recurses through the map entry into
			// the value, so it is visited and written without new machinery here.
			// A CRD expresses this as additionalProperties with an object schema.
			// A message with a special-cased Go type takes that type here too,
			// rather than the struct name it does not have. google.protobuf.Struct
			// and Value both land on apiextensionsv1.JSON, which is what upstream
			// writes by hand for Firestore's Document.fields, a map<string, Value>.
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

func WriteField(out io.Writer, field protoreflect.FieldDescriptor, msg protoreflect.MessageDescriptor, fieldIndex int, isTransitiveOutput bool, opts WriteOptions) {
	sourceLocations := msg.ParentFile().SourceLocations().ByDescriptor(field)

	jsonName := getJSONForKRM(field, opts)
	GoFieldName := goFieldNameOpts(field, opts)

	goType, err := GoTypeForField(field, isTransitiveOutput)
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

// GetJSONForKRM returns the KRM JSON name for the field,
// honoring KRM conventions
func GetJSONForKRM(protoField protoreflect.FieldDescriptor) string {
	return getJSONForKRM(protoField, WriteOptions{})
}

func getJSONForKRM(protoField protoreflect.FieldDescriptor, opts WriteOptions) string {
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

// scanUnsupported pulls the "// TODO: <field>: <reason>" markers out of a
// rendered message body.
func scanUnsupported(msgName, body string) []UnsupportedField {
	var out []UnsupportedField
	for _, line := range strings.Split(body, "\n") {
		after, ok := strings.CutPrefix(strings.TrimSpace(line), "// TODO: ")
		if !ok {
			continue
		}
		field, reason, found := strings.Cut(after, ": ")
		if !found {
			field, reason = "", after
		}
		out = append(out, UnsupportedField{Message: msgName, Field: field, Reason: reason})
	}
	return out
}
