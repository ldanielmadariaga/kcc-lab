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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var BigQueryReservationCapacityCommitmentGVK = GroupVersion.WithKind("BigQueryReservationCapacityCommitment")

// BigQueryReservationCapacityCommitmentSpec defines the desired state of BigQueryReservationCapacityCommitment
// +kcc:spec:proto=google.cloud.bigquery.reservation.v1.CapacityCommitment
type BigQueryReservationCapacityCommitmentSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location string `json:"location"`

	// The BigQueryReservationCapacityCommitment name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Number of slots in this commitment.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.slot_count
	SlotCount *int64 `json:"slotCount,omitempty"`

	// Capacity commitment commitment plan.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.plan
	Plan *string `json:"plan,omitempty"`

	// The plan this capacity commitment is converted to after commitment_end_time
	//  passes. Once the plan is changed, committed period is extended according to
	//  commitment plan. Only applicable for ANNUAL and TRIAL commitments.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.renewal_plan
	RenewalPlan *string `json:"renewalPlan,omitempty"`

	// Applicable only for commitments located within one of the BigQuery
	//  multi-regions (US or EU).
	//
	//  If set to true, this commitment is placed in the organization's
	//  secondary region which is designated for disaster recovery purposes.
	//  If false, this commitment is placed in the organization's default region.
	//
	//  NOTE: this is a preview feature. Project must be allow-listed in order to
	//  set this field.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.multi_region_auxiliary
	MultiRegionAuxiliary *bool `json:"multiRegionAuxiliary,omitempty"`

	// Edition of the capacity commitment.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.edition
	Edition *string `json:"edition,omitempty"`
}

// BigQueryReservationCapacityCommitmentStatus defines the config connector machine state of BigQueryReservationCapacityCommitment
type BigQueryReservationCapacityCommitmentStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the BigQueryReservationCapacityCommitment resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *BigQueryReservationCapacityCommitmentObservedState `json:"observedState,omitempty"`
}

// BigQueryReservationCapacityCommitmentObservedState is the state of the BigQueryReservationCapacityCommitment resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.bigquery.reservation.v1.CapacityCommitment
type BigQueryReservationCapacityCommitmentObservedState struct {
	// Output only. State of the commitment.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.state
	State *string `json:"state,omitempty"`

	// Output only. The start of the current commitment period. It is applicable
	//  only for ACTIVE capacity commitments. Note after the commitment is renewed,
	//  commitment_start_time won't be changed. It refers to the start time of the
	//  original commitment.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.commitment_start_time
	CommitmentStartTime *string `json:"commitmentStartTime,omitempty"`

	// Output only. The end of the current commitment period. It is applicable
	//  only for ACTIVE capacity commitments. Note after renewal,
	//  commitment_end_time is the time the renewed commitment expires. So it would
	//  be at a time after commitment_start_time + committed period, because we
	//  don't change commitment_start_time ,
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.commitment_end_time
	CommitmentEndTime *string `json:"commitmentEndTime,omitempty"`

	// Output only. For FAILED commitment plan, provides the reason of failure.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.failure_status
	FailureStatus *common.Status `json:"failureStatus,omitempty"`

	// Output only. If true, the commitment is a flat-rate commitment, otherwise,
	//  it's an edition commitment.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.CapacityCommitment.is_flat_rate
	IsFlatRate *bool `json:"isFlatRate,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpbigqueryreservationcapacitycommitment;gcpbigqueryreservationcapacitycommitments
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// BigQueryReservationCapacityCommitment is the Schema for the BigQueryReservationCapacityCommitment API
// +k8s:openapi-gen=true
type BigQueryReservationCapacityCommitment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   BigQueryReservationCapacityCommitmentSpec   `json:"spec,omitempty"`
	Status BigQueryReservationCapacityCommitmentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// BigQueryReservationCapacityCommitmentList contains a list of BigQueryReservationCapacityCommitment
type BigQueryReservationCapacityCommitmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BigQueryReservationCapacityCommitment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BigQueryReservationCapacityCommitment{}, &BigQueryReservationCapacityCommitmentList{})
}
