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

var CloudBuildConnectionGVK = GroupVersion.WithKind("CloudBuildConnection")

// CloudBuildConnectionSpec defines the desired state of CloudBuildConnection
// +kcc:spec:proto=google.devtools.cloudbuild.v2.Connection
type CloudBuildConnectionSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location string `json:"location"`

	// The CloudBuildConnection name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Configuration for connections to github.com.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.github_config
	GithubConfig *GitHubConfig `json:"githubConfig,omitempty"`

	// Configuration for connections to an instance of GitHub Enterprise.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.github_enterprise_config
	GithubEnterpriseConfig *GitHubEnterpriseConfig `json:"githubEnterpriseConfig,omitempty"`

	// Configuration for connections to gitlab.com or an instance of GitLab
	//  Enterprise.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.gitlab_config
	GitlabConfig *GitLabConfig `json:"gitlabConfig,omitempty"`

	// Configuration for connections to Bitbucket Data Center.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.bitbucket_data_center_config
	BitbucketDataCenterConfig *BitbucketDataCenterConfig `json:"bitbucketDataCenterConfig,omitempty"`

	// Configuration for connections to Bitbucket Cloud.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.bitbucket_cloud_config
	BitbucketCloudConfig *BitbucketCloudConfig `json:"bitbucketCloudConfig,omitempty"`

	// If disabled is set to true, functionality is disabled for this connection.
	//  Repository based API methods and webhooks processing for repositories in
	//  this connection will be disabled.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.disabled
	Disabled *bool `json:"disabled,omitempty"`

	// Allows clients to store small amounts of arbitrary data.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// This checksum is computed by the server based on the value of other
	//  fields, and may be sent on update and delete requests to ensure the
	//  client has an up-to-date value before proceeding.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.etag
	Etag *string `json:"etag,omitempty"`
}

// CloudBuildConnectionStatus defines the config connector machine state of CloudBuildConnection
type CloudBuildConnectionStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the CloudBuildConnection resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *CloudBuildConnectionObservedState `json:"observedState,omitempty"`
}

// CloudBuildConnectionObservedState is the state of the CloudBuildConnection resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.devtools.cloudbuild.v2.Connection
type CloudBuildConnectionObservedState struct {
	// Output only. Server assigned timestamp for when the connection was created.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Server assigned timestamp for when the connection was updated.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Configuration for connections to github.com.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.github_config
	GithubConfig *GitHubConfigObservedState `json:"githubConfig,omitempty"`

	// Configuration for connections to an instance of GitHub Enterprise.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.github_enterprise_config
	GithubEnterpriseConfig *GitHubEnterpriseConfigObservedState `json:"githubEnterpriseConfig,omitempty"`

	// Configuration for connections to gitlab.com or an instance of GitLab
	//  Enterprise.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.gitlab_config
	GitlabConfig *GitLabConfigObservedState `json:"gitlabConfig,omitempty"`

	// Configuration for connections to Bitbucket Data Center.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.bitbucket_data_center_config
	BitbucketDataCenterConfig *BitbucketDataCenterConfigObservedState `json:"bitbucketDataCenterConfig,omitempty"`

	// Configuration for connections to Bitbucket Cloud.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.bitbucket_cloud_config
	BitbucketCloudConfig *BitbucketCloudConfigObservedState `json:"bitbucketCloudConfig,omitempty"`

	// Output only. Installation state of the Connection.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.installation_state
	InstallationState *InstallationStateObservedState `json:"installationState,omitempty"`

	// Output only. Set to true when the connection is being set up or updated in
	//  the background.
	// +kcc:proto:field=google.devtools.cloudbuild.v2.Connection.reconciling
	Reconciling *bool `json:"reconciling,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcloudbuildconnection;gcpcloudbuildconnections
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// CloudBuildConnection is the Schema for the CloudBuildConnection API
// +k8s:openapi-gen=true
type CloudBuildConnection struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   CloudBuildConnectionSpec   `json:"spec,omitempty"`
	Status CloudBuildConnectionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// CloudBuildConnectionList contains a list of CloudBuildConnection
type CloudBuildConnectionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CloudBuildConnection `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CloudBuildConnection{}, &CloudBuildConnectionList{})
}
