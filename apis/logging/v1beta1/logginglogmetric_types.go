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

var LoggingLogMetricGVK = GroupVersion.WithKind("LoggingLogMetric")

// LoggingLogMetricSpec defines the desired state of LoggingLogMetric
// +kcc:spec:proto=google.logging.v2.LogMetric
type LoggingLogMetricSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The LoggingLogMetric name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. A description of this metric, which is used in documentation.
	//  The maximum length of the description is 8000 characters.
	// +kcc:proto:field=google.logging.v2.LogMetric.description
	Description *string `json:"description,omitempty"`

	// Required. An [advanced logs
	//  filter](https://cloud.google.com/logging/docs/view/advanced_filters) which
	//  is used to match log entries. Example:
	//
	//      "resource.type=gae_app AND severity>=ERROR"
	//
	//  The maximum length of the filter is 20000 characters.
	// +kcc:proto:field=google.logging.v2.LogMetric.filter
	// +required
	Filter *string `json:"filter,omitempty"`

	// Optional. The resource name of the Log Bucket that owns the Log Metric.
	//  Only Log Buckets in projects are supported. The bucket has to be in the
	//  same project as the metric.
	//
	//  For example:
	//
	//    `projects/my-project/locations/global/buckets/my-bucket`
	//
	//  If empty, then the Log Metric is considered a non-Bucket Log Metric.
	// +kcc:proto:field=google.logging.v2.LogMetric.bucket_name
	BucketName *string `json:"bucketName,omitempty"`

	// Optional. If set to True, then this metric is disabled and it does not
	//  generate any points.
	// +kcc:proto:field=google.logging.v2.LogMetric.disabled
	Disabled *bool `json:"disabled,omitempty"`

	// Optional. The metric descriptor associated with the logs-based metric.
	//  If unspecified, it uses a default metric descriptor with a DELTA metric
	//  kind, INT64 value type, with no labels and a unit of "1". Such a metric
	//  counts the number of log entries matching the `filter` expression.
	//
	//  The `name`, `type`, and `description` fields in the `metric_descriptor`
	//  are output only, and is constructed using the `name` and `description`
	//  field in the LogMetric.
	//
	//  To create a logs-based metric that records a distribution of log values, a
	//  DELTA metric kind with a DISTRIBUTION value type must be used along with
	//  a `value_extractor` expression in the LogMetric.
	//
	//  Each label in the metric descriptor must have a matching label
	//  name as the key and an extractor expression as the value in the
	//  `label_extractors` map.
	//
	//  The `metric_kind` and `value_type` fields in the `metric_descriptor` cannot
	//  be updated once initially configured. New labels can be added in the
	//  `metric_descriptor`, but existing labels cannot be modified except for
	//  their description.
	// +kcc:proto:field=google.logging.v2.LogMetric.metric_descriptor
	MetricDescriptor *MetricDescriptor `json:"metricDescriptor,omitempty"`

	// Optional. A `value_extractor` is required when using a distribution
	//  logs-based metric to extract the values to record from a log entry.
	//  Two functions are supported for value extraction: `EXTRACT(field)` or
	//  `REGEXP_EXTRACT(field, regex)`. The arguments are:
	//
	//    1. field: The name of the log entry field from which the value is to be
	//       extracted.
	//    2. regex: A regular expression using the Google RE2 syntax
	//       (https://github.com/google/re2/wiki/Syntax) with a single capture
	//       group to extract data from the specified log entry field. The value
	//       of the field is converted to a string before applying the regex.
	//       It is an error to specify a regex that does not include exactly one
	//       capture group.
	//
	//  The result of the extraction must be convertible to a double type, as the
	//  distribution always records double values. If either the extraction or
	//  the conversion to double fails, then those values are not recorded in the
	//  distribution.
	//
	//  Example: `REGEXP_EXTRACT(jsonPayload.request, ".*quantity=(\d+).*")`
	// +kcc:proto:field=google.logging.v2.LogMetric.value_extractor
	ValueExtractor *string `json:"valueExtractor,omitempty"`

	// Optional. A map from a label key string to an extractor expression which is
	//  used to extract data from a log entry field and assign as the label value.
	//  Each label key specified in the LabelDescriptor must have an associated
	//  extractor expression in this map. The syntax of the extractor expression
	//  is the same as for the `value_extractor` field.
	//
	//  The extracted value is converted to the type defined in the label
	//  descriptor. If either the extraction or the type conversion fails,
	//  the label will have a default value. The default value for a string
	//  label is an empty string, for an integer label its 0, and for a boolean
	//  label its `false`.
	//
	//  Note that there are upper bounds on the maximum number of labels and the
	//  number of active time series that are allowed in a project.
	// +kcc:proto:field=google.logging.v2.LogMetric.label_extractors
	LabelExtractors map[string]string `json:"labelExtractors,omitempty"`

	// Optional. The `bucket_options` are required when the logs-based metric is
	//  using a DISTRIBUTION value type and it describes the bucket boundaries
	//  used to create a histogram of the extracted values.
	// +kcc:proto:field=google.logging.v2.LogMetric.bucket_options
	BucketOptions *Distribution_BucketOptions `json:"bucketOptions,omitempty"`

	// Deprecated. The API version that created or updated this metric.
	//  The v2 format is used by default and cannot be changed.
	// +kcc:proto:field=google.logging.v2.LogMetric.version
	Version *string `json:"version,omitempty"`
}

// LoggingLogMetricStatus defines the config connector machine state of LoggingLogMetric
type LoggingLogMetricStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the LoggingLogMetric resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *LoggingLogMetricObservedState `json:"observedState,omitempty"`
}

// LoggingLogMetricObservedState is the state of the LoggingLogMetric resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.logging.v2.LogMetric
type LoggingLogMetricObservedState struct {
	// Output only. The creation timestamp of the metric.
	//
	//  This field may not be present for older metrics.
	// +kcc:proto:field=google.logging.v2.LogMetric.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The last update timestamp of the metric.
	//
	//  This field may not be present for older metrics.
	// +kcc:proto:field=google.logging.v2.LogMetric.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcplogginglogmetric;gcplogginglogmetrics
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// LoggingLogMetric is the Schema for the LoggingLogMetric API
// +k8s:openapi-gen=true
type LoggingLogMetric struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   LoggingLogMetricSpec   `json:"spec,omitempty"`
	Status LoggingLogMetricStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// LoggingLogMetricList contains a list of LoggingLogMetric
type LoggingLogMetricList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LoggingLogMetric `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LoggingLogMetric{}, &LoggingLogMetricList{})
}
