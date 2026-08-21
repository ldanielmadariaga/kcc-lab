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

var SecurityCenterBigQueryExportGVK = GroupVersion.WithKind("SecurityCenterBigQueryExport")

// SecurityCenterBigQueryExportSpec defines the desired state of SecurityCenterBigQueryExport
// +kcc:spec:proto=google.cloud.securitycenter.v1.BigQueryExport
type SecurityCenterBigQueryExportSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The SecurityCenterBigQueryExport name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The description of the export (max of 1024 characters).
	// +kcc:proto:field=google.cloud.securitycenter.v1.BigQueryExport.description
	Description *string `json:"description,omitempty"`

	// Expression that defines the filter to apply across create/update events
	//  of findings. The expression is a list of zero or more restrictions combined
	//  via logical operators `AND` and `OR`. Parentheses are supported, and `OR`
	//  has higher precedence than `AND`.
	//
	//  Restrictions have the form `<field> <operator> <value>` and may have a
	//  `-` character in front of them to indicate negation. The fields map to
	//  those defined in the corresponding resource.
	//
	//  The supported operators are:
	//
	//  * `=` for all value types.
	//  * `>`, `<`, `>=`, `<=` for integer values.
	//  * `:`, meaning substring matching, for strings.
	//
	//  The supported value types are:
	//
	//  * string literals in quotes.
	//  * integer literals without quotes.
	//  * boolean literals `true` and `false` without quotes.
	// +kcc:proto:field=google.cloud.securitycenter.v1.BigQueryExport.filter
	Filter *string `json:"filter,omitempty"`

	// The dataset to write findings' updates to. Its format is
	//  "projects/[project_id]/datasets/[bigquery_dataset_id]".
	//  BigQuery Dataset unique ID  must contain only letters (a-z, A-Z), numbers
	//  (0-9), or underscores (_).
	// +kcc:proto:field=google.cloud.securitycenter.v1.BigQueryExport.dataset
	Dataset *string `json:"dataset,omitempty"`
}

// SecurityCenterBigQueryExportStatus defines the config connector machine state of SecurityCenterBigQueryExport
type SecurityCenterBigQueryExportStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the SecurityCenterBigQueryExport resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *SecurityCenterBigQueryExportObservedState `json:"observedState,omitempty"`
}

// SecurityCenterBigQueryExportObservedState is the state of the SecurityCenterBigQueryExport resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.securitycenter.v1.BigQueryExport
type SecurityCenterBigQueryExportObservedState struct {
	// Output only. The time at which the BigQuery export was created.
	//  This field is set by the server and will be ignored if provided on export
	//  on creation.
	// +kcc:proto:field=google.cloud.securitycenter.v1.BigQueryExport.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The most recent time at which the BigQuery export was updated.
	//  This field is set by the server and will be ignored if provided on export
	//  creation or update.
	// +kcc:proto:field=google.cloud.securitycenter.v1.BigQueryExport.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Email address of the user who last edited the BigQuery export.
	//  This field is set by the server and will be ignored if provided on export
	//  creation or update.
	// +kcc:proto:field=google.cloud.securitycenter.v1.BigQueryExport.most_recent_editor
	MostRecentEditor *string `json:"mostRecentEditor,omitempty"`

	// Output only. The service account that needs permission to create table and
	//  upload data to the BigQuery dataset.
	// +kcc:proto:field=google.cloud.securitycenter.v1.BigQueryExport.principal
	Principal *string `json:"principal,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpsecuritycenterbigqueryexport;gcpsecuritycenterbigqueryexports
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// SecurityCenterBigQueryExport is the Schema for the SecurityCenterBigQueryExport API
// +k8s:openapi-gen=true
type SecurityCenterBigQueryExport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   SecurityCenterBigQueryExportSpec   `json:"spec,omitempty"`
	Status SecurityCenterBigQueryExportStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// SecurityCenterBigQueryExportList contains a list of SecurityCenterBigQueryExport
type SecurityCenterBigQueryExportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityCenterBigQueryExport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecurityCenterBigQueryExport{}, &SecurityCenterBigQueryExportList{})
}
