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
		if base := strings.TrimSuffix(fullName, "."+lastSegment(fullName)); base != "" {
			for _, pkg := range strings.Split(a.PackageProtoTag, ",") {
				candidate := pkg + "." + lastSegment(fullName)
				if alt, altErr := a.Proto.Files().FindDescriptorByName(protoreflect.FullName(candidate)); altErr == nil {
					d = alt
					break
				}
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
	// Default for every path, including the stub with no proto at all: nearly
	// every resource is project-rooted, and AddTypeFile overrides this when the
	// pattern says otherwise.
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
// When prepopulated is non-nil the Spec is filled in from the proto rather than
// left as a three-field stub. The descriptor is passed in rather than held on the
// scaffolder because only the caller knows which proto message a resource maps to.
func (a *APIScaffolder) AddTypeFile(resource options.Resource, prepopulated *PrepopulateResult) error {
	typeFilePath := a.PathToTypeFile(resource)
	cArgs := a.buildAPIArgs(&resource)
	if prepopulated != nil {
		// Name every part of the resource's name that the Spec does not carry.
		//
		// The template emits projectRef and resourceID always, and location only
		// for a projects/locations parent. Every other variable segment is dropped
		// in silence, and a user cannot name the resource without it. Measured on
		// the 189-resource bulk run, the omissions were spec.collection,
		// spec.collectionGroup, spec.tenant, spec.parent and spec.region -- about
		// fifteen fields upstream has and we did not mention anywhere.
		//
		// Omitting rather than emitting is still the right default. Adding a field
		// later is easy; removing a required one is a breaking change. The point is
		// only that the choice gets written down instead of made silently.
		// Emit every part of the resource's name that the Spec can carry, and
		// queue every one of them.
		//
		// Each field emitted here carries a +kcc:guess marker, because none of it
		// is justified by the proto: a segment's field name is taken from the
		// pattern's placeholder, and a ref's target type is assumed from the
		// collection segment. Anything marked as a guess belongs in the judgement
		// queue, typed refs included -- "very sure" is not a state the generator
		// can be in about a target it inferred from a plural noun.
		//
		// An earlier version recorded emitted segments and suppressed their queue
		// entries. That traded detection for an unflagged guess: BigtableCluster
		// got "Instance *string" where upstream has spec.instanceRef, and nothing
		// said so. The compiler found 32 fields missing while the queue named 2.
		known := refTypesInPackage(a.repoRoot(), filepath.Join(a.BaseDir, a.GoPackage))
		var parentRefs bytes.Buffer
		var guesses []JudgementItem
		segments := parentSegments(cArgs.ResourcePattern)
		for i, seg := range segments {
			collection, variable := seg[0], seg[1]
			// The root segment becomes projectRef / organizationRef / folderRef,
			// which the template renders.
			if i == 0 {
				continue
			}
			if field, ok := locationFieldNames[collection]; ok {
				// Required only for the projects/locations shape the template used
				// to render, so that case stays byte-identical. Everywhere else the
				// parent already fixes a location, so requiring it would be new.
				required := cArgs.ParentStyle == string(protoapi.ParentProjectLocation)
				parentRefs.WriteString(locationField(field, cArgs.ResourcePattern, required))
				guesses = append(guesses, JudgementItem{
					FieldPath: ".spec." + field,
					Reason:    "parent-location-guessed",
					Detail: "emitted from the location segment of " + cArgs.ResourcePattern +
						"; confirm the resource is regional and that this is the name for it",
				})
				continue
			}
			name := lowerCamel(variable)
			goType := strings.ToUpper(name[:1]) + name[1:] + "Ref"
			qualifier, ok := known[goType]
			reason, detail := "parent-ref-guessed",
				"emitted as a reference to "+goType+", assumed from the collection segment of "+
					cArgs.ResourcePattern+"; confirm the target type"
			if !ok {
				// No ref type anywhere, so a plain string. Still better than
				// nothing, and upstream may well want a reference here.
				goType = ""
				reason, detail = "parent-segment-guessed",
					"emitted as a plain string from the pattern "+cArgs.ResourcePattern+
						"; upstream may model this as a reference instead"
			}
			parentRefs.WriteString(parentRefField(name, goType, qualifier, cArgs.ResourcePattern))
			suffix := ""
			if goType != "" {
				suffix = "Ref"
			}
			guesses = append(guesses, JudgementItem{
				FieldPath: ".spec." + name + suffix,
				Reason:    reason,
				Detail:    detail,
			})
		}
		prepopulated.Judgement = append(prepopulated.Judgement, guesses...)

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

		// Only the nothing-was-emitted case is left for parentSegmentJudgement:
		// a proto with no google.api.resource, where there is no pattern to walk
		// and so nothing to emit or to name precisely.
		if len(segments) == 0 {
			prepopulated.Judgement = append(prepopulated.Judgement,
				parentSegmentJudgement(cArgs.ResourcePattern, cArgs.ParentStyle)...)
		}

		cArgs.SpecFields = prepopulated.SpecFields
		cArgs.ObservedStateFields = prepopulated.ObservedStateFields
		// Computed from both bodies here rather than taken from the result,
		// because a special-cased type can appear in either one.
		cArgs.ExtraImports = ExtraImportsFor(prepopulated.SpecFields, prepopulated.ObservedStateFields)
	}
	return scaffoldTypeFile(typeFilePath, cArgs)
}

// parentSegmentJudgement reports the parts of a resource's name that the Spec
// does not carry.
//
// GrafeasNote is the worked example for the unknown case: unannotated, project
// parented, and upstream gives it no location either.
// sharedRefsPackage holds the ref types every service can point at --
// OrganizationRef, FolderRef, ProjectRef and about thirty more.
const sharedRefsPackage = "apis/refs/v1beta1"

// refTypesInPackage lists the <X>Ref types available to the target package,
// mapping each to the qualifier it must be written with.
//
// Two directories, because a ref type can live in either. scaffoldRefsFile
// writes one per resource into the service package, so a resource whose parent
// has been generated has a local type to point at. The shared package holds the
// ones every service needs. Missing the shared package was the single biggest
// hole this scanner had: OrganizationRef accounts for 14 of the compile errors
// and has existed all along.
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

// parentSegments returns the (collection, placeholder) pairs of a pattern's
// parent, in order. The resource's own trailing segment is dropped.
//
//	projects/{project}/regions/{location}/jobs/{job}
//	  -> [{projects project} {regions location}]
//
// The collection matters as well as the placeholder because they disagree:
// Dataproc's placeholder is "location" while the collection is "regions", and
// upstream names that field "region".
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

// locationFieldNames maps a location-ish collection to the field name upstream
// uses for it. Anything not here is not a location.
var locationFieldNames = map[string]string{
	"locations": "location",
	"regions":   "region",
	"zones":     "zone",
}

// parentRefField renders a spec field pointing at part of the resource's
// parent.
//
// The types template used to emit projectRef and resourceID and nothing else,
// so a resource nested under another resource -- a Bigtable Cluster under an
// Instance, a Firestore Field under a Database -- had no way to say which
// parent it belongs to. Upstream carries these for exactly that.
//
// goType empty means no ref type resolved, and the segment is emitted as a
// plain string: the compiler asks for Spec.Tenant and Spec.KeyRing by those
// names, not as refs, so that is how upstream models them.
//
// The first line is prose and becomes the CRD description, so it says what the
// field is for. The rest is a +kcc: marker, which controller-gen strips from
// the description -- a reviewer reading the type sees the guess, a user running
// kubectl explain is not told our TODO.
func parentRefField(segment, goType, qualifier, pattern string) string {
	name := strings.ToUpper(segment[:1]) + segment[1:]
	if goType == "" {
		return fmt.Sprintf(`
	// The %s that this resource belongs to.
	// +kcc:guess=parent-segment pattern=%s
	%s *string `+"`"+`json:"%s,omitempty"`+"`"+`
`, name, pattern, name, segment)
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
		// No google.api.resource, so there is no pattern to walk. Say that much,
		// naming location because a regional resource is the case that bites.
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
			// Keep the more specific advice for the case that has it: upstream is
			// split 8 to 7 on whether a nested resource repeats its parent's
			// location, so there is no convention to copy.
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
// Deliberately not the type generator's field-name casing, which also applies
// the acronym table. Pattern placeholders are plain words -- collection,
// tenant, data_store -- and borrowing the acronym rules here would couple this
// to a setting that is opt-in per service.
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

// lastSegment returns the final dot-separated component of a proto full name.
func lastSegment(fullName string) string {
	if i := strings.LastIndex(fullName, "."); i >= 0 {
		return fullName[i+1:]
	}
	return fullName
}
