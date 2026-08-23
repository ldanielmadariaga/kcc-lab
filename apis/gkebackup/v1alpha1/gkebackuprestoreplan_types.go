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

var GKEBackupRestorePlanGVK = GroupVersion.WithKind("GKEBackupRestorePlan")

// GKEBackupRestorePlanSpec defines the desired state of GKEBackupRestorePlan
// +kcc:spec:proto=google.cloud.gkebackup.v1.RestorePlan
type GKEBackupRestorePlanSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/restorePlans/{restore_plan}
	Location *string `json:"location"`

	// The GKEBackupRestorePlan name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. User specified descriptive string for this RestorePlan.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.description
	Description *string `json:"description,omitempty"`

	// Required. Immutable. A reference to the
	//  [BackupPlan][google.cloud.gkebackup.v1.BackupPlan] from which Backups may
	//  be used as the source for Restores created via this RestorePlan. Format:
	//  `projects/*/locations/*/backupPlans/*`.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.backup_plan
	// +required
	BackupPlan *string `json:"backupPlan,omitempty"`

	// Required. Immutable. The target cluster into which Restores created via
	//  this RestorePlan will restore data. NOTE: the cluster's region must be the
	//  same as the RestorePlan. Valid formats:
	//
	//    - `projects/*/locations/*/clusters/*`
	//    - `projects/*/zones/*/clusters/*`
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.cluster
	// +required
	Cluster *string `json:"cluster,omitempty"`

	// Required. Configuration of Restores created via this RestorePlan.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.restore_config
	// +required
	RestoreConfig *RestoreConfig `json:"restoreConfig,omitempty"`

	// Optional. A set of custom labels supplied by user.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.labels
	Labels map[string]string `json:"labels,omitempty"`
}

// GKEBackupRestorePlanStatus defines the config connector machine state of GKEBackupRestorePlan
type GKEBackupRestorePlanStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the GKEBackupRestorePlan resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *GKEBackupRestorePlanObservedState `json:"observedState,omitempty"`
}

// GKEBackupRestorePlanObservedState is the state of the GKEBackupRestorePlan resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.gkebackup.v1.RestorePlan
type GKEBackupRestorePlanObservedState struct {
	// Output only. Server generated global unique identifier of
	//  [UUID](https://en.wikipedia.org/wiki/Universally_unique_identifier) format.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. The timestamp when this RestorePlan resource was
	//  created.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The timestamp when this RestorePlan resource was last
	//  updated.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. `etag` is used for optimistic concurrency control as a way to
	//  help prevent simultaneous updates of a restore from overwriting each other.
	//  It is strongly suggested that systems make use of the `etag` in the
	//  read-modify-write cycle to perform restore updates in order to avoid
	//  race conditions: An `etag` is returned in the response to `GetRestorePlan`,
	//  and systems are expected to put that etag in the request to
	//  `UpdateRestorePlan` or `DeleteRestorePlan` to ensure that their change
	//  will be applied to the same version of the resource.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.etag
	Etag *string `json:"etag,omitempty"`

	// Output only. State of the RestorePlan. This State field reflects the
	//  various stages a RestorePlan can be in
	//  during the Create operation.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.state
	State *string `json:"state,omitempty"`

	// Output only. Human-readable description of why RestorePlan is in the
	//  current `state`. This field is only meant for human readability and should
	//  not be used programmatically as this field is not guaranteed to be
	//  consistent.
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.state_reason
	StateReason *string `json:"stateReason,omitempty"`

	// Output only. The fully qualified name of the RestoreChannel to be used to
	//  create a RestorePlan. This field is set only if the `backup_plan` is in a
	//  different project than the RestorePlan. Format:
	//  `projects/*/locations/*/restoreChannels/*`
	// +kcc:proto:field=google.cloud.gkebackup.v1.RestorePlan.restore_channel
	RestoreChannel *string `json:"restoreChannel,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpgkebackuprestoreplan;gcpgkebackuprestoreplans
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// GKEBackupRestorePlan is the Schema for the GKEBackupRestorePlan API
// +k8s:openapi-gen=true
type GKEBackupRestorePlan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   GKEBackupRestorePlanSpec   `json:"spec,omitempty"`
	Status GKEBackupRestorePlanStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// GKEBackupRestorePlanList contains a list of GKEBackupRestorePlan
type GKEBackupRestorePlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GKEBackupRestorePlan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GKEBackupRestorePlan{}, &GKEBackupRestorePlanList{})
}
