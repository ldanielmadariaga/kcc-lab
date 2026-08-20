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

var AppOptimizeReportGVK = GroupVersion.WithKind("AppOptimizeReport")

// AppOptimizeReportSpec defines the desired state of AppOptimizeReport
// +kcc:spec:proto=google.cloud.appoptimize.v1beta.Report
type AppOptimizeReportSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The AppOptimizeReport name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Required. A list of dimensions to include in the report. Supported values:
	//
	//    * `project`
	//    * `application`
	//    * `service_or_workload`
	//    * `resource`
	//    * `resource_type`
	//    * `location`
	//    * `product_display_name`
	//    * `sku`
	//    * `month`
	//    * `day`
	//    * `hour`
	//
	//  To aggregate results by time, specify at least one time dimension
	//  (`month`, `day`, or `hour`). All time dimensions use Pacific Time,
	//  respect Daylight Saving Time (DST), and follow these ISO 8601 formats:
	//
	//    * `month`: `YYYY-MM` (e.g., `2024-01`)
	//    * `day`: `YYYY-MM-DD` (e.g., `2024-01-10`)
	//    * `hour`: `YYYY-MM-DDTHH` (e.g., `2024-01-10T00`)
	//
	//  If the time range filter does not align with the selected time dimension,
	//  the range is expanded to encompass the full period of the finest-grained
	//  time dimension.
	//
	//  For example, if the filter is `2026-01-10` through `2026-01-12` and the
	//  `month` dimension is selected, the effective time range expands to include
	//  all of January (`2026-01-01` to `2026-02-01`).
	// +kcc:proto:field=google.cloud.appoptimize.v1beta.Report.dimensions
	// +required
	Dimensions []string `json:"dimensions,omitempty"`

	// Required. A list of metrics to include in the report. Supported values:
	//
	//    * `cost`
	//    * `cpu_mean_utilization`
	//    * `cpu_usage_core_seconds`
	//    * `cpu_allocation_core_seconds`
	//    * `cpu_p95_utilization`
	//    * `memory_mean_utilization`
	//    * `memory_usage_byte_seconds`
	//    * `memory_allocation_byte_seconds`
	//    * `memory_p95_utilization`
	// +kcc:proto:field=google.cloud.appoptimize.v1beta.Report.metrics
	// +required
	Metrics []string `json:"metrics,omitempty"`

	// Optional. The resource containers for which to fetch data. Default is the
	//  project specified in the report's parent.
	// +kcc:proto:field=google.cloud.appoptimize.v1beta.Report.scopes
	Scopes []Scope `json:"scopes,omitempty"`

	// Optional. A Common Expression Language (CEL) expression used to filter the
	//  data for the report.
	//
	//  Predicates may refer to any dimension. Filtering must conform to these
	//  constraints:
	//
	//    * All string field predicates must use exact string matches.
	//    * Multiple predicates referring to the same string field must be joined
	//      using the logical OR operator ('||').
	//    * All other predicates must be joined using the logical AND operator
	//      (`&&`).
	//    * A predicate on a time dimension (e.g., `day`) specifying the start time
	//      must use a greater-than-or-equal-to comparison (`>=`).
	//    * A predicate on a time dimension specifying the end time must use a
	//      less-than comparison (`<`).
	//
	//  Examples:
	//
	//    1. Filter by a specific resource type:
	//       `"resource_type == 'compute.googleapis.com/Instance'"`
	//
	//    2. Filter data points that fall within a specific absolute time interval:
	//       `"hour >= timestamp('2024-01-01T00:00:00Z') &&
	//       hour < timestamp('2024-02-01T00:00:00Z')"`
	//
	//    3. Filter data points that fall within the past 72 hours:
	//       `"hour >= now - duration('72h')"`
	//
	//    4. Combine string predicate with time interval predicate:
	//       `"(location == 'us-east1' || location == 'us-west1') &&
	//        hour >= timestamp('2023-12-01T00:00:00Z') &&
	//        hour < timestamp('2024-02-01T00:00:00Z')"`
	//
	//  If the filter omits time dimensions (`month`, `day`, `hour`), the report
	//  defaults to a 7-day range ending at the previous Pacific Time midnight,
	//  with Daylight Saving Time (DST) applied.
	//
	//  For example, if the current Pacific Time is `2026-01-05T12:00:00`,
	//  the default range is `2025-12-29T00:00:00` to `2026-01-05T00:00:00` Pacific
	//  time.
	// +kcc:proto:field=google.cloud.appoptimize.v1beta.Report.filter
	Filter *string `json:"filter,omitempty"`
}

// AppOptimizeReportStatus defines the config connector machine state of AppOptimizeReport
type AppOptimizeReportStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the AppOptimizeReport resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *AppOptimizeReportObservedState `json:"observedState,omitempty"`
}

// AppOptimizeReportObservedState is the state of the AppOptimizeReport resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.appoptimize.v1beta.Report
type AppOptimizeReportObservedState struct {
	// Output only. Timestamp in UTC of when this report expires. Once the
	//  report expires, it will no longer be accessible and the report's
	//  underlying data will be deleted.
	// +kcc:proto:field=google.cloud.appoptimize.v1beta.Report.expire_time
	ExpireTime *string `json:"expireTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpappoptimizereport;gcpappoptimizereports
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// AppOptimizeReport is the Schema for the AppOptimizeReport API
// +k8s:openapi-gen=true
type AppOptimizeReport struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   AppOptimizeReportSpec   `json:"spec,omitempty"`
	Status AppOptimizeReportStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AppOptimizeReportList contains a list of AppOptimizeReport
type AppOptimizeReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppOptimizeReport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppOptimizeReport{}, &AppOptimizeReportList{})
}
