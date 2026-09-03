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

var DialogflowSecuritySettingsGVK = GroupVersion.WithKind("DialogflowSecuritySettings")

// DialogflowSecuritySettingsSpec defines the desired state of DialogflowSecuritySettings
// +kcc:spec:proto=google.cloud.dialogflow.cx.v3.SecuritySettings
type DialogflowSecuritySettingsSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/securitySettings/{security_settings}
	Location *string `json:"location"`

	// The DialogflowSecuritySettings name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Required. The human-readable name of the security settings, unique within
	//  the location.
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// Strategy that defines how we do redaction.
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.redaction_strategy
	RedactionStrategy *string `json:"redactionStrategy,omitempty"`

	// Defines the data for which Dialogflow applies redaction. Dialogflow does
	//  not redact data that it does not have access to – for example, Cloud
	//  logging.
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.redaction_scope
	RedactionScope *string `json:"redactionScope,omitempty"`

	// [DLP](https://cloud.google.com/dlp/docs) inspect template name. Use this
	//  template to define inspect base settings.
	//
	//  The `DLP Inspect Templates Reader` role is needed on the Dialogflow
	//  service identity service account (has the form
	//  `service-PROJECT_NUMBER@gcp-sa-dialogflow.iam.gserviceaccount.com`)
	//  for your agent's project.
	//
	//  If empty, we use the default DLP inspect config.
	//
	//  The template name will have one of the following formats:
	//  `projects/<ProjectID>/locations/<LocationID>/inspectTemplates/<TemplateID>`
	//  OR
	//  `organizations/<OrganizationID>/locations/<LocationID>/inspectTemplates/<TemplateID>`
	//
	//  Note: `inspect_template` must be located in the same region as the
	//  `SecuritySettings`.
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.inspect_template
	InspectTemplate *string `json:"inspectTemplate,omitempty"`

	// [DLP](https://cloud.google.com/dlp/docs) deidentify template name. Use this
	//  template to define de-identification configuration for the content.
	//
	//  The `DLP De-identify Templates Reader` role is needed on the Dialogflow
	//  service identity service account (has the form
	//  `service-PROJECT_NUMBER@gcp-sa-dialogflow.iam.gserviceaccount.com`)
	//  for your agent's project.
	//
	//  If empty, Dialogflow replaces sensitive info with `[redacted]` text.
	//
	//  The template name will have one of the following formats:
	//  `projects/<ProjectID>/locations/<LocationID>/deidentifyTemplates/<TemplateID>`
	//  OR
	//  `organizations/<OrganizationID>/locations/<LocationID>/deidentifyTemplates/<TemplateID>`
	//
	//  Note: `deidentify_template` must be located in the same region as the
	//  `SecuritySettings`.
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.deidentify_template
	DeidentifyTemplate *string `json:"deidentifyTemplate,omitempty"`

	// Retains the data for the specified number of days.
	//  User must set a value lower than Dialogflow's default 365d TTL (30 days
	//  for Agent Assist traffic), higher value will be ignored and use default.
	//  Setting a value higher than that has no effect. A missing value or
	//  setting to 0 also means we use default TTL.
	//  When data retention configuration is changed, it only applies to the data
	//  created after the change; the TTL of existing data created before the
	//  change stays intact.
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.retention_window_days
	RetentionWindowDays *int32 `json:"retentionWindowDays,omitempty"`

	// Specifies the retention behavior defined by
	//  [SecuritySettings.RetentionStrategy][google.cloud.dialogflow.cx.v3.SecuritySettings.RetentionStrategy].
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.retention_strategy
	RetentionStrategy *string `json:"retentionStrategy,omitempty"`

	// List of types of data to remove when retention settings triggers purge.
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.purge_data_types
	PurgeDataTypes []string `json:"purgeDataTypes,omitempty"`

	// Controls audio export settings for post-conversation analytics when
	//  ingesting audio to conversations via [Participants.AnalyzeContent][] or
	//  [Participants.StreamingAnalyzeContent][].
	//
	//  If
	//  [retention_strategy][google.cloud.dialogflow.cx.v3.SecuritySettings.retention_strategy]
	//  is set to REMOVE_AFTER_CONVERSATION or [audio_export_settings.gcs_bucket][]
	//  is empty, audio export is disabled.
	//
	//  If audio export is enabled, audio is recorded and saved to
	//  [audio_export_settings.gcs_bucket][], subject to retention policy of
	//  [audio_export_settings.gcs_bucket][].
	//
	//  This setting won't effect audio input for implicit sessions via
	//  [Sessions.DetectIntent][google.cloud.dialogflow.cx.v3.Sessions.DetectIntent]
	//  or
	//  [Sessions.StreamingDetectIntent][google.cloud.dialogflow.cx.v3.Sessions.StreamingDetectIntent].
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.audio_export_settings
	AudioExportSettings *SecuritySettings_AudioExportSettings `json:"audioExportSettings,omitempty"`

	// Controls conversation exporting settings to Insights after conversation is
	//  completed.
	//
	//  If
	//  [retention_strategy][google.cloud.dialogflow.cx.v3.SecuritySettings.retention_strategy]
	//  is set to REMOVE_AFTER_CONVERSATION, Insights export is disabled no matter
	//  what you configure here.
	// +kcc:proto:field=google.cloud.dialogflow.cx.v3.SecuritySettings.insights_export_settings
	InsightsExportSettings *SecuritySettings_InsightsExportSettings `json:"insightsExportSettings,omitempty"`
}

// DialogflowSecuritySettingsStatus defines the config connector machine state of DialogflowSecuritySettings
type DialogflowSecuritySettingsStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DialogflowSecuritySettings resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DialogflowSecuritySettingsObservedState `json:"observedState,omitempty"`
}

// DialogflowSecuritySettingsObservedState is the state of the DialogflowSecuritySettings resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.dialogflow.cx.v3.SecuritySettings
type DialogflowSecuritySettingsObservedState struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdialogflowsecuritysettings;gcpdialogflowsecuritysettingss
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DialogflowSecuritySettings is the Schema for the DialogflowSecuritySettings API
// +k8s:openapi-gen=true
type DialogflowSecuritySettings struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DialogflowSecuritySettingsSpec   `json:"spec,omitempty"`
	Status DialogflowSecuritySettingsStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DialogflowSecuritySettingsList contains a list of DialogflowSecuritySettings
type DialogflowSecuritySettingsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DialogflowSecuritySettings `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DialogflowSecuritySettings{}, &DialogflowSecuritySettingsList{})
}
