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

var SecureSourceManagerInstanceGVK = GroupVersion.WithKind("SecureSourceManagerInstance")

// SecureSourceManagerInstanceSpec defines the desired state of SecureSourceManagerInstance
// +kcc:spec:proto=google.cloud.securesourcemanager.v1.Instance
type SecureSourceManagerInstanceSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/instances/{instance}
	Location *string `json:"location"`

	// The SecureSourceManagerInstance name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Labels as key value pairs.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. Private settings for private instance.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.private_config
	PrivateConfig *Instance_PrivateConfig `json:"privateConfig,omitempty"`

	// Optional. Immutable. Customer-managed encryption key name, in the format
	//  projects/*/locations/*/keyRings/*/cryptoKeys/*.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.kms_key
	KMSKey *string `json:"kmsKey,omitempty"`

	// Optional. Configuration for Workforce Identity Federation to support
	//  third party identity provider. If unset, defaults to the Google OIDC IdP.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.workforce_identity_federation_config
	WorkforceIdentityFederationConfig *Instance_WorkforceIdentityFederationConfig `json:"workforceIdentityFederationConfig,omitempty"`
}

// SecureSourceManagerInstanceStatus defines the config connector machine state of SecureSourceManagerInstance
type SecureSourceManagerInstanceStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the SecureSourceManagerInstance resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *SecureSourceManagerInstanceObservedState `json:"observedState,omitempty"`
}

// SecureSourceManagerInstanceObservedState is the state of the SecureSourceManagerInstance resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.securesourcemanager.v1.Instance
type SecureSourceManagerInstanceObservedState struct {
	// Output only. Create timestamp.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Update timestamp.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Optional. Private settings for private instance.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.private_config
	PrivateConfig *Instance_PrivateConfigObservedState `json:"privateConfig,omitempty"`

	// Output only. Current state of the instance.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.state
	State *string `json:"state,omitempty"`

	// Output only. An optional field providing information about the current
	//  instance state.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.state_note
	StateNote *string `json:"stateNote,omitempty"`

	// Output only. A list of hostnames for this instance.
	// +kcc:proto:field=google.cloud.securesourcemanager.v1.Instance.host_config
	HostConfig *Instance_HostConfigObservedState `json:"hostConfig,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpsecuresourcemanagerinstance;gcpsecuresourcemanagerinstances
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// SecureSourceManagerInstance is the Schema for the SecureSourceManagerInstance API
// +k8s:openapi-gen=true
type SecureSourceManagerInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   SecureSourceManagerInstanceSpec   `json:"spec,omitempty"`
	Status SecureSourceManagerInstanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// SecureSourceManagerInstanceList contains a list of SecureSourceManagerInstance
type SecureSourceManagerInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecureSourceManagerInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecureSourceManagerInstance{}, &SecureSourceManagerInstanceList{})
}
