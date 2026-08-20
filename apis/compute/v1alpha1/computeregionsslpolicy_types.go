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

var ComputeRegionSSLPolicyGVK = GroupVersion.WithKind("ComputeRegionSSLPolicy")

// ComputeRegionSSLPolicySpec defines the desired state of ComputeRegionSSLPolicy
// +kcc:spec:proto=google.cloud.compute.v1.SslPolicy
type ComputeRegionSSLPolicySpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The ComputeRegionSSLPolicy name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// [Output Only] Creation timestamp in RFC3339 text format.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.creation_timestamp
	CreationTimestamp *string `json:"creationTimestamp,omitempty"`

	// A list of features enabled when the selected profile is CUSTOM. The method returns the set of features that can be specified in this list. This field must be empty if the profile is not CUSTOM.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.custom_features
	CustomFeatures []string `json:"customFeatures,omitempty"`

	// An optional description of this resource. Provide this property when you create the resource.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.description
	Description *string `json:"description,omitempty"`

	// [Output Only] The list of features enabled in the SSL policy.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.enabled_features
	EnabledFeatures []string `json:"enabledFeatures,omitempty"`

	// Fingerprint of this resource. A hash of the contents stored in this object. This field is used in optimistic locking. This field will be ignored when inserting a SslPolicy. An up-to-date fingerprint must be provided in order to update the SslPolicy, otherwise the request will fail with error 412 conditionNotMet. To see the latest fingerprint, make a get() request to retrieve an SslPolicy.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.fingerprint
	Fingerprint *string `json:"fingerprint,omitempty"`

	// [Output Only] The unique identifier for the resource. This identifier is defined by the server.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.id
	ID *uint64 `json:"id,omitempty"`

	// [Output only] Type of the resource. Always compute#sslPolicyfor SSL policies.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.kind
	Kind *string `json:"kind,omitempty"`

	// The minimum version of SSL protocol that can be used by the clients to establish a connection with the load balancer. This can be one of TLS_1_0, TLS_1_1, TLS_1_2.
	//  Check the MinTlsVersion enum for the list of possible values.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.min_tls_version
	MinTLSVersion *string `json:"minTLSVersion,omitempty"`

	// Profile specifies the set of SSL features that can be used by the load balancer when negotiating SSL with clients. This can be one of COMPATIBLE, MODERN, RESTRICTED, or CUSTOM. If using CUSTOM, the set of SSL features to enable must be specified in the customFeatures field.
	//  Check the Profile enum for the list of possible values.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.profile
	Profile *string `json:"profile,omitempty"`

	// [Output Only] URL of the region where the regional SSL policy resides. This field is not applicable to global SSL policies.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.region
	Region *string `json:"region,omitempty"`

	// [Output Only] Server-defined URL for the resource.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.self_link
	SelfLink *string `json:"selfLink,omitempty"`

	// [Output Only] If potential misconfigurations are detected for this SSL policy, this field will be populated with warning messages.
	// +kcc:proto:field=google.cloud.compute.v1.SslPolicy.warnings
	Warnings []Warnings `json:"warnings,omitempty"`
}

// ComputeRegionSSLPolicyStatus defines the config connector machine state of ComputeRegionSSLPolicy
type ComputeRegionSSLPolicyStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the ComputeRegionSSLPolicy resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *ComputeRegionSSLPolicyObservedState `json:"observedState,omitempty"`
}

// ComputeRegionSSLPolicyObservedState is the state of the ComputeRegionSSLPolicy resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.compute.v1.SslPolicy
type ComputeRegionSSLPolicyObservedState struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcomputeregionsslpolicy;gcpcomputeregionsslpolicys
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// ComputeRegionSSLPolicy is the Schema for the ComputeRegionSSLPolicy API
// +k8s:openapi-gen=true
type ComputeRegionSSLPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ComputeRegionSSLPolicySpec   `json:"spec,omitempty"`
	Status ComputeRegionSSLPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ComputeRegionSSLPolicyList contains a list of ComputeRegionSSLPolicy
type ComputeRegionSSLPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeRegionSSLPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeRegionSSLPolicy{}, &ComputeRegionSSLPolicyList{})
}
