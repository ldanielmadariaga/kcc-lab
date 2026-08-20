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

var DatastreamConnectionProfileGVK = GroupVersion.WithKind("DatastreamConnectionProfile")

// DatastreamConnectionProfileSpec defines the desired state of DatastreamConnectionProfile
// +kcc:spec:proto=google.cloud.datastream.v1.ConnectionProfile
type DatastreamConnectionProfileSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`


	// The DatastreamConnectionProfile name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Labels.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. Display name.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// Oracle ConnectionProfile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.oracle_profile
	OracleProfile *OracleProfile `json:"oracleProfile,omitempty"`

	// Cloud Storage ConnectionProfile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.gcs_profile
	GCSProfile *GCSProfile `json:"gcsProfile,omitempty"`

	// MySQL ConnectionProfile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.mysql_profile
	MysqlProfile *MysqlProfile `json:"mysqlProfile,omitempty"`

	// BigQuery Connection Profile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.bigquery_profile
	BigqueryProfile *BigQueryProfile `json:"bigqueryProfile,omitempty"`

	// PostgreSQL Connection Profile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.postgresql_profile
	PostgresqlProfile *PostgresqlProfile `json:"postgresqlProfile,omitempty"`

	// SQLServer Connection Profile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.sql_server_profile
	SQLServerProfile *SQLServerProfile `json:"sqlServerProfile,omitempty"`

	// Salesforce Connection Profile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.salesforce_profile
	SalesforceProfile *SalesforceProfile `json:"salesforceProfile,omitempty"`

	// MongoDB Connection Profile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.mongodb_profile
	MongodbProfile *MongodbProfile `json:"mongodbProfile,omitempty"`

	// Static Service IP connectivity.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.static_service_ip_connectivity
	StaticServiceIPConnectivity *StaticServiceIPConnectivity `json:"staticServiceIPConnectivity,omitempty"`

	// Forward SSH tunnel connectivity.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.forward_ssh_connectivity
	ForwardSSHConnectivity *ForwardSSHTunnelConnectivity `json:"forwardSSHConnectivity,omitempty"`

	// Private connectivity.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.private_connectivity
	PrivateConnectivity *PrivateConnectivity `json:"privateConnectivity,omitempty"`
}

// DatastreamConnectionProfileStatus defines the config connector machine state of DatastreamConnectionProfile
type DatastreamConnectionProfileStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DatastreamConnectionProfile resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DatastreamConnectionProfileObservedState `json:"observedState,omitempty"`
}

// DatastreamConnectionProfileObservedState is the state of the DatastreamConnectionProfile resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.datastream.v1.ConnectionProfile
type DatastreamConnectionProfileObservedState struct {
	// Output only. The create time of the resource.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The update time of the resource.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.satisfies_pzs
	SatisfiesPzs *bool `json:"satisfiesPzs,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.satisfies_pzi
	SatisfiesPzi *bool `json:"satisfiesPzi,omitempty"`

	// Oracle ConnectionProfile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.oracle_profile
	OracleProfile *OracleProfileObservedState `json:"oracleProfile,omitempty"`

	// MySQL ConnectionProfile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.mysql_profile
	MysqlProfile *MysqlProfileObservedState `json:"mysqlProfile,omitempty"`

	// MongoDB Connection Profile configuration.
	// +kcc:proto:field=google.cloud.datastream.v1.ConnectionProfile.mongodb_profile
	MongodbProfile *MongodbProfileObservedState `json:"mongodbProfile,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdatastreamconnectionprofile;gcpdatastreamconnectionprofiles
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DatastreamConnectionProfile is the Schema for the DatastreamConnectionProfile API
// +k8s:openapi-gen=true
type DatastreamConnectionProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DatastreamConnectionProfileSpec   `json:"spec,omitempty"`
	Status DatastreamConnectionProfileStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DatastreamConnectionProfileList contains a list of DatastreamConnectionProfile
type DatastreamConnectionProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatastreamConnectionProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatastreamConnectionProfile{}, &DatastreamConnectionProfileList{})
}
