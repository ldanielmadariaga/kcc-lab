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
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/codegen"
	"github.com/GoogleCloudPlatform/k8s-config-connector/dev/tools/controllerbuilder/pkg/protoapi"
)

// sharedRefsPackage holds the reference types services use in common, relative
// to the repo root.
const sharedRefsPackage = "apis/refs/v1beta1"

// parentRef renders the Spec field naming a resource's direct parent and the
// queue entry that goes with it, as referenceTo describes. Both are empty where
// the parent is a location or the root of the name, which the location field
// and rootRef already carry.
func (a *APIScaffolder) parentRef(pattern string) (field string, item *JudgementItem) {
	collection, placeholder := protoapi.ParentPair(pattern)
	switch collection {
	case "", "locations", "regions", "zones", "global":
		return "", nil
	}
	path := parentPath(pattern, collection, placeholder)
	if strings.Count(path, "/") == 1 {
		return "", nil
	}

	// The collection segment names the field, not the placeholder beside it. See
	// ParentPair. Reading the collection is right 32 times out of 46, against 28
	// for the placeholder.
	return a.referenceTo("parent", collection, path, pattern)
}

// rootRef renders the Spec field naming the root of a resource's name, and the
// queue entry that goes with it when the root is one KCC has no fixed field
// for.
//
// A project, and a resource with no pattern, get projectRef; an organization
// or folder gets organizationRef or folderRef. Any other root, such as
// properties/{property} in Analytics, is looked up the way parentRef looks up
// a parent. A pattern that is only the root itself, such as
// billingAccounts/{billing_account}, names no root to point at, so it gets no
// field.
//
// It reads the first segment rather than protoapi.ParentStyle, which calls
// "organizations/{organization}/locations/{location}/..." ParentOther.
func (a *APIScaffolder) rootRef(pattern string) (field string, item *JudgementItem) {
	segs := strings.Split(pattern, "/")
	switch {
	case pattern == "":
		return fixedRootField("ProjectRef", "projectRef", "project"), nil
	case len(segs) < 3:
		return "", nil
	}
	switch segs[0] {
	case "projects":
		return fixedRootField("ProjectRef", "projectRef", "project"), nil
	case "organizations":
		return fixedRootField("OrganizationRef", "organizationRef", "organization"), nil
	case "folders":
		return fixedRootField("FolderRef", "folderRef", "folder"), nil
	}
	return a.referenceTo("root", segs[0], strings.Join(segs[:2], "/"), pattern)
}

// locationRef renders the Spec field naming the location in a resource's
// name, and returns a queue entry when that field is a guess or needs
// follow-up.
//
// The field is written only when the name has a location, region or zone
// segment with a placeholder, and is named after that segment. The first such
// segment counts, so a zones collection below a location, as in Dataplex, is
// not taken for the location. The field is required when that segment is the
// resource's direct parent. Higher up, as for
// .../locations/{location}/clusters/{cluster}/instances/{instance}, the parent
// already implies it, and upstream is split 17 to 30 on restating it, so it is
// optional and queued. With no pattern the parent shape is unknown, and the
// field is the required location the template has always written, with a
// queue entry asking whether the resource is regional.
//
// The type stays string because template/apis/identity.go reads Spec.Location
// as one.
func (a *APIScaffolder) locationRef(pattern string) (field string, item *JudgementItem) {
	if pattern == "" {
		return renderLocationField("location", true, false), &JudgementItem{
			FieldPath: ".spec.location",
			Reason:    "location-parent-unknown",
			Detail: "the proto declares no google.api.resource, so the parent shape is unknown; " +
				"drop location if the resource is not regional",
		}
	}

	segs := strings.Split(pattern, "/")
	collection, placeholder := "", ""
	for i := 0; i+1 < len(segs); i++ {
		switch segs[i] {
		case "locations", "regions", "zones":
			if strings.HasPrefix(segs[i+1], "{") {
				collection, placeholder = segs[i], strings.Trim(segs[i+1], "{}")
			}
		}
		if collection != "" {
			// The resource's own ID, as for projects/{project}/locations/{location},
			// is resourceID, not a location field.
			if i+2 == len(segs) {
				return "", nil
			}
			break
		}
	}
	if collection == "" {
		// The name has no location, or only a fixed one such as locations/global.
		return "", nil
	}

	name := codegen.Singular(collection)
	parentCollection, parentPlaceholder := protoapi.ParentPair(pattern)
	direct := parentCollection == collection && parentPlaceholder == placeholder
	field = renderLocationField(name, direct, !direct)

	switch {
	case !direct:
		item = &JudgementItem{
			FieldPath: ".spec." + name,
			Reason:    "location-guessed",
			Detail: fmt.Sprintf("the %s in %s belongs to an ancestor, which the parent "+
				"already names; emitted as optional, so keep it only if the API needs it stated", name, pattern),
		}
	case name != "location":
		item = &JudgementItem{
			FieldPath: ".spec." + name,
			Reason:    "location-renamed",
			Detail: fmt.Sprintf("named %s after %s; the identity template reads spec.location, "+
				"so the generated identity needs the same rename", name, pattern),
		}
	}
	return field, item
}

// renderLocationField renders a string field called name, marked +kcc:guess
// when guess is set. A required "location" renders exactly as the types
// template wrote it before, so project/location resources scaffold unchanged.
func renderLocationField(name string, required, guess bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\t// The %s of this resource.\n", name)
	if guess {
		b.WriteString("\t// +kcc:guess\n")
	}
	tag := name
	if !required {
		tag += ",omitempty"
	}
	fmt.Fprintf(&b, "\t%s string `json:%q`", exportedName(name), tag)
	return b.String()
}

// fixedRootField renders a required reference to one of the roots every KCC
// resource can point at, using the shared type of that name.
func fixedRootField(refType, jsonName, noun string) string {
	return fmt.Sprintf("\t// The %s that this resource belongs to.\n"+
		"\t%s *refsv1beta1.%s `json:%q`", noun, refType, refType, jsonName)
}

// referenceTo renders a Spec field pointing at the resource at path, whose
// collection segment is collection, and the queue entry for it. role says
// which ancestor this is, "parent" or "root", in the entry.
//
// The field is written only where exactly one reference type matches. Where
// none does, or several do, the entry names the path and no field is written,
// because a wrong reference is harder to catch in review than an absent one.
func (a *APIScaffolder) referenceTo(role, collection, path, pattern string) (field string, item *JudgementItem) {
	name := codegen.Singular(collection)
	candidates := parentRefTypes(a.repoRoot(), filepath.Join(a.BaseDir, a.GoPackage), name)
	if len(candidates) != 1 {
		detail := fmt.Sprintf("the %s is %s, and no %sRef type exists to point at; "+
			"add the %s resource first, then model this as a reference", role, path, exportedName(name), role)
		if len(candidates) > 1 {
			detail = fmt.Sprintf("the %s is %s, and several reference types match it (%s); "+
				"pick one and add the field by hand", role, path, strings.Join(sortedNames(candidates), ", "))
		}
		return "", &JudgementItem{
			Reason: role + "-ref-not-modelled",
			Detail: detail,
		}
	}

	typeName := sortedNames(candidates)[0]
	goType := typeName
	if qualifier := candidates[typeName]; qualifier != "" {
		goType = qualifier + "." + typeName
	}

	field = fmt.Sprintf("\t// A reference to the %s this resource belongs to.\n"+
		"\t// +kcc:guess\n"+
		"\t%sRef *%s `json:%q`", path, exportedName(name), goType, name+"Ref,omitempty")
	return field, &JudgementItem{
		FieldPath: ".spec." + name + "Ref",
		Reason:    role + "-ref-guessed",
		Detail: fmt.Sprintf("emitted as a reference to %s, read from the %s segment of %s; "+
			"the field name comes from the pattern rather than the proto, so confirm both the name and the target",
			typeName, collection, pattern),
	}
}

// parentRefTypes returns the reference types that could name a parent whose
// collection singularises to name, as a map of type name to import qualifier.
//
// Services spell a reference type either as a bare noun, ClusterRef in alloydb,
// or with the Kind's own prefix, as in DNSManagedZoneRef, so the match is on
// the suffix. The service's own package wins over the shared one, and
// unexported types are skipped: apis/kms/v1beta1 declares kmsCryptoKeyRef
// beside the real KMSCryptoKeyRef, which would make a single match look
// ambiguous.
func parentRefTypes(repoRoot, serviceDir, name string) map[string]string {
	want := strings.ToLower(name) + "ref"
	re := regexp.MustCompile(`(?m)^type ([A-Z]\w*Ref) struct`)

	scan := func(dir, qualifier string) map[string]string {
		found := map[string]string{}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return found
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
				continue
			}
			body, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			for _, m := range re.FindAllStringSubmatch(string(body), -1) {
				if strings.HasSuffix(strings.ToLower(m[1]), want) {
					found[m[1]] = qualifier
				}
			}
		}
		return found
	}

	if found := scan(serviceDir, ""); len(found) > 0 {
		return found
	}
	return scan(filepath.Join(repoRoot, sharedRefsPackage), "refsv1beta1")
}

// repoRoot is the tree BaseDir sits in, so the shared refs package can be
// found. BaseDir is the output API directory, normally <repo>/apis.
func (a *APIScaffolder) repoRoot() string {
	if filepath.Base(a.BaseDir) == "apis" {
		return filepath.Dir(a.BaseDir)
	}
	return a.BaseDir
}

// parentPath is the pattern truncated at the parent, which is the whole path a
// reference to it carries in its External field, per AIP-122.
func parentPath(pattern, collection, placeholder string) string {
	segs := strings.Split(pattern, "/")
	for i := 0; i+1 < len(segs); i++ {
		if segs[i] == collection && segs[i+1] == "{"+placeholder+"}" {
			return strings.Join(segs[:i+2], "/")
		}
	}
	return pattern
}

func sortedNames(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func exportedName(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
