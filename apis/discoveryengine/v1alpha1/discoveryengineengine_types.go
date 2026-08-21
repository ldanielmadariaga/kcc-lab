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

var DiscoveryEngineEngineGVK = GroupVersion.WithKind("DiscoveryEngineEngine")

// DiscoveryEngineEngineSpec defines the desired state of DiscoveryEngineEngine
// +kcc:spec:proto=google.cloud.discoveryengine.v1.Engine
type DiscoveryEngineEngineSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The DiscoveryEngineEngine name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Configurations for the Chat Engine. Only applicable if
	//  [solution_type][google.cloud.discoveryengine.v1.Engine.solution_type] is
	//  [SOLUTION_TYPE_CHAT][google.cloud.discoveryengine.v1.SolutionType.SOLUTION_TYPE_CHAT].
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.chat_engine_config
	ChatEngineConfig *Engine_ChatEngineConfig `json:"chatEngineConfig,omitempty"`

	// Configurations for the Search Engine. Only applicable if
	//  [solution_type][google.cloud.discoveryengine.v1.Engine.solution_type] is
	//  [SOLUTION_TYPE_SEARCH][google.cloud.discoveryengine.v1.SolutionType.SOLUTION_TYPE_SEARCH].
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.search_engine_config
	SearchEngineConfig *Engine_SearchEngineConfig `json:"searchEngineConfig,omitempty"`

	// Configurations for the Media Engine. Only applicable on the data
	//  stores with
	//  [solution_type][google.cloud.discoveryengine.v1.Engine.solution_type]
	//  [SOLUTION_TYPE_RECOMMENDATION][google.cloud.discoveryengine.v1.SolutionType.SOLUTION_TYPE_RECOMMENDATION]
	//  and
	//  [IndustryVertical.MEDIA][google.cloud.discoveryengine.v1.IndustryVertical.MEDIA]
	//  vertical.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.media_recommendation_engine_config
	MediaRecommendationEngineConfig *Engine_MediaRecommendationEngineConfig `json:"mediaRecommendationEngineConfig,omitempty"`

	// Required. The display name of the engine. Should be human readable. UTF-8
	//  encoded string with limit of 1024 characters.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. The data stores associated with this engine.
	//
	//  For
	//  [SOLUTION_TYPE_SEARCH][google.cloud.discoveryengine.v1.SolutionType.SOLUTION_TYPE_SEARCH]
	//  and
	//  [SOLUTION_TYPE_RECOMMENDATION][google.cloud.discoveryengine.v1.SolutionType.SOLUTION_TYPE_RECOMMENDATION]
	//  type of engines, they can only associate with at most one data store.
	//
	//  If [solution_type][google.cloud.discoveryengine.v1.Engine.solution_type] is
	//  [SOLUTION_TYPE_CHAT][google.cloud.discoveryengine.v1.SolutionType.SOLUTION_TYPE_CHAT],
	//  multiple [DataStore][google.cloud.discoveryengine.v1.DataStore]s in the
	//  same [Collection][google.cloud.discoveryengine.v1.Collection] can be
	//  associated here.
	//
	//  Note that when used in
	//  [CreateEngineRequest][google.cloud.discoveryengine.v1.CreateEngineRequest],
	//  one DataStore id must be provided as the system will use it for necessary
	//  initializations.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.data_store_ids
	DataStoreIDs []string `json:"dataStoreIDs,omitempty"`

	// Required. The solutions of the engine.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.solution_type
	// +required
	SolutionType *string `json:"solutionType,omitempty"`

	// Optional. The industry vertical that the engine registers.
	//  The restriction of the Engine industry vertical is based on
	//  [DataStore][google.cloud.discoveryengine.v1.DataStore]: Vertical on Engine
	//  has to match vertical of the DataStore linked to the engine.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.industry_vertical
	IndustryVertical *string `json:"industryVertical,omitempty"`

	// Common config spec that specifies the metadata of the engine.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.common_config
	CommonConfig *Engine_CommonConfig `json:"commonConfig,omitempty"`

	// Optional. Whether to disable analytics for searches performed on this
	//  engine.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.disable_analytics
	DisableAnalytics *bool `json:"disableAnalytics,omitempty"`
}

// DiscoveryEngineEngineStatus defines the config connector machine state of DiscoveryEngineEngine
type DiscoveryEngineEngineStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DiscoveryEngineEngine resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DiscoveryEngineEngineObservedState `json:"observedState,omitempty"`
}

// DiscoveryEngineEngineObservedState is the state of the DiscoveryEngineEngine resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.discoveryengine.v1.Engine
type DiscoveryEngineEngineObservedState struct {
	// Output only. Additional information of the Chat Engine. Only applicable
	//  if [solution_type][google.cloud.discoveryengine.v1.Engine.solution_type]
	//  is
	//  [SOLUTION_TYPE_CHAT][google.cloud.discoveryengine.v1.SolutionType.SOLUTION_TYPE_CHAT].
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.chat_engine_metadata
	ChatEngineMetadata *Engine_ChatEngineMetadata `json:"chatEngineMetadata,omitempty"`

	// Output only. Timestamp the Recommendation Engine was created at.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Timestamp the Recommendation Engine was last updated.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.Engine.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdiscoveryengineengine;gcpdiscoveryengineengines
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DiscoveryEngineEngine is the Schema for the DiscoveryEngineEngine API
// +k8s:openapi-gen=true
type DiscoveryEngineEngine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DiscoveryEngineEngineSpec   `json:"spec,omitempty"`
	Status DiscoveryEngineEngineStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DiscoveryEngineEngineList contains a list of DiscoveryEngineEngine
type DiscoveryEngineEngineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiscoveryEngineEngine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiscoveryEngineEngine{}, &DiscoveryEngineEngineList{})
}
