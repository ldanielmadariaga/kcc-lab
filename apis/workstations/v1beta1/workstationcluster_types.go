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
	common "github.com/GoogleCloudPlatform/k8s-config-connector/apis/common"
	refsv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/apis/refs/v1beta1"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var WorkstationClusterGVK = GroupVersion.WithKind("WorkstationCluster")

// WorkstationClusterSpec defines the desired state of WorkstationCluster
// +kcc:spec:proto=google.cloud.workstations.v1.WorkstationCluster
type WorkstationClusterSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/workstationClusters/{workstation_cluster}
	Location *string `json:"location"`

	// The WorkstationCluster name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Human-readable name for this workstation cluster.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. Client-specified annotations.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Optional.
	//  [Labels](https://cloud.google.com/workstations/docs/label-resources) that
	//  are applied to the workstation cluster and that are also propagated to the
	//  underlying Compute Engine resources.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. Checksum computed by the server. May be sent on update and delete
	//  requests to make sure that the client has an up-to-date value before
	//  proceeding.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.etag
	Etag *string `json:"etag,omitempty"`

	// Immutable. Name of the Compute Engine network in which instances associated
	//  with this workstation cluster will be created.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.network
	Network *string `json:"network,omitempty"`

	// Immutable. Name of the Compute Engine subnetwork in which instances
	//  associated with this workstation cluster will be created. Must be part of
	//  the subnetwork specified for this workstation cluster.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.subnetwork
	Subnetwork *string `json:"subnetwork,omitempty"`

	// Optional. Configuration for private workstation cluster.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.private_cluster_config
	PrivateClusterConfig *WorkstationCluster_PrivateClusterConfig `json:"privateClusterConfig,omitempty"`
}

// WorkstationClusterStatus defines the config connector machine state of WorkstationCluster
type WorkstationClusterStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the WorkstationCluster resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *WorkstationClusterObservedState `json:"observedState,omitempty"`
}

// WorkstationClusterObservedState is the state of the WorkstationCluster resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.workstations.v1.WorkstationCluster
type WorkstationClusterObservedState struct {
	// Output only. A system-assigned unique identifier for this workstation
	//  cluster.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. Indicates whether this workstation cluster is currently being
	//  updated to match its intended state.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.reconciling
	Reconciling *bool `json:"reconciling,omitempty"`

	// Output only. Time when this workstation cluster was created.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Time when this workstation cluster was most recently updated.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Time when this workstation cluster was soft-deleted.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.delete_time
	DeleteTime *string `json:"deleteTime,omitempty"`

	// Output only. The private IP address of the control plane for this
	//  workstation cluster. Workstation VMs need access to this IP address to work
	//  with the service, so make sure that your firewall rules allow egress from
	//  the workstation VMs to this address.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.control_plane_ip
	ControlPlaneIP *string `json:"controlPlaneIP,omitempty"`

	// Optional. Configuration for private workstation cluster.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.private_cluster_config
	PrivateClusterConfig *WorkstationCluster_PrivateClusterConfigObservedState `json:"privateClusterConfig,omitempty"`

	// Output only. Whether this workstation cluster is in degraded mode, in which
	//  case it may require user action to restore full functionality. Details can
	//  be found in
	//  [conditions][google.cloud.workstations.v1.WorkstationCluster.conditions].
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.degraded
	Degraded *bool `json:"degraded,omitempty"`

	// Output only. Status conditions describing the workstation cluster's current
	//  state.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationCluster.conditions
	Conditions []common.Status `json:"conditions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpworkstationcluster;gcpworkstationclusters
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// WorkstationCluster is the Schema for the WorkstationCluster API
// +k8s:openapi-gen=true
type WorkstationCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   WorkstationClusterSpec   `json:"spec,omitempty"`
	Status WorkstationClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// WorkstationClusterList contains a list of WorkstationCluster
type WorkstationClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkstationCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkstationCluster{}, &WorkstationClusterList{})
}
