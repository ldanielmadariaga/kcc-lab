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

var ComputeAutoscalerGVK = GroupVersion.WithKind("ComputeAutoscaler")

// ComputeAutoscalerSpec defines the desired state of ComputeAutoscaler
// +kcc:spec:proto=google.cloud.compute.v1.Autoscaler
type ComputeAutoscalerSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`


	// The ComputeAutoscaler name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The configuration parameters for the autoscaling algorithm. You can define one or more signals for an autoscaler: cpuUtilization, customMetricUtilizations, and loadBalancingUtilization. If none of these are specified, the default will be to autoscale based on cpuUtilization to 0.6 or 60%.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.autoscaling_policy
	AutoscalingPolicy *AutoscalingPolicy `json:"autoscalingPolicy,omitempty"`

	// [Output Only] Creation timestamp in RFC3339 text format.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.creation_timestamp
	CreationTimestamp *string `json:"creationTimestamp,omitempty"`

	// An optional description of this resource. Provide this property when you create the resource.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.description
	Description *string `json:"description,omitempty"`

	// [Output Only] The unique identifier for the resource. This identifier is defined by the server.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.id
	ID *uint64 `json:"id,omitempty"`

	// [Output Only] Type of the resource. Always compute#autoscaler for autoscalers.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.kind
	Kind *string `json:"kind,omitempty"`

	// [Output Only] Target recommended MIG size (number of instances) computed by autoscaler. Autoscaler calculates the recommended MIG size even when the autoscaling policy mode is different from ON. This field is empty when autoscaler is not connected to an existing managed instance group or autoscaler did not generate its prediction.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.recommended_size
	RecommendedSize *int32 `json:"recommendedSize,omitempty"`

	// [Output Only] URL of the region where the instance group resides (for autoscalers living in regional scope).
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.region
	Region *string `json:"region,omitempty"`

	// TODO: unsupported map type with key string and value message

	// [Output Only] Server-defined URL for the resource.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.self_link
	SelfLink *string `json:"selfLink,omitempty"`

	// [Output Only] The status of the autoscaler configuration. Current set of possible values: - PENDING: Autoscaler backend hasn't read new/updated configuration. - DELETING: Configuration is being deleted. - ACTIVE: Configuration is acknowledged to be effective. Some warnings might be present in the statusDetails field. - ERROR: Configuration has errors. Actionable for users. Details are present in the statusDetails field. New values might be added in the future.
	//  Check the Status enum for the list of possible values.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.status
	Status *string `json:"status,omitempty"`

	// [Output Only] Human-readable details about the current state of the autoscaler. Read the documentation for Commonly returned status messages for examples of status messages you might encounter.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.status_details
	StatusDetails []AutoscalerStatusDetails `json:"statusDetails,omitempty"`

	// URL of the managed instance group that this autoscaler will scale. This field is required when creating an autoscaler.
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.target
	Target *string `json:"target,omitempty"`

	// [Output Only] URL of the zone where the instance group resides (for autoscalers living in zonal scope).
	// +kcc:proto:field=google.cloud.compute.v1.Autoscaler.zone
	Zone *string `json:"zone,omitempty"`
}

// ComputeAutoscalerStatus defines the config connector machine state of ComputeAutoscaler
type ComputeAutoscalerStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the ComputeAutoscaler resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *ComputeAutoscalerObservedState `json:"observedState,omitempty"`
}

// ComputeAutoscalerObservedState is the state of the ComputeAutoscaler resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.compute.v1.Autoscaler
type ComputeAutoscalerObservedState struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcomputeautoscaler;gcpcomputeautoscalers
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// ComputeAutoscaler is the Schema for the ComputeAutoscaler API
// +k8s:openapi-gen=true
type ComputeAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ComputeAutoscalerSpec   `json:"spec,omitempty"`
	Status ComputeAutoscalerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ComputeAutoscalerList contains a list of ComputeAutoscaler
type ComputeAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeAutoscaler{}, &ComputeAutoscalerList{})
}
