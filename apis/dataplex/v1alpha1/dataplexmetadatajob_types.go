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

var DataplexMetadataJobGVK = GroupVersion.WithKind("DataplexMetadataJob")

// DataplexMetadataJobSpec defines the desired state of DataplexMetadataJob
// +kcc:spec:proto=google.cloud.dataplex.v1.MetadataJob
type DataplexMetadataJobSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/metadataJobs/{metadataJob}
	Location *string `json:"location"`

	// The DataplexMetadataJob name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. User-defined labels.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. Metadata job type.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.type
	// +required
	Type *string `json:"type,omitempty"`

	// Import job specification.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.import_spec
	ImportSpec *MetadataJob_ImportJobSpec `json:"importSpec,omitempty"`

	// Export job specification.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.export_spec
	ExportSpec *MetadataJob_ExportJobSpec `json:"exportSpec,omitempty"`
}

// DataplexMetadataJobStatus defines the config connector machine state of DataplexMetadataJob
type DataplexMetadataJobStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DataplexMetadataJob resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DataplexMetadataJobObservedState `json:"observedState,omitempty"`
}

// DataplexMetadataJobObservedState is the state of the DataplexMetadataJob resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.dataplex.v1.MetadataJob
type DataplexMetadataJobObservedState struct {
	// Output only. A system-generated, globally unique ID for the metadata job.
	//  If the metadata job is deleted and then re-created with the same name, this
	//  ID is different.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. The time when the metadata job was created.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The time when the metadata job was updated.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Import job result.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.import_result
	ImportResult *MetadataJob_ImportJobResultObservedState `json:"importResult,omitempty"`

	// Output only. Export job result.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.export_result
	ExportResult *MetadataJob_ExportJobResultObservedState `json:"exportResult,omitempty"`

	// Output only. Metadata job status.
	// +kcc:proto:field=google.cloud.dataplex.v1.MetadataJob.status
	Status *MetadataJob_StatusObservedState `json:"status,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdataplexmetadatajob;gcpdataplexmetadatajobs
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DataplexMetadataJob is the Schema for the DataplexMetadataJob API
// +k8s:openapi-gen=true
type DataplexMetadataJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DataplexMetadataJobSpec   `json:"spec,omitempty"`
	Status DataplexMetadataJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DataplexMetadataJobList contains a list of DataplexMetadataJob
type DataplexMetadataJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataplexMetadataJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataplexMetadataJob{}, &DataplexMetadataJobList{})
}
