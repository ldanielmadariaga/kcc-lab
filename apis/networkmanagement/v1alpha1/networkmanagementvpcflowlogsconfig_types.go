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

var NetworkManagementVPCFlowLogsConfigGVK = GroupVersion.WithKind("NetworkManagementVPCFlowLogsConfig")

// NetworkManagementVPCFlowLogsConfigSpec defines the desired state of NetworkManagementVPCFlowLogsConfig
// +kcc:spec:proto=google.cloud.networkmanagement.v1.VpcFlowLogsConfig
type NetworkManagementVPCFlowLogsConfigSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/vpcFlowLogsConfigs/{vpc_flow_logs_config}
	Location *string `json:"location"`

	// The NetworkManagementVPCFlowLogsConfig name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. The user-supplied description of the VPC Flow Logs configuration.
	//  Maximum of 512 characters.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.description
	Description *string `json:"description,omitempty"`

	// Optional. The state of the VPC Flow Log configuration. Default value is
	//  ENABLED. When creating a new configuration, it must be enabled.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.state
	State *string `json:"state,omitempty"`

	// Optional. The aggregation interval for the logs. Default value is
	//  INTERVAL_5_SEC.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.aggregation_interval
	AggregationInterval *string `json:"aggregationInterval,omitempty"`

	// Optional. The value of the field must be in (0, 1]. The sampling rate of
	//  VPC Flow Logs where 1.0 means all collected logs are reported. Setting the
	//  sampling rate to 0.0 is not allowed. If you want to disable VPC Flow Logs,
	//  use the state field instead. Default value is 1.0.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.flow_sampling
	FlowSampling *float32 `json:"flowSampling,omitempty"`

	// Optional. Configures whether all, none or a subset of metadata fields
	//  should be added to the reported VPC flow logs. Default value is
	//  INCLUDE_ALL_METADATA.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.metadata
	Metadata *string `json:"metadata,omitempty"`

	// Optional. Custom metadata fields to include in the reported VPC flow logs.
	//  Can only be specified if "metadata" was set to CUSTOM_METADATA.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.metadata_fields
	MetadataFields []string `json:"metadataFields,omitempty"`

	// Optional. Export filter used to define which VPC Flow Logs should be
	//  logged.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.filter_expr
	FilterExpr *string `json:"filterExpr,omitempty"`

	// Traffic will be logged from the Interconnect Attachment.
	//  Format:
	//  projects/{project_id}/regions/{region}/interconnectAttachments/{name}
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.interconnect_attachment
	InterconnectAttachment *string `json:"interconnectAttachment,omitempty"`

	// Traffic will be logged from the VPN Tunnel.
	//  Format: projects/{project_id}/regions/{region}/vpnTunnels/{name}
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.vpn_tunnel
	VPNTunnel *string `json:"vpnTunnel,omitempty"`

	// Optional. Resource labels to represent user-provided metadata.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.labels
	Labels map[string]string `json:"labels,omitempty"`
}

// NetworkManagementVPCFlowLogsConfigStatus defines the config connector machine state of NetworkManagementVPCFlowLogsConfig
type NetworkManagementVPCFlowLogsConfigStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the NetworkManagementVPCFlowLogsConfig resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *NetworkManagementVPCFlowLogsConfigObservedState `json:"observedState,omitempty"`
}

// NetworkManagementVPCFlowLogsConfigObservedState is the state of the NetworkManagementVPCFlowLogsConfig resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.networkmanagement.v1.VpcFlowLogsConfig
type NetworkManagementVPCFlowLogsConfigObservedState struct {
	// Output only. A diagnostic bit - describes the state of the configured
	//  target resource for diagnostic purposes.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.target_resource_state
	TargetResourceState *string `json:"targetResourceState,omitempty"`

	// Output only. The time the config was created.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The time the config was updated.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.VpcFlowLogsConfig.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpnetworkmanagementvpcflowlogsconfig;gcpnetworkmanagementvpcflowlogsconfigs
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// NetworkManagementVPCFlowLogsConfig is the Schema for the NetworkManagementVPCFlowLogsConfig API
// +k8s:openapi-gen=true
type NetworkManagementVPCFlowLogsConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   NetworkManagementVPCFlowLogsConfigSpec   `json:"spec,omitempty"`
	Status NetworkManagementVPCFlowLogsConfigStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkManagementVPCFlowLogsConfigList contains a list of NetworkManagementVPCFlowLogsConfig
type NetworkManagementVPCFlowLogsConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkManagementVPCFlowLogsConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkManagementVPCFlowLogsConfig{}, &NetworkManagementVPCFlowLogsConfigList{})
}
