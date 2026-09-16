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

// Package refs decides whether a CRD spec field ought to be a KCC reference,
// from the two things a CRD keeps: the field's path and its description.
//
// TestMissingRefs in tests/apichecks applies these rules to every CRD, and
// missingrefs.txt, a ratchet, records what they find. The rules live in the
// generator module, which tests/apichecks already imports, so the generator
// can apply the same rules while it writes a resource.
package refs

import (
	"strings"
)

// Verdict is what Classify concluded about a field.
type Verdict int

const (
	// NotAReference means nothing suggests the field names another resource.
	NotAReference Verdict = iota
	// IsReference means the field names another resource and a KCC reference
	// can express it.
	IsReference
	// NotRepresentable means the field names another resource, but KCC cannot
	// model it as a reference today.
	NotRepresentable
)

// IsReferenceFieldPath reports whether fieldPath is already a KCC reference, or
// the external or name field inside one.
func IsReferenceFieldPath(fieldPath string) bool {
	for _, suffix := range []string{"Ref", "Refs[]", "Refs", "Ref.external", "Refs[].external", "Ref.name"} {
		if strings.HasSuffix(fieldPath, suffix) {
			return true
		}
	}
	return false
}

// Classify returns the verdict for a spec field from its path and description,
// and for NotRepresentable the reason. Callers skip status fields and paths
// IsReferenceFieldPath accepts before calling it.
func Classify(fieldPath, desc string) (Verdict, string) {
	// Pattern/matcher fields describe a SET of resources, never one, so
	// they are not references in either sense. e.g. dlp's
	// cloudStorageRegex.bucketNameRegex: "Regex to test the bucket name
	// against ... (marketing)\d{4}".
	if isPatternField(fieldPath, desc) {
		return NotAReference, ""
	}

	if reason := notRepresentableReason(fieldPath, desc); reason != "" {
		return NotRepresentable, reason
	}

	isRef := false

	// NOTE: (google.api.resource_reference) is deliberately NOT consulted
	// here. It is upstream ground truth and names the exact target type,
	// so it looks like the obvious signal - but matching it by field NAME
	// was tried and measured: 2,164 findings against 78 for the description
	// heuristics alone. A name like "network" is annotated in one service
	// and appears in hundreds of unrelated CRD fields elsewhere.
	//
	// A CRD does not record which proto field it came from, so the leaf
	// name is the only bridge available at this layer, and it is too weak.
	// Using the annotation safely needs the CRD field mapped to its
	// originating proto field via the +kcc:proto:field= markers the
	// generator emits into _types.go, then looking up the fully-qualified
	// proto path. That is its own change, and it would not reuse a
	// leaf-name list.

	// Signal: a resource-name path template in the description,
	// e.g. "should be of the form projects/{projectID}/locations/{location}/bars/{name}".
	if hasResourceNameTemplate(desc) {
		isRef = true
	}

	if strings.HasSuffix(fieldPath, "erviceAccount") {
		isRef = true
	}
	// TODO: how to detect KMS Key

	// Cloud Storage BUCKET references are expressible today:
	// StorageBucketIdentity.FromExternal accepts the bare "gs://<bucket>"
	// form (apis/storage/v1beta1/storagebucket_identity.go). Object paths
	// are handled above as not-representable.
	if hasToken(fieldPath, "bucket") && mentionsCloudStorage(desc) {
		isRef = true
	}

	if !isRef {
		return NotAReference, ""
	}

	// We don't require refs for zones or regions, nor for instanceTypes
	switch {
	case strings.HasSuffix(fieldPath, ".zone"),
		strings.HasSuffix(fieldPath, ".location"),
		strings.HasSuffix(fieldPath, ".machineType"),
		strings.HasSuffix(fieldPath, ".acceleratorType"):
		return NotAReference, ""
	}
	return IsReference, ""
}

// resourceNamePrefixes are the collection segments that start a GCP resource
// name. A description containing one of these followed by a placeholder is
// describing a resource name, i.e. a reference.
var resourceNamePrefixes = []string{"projects/", "locations/", "zones/", "regions/", "organizations/", "folders/"}

// hasResourceNameTemplate reports whether desc contains a resource-name path
// template.
//
// Placeholder syntax varies across APIs and the delimiter before the path
// varies too, which is how notification_channels was missed: its description
// reads "Must be of the format `projects/<project_id_or_number>/notificationChannels/<channel_id>`"
// - a backtick rather than a space, and <> rather than {}. Both forms count.
func hasResourceNameTemplate(desc string) bool {
	// A placeholder immediately after a collection segment is unambiguous: the
	// description is spelling out a resource name. Both brace and angle-bracket
	// syntaxes are used upstream.
	for _, prefix := range resourceNamePrefixes {
		if strings.Contains(desc, prefix+"{") || strings.Contains(desc, prefix+"<") {
			return true
		}
	}

	// Without a placeholder, only a space-delimited "projects/" is acted on. This
	// is the original rule and is kept deliberately narrow: widening it to other
	// delimiters or to "locations/" and friends matches ordinary prose, producing
	// findings on container.username and allowedLocations, which are plainly not
	// references.
	return strings.Contains(desc, " projects/")
}

// hasToken reports whether name contains tok as a whole word, splitting on
// camelCase boundaries, "." and "[]". Substring matching is wrong here:
// "security" contains "uri".
func hasToken(fieldPath string, tok string) bool {
	// Take the leaf segment, then split camelCase into lowercase words.
	leaf := fieldPath
	if i := strings.LastIndex(leaf, "."); i >= 0 {
		leaf = leaf[i+1:]
	}
	leaf = strings.TrimSuffix(leaf, "[]")
	leaf = strings.ReplaceAll(leaf, "_", " ")

	for _, w := range splitCamelWords(leaf) {
		if w == tok {
			return true
		}
	}
	return false
}

// splitCamelWords lowercases and splits identifiers on camelCase boundaries,
// keeping acronym runs intact: "outputURIPrefix" -> [output uri prefix].
// A naive split would yield [output u r i prefix] and never match "uri".
func splitCamelWords(s string) []string {
	isUpper := func(r rune) bool { return r >= 'A' && r <= 'Z' }
	isLower := func(r rune) bool { return r >= 'a' && r <= 'z' }

	runes := []rune(s)
	var words []string
	start := 0
	flush := func(end int) {
		w := strings.ToLower(strings.TrimSpace(string(runes[start:end])))
		if w != "" {
			words = append(words, w)
		}
		start = end
	}
	for i := 1; i < len(runes); i++ {
		prev, cur := runes[i-1], runes[i]
		switch {
		case isLower(prev) && isUpper(cur):
			// output|URI
			flush(i)
		case isUpper(prev) && isUpper(cur) && i+1 < len(runes) && isLower(runes[i+1]):
			// URI|Prefix
			flush(i)
		case cur == ' ' || prev == ' ':
			flush(i)
		}
	}
	flush(len(runes))
	return words
}

func mentionsCloudStorage(desc string) bool {
	return strings.Contains(desc, "Cloud Storage") || strings.Contains(desc, "gs://")
}

// isPatternField reports whether a field holds a matcher (regex/glob/filter)
// rather than the identity of one resource. These must not be reported in
// either bucket.
func isPatternField(fieldPath, desc string) bool {
	for _, tok := range []string{"regex", "pattern", "filter", "matcher", "matchers"} {
		if hasToken(fieldPath, tok) {
			return true
		}
	}
	trimmed := strings.TrimSpace(desc)
	if strings.HasPrefix(trimmed, "Regex") ||
		strings.HasPrefix(trimmed, "Optional. Regex") ||
		strings.Contains(desc, "Regex to test") {
		return true
	}

	// A glob names a SET of objects, so it is not a reference in any form: a Ref
	// identifies one resource. This is the same reasoning as the regex cases
	// above, applied to wildcard paths - e.g. aiplatform's GcsSource.uris, "Google
	// Cloud Storage URI(-s) to the input file(s). May contain wildcards."
	//
	// NOTE: this rule is ours, not upstream policy. See
	// docs/ai/refs-decision-guide.md for what is team-vetted and what is not.
	return mentionsWildcard(desc)
}

// mentionsWildcard reports whether a description says its value may be a glob.
//
// Deliberately conservative: it requires the description to say so, rather than
// inferring from a literal "*" anywhere, because "*" appears in ordinary prose
// (markdown emphasis, footnote markers) far more often than it denotes a glob.
func mentionsWildcard(desc string) bool {
	lower := strings.ToLower(desc)
	for _, phrase := range []string{"wildcard", "may contain wildcards", "glob pattern"} {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

// mentionsContainerImage reports whether a description names a container image
// registry. Checked BEFORE Cloud Storage: image descriptions frequently mention
// GCS incidentally (e.g. executorImageURI, "an image in Artifact Registry ...
// Google Cloud Storage paths ... will be mapped to local paths"), which would
// otherwise misfile them as GCS object paths.
func mentionsContainerImage(desc string) bool {
	return strings.Contains(desc, "Artifact Registry") ||
		strings.Contains(desc, "Container Registry") ||
		strings.Contains(desc, "gcr.io") ||
		strings.Contains(desc, "-docker.pkg.dev") ||
		strings.Contains(desc, "container image")
}

// notRepresentableReason returns a non-empty reason when a field looks like a
// reference but cannot be modeled as a KCC ref today. These are recorded rather
// than reported as missing refs, so that missingrefs.txt stays actionable.
func notRepresentableReason(fieldPath, desc string) string {
	isURIField := hasToken(fieldPath, "uri") || hasToken(fieldPath, "uris") ||
		hasToken(fieldPath, "url") || hasToken(fieldPath, "urls")

	// Scheme-specific checks run before the generic Cloud Storage check, so a
	// field that names its own registry/scheme is not misattributed to GCS.
	if isURIField && mentionsContainerImage(desc) {
		return "container-image-uri-not-a-storage-object"
	}

	// bq:// is not a GCP resource name. Its arity is ambiguous within a single
	// field (bq://p, bq://p.d, bq://p.d.t address three different kinds) and the
	// scheme is not consistent across services (aiplatform uses dots,
	// datalabeling documents slashes). KCC has no bq:// parsing at all.
	if strings.Contains(desc, "bq://") {
		return "bq-scheme-not-a-gcp-resource-name"
	}
	if isURIField && strings.Contains(desc, "BigQuery") {
		return "bigquery-uri-not-a-gcp-resource-name"
	}

	// Cloud Storage object paths and prefixes are not addressable KCC resources:
	// there is no StorageObject CRD, and StorageBucketIdentity explicitly rejects
	// anything with a "/" after the bucket. Bucket-only fields are handled as
	// real refs above.
	if isURIField && mentionsCloudStorage(desc) && !hasToken(fieldPath, "bucket") {
		// Two different situations, deliberately given different reasons so the
		// list can be worked down rather than treated as one opaque bucket:
		//
		//   - A prefix or directory addresses a location *within* a bucket, so it
		//     is expressible as bucketRef + path with an API change. Actionable,
		//     pending design.
		//   - A concrete object path or wildcard is not addressable at all today:
		//     there is no StorageObject CRD.
		if hasToken(fieldPath, "prefix") || hasToken(fieldPath, "directory") ||
			strings.Contains(desc, "output directory") || strings.Contains(desc, "directory path") {
			return "gcs-prefix-needs-bucket-ref-plus-path"
		}
		// A single object path. Kept as a string for now, but recorded rather than
		// dropped: this is a deferred design decision, NOT a claim that the field
		// cannot be a reference. KCC already manages StorageFolder and
		// StorageManagedFolder inside buckets, so objects are not categorically out
		// of scope. The same bucketRef + path decomposition would work here, with
		// composerenvironment (BucketRef + DagGCSPrefix) as precedent.
		//
		// refs_not_representable.txt is a golden file, not a ratchet, so this is a
		// warning: visible and countable, never blocking.
		return "gcs-object-path-string-for-now-decomposable-as-bucketref-plus-path"
	}

	return ""
}
