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

var VertexAIFeatureOnlineStoreGVK = GroupVersion.WithKind("VertexAIFeatureOnlineStore")

// VertexAIFeatureOnlineStoreSpec defines the desired state of VertexAIFeatureOnlineStore
// +kcc:spec:proto=google.cloud.aiplatform.v1.FeatureOnlineStore
type VertexAIFeatureOnlineStoreSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/featureOnlineStores/{feature_online_store}
	Location *string `json:"location"`

	// The VertexAIFeatureOnlineStore name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Contains settings for the Cloud Bigtable instance that will be created
	//  to serve featureValues for all FeatureViews under this
	//  FeatureOnlineStore.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.bigtable
	Bigtable *FeatureOnlineStore_Bigtable `json:"bigtable,omitempty"`

	// Contains settings for the Optimized store that will be created
	//  to serve featureValues for all FeatureViews under this
	//  FeatureOnlineStore. When choose Optimized storage type, need to set
	//  [PrivateServiceConnectConfig.enable_private_service_connect][google.cloud.aiplatform.v1.PrivateServiceConnectConfig.enable_private_service_connect]
	//  to use private endpoint. Otherwise will use public endpoint by default.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.optimized
	Optimized *FeatureOnlineStore_Optimized `json:"optimized,omitempty"`

	// Optional. Used to perform consistent read-modify-write updates. If not set,
	//  a blind "overwrite" update happens.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.etag
	Etag *string `json:"etag,omitempty"`

	// Optional. The labels with user-defined metadata to organize your
	//  FeatureOnlineStore.
	//
	//  Label keys and values can be no longer than 64 characters
	//  (Unicode codepoints), can only contain lowercase letters, numeric
	//  characters, underscores and dashes. International characters are allowed.
	//
	//  See https://goo.gl/xmQnxf for more information on and examples of labels.
	//  No more than 64 user labels can be associated with one
	//  FeatureOnlineStore(System labels are excluded)." System reserved label keys
	//  are prefixed with "aiplatform.googleapis.com/" and are immutable.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. The dedicated serving endpoint for this FeatureOnlineStore, which
	//  is different from common Vertex service endpoint.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.dedicated_serving_endpoint
	DedicatedServingEndpoint *FeatureOnlineStore_DedicatedServingEndpoint `json:"dedicatedServingEndpoint,omitempty"`

	// Optional. Customer-managed encryption key spec for data storage. If set,
	//  online store will be secured by this key.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.encryption_spec
	EncryptionSpec *EncryptionSpec `json:"encryptionSpec,omitempty"`
}

// VertexAIFeatureOnlineStoreStatus defines the config connector machine state of VertexAIFeatureOnlineStore
type VertexAIFeatureOnlineStoreStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the VertexAIFeatureOnlineStore resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *VertexAIFeatureOnlineStoreObservedState `json:"observedState,omitempty"`
}

// VertexAIFeatureOnlineStoreObservedState is the state of the VertexAIFeatureOnlineStore resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.aiplatform.v1.FeatureOnlineStore
type VertexAIFeatureOnlineStoreObservedState struct {
	// Output only. Timestamp when this FeatureOnlineStore was created.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Timestamp when this FeatureOnlineStore was last updated.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. State of the featureOnlineStore.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.state
	State *string `json:"state,omitempty"`

	// Optional. The dedicated serving endpoint for this FeatureOnlineStore, which
	//  is different from common Vertex service endpoint.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.dedicated_serving_endpoint
	DedicatedServingEndpoint *FeatureOnlineStore_DedicatedServingEndpointObservedState `json:"dedicatedServingEndpoint,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.satisfies_pzs
	SatisfiesPzs *bool `json:"satisfiesPzs,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.aiplatform.v1.FeatureOnlineStore.satisfies_pzi
	SatisfiesPzi *bool `json:"satisfiesPzi,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpvertexaifeatureonlinestore;gcpvertexaifeatureonlinestores
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// VertexAIFeatureOnlineStore is the Schema for the VertexAIFeatureOnlineStore API
// +k8s:openapi-gen=true
type VertexAIFeatureOnlineStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   VertexAIFeatureOnlineStoreSpec   `json:"spec,omitempty"`
	Status VertexAIFeatureOnlineStoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VertexAIFeatureOnlineStoreList contains a list of VertexAIFeatureOnlineStore
type VertexAIFeatureOnlineStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VertexAIFeatureOnlineStore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VertexAIFeatureOnlineStore{}, &VertexAIFeatureOnlineStoreList{})
}
