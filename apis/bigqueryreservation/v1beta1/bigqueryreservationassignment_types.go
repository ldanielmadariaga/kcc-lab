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

var BigQueryReservationAssignmentGVK = GroupVersion.WithKind("BigQueryReservationAssignment")

// BigQueryReservationAssignmentSpec defines the desired state of BigQueryReservationAssignment
// +kcc:spec:proto=google.cloud.bigquery.reservation.v1.Assignment
type BigQueryReservationAssignmentSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/reservations/{reservation}/assignments/{assignment}
	Location *string `json:"location,omitempty"`

	// The Reservation that this resource belongs to.
	// +kcc:guess=parent-ref target=ReservationRef pattern=projects/{project}/locations/{location}/reservations/{reservation}/assignments/{assignment}
	ReservationRef *ReservationRef `json:"reservationRef,omitempty"`

	// The BigQueryReservationAssignment name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The resource which will use the reservation. E.g.
	//  `projects/myproject`, `folders/123`, or `organizations/456`.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Assignment.assignee
	Assignee *string `json:"assignee,omitempty"`

	// Which type of jobs will use the reservation.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Assignment.job_type
	JobType *string `json:"jobType,omitempty"`

	// Optional. This field controls if "Gemini in BigQuery"
	//  (https://cloud.google.com/gemini/docs/bigquery/overview) features should be
	//  enabled for this reservation assignment, which is not on by default.
	//  "Gemini in BigQuery" has a distinct compliance posture from BigQuery.  If
	//  this field is set to true, the assignment job type is QUERY, and
	//  the parent reservation edition is ENTERPRISE_PLUS, then the assignment will
	//  give the grantee project/organization access to "Gemini in BigQuery"
	//  features.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Assignment.enable_gemini_in_bigquery
	EnableGeminiInBigquery *bool `json:"enableGeminiInBigquery,omitempty"`
}

// BigQueryReservationAssignmentStatus defines the config connector machine state of BigQueryReservationAssignment
type BigQueryReservationAssignmentStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the BigQueryReservationAssignment resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *BigQueryReservationAssignmentObservedState `json:"observedState,omitempty"`
}

// BigQueryReservationAssignmentObservedState is the state of the BigQueryReservationAssignment resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.bigquery.reservation.v1.Assignment
type BigQueryReservationAssignmentObservedState struct {
	// Output only. State of the assignment.
	// +kcc:proto:field=google.cloud.bigquery.reservation.v1.Assignment.state
	State *string `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpbigqueryreservationassignment;gcpbigqueryreservationassignments
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// BigQueryReservationAssignment is the Schema for the BigQueryReservationAssignment API
// +k8s:openapi-gen=true
type BigQueryReservationAssignment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   BigQueryReservationAssignmentSpec   `json:"spec,omitempty"`
	Status BigQueryReservationAssignmentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// BigQueryReservationAssignmentList contains a list of BigQueryReservationAssignment
type BigQueryReservationAssignmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BigQueryReservationAssignment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BigQueryReservationAssignment{}, &BigQueryReservationAssignmentList{})
}
