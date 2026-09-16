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

package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/codegen"
	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/options"
	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/protoapi"
	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/template/apis"
	"github.com/fatih/color"
	"google.golang.org/protobuf/reflect/protoreflect"
	"k8s.io/klog/v2"
)

type APIScaffolder struct {
	BaseDir         string
	GoPackage       string
	Group           string
	Version         string
	PackageProtoTag string

	// Proto is optional. When set, the scaffolder reads google.api.resource to
	// learn the resource's real collection segment and parent shape instead of
	// guessing them. Without it the templates fall back to their old guesses,
	// which are wrong for most resources: the collection segment disagrees with
	// the declared pattern for 752 of 1417 annotated messages, and the assumed
	// projects/locations parent holds for about a third.
	Proto *protoapi.Proto

	// Siblings maps a lowercased Kind suffix to a Kind this service declares, so
	// a parent segment naming one can be flagged as a probable reference. See
	// codegen.SiblingResourceByName.
	Siblings map[string]string
}

// resourceMetadata looks up what the proto states about a resource, or nil if we
// have no proto loaded or the message is not an annotated resource.
func (a *APIScaffolder) resourceMetadata(fullName string) *protoapi.ResourceMetadata {
	if a.Proto == nil || fullName == "" {
		return nil
	}
	d, err := a.Proto.Files().FindDescriptorByName(protoreflect.FullName(fullName))
	if err != nil {
		// PackageProtoTag may name several proto packages, and ProtoMessageFullName
		// composes the name from the first one. grafeas is the case: it is invoked
		// with "google.cloud.grafeas.v1,grafeas.v1" while Note lives in the second,
		// so the first guess never resolves. Try the rest before giving up.
		d = nil
		for _, pkg := range strings.Split(a.PackageProtoTag, ",") {
			candidate := pkg + "." + lastSegment(fullName)
			if alt, altErr := a.Proto.Files().FindDescriptorByName(protoreflect.FullName(candidate)); altErr == nil {
				d = alt
				break
			}
		}
		if d == nil {
			klog.V(2).Infof("no descriptor for %q, scaffolding will guess: %v", fullName, err)
			return nil
		}
	}
	msg, ok := d.(protoreflect.MessageDescriptor)
	if !ok {
		return nil
	}
	return protoapi.GetResourceMetadata(msg)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	if err == nil {
		return true
	}
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	klog.Fatalf("unexpected error checking for file %q: %v", p, err)
	return false
}

func (a *APIScaffolder) RefsFileExist(resource options.Resource) bool {
	return fileExists(a.PathToRefsFile(resource))
}

func (a *APIScaffolder) PathToRefsFile(resource options.Resource) string {
	fileName := strings.ToLower(resource.Kind) + "_reference.go"
	return filepath.Join(a.BaseDir, a.GoPackage, fileName)
}

func (a *APIScaffolder) AddRefsFile(resource options.Resource) error {
	refsFilePath := a.PathToRefsFile(resource)
	cArgs := a.buildAPIArgs(&resource)
	return scaffoldRefsFile(refsFilePath, cArgs)
}

func scaffoldIdentityFile(path string, cArgs *apis.APIArgs) error {
	tmpl, err := template.New(cArgs.Kind).Funcs(funcMap).Parse(apis.IdentityTemplate)
	if err != nil {
		return fmt.Errorf("parse %s_identity.go template: %w", strings.ToLower(cArgs.ProtoResource), err)
	}
	// Apply the APIArgs args to the template
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, cArgs); err != nil {
		return err
	}
	// Format on write. The template now branches on the parent shape, and getting
	// blank lines right by hand across those branches is fiddly and not worth it.
	if err := FormatImports(path, out.Bytes()); err != nil {
		return err
	}
	color.HiGreen("New identity file added %s\nPlease EDIT it!\n", path)
	return nil
}

func (a *APIScaffolder) IdentityFileExist(resource options.Resource) bool {
	return fileExists(a.PathToIdentityFile(resource))
}

func (a *APIScaffolder) PathToIdentityFile(resource options.Resource) string {
	fileName := strings.ToLower(resource.Kind) + "_identity.go"
	return filepath.Join(a.BaseDir, a.GoPackage, fileName)
}

// Populates an APIArgs for templating.  Arguments are optional.
func (a *APIScaffolder) buildAPIArgs(resource *options.Resource) *apis.APIArgs {
	args := &apis.APIArgs{
		Group:           a.Group,
		Version:         a.Version,
		PackageProtoTag: a.PackageProtoTag,
	}

	if resource != nil {
		args.Kind = resource.Kind

		if strings.Contains(resource.ProtoName, ".") {
			args.KindProtoTag = resource.ProtoName
		} else {
			pkg := a.PackageProtoTag
			if strings.Contains(pkg, ",") {
				pkg = strings.Split(pkg, ",")[0]
			}
			args.KindProtoTag = pkg + "." + resource.ProtoName
		}
		args.ProtoResource = resource.ProtoName

		args.ProtoMessageName = resource.ProtoMessageName()
		args.ProtoMessageFullName = resource.ProtoMessageFullName(a.PackageProtoTag)

		if md := a.resourceMetadata(args.ProtoMessageFullName); md != nil {
			args.Collection = md.Collection
			args.ParentStyle = string(md.ParentStyle)
			args.ResourcePattern = md.Pattern
		}
		if args.Collection == "" {
			// No pattern to read. Fall back to what the template did before, so
			// behavior is unchanged for anything the proto does not describe.
			args.Collection = strings.ToLower(args.ProtoMessageName) + "s"
		}
		if args.ParentStyle == "" {
			args.ParentStyle = string(protoapi.ParentUnknown)
		}
	}
	// The default for every path, including the stub with no proto at all:
	// nearly every resource is project-rooted, and AddTypeFile overrides this
	// where the pattern says otherwise.
	if args.RootRefType == "" {
		args.RootRefType, args.RootRefField = "ProjectRef", "projectRef"
		args.RootRefDescription = "The project that this resource belongs to."
	}

	return args
}

// repoRoot is the tree BaseDir sits in, so the shared refs package can be found
// whether we are writing to apis/ or to a scratch tree via --output-api.
func (a *APIScaffolder) repoRoot() string {
	if abs, err := filepath.Abs(a.BaseDir); err == nil {
		return filepath.Dir(abs)
	}
	return filepath.Dir(a.BaseDir)
}

func (a *APIScaffolder) AddIdentityFile(resource options.Resource) error {
	refsFilePath := a.PathToIdentityFile(resource)
	cArgs := a.buildAPIArgs(&resource)
	return scaffoldIdentityFile(refsFilePath, cArgs)
}

func scaffoldRefsFile(path string, cArgs *apis.APIArgs) error {
	tmpl, err := template.New(cArgs.Kind).Funcs(funcMap).Parse(apis.RefsHeaderTemplate)
	if err != nil {
		return fmt.Errorf("parse %s_reference.go template: %w", strings.ToLower(cArgs.ProtoResource), err)
	}
	// Apply the APIArgs args to the template
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, cArgs); err != nil {
		return err
	}
	// Write the generated <kind>_types.go
	if err := WriteToFile(path, out.Bytes()); err != nil {
		return err
	}
	color.HiGreen("New reference file added %s\nPlease EDIT it!\n", path)
	return nil
}

func (a *APIScaffolder) TypeFileExists(resource options.Resource) bool {
	return fileExists(a.PathToTypeFile(resource))
}

func (a *APIScaffolder) PathToTypeFile(resource options.Resource) string {
	fileName := strings.ToLower(resource.Kind) + "_types.go"
	return filepath.Join(a.BaseDir, a.GoPackage, fileName)
}

// AddTypeFile scaffolds <kind>_types.go.
//
// When prepopulated is non-nil, the Spec and ObservedState bodies come from the
// proto rather than the three-field stub, and the file imports what those
// bodies reference. The caller builds prepopulated, because only it knows which
// proto message a resource maps to.
func (a *APIScaffolder) AddTypeFile(resource options.Resource, prepopulated *PrepopulateResult) error {
	typeFilePath := a.PathToTypeFile(resource)
	cArgs := a.buildAPIArgs(&resource)
	cArgs.SkipGVK = packageDeclaresGVK(filepath.Join(a.BaseDir, a.GoPackage), cArgs.Kind)
	if prepopulated != nil {
		known := refTypesInPackage(a.repoRoot(), filepath.Join(a.BaseDir, a.GoPackage))
		var parentRefs bytes.Buffer
		var emitted, siblings []string
		segments := parentSegments(cArgs.ResourcePattern)
		for i, seg := range segments {
			collection, variable := seg[0], seg[1]
			// The root segment becomes projectRef / organizationRef / folderRef,
			// which the template renders.
			if i == 0 {
				continue
			}
			if field, ok := locationFieldNames[collection]; ok {
				// Required only for the projects/locations shape, which the
				// template renders the same way, so that output is unchanged.
				// Elsewhere the parent already fixes a location, and requiring
				// one would add a constraint upstream does not have.
				required := cArgs.ParentStyle == string(protoapi.ParentProjectLocation)
				parentRefs.WriteString(locationField(field, cArgs.ResourcePattern, required))
				emitted = append(emitted, ".spec."+field)
				continue
			}
			name := lowerCamel(variable)
			goType := strings.ToUpper(name[:1]) + name[1:] + "Ref"
			qualifier, ok := known[goType]
			// A parent segment is synthesised from the resource pattern, so no
			// proto field carries its name and the sibling rule has to be asked
			// about the name directly.
			sibling, _ := codegen.SiblingResourceByName(name, a.Siblings)
			suffix := "Ref"
			if !ok {
				// No ref type anywhere, so a plain string. Still better than
				// nothing, and upstream may well want a reference here.
				goType, suffix = "", ""
			}
			parentRefs.WriteString(parentRefField(name, goType, qualifier, cArgs.ResourcePattern, sibling))
			emitted = append(emitted, ".spec."+name+suffix)
			if sibling != "" {
				siblings = append(siblings, name+" matches "+sibling)
			}
		}
		// One entry names the whole parent. Of the 44 hand-written kinds whose
		// pattern has a segment below project and location, about 32 name the
		// direct parent with a typed ref and five inline a Parent struct holding
		// one. None carries a field per segment, so the decision a reader makes
		// is about the shape of the parent, not about each segment in turn.
		if len(emitted) > 0 {
			detail := "emitted from the pattern " + cArgs.ResourcePattern + ": " +
				strings.Join(emitted, ", ") + ". Upstream usually names the direct " +
				"parent as a reference and carries project and location as plain " +
				"fields; confirm the shape, and whether other references in the API " +
				"tree belong here"
			if len(siblings) > 0 {
				detail += " (" + strings.Join(siblings, "; ") + ", declared by this service)"
			}
			prepopulated.Judgement = append(prepopulated.Judgement, JudgementItem{
				Reason: "parent-fields-guessed",
				Detail: detail,
			})
		}

		cArgs.ParentRefFields = parentRefs.String()

		// An organization- or folder-rooted resource has no project, and emitting
		// projectRef for one gives it a field upstream does not have while leaving
		// out the one it does.
		if len(segments) > 0 {
			switch segments[0][0] {
			case "organizations":
				cArgs.RootRefType, cArgs.RootRefField = "OrganizationRef", "organizationRef"
				cArgs.RootRefDescription = "The organization that this resource belongs to."
			case "folders":
				cArgs.RootRefType, cArgs.RootRefField = "FolderRef", "folderRef"
				cArgs.RootRefDescription = "The folder that this resource belongs to."
			}
		}

		// parentSegmentJudgement covers the remaining case: a proto with no
		// google.api.resource, where there is no pattern to walk and so no
		// segment to emit or to name.
		if len(segments) == 0 {
			prepopulated.Judgement = append(prepopulated.Judgement,
				parentSegmentJudgement(cArgs.ResourcePattern, cArgs.ParentStyle)...)
		}

		cArgs.SpecFields = prepopulated.SpecFields
		cArgs.ObservedStateFields = prepopulated.ObservedStateFields
		cArgs.ExtraImports = prepopulated.ExtraImports
	}
	return scaffoldTypeFile(typeFilePath, cArgs)
}

func scaffoldTypeFile(path string, cArgs *apis.APIArgs) error {
	tmpl, err := template.New(cArgs.Kind).Funcs(funcMap).Parse(apis.TypesTemplate)
	if err != nil {
		return fmt.Errorf("parse %s_types.go template: %w", strings.ToLower(cArgs.ProtoResource), err)
	}
	// Apply the APIArgs args to the template
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, cArgs); err != nil {
		return err
	}
	// Write the generated <kind>_types.go
	if err := WriteToFile(path, out.Bytes()); err != nil {
		return err
	}
	// Format and adjust the go imports in the generated files.
	if err := FormatImports(path, out.Bytes()); err != nil {
		return err
	}
	color.HiGreen("New API file added %s\nPlease EDIT it!\n", path)
	return nil
}

func (a *APIScaffolder) GroupVersionFileNotExist() bool {
	docFilePath := filepath.Join(a.BaseDir, a.GoPackage, "groupversion_info.go")
	_, err := os.Stat(docFilePath)
	if err == nil {
		return false
	}
	return errors.Is(err, os.ErrNotExist)
}

func (a *APIScaffolder) AddGroupVersionFile() error {
	docFilePath := filepath.Join(a.BaseDir, a.GoPackage, "groupversion_info.go")
	cArgs := a.buildAPIArgs(nil)
	return scaffoldGroupVersionFile(docFilePath, cArgs)
}

func (a *APIScaffolder) DocFileNotExist() bool {
	docFilePath := filepath.Join(a.BaseDir, a.GoPackage, "doc.go")
	_, err := os.Stat(docFilePath)
	if err == nil {
		return false
	}
	return errors.Is(err, os.ErrNotExist)
}

func (a *APIScaffolder) AddDocFile() error {
	docFilePath := filepath.Join(a.BaseDir, a.GoPackage, "doc.go")
	cArgs := a.buildAPIArgs(nil)
	return scaffoldDocFile(docFilePath, cArgs)
}

func scaffoldDocFile(path string, cArgs *apis.APIArgs) error {
	tmpl, err := template.New("doc.go").Parse(apis.DocTemplate)
	if err != nil {
		return fmt.Errorf("parse doc.go template: %w", err)
	}
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, cArgs); err != nil {
		return err
	}
	if err := WriteToFile(path, out.Bytes()); err != nil {
		return err
	}
	color.HiGreen("New file added %q\n", path)
	return nil
}

func scaffoldGroupVersionFile(path string, cArgs *apis.APIArgs) error {
	tmpl, err := template.New("groupversioninfo.go").Parse(apis.GroupVersionInfoTemplate)
	if err != nil {
		return fmt.Errorf("parse groupversion_info.go template: %w", err)
	}
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, cArgs); err != nil {
		return err
	}
	if err := WriteToFile(path, out.Bytes()); err != nil {
		return err
	}
	color.HiGreen("New file added %q\n", path)
	return nil
}

const sharedRefsPackage = "apis/refs/v1beta1"

// refTypesInPackage lists the <X>Ref types available to the target package,
// mapping each to the qualifier it must be written with.
//
// Both directories are scanned, because a ref type can live in either.
// scaffoldRefsFile writes one per resource into the service package, so a
// resource whose parent is already generated has a local type to name. The
// shared package holds the ones every service needs, OrganizationRef among
// them, which 14 resources use.
func refTypesInPackage(repoRoot, serviceDir string) map[string]string {
	out := map[string]string{}
	scan := func(dir, qualifier string) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		re := regexp.MustCompile(`(?m)^type (\w+Ref) struct`)
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
				continue
			}
			body, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			for _, m := range re.FindAllStringSubmatch(string(body), -1) {
				// A type in the service package wins: it is more specific than
				// the shared one and needs no import.
				if _, seen := out[m[1]]; !seen || qualifier == "" {
					out[m[1]] = qualifier
				}
			}
		}
	}
	scan(filepath.Join(repoRoot, sharedRefsPackage), "refsv1beta1")
	scan(serviceDir, "")
	return out
}

// packageDeclaresGVK reports whether the target package already declares
// <Kind>GVK.
//
// scaffoldRefsFile writes the GVK into <kind>_reference.go, where upstream
// keeps it, and the types template writes one as well. A resource with both
// files would declare it twice, and the package would not compile.
func packageDeclaresGVK(dir, kind string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	want := regexp.MustCompile(`(?m)^var ` + regexp.QuoteMeta(kind) + `GVK\b`)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") ||
			strings.HasSuffix(e.Name(), "_types.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err == nil && want.Match(body) {
			return true
		}
	}
	return false
}

func SiblingResources(dir, service string, alsoKnown ...string) map[string]string {
	out := map[string]string{}
	add := func(kind string) {
		trimmed := strings.TrimPrefix(strings.ToLower(kind), strings.ToLower(service))
		if trimmed != "" && trimmed != strings.ToLower(kind) {
			out[trimmed] = kind
		}
	}
	for _, kind := range alsoKnown {
		add(kind)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	kindRe := regexp.MustCompile(`(?m)^type (\w+)Spec struct`)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_types.go") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		for _, m := range kindRe.FindAllStringSubmatch(string(body), -1) {
			add(m[1])
		}
	}
	return out
}

func parentSegments(pattern string) [][2]string {
	toks := strings.Split(pattern, "/")
	var pairs [][2]string
	for i := 0; i+1 < len(toks); i += 2 {
		if !strings.HasPrefix(toks[i+1], "{") {
			break
		}
		pairs = append(pairs, [2]string{toks[i], strings.Trim(toks[i+1], "{}")})
	}
	if len(pairs) > 0 && strings.HasSuffix(pattern, "}") {
		pairs = pairs[:len(pairs)-1] // the resource's own id
	}
	return pairs
}

var locationFieldNames = map[string]string{
	"locations": "location",
	"regions":   "region",
	"zones":     "zone",
}

// parentRefField returns a spec field naming one segment of the resource's
// parent: a typed ref when goType is set, a plain string when it is empty.
//
// A plain string is what upstream uses where no ref type exists. BigtableTable
// carries Spec.Instance and KMSCryptoKeyVersion Spec.CryptoKey by those names.
//
// The first line becomes the CRD description, so it says what the field is
// for. The +kcc: marker below it records that the generator chose the field
// from the pattern; controller-gen strips markers, so a reviewer reading the
// type sees the guess while kubectl explain shows only the description.
func parentRefField(segment, goType, qualifier, pattern, sibling string) string {
	name := strings.ToUpper(segment[:1]) + segment[1:]
	if goType == "" {
		// The marker names the sibling when there is one, because that is the
		// finding: this string is probably a ref to FirestoreDatabase. It
		// replaces the pattern rather than joining it, since a marker cannot be
		// wrapped and a long pattern pushes the line past 80 columns on its own.
		// The queue entry carries the pattern.
		detail := "pattern=" + pattern
		if sibling != "" {
			detail = "target=" + sibling
		}
		return fmt.Sprintf(`
	// The %s that this resource belongs to.
	// +kcc:guess=parent-segment %s
	%s *string `+"`"+`json:"%s,omitempty"`+"`"+`
`, name, detail, name, segment)
	}
	qualified := goType
	if qualifier != "" {
		qualified = qualifier + "." + goType
	}
	return fmt.Sprintf(`
	// The %s that this resource belongs to.
	// +kcc:guess=parent-ref target=%s pattern=%s
	%sRef *%s `+"`"+`json:"%sRef,omitempty"`+"`"+`
`, name, qualified, pattern, name, qualified, segment)
}

// locationField renders the resource's own location, named as upstream does.
func locationField(name, pattern string, required bool) string {
	tag := name + ",omitempty"
	if required {
		tag = name
	}
	return fmt.Sprintf(`
	// The location of this resource.
	// +kcc:guess=parent-location pattern=%s
	%s *string `+"`"+`json:"%s"`+"`"+`
`, pattern, strings.ToUpper(name[:1])+name[1:], tag)
}

func parentSegmentJudgement(pattern, parentStyle string) []JudgementItem {
	if parentStyle == string(protoapi.ParentUnknown) || pattern == "" {
		// No google.api.resource, so there is no pattern to walk. The entry says
		// that, and names location, because a regional resource whose location
		// is missing cannot be addressed at all.
		return []JudgementItem{{
			FieldPath: ".spec.location",
			Reason:    "location-omitted-unknown-parent",
			Detail: "the proto declares no google.api.resource, so the parent shape is unknown; " +
				"add location if the resource is regional",
		}}
	}

	// projectRef and resourceID are always emitted; location as well, but only
	// when the parent is exactly projects/locations.
	produced := map[string]bool{"project": true}
	if parentStyle == string(protoapi.ParentProjectLocation) {
		produced["location"] = true
	}

	var out []JudgementItem
	for _, v := range protoapi.ParentVariables(pattern) {
		if produced[v] {
			continue
		}
		item := JudgementItem{
			FieldPath: ".spec." + lowerCamel(v),
			Reason:    "parent-segment-omitted",
			Detail: "the resource pattern is " + pattern +
				"; upstream carries each part of the name as a spec field",
		}
		if v == "location" {
			// Location gets the more specific advice, because upstream is split
			// 8 to 7 on whether a nested resource repeats its parent's location.
			// There is no convention to copy, so the reader has to decide.
			item.Reason = "location-omitted-nested-parent"
			item.Detail = "parent is " + pattern +
				"; location is implied by the parent, add it only if the API needs it stated"
		}
		out = append(out, item)
	}
	return out
}

// lowerCamel converts a pattern placeholder to the JSON name upstream uses:
// "collection_group" -> "collectionGroup".
//
// This is not the type generator's field-name casing, which also applies the
// acronym table. Pattern placeholders are plain words, such as collection,
// tenant and data_store, and reusing the acronym rules would tie this to a
// setting each service opts into separately.
func lowerCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, "")
}

// lastSegment returns the final dot-separated component of a proto full name.
func lastSegment(fullName string) string {
	if i := strings.LastIndex(fullName, "."); i >= 0 {
		return fullName[i+1:]
	}
	return fullName
}
