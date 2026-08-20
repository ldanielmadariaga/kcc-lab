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

var DatastreamPrivateConnectionGVK = GroupVersion.WithKind("DatastreamPrivateConnection")

// DatastreamPrivateConnectionSpec defines the desired state of DatastreamPrivateConnection
// +kcc:spec:proto=google.cloud.datastream.v1.PrivateConnection
type DatastreamPrivateConnectionSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location string `json:"location"`

	// The DatastreamPrivateConnection name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Labels.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. Display name.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// VPC Peering Config.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.vpc_peering_config
	VPCPeeringConfig *VPCPeeringConfig `json:"vpcPeeringConfig,omitempty"`

	// PSC Interface Config.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.psc_interface_config
	PSCInterfaceConfig *PSCInterfaceConfig `json:"pscInterfaceConfig,omitempty"`
}

// DatastreamPrivateConnectionStatus defines the config connector machine state of DatastreamPrivateConnection
type DatastreamPrivateConnectionStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DatastreamPrivateConnection resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DatastreamPrivateConnectionObservedState `json:"observedState,omitempty"`
}

// DatastreamPrivateConnectionObservedState is the state of the DatastreamPrivateConnection resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.datastream.v1.PrivateConnection
type DatastreamPrivateConnectionObservedState struct {
	// Output only. The create time of the resource.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The update time of the resource.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The state of the Private Connection.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.state
	State *string `json:"state,omitempty"`

	// Output only. In case of error, the details of the error in a user-friendly
	//  format.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.error
	Error *Error `json:"error,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.satisfies_pzs
	SatisfiesPzs *bool `json:"satisfiesPzs,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.datastream.v1.PrivateConnection.satisfies_pzi
	SatisfiesPzi *bool `json:"satisfiesPzi,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdatastreamprivateconnection;gcpdatastreamprivateconnections
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DatastreamPrivateConnection is the Schema for the DatastreamPrivateConnection API
// +k8s:openapi-gen=true
type DatastreamPrivateConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DatastreamPrivateConnectionSpec   `json:"spec,omitempty"`
	Status DatastreamPrivateConnectionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DatastreamPrivateConnectionList contains a list of DatastreamPrivateConnection
type DatastreamPrivateConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatastreamPrivateConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatastreamPrivateConnection{}, &DatastreamPrivateConnectionList{})
}
