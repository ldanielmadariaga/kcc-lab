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

var KMSCryptoKeyVersionGVK = GroupVersion.WithKind("KMSCryptoKeyVersion")

// KMSCryptoKeyVersionSpec defines the desired state of KMSCryptoKeyVersion
// +kcc:spec:proto=google.cloud.kms.v1.CryptoKeyVersion
type KMSCryptoKeyVersionSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`


	// The KMSCryptoKeyVersion name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The current state of the
	//  [CryptoKeyVersion][google.cloud.kms.v1.CryptoKeyVersion].
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.state
	State *string `json:"state,omitempty"`

	// ExternalProtectionLevelOptions stores a group of additional fields for
	//  configuring a [CryptoKeyVersion][google.cloud.kms.v1.CryptoKeyVersion] that
	//  are specific to the
	//  [EXTERNAL][google.cloud.kms.v1.ProtectionLevel.EXTERNAL] protection level
	//  and [EXTERNAL_VPC][google.cloud.kms.v1.ProtectionLevel.EXTERNAL_VPC]
	//  protection levels.
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.external_protection_level_options
	ExternalProtectionLevelOptions *ExternalProtectionLevelOptions `json:"externalProtectionLevelOptions,omitempty"`
}

// KMSCryptoKeyVersionStatus defines the config connector machine state of KMSCryptoKeyVersion
type KMSCryptoKeyVersionStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the KMSCryptoKeyVersion resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *KMSCryptoKeyVersionObservedState `json:"observedState,omitempty"`
}

// KMSCryptoKeyVersionObservedState is the state of the KMSCryptoKeyVersion resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.kms.v1.CryptoKeyVersion
type KMSCryptoKeyVersionObservedState struct {
	// Output only. The [ProtectionLevel][google.cloud.kms.v1.ProtectionLevel]
	//  describing how crypto operations are performed with this
	//  [CryptoKeyVersion][google.cloud.kms.v1.CryptoKeyVersion].
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.protection_level
	ProtectionLevel *string `json:"protectionLevel,omitempty"`

	// Output only. The
	//  [CryptoKeyVersionAlgorithm][google.cloud.kms.v1.CryptoKeyVersion.CryptoKeyVersionAlgorithm]
	//  that this [CryptoKeyVersion][google.cloud.kms.v1.CryptoKeyVersion]
	//  supports.
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.algorithm
	Algorithm *string `json:"algorithm,omitempty"`

	// Output only. Statement that was generated and signed by the HSM at key
	//  creation time. Use this statement to verify attributes of the key as stored
	//  on the HSM, independently of Google. Only provided for key versions with
	//  [protection_level][google.cloud.kms.v1.CryptoKeyVersion.protection_level]
	//  [HSM][google.cloud.kms.v1.ProtectionLevel.HSM].
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.attestation
	Attestation *KeyOperationAttestationObservedState `json:"attestation,omitempty"`

	// Output only. The time at which this
	//  [CryptoKeyVersion][google.cloud.kms.v1.CryptoKeyVersion] was created.
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The time this
	//  [CryptoKeyVersion][google.cloud.kms.v1.CryptoKeyVersion]'s key material was
	//  generated.
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.generate_time
	GenerateTime *string `json:"generateTime,omitempty"`

	// Output only. The time this
	//  [CryptoKeyVersion][google.cloud.kms.v1.CryptoKeyVersion]'s key material is
	//  scheduled for destruction. Only present if
	//  [state][google.cloud.kms.v1.CryptoKeyVersion.state] is
	//  [DESTROY_SCHEDULED][google.cloud.kms.v1.CryptoKeyVersion.CryptoKeyVersionState.DESTROY_SCHEDULED].
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.destroy_time
	DestroyTime *string `json:"destroyTime,omitempty"`

	// Output only. The time this CryptoKeyVersion's key material was
	//  destroyed. Only present if
	//  [state][google.cloud.kms.v1.CryptoKeyVersion.state] is
	//  [DESTROYED][google.cloud.kms.v1.CryptoKeyVersion.CryptoKeyVersionState.DESTROYED].
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.destroy_event_time
	DestroyEventTime *string `json:"destroyEventTime,omitempty"`

	// Output only. The name of the [ImportJob][google.cloud.kms.v1.ImportJob]
	//  used in the most recent import of this
	//  [CryptoKeyVersion][google.cloud.kms.v1.CryptoKeyVersion]. Only present if
	//  the underlying key material was imported.
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.import_job
	ImportJob *string `json:"importJob,omitempty"`

	// Output only. The time at which this
	//  [CryptoKeyVersion][google.cloud.kms.v1.CryptoKeyVersion]'s key material was
	//  most recently imported.
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.import_time
	ImportTime *string `json:"importTime,omitempty"`

	// Output only. The root cause of the most recent import failure. Only present
	//  if [state][google.cloud.kms.v1.CryptoKeyVersion.state] is
	//  [IMPORT_FAILED][google.cloud.kms.v1.CryptoKeyVersion.CryptoKeyVersionState.IMPORT_FAILED].
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.import_failure_reason
	ImportFailureReason *string `json:"importFailureReason,omitempty"`

	// Output only. The root cause of the most recent generation failure. Only
	//  present if [state][google.cloud.kms.v1.CryptoKeyVersion.state] is
	//  [GENERATION_FAILED][google.cloud.kms.v1.CryptoKeyVersion.CryptoKeyVersionState.GENERATION_FAILED].
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.generation_failure_reason
	GenerationFailureReason *string `json:"generationFailureReason,omitempty"`

	// Output only. The root cause of the most recent external destruction
	//  failure. Only present if
	//  [state][google.cloud.kms.v1.CryptoKeyVersion.state] is
	//  [EXTERNAL_DESTRUCTION_FAILED][google.cloud.kms.v1.CryptoKeyVersion.CryptoKeyVersionState.EXTERNAL_DESTRUCTION_FAILED].
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.external_destruction_failure_reason
	ExternalDestructionFailureReason *string `json:"externalDestructionFailureReason,omitempty"`

	// Output only. Whether or not this key version is eligible for reimport, by
	//  being specified as a target in
	//  [ImportCryptoKeyVersionRequest.crypto_key_version][google.cloud.kms.v1.ImportCryptoKeyVersionRequest.crypto_key_version].
	// +kcc:proto:field=google.cloud.kms.v1.CryptoKeyVersion.reimport_eligible
	ReimportEligible *bool `json:"reimportEligible,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpkmscryptokeyversion;gcpkmscryptokeyversions
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// KMSCryptoKeyVersion is the Schema for the KMSCryptoKeyVersion API
// +k8s:openapi-gen=true
type KMSCryptoKeyVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   KMSCryptoKeyVersionSpec   `json:"spec,omitempty"`
	Status KMSCryptoKeyVersionStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// KMSCryptoKeyVersionList contains a list of KMSCryptoKeyVersion
type KMSCryptoKeyVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KMSCryptoKeyVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KMSCryptoKeyVersion{}, &KMSCryptoKeyVersionList{})
}
