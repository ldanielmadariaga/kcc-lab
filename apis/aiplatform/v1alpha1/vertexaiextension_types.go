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

var VertexAIExtensionGVK = GroupVersion.WithKind("VertexAIExtension")

// VertexAIExtensionSpec defines the desired state of VertexAIExtension
// +kcc:spec:proto=google.cloud.aiplatform.v1.Extension
type VertexAIExtensionSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The VertexAIExtension name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Required. The display name of the Extension.
	//  The name can be up to 128 characters long and can consist of any UTF-8
	//  characters.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. The description of the Extension.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.description
	Description *string `json:"description,omitempty"`

	// Optional. Used to perform consistent read-modify-write updates. If not set,
	//  a blind "overwrite" update happens.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.etag
	Etag *string `json:"etag,omitempty"`

	// Required. Manifest of the Extension.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.manifest
	// +required
	Manifest *ExtensionManifest `json:"manifest,omitempty"`

	// Optional. Runtime config controlling the runtime behavior of this
	//  Extension.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.runtime_config
	RuntimeConfig *RuntimeConfig `json:"runtimeConfig,omitempty"`

	// Optional. Examples to illustrate the usage of the extension as a tool.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.tool_use_examples
	ToolUseExamples []ToolUseExample `json:"toolUseExamples,omitempty"`

	// Optional. The PrivateServiceConnect config for the extension.
	//  If specified, the service endpoints associated with the
	//  Extension should be registered with private network access in the provided
	//  Service Directory
	//  (https://cloud.google.com/service-directory/docs/configuring-private-network-access).
	//
	//  If the service contains more than one endpoint with a network, the service
	//  will arbitrarilty choose one of the endpoints to use for extension
	//  execution.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.private_service_connect_config
	PrivateServiceConnectConfig *ExtensionPrivateServiceConnectConfig `json:"privateServiceConnectConfig,omitempty"`
}

// VertexAIExtensionStatus defines the config connector machine state of VertexAIExtension
type VertexAIExtensionStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the VertexAIExtension resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *VertexAIExtensionObservedState `json:"observedState,omitempty"`
}

// VertexAIExtensionObservedState is the state of the VertexAIExtension resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.aiplatform.v1.Extension
type VertexAIExtensionObservedState struct {
	// Output only. Timestamp when this Extension was created.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Timestamp when this Extension was most recently updated.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Supported operations.
	// +kcc:proto:field=google.cloud.aiplatform.v1beta1.Extension.extension_operations
	ExtensionOperations []ExtensionOperationObservedState `json:"extensionOperations,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpvertexaiextension;gcpvertexaiextensions
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// VertexAIExtension is the Schema for the VertexAIExtension API
// +k8s:openapi-gen=true
type VertexAIExtension struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   VertexAIExtensionSpec   `json:"spec,omitempty"`
	Status VertexAIExtensionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VertexAIExtensionList contains a list of VertexAIExtension
type VertexAIExtensionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VertexAIExtension `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VertexAIExtension{}, &VertexAIExtensionList{})
}
