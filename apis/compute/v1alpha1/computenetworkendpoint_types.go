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

var ComputeNetworkEndpointGVK = GroupVersion.WithKind("ComputeNetworkEndpoint")

// ComputeNetworkEndpointSpec defines the desired state of ComputeNetworkEndpoint
// +kcc:spec:proto=google.cloud.compute.v1.NetworkEndpoint
type ComputeNetworkEndpointSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`


	// The ComputeNetworkEndpoint name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional metadata defined as annotations on the network endpoint.
	// +kcc:proto:field=google.cloud.compute.v1.NetworkEndpoint.annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Represents the port number to which PSC consumer sends packets. Optional. Only valid for network endpoint groups created with GCE_VM_IP_PORTMAP endpoint type.
	// +kcc:proto:field=google.cloud.compute.v1.NetworkEndpoint.client_destination_port
	ClientDestinationPort *int32 `json:"clientDestinationPort,omitempty"`

	// Optional fully qualified domain name of network endpoint. This can only be specified when NetworkEndpointGroup.network_endpoint_type is NON_GCP_FQDN_PORT.
	// +kcc:proto:field=google.cloud.compute.v1.NetworkEndpoint.fqdn
	FQDN *string `json:"fqdn,omitempty"`

	// The name or a URL of VM instance of this network endpoint. Optional, the field presence depends on the network endpoint type. The field is required for network endpoints of type GCE_VM_IP and GCE_VM_IP_PORT. The instance must be in the same zone of network endpoint group (for zonal NEGs) or in the zone within the region of the NEG (for regional NEGs). If the ipAddress is specified, it must belongs to the VM instance. The name must be 1-63 characters long, and comply with RFC1035 or be a valid URL pointing to an existing instance.
	// +kcc:proto:field=google.cloud.compute.v1.NetworkEndpoint.instance
	Instance *string `json:"instance,omitempty"`

	// Optional IPv4 address of network endpoint. The IP address must belong to a VM in Compute Engine (either the primary IP or as part of an aliased IP range). If the IP address is not specified, then the primary IP address for the VM instance in the network that the network endpoint group belongs to will be used. This field is redundant and need not be set for network endpoints of type GCE_VM_IP. If set, it must be set to the primary internal IP address of the attached VM instance that matches the subnetwork of the NEG. The primary internal IP address from any NIC of a multi-NIC VM instance can be added to a NEG as long as it matches the NEG subnetwork.
	// +kcc:proto:field=google.cloud.compute.v1.NetworkEndpoint.ip_address
	IPAddress *string `json:"ipAddress,omitempty"`

	// Optional IPv6 address of network endpoint.
	// +kcc:proto:field=google.cloud.compute.v1.NetworkEndpoint.ipv6_address
	IPV6Address *string `json:"ipv6Address,omitempty"`

	// Optional port number of network endpoint. If not specified, the defaultPort for the network endpoint group will be used. This field can not be set for network endpoints of type GCE_VM_IP.
	// +kcc:proto:field=google.cloud.compute.v1.NetworkEndpoint.port
	Port *int32 `json:"port,omitempty"`
}

// ComputeNetworkEndpointStatus defines the config connector machine state of ComputeNetworkEndpoint
type ComputeNetworkEndpointStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the ComputeNetworkEndpoint resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *ComputeNetworkEndpointObservedState `json:"observedState,omitempty"`
}

// ComputeNetworkEndpointObservedState is the state of the ComputeNetworkEndpoint resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.compute.v1.NetworkEndpoint
type ComputeNetworkEndpointObservedState struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcomputenetworkendpoint;gcpcomputenetworkendpoints
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// ComputeNetworkEndpoint is the Schema for the ComputeNetworkEndpoint API
// +k8s:openapi-gen=true
type ComputeNetworkEndpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ComputeNetworkEndpointSpec   `json:"spec,omitempty"`
	Status ComputeNetworkEndpointStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ComputeNetworkEndpointList contains a list of ComputeNetworkEndpoint
type ComputeNetworkEndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeNetworkEndpoint `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeNetworkEndpoint{}, &ComputeNetworkEndpointList{})
}
