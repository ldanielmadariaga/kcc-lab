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

package v1beta1

import (
	refsv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/apis/refs/v1beta1"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var KMSAutokeyConfigGVK = GroupVersion.WithKind("KMSAutokeyConfig")

// KMSAutokeyConfigSpec defines the desired state of KMSAutokeyConfig
// +kcc:spec:proto=google.cloud.kms.v1.AutokeyConfig
type KMSAutokeyConfigSpec struct {
	// The folder that this resource belongs to.
	FolderRef *refsv1beta1.FolderRef `json:"folderRef"`

	// The KMSAutokeyConfig name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Name of the key project, e.g. `projects/{PROJECT_ID}` or
	//  `projects/{PROJECT_NUMBER}`, where Cloud KMS Autokey will provision a new
	//  [CryptoKey][google.cloud.kms.v1.CryptoKey] when a
	//  [KeyHandle][google.cloud.kms.v1.KeyHandle] is created. On
	//  [UpdateAutokeyConfig][google.cloud.kms.v1.AutokeyAdmin.UpdateAutokeyConfig],
	//  the caller will require `cloudkms.cryptoKeys.setIamPolicy` permission on
	//  this key project. Once configured, for Cloud KMS Autokey to function
	//  properly, this key project must have the Cloud KMS API activated and the
	//  Cloud KMS Service Agent for this key project must be granted the
	//  `cloudkms.admin` role (or pertinent permissions). A request with an empty
	//  key project field will clear the configuration.
	// +kcc:proto:field=google.cloud.kms.v1.AutokeyConfig.key_project
	KeyProject *string `json:"keyProject,omitempty"`

	// Optional. A checksum computed by the server based on the value of other
	//  fields. This may be sent on update requests to ensure that the client has
	//  an up-to-date value before proceeding. The request will be rejected with an
	//  ABORTED error on a mismatched etag.
	// +kcc:proto:field=google.cloud.kms.v1.AutokeyConfig.etag
	Etag *string `json:"etag,omitempty"`
}

// KMSAutokeyConfigStatus defines the config connector machine state of KMSAutokeyConfig
type KMSAutokeyConfigStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the KMSAutokeyConfig resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *KMSAutokeyConfigObservedState `json:"observedState,omitempty"`
}

// KMSAutokeyConfigObservedState is the state of the KMSAutokeyConfig resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.kms.v1.AutokeyConfig
type KMSAutokeyConfigObservedState struct {
	// Output only. The state for the AutokeyConfig.
	// +kcc:proto:field=google.cloud.kms.v1.AutokeyConfig.state
	State *string `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpkmsautokeyconfig;gcpkmsautokeyconfigs
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// KMSAutokeyConfig is the Schema for the KMSAutokeyConfig API
// +k8s:openapi-gen=true
type KMSAutokeyConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   KMSAutokeyConfigSpec   `json:"spec,omitempty"`
	Status KMSAutokeyConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// KMSAutokeyConfigList contains a list of KMSAutokeyConfig
type KMSAutokeyConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KMSAutokeyConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KMSAutokeyConfig{}, &KMSAutokeyConfigList{})
}
