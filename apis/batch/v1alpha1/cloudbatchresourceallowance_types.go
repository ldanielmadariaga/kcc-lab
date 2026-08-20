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

var CloudBatchResourceAllowanceGVK = GroupVersion.WithKind("CloudBatchResourceAllowance")

// CloudBatchResourceAllowanceSpec defines the desired state of CloudBatchResourceAllowance
// +kcc:spec:proto=google.cloud.batch.v1alpha.ResourceAllowance
type CloudBatchResourceAllowanceSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The CloudBatchResourceAllowance name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The detail of usage resource allowance.
	// +kcc:proto:field=google.cloud.batch.v1alpha.ResourceAllowance.usage_resource_allowance
	UsageResourceAllowance *UsageResourceAllowance `json:"usageResourceAllowance,omitempty"`

	// Optional. Labels are attributes that can be set and used by both the
	//  user and by Batch. Labels must meet the following constraints:
	//
	//  * Keys and values can contain only lowercase letters, numeric characters,
	//  underscores, and dashes.
	//  * All characters must use UTF-8 encoding, and international characters are
	//  allowed.
	//  * Keys must start with a lowercase letter or international character.
	//  * Each resource is limited to a maximum of 64 labels.
	//
	//  Both keys and values are additionally constrained to be <= 128 bytes.
	// +kcc:proto:field=google.cloud.batch.v1alpha.ResourceAllowance.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. Notification configurations.
	// +kcc:proto:field=google.cloud.batch.v1alpha.ResourceAllowance.notifications
	Notifications []Notification `json:"notifications,omitempty"`
}

// CloudBatchResourceAllowanceStatus defines the config connector machine state of CloudBatchResourceAllowance
type CloudBatchResourceAllowanceStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the CloudBatchResourceAllowance resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *CloudBatchResourceAllowanceObservedState `json:"observedState,omitempty"`
}

// CloudBatchResourceAllowanceObservedState is the state of the CloudBatchResourceAllowance resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.batch.v1alpha.ResourceAllowance
type CloudBatchResourceAllowanceObservedState struct {
	// The detail of usage resource allowance.
	// +kcc:proto:field=google.cloud.batch.v1alpha.ResourceAllowance.usage_resource_allowance
	UsageResourceAllowance *UsageResourceAllowanceObservedState `json:"usageResourceAllowance,omitempty"`

	// Output only. A system generated unique ID (in UUID4 format) for the
	//  ResourceAllowance.
	// +kcc:proto:field=google.cloud.batch.v1alpha.ResourceAllowance.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. Time when the ResourceAllowance was created.
	// +kcc:proto:field=google.cloud.batch.v1alpha.ResourceAllowance.create_time
	CreateTime *string `json:"createTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcloudbatchresourceallowance;gcpcloudbatchresourceallowances
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// CloudBatchResourceAllowance is the Schema for the CloudBatchResourceAllowance API
// +k8s:openapi-gen=true
type CloudBatchResourceAllowance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   CloudBatchResourceAllowanceSpec   `json:"spec,omitempty"`
	Status CloudBatchResourceAllowanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// CloudBatchResourceAllowanceList contains a list of CloudBatchResourceAllowance
type CloudBatchResourceAllowanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudBatchResourceAllowance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudBatchResourceAllowance{}, &CloudBatchResourceAllowanceList{})
}
