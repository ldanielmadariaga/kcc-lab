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

var DiscoveryEngineDataStoreGVK = GroupVersion.WithKind("DiscoveryEngineDataStore")

// DiscoveryEngineDataStoreSpec defines the desired state of DiscoveryEngineDataStore
// +kcc:spec:proto=google.cloud.discoveryengine.v1.DataStore
type DiscoveryEngineDataStoreSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/dataStores/{data_store}
	Location *string `json:"location"`

	// The DiscoveryEngineDataStore name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Required. The data store display name.
	//
	//  This field must be a UTF-8 encoded string with a length limit of 128
	//  characters. Otherwise, an INVALID_ARGUMENT error is returned.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// Immutable. The industry vertical that the data store registers.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.industry_vertical
	IndustryVertical *string `json:"industryVertical,omitempty"`

	// The solutions that the data store enrolls. Available solutions for each
	//  [industry_vertical][google.cloud.discoveryengine.v1.DataStore.industry_vertical]:
	//
	//  * `MEDIA`: `SOLUTION_TYPE_RECOMMENDATION` and `SOLUTION_TYPE_SEARCH`.
	//  * `SITE_SEARCH`: `SOLUTION_TYPE_SEARCH` is automatically enrolled. Other
	//    solutions cannot be enrolled.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.solution_types
	SolutionTypes []string `json:"solutionTypes,omitempty"`

	// Immutable. The content config of the data store. If this field is unset,
	//  the server behavior defaults to
	//  [ContentConfig.NO_CONTENT][google.cloud.discoveryengine.v1.DataStore.ContentConfig.NO_CONTENT].
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.content_config
	ContentConfig *string `json:"contentConfig,omitempty"`

	// Optional. Configuration for advanced site search.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.advanced_site_search_config
	AdvancedSiteSearchConfig *AdvancedSiteSearchConfig `json:"advancedSiteSearchConfig,omitempty"`

	// Optional. Configuration for Natural Language Query Understanding.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.natural_language_query_understanding_config
	NaturalLanguageQueryUnderstandingConfig *NaturalLanguageQueryUnderstandingConfig `json:"naturalLanguageQueryUnderstandingConfig,omitempty"`

	// Input only. The KMS key to be used to protect this DataStore at creation
	//  time.
	//
	//  Must be set for requests that need to comply with CMEK Org Policy
	//  protections.
	//
	//  If this field is set and processed successfully, the DataStore will be
	//  protected by the KMS key, as indicated in the cmek_config field.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.kms_key_name
	KMSKeyName *string `json:"kmsKeyName,omitempty"`

	// Immutable. Whether data in the
	//  [DataStore][google.cloud.discoveryengine.v1.DataStore] has ACL information.
	//  If set to `true`, the source data must have ACL. ACL will be ingested when
	//  data is ingested by
	//  [DocumentService.ImportDocuments][google.cloud.discoveryengine.v1.DocumentService.ImportDocuments]
	//  methods.
	//
	//  When ACL is enabled for the
	//  [DataStore][google.cloud.discoveryengine.v1.DataStore],
	//  [Document][google.cloud.discoveryengine.v1.Document] can't be accessed by
	//  calling
	//  [DocumentService.GetDocument][google.cloud.discoveryengine.v1.DocumentService.GetDocument]
	//  or
	//  [DocumentService.ListDocuments][google.cloud.discoveryengine.v1.DocumentService.ListDocuments].
	//
	//  Currently ACL is only supported in `GENERIC` industry vertical with
	//  non-`PUBLIC_WEBSITE` content config.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.acl_enabled
	AclEnabled *bool `json:"aclEnabled,omitempty"`

	// Config to store data store type configuration for workspace data. This
	//  must be set when
	//  [DataStore.content_config][google.cloud.discoveryengine.v1.DataStore.content_config]
	//  is set as
	//  [DataStore.ContentConfig.GOOGLE_WORKSPACE][google.cloud.discoveryengine.v1.DataStore.ContentConfig.GOOGLE_WORKSPACE].
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.workspace_config
	WorkspaceConfig *WorkspaceConfig `json:"workspaceConfig,omitempty"`

	// Configuration for Document understanding and enrichment.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.document_processing_config
	DocumentProcessingConfig *DocumentProcessingConfig `json:"documentProcessingConfig,omitempty"`

	// The start schema to use for this
	//  [DataStore][google.cloud.discoveryengine.v1.DataStore] when provisioning
	//  it. If unset, a default vertical specialized schema will be used.
	//
	//  This field is only used by
	//  [CreateDataStore][google.cloud.discoveryengine.v1.DataStoreService.CreateDataStore]
	//  API, and will be ignored if used in other APIs. This field will be omitted
	//  from all API responses including
	//  [CreateDataStore][google.cloud.discoveryengine.v1.DataStoreService.CreateDataStore]
	//  API. To retrieve a schema of a
	//  [DataStore][google.cloud.discoveryengine.v1.DataStore], use
	//  [SchemaService.GetSchema][google.cloud.discoveryengine.v1.SchemaService.GetSchema]
	//  API instead.
	//
	//  The provided schema will be validated against certain rules on schema.
	//  Learn more from [this
	//  doc](https://cloud.google.com/generative-ai-app-builder/docs/provide-schema).
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.starting_schema
	StartingSchema *Schema `json:"startingSchema,omitempty"`

	// Optional. Configuration for `HEALTHCARE_FHIR` vertical.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.healthcare_fhir_config
	HealthcareFhirConfig *HealthcareFhirConfig `json:"healthcareFhirConfig,omitempty"`

	// Immutable. The fully qualified resource name of the associated
	//  [IdentityMappingStore][google.cloud.discoveryengine.v1.IdentityMappingStore].
	//  This field can only be set for acl_enabled DataStores with `THIRD_PARTY` or
	//  `GSUITE` IdP. Format:
	//  `projects/{project}/locations/{location}/identityMappingStores/{identity_mapping_store}`.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.identity_mapping_store
	IdentityMappingStore *string `json:"identityMappingStore,omitempty"`
}

// DiscoveryEngineDataStoreStatus defines the config connector machine state of DiscoveryEngineDataStore
type DiscoveryEngineDataStoreStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DiscoveryEngineDataStore resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DiscoveryEngineDataStoreObservedState `json:"observedState,omitempty"`
}

// DiscoveryEngineDataStoreObservedState is the state of the DiscoveryEngineDataStore resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.discoveryengine.v1.DataStore
type DiscoveryEngineDataStoreObservedState struct {
	// Output only. The id of the default
	//  [Schema][google.cloud.discoveryengine.v1.Schema] associated to this data
	//  store.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.default_schema_id
	DefaultSchemaID *string `json:"defaultSchemaID,omitempty"`

	// Output only. Timestamp the
	//  [DataStore][google.cloud.discoveryengine.v1.DataStore] was created at.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. CMEK-related information for the DataStore.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.cmek_config
	CmekConfig *CmekConfigObservedState `json:"cmekConfig,omitempty"`

	// Output only. Data size estimation for billing.
	// +kcc:proto:field=google.cloud.discoveryengine.v1.DataStore.billing_estimation
	BillingEstimation *DataStore_BillingEstimation `json:"billingEstimation,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdiscoveryenginedatastore;gcpdiscoveryenginedatastores
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DiscoveryEngineDataStore is the Schema for the DiscoveryEngineDataStore API
// +k8s:openapi-gen=true
type DiscoveryEngineDataStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DiscoveryEngineDataStoreSpec   `json:"spec,omitempty"`
	Status DiscoveryEngineDataStoreStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DiscoveryEngineDataStoreList contains a list of DiscoveryEngineDataStore
type DiscoveryEngineDataStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DiscoveryEngineDataStore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DiscoveryEngineDataStore{}, &DiscoveryEngineDataStoreList{})
}
