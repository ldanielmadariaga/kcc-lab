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

var CCInsightsPhraseMatcherGVK = GroupVersion.WithKind("CCInsightsPhraseMatcher")

// CCInsightsPhraseMatcherSpec defines the desired state of CCInsightsPhraseMatcher
// +kcc:spec:proto=google.cloud.contactcenterinsights.v1.PhraseMatcher
type CCInsightsPhraseMatcherSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/phraseMatchers/{phrase_matcher}
	Location *string `json:"location"`

	// The CCInsightsPhraseMatcher name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The customized version tag to use for the phrase matcher. If not specified,
	//  it will default to `revision_id`.
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.version_tag
	VersionTag *string `json:"versionTag,omitempty"`

	// The human-readable name of the phrase matcher.
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Required. The type of this phrase matcher.
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.type
	// +required
	Type *string `json:"type,omitempty"`

	// Applies the phrase matcher only when it is active.
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.active
	Active *bool `json:"active,omitempty"`

	// A list of phase match rule groups that are included in this matcher.
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.phrase_match_rule_groups
	PhraseMatchRuleGroups []PhraseMatchRuleGroup `json:"phraseMatchRuleGroups,omitempty"`

	// The role whose utterances the phrase matcher should be matched
	//  against. If the role is ROLE_UNSPECIFIED it will be matched against any
	//  utterances in the transcript.
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.role_match
	RoleMatch *string `json:"roleMatch,omitempty"`
}

// CCInsightsPhraseMatcherStatus defines the config connector machine state of CCInsightsPhraseMatcher
type CCInsightsPhraseMatcherStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the CCInsightsPhraseMatcher resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *CCInsightsPhraseMatcherObservedState `json:"observedState,omitempty"`
}

// CCInsightsPhraseMatcherObservedState is the state of the CCInsightsPhraseMatcher resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.contactcenterinsights.v1.PhraseMatcher
type CCInsightsPhraseMatcherObservedState struct {
	// Output only. Immutable. The revision ID of the phrase matcher.
	//  A new revision is committed whenever the matcher is changed, except when it
	//  is activated or deactivated. A server generated random ID will be used.
	//  Example: locations/global/phraseMatchers/my-first-matcher@1234567
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.revision_id
	RevisionID *string `json:"revisionID,omitempty"`

	// Output only. The timestamp of when the revision was created. It is also the
	//  create time when a new matcher is added.
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.revision_create_time
	RevisionCreateTime *string `json:"revisionCreateTime,omitempty"`

	// Output only. The most recent time at which the activation status was
	//  updated.
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.activation_update_time
	ActivationUpdateTime *string `json:"activationUpdateTime,omitempty"`

	// Output only. The most recent time at which the phrase matcher was updated.
	// +kcc:proto:field=google.cloud.contactcenterinsights.v1.PhraseMatcher.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpccinsightsphrasematcher;gcpccinsightsphrasematchers
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// CCInsightsPhraseMatcher is the Schema for the CCInsightsPhraseMatcher API
// +k8s:openapi-gen=true
type CCInsightsPhraseMatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   CCInsightsPhraseMatcherSpec   `json:"spec,omitempty"`
	Status CCInsightsPhraseMatcherStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// CCInsightsPhraseMatcherList contains a list of CCInsightsPhraseMatcher
type CCInsightsPhraseMatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CCInsightsPhraseMatcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CCInsightsPhraseMatcher{}, &CCInsightsPhraseMatcherList{})
}
