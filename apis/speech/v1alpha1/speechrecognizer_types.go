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

var SpeechRecognizerGVK = GroupVersion.WithKind("SpeechRecognizer")

// SpeechRecognizerSpec defines the desired state of SpeechRecognizer
// +kcc:spec:proto=google.cloud.speech.v2.Recognizer
type SpeechRecognizerSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/recognizers/{recognizer}
	Location *string `json:"location"`

	// The SpeechRecognizer name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// User-settable, human-readable name for the Recognizer. Must be 63
	//  characters or less.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. This field is now deprecated. Prefer the
	//  [`model`][google.cloud.speech.v2.RecognitionConfig.model] field in the
	//  [`RecognitionConfig`][google.cloud.speech.v2.RecognitionConfig] message.
	//
	//  Which model to use for recognition requests. Select the model best suited
	//  to your domain to get best results.
	//
	//  Guidance for choosing which model to use can be found in the [Transcription
	//  Models
	//  Documentation](https://cloud.google.com/speech-to-text/v2/docs/transcription-model)
	//  and the models supported in each region can be found in the [Table Of
	//  Supported
	//  Models](https://cloud.google.com/speech-to-text/v2/docs/speech-to-text-supported-languages).
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.model
	Model *string `json:"model,omitempty"`

	// Optional. This field is now deprecated. Prefer the
	//  [`language_codes`][google.cloud.speech.v2.RecognitionConfig.language_codes]
	//  field in the
	//  [`RecognitionConfig`][google.cloud.speech.v2.RecognitionConfig] message.
	//
	//  The language of the supplied audio as a
	//  [BCP-47](https://www.rfc-editor.org/rfc/bcp/bcp47.txt) language tag.
	//
	//  Supported languages for each model are listed in the [Table of Supported
	//  Models](https://cloud.google.com/speech-to-text/v2/docs/speech-to-text-supported-languages).
	//
	//  If additional languages are provided, recognition result will contain
	//  recognition in the most likely language detected. The recognition result
	//  will include the language tag of the language detected in the audio.
	//  When you create or update a Recognizer, these values are
	//  stored in normalized BCP-47 form. For example, "en-us" is stored as
	//  "en-US".
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.language_codes
	LanguageCodes []string `json:"languageCodes,omitempty"`

	// Default configuration to use for requests with this Recognizer.
	//  This can be overwritten by inline configuration in the
	//  [RecognizeRequest.config][google.cloud.speech.v2.RecognizeRequest.config]
	//  field.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.default_recognition_config
	DefaultRecognitionConfig *RecognitionConfig `json:"defaultRecognitionConfig,omitempty"`

	// Allows users to store small amounts of arbitrary data.
	//  Both the key and the value must be 63 characters or less each.
	//  At most 100 annotations.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.annotations
	Annotations map[string]string `json:"annotations,omitempty"`
}

// SpeechRecognizerStatus defines the config connector machine state of SpeechRecognizer
type SpeechRecognizerStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the SpeechRecognizer resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *SpeechRecognizerObservedState `json:"observedState,omitempty"`
}

// SpeechRecognizerObservedState is the state of the SpeechRecognizer resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.speech.v2.Recognizer
type SpeechRecognizerObservedState struct {
	// Output only. System-assigned unique identifier for the Recognizer.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.uid
	Uid *string `json:"uid,omitempty"`

	// Default configuration to use for requests with this Recognizer.
	//  This can be overwritten by inline configuration in the
	//  [RecognizeRequest.config][google.cloud.speech.v2.RecognizeRequest.config]
	//  field.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.default_recognition_config
	DefaultRecognitionConfig *RecognitionConfigObservedState `json:"defaultRecognitionConfig,omitempty"`

	// Output only. The Recognizer lifecycle state.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.state
	State *string `json:"state,omitempty"`

	// Output only. Creation time.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The most recent time this Recognizer was modified.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The time at which this Recognizer was requested for deletion.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.delete_time
	DeleteTime *string `json:"deleteTime,omitempty"`

	// Output only. The time at which this Recognizer will be purged.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.expire_time
	ExpireTime *string `json:"expireTime,omitempty"`

	// Output only. This checksum is computed by the server based on the value of
	//  other fields. This may be sent on update, undelete, and delete requests to
	//  ensure the client has an up-to-date value before proceeding.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.etag
	Etag *string `json:"etag,omitempty"`

	// Output only. Whether or not this Recognizer is in the process of being
	//  updated.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.reconciling
	Reconciling *bool `json:"reconciling,omitempty"`

	// Output only. The [KMS key
	//  name](https://cloud.google.com/kms/docs/resource-hierarchy#keys) with which
	//  the Recognizer is encrypted. The expected format is
	//  `projects/{project}/locations/{location}/keyRings/{key_ring}/cryptoKeys/{crypto_key}`.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.kms_key_name
	KMSKeyName *string `json:"kmsKeyName,omitempty"`

	// Output only. The [KMS key version
	//  name](https://cloud.google.com/kms/docs/resource-hierarchy#key_versions)
	//  with which the Recognizer is encrypted. The expected format is
	//  `projects/{project}/locations/{location}/keyRings/{key_ring}/cryptoKeys/{crypto_key}/cryptoKeyVersions/{crypto_key_version}`.
	// +kcc:proto:field=google.cloud.speech.v2.Recognizer.kms_key_version_name
	KMSKeyVersionName *string `json:"kmsKeyVersionName,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpspeechrecognizer;gcpspeechrecognizers
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// SpeechRecognizer is the Schema for the SpeechRecognizer API
// +k8s:openapi-gen=true
type SpeechRecognizer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   SpeechRecognizerSpec   `json:"spec,omitempty"`
	Status SpeechRecognizerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// SpeechRecognizerList contains a list of SpeechRecognizer
type SpeechRecognizerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SpeechRecognizer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SpeechRecognizer{}, &SpeechRecognizerList{})
}
