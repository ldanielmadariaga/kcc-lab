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

var MapManagementMapConfigGVK = GroupVersion.WithKind("MapManagementMapConfig")

// MapManagementMapConfigSpec defines the desired state of MapManagementMapConfig
// +kcc:spec:proto=google.maps.mapmanagement.v2beta.MapConfig
type MapManagementMapConfigSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`


	// The MapManagementMapConfig name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. The display name of this MapConfig, as specified by the user.
	// +kcc:proto:field=google.maps.mapmanagement.v2beta.MapConfig.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. The description of this MapConfig, as specified by the user.
	// +kcc:proto:field=google.maps.mapmanagement.v2beta.MapConfig.description
	Description *string `json:"description,omitempty"`

	// Optional. The Map Features that apply to this Map Config.
	// +kcc:proto:field=google.maps.mapmanagement.v2beta.MapConfig.map_features
	MapFeatures *MapFeatures `json:"mapFeatures,omitempty"`

	// Optional. Represents the Map Type of the MapConfig. If this is unset, the
	//  default behavior is to use the raster map type.
	// +kcc:proto:field=google.maps.mapmanagement.v2beta.MapConfig.map_type
	MapType *string `json:"mapType,omitempty"`
}

// MapManagementMapConfigStatus defines the config connector machine state of MapManagementMapConfig
type MapManagementMapConfigStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the MapManagementMapConfig resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *MapManagementMapConfigObservedState `json:"observedState,omitempty"`
}

// MapManagementMapConfigObservedState is the state of the MapManagementMapConfig resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.maps.mapmanagement.v2beta.MapConfig
type MapManagementMapConfigObservedState struct {
	// Output only. The Map ID of this MapConfig, used to identify the map in
	//  client applications. This read-only field is generated when the MapConfig
	//  is created. Output only.
	// +kcc:proto:field=google.maps.mapmanagement.v2beta.MapConfig.map_id
	MapID *string `json:"mapID,omitempty"`

	// Output only. Denotes the creation time of the Map Config. Output only.
	// +kcc:proto:field=google.maps.mapmanagement.v2beta.MapConfig.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Denotes the last update time of the Map Config. Output only.
	// +kcc:proto:field=google.maps.mapmanagement.v2beta.MapConfig.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpmapmanagementmapconfig;gcpmapmanagementmapconfigs
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// MapManagementMapConfig is the Schema for the MapManagementMapConfig API
// +k8s:openapi-gen=true
type MapManagementMapConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   MapManagementMapConfigSpec   `json:"spec,omitempty"`
	Status MapManagementMapConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MapManagementMapConfigList contains a list of MapManagementMapConfig
type MapManagementMapConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MapManagementMapConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MapManagementMapConfig{}, &MapManagementMapConfigList{})
}
