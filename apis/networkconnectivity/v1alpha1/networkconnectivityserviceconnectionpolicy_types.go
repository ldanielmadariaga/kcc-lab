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

var NetworkConnectivityServiceConnectionPolicyGVK = GroupVersion.WithKind("NetworkConnectivityServiceConnectionPolicy")

// NetworkConnectivityServiceConnectionPolicySpec defines the desired state of NetworkConnectivityServiceConnectionPolicy
// +kcc:spec:proto=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy
type NetworkConnectivityServiceConnectionPolicySpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The NetworkConnectivityServiceConnectionPolicy name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Output only. Information for the automatically created subnetwork and its associated IR.
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.auto_created_subnet_info
	AutoCreatedSubnetInfo *AutoCreatedSubnetworkInfo `json:"autoCreatedSubnetInfo,omitempty"`

	// A description of this resource.
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.description
	Description *string `json:"description,omitempty"`

	// Output only. The type of underlying resources used to create the connection.
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.infrastructure
	Infrastructure *string `json:"infrastructure,omitempty"`

	// User-defined labels.
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.labels
	Labels map[string]string `json:"labels,omitempty"`

	// The resource path of the consumer network. Example: - projects/{projectNumOrId}/global/networks/{resourceId}.
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.network
	Network *string `json:"network,omitempty"`

	// Configuration used for Private Service Connect connections. Used when Infrastructure is PSC.
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.psc_config
	PSCConfig *PSCConfig `json:"pscConfig,omitempty"`

	// Output only. [Output only] Information about each Private Service Connect connection.
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.psc_connections
	PSCConnections []PSCConnection `json:"pscConnections,omitempty"`

	// The service class identifier for which this ServiceConnectionPolicy is for. The service class identifier is a unique, symbolic representation of a ServiceClass. It is provided by the Service Producer. Google services have a prefix of gcp or google-cloud. For example, gcp-memorystore-redis or google-cloud-sql. 3rd party services do not. For example, test-service-a3dfcx.
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.service_class
	ServiceClass *string `json:"serviceClass,omitempty"`
}

// NetworkConnectivityServiceConnectionPolicyStatus defines the config connector machine state of NetworkConnectivityServiceConnectionPolicy
type NetworkConnectivityServiceConnectionPolicyStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the NetworkConnectivityServiceConnectionPolicy resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *NetworkConnectivityServiceConnectionPolicyObservedState `json:"observedState,omitempty"`
}

// NetworkConnectivityServiceConnectionPolicyObservedState is the state of the NetworkConnectivityServiceConnectionPolicy resource as most recently observed in GCP.
// +kcc:observedstate:proto=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy
type NetworkConnectivityServiceConnectionPolicyObservedState struct {
	// Output only. Time when the ServiceConnectionPolicy was created.
	// +kcc:guess=placement reason=no-field-behavior-on-message
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Optional. The etag is computed by the server, and may be sent on update and delete requests to ensure the client has an up-to-date value before proceeding.
	// +kcc:guess=placement reason=no-field-behavior-on-message
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.etag
	Etag *string `json:"etag,omitempty"`

	// Output only. Time when the ServiceConnectionPolicy was updated.
	// +kcc:guess=placement reason=no-field-behavior-on-message
	// +kcc:proto:field=mockgcp.cloud.networkconnectivity.v1.ServiceConnectionPolicy.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpnetworkconnectivityserviceconnectionpolicy;gcpnetworkconnectivityserviceconnectionpolicys
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// NetworkConnectivityServiceConnectionPolicy is the Schema for the NetworkConnectivityServiceConnectionPolicy API
// +k8s:openapi-gen=true
type NetworkConnectivityServiceConnectionPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   NetworkConnectivityServiceConnectionPolicySpec   `json:"spec,omitempty"`
	Status NetworkConnectivityServiceConnectionPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkConnectivityServiceConnectionPolicyList contains a list of NetworkConnectivityServiceConnectionPolicy
type NetworkConnectivityServiceConnectionPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkConnectivityServiceConnectionPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkConnectivityServiceConnectionPolicy{}, &NetworkConnectivityServiceConnectionPolicyList{})
}
