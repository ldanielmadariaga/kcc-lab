// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Conformance checks for resources produced by the experimental bulk greenfield
// workflow. Scope comes from testdata/greenfield_bulk.txt; everything not listed
// there is out of scope, because it predates this bar.
//
// SANDBOX-ONLY. These checks do not exist upstream.

package lint

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/crd/crdloader"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/test"
	"github.com/GoogleCloudPlatform/k8s-config-connector/tests/apichecks/greenfield"
)

const (
	bulkManifestPath = "testdata/greenfield_bulk.txt"
	repoRoot         = "../.."
)

func loadBulkManifest(t *testing.T) *greenfield.Manifest {
	t.Helper()
	m, err := greenfield.Load(bulkManifestPath)
	if err != nil {
		t.Fatalf("loading bulk manifest: %v", err)
	}
	return m
}

// TestGreenfieldBulkManifestIsResolvable checks that every entry in the manifest
// corresponds to a real CRD and at least one on-disk file. A typo here would
// silently narrow the scope of every other check in this file, so it is checked
// first and explicitly.
func TestGreenfieldBulkManifestIsResolvable(t *testing.T) {
	t.Parallel()

	m := loadBulkManifest(t)
	crds, err := crdloader.LoadAllCRDs()
	if err != nil {
		t.Fatalf("loading CRDs: %v", err)
	}

	for _, r := range m.Resources() {
		if _, ok := greenfield.FindCRD(crds, r); !ok {
			t.Errorf("FAIL: %s/%s is in %s but no matching CRD was found. Check the group and Kind spelling.",
				r.Group, r.Kind, bulkManifestPath)
			continue
		}
		files, err := m.FilesFor(repoRoot, r)
		if err != nil {
			t.Errorf("FAIL: resolving files for %s: %v", r.Kind, err)
			continue
		}
		if len(files) == 0 {
			t.Errorf("FAIL: %s/%s is in %s but no per-resource files were found. Expected files named %s_*.go.",
				r.Group, r.Kind, bulkManifestPath, strings.ToLower(r.Kind))
		}
	}
}

// TestGreenfieldBulkTypesConformance checks the hand-edited per-resource Go
// files against the bulk generation bar.
//
// These are all syntax-level facts, so the files are parsed with go/parser
// rather than run through a full analysis pass.
func TestGreenfieldBulkTypesConformance(t *testing.T) {
	t.Parallel()

	m := loadBulkManifest(t)

	for _, r := range m.Resources() {
		files, err := m.FilesFor(repoRoot, r)
		if err != nil {
			t.Fatalf("resolving files for %s: %v", r.Kind, err)
		}
		for _, path := range files {
			problems, err := greenfield.CheckGoFile(path)
			if err != nil {
				t.Errorf("FAIL: %s: %v", path, err)
				continue
			}
			for _, problem := range problems {
				t.Errorf("FAIL: %s: %s", path, problem)
			}
		}
	}
}

// TestGreenfieldBulkCRDConformance checks the generated CRDs for the bulk
// resources. Greenfield resources are v1alpha1 only.
func TestGreenfieldBulkCRDConformance(t *testing.T) {
	t.Parallel()

	m := loadBulkManifest(t)
	crds, err := crdloader.LoadAllCRDs()
	if err != nil {
		t.Fatalf("loading CRDs: %v", err)
	}

	// Required CRD labels. managed-by-kcc and system come from the base types
	// skill; stability-level=alpha from the greenfield skill.
	requiredLabels := map[string]string{
		"cnrm.cloud.google.com/managed-by-kcc":  "true",
		"cnrm.cloud.google.com/system":          "true",
		"cnrm.cloud.google.com/stability-level": "alpha",
	}

	for _, r := range m.Resources() {
		crd, ok := greenfield.FindCRD(crds, r)
		if !ok {
			continue // reported by TestGreenfieldBulkManifestIsResolvable
		}
		for _, version := range crd.Spec.Versions {
			if version.Name != "v1alpha1" {
				t.Errorf("FAIL: crd=%s has version %q; greenfield resources must be v1alpha1 only",
					crd.Name, version.Name)
			}
		}
		for key, want := range requiredLabels {
			got, present := crd.Labels[key]
			if !present {
				t.Errorf("FAIL: crd=%s is missing label %q (add `// +kubebuilder:metadata:labels=\"%s=%s\"` to the Kind)",
					crd.Name, key, key, want)
				continue
			}
			if got != want {
				t.Errorf("FAIL: crd=%s has label %s=%q, want %q", crd.Name, key, got, want)
			}
		}
	}
}

// TestGreenfieldNoNewDeprecatedRefs guards the deprecated apis/refs/v1beta1
// directory. New Ref types belong in apis/<service>/<version>/<kind>_reference.go.
//
// Scoped by a committed baseline of the files that exist today rather than by
// the bulk manifest: the rule is about where a file is placed, so it must catch
// additions regardless of which resource prompted them.
func TestGreenfieldNoNewDeprecatedRefs(t *testing.T) {
	t.Parallel()

	const dir = "../../apis/refs/v1beta1"
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}
	var got []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		got = append(got, e.Name())
	}
	sort.Strings(got)

	test.CompareRatchetFile(t, "testdata/exceptions/deprecated_refs_v1beta1.txt", strings.Join(got, "\n"))
}

// TestGreenfieldGenerateScriptCopyright applies the 2026 copyright rule to the
// service generate.sh, which the .go-only check missed.
func TestGreenfieldGenerateScriptCopyright(t *testing.T) {
	t.Parallel()

	m := loadBulkManifest(t)
	seen := map[string]bool{}
	for _, r := range m.Resources() {
		path := greenfield.GenerateScriptFor(repoRoot, r)
		if seen[path] {
			continue // one generate.sh per service, not per resource
		}
		seen[path] = true

		problems, err := greenfield.CheckShellFile(path)
		if err != nil {
			t.Errorf("FAIL: %s: %v", path, err)
			continue
		}
		for _, p := range problems {
			t.Errorf("FAIL: %s: %s", path, p)
		}
	}
}

// TestGreenfieldDroppedFields reports proto fields that have no representation
// at all in a bulk resource's KRM types.
//
// This is a different concern from alpha-missingfields.txt. That file records
// fields that exist in the CRD but no test fixture sets. A dropped field is not
// in the CRD at all, so nothing else in the repo can see it: it can be commented
// out, or simply never written, and disappear silently.
//
// The baseline is a ratchet, not a golden file: new drops fail even under
// WRITE_GOLDEN_OUTPUT, so a field cannot be regenerated out of existence. Fixed
// drops are pruned, so the list can only shrink.
func TestGreenfieldDroppedFields(t *testing.T) {
	t.Parallel()

	m := loadBulkManifest(t)
	crds, err := crdloader.LoadAllCRDs()
	if err != nil {
		t.Fatalf("loading CRDs: %v", err)
	}

	// The baseline carries a "reason=" suffix per entry, but detection only knows
	// the entry itself. Re-emit the baseline's line verbatim when the drop is
	// still present, so an already-justified drop compares equal. A newly
	// detected drop has no baseline line, so it appears bare, is not in the
	// baseline, and fails - which is the point.
	baselineByKey := map[string]string{}
	baselineLines, err := greenfield.BaselineLines("testdata/exceptions/greenfield_dropped_fields.txt")
	if err != nil {
		t.Fatalf("reading greenfield_dropped_fields.txt: %v", err)
	}
	for _, line := range baselineLines {
		key, _, _ := strings.Cut(line, " reason=")
		baselineByKey[strings.TrimSpace(key)] = line
	}

	var got []string
	for _, r := range m.Resources() {
		crd, ok := greenfield.FindCRD(crds, r)
		if !ok {
			continue // reported by TestGreenfieldBulkManifestIsResolvable
		}
		dropped, err := greenfield.DroppedFields(greenfield.MapperPath(repoRoot, r), r.Kind)
		if err != nil {
			t.Fatalf("finding dropped fields for %s: %v", r.Kind, err)
		}
		for _, field := range dropped {
			key := fmt.Sprintf("[dropped] crd=%s: field %q", crd.Name, field)
			if line, ok := baselineByKey[key]; ok {
				got = append(got, line)
			} else {
				got = append(got, key)
			}
		}
	}

	// Nested and shared types, keyed by service rather than by Kind: a nested
	// type belongs to every resource in its service, so attributing a drop to one
	// Kind would be a fiction. Collected per service so a type shared by two
	// manifest Kinds is reported once.
	kindsByService := map[string][]string{}
	for _, r := range m.Resources() {
		kindsByService[r.Service()] = append(kindsByService[r.Service()], r.Kind)
	}
	for _, r := range m.Resources() {
		svc := r.Service()
		kinds, pending := kindsByService[svc]
		if !pending {
			continue // already processed this service
		}
		delete(kindsByService, svc)

		nested, err := greenfield.NestedDroppedFields(greenfield.MapperPath(repoRoot, r), kinds)
		if err != nil {
			t.Fatalf("finding nested dropped fields for %s: %v", svc, err)
		}
		for _, d := range nested {
			key := fmt.Sprintf("[dropped] service=%s version=v1alpha1 type=%s: field %q", svc, d.Type, d.Field)
			if line, ok := baselineByKey[key]; ok {
				got = append(got, line)
			} else {
				got = append(got, key)
			}
		}
	}

	sort.Strings(got)

	test.CompareRatchetFile(t, "testdata/exceptions/greenfield_dropped_fields.txt", strings.Join(got, "\n"))
}

// TestGreenfieldDroppedFieldsHaveReasons requires every recorded drop to carry a
// reason. Recording a drop is a decision, and a decision without a stated reason
// is indistinguishable from an oversight.
func TestGreenfieldDroppedFieldsHaveReasons(t *testing.T) {
	t.Parallel()

	lines, err := greenfield.BaselineLines("testdata/exceptions/greenfield_dropped_fields.txt")
	if err != nil {
		t.Fatalf("reading greenfield_dropped_fields.txt: %v", err)
	}
	for _, line := range lines {
		if !strings.Contains(line, "reason=") {
			t.Errorf("FAIL: dropped-field entry has no reason=: %q\n"+
				"Either implement the field, or say why it is intentionally absent.", line)
		}
	}
}

// TestGreenfieldBulkFieldCoverage fails when a bulk resource still has fields
// recorded in alpha-missingfields.txt, i.e. fields no test fixture exercises.
//
// This is a local-only nudge, not a CI gate: Step 1 of the bulk workflow ships
// types and CRDs with no fixtures, so every new resource legitimately fails this
// until a later phase adds them. Run it deliberately with:
//
//	GREENFIELD_STRICT=1 go test ./tests/apichecks/... -run TestGreenfieldBulkFieldCoverage
//
// Opt-in rather than skip-when-CI-detected: if the environment check were ever
// wrong, opt-in simply does not run, instead of breaking CI.
//
// Fields that genuinely cannot be covered belong in
// testdata/exceptions/greenfield_fields_accepted.txt with a reason.
func TestGreenfieldBulkFieldCoverage(t *testing.T) {
	t.Parallel()

	if os.Getenv("GREENFIELD_STRICT") == "" {
		t.Skip("set GREENFIELD_STRICT=1 to run; this is a local-only check, not a CI gate")
	}

	m := loadBulkManifest(t)
	crds, err := crdloader.LoadAllCRDs()
	if err != nil {
		t.Fatalf("loading CRDs: %v", err)
	}

	missing, err := greenfield.BaselineLines("testdata/exceptions/alpha-missingfields.txt")
	if err != nil {
		t.Fatalf("reading alpha-missingfields.txt: %v", err)
	}
	accepted, err := greenfield.BaselineLines("testdata/exceptions/greenfield_fields_accepted.txt")
	if err != nil {
		t.Fatalf("reading greenfield_fields_accepted.txt: %v", err)
	}

	// Index accepted fields by "crd=<name>|<field path>".
	acceptedFields := make(map[string]bool)
	for _, line := range accepted {
		crdName, fieldPath, ok := greenfield.ParseBaselineEntry(line)
		if !ok {
			t.Errorf("FAIL: unparseable entry in greenfield_fields_accepted.txt: %q", line)
			continue
		}
		if !strings.Contains(line, "reason=") {
			t.Errorf("FAIL: entry in greenfield_fields_accepted.txt has no reason=: %q", line)
			continue
		}
		acceptedFields[crdName+"|"+fieldPath] = true
	}

	for _, r := range m.Resources() {
		crd, ok := greenfield.FindCRD(crds, r)
		if !ok {
			continue
		}
		var uncovered []string
		for _, line := range missing {
			crdName, fieldPath, ok := greenfield.ParseBaselineEntry(line)
			if !ok || crdName != crd.Name {
				continue
			}
			if acceptedFields[crdName+"|"+fieldPath] {
				continue
			}
			uncovered = append(uncovered, fieldPath)
		}
		if len(uncovered) > 0 {
			t.Errorf("FAIL: crd=%s has %d field(s) not exercised by any test fixture:\n  %s\n\n"+
				"Add fixture coverage, or record genuinely uncoverable fields in "+
				"testdata/exceptions/greenfield_fields_accepted.txt with a reason=.",
				crd.Name, len(uncovered), strings.Join(uncovered, "\n  "))
		}
	}
}
