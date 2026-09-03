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

var DataformRepositoryGVK = GroupVersion.WithKind("DataformRepository")

// DataformRepositorySpec defines the desired state of DataformRepository
// +kcc:spec:proto=google.cloud.dataform.v1beta1.Repository
type DataformRepositorySpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/repositories/{repository}
	Location *string `json:"location"`

	// The DataformRepository name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. The name of the containing folder of the repository.
	//  The field is immutable and it can be modified via a MoveRepository
	//  operation.
	//  Format: `projects/*/locations/*/folders/*`. or
	//  `projects/*/locations/*/teamFolders/*`.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.containing_folder
	ContainingFolder *string `json:"containingFolder,omitempty"`

	// Optional. The repository's user-friendly name.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. If set, configures this repository to be linked to a Git remote.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.git_remote_settings
	GitRemoteSettings *Repository_GitRemoteSettings `json:"gitRemoteSettings,omitempty"`

	// Optional. The name of the Secret Manager secret version to be used to
	//  interpolate variables into the .npmrc file for package installation
	//  operations. Must be in the format `projects/*/secrets/*/versions/*`. The
	//  file itself must be in a JSON format.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.npmrc_environment_variables_secret_version
	NpmrcEnvironmentVariablesSecretVersion *string `json:"npmrcEnvironmentVariablesSecretVersion,omitempty"`

	// Optional. If set, fields of `workspace_compilation_overrides` override the
	//  default compilation settings that are specified in dataform.json when
	//  creating workspace-scoped compilation results. See documentation for
	//  `WorkspaceCompilationOverrides` for more information.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.workspace_compilation_overrides
	WorkspaceCompilationOverrides *Repository_WorkspaceCompilationOverrides `json:"workspaceCompilationOverrides,omitempty"`

	// Optional. Repository user labels.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. Input only. If set to true, the authenticated user will be
	//  granted the roles/dataform.admin role on the created repository. To modify
	//  access to the created repository later apply setIamPolicy from
	//  https://cloud.google.com/dataform/reference/rest#rest-resource:-v1beta1.projects.locations.repositories
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.set_authenticated_user_admin
	SetAuthenticatedUserAdmin *bool `json:"setAuthenticatedUserAdmin,omitempty"`

	// Optional. The service account to run workflow invocations under.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.service_account
	ServiceAccount *string `json:"serviceAccount,omitempty"`

	// Optional. The reference to a KMS encryption key. If provided, it will be
	//  used to encrypt user data in the repository and all child resources. It is
	//  not possible to add or update the encryption key after the repository is
	//  created. Example:
	//  `projects/{kms_project}/locations/{location}/keyRings/{key_location}/cryptoKeys/{key}`
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.kms_key_name
	KMSKeyName *string `json:"kmsKeyName,omitempty"`
}

// DataformRepositoryStatus defines the config connector machine state of DataformRepository
type DataformRepositoryStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DataformRepository resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DataformRepositoryObservedState `json:"observedState,omitempty"`
}

// DataformRepositoryObservedState is the state of the DataformRepository resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.dataform.v1beta1.Repository
type DataformRepositoryObservedState struct {
	// Output only. The resource name of the TeamFolder that this Repository is
	//  associated with. This should take the format:
	//  projects/{project}/locations/{location}/teamFolders/{teamFolder}. If this
	//  is not set, the Repository is not associated with a TeamFolder.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.team_folder_name
	TeamFolderName *string `json:"teamFolderName,omitempty"`

	// Output only. The timestamp of when the repository was created.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Optional. If set, configures this repository to be linked to a Git remote.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.git_remote_settings
	GitRemoteSettings *Repository_GitRemoteSettingsObservedState `json:"gitRemoteSettings,omitempty"`

	// Output only. A data encryption state of a Git repository if this Repository
	//  is protected by a KMS key.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.data_encryption_state
	DataEncryptionState *DataEncryptionStateObservedState `json:"dataEncryptionState,omitempty"`

	// Output only. All the metadata information that is used internally to serve
	//  the resource. For example: timestamps, flags, status fields, etc. The
	//  format of this field is a JSON string.
	// +kcc:proto:field=google.cloud.dataform.v1beta1.Repository.internal_metadata
	InternalMetadata *string `json:"internalMetadata,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdataformrepository;gcpdataformrepositorys
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DataformRepository is the Schema for the DataformRepository API
// +k8s:openapi-gen=true
type DataformRepository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DataformRepositorySpec   `json:"spec,omitempty"`
	Status DataformRepositoryStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DataformRepositoryList contains a list of DataformRepository
type DataformRepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataformRepository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataformRepository{}, &DataformRepositoryList{})
}
