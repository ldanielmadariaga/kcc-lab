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

var NetworkSecurityMirroringEndpointGroupGVK = GroupVersion.WithKind("NetworkSecurityMirroringEndpointGroup")

// NetworkSecurityMirroringEndpointGroupSpec defines the desired state of NetworkSecurityMirroringEndpointGroup
// +kcc:spec:proto=google.cloud.networksecurity.v1.MirroringEndpointGroup
type NetworkSecurityMirroringEndpointGroupSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The NetworkSecurityMirroringEndpointGroup name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Labels are key/value pairs that help to organize and filter
	//  resources.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Immutable. The deployment group that this DIRECT endpoint group is
	//  connected to, for example:
	//  `projects/123456789/locations/global/mirroringDeploymentGroups/my-dg`.
	//  See https://google.aip.dev/124.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.mirroring_deployment_group
	MirroringDeploymentGroup *string `json:"mirroringDeploymentGroup,omitempty"`

	// Immutable. The type of the endpoint group.
	//  If left unspecified, defaults to DIRECT.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.type
	Type *string `json:"type,omitempty"`

	// Optional. User-provided description of the endpoint group.
	//  Used as additional context for the endpoint group.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.description
	Description *string `json:"description,omitempty"`
}

// NetworkSecurityMirroringEndpointGroupStatus defines the config connector machine state of NetworkSecurityMirroringEndpointGroup
type NetworkSecurityMirroringEndpointGroupStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the NetworkSecurityMirroringEndpointGroup resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *NetworkSecurityMirroringEndpointGroupObservedState `json:"observedState,omitempty"`
}

// NetworkSecurityMirroringEndpointGroupObservedState is the state of the NetworkSecurityMirroringEndpointGroup resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.networksecurity.v1.MirroringEndpointGroup
type NetworkSecurityMirroringEndpointGroupObservedState struct {
	// Output only. The timestamp when the resource was created.
	//  See https://google.aip.dev/148#timestamps.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The timestamp when the resource was most recently updated.
	//  See https://google.aip.dev/148#timestamps.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. List of details about the connected deployment groups to this
	//  endpoint group.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.connected_deployment_groups
	ConnectedDeploymentGroups []MirroringEndpointGroup_ConnectedDeploymentGroupObservedState `json:"connectedDeploymentGroups,omitempty"`

	// Output only. The current state of the endpoint group.
	//  See https://google.aip.dev/216.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.state
	State *string `json:"state,omitempty"`

	// Output only. The current state of the resource does not match the user's
	//  intended state, and the system is working to reconcile them. This is part
	//  of the normal operation (e.g. adding a new association to the group). See
	//  https://google.aip.dev/128.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.reconciling
	Reconciling *bool `json:"reconciling,omitempty"`

	// Output only. List of associations to this endpoint group.
	// +kcc:proto:field=google.cloud.networksecurity.v1.MirroringEndpointGroup.associations
	Associations []MirroringEndpointGroup_AssociationDetailsObservedState `json:"associations,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpnetworksecuritymirroringendpointgroup;gcpnetworksecuritymirroringendpointgroups
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// NetworkSecurityMirroringEndpointGroup is the Schema for the NetworkSecurityMirroringEndpointGroup API
// +k8s:openapi-gen=true
type NetworkSecurityMirroringEndpointGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   NetworkSecurityMirroringEndpointGroupSpec   `json:"spec,omitempty"`
	Status NetworkSecurityMirroringEndpointGroupStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkSecurityMirroringEndpointGroupList contains a list of NetworkSecurityMirroringEndpointGroup
type NetworkSecurityMirroringEndpointGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkSecurityMirroringEndpointGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkSecurityMirroringEndpointGroup{}, &NetworkSecurityMirroringEndpointGroupList{})
}
