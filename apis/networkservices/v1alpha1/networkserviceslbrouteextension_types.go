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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var NetworkServicesLBRouteExtensionGVK = GroupVersion.WithKind("NetworkServicesLBRouteExtension")

// NetworkServicesLBRouteExtensionSpec defines the desired state of NetworkServicesLBRouteExtension
// +kcc:spec:proto=google.cloud.networkservices.v1.LbRouteExtension
type NetworkServicesLBRouteExtensionSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The NetworkServicesLBRouteExtension name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. A human-readable description of the resource.
	// +kcc:proto:field=google.cloud.networkservices.v1.LbRouteExtension.description
	Description *string `json:"description,omitempty"`

	// Optional. Set of labels associated with the `LbRouteExtension` resource.
	//
	//  The format must comply with [the requirements for
	//  labels](https://cloud.google.com/compute/docs/labeling-resources#requirements)
	//  for Google Cloud resources.
	// +kcc:proto:field=google.cloud.networkservices.v1.LbRouteExtension.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. A list of references to the forwarding rules to which this
	//  service extension is attached. At least one forwarding rule is required.
	//  Only one `LbRouteExtension` resource can be associated with a forwarding
	//  rule.
	// +kcc:proto:field=google.cloud.networkservices.v1.LbRouteExtension.forwarding_rules
	// +required
	ForwardingRules []string `json:"forwardingRules,omitempty"`

	// Required. A set of ordered extension chains that contain the match
	//  conditions and extensions to execute. Match conditions for each extension
	//  chain are evaluated in sequence for a given request. The first extension
	//  chain that has a condition that matches the request is executed.
	//  Any subsequent extension chains do not execute.
	//  Limited to 5 extension chains per resource.
	// +kcc:proto:field=google.cloud.networkservices.v1.LbRouteExtension.extension_chains
	// +required
	ExtensionChains []ExtensionChain `json:"extensionChains,omitempty"`

	// Required. All backend services and forwarding rules referenced by this
	//  extension must share the same load balancing scheme. Supported values:
	//  `INTERNAL_MANAGED`, `EXTERNAL_MANAGED`. For more information, refer to
	//  [Backend services
	//  overview](https://cloud.google.com/load-balancing/docs/backend-service).
	// +kcc:proto:field=google.cloud.networkservices.v1.LbRouteExtension.load_balancing_scheme
	// +required
	LoadBalancingScheme *string `json:"loadBalancingScheme,omitempty"`

	// Optional. The metadata provided here is included as part of the
	//  `metadata_context` (of type `google.protobuf.Struct`) in the
	//  `ProcessingRequest` message sent to the extension server.
	//
	//  The metadata applies to all extensions in all extensions chains in this
	//  resource.
	//
	//  The metadata is available under the key
	//  `com.google.lb_route_extension.<resource_name>`.
	//
	//  The following variables are supported in the metadata:
	//
	//  `{forwarding_rule_id}` - substituted with the forwarding rule's fully
	//    qualified resource name.
	//
	//  This field must not be set if at least one of the extension chains
	//  contains plugin extensions. Setting it results in a validation error.
	//
	//  You can set metadata at either the resource level or the extension level.
	//  The extension level metadata is recommended because you can pass a
	//  different set of metadata through each extension to the backend.
	// +kcc:proto:field=google.cloud.networkservices.v1.LbRouteExtension.metadata
	Metadata apiextensionsv1.JSON `json:"metadata,omitempty"`
}

// NetworkServicesLBRouteExtensionStatus defines the config connector machine state of NetworkServicesLBRouteExtension
type NetworkServicesLBRouteExtensionStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the NetworkServicesLBRouteExtension resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *NetworkServicesLBRouteExtensionObservedState `json:"observedState,omitempty"`
}

// NetworkServicesLBRouteExtensionObservedState is the state of the NetworkServicesLBRouteExtension resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.networkservices.v1.LbRouteExtension
type NetworkServicesLBRouteExtensionObservedState struct {
	// Output only. The timestamp when the resource was created.
	// +kcc:proto:field=google.cloud.networkservices.v1.LbRouteExtension.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The timestamp when the resource was updated.
	// +kcc:proto:field=google.cloud.networkservices.v1.LbRouteExtension.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpnetworkserviceslbrouteextension;gcpnetworkserviceslbrouteextensions
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// NetworkServicesLBRouteExtension is the Schema for the NetworkServicesLBRouteExtension API
// +k8s:openapi-gen=true
type NetworkServicesLBRouteExtension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   NetworkServicesLBRouteExtensionSpec   `json:"spec,omitempty"`
	Status NetworkServicesLBRouteExtensionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkServicesLBRouteExtensionList contains a list of NetworkServicesLBRouteExtension
type NetworkServicesLBRouteExtensionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkServicesLBRouteExtension `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkServicesLBRouteExtension{}, &NetworkServicesLBRouteExtensionList{})
}
