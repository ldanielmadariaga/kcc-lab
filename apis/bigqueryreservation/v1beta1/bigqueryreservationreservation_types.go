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

var BigQueryReservationReservationGVK = GroupVersion.WithKind("BigQueryReservationReservation")

// BigQueryReservationReservationSpec defines the desired state of BigQueryReservationReservation
// +kcc:spec:proto=google.cloud.bigquery.reservation.v1.Reservation
type BigQueryReservationReservationSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/reservations/{reservation}
	Location *string `json:"location"`

	// The BigQueryReservationReservation name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Baseline slots available to this reservation. A slot is a unit of
	//  computational power in BigQuery, and serves as the unit of parallelism.
	//
	//  Queries using this reservation might use more slots during runtime if
	//  ignore_idle_slots is set to false, or autoscaling is enabled.
	//
	//  The total slot_capacity of the reservation and its siblings
	//  may exceed the total slot_count of capacity commitments. In that case, the
	//  exceeding slots will be charged with the autoscale SKU. You can increase
	//  the number of baseline slots in a reservation every few minutes. If you
	//  want to decrease your baseline slots, you are limited to once an hour if
	//  you have recently changed your baseline slot capacity and your baseline
	//  slots exceed your committed slots. Otherwise, you can decrease your
	//  baseline slots every few minutes.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.slot_capacity
	SlotCapacity *int64 `json:"slotCapacity,omitempty"`

	// If false, any query or pipeline job using this reservation will use idle
	//  slots from other reservations within the same admin project. If true, a
	//  query or pipeline job using this reservation will execute with the slot
	//  capacity specified in the slot_capacity field at most.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.ignore_idle_slots
	IgnoreIdleSlots *bool `json:"ignoreIdleSlots,omitempty"`

	// The configuration parameters for the auto scaling feature.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.autoscale
	Autoscale *Reservation_Autoscale `json:"autoscale,omitempty"`

	// Job concurrency target which sets a soft upper bound on the number of jobs
	//  that can run concurrently in this reservation. This is a soft target due to
	//  asynchronous nature of the system and various optimizations for small
	//  queries.
	//  Default value is 0 which means that concurrency target will be
	//  automatically computed by the system.
	//  NOTE: this field is exposed as target job concurrency in the Information
	//  Schema, DDL and BigQuery CLI.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.concurrency
	Concurrency *int64 `json:"concurrency,omitempty"`

	// Applicable only for reservations located within one of the BigQuery
	//  multi-regions (US or EU).
	//
	//  If set to true, this reservation is placed in the organization's
	//  secondary region which is designated for disaster recovery purposes.
	//  If false, this reservation is placed in the organization's default region.
	//
	//  NOTE: this is a preview feature. Project must be allow-listed in order to
	//  set this field.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.multi_region_auxiliary
	MultiRegionAuxiliary *bool `json:"multiRegionAuxiliary,omitempty"`

	// Edition of the reservation.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.edition
	Edition *string `json:"edition,omitempty"`

	// Optional. The current location of the reservation's secondary replica. This
	//  field is only set for reservations using the managed disaster recovery
	//  feature. Users can set this in create reservation calls
	//  to create a failover reservation or in update reservation calls to convert
	//  a non-failover reservation to a failover reservation(or vice versa).
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.secondary_location
	SecondaryLocation *string `json:"secondaryLocation,omitempty"`

	// Optional. The overall max slots for the reservation, covering slot_capacity
	//  (baseline), idle slots (if ignore_idle_slots is false) and scaled slots.
	//  If present, the reservation won't use more than the specified number of
	//  slots, even if there is demand and supply (from idle slots).
	//  NOTE: capping a reservation's idle slot usage is best effort and its
	//  usage may exceed the max_slots value. However, in terms of
	//  autoscale.current_slots (which accounts for the additional added slots), it
	//  will never exceed the max_slots - baseline.
	//
	//
	//  This field must be set together with the scaling_mode enum value.
	//
	//  If the max_slots and scaling_mode are set, the autoscale or
	//  autoscale.max_slots field must be unset. However, the
	//  autoscale field may still be in the output. The autopscale.max_slots will
	//  always show as 0 and the autoscaler.current_slots will represent the
	//  current slots from autoscaler excluding idle slots.
	//  For example, if the max_slots is 1000 and scaling_mode is AUTOSCALE_ONLY,
	//  then in the output, the autoscaler.max_slots will be 0 and the
	//  autoscaler.current_slots may be any value between 0 and 1000.
	//
	//  If the max_slots is 1000, scaling_mode is ALL_SLOTS, the baseline is 100
	//  and idle slots usage is 200, then in the output, the autoscaler.max_slots
	//  will be 0 and the autoscaler.current_slots will not be higher than 700.
	//
	//  If the max_slots is 1000, scaling_mode is IDLE_SLOTS_ONLY, then in the
	//  output, the autoscaler field will be null.
	//
	//  If the max_slots and scaling_mode are set, then the ignore_idle_slots field
	//  must be aligned with the scaling_mode enum value.(See details in
	//  ScalingMode comments).
	//
	//  Please note,  the max_slots is for user to manage the part of slots greater
	//  than the baseline. Therefore, we don't allow users to set max_slots smaller
	//  or equal to the baseline as it will not be meaningful. If the field is
	//  present and slot_capacity>=max_slots.
	//
	//  Please note that if max_slots is set to 0, we will treat it as unset.
	//  Customers can set max_slots to 0 and set scaling_mode to
	//  SCALING_MODE_UNSPECIFIED to disable the max_slots feature.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.max_slots
	MaxSlots *int64 `json:"maxSlots,omitempty"`

	// Optional. The scaling mode for the reservation.
	//  If the field is present but max_slots is not present.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.scaling_mode
	ScalingMode *string `json:"scalingMode,omitempty"`
}

// BigQueryReservationReservationStatus defines the config connector machine state of BigQueryReservationReservation
type BigQueryReservationReservationStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the BigQueryReservationReservation resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *BigQueryReservationReservationObservedState `json:"observedState,omitempty"`
}

// BigQueryReservationReservationObservedState is the state of the BigQueryReservationReservation resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.bigquery.reservation.v1.Reservation
type BigQueryReservationReservationObservedState struct {
	// The configuration parameters for the auto scaling feature.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.autoscale
	Autoscale *Reservation_AutoscaleObservedState `json:"autoscale,omitempty"`

	// Output only. Creation time of the reservation.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.creation_time
	CreationTime *string `json:"creationTime,omitempty"`

	// Output only. Last update time of the reservation.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The current location of the reservation's primary replica.
	//  This field is only set for reservations using the managed disaster recovery
	//  feature.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.primary_location
	PrimaryLocation *string `json:"primaryLocation,omitempty"`

	// Output only. The location where the reservation was originally created.
	//  This is set only during the failover reservation's creation. All billing
	//  charges for the failover reservation will be applied to this location.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.original_primary_location
	OriginalPrimaryLocation *string `json:"originalPrimaryLocation,omitempty"`

	// Output only. The Disaster Recovery(DR) replication status of the
	//  reservation. This is only available for the primary replicas of DR/failover
	//  reservations and provides information about the both the staleness of the
	//  secondary and the last error encountered while trying to replicate changes
	//  from the primary to the secondary. If this field is blank, it means that
	//  the reservation is either not a DR reservation or the reservation is a DR
	//  secondary or that any replication operations on the reservation have
	//  succeeded.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Reservation.replication_status
	ReplicationStatus *Reservation_ReplicationStatusObservedState `json:"replicationStatus,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpbigqueryreservationreservation;gcpbigqueryreservationreservations
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// BigQueryReservationReservation is the Schema for the BigQueryReservationReservation API
// +k8s:openapi-gen=true
type BigQueryReservationReservation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   BigQueryReservationReservationSpec   `json:"spec,omitempty"`
	Status BigQueryReservationReservationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// BigQueryReservationReservationList contains a list of BigQueryReservationReservation
type BigQueryReservationReservationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BigQueryReservationReservation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BigQueryReservationReservation{}, &BigQueryReservationReservationList{})
}
