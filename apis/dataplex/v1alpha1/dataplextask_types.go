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

var DataplexTaskGVK = GroupVersion.WithKind("DataplexTask")

// DataplexTaskSpec defines the desired state of DataplexTask
// +kcc:spec:proto=google.cloud.dataplex.v1.Task
type DataplexTaskSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/lakes/{lake}/tasks/{task}
	Location *string `json:"location,omitempty"`

	// The Lake that this resource belongs to.
	// +kcc:guess=parent-ref target=LakeRef pattern=projects/{project}/locations/{location}/lakes/{lake}/tasks/{task}
	LakeRef *LakeRef `json:"lakeRef,omitempty"`

	// The DataplexTask name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Description of the task.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.description
	Description *string `json:"description,omitempty"`

	// Optional. User friendly display name.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. User-defined labels for the task.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. Spec related to how often and when a task should be triggered.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.trigger_spec
	// +required
	TriggerSpec *Task_TriggerSpec `json:"triggerSpec,omitempty"`

	// Required. Spec related to how a task is executed.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.execution_spec
	// +required
	ExecutionSpec *Task_ExecutionSpec `json:"executionSpec,omitempty"`

	// Config related to running custom Spark tasks.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.spark
	Spark *Task_SparkTaskConfig `json:"spark,omitempty"`

	// Config related to running scheduled Notebooks.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.notebook
	Notebook *Task_NotebookTaskConfig `json:"notebook,omitempty"`
}

// DataplexTaskStatus defines the config connector machine state of DataplexTask
type DataplexTaskStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DataplexTask resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DataplexTaskObservedState `json:"observedState,omitempty"`
}

// DataplexTaskObservedState is the state of the DataplexTask resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.dataplex.v1.Task
type DataplexTaskObservedState struct {
	// Output only. System generated globally unique ID for the task. This ID will
	//  be different if the task is deleted and re-created with the same name.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. The time when the task was created.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The time when the task was last updated.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Current state of the task.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.state
	State *string `json:"state,omitempty"`

	// Output only. Status of the latest task executions.
	// +kcc:proto:field=google.cloud.dataplex.v1.Task.execution_status
	ExecutionStatus *Task_ExecutionStatusObservedState `json:"executionStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdataplextask;gcpdataplextasks
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DataplexTask is the Schema for the DataplexTask API
// +k8s:openapi-gen=true
type DataplexTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DataplexTaskSpec   `json:"spec,omitempty"`
	Status DataplexTaskStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DataplexTaskList contains a list of DataplexTask
type DataplexTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataplexTask `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataplexTask{}, &DataplexTaskList{})
}
