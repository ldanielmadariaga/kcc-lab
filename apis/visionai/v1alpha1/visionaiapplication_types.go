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

var VisionAIApplicationGVK = GroupVersion.WithKind("VisionAIApplication")

// VisionAIApplicationSpec defines the desired state of VisionAIApplication
// +kcc:spec:proto=google.cloud.visionai.v1.Application
type VisionAIApplicationSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/applications/{application}
	Location *string `json:"location"`

	// The VisionAIApplication name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Labels as key value pairs
	// +kcc:proto:field=google.cloud.visionai.v1.Application.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. A user friendly display name for the solution.
	// +kcc:proto:field=google.cloud.visionai.v1.Application.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// A description for this application.
	// +kcc:proto:field=google.cloud.visionai.v1.Application.description
	Description *string `json:"description,omitempty"`

	// Application graph configuration.
	// +kcc:proto:field=google.cloud.visionai.v1.Application.application_configs
	ApplicationConfigs *ApplicationConfigs `json:"applicationConfigs,omitempty"`

	// Billing mode of the application.
	// +kcc:proto:field=google.cloud.visionai.v1.Application.billing_mode
	BillingMode *string `json:"billingMode,omitempty"`
}

// VisionAIApplicationStatus defines the config connector machine state of VisionAIApplication
type VisionAIApplicationStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the VisionAIApplication resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *VisionAIApplicationObservedState `json:"observedState,omitempty"`
}

// VisionAIApplicationObservedState is the state of the VisionAIApplication resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.visionai.v1.Application
type VisionAIApplicationObservedState struct {
	// Output only. [Output only] Create timestamp
	// +kcc:proto:field=google.cloud.visionai.v1.Application.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. [Output only] Update timestamp
	// +kcc:proto:field=google.cloud.visionai.v1.Application.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Application graph runtime info. Only exists when application
	//  state equals to DEPLOYED.
	// +kcc:proto:field=google.cloud.visionai.v1.Application.runtime_info
	RuntimeInfo *Application_ApplicationRuntimeInfo `json:"runtimeInfo,omitempty"`

	// Output only. State of the application.
	// +kcc:proto:field=google.cloud.visionai.v1.Application.state
	State *string `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpvisionaiapplication;gcpvisionaiapplications
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// VisionAIApplication is the Schema for the VisionAIApplication API
// +k8s:openapi-gen=true
type VisionAIApplication struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   VisionAIApplicationSpec   `json:"spec,omitempty"`
	Status VisionAIApplicationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VisionAIApplicationList contains a list of VisionAIApplication
type VisionAIApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VisionAIApplication `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VisionAIApplication{}, &VisionAIApplicationList{})
}
