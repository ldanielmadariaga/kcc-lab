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

var WorkstationGVK = GroupVersion.WithKind("Workstation")

// WorkstationSpec defines the desired state of Workstation
// +kcc:spec:proto=google.cloud.workstations.v1.Workstation
type WorkstationSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/workstationClusters/{workstation_cluster}/workstationConfigs/{workstation_config}/workstations/{workstation}
	Location *string `json:"location,omitempty"`

	// The WorkstationCluster that this resource belongs to.
	// +kcc:guess=parent-ref target=WorkstationClusterRef pattern=projects/{project}/locations/{location}/workstationClusters/{workstation_cluster}/workstationConfigs/{workstation_config}/workstations/{workstation}
	WorkstationClusterRef *WorkstationClusterRef `json:"workstationClusterRef,omitempty"`

	// The WorkstationConfig that this resource belongs to.
	// +kcc:guess=parent-ref target=WorkstationConfigRef pattern=projects/{project}/locations/{location}/workstationClusters/{workstation_cluster}/workstationConfigs/{workstation_config}/workstations/{workstation}
	WorkstationConfigRef *WorkstationConfigRef `json:"workstationConfigRef,omitempty"`

	// The Workstation name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Human-readable name for this workstation.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. Client-specified annotations.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Optional.
	//  [Labels](https://cloud.google.com/workstations/docs/label-resources) that
	//  are applied to the workstation and that are also propagated to the
	//  underlying Compute Engine resources.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. Checksum computed by the server. May be sent on update and delete
	//  requests to make sure that the client has an up-to-date value before
	//  proceeding.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.etag
	Etag *string `json:"etag,omitempty"`
}

// WorkstationStatus defines the config connector machine state of Workstation
type WorkstationStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the Workstation resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *WorkstationObservedState `json:"observedState,omitempty"`
}

// WorkstationObservedState is the state of the Workstation resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.workstations.v1.Workstation
type WorkstationObservedState struct {
	// Output only. A system-assigned unique identifier for this workstation.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. Indicates whether this workstation is currently being updated
	//  to match its intended state.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.reconciling
	Reconciling *bool `json:"reconciling,omitempty"`

	// Output only. Time when this workstation was created.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Time when this workstation was most recently updated.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Time when this workstation was most recently successfully
	//  started, regardless of the workstation's initial state.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.start_time
	StartTime *string `json:"startTime,omitempty"`

	// Output only. Time when this workstation was soft-deleted.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.delete_time
	DeleteTime *string `json:"deleteTime,omitempty"`

	// Output only. Current state of the workstation.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.state
	State *string `json:"state,omitempty"`

	// Output only. Host to which clients can send HTTPS traffic that will be
	//  received by the workstation. Authorized traffic will be received to the
	//  workstation as HTTP on port 80. To send traffic to a different port,
	//  clients may prefix the host with the destination port in the format
	//  `{port}-{host}`.
	// +kcc:proto:field=google.cloud.workstations.v1.Workstation.host
	Host *string `json:"host,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpworkstation;gcpworkstations
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// Workstation is the Schema for the Workstation API
// +k8s:openapi-gen=true
type Workstation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   WorkstationSpec   `json:"spec,omitempty"`
	Status WorkstationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// WorkstationList contains a list of Workstation
type WorkstationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workstation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workstation{}, &WorkstationList{})
}
