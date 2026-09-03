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
	common "github.com/GoogleCloudPlatform/k8s-config-connector/apis/common"
	refsv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/apis/refs/v1beta1"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var WorkstationConfigGVK = GroupVersion.WithKind("WorkstationConfig")

// WorkstationConfigSpec defines the desired state of WorkstationConfig
// +kcc:spec:proto=google.cloud.workstations.v1.WorkstationConfig
type WorkstationConfigSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/workstationClusters/{workstation_cluster}/workstationConfigs/{workstation_config}
	Location *string `json:"location,omitempty"`

	// The WorkstationCluster that this resource belongs to.
	// +kcc:guess=parent-ref target=WorkstationClusterRef pattern=projects/{project}/locations/{location}/workstationClusters/{workstation_cluster}/workstationConfigs/{workstation_config}
	WorkstationClusterRef *WorkstationClusterRef `json:"workstationClusterRef,omitempty"`

	// The WorkstationConfig name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Human-readable name for this workstation configuration.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. Client-specified annotations.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Optional.
	//  [Labels](https://cloud.google.com/workstations/docs/label-resources) that
	//  are applied to the workstation configuration and that are also propagated
	//  to the underlying Compute Engine resources.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. Checksum computed by the server. May be sent on update and delete
	//  requests to make sure that the client has an up-to-date value before
	//  proceeding.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.etag
	Etag *string `json:"etag,omitempty"`

	// Optional. Number of seconds to wait before automatically stopping a
	//  workstation after it last received user traffic.
	//
	//  A value of `"0s"` indicates that Cloud Workstations VMs created with this
	//  configuration should never time out due to idleness.
	//  Provide
	//  [duration](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf#duration)
	//  terminated by `s` for seconds—for example, `"7200s"` (2 hours).
	//  The default is `"1200s"` (20 minutes).
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.idle_timeout
	IdleTimeout *string `json:"idleTimeout,omitempty"`

	// Optional. Number of seconds that a workstation can run until it is
	//  automatically shut down. We recommend that workstations be shut down daily
	//  to reduce costs and so that security updates can be applied upon restart.
	//  The
	//  [idle_timeout][google.cloud.workstations.v1.WorkstationConfig.idle_timeout]
	//  and
	//  [running_timeout][google.cloud.workstations.v1.WorkstationConfig.running_timeout]
	//  fields are independent of each other. Note that the
	//  [running_timeout][google.cloud.workstations.v1.WorkstationConfig.running_timeout]
	//  field shuts down VMs after the specified time, regardless of whether or not
	//  the VMs are idle.
	//
	//  Provide duration terminated by `s` for seconds—for example, `"54000s"`
	//  (15 hours). Defaults to `"43200s"` (12 hours). A value of `"0s"` indicates
	//  that workstations using this configuration should never time out. If
	//  [encryption_key][google.cloud.workstations.v1.WorkstationConfig.encryption_key]
	//  is set, it must be greater than `"0s"` and less than
	//  `"86400s"` (24 hours).
	//
	//  Warning: A value of `"0s"` indicates that Cloud Workstations VMs created
	//  with this configuration have no maximum running time. This is strongly
	//  discouraged because you incur costs and will not pick up security updates.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.running_timeout
	RunningTimeout *string `json:"runningTimeout,omitempty"`

	// Optional. Runtime host for the workstation.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.host
	Host *WorkstationConfig_Host `json:"host,omitempty"`

	// Optional. Directories to persist across workstation sessions.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.persistent_directories
	PersistentDirectories []WorkstationConfig_PersistentDirectory `json:"persistentDirectories,omitempty"`

	// Optional. Container that runs upon startup for each workstation using this
	//  workstation configuration.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.container
	Container *WorkstationConfig_Container `json:"container,omitempty"`

	// Immutable. Encrypts resources of this workstation configuration using a
	//  customer-managed encryption key (CMEK).
	//
	//  If specified, the boot disk of the Compute Engine instance and the
	//  persistent disk are encrypted using this encryption key. If
	//  this field is not set, the disks are encrypted using a generated
	//  key. Customer-managed encryption keys do not protect disk metadata.
	//
	//  If the customer-managed encryption key is rotated, when the workstation
	//  instance is stopped, the system attempts to recreate the
	//  persistent disk with the new version of the key. Be sure to keep
	//  older versions of the key until the persistent disk is recreated.
	//  Otherwise, data on the persistent disk might be lost.
	//
	//  If the encryption key is revoked, the workstation session automatically
	//  stops within 7 hours.
	//
	//  Immutable after the workstation configuration is created.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.encryption_key
	EncryptionKey *WorkstationConfig_CustomerEncryptionKey `json:"encryptionKey,omitempty"`

	// Optional. Readiness checks to perform when starting a workstation using
	//  this workstation configuration. Mark a workstation as running only after
	//  all specified readiness checks return 200 status codes.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.readiness_checks
	ReadinessChecks []WorkstationConfig_ReadinessCheck `json:"readinessChecks,omitempty"`

	// Optional. Immutable. Specifies the zones used to replicate the VM and disk
	//  resources within the region. If set, exactly two zones within the
	//  workstation cluster's region must be specified—for example,
	//  `['us-central1-a', 'us-central1-f']`. If this field is empty, two default
	//  zones within the region are used.
	//
	//  Immutable after the workstation configuration is created.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.replica_zones
	ReplicaZones []string `json:"replicaZones,omitempty"`
}

// WorkstationConfigStatus defines the config connector machine state of WorkstationConfig
type WorkstationConfigStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the WorkstationConfig resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *WorkstationConfigObservedState `json:"observedState,omitempty"`
}

// WorkstationConfigObservedState is the state of the WorkstationConfig resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.workstations.v1.WorkstationConfig
type WorkstationConfigObservedState struct {
	// Output only. A system-assigned unique identifier for this workstation
	//  configuration.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. Indicates whether this workstation configuration is currently
	//  being updated to match its intended state.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.reconciling
	Reconciling *bool `json:"reconciling,omitempty"`

	// Output only. Time when this workstation configuration was created.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Time when this workstation configuration was most recently
	//  updated.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Time when this workstation configuration was soft-deleted.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.delete_time
	DeleteTime *string `json:"deleteTime,omitempty"`

	// Optional. Runtime host for the workstation.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.host
	Host *WorkstationConfig_HostObservedState `json:"host,omitempty"`

	// Output only. Whether this resource is degraded, in which case it may
	//  require user action to restore full functionality. See also the
	//  [conditions][google.cloud.workstations.v1.WorkstationConfig.conditions]
	//  field.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.degraded
	Degraded *bool `json:"degraded,omitempty"`

	// Output only. Status conditions describing the current resource state.
	// +kcc:proto:field=google.cloud.workstations.v1.WorkstationConfig.conditions
	Conditions []common.Status `json:"conditions,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpworkstationconfig;gcpworkstationconfigs
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// WorkstationConfig is the Schema for the WorkstationConfig API
// +k8s:openapi-gen=true
type WorkstationConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   WorkstationConfigSpec   `json:"spec,omitempty"`
	Status WorkstationConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// WorkstationConfigList contains a list of WorkstationConfig
type WorkstationConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkstationConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkstationConfig{}, &WorkstationConfigList{})
}
