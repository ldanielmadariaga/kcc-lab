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

var VMwareEnginePrivateCloudGVK = GroupVersion.WithKind("VMwareEnginePrivateCloud")

// VMwareEnginePrivateCloudSpec defines the desired state of VMwareEnginePrivateCloud
// +kcc:spec:proto=google.cloud.vmwareengine.v1.PrivateCloud
type VMwareEnginePrivateCloudSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/privateClouds/{private_cloud}
	Location *string `json:"location"`

	// The VMwareEnginePrivateCloud name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Required. Network configuration of the private cloud.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.network_config
	// +required
	NetworkConfig *NetworkConfig `json:"networkConfig,omitempty"`

	// Required. Input only. The management cluster for this private cloud.
	//  This field is required during creation of the private cloud to provide
	//  details for the default cluster.
	//
	//  The following fields can't be changed after private cloud creation:
	//  `ManagementCluster.clusterId`, `ManagementCluster.nodeTypeId`.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.management_cluster
	// +required
	ManagementCluster *PrivateCloud_ManagementCluster `json:"managementCluster,omitempty"`

	// User-provided description for this private cloud.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.description
	Description *string `json:"description,omitempty"`

	// Optional. Type of the private cloud. Defaults to STANDARD.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.type
	Type *string `json:"type,omitempty"`
}

// VMwareEnginePrivateCloudStatus defines the config connector machine state of VMwareEnginePrivateCloud
type VMwareEnginePrivateCloudStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the VMwareEnginePrivateCloud resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *VMwareEnginePrivateCloudObservedState `json:"observedState,omitempty"`
}

// VMwareEnginePrivateCloudObservedState is the state of the VMwareEnginePrivateCloud resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.vmwareengine.v1.PrivateCloud
type VMwareEnginePrivateCloudObservedState struct {
	// Output only. Creation time of this resource.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Last update time of this resource.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Time when the resource was scheduled for deletion.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.delete_time
	DeleteTime *string `json:"deleteTime,omitempty"`

	// Output only. Time when the resource will be irreversibly deleted.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.expire_time
	ExpireTime *string `json:"expireTime,omitempty"`

	// Output only. State of the resource. New values may be added to this enum
	//  when appropriate.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.state
	State *string `json:"state,omitempty"`

	// Required. Network configuration of the private cloud.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.network_config
	NetworkConfig *NetworkConfigObservedState `json:"networkConfig,omitempty"`

	// Output only. HCX appliance.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.hcx
	Hcx *HcxObservedState `json:"hcx,omitempty"`

	// Output only. NSX appliance.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.nsx
	Nsx *NsxObservedState `json:"nsx,omitempty"`

	// Output only. Vcenter appliance.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.vcenter
	Vcenter *VcenterObservedState `json:"vcenter,omitempty"`

	// Output only. System-generated unique identifier for the resource.
	// +kcc:proto:field=google.cloud.vmwareengine.v1.PrivateCloud.uid
	Uid *string `json:"uid,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpvmwareengineprivatecloud;gcpvmwareengineprivateclouds
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// VMwareEnginePrivateCloud is the Schema for the VMwareEnginePrivateCloud API
// +k8s:openapi-gen=true
type VMwareEnginePrivateCloud struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   VMwareEnginePrivateCloudSpec   `json:"spec,omitempty"`
	Status VMwareEnginePrivateCloudStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// VMwareEnginePrivateCloudList contains a list of VMwareEnginePrivateCloud
type VMwareEnginePrivateCloudList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VMwareEnginePrivateCloud `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VMwareEnginePrivateCloud{}, &VMwareEnginePrivateCloudList{})
}
