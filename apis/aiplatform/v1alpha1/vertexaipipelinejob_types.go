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
	common "github.com/GoogleCloudPlatform/k8s-config-connector/apis/common"
	refsv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/apis/refs/v1beta1"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var VertexAIPipelineJobGVK = GroupVersion.WithKind("VertexAIPipelineJob")

// VertexAIPipelineJobSpec defines the desired state of VertexAIPipelineJob
// +kcc:spec:proto=google.cloud.aiplatform.v1.PipelineJob
type VertexAIPipelineJobSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location string `json:"location"`

	// The VertexAIPipelineJob name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The display name of the Pipeline.
	//  The name can be up to 128 characters long and can consist of any UTF-8
	//  characters.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// The spec of the pipeline.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.pipeline_spec
	PipelineSpec apiextensionsv1.JSON `json:"pipelineSpec,omitempty"`

	// The labels with user-defined metadata to organize PipelineJob.
	//
	//  Label keys and values can be no longer than 64 characters
	//  (Unicode codepoints), can only contain lowercase letters, numeric
	//  characters, underscores and dashes. International characters are allowed.
	//
	//  See https://goo.gl/xmQnxf for more information and examples of labels.
	//
	//  Note there is some reserved label key for Vertex AI Pipelines.
	//  - `vertex-ai-pipelines-run-billing-id`, user set value will get overrided.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Runtime config of the pipeline.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.runtime_config
	RuntimeConfig *PipelineJob_RuntimeConfig `json:"runtimeConfig,omitempty"`

	// Customer-managed encryption key spec for a pipelineJob. If set, this
	//  PipelineJob and all of its sub-resources will be secured by this key.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.encryption_spec
	EncryptionSpec *EncryptionSpec `json:"encryptionSpec,omitempty"`

	// The service account that the pipeline workload runs as.
	//  If not specified, the Compute Engine default service account in the project
	//  will be used.
	//  See
	//  https://cloud.google.com/compute/docs/access/service-accounts#default_service_account
	//
	//  Users starting the pipeline must have the `iam.serviceAccounts.actAs`
	//  permission on this service account.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.service_account
	ServiceAccount *string `json:"serviceAccount,omitempty"`

	// The full name of the Compute Engine
	//  [network](/compute/docs/networks-and-firewalls#networks) to which the
	//  Pipeline Job's workload should be peered. For example,
	//  `projects/12345/global/networks/myVPC`.
	//  [Format](/compute/docs/reference/rest/v1/networks/insert)
	//  is of the form `projects/{project}/global/networks/{network}`.
	//  Where {project} is a project number, as in `12345`, and {network} is a
	//  network name.
	//
	//  Private services access must already be configured for the network.
	//  Pipeline job will apply the network configuration to the Google Cloud
	//  resources being launched, if applied, such as Vertex AI
	//  Training or Dataflow job. If left unspecified, the workload is not peered
	//  with any network.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.network
	Network *string `json:"network,omitempty"`

	// A list of names for the reserved ip ranges under the VPC network
	//  that can be used for this Pipeline Job's workload.
	//
	//  If set, we will deploy the Pipeline Job's workload within the provided ip
	//  ranges. Otherwise, the job will be deployed to any ip ranges under the
	//  provided VPC network.
	//
	//  Example: ['vertex-ai-ip-range'].
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.reserved_ip_ranges
	ReservedIPRanges []string `json:"reservedIPRanges,omitempty"`

	// Optional. Configuration for PSC-I for PipelineJob.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.psc_interface_config
	PSCInterfaceConfig *PSCInterfaceConfig `json:"pscInterfaceConfig,omitempty"`

	// A template uri from where the
	//  [PipelineJob.pipeline_spec][google.cloud.aiplatform.v1.PipelineJob.pipeline_spec],
	//  if empty, will be downloaded. Currently, only uri from Vertex Template
	//  Registry & Gallery is supported. Reference to
	//  https://cloud.google.com/vertex-ai/docs/pipelines/create-pipeline-template.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.template_uri
	TemplateURI *string `json:"templateURI,omitempty"`

	// Optional. Whether to do component level validations before job creation.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.preflight_validations
	PreflightValidations *bool `json:"preflightValidations,omitempty"`
}

// VertexAIPipelineJobStatus defines the config connector machine state of VertexAIPipelineJob
type VertexAIPipelineJobStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the VertexAIPipelineJob resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *VertexAIPipelineJobObservedState `json:"observedState,omitempty"`
}

// VertexAIPipelineJobObservedState is the state of the VertexAIPipelineJob resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.aiplatform.v1.PipelineJob
type VertexAIPipelineJobObservedState struct {
	// Output only. Pipeline creation time.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Pipeline start time.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.start_time
	StartTime *string `json:"startTime,omitempty"`

	// Output only. Pipeline end time.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.end_time
	EndTime *string `json:"endTime,omitempty"`

	// Output only. Timestamp when this PipelineJob was most recently updated.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The detailed state of the job.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.state
	State *string `json:"state,omitempty"`

	// Output only. The details of pipeline run. Not available in the list view.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.job_detail
	JobDetail *PipelineJobDetailObservedState `json:"jobDetail,omitempty"`

	// Output only. The error that occurred during pipeline execution.
	//  Only populated when the pipeline's state is FAILED or CANCELLED.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.error
	Error *common.Status `json:"error,omitempty"`

	// Output only. Pipeline template metadata. Will fill up fields if
	//  [PipelineJob.template_uri][google.cloud.aiplatform.v1.PipelineJob.template_uri]
	//  is from supported template registry.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.template_metadata
	TemplateMetadata *PipelineTemplateMetadata `json:"templateMetadata,omitempty"`

	// Output only. The schedule resource name.
	//  Only returned if the Pipeline is created by Schedule API.
	// +kcc:proto:field=google.cloud.aiplatform.v1.PipelineJob.schedule_name
	ScheduleName *string `json:"scheduleName,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpvertexaipipelinejob;gcpvertexaipipelinejobs
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// VertexAIPipelineJob is the Schema for the VertexAIPipelineJob API
// +k8s:openapi-gen=true
type VertexAIPipelineJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   VertexAIPipelineJobSpec   `json:"spec,omitempty"`
	Status VertexAIPipelineJobStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VertexAIPipelineJobList contains a list of VertexAIPipelineJob
type VertexAIPipelineJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VertexAIPipelineJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VertexAIPipelineJob{}, &VertexAIPipelineJobList{})
}
