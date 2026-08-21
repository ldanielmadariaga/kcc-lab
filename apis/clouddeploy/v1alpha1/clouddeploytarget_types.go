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

var CloudDeployTargetGVK = GroupVersion.WithKind("CloudDeployTarget")

// CloudDeployTargetSpec defines the desired state of CloudDeployTarget
// +kcc:spec:proto=google.cloud.deploy.v1.Target
type CloudDeployTargetSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The CloudDeployTarget name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Description of the `Target`. Max length is 255 characters.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.description
	Description *string `json:"description,omitempty"`

	// Optional. User annotations. These attributes can only be set and used by
	//  the user, and not by Cloud Deploy. See
	//  https://google.aip.dev/128#annotations for more details such as format and
	//  size limitations.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Optional. Labels are attributes that can be set and used by both the
	//  user and by Cloud Deploy. Labels must meet the following constraints:
	//
	//  * Keys and values can contain only lowercase letters, numeric characters,
	//  underscores, and dashes.
	//  * All characters must use UTF-8 encoding, and international characters are
	//  allowed.
	//  * Keys must start with a lowercase letter or international character.
	//  * Each resource is limited to a maximum of 64 labels.
	//
	//  Both keys and values are additionally constrained to be <= 128 bytes.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. Whether or not the `Target` requires approval.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.require_approval
	RequireApproval *bool `json:"requireApproval,omitempty"`

	// Optional. Information specifying a GKE Cluster.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.gke
	GKE *GKECluster `json:"gke,omitempty"`

	// Optional. Information specifying an Anthos Cluster.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.anthos_cluster
	AnthosCluster *AnthosCluster `json:"anthosCluster,omitempty"`

	// Optional. Information specifying a Cloud Run deployment target.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.run
	Run *CloudRunLocation `json:"run,omitempty"`

	// Optional. Information specifying a multiTarget.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.multi_target
	MultiTarget *MultiTarget `json:"multiTarget,omitempty"`

	// Optional. Information specifying a Custom Target.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.custom_target
	CustomTarget *CustomTarget `json:"customTarget,omitempty"`

	// Optional. Map of entity IDs to their associated entities. Associated
	//  entities allows specifying places other than the deployment target for
	//  specific features. For example, the Gateway API canary can be configured to
	//  deploy the HTTPRoute to a different cluster(s) than the deployment cluster
	//  using associated entities. An entity ID must consist of lower-case letters,
	//  numbers, and hyphens, start with a letter and end with a letter or a
	//  number, and have a max length of 63 characters. In other words, it must
	//  match the following regex: `^[a-z]([a-z0-9-]{0,61}[a-z0-9])?$`.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.associated_entities
	AssociatedEntities map[string]AssociatedEntities `json:"associatedEntities,omitempty"`

	// Optional. This checksum is computed by the server based on the value of
	//  other fields, and may be sent on update and delete requests to ensure the
	//  client has an up-to-date value before proceeding.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.etag
	Etag *string `json:"etag,omitempty"`

	// Optional. Configurations for all execution that relates to this `Target`.
	//  Each `ExecutionEnvironmentUsage` value may only be used in a single
	//  configuration; using the same value multiple times is an error.
	//  When one or more configurations are specified, they must include the
	//  `RENDER` and `DEPLOY` `ExecutionEnvironmentUsage` values.
	//  When no configurations are specified, execution will use the default
	//  specified in `DefaultPool`.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.execution_configs
	ExecutionConfigs []ExecutionConfig `json:"executionConfigs,omitempty"`

	// Optional. The deploy parameters to use for this target.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.deploy_parameters
	DeployParameters map[string]string `json:"deployParameters,omitempty"`
}

// CloudDeployTargetStatus defines the config connector machine state of CloudDeployTarget
type CloudDeployTargetStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the CloudDeployTarget resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *CloudDeployTargetObservedState `json:"observedState,omitempty"`
}

// CloudDeployTargetObservedState is the state of the CloudDeployTarget resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.deploy.v1.Target
type CloudDeployTargetObservedState struct {
	// Output only. Resource id of the `Target`.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.target_id
	TargetID *string `json:"targetID,omitempty"`

	// Output only. Unique identifier of the `Target`.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. Time at which the `Target` was created.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Most recent time at which the `Target` was updated.
	// +kcc:proto:field=google.cloud.deploy.v1.Target.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpclouddeploytarget;gcpclouddeploytargets
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// CloudDeployTarget is the Schema for the CloudDeployTarget API
// +k8s:openapi-gen=true
type CloudDeployTarget struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   CloudDeployTargetSpec   `json:"spec,omitempty"`
	Status CloudDeployTargetStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// CloudDeployTargetList contains a list of CloudDeployTarget
type CloudDeployTargetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudDeployTarget `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudDeployTarget{}, &CloudDeployTargetList{})
}
