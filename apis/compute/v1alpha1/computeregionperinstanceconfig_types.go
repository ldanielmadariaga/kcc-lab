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

var ComputeRegionPerInstanceConfigGVK = GroupVersion.WithKind("ComputeRegionPerInstanceConfig")

// ComputeRegionPerInstanceConfigSpec defines the desired state of ComputeRegionPerInstanceConfig
// +kcc:spec:proto=google.cloud.compute.v1.PerInstanceConfig
type ComputeRegionPerInstanceConfigSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`


	// The ComputeRegionPerInstanceConfig name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Fingerprint of this per-instance config. This field can be used in optimistic locking. It is ignored when inserting a per-instance config. An up-to-date fingerprint must be provided in order to update an existing per-instance configuration or the field needs to be unset.
	// +kcc:proto:field=google.cloud.compute.v1.PerInstanceConfig.fingerprint
	Fingerprint *string `json:"fingerprint,omitempty"`

	// The intended preserved state for the given instance. Does not contain preserved state generated from a stateful policy.
	// +kcc:proto:field=google.cloud.compute.v1.PerInstanceConfig.preserved_state
	PreservedState *PreservedState `json:"preservedState,omitempty"`

	// The status of applying this per-instance configuration on the corresponding managed instance.
	//  Check the Status enum for the list of possible values.
	// +kcc:proto:field=google.cloud.compute.v1.PerInstanceConfig.status
	Status *string `json:"status,omitempty"`
}

// ComputeRegionPerInstanceConfigStatus defines the config connector machine state of ComputeRegionPerInstanceConfig
type ComputeRegionPerInstanceConfigStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the ComputeRegionPerInstanceConfig resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *ComputeRegionPerInstanceConfigObservedState `json:"observedState,omitempty"`
}

// ComputeRegionPerInstanceConfigObservedState is the state of the ComputeRegionPerInstanceConfig resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.compute.v1.PerInstanceConfig
type ComputeRegionPerInstanceConfigObservedState struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcomputeregionperinstanceconfig;gcpcomputeregionperinstanceconfigs
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// ComputeRegionPerInstanceConfig is the Schema for the ComputeRegionPerInstanceConfig API
// +k8s:openapi-gen=true
type ComputeRegionPerInstanceConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ComputeRegionPerInstanceConfigSpec   `json:"spec,omitempty"`
	Status ComputeRegionPerInstanceConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ComputeRegionPerInstanceConfigList contains a list of ComputeRegionPerInstanceConfig
type ComputeRegionPerInstanceConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeRegionPerInstanceConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeRegionPerInstanceConfig{}, &ComputeRegionPerInstanceConfigList{})
}
