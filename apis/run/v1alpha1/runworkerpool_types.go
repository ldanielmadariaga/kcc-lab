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

var RunWorkerPoolGVK = GroupVersion.WithKind("RunWorkerPool")

// RunWorkerPoolSpec defines the desired state of RunWorkerPool
// +kcc:spec:proto=google.cloud.run.v2.WorkerPool
type RunWorkerPoolSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The RunWorkerPool name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// User-provided description of the WorkerPool. This field currently has a
	//  512-character limit.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.description
	Description *string `json:"description,omitempty"`

	// Optional. Unstructured key value map that can be used to organize and
	//  categorize objects. User-provided labels are shared with Google's billing
	//  system, so they can be used to filter, or break down billing charges by
	//  team, component, environment, state, etc. For more information, visit
	//  https://cloud.google.com/resource-manager/docs/creating-managing-labels or
	//  https://cloud.google.com/run/docs/configuring/labels.
	//
	//  Cloud Run API v2 does not support labels with  `run.googleapis.com`,
	//  `cloud.googleapis.com`, `serving.knative.dev`, or `autoscaling.knative.dev`
	//  namespaces, and they will be rejected. All system labels in v1 now have a
	//  corresponding field in v2 WorkerPool.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. Unstructured key value map that may be set by external tools to
	//  store and arbitrary metadata. They are not queryable and should be
	//  preserved when modifying objects.
	//
	//  Cloud Run API v2 does not support annotations with `run.googleapis.com`,
	//  `cloud.googleapis.com`, `serving.knative.dev`, or `autoscaling.knative.dev`
	//  namespaces, and they will be rejected in new resources. All system
	//  annotations in v1 now have a corresponding field in v2 WorkerPool.
	//
	//  <p>This field follows Kubernetes
	//  annotations' namespacing, limits, and rules.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Arbitrary identifier for the API client.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.client
	Client *string `json:"client,omitempty"`

	// Arbitrary version identifier for the API client.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.client_version
	ClientVersion *string `json:"clientVersion,omitempty"`

	// Optional. The launch stage as defined by [Google Cloud Platform
	//   Launch Stages](https://cloud.google.com/terms/launch-stages).
	//   Cloud Run supports `ALPHA`, `BETA`, and `GA`. If no value is specified, GA
	//   is assumed.
	//   Set the launch stage to a preview stage on input to allow use of preview
	//   features in that stage. On read (or output), describes whether the
	//   resource uses preview features.
	//
	//   For example, if ALPHA is provided as input, but only BETA and GA-level
	//   features are used, this field will be BETA on output.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.launch_stage
	LaunchStage *string `json:"launchStage,omitempty"`

	// Optional. Settings for the Binary Authorization feature.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.binary_authorization
	BinaryAuthorization *BinaryAuthorization `json:"binaryAuthorization,omitempty"`

	// Required. The template used to create revisions for this WorkerPool.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.template
	// +required
	Template *WorkerPoolRevisionTemplate `json:"template,omitempty"`

	// Optional. Specifies how to distribute instances over a collection of
	//  Revisions belonging to the WorkerPool. If instance split is empty or not
	//  provided, defaults to 100% instances assigned to the latest `Ready`
	//  Revision.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.instance_splits
	InstanceSplits []InstanceSplit `json:"instanceSplits,omitempty"`

	// Optional. Specifies worker-pool-level scaling settings
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.scaling
	Scaling *WorkerPoolScaling `json:"scaling,omitempty"`

	// One or more custom audiences that you want this worker pool to support.
	//  Specify each custom audience as the full URL in a string. The custom
	//  audiences are encoded in the token and used to authenticate requests. For
	//  more information, see
	//  https://cloud.google.com/run/docs/configuring/custom-audiences.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.custom_audiences
	CustomAudiences []string `json:"customAudiences,omitempty"`
}

// RunWorkerPoolStatus defines the config connector machine state of RunWorkerPool
type RunWorkerPoolStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the RunWorkerPool resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *RunWorkerPoolObservedState `json:"observedState,omitempty"`
}

// RunWorkerPoolObservedState is the state of the RunWorkerPool resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.run.v2.WorkerPool
type RunWorkerPoolObservedState struct {
	// Output only. Server assigned unique identifier for the trigger. The value
	//  is a UUID4 string and guaranteed to remain unchanged until the resource is
	//  deleted.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. A number that monotonically increases every time the user
	//  modifies the desired state.
	//  Please note that unlike v1, this is an int64 value. As with most Google
	//  APIs, its JSON representation will be a `string` instead of an `integer`.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.generation
	Generation *int64 `json:"generation,omitempty"`

	// Output only. The creation time.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The last-modified time.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The deletion time. It is only populated as a response to a
	//  Delete request.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.delete_time
	DeleteTime *string `json:"deleteTime,omitempty"`

	// Output only. For a deleted resource, the time after which it will be
	//  permamently deleted.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.expire_time
	ExpireTime *string `json:"expireTime,omitempty"`

	// Output only. Email address of the authenticated creator.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.creator
	Creator *string `json:"creator,omitempty"`

	// Output only. Email address of the last authenticated modifier.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.last_modifier
	LastModifier *string `json:"lastModifier,omitempty"`

	// Required. The template used to create revisions for this WorkerPool.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.template
	Template *WorkerPoolRevisionTemplateObservedState `json:"template,omitempty"`

	// Output only. The generation of this WorkerPool currently serving traffic.
	//  See comments in `reconciling` for additional information on reconciliation
	//  process in Cloud Run. Please note that unlike v1, this is an int64 value.
	//  As with most Google APIs, its JSON representation will be a `string`
	//  instead of an `integer`.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.observed_generation
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// Output only. The Condition of this WorkerPool, containing its readiness
	//  status, and detailed error information in case it did not reach a serving
	//  state. See comments in `reconciling` for additional information on
	//  reconciliation process in Cloud Run.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.terminal_condition
	TerminalCondition *ConditionObservedState `json:"terminalCondition,omitempty"`

	// Output only. The Conditions of all other associated sub-resources. They
	//  contain additional diagnostics information in case the WorkerPool does not
	//  reach its Serving state. See comments in `reconciling` for additional
	//  information on reconciliation process in Cloud Run.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.conditions
	Conditions []ConditionObservedState `json:"conditions,omitempty"`

	// Output only. Name of the latest revision that is serving traffic. See
	//  comments in `reconciling` for additional information on reconciliation
	//  process in Cloud Run.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.latest_ready_revision
	LatestReadyRevision *string `json:"latestReadyRevision,omitempty"`

	// Output only. Name of the last created revision. See comments in
	//  `reconciling` for additional information on reconciliation process in Cloud
	//  Run.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.latest_created_revision
	LatestCreatedRevision *string `json:"latestCreatedRevision,omitempty"`

	// Output only. Detailed status information for corresponding instance splits.
	//  See comments in `reconciling` for additional information on reconciliation
	//  process in Cloud Run.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.instance_split_statuses
	InstanceSplitStatuses []InstanceSplitStatus `json:"instanceSplitStatuses,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.satisfies_pzs
	SatisfiesPzs *bool `json:"satisfiesPzs,omitempty"`

	// Output only. Returns true if the WorkerPool is currently being acted upon
	//  by the system to bring it into the desired state.
	//
	//  When a new WorkerPool is created, or an existing one is updated, Cloud Run
	//  will asynchronously perform all necessary steps to bring the WorkerPool to
	//  the desired serving state. This process is called reconciliation. While
	//  reconciliation is in process, `observed_generation`,
	//  `latest_ready_revison`, `traffic_statuses`, and `uri` will have transient
	//  values that might mismatch the intended state: Once reconciliation is over
	//  (and this field is false), there are two possible outcomes: reconciliation
	//  succeeded and the serving state matches the WorkerPool, or there was an
	//  error, and reconciliation failed. This state can be found in
	//  `terminal_condition.state`.
	//
	//  If reconciliation succeeded, the following fields will match: `traffic` and
	//  `traffic_statuses`, `observed_generation` and `generation`,
	//  `latest_ready_revision` and `latest_created_revision`.
	//
	//  If reconciliation failed, `traffic_statuses`, `observed_generation`, and
	//  `latest_ready_revision` will have the state of the last serving revision,
	//  or empty for newly created WorkerPools. Additional information on the
	//  failure can be found in `terminal_condition` and `conditions`.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.reconciling
	Reconciling *bool `json:"reconciling,omitempty"`

	// Output only. A system-generated fingerprint for this version of the
	//  resource. May be used to detect modification conflict during updates.
	// +kcc:proto:field=google.cloud.run.v2.WorkerPool.etag
	Etag *string `json:"etag,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcprunworkerpool;gcprunworkerpools
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// RunWorkerPool is the Schema for the RunWorkerPool API
// +k8s:openapi-gen=true
type RunWorkerPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   RunWorkerPoolSpec   `json:"spec,omitempty"`
	Status RunWorkerPoolStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// RunWorkerPoolList contains a list of RunWorkerPool
type RunWorkerPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RunWorkerPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RunWorkerPool{}, &RunWorkerPoolList{})
}
