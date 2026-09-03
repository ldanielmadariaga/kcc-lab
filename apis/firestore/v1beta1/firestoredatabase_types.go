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

var FirestoreDatabaseGVK = GroupVersion.WithKind("FirestoreDatabase")

// FirestoreDatabaseSpec defines the desired state of FirestoreDatabase
// +kcc:spec:proto=google.firestore.admin.v1.Database
type FirestoreDatabaseSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The FirestoreDatabase name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The location of the database. Available locations are listed at
	//  https://cloud.google.com/firestore/docs/locations.
	// +kcc:proto:field=google.firestore.admin.v1.Database.location_id
	LocationID *string `json:"locationID,omitempty"`

	// The type of the database.
	//  See https://cloud.google.com/datastore/docs/firestore-or-datastore for
	//  information about how to choose.
	// +kcc:proto:field=google.firestore.admin.v1.Database.type
	Type *string `json:"type,omitempty"`

	// The concurrency control mode to use for this database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.concurrency_mode
	ConcurrencyMode *string `json:"concurrencyMode,omitempty"`

	// Whether to enable the PITR feature on this database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.point_in_time_recovery_enablement
	PointInTimeRecoveryEnablement *string `json:"pointInTimeRecoveryEnablement,omitempty"`

	// The App Engine integration mode to use for this database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.app_engine_integration_mode
	AppEngineIntegrationMode *string `json:"appEngineIntegrationMode,omitempty"`

	// State of delete protection for the database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.delete_protection_state
	DeleteProtectionState *string `json:"deleteProtectionState,omitempty"`

	// Optional. Presence indicates CMEK is enabled for this database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.cmek_config
	CmekConfig *Database_CmekConfig `json:"cmekConfig,omitempty"`

	// Optional. Input only. Immutable. Tag keys/values directly bound to this
	//  resource. For example:
	//    "123/environment": "production",
	//    "123/costCenter": "marketing"
	// +kcc:proto:field=google.firestore.admin.v1.Database.tags
	Tags map[string]string `json:"tags,omitempty"`

	// This checksum is computed by the server based on the value of other
	//  fields, and may be sent on update and delete requests to ensure the
	//  client has an up-to-date value before proceeding.
	// +kcc:proto:field=google.firestore.admin.v1.Database.etag
	Etag *string `json:"etag,omitempty"`

	// Immutable. The edition of the database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.database_edition
	DatabaseEdition *string `json:"databaseEdition,omitempty"`
}

// FirestoreDatabaseStatus defines the config connector machine state of FirestoreDatabase
type FirestoreDatabaseStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the FirestoreDatabase resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *FirestoreDatabaseObservedState `json:"observedState,omitempty"`
}

// FirestoreDatabaseObservedState is the state of the FirestoreDatabase resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.firestore.admin.v1.Database
type FirestoreDatabaseObservedState struct {
	// Output only. The system-generated UUID4 for this Database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. The timestamp at which this database was created. Databases
	//  created before 2016 do not populate create_time.
	// +kcc:proto:field=google.firestore.admin.v1.Database.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The timestamp at which this database was most recently
	//  updated. Note this only includes updates to the database resource and not
	//  data contained by the database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The timestamp at which this database was deleted. Only set if
	//  the database has been deleted.
	// +kcc:proto:field=google.firestore.admin.v1.Database.delete_time
	DeleteTime *string `json:"deleteTime,omitempty"`

	// Output only. The period during which past versions of data are retained in
	//  the database.
	//
	//  Any [read][google.firestore.v1.GetDocumentRequest.read_time]
	//  or [query][google.firestore.v1.ListDocumentsRequest.read_time] can specify
	//  a `read_time` within this window, and will read the state of the database
	//  at that time.
	//
	//  If the PITR feature is enabled, the retention period is 7 days. Otherwise,
	//  the retention period is 1 hour.
	// +kcc:proto:field=google.firestore.admin.v1.Database.version_retention_period
	VersionRetentionPeriod *string `json:"versionRetentionPeriod,omitempty"`

	// Output only. The earliest timestamp at which older versions of the data can
	//  be read from the database. See [version_retention_period] above; this field
	//  is populated with `now - version_retention_period`.
	//
	//  This value is continuously updated, and becomes stale the moment it is
	//  queried. If you are using this value to recover data, make sure to account
	//  for the time from the moment when the value is queried to the moment when
	//  you initiate the recovery.
	// +kcc:proto:field=google.firestore.admin.v1.Database.earliest_version_time
	EarliestVersionTime *string `json:"earliestVersionTime,omitempty"`

	// Output only. The key_prefix for this database. This key_prefix is used, in
	//  combination with the project ID ("<key prefix>~<project id>") to construct
	//  the application ID that is returned from the Cloud Datastore APIs in Google
	//  App Engine first generation runtimes.
	//
	//  This value may be empty in which case the appid to use for URL-encoded keys
	//  is the project_id (eg: foo instead of v~foo).
	// +kcc:proto:field=google.firestore.admin.v1.Database.key_prefix
	KeyPrefix *string `json:"keyPrefix,omitempty"`

	// Optional. Presence indicates CMEK is enabled for this database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.cmek_config
	CmekConfig *Database_CmekConfigObservedState `json:"cmekConfig,omitempty"`

	// Output only. The database resource's prior database ID. This field is only
	//  populated for deleted databases.
	// +kcc:proto:field=google.firestore.admin.v1.Database.previous_id
	PreviousID *string `json:"previousID,omitempty"`

	// Output only. Information about the provenance of this database.
	// +kcc:proto:field=google.firestore.admin.v1.Database.source_info
	SourceInfo *Database_SourceInfo `json:"sourceInfo,omitempty"`

	// Output only. Background: Free tier is the ability of a Firestore database
	//  to use a small amount of resources every day without being charged. Once
	//  usage exceeds the free tier limit further usage is charged.
	//
	//  Whether this database can make use of the free tier. Only one database
	//  per project can be eligible for the free tier.
	//
	//  The first (or next) database that is created in a project without a free
	//  tier database will be marked as eligible for the free tier. Databases that
	//  are created while there is a free tier database will not be eligible for
	//  the free tier.
	// +kcc:proto:field=google.firestore.admin.v1.Database.free_tier
	FreeTier *bool `json:"freeTier,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpfirestoredatabase;gcpfirestoredatabases
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// FirestoreDatabase is the Schema for the FirestoreDatabase API
// +k8s:openapi-gen=true
type FirestoreDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   FirestoreDatabaseSpec   `json:"spec,omitempty"`
	Status FirestoreDatabaseStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// FirestoreDatabaseList contains a list of FirestoreDatabase
type FirestoreDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirestoreDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirestoreDatabase{}, &FirestoreDatabaseList{})
}
