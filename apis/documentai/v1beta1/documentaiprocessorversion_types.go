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

package v1beta1

import (
	refsv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/apis/refs/v1beta1"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DocumentAIProcessorVersionGVK = GroupVersion.WithKind("DocumentAIProcessorVersion")

// DocumentAIProcessorVersionSpec defines the desired state of DocumentAIProcessorVersion
// +kcc:spec:proto=google.cloud.documentai.v1.ProcessorVersion
type DocumentAIProcessorVersionSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/processors/{processor}/processorVersions/{processor_version}
	Location *string `json:"location,omitempty"`

	// The Processor that this resource belongs to.
	// +kcc:guess=parent-segment pattern=projects/{project}/locations/{location}/processors/{processor}/processorVersions/{processor_version}
	Processor *string `json:"processor,omitempty"`

	// The DocumentAIProcessorVersion name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The display name of the processor version.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.display_name
	DisplayName *string `json:"displayName,omitempty"`
}

// DocumentAIProcessorVersionStatus defines the config connector machine state of DocumentAIProcessorVersion
type DocumentAIProcessorVersionStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DocumentAIProcessorVersion resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DocumentAIProcessorVersionObservedState `json:"observedState,omitempty"`
}

// DocumentAIProcessorVersionObservedState is the state of the DocumentAIProcessorVersion resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.documentai.v1.ProcessorVersion
type DocumentAIProcessorVersionObservedState struct {
	// Output only. The schema of the processor version. Describes the output.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.document_schema
	DocumentSchema *DocumentSchema `json:"documentSchema,omitempty"`

	// Output only. The state of the processor version.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.state
	State *string `json:"state,omitempty"`

	// Output only. The time the processor version was created.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The most recently invoked evaluation for the processor
	//  version.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.latest_evaluation
	LatestEvaluation *EvaluationReference `json:"latestEvaluation,omitempty"`

	// Output only. The KMS key name used for encryption.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.kms_key_name
	KMSKeyName *string `json:"kmsKeyName,omitempty"`

	// Output only. The KMS key version with which data is encrypted.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.kms_key_version_name
	KMSKeyVersionName *string `json:"kmsKeyVersionName,omitempty"`

	// Output only. Denotes that this `ProcessorVersion` is managed by Google.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.google_managed
	GoogleManaged *bool `json:"googleManaged,omitempty"`

	// Output only. If set, information about the eventual deprecation of this
	//  version.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.deprecation_info
	DeprecationInfo *ProcessorVersion_DeprecationInfo `json:"deprecationInfo,omitempty"`

	// Output only. The model type of this processor version.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.model_type
	ModelType *string `json:"modelType,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.satisfies_pzs
	SatisfiesPzs *bool `json:"satisfiesPzs,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.satisfies_pzi
	SatisfiesPzi *bool `json:"satisfiesPzi,omitempty"`

	// Output only. Information about Generative AI model-based processor
	//  versions.
	// +kcc:proto:field=google.cloud.documentai.v1.ProcessorVersion.gen_ai_model_info
	GenAiModelInfo *ProcessorVersion_GenAiModelInfo `json:"genAiModelInfo,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdocumentaiprocessorversion;gcpdocumentaiprocessorversions
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DocumentAIProcessorVersion is the Schema for the DocumentAIProcessorVersion API
// +k8s:openapi-gen=true
type DocumentAIProcessorVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DocumentAIProcessorVersionSpec   `json:"spec,omitempty"`
	Status DocumentAIProcessorVersionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DocumentAIProcessorVersionList contains a list of DocumentAIProcessorVersion
type DocumentAIProcessorVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DocumentAIProcessorVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DocumentAIProcessorVersion{}, &DocumentAIProcessorVersionList{})
}
