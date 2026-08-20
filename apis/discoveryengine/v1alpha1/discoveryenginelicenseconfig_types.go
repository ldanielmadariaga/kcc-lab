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

var DiscoveryEngineLicenseConfigGVK = GroupVersion.WithKind("DiscoveryEngineLicenseConfig")

// DiscoveryEngineLicenseConfigSpec defines the desired state of DiscoveryEngineLicenseConfig
// +kcc:spec:proto=google.cloud.discoveryengine.v1beta.LicenseConfig
type DiscoveryEngineLicenseConfigSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location string `json:"location"`

	// The DiscoveryEngineLicenseConfig name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Required. Number of licenses purchased.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.license_count
	// +required
	LicenseCount *int64 `json:"licenseCount,omitempty"`

	// Required. Subscription tier information for the license config.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.subscription_tier
	// +required
	SubscriptionTier *string `json:"subscriptionTier,omitempty"`

	// Optional. Whether the license config should be auto renewed when it reaches
	//  the end date.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.auto_renew
	AutoRenew *bool `json:"autoRenew,omitempty"`

	// Required. The start date.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.start_date
	// +required
	StartDate *Date `json:"startDate,omitempty"`

	// Optional. The planed end date.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.end_date
	EndDate *Date `json:"endDate,omitempty"`

	// Required. Subscription term.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.subscription_term
	// +required
	SubscriptionTerm *string `json:"subscriptionTerm,omitempty"`

	// Optional. Whether the license config is for free trial.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.free_trial
	FreeTrial *bool `json:"freeTrial,omitempty"`
}

// DiscoveryEngineLicenseConfigStatus defines the config connector machine state of DiscoveryEngineLicenseConfig
type DiscoveryEngineLicenseConfigStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DiscoveryEngineLicenseConfig resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DiscoveryEngineLicenseConfigObservedState `json:"observedState,omitempty"`
}

// DiscoveryEngineLicenseConfigObservedState is the state of the DiscoveryEngineLicenseConfig resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.discoveryengine.v1beta.LicenseConfig
type DiscoveryEngineLicenseConfigObservedState struct {
	// Output only. The state of the license config.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.state
	State *string `json:"state,omitempty"`

	// Output only. Whether the license config is for Gemini bundle.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.gemini_bundle
	GeminiBundle *bool `json:"geminiBundle,omitempty"`

	// Output only. Indication of whether the subscription is terminated earlier
	//  than the expiration date. This is usually terminated by pipeline once the
	//  subscription gets terminated from subsv3.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.early_terminated
	EarlyTerminated *bool `json:"earlyTerminated,omitempty"`

	// Output only. The date when the subscription is terminated earlier than the
	//  expiration date.
	// +kcc:proto:field=google.cloud.discoveryengine.v1beta.LicenseConfig.early_termination_date
	EarlyTerminationDate *Date `json:"earlyTerminationDate,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdiscoveryenginelicenseconfig;gcpdiscoveryenginelicenseconfigs
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DiscoveryEngineLicenseConfig is the Schema for the DiscoveryEngineLicenseConfig API
// +k8s:openapi-gen=true
type DiscoveryEngineLicenseConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DiscoveryEngineLicenseConfigSpec   `json:"spec,omitempty"`
	Status DiscoveryEngineLicenseConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DiscoveryEngineLicenseConfigList contains a list of DiscoveryEngineLicenseConfig
type DiscoveryEngineLicenseConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiscoveryEngineLicenseConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiscoveryEngineLicenseConfig{}, &DiscoveryEngineLicenseConfigList{})
}
