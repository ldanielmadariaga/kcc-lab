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

var MonitoringDashboardGVK = GroupVersion.WithKind("MonitoringDashboard")

// MonitoringDashboardSpec defines the desired state of MonitoringDashboard
// +kcc:spec:proto=google.monitoring.v3.Dashboard
type MonitoringDashboardSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The MonitoringDashboard name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Required. The mutable, human-readable name.
	// +kcc:proto:field=google.monitoring.dashboard.v1.Dashboard.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// `etag` is used for optimistic concurrency control as a way to help
	//  prevent simultaneous updates of a policy from overwriting each other.
	//  An `etag` is returned in the response to `GetDashboard`, and
	//  users are expected to put that etag in the request to `UpdateDashboard` to
	//  ensure that their change will be applied to the same version of the
	//  Dashboard configuration. The field should not be passed during
	//  dashboard creation.
	// +kcc:proto:field=google.monitoring.dashboard.v1.Dashboard.etag
	Etag *string `json:"etag,omitempty"`

	// Content is arranged with a basic layout that re-flows a simple list of
	//  informational elements like widgets or tiles.
	// +kcc:proto:field=google.monitoring.dashboard.v1.Dashboard.grid_layout
	GridLayout *GridLayout `json:"gridLayout,omitempty"`

	// The content is arranged as a grid of tiles, with each content widget
	//  occupying one or more grid blocks.
	// +kcc:proto:field=google.monitoring.dashboard.v1.Dashboard.mosaic_layout
	MosaicLayout *MosaicLayout `json:"mosaicLayout,omitempty"`

	// The content is divided into equally spaced rows and the widgets are
	//  arranged horizontally.
	// +kcc:proto:field=google.monitoring.dashboard.v1.Dashboard.row_layout
	RowLayout *RowLayout `json:"rowLayout,omitempty"`

	// The content is divided into equally spaced columns and the widgets are
	//  arranged vertically.
	// +kcc:proto:field=google.monitoring.dashboard.v1.Dashboard.column_layout
	ColumnLayout *ColumnLayout `json:"columnLayout,omitempty"`

	// Filters to reduce the amount of data charted based on the filter criteria.
	// +kcc:proto:field=google.monitoring.dashboard.v1.Dashboard.dashboard_filters
	DashboardFilters []DashboardFilter `json:"dashboardFilters,omitempty"`

	// Labels applied to the dashboard
	// +kcc:proto:field=google.monitoring.dashboard.v1.Dashboard.labels
	Labels map[string]string `json:"labels,omitempty"`
}

// MonitoringDashboardStatus defines the config connector machine state of MonitoringDashboard
type MonitoringDashboardStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the MonitoringDashboard resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *MonitoringDashboardObservedState `json:"observedState,omitempty"`
}

// MonitoringDashboardObservedState is the state of the MonitoringDashboard resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.monitoring.v3.Dashboard
type MonitoringDashboardObservedState struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpmonitoringdashboard;gcpmonitoringdashboards
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// MonitoringDashboard is the Schema for the MonitoringDashboard API
// +k8s:openapi-gen=true
type MonitoringDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   MonitoringDashboardSpec   `json:"spec,omitempty"`
	Status MonitoringDashboardStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// MonitoringDashboardList contains a list of MonitoringDashboard
type MonitoringDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MonitoringDashboard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MonitoringDashboard{}, &MonitoringDashboardList{})
}
