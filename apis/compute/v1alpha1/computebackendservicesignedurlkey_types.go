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

var ComputeBackendServiceSignedURLKeyGVK = GroupVersion.WithKind("ComputeBackendServiceSignedURLKey")

// ComputeBackendServiceSignedURLKeySpec defines the desired state of ComputeBackendServiceSignedURLKey
// +kcc:spec:proto=google.cloud.compute.v1.SignedUrlKey
type ComputeBackendServiceSignedURLKeySpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The ComputeBackendServiceSignedURLKey name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Name of the key. The name must be 1-63 characters long, and comply with RFC1035. Specifically, the name must be 1-63 characters long and match the regular expression `[a-z]([-a-z0-9]*[a-z0-9])?` which means the first character must be a lowercase letter, and all following characters must be a dash, lowercase letter, or digit, except the last character, which cannot be a dash.
	// +kcc:proto:field=google.cloud.compute.v1.SignedUrlKey.key_name
	KeyName *string `json:"keyName,omitempty"`

	// 128-bit key value used for signing the URL. The key value must be a valid RFC 4648 Section 5 base64url encoded string.
	// +kcc:proto:field=google.cloud.compute.v1.SignedUrlKey.key_value
	KeyValue *string `json:"keyValue,omitempty"`
}

// ComputeBackendServiceSignedURLKeyStatus defines the config connector machine state of ComputeBackendServiceSignedURLKey
type ComputeBackendServiceSignedURLKeyStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the ComputeBackendServiceSignedURLKey resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *ComputeBackendServiceSignedURLKeyObservedState `json:"observedState,omitempty"`
}

// ComputeBackendServiceSignedURLKeyObservedState is the state of the ComputeBackendServiceSignedURLKey resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.compute.v1.SignedUrlKey
type ComputeBackendServiceSignedURLKeyObservedState struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcomputebackendservicesignedurlkey;gcpcomputebackendservicesignedurlkeys
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// ComputeBackendServiceSignedURLKey is the Schema for the ComputeBackendServiceSignedURLKey API
// +k8s:openapi-gen=true
type ComputeBackendServiceSignedURLKey struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ComputeBackendServiceSignedURLKeySpec   `json:"spec,omitempty"`
	Status ComputeBackendServiceSignedURLKeyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ComputeBackendServiceSignedURLKeyList contains a list of ComputeBackendServiceSignedURLKey
type ComputeBackendServiceSignedURLKeyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeBackendServiceSignedURLKey `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeBackendServiceSignedURLKey{}, &ComputeBackendServiceSignedURLKeyList{})
}
