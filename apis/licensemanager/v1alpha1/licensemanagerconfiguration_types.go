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

var LicenseManagerConfigurationGVK = GroupVersion.WithKind("LicenseManagerConfiguration")

// LicenseManagerConfigurationSpec defines the desired state of LicenseManagerConfiguration
// +kcc:spec:proto=google.cloud.licensemanager.v1.Configuration
type LicenseManagerConfigurationSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location string `json:"location"`

	// The LicenseManagerConfiguration name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Required. User given name.
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// Required. Name field (with URL) of the Product offered for SPLA.
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.product
	// +required
	Product *string `json:"product,omitempty"`

	// Required. LicenseType to be applied for billing
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.license_type
	// +required
	LicenseType *string `json:"licenseType,omitempty"`

	// Required. Billing information applicable till end of the current month.
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.current_billing_info
	// +required
	CurrentBillingInfo *BillingInfo `json:"currentBillingInfo,omitempty"`

	// Required. Billing information applicable for next month.
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.next_billing_info
	// +required
	NextBillingInfo *BillingInfo `json:"nextBillingInfo,omitempty"`

	// Optional. Labels as key value pairs
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.labels
	Labels map[string]string `json:"labels,omitempty"`
}

// LicenseManagerConfigurationStatus defines the config connector machine state of LicenseManagerConfiguration
type LicenseManagerConfigurationStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the LicenseManagerConfiguration resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *LicenseManagerConfigurationObservedState `json:"observedState,omitempty"`
}

// LicenseManagerConfigurationObservedState is the state of the LicenseManagerConfiguration resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.licensemanager.v1.Configuration
type LicenseManagerConfigurationObservedState struct {
	// Required. Billing information applicable till end of the current month.
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.current_billing_info
	CurrentBillingInfo *BillingInfoObservedState `json:"currentBillingInfo,omitempty"`

	// Required. Billing information applicable for next month.
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.next_billing_info
	NextBillingInfo *BillingInfoObservedState `json:"nextBillingInfo,omitempty"`

	// Output only. [Output only] Create time stamp
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. [Output only] Update time stamp
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. State of the configuration.
	// +kcc:proto:field=google.cloud.licensemanager.v1.Configuration.state
	State *string `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcplicensemanagerconfiguration;gcplicensemanagerconfigurations
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// LicenseManagerConfiguration is the Schema for the LicenseManagerConfiguration API
// +k8s:openapi-gen=true
type LicenseManagerConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   LicenseManagerConfigurationSpec   `json:"spec,omitempty"`
	Status LicenseManagerConfigurationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// LicenseManagerConfigurationList contains a list of LicenseManagerConfiguration
type LicenseManagerConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LicenseManagerConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LicenseManagerConfiguration{}, &LicenseManagerConfigurationList{})
}
