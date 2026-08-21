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

var ConfigDeliveryFleetPackageGVK = GroupVersion.WithKind("ConfigDeliveryFleetPackage")

// ConfigDeliveryFleetPackageSpec defines the desired state of ConfigDeliveryFleetPackage
// +kcc:spec:proto=google.cloud.configdelivery.v1.FleetPackage
type ConfigDeliveryFleetPackageSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The ConfigDeliveryFleetPackage name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Labels are attributes that can be set and used by both the
	//  user and by Config Delivery. Labels must meet the following constraints:
	//
	//  * Keys and values can contain only lowercase letters, numeric characters,
	//  underscores, and dashes.
	//  * All characters must use UTF-8 encoding, and international characters are
	//  allowed.
	//  * Keys must start with a lowercase letter or international character.
	//  * Each resource is limited to a maximum of 64 labels.
	//
	//  Both keys and values are additionally constrained to be <= 128 bytes.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. Information specifying the source of kubernetes configuration to
	//  deploy.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.resource_bundle_selector
	// +required
	ResourceBundleSelector *FleetPackage_ResourceBundleSelector `json:"resourceBundleSelector,omitempty"`

	// Optional. Configuration to select target clusters to deploy kubernetes
	//  configuration to.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.target
	Target *FleetPackage_Target `json:"target,omitempty"`

	// Optional. The strategy to use to deploy kubernetes configuration to
	//  clusters.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.rollout_strategy
	RolloutStrategy *RolloutStrategy `json:"rolloutStrategy,omitempty"`

	// Required. Information specifying how to map a `ResourceBundle` variant to a
	//  target cluster.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.variant_selector
	// +required
	VariantSelector *FleetPackage_VariantSelector `json:"variantSelector,omitempty"`

	// Optional. Information around how to handle kubernetes resources at the
	//  target clusters when the `FleetPackage` is deleted.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.deletion_propagation_policy
	DeletionPropagationPolicy *string `json:"deletionPropagationPolicy,omitempty"`

	// Optional. The desired state of the fleet package.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.state
	State *string `json:"state,omitempty"`
}

// ConfigDeliveryFleetPackageStatus defines the config connector machine state of ConfigDeliveryFleetPackage
type ConfigDeliveryFleetPackageStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the ConfigDeliveryFleetPackage resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *ConfigDeliveryFleetPackageObservedState `json:"observedState,omitempty"`
}

// ConfigDeliveryFleetPackageObservedState is the state of the ConfigDeliveryFleetPackage resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.configdelivery.v1.FleetPackage
type ConfigDeliveryFleetPackageObservedState struct {
	// Output only. Time at which the `FleetPackage` was created.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Most recent time at which the `FleetPackage` was updated.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Information containing the rollout status of the
	//  `FleetPackage` across all the target clusters.
	// +kcc:proto:field=google.cloud.configdelivery.v1.FleetPackage.info
	Info *FleetPackageInfoObservedState `json:"info,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpconfigdeliveryfleetpackage;gcpconfigdeliveryfleetpackages
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// ConfigDeliveryFleetPackage is the Schema for the ConfigDeliveryFleetPackage API
// +k8s:openapi-gen=true
type ConfigDeliveryFleetPackage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ConfigDeliveryFleetPackageSpec   `json:"spec,omitempty"`
	Status ConfigDeliveryFleetPackageStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ConfigDeliveryFleetPackageList contains a list of ConfigDeliveryFleetPackage
type ConfigDeliveryFleetPackageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigDeliveryFleetPackage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigDeliveryFleetPackage{}, &ConfigDeliveryFleetPackageList{})
}
