// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	refsv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/apis/refs/v1beta1"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var GrafeasNoteGVK = GroupVersion.WithKind("GrafeasNote")

// GrafeasNoteSpec defines the desired state of GrafeasNote
// +kcc:spec:proto=google.cloud.grafeas.v1.Note
type GrafeasNoteSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The GrafeasNote name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// A one sentence description of this note.
	// +kcc:proto:field=grafeas.v1.Note.short_description
	ShortDescription *string `json:"shortDescription,omitempty"`

	// A detailed description of this note.
	// +kcc:proto:field=grafeas.v1.Note.long_description
	LongDescription *string `json:"longDescription,omitempty"`

	// URLs associated with this note.
	// +kcc:proto:field=grafeas.v1.Note.related_url
	RelatedURL []RelatedURL `json:"relatedURL,omitempty"`

	// Time of expiration for this note. Empty if note does not expire.
	// +kcc:proto:field=grafeas.v1.Note.expiration_time
	ExpirationTime *string `json:"expirationTime,omitempty"`

	// Other notes related to this note.
	// +kcc:proto:field=grafeas.v1.Note.related_note_names
	RelatedNoteNames []string `json:"relatedNoteNames,omitempty"`

	// A note describing a package vulnerability.
	// +kcc:proto:field=grafeas.v1.Note.vulnerability
	Vulnerability *VulnerabilityNote `json:"vulnerability,omitempty"`

	// A note describing build provenance for a verifiable build.
	// +kcc:proto:field=grafeas.v1.Note.build
	Build *BuildNote `json:"build,omitempty"`

	// A note describing a base image.
	// +kcc:proto:field=grafeas.v1.Note.image
	Image *ImageNote `json:"image,omitempty"`

	// A note describing a package hosted by various package managers.
	// +kcc:proto:field=grafeas.v1.Note.package
	Package *PackageNote `json:"package,omitempty"`

	// A note describing something that can be deployed.
	// +kcc:proto:field=grafeas.v1.Note.deployment
	Deployment *DeploymentNote `json:"deployment,omitempty"`

	// A note describing the initial analysis of a resource.
	// +kcc:proto:field=grafeas.v1.Note.discovery
	Discovery *DiscoveryNote `json:"discovery,omitempty"`

	// A note describing an attestation role.
	// +kcc:proto:field=grafeas.v1.Note.attestation
	Attestation *AttestationNote `json:"attestation,omitempty"`

	// A note describing available package upgrades.
	// +kcc:proto:field=grafeas.v1.Note.upgrade
	Upgrade *UpgradeNote `json:"upgrade,omitempty"`

	// A note describing a compliance check.
	// +kcc:proto:field=grafeas.v1.Note.compliance
	Compliance *ComplianceNote `json:"compliance,omitempty"`

	// A note describing a dsse attestation note.
	// +kcc:proto:field=grafeas.v1.Note.dsse_attestation
	DsseAttestation *DsseAttestationNote `json:"dsseAttestation,omitempty"`

	// A note describing a vulnerability assessment.
	// +kcc:proto:field=grafeas.v1.Note.vulnerability_assessment
	VulnerabilityAssessment *VulnerabilityAssessmentNote `json:"vulnerabilityAssessment,omitempty"`

	// A note describing an SBOM reference.
	// +kcc:proto:field=grafeas.v1.Note.sbom_reference
	SbomReference *SbomReferenceNote `json:"sbomReference,omitempty"`
}

// GrafeasNoteStatus defines the config connector machine state of GrafeasNote
type GrafeasNoteStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the GrafeasNote resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *GrafeasNoteObservedState `json:"observedState,omitempty"`
}

// GrafeasNoteObservedState is the state of the GrafeasNote resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.grafeas.v1.Note
type GrafeasNoteObservedState struct {
	// Output only. The type of analysis. This field can be used as a filter in
	//  list requests.
	// +kcc:proto:field=grafeas.v1.Note.kind
	Kind *string `json:"kind,omitempty"`

	// Output only. The time this note was created. This field can be used as a
	//  filter in list requests.
	// +kcc:proto:field=grafeas.v1.Note.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The time this note was last updated. This field can be used as
	//  a filter in list requests.
	// +kcc:proto:field=grafeas.v1.Note.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +kcc:proto=grafeas.v1.WindowsUpdate
type WindowsUpdate struct {
	// Required - The unique identifier for the update.
	// +kcc:proto:field=grafeas.v1.WindowsUpdate.identity
	Identity *WindowsUpdate_Identity `json:"identity,omitempty"`

	// The localized title of the update.
	// +kcc:proto:field=grafeas.v1.WindowsUpdate.title
	Title *string `json:"title,omitempty"`

	// The localized description of the update.
	// +kcc:proto:field=grafeas.v1.WindowsUpdate.description
	Description *string `json:"description,omitempty"`

	// The list of categories to which the update belongs.
	// +kcc:proto:field=grafeas.v1.WindowsUpdate.categories
	Categories []WindowsUpdate_Category `json:"categories,omitempty"`

	// The Microsoft Knowledge Base article IDs that are associated with the
	//  update.
	// +kcc:proto:field=grafeas.v1.WindowsUpdate.kb_article_ids
	KbArticleIDs []string `json:"kbArticleIDs,omitempty"`

	// The hyperlink to the support information for the update.
	// +kcc:proto:field=grafeas.v1.WindowsUpdate.support_url
	SupportURL *string `json:"supportURL,omitempty"`

	// The last published timestamp of the update.
	// +kcc:proto:field=grafeas.v1.WindowsUpdate.last_published_timestamp
	LastPublishedTimestamp *string `json:"lastPublishedTimestamp,omitempty"`
}

// +kcc:proto=grafeas.v1.VulnerabilityAssessmentNote.Assessment
type VulnerabilityAssessmentNote_Assessment struct {
	// Holds the MITRE standard Common Vulnerabilities and Exposures (CVE)
	//  tracking number for the vulnerability.
	//  Deprecated: Use vulnerability_id instead to denote CVEs.
	// +kcc:proto:field=grafeas.v1.VulnerabilityAssessmentNote.Assessment.cve
	Cve *string `json:"cve,omitempty"`

	// The vulnerability identifier for this Assessment. Will hold one of
	//  common identifiers e.g. CVE, GHSA etc.
	// +kcc:proto:field=grafeas.v1.VulnerabilityAssessmentNote.Assessment.vulnerability_id
	VulnerabilityID *string `json:"vulnerabilityID,omitempty"`

	// A one sentence description of this Vex.
	// +kcc:proto:field=grafeas.v1.VulnerabilityAssessmentNote.Assessment.short_description
	ShortDescription *string `json:"shortDescription,omitempty"`

	// A detailed description of this Vex.
	// +kcc:proto:field=grafeas.v1.VulnerabilityAssessmentNote.Assessment.long_description
	LongDescription *string `json:"longDescription,omitempty"`

	// Holds a list of references associated with this vulnerability item and
	//  assessment. These uris have additional information about the
	//  vulnerability and the assessment itself. E.g. Link to a document which
	//  details how this assessment concluded the state of this vulnerability.
	// +kcc:proto:field=grafeas.v1.VulnerabilityAssessmentNote.Assessment.related_uris
	RelatedURIs []RelatedURL `json:"relatedURIs,omitempty"`

	// Provides the state of this Vulnerability assessment.
	// +kcc:proto:field=grafeas.v1.VulnerabilityAssessmentNote.Assessment.state
	State *string `json:"state,omitempty"`

	// Contains information about the impact of this vulnerability,
	//  this will change with time.
	// +kcc:proto:field=grafeas.v1.VulnerabilityAssessmentNote.Assessment.impacts
	Impacts []string `json:"impacts,omitempty"`

	// Justification provides the justification when the state of the
	//  assessment if NOT_AFFECTED.
	// +kcc:proto:field=grafeas.v1.VulnerabilityAssessmentNote.Assessment.justification
	Justification *VulnerabilityAssessmentNote_Assessment_Justification `json:"justification,omitempty"`

	// Specifies details on how to handle (and presumably, fix) a vulnerability.
	// +kcc:proto:field=grafeas.v1.VulnerabilityAssessmentNote.Assessment.remediations
	Remediations []VulnerabilityAssessmentNote_Assessment_Remediation `json:"remediations,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpgrafeasnote;gcpgrafeasnotes
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// GrafeasNote is the Schema for the GrafeasNote API
// +k8s:openapi-gen=true
type GrafeasNote struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   GrafeasNoteSpec   `json:"spec,omitempty"`
	Status GrafeasNoteStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// GrafeasNoteList contains a list of GrafeasNote
type GrafeasNoteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafeasNote `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafeasNote{}, &GrafeasNoteList{})
}
