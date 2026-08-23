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

var DataplexDataScanGVK = GroupVersion.WithKind("DataplexDataScan")

// DataplexDataScanSpec defines the desired state of DataplexDataScan
// +kcc:spec:proto=google.cloud.dataplex.v1.DataScan
type DataplexDataScanSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/dataScans/{dataScan}
	Location *string `json:"location"`

	// The DataplexDataScan name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Description of the scan.
	//
	//  * Must be between 1-1024 characters.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.description
	Description *string `json:"description,omitempty"`

	// Optional. User friendly display name.
	//
	//  * Must be between 1-256 characters.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. User-defined labels for the scan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. The data source for DataScan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.data
	// +required
	Data *DataSource `json:"data,omitempty"`

	// Optional. DataScan execution settings.
	//
	//  If not specified, the fields in it will use their default values.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.execution_spec
	ExecutionSpec *DataScan_ExecutionSpec `json:"executionSpec,omitempty"`

	// Settings for a data quality scan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.data_quality_spec
	DataQualitySpec *DataQualitySpec `json:"dataQualitySpec,omitempty"`

	// Settings for a data profile scan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.data_profile_spec
	DataProfileSpec *DataProfileSpec `json:"dataProfileSpec,omitempty"`

	// Settings for a data discovery scan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.data_discovery_spec
	DataDiscoverySpec *DataDiscoverySpec `json:"dataDiscoverySpec,omitempty"`
}

// DataplexDataScanStatus defines the config connector machine state of DataplexDataScan
type DataplexDataScanStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DataplexDataScan resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DataplexDataScanObservedState `json:"observedState,omitempty"`
}

// DataplexDataScanObservedState is the state of the DataplexDataScan resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.dataplex.v1.DataScan
type DataplexDataScanObservedState struct {
	// Output only. System generated globally unique ID for the scan. This ID will
	//  be different if the scan is deleted and re-created with the same name.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. Current state of the DataScan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.state
	State *string `json:"state,omitempty"`

	// Output only. The time when the scan was created.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The time when the scan was last updated.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Status of the data scan execution.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.execution_status
	ExecutionStatus *DataScan_ExecutionStatus `json:"executionStatus,omitempty"`

	// Output only. The type of DataScan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.type
	Type *string `json:"type,omitempty"`

	// Output only. The result of a data quality scan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.data_quality_result
	DataQualityResult *DataQualityResultObservedState `json:"dataQualityResult,omitempty"`

	// Output only. The result of a data profile scan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.data_profile_result
	DataProfileResult *DataProfileResultObservedState `json:"dataProfileResult,omitempty"`

	// Output only. The result of a data discovery scan.
	// +kcc:proto:field=google.cloud.dataplex.v1.DataScan.data_discovery_result
	DataDiscoveryResult *DataDiscoveryResultObservedState `json:"dataDiscoveryResult,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdataplexdatascan;gcpdataplexdatascans
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DataplexDataScan is the Schema for the DataplexDataScan API
// +k8s:openapi-gen=true
type DataplexDataScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DataplexDataScanSpec   `json:"spec,omitempty"`
	Status DataplexDataScanStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DataplexDataScanList contains a list of DataplexDataScan
type DataplexDataScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataplexDataScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataplexDataScan{}, &DataplexDataScanList{})
}
