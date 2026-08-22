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

// Command queue-ref-hints seeds the judgement queue with the reference fields a
// generated resource probably needs.
//
// Why this runs after generation rather than inside it: the generator holds the
// proto and could consult google.api.resource_reference directly, which would be
// more precise. But dev/tools/controllerbuilder is a separate Go module with no
// dependency on the main one, so it cannot use the rules in tests/apichecks/refs
// -- and copying them would leave two tuned implementations to drift apart. The
// rules already work on CRDs, and descriptions survive into the CRD, so applying
// them to freshly generated CRDs costs nothing and keeps one implementation.
//
// Measured on the 239-resource run before this existed: the queue named 11 of
// the 168 reference fields those resources actually needed, so an agent working
// from it saw almost none of the work.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/crd/crdloader"
	"github.com/GoogleCloudPlatform/k8s-config-connector/tests/apichecks/refs"
)

func main() {
	apisDir := flag.String("apis-dir", "apis", "directory holding the per-service API packages")
	dryRun := flag.Bool("dry-run", false, "report what would be written without writing it")
	listHints := flag.Bool("list", false, "with -dry-run, print each hint line instead of per-service counts")
	flag.Parse()

	if err := run(*apisDir, *dryRun, *listHints); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(apisDir string, dryRun, listHints bool) error {
	crds, err := crdloader.LoadAllCRDs()
	if err != nil {
		return fmt.Errorf("loading CRDs: %w", err)
	}

	// service -> entry lines
	hints := map[string][]string{}
	for _, crd := range crds {
		service := serviceFromGroup(crd.Spec.Group)
		if service == "" {
			continue
		}
		// Only seed a Kind that is already in the queue. The queue is written by
		// generate-types --prepopulate-spec, so an entry is what marks a resource
		// as bulk-generated; seeding others would add entries to resources nobody
		// is generating, and every entry suppresses [refs] for the Kind it names.
		//
		// The gate is per Kind rather than per service because a service is not
		// uniformly bulk-generated. compute holds both freshly generated kinds and
		// ComputeInstance, which upstream has maintained by hand for years. A
		// service-level gate let 30 such resources through -- ComputeInstance,
		// ComputeInstanceTemplate, CloudBuildTrigger, TPUNode among them -- and
		// queueing one of those does more than add a note: a queued resource
		// contributes no [refs] findings, so its existing missingrefs.txt entries
		// read as fixed and get pruned, then return as fresh violations the moment
		// it graduates.
		queued, err := queuedKinds(queuePath(apisDir, service))
		if err != nil {
			continue
		}
		if !queued[crd.Spec.Names.Kind] {
			continue
		}
		for _, version := range crd.Spec.Versions {
			if version.Schema == nil || version.Schema.OpenAPIV3Schema == nil {
				continue
			}
			for _, line := range hintsFor(crd, version) {
				hints[service] = append(hints[service], line)
			}
		}
	}

	services := make([]string, 0, len(hints))
	for s := range hints {
		services = append(services, s)
	}
	sort.Strings(services)

	total := 0
	for _, service := range services {
		lines := dedupe(hints[service])
		total += len(lines)
		if dryRun {
			if listHints {
				for _, l := range lines {
					fmt.Println(l)
				}
			} else {
				fmt.Printf("%s: %d hint(s)\n", service, len(lines))
			}
			continue
		}
		if err := appendHints(queuePath(apisDir, service), lines); err != nil {
			return fmt.Errorf("%s: %w", service, err)
		}
	}
	fmt.Printf("%d reference hint(s) across %d service(s)\n", total, len(services))
	return nil
}

// hintsFor applies the shared rules to one CRD version.
//
// Unlike TestMissingRefs this deliberately does NOT skip resources that already
// have queue entries. Those are precisely the untriaged resources the hints are
// for; the check skips them so a half-finished resource can still merge.
// queuedKinds reports which Kinds a service's queue file already names.
func queuedKinds(path string) (map[string]bool, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := map[string]bool{}
	for _, line := range strings.Split(string(body), "\n") {
		if !strings.HasPrefix(line, "kind=") {
			continue
		}
		rest := line[len("kind="):]
		if i := strings.IndexAny(rest, " \t"); i >= 0 {
			rest = rest[:i]
		}
		if rest != "" {
			out[rest] = true
		}
	}
	return out, nil
}

func hintsFor(crd apiextensions.CustomResourceDefinition, version apiextensions.CustomResourceDefinitionVersion) []string {
	var out []string
	visitProps(version.Schema.OpenAPIV3Schema, "", func(path string, props *apiextensions.JSONSchemaProps) {
		if strings.HasPrefix(path, ".status.") || refs.IsReferenceFieldPath(path) {
			return
		}
		verdict, _ := refs.Classify(path, props.Description)
		reason := "possible-reference-by-description"
		if verdict != refs.IsReference {
			// The description rules miss a whole class: a Secret Manager version
			// or a VPC is often named plainly with no resource-name template to
			// go on. The name rules cover those, and are applied here rather
			// than inside Classify because Classify also feeds a ratchet.
			target, ok := refs.MatchName(path)
			if !ok {
				return
			}
			reason = "possible-reference-by-name (" + target + ")"
		}
		out = append(out, fmt.Sprintf(
			"kind=%s group=%s: field %q reason=%s",
			crd.Spec.Names.Kind, crd.Spec.Group, path, reason))
	})
	return out
}

func visitProps(props *apiextensions.JSONSchemaProps, path string, fn func(string, *apiextensions.JSONSchemaProps)) {
	if props == nil {
		return
	}
	// A reference-shaped object is already a reference, whatever it is called,
	// so neither it nor its external/name/namespace children are candidates.
	// Pruning the subtree here rather than filtering by name catches the ones
	// upstream did not suffix with Ref, such as
	// NetworkManagementConnectivityTest's destination.sqlInstance.
	if path != "" && refs.HasReferenceShape(propertyNames(props)) {
		return
	}
	if path != "" {
		fn(path, props)
	}
	for name, child := range props.Properties {
		visitProps(&child, path+"."+name, fn)
	}
	if props.Items != nil && props.Items.Schema != nil {
		visitProps(props.Items.Schema, path+"[]", fn)
	}
}

func propertyNames(props *apiextensions.JSONSchemaProps) []string {
	out := make([]string, 0, len(props.Properties))
	for name := range props.Properties {
		out = append(out, name)
	}
	return out
}

func serviceFromGroup(group string) string {
	name, _, found := strings.Cut(group, ".")
	if !found {
		return ""
	}
	return name
}

func queuePath(apisDir, service string) string {
	return filepath.Join(apisDir, service, "needs_judgement_call.txt")
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// appendHints adds only lines the file does not already have, so the command is
// safe to re-run after each generation.
func appendHints(path string, lines []string) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	body := string(existing)
	var add []string
	for _, l := range lines {
		if !strings.Contains(body, l) {
			add = append(add, l)
		}
	}
	if len(add) == 0 {
		return nil
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	body += strings.Join(add, "\n") + "\n"
	return os.WriteFile(path, []byte(body), 0644)
}
