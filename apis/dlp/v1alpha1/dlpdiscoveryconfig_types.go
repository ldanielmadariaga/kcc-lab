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

var DLPDiscoveryConfigGVK = GroupVersion.WithKind("DLPDiscoveryConfig")

// DLPDiscoveryConfigSpec defines the desired state of DLPDiscoveryConfig
// +kcc:spec:proto=google.privacy.dlp.v2.DiscoveryConfig
type DLPDiscoveryConfigSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The DLPDiscoveryConfig name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Display name (max 100 chars)
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Only set when the parent is an org.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.org_config
	OrgConfig *DiscoveryConfig_OrgConfig `json:"orgConfig,omitempty"`

	// Must be set only when scanning other clouds.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.other_cloud_starting_location
	OtherCloudStartingLocation *OtherCloudDiscoveryStartingLocation `json:"otherCloudStartingLocation,omitempty"`

	// Detection logic for profile generation.
	//
	//  Not all template features are used by Discovery. FindingLimits,
	//  include_quote and exclude_info_types have no impact on
	//  Discovery.
	//
	//  Multiple templates may be provided if there is data in multiple regions.
	//  At most one template must be specified per-region (including "global").
	//  Each region is scanned using the applicable template. If no region-specific
	//  template is specified, but a "global" template is specified, it will be
	//  copied to that region and used instead. If no global or region-specific
	//  template is provided for a region with data, that region's data will not be
	//  scanned.
	//
	//  For more information, see
	//  https://cloud.google.com/sensitive-data-protection/docs/data-profiles#data-residency.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.inspect_templates
	InspectTemplates []string `json:"inspectTemplates,omitempty"`

	// Actions to execute at the completion of scanning.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.actions
	Actions []DataProfileAction `json:"actions,omitempty"`

	// Target to match against for determining what to scan and how frequently.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.targets
	Targets []DiscoveryTarget `json:"targets,omitempty"`

	// Required. A status for this configuration.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.status
	// +required
	Status *string `json:"status,omitempty"`

	// Optional. Processing location configuration. Vertex AI dataset scanning
	//  will set processing_location.image_fallback_type to MultiRegionProcessing
	//  by default.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.processing_location
	ProcessingLocation *ProcessingLocation `json:"processingLocation,omitempty"`
}

// DLPDiscoveryConfigStatus defines the config connector machine state of DLPDiscoveryConfig
type DLPDiscoveryConfigStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DLPDiscoveryConfig resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DLPDiscoveryConfigObservedState `json:"observedState,omitempty"`
}

// DLPDiscoveryConfigObservedState is the state of the DLPDiscoveryConfig resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.privacy.dlp.v2.DiscoveryConfig
type DLPDiscoveryConfigObservedState struct {
	// Target to match against for determining what to scan and how frequently.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.targets
	Targets []DiscoveryTargetObservedState `json:"targets,omitempty"`

	// Output only. A stream of errors encountered when the config was activated.
	//  Repeated errors may result in the config automatically being paused. Output
	//  only field. Will return the last 100 errors. Whenever the config is
	//  modified this list will be cleared.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.errors
	Errors []Error `json:"errors,omitempty"`

	// Output only. The creation timestamp of a DiscoveryConfig.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The last update timestamp of a DiscoveryConfig.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The timestamp of the last time this config was executed.
	// +kcc:proto:field=google.privacy.dlp.v2.DiscoveryConfig.last_run_time
	LastRunTime *string `json:"lastRunTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdlpdiscoveryconfig;gcpdlpdiscoveryconfigs
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DLPDiscoveryConfig is the Schema for the DLPDiscoveryConfig API
// +k8s:openapi-gen=true
type DLPDiscoveryConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DLPDiscoveryConfigSpec   `json:"spec,omitempty"`
	Status DLPDiscoveryConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DLPDiscoveryConfigList contains a list of DLPDiscoveryConfig
type DLPDiscoveryConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DLPDiscoveryConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DLPDiscoveryConfig{}, &DLPDiscoveryConfigList{})
}
