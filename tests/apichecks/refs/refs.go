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

// Package refs decides whether a CRD spec field ought to be a KCC reference.
//
// The rules live here rather than in the check that first used them because
// there are now two callers with opposite purposes: TestMissingRefs gates
// resources that have graduated from the judgement queue, and the queue seeder
// hints at work for resources still in it. One implementation, so the two cannot
// drift.
//
// Everything here works on a field path and its description, the two things a
// CRD retains. Note what that rules out: a CRD does not record which proto field
// it came from, so google.api.resource_reference -- which names the target type
// exactly -- is unavailable at this layer. Matching it by leaf name instead was
// tried and measured at 2,164 findings against 78 for the description rules, and
// abandoned.
package refs

import "strings"

// Verdict is what the rules concluded about a field.
type Verdict int

const (
	// NotAReference: nothing suggests this field names another resource.
	NotAReference Verdict = iota
	// IsReference: this field names another resource and a KCC ref can express it.
	IsReference
	// NotRepresentable: it names another resource, but KCC cannot model it today.
	// Recorded with a reason rather than reported as missing, so the actionable
	// list stays actionable.
	NotRepresentable
)

// exemptSuffixes name things that look like references but are values, not
// resources: a zone or location is a coordinate, and a machine or accelerator
// type is a catalogue entry rather than an object anyone owns.
var exemptSuffixes = []string{".zone", ".location", ".machineType", ".acceleratorType"}

// Classify applies the rules to one CRD spec field. reason is set only for
// NotRepresentable.
//
// Callers are expected to have skipped status fields and fields that are already
// references.
func Classify(fieldPath, desc string) (verdict Verdict, reason string) {
	// A pattern or matcher describes a SET of resources, never one, so it is not
	// a reference in either sense.
	if isPatternField(fieldPath, desc) {
		return NotAReference, ""
	}
	if r := notRepresentableReason(fieldPath, desc); r != "" {
		return NotRepresentable, r
	}

	isRef := false
	// A resource-name path template in the description, e.g.
	// "projects/{projectID}/locations/{location}/bars/{name}".
	if hasResourceNameTemplate(desc) {
		isRef = true
	}
	if strings.HasSuffix(fieldPath, "erviceAccount") {
		isRef = true
	}
	// Cloud Storage buckets are expressible: StorageBucketIdentity.FromExternal
	// accepts the bare "gs://<bucket>" form. Object paths are caught above as
	// not-representable.
	if hasToken(fieldPath, "bucket") && mentionsCloudStorage(desc) {
		isRef = true
	}
	if !isRef {
		return NotAReference, ""
	}
	for _, s := range exemptSuffixes {
		if strings.HasSuffix(fieldPath, s) {
			return NotAReference, ""
		}
	}
	return IsReference, ""
}

// IsReferenceFieldPath reports whether a path is already modelled as a reference,
// including the external/name children a ref expands into.
func IsReferenceFieldPath(fieldPath string) bool {
	for _, s := range []string{"Ref", "Refs", "Refs[]", "Ref.external", "Refs[].external", "Ref.name"} {
		if strings.HasSuffix(fieldPath, s) {
			return true
		}
	}
	return false
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

// referenceShapeChildren are the properties a KCC reference object carries.
// "external" is the discriminating one: it is what makes a reference a
// reference, and no ordinary message has it alongside name and namespace.
var referenceShapeChildren = []string{"external", "name", "namespace"}

// HasReferenceShape reports whether an object's property names are those of a
// KCC reference.
//
// Needed because the name is not reliable. IsReferenceFieldPath keys off a Ref
// or Refs suffix, which most references have, but not all: upstream models
// NetworkManagementConnectivityTest's destination.sqlInstance as a full
// reference while calling it "sqlInstance". Judged by name alone its external,
// name and namespace children look like three ordinary missing fields, which is
// how 25 reference children ended up miscounted as unexplained drops.
//
// Callers pass the property names of the object being visited.
func HasReferenceShape(propertyNames []string) bool {
	have := make(map[string]bool, len(propertyNames))
	for _, n := range propertyNames {
		have[n] = true
	}
	if !have["external"] {
		return false
	}
	for _, n := range referenceShapeChildren {
		if !have[n] {
			return false
		}
	}
	return true
}

// NameRule matches a field whose name identifies what it points at.
//
// These exist because the description signals miss a whole class: a Secret
// Manager version or a VPC is often named plainly, with no resource-name
// template in its description, so hasResourceNameTemplate never fires.
// Measured on the 239-resource run, that class is half of what the rules were
// missing.
//
// On reading their measured precision: 20 of the 21 hints the answer key calls
// wrong are fields whose exact name upstream turned into a reference in another
// resource -- "network" is a reference in nine resources and a plain string in
// BlockchainNodeEngineBlockchainNode, userTokenSecretVersion is a reference in
// CloudBuildConnection and a string in DevConnectConnection. So the raw figure
// measures upstream's inconsistency more than this code's error rate, and for a
// queue that proposes work to a person, surfacing them is the point. Only one
// hint had a genuinely novel name.
//
// Kept as a short list of specific targets rather than a general vocabulary.
// Matching reference names generically was tried and rejected at 2,164
// findings. What makes these safe is that each names one target type and is
// measured against the answer key in
// hack/tools/greenfield/reference_testset.tsv, with any rule that does not earn
// its place removed.
type NameRule struct {
	// Target is the KCC ref type the field points at, for the hint text.
	Target string
	// Match takes the field's leaf name, already stripped of any list suffix.
	Match func(leaf string) bool
}

func eq(names ...string) func(string) bool {
	return func(leaf string) bool {
		for _, n := range names {
			if strings.EqualFold(leaf, n) {
				return true
			}
		}
		return false
	}
}

func hasSuffix(suffix string) func(string) bool {
	return func(leaf string) bool {
		return len(leaf) > len(suffix) && strings.EqualFold(leaf[len(leaf)-len(suffix):], suffix)
	}
}

// NameRules is ordered most specific first; the first match wins.
var NameRules = []NameRule{
	// Every confirmed instance ends in SecretVersion, and the suffix is long
	// enough not to collide with anything else in the corpus.
	{Target: "SecretManagerSecretVersionRef", Match: hasSuffix("SecretVersion")},
	// "network" is the exact name the rejected heuristic was built on, so it is
	// admitted only as a whole leaf, never as a substring: a field called
	// networkConfig or networkPolicy is not a network.
	{Target: "ComputeNetworkRef", Match: eq("network", "vpc", "vpcName")},
	{Target: "KMSCryptoKeyRef", Match: eq("kmsKey", "cmekKeyName", "encryptionKey", "kmsKeyName")},
	// A field whose name calls itself a secret, other than the SecretVersion
	// suffix the rule above already owns. Measured over fields upstream actually
	// modelled: 4 references, 0 kept plain. The exclusion matters -- without it
	// the same measurement reads 16 and 16, because it picks up DevConnect's
	// userTokenSecretVersion fields, which the suffix rule already hints and
	// which upstream did keep plain.
	//
	// Names around these were measured and rejected: "ends Certificate" is 0 for
	// 6 and "ends PrivateKey" 0 for 3, because pemCertificate and privateKey are
	// contents rather than references, whatever ConnectorsConnection does with
	// its own.
	{Target: "SecretManagerSecretVersionRef", Match: wholeWordNotSuffix("secret", "SecretVersion")},
	// Measured 4 for 4. Kept to the exact leaf: "projectNumber" and "projectID"
	// are values, and this must not become a substring rule.
	{Target: "ProjectRef", Match: eq("project")},
}

// wholeWordNotSuffix matches a leaf containing word as a camelCase word, unless
// the leaf ends in exclude -- which another rule owns.
func wholeWordNotSuffix(word, exclude string) func(string) bool {
	return func(leaf string) bool {
		if strings.EqualFold(leaf[max(0, len(leaf)-len(exclude)):], exclude) {
			return false
		}
		for _, w := range splitCamelWords(leaf) {
			if w == word {
				return true
			}
		}
		return false
	}
}

// MatchName returns the reference target a field's name indicates, if any.
//
// Deliberately NOT consulted by Classify. TestMissingRefs writes its findings
// to missingrefs.txt, a ratchet that refuses new entries even under
// WRITE_GOLDEN_OUTPUT, and these rules add 28 of them -- correct ones, such as
// ComputePacketMirroring's .spec.network, but a gate nobody can pass until
// every one is implemented. The queue seeder proposes work and gates nothing,
// so it can use them today. Tightening the check is a separate decision, and
// has to come with the implementations or reviewed refs_deferred.txt entries.
func MatchName(fieldPath string) (string, bool) {
	leaf := fieldPath
	if i := strings.LastIndex(leaf, "."); i >= 0 {
		leaf = leaf[i+1:]
	}
	leaf = strings.TrimSuffix(leaf, "[]")
	for _, r := range NameRules {
		if r.Match(leaf) {
			return r.Target, true
		}
	}
	return "", false
}
