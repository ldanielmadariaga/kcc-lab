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

var ManagedKafkaConnectClusterGVK = GroupVersion.WithKind("ManagedKafkaConnectCluster")

// ManagedKafkaConnectClusterSpec defines the desired state of ManagedKafkaConnectCluster
// +kcc:spec:proto=google.cloud.managedkafka.v1.ConnectCluster
type ManagedKafkaConnectClusterSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The ManagedKafkaConnectCluster name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Required. Configuration properties for a Kafka Connect cluster deployed
	//  to Google Cloud Platform.
	// +kcc:proto:field=google.cloud.managedkafka.v1.ConnectCluster.gcp_config
	// +required
	GcpConfig *ConnectGcpConfig `json:"gcpConfig,omitempty"`

	// Required. Immutable. The name of the Kafka cluster this Kafka Connect
	//  cluster is attached to. Structured like:
	//  projects/{project}/locations/{location}/clusters/{cluster}
	// +kcc:proto:field=google.cloud.managedkafka.v1.ConnectCluster.kafka_cluster
	// +required
	KafkaCluster *string `json:"kafkaCluster,omitempty"`

	// Optional. Labels as key value pairs.
	// +kcc:proto:field=google.cloud.managedkafka.v1.ConnectCluster.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. Capacity configuration for the Kafka Connect cluster.
	// +kcc:proto:field=google.cloud.managedkafka.v1.ConnectCluster.capacity_config
	// +required
	CapacityConfig *CapacityConfig `json:"capacityConfig,omitempty"`

	// Optional. Configurations for the worker that are overridden from the
	//  defaults. The key of the map is a Kafka Connect worker property name, for
	//  example: `exactly.once.source.support`.
	// +kcc:proto:field=google.cloud.managedkafka.v1.ConnectCluster.config
	Config map[string]string `json:"config,omitempty"`
}

// ManagedKafkaConnectClusterStatus defines the config connector machine state of ManagedKafkaConnectCluster
type ManagedKafkaConnectClusterStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the ManagedKafkaConnectCluster resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *ManagedKafkaConnectClusterObservedState `json:"observedState,omitempty"`
}

// ManagedKafkaConnectClusterObservedState is the state of the ManagedKafkaConnectCluster resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.managedkafka.v1.ConnectCluster
type ManagedKafkaConnectClusterObservedState struct {
	// Output only. The time when the cluster was created.
	// +kcc:proto:field=google.cloud.managedkafka.v1.ConnectCluster.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The time when the cluster was last updated.
	// +kcc:proto:field=google.cloud.managedkafka.v1.ConnectCluster.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The current state of the cluster.
	// +kcc:proto:field=google.cloud.managedkafka.v1.ConnectCluster.state
	State *string `json:"state,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpmanagedkafkaconnectcluster;gcpmanagedkafkaconnectclusters
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// ManagedKafkaConnectCluster is the Schema for the ManagedKafkaConnectCluster API
// +k8s:openapi-gen=true
type ManagedKafkaConnectCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ManagedKafkaConnectClusterSpec   `json:"spec,omitempty"`
	Status ManagedKafkaConnectClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ManagedKafkaConnectClusterList contains a list of ManagedKafkaConnectCluster
type ManagedKafkaConnectClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedKafkaConnectCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedKafkaConnectCluster{}, &ManagedKafkaConnectClusterList{})
}
