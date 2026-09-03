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

var BigQueryConnectionConnectionGVK = GroupVersion.WithKind("BigQueryConnectionConnection")

// BigQueryConnectionConnectionSpec defines the desired state of BigQueryConnectionConnection
// +kcc:spec:proto=google.cloud.bigquery.connection.v1.Connection
type BigQueryConnectionConnectionSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/connections/{connection}
	Location *string `json:"location"`

	// The BigQueryConnectionConnection name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// User provided display name for the connection.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.friendly_name
	FriendlyName *string `json:"friendlyName,omitempty"`

	// User provided description.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.description
	Description *string `json:"description,omitempty"`

	// Cloud SQL properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.cloud_sql
	CloudSQL *CloudSQLProperties `json:"cloudSQL,omitempty"`

	// Amazon Web Services (AWS) properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.aws
	Aws *AwsProperties `json:"aws,omitempty"`

	// Azure properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.azure
	Azure *AzureProperties `json:"azure,omitempty"`

	// Cloud Spanner properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.cloud_spanner
	CloudSpanner *CloudSpannerProperties `json:"cloudSpanner,omitempty"`

	// Cloud Resource properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.cloud_resource
	CloudResource *CloudResourceProperties `json:"cloudResource,omitempty"`

	// Spark properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.spark
	Spark *SparkProperties `json:"spark,omitempty"`

	// Optional. Salesforce DataCloud properties. This field is intended for
	//  use only by Salesforce partner projects. This field contains properties
	//  for your Salesforce DataCloud connection.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.salesforce_data_cloud
	SalesforceDataCloud *SalesforceDataCloudProperties `json:"salesforceDataCloud,omitempty"`
}

// BigQueryConnectionConnectionStatus defines the config connector machine state of BigQueryConnectionConnection
type BigQueryConnectionConnectionStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the BigQueryConnectionConnection resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *BigQueryConnectionConnectionObservedState `json:"observedState,omitempty"`
}

// BigQueryConnectionConnectionObservedState is the state of the BigQueryConnectionConnection resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.bigquery.connection.v1.Connection
type BigQueryConnectionConnectionObservedState struct {
	// Cloud SQL properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.cloud_sql
	CloudSQL *CloudSQLPropertiesObservedState `json:"cloudSQL,omitempty"`

	// Amazon Web Services (AWS) properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.aws
	Aws *AwsPropertiesObservedState `json:"aws,omitempty"`

	// Azure properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.azure
	Azure *AzurePropertiesObservedState `json:"azure,omitempty"`

	// Cloud Resource properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.cloud_resource
	CloudResource *CloudResourcePropertiesObservedState `json:"cloudResource,omitempty"`

	// Spark properties.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.spark
	Spark *SparkPropertiesObservedState `json:"spark,omitempty"`

	// Optional. Salesforce DataCloud properties. This field is intended for
	//  use only by Salesforce partner projects. This field contains properties
	//  for your Salesforce DataCloud connection.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.salesforce_data_cloud
	SalesforceDataCloud *SalesforceDataCloudPropertiesObservedState `json:"salesforceDataCloud,omitempty"`

	// Output only. The creation timestamp of the connection.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.creation_time
	CreationTime *int64 `json:"creationTime,omitempty"`

	// Output only. The last update timestamp of the connection.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.last_modified_time
	LastModifiedTime *int64 `json:"lastModifiedTime,omitempty"`

	// Output only. True, if credential is configured for this connection.
	// +kcc:proto:field=google.cloud.bigquery.connection.v1.Connection.has_credential
	HasCredential *bool `json:"hasCredential,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpbigqueryconnectionconnection;gcpbigqueryconnectionconnections
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// BigQueryConnectionConnection is the Schema for the BigQueryConnectionConnection API
// +k8s:openapi-gen=true
type BigQueryConnectionConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   BigQueryConnectionConnectionSpec   `json:"spec,omitempty"`
	Status BigQueryConnectionConnectionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// BigQueryConnectionConnectionList contains a list of BigQueryConnectionConnection
type BigQueryConnectionConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BigQueryConnectionConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BigQueryConnectionConnection{}, &BigQueryConnectionConnectionList{})
}
