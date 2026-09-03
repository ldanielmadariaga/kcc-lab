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

var BackupDRBackupPlanGVK = GroupVersion.WithKind("BackupDRBackupPlan")

// BackupDRBackupPlanSpec defines the desired state of BackupDRBackupPlan
// +kcc:spec:proto=google.cloud.backupdr.v1.BackupPlan
type BackupDRBackupPlanSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/backupPlans/{backup_plan}
	Location *string `json:"location"`

	// The BackupDRBackupPlan name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. The description of the `BackupPlan` resource.
	//
	//  The description allows for additional details about `BackupPlan` and its
	//  use cases to be provided. An example description is the following:  "This
	//  is a backup plan that performs a daily backup at 6pm and retains data for 3
	//  months". The description must be at most 2048 characters.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.description
	Description *string `json:"description,omitempty"`

	// Optional. This collection of key/value pairs allows for custom labels to be
	//  supplied by the user.  Example, {"tag": "Weekly"}.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. The backup rules for this `BackupPlan`. There must be at least
	//  one `BackupRule` message.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.backup_rules
	// +required
	BackupRules []BackupRule `json:"backupRules,omitempty"`

	// Required. The resource type to which the `BackupPlan` will be applied.
	//  Examples include, "compute.googleapis.com/Instance",
	//  "sqladmin.googleapis.com/Instance", "alloydb.googleapis.com/Cluster",
	//  "compute.googleapis.com/Disk".
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.resource_type
	// +required
	ResourceType *string `json:"resourceType,omitempty"`

	// Optional. `etag` is returned from the service in the response. As a user of
	//  the service, you may provide an etag value in this field to prevent stale
	//  resources.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.etag
	Etag *string `json:"etag,omitempty"`

	// Required. Resource name of backup vault which will be used as storage
	//  location for backups. Format:
	//  projects/{project}/locations/{location}/backupVaults/{backupvault}
	// +kcc:guess=possible-reference target=BackupDRBackupVault
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.backup_vault
	// +required
	BackupVault *string `json:"backupVault,omitempty"`

	// Optional. Applicable only for CloudSQL resource_type.
	//
	//  Configures how long logs will be stored. It is defined in “days”. This
	//  value should be greater than or equal to minimum enforced log retention
	//  duration of the backup vault.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.log_retention_days
	LogRetentionDays *int64 `json:"logRetentionDays,omitempty"`
}

// BackupDRBackupPlanStatus defines the config connector machine state of BackupDRBackupPlan
type BackupDRBackupPlanStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the BackupDRBackupPlan resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *BackupDRBackupPlanObservedState `json:"observedState,omitempty"`
}

// BackupDRBackupPlanObservedState is the state of the BackupDRBackupPlan resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.backupdr.v1.BackupPlan
type BackupDRBackupPlanObservedState struct {
	// Output only. When the `BackupPlan` was created.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. When the `BackupPlan` was last updated.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The `State` for the `BackupPlan`.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.state
	State *string `json:"state,omitempty"`

	// Output only. The Google Cloud Platform Service Account to be used by the
	//  BackupVault for taking backups. Specify the email address of the Backup
	//  Vault Service Account.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.backup_vault_service_account
	BackupVaultServiceAccount *string `json:"backupVaultServiceAccount,omitempty"`

	// Output only. All resource types to which backupPlan can be applied.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.supported_resource_types
	SupportedResourceTypes []string `json:"supportedResourceTypes,omitempty"`

	// Output only. The user friendly revision ID of the `BackupPlanRevision`.
	//
	//  Example: v0, v1, v2, etc.
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.revision_id
	RevisionID *string `json:"revisionID,omitempty"`

	// Output only. The resource id of the `BackupPlanRevision`.
	//
	//  Format:
	//  `projects/{project}/locations/{location}/backupPlans/{backup_plan}/revisions/{revision_id}`
	// +kcc:proto:field=google.cloud.backupdr.v1.BackupPlan.revision_name
	RevisionName *string `json:"revisionName,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpbackupdrbackupplan;gcpbackupdrbackupplans
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// BackupDRBackupPlan is the Schema for the BackupDRBackupPlan API
// +k8s:openapi-gen=true
type BackupDRBackupPlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   BackupDRBackupPlanSpec   `json:"spec,omitempty"`
	Status BackupDRBackupPlanStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// BackupDRBackupPlanList contains a list of BackupDRBackupPlan
type BackupDRBackupPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupDRBackupPlan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupDRBackupPlan{}, &BackupDRBackupPlanList{})
}
