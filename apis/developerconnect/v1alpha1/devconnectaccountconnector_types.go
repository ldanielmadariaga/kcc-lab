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

var DevConnectAccountConnectorGVK = GroupVersion.WithKind("DevConnectAccountConnector")

// DevConnectAccountConnectorSpec defines the desired state of DevConnectAccountConnector
// +kcc:spec:proto=google.cloud.developerconnect.v1.AccountConnector
type DevConnectAccountConnectorSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/accountConnectors/{account_connector}
	Location *string `json:"location"`

	// The DevConnectAccountConnector name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Provider OAuth config.
	// +kcc:proto:field=google.cloud.developerconnect.v1.AccountConnector.provider_oauth_config
	ProviderOauthConfig *ProviderOAuthConfig `json:"providerOauthConfig,omitempty"`

	// Optional. Allows users to store small amounts of arbitrary data.
	// +kcc:proto:field=google.cloud.developerconnect.v1.AccountConnector.annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Optional. This checksum is computed by the server based on the value of
	//  other fields, and may be sent on update and delete requests to ensure the
	//  client has an up-to-date value before proceeding.
	// +kcc:proto:field=google.cloud.developerconnect.v1.AccountConnector.etag
	Etag *string `json:"etag,omitempty"`

	// Optional. Labels as key value pairs
	// +kcc:proto:field=google.cloud.developerconnect.v1.AccountConnector.labels
	Labels map[string]string `json:"labels,omitempty"`
}

// DevConnectAccountConnectorStatus defines the config connector machine state of DevConnectAccountConnector
type DevConnectAccountConnectorStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DevConnectAccountConnector resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DevConnectAccountConnectorObservedState `json:"observedState,omitempty"`
}

// DevConnectAccountConnectorObservedState is the state of the DevConnectAccountConnector resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.developerconnect.v1.AccountConnector
type DevConnectAccountConnectorObservedState struct {
	// Output only. The timestamp when the accountConnector was created.
	// +kcc:proto:field=google.cloud.developerconnect.v1.AccountConnector.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The timestamp when the accountConnector was updated.
	// +kcc:proto:field=google.cloud.developerconnect.v1.AccountConnector.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Start OAuth flow by clicking on this URL.
	// +kcc:proto:field=google.cloud.developerconnect.v1.AccountConnector.oauth_start_uri
	OauthStartURI *string `json:"oauthStartURI,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdevconnectaccountconnector;gcpdevconnectaccountconnectors
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DevConnectAccountConnector is the Schema for the DevConnectAccountConnector API
// +k8s:openapi-gen=true
type DevConnectAccountConnector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DevConnectAccountConnectorSpec   `json:"spec,omitempty"`
	Status DevConnectAccountConnectorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DevConnectAccountConnectorList contains a list of DevConnectAccountConnector
type DevConnectAccountConnectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DevConnectAccountConnector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DevConnectAccountConnector{}, &DevConnectAccountConnectorList{})
}
