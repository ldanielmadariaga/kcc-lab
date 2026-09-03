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

var AlloyDBInstanceGVK = GroupVersion.WithKind("AlloyDBInstance")

// AlloyDBInstanceSpec defines the desired state of AlloyDBInstance
// +kcc:spec:proto=google.cloud.alloydb.v1beta.Instance
type AlloyDBInstanceSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/clusters/{cluster}/instances/{instance}
	Location *string `json:"location,omitempty"`

	// The Cluster that this resource belongs to.
	// +kcc:guess=parent-ref target=ClusterRef pattern=projects/{project}/locations/{location}/clusters/{cluster}/instances/{instance}
	ClusterRef *ClusterRef `json:"clusterRef,omitempty"`

	// The AlloyDBInstance name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// User-settable and human-readable display name for the Instance.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Labels as key value pairs
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. The type of the instance. Specified at creation time.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.instance_type
	// +required
	InstanceType *string `json:"instanceType,omitempty"`

	// Configurations for the machines that host the underlying
	//  database engine.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.machine_config
	MachineConfig *Instance_MachineConfig `json:"machineConfig,omitempty"`

	// Availability type of an Instance.
	//  If empty, defaults to REGIONAL for primary instances.
	//  For read pools, availability_type is always UNSPECIFIED. Instances in the
	//  read pools are evenly distributed across available zones within the region
	//  (i.e. read pools with more than one node will have a node in at
	//  least two zones).
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.availability_type
	AvailabilityType *string `json:"availabilityType,omitempty"`

	// The Compute Engine zone that the instance should serve from, per
	//  https://cloud.google.com/compute/docs/regions-zones
	//  This can ONLY be specified for ZONAL instances.
	//  If present for a REGIONAL instance, an error will be thrown.
	//  If this is absent for a ZONAL instance, instance is created in a random
	//  zone with available capacity.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.gce_zone
	GCEZone *string `json:"gceZone,omitempty"`

	// Database flags. Set at the instance level.
	//  They are copied from the primary instance on secondary instance creation.
	//  Flags that have restrictions default to the value at primary
	//  instance on read instances during creation. Read instances can set new
	//  flags or override existing flags that are relevant for reads, for example,
	//  for enabling columnar cache on a read instance. Flags set on read instance
	//  might or might not be present on the primary instance.
	//
	//
	//  This is a list of "key": "value" pairs.
	//  "key": The name of the flag. These flags are passed at instance setup time,
	//  so include both server options and system variables for Postgres. Flags are
	//  specified with underscores, not hyphens.
	//  "value": The value of the flag. Booleans are set to **on** for true
	//  and **off** for false. This field must be omitted if the flag
	//  doesn't take a value.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.database_flags
	DatabaseFlags map[string]string `json:"databaseFlags,omitempty"`

	// Configuration for query insights.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.query_insights_config
	QueryInsightsConfig *Instance_QueryInsightsInstanceConfig `json:"queryInsightsConfig,omitempty"`

	// Configuration for observability.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.observability_config
	ObservabilityConfig *Instance_ObservabilityInstanceConfig `json:"observabilityConfig,omitempty"`

	// Read pool instance configuration.
	//  This is required if the value of instanceType is READ_POOL.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.read_pool_config
	ReadPoolConfig *Instance_ReadPoolConfig `json:"readPoolConfig,omitempty"`

	// For Resource freshness validation (https://google.aip.dev/154)
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.etag
	Etag *string `json:"etag,omitempty"`

	// Annotations to allow client tools to store small amount of arbitrary data.
	//  This is distinct from labels.
	//  https://google.aip.dev/128
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.annotations
	Annotations map[string]string `json:"annotations,omitempty"`

	// Update policy that will be applied during instance update.
	//  This field is not persisted when you update the instance.
	//  To use a non-default update policy, you must
	//  specify explicitly specify the value in each update request.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.update_policy
	UpdatePolicy *Instance_UpdatePolicy `json:"updatePolicy,omitempty"`

	// Optional. Client connection specific configurations
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.client_connection_config
	ClientConnectionConfig *Instance_ClientConnectionConfig `json:"clientConnectionConfig,omitempty"`

	// Optional. The configuration for Private Service Connect (PSC) for the
	//  instance.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.psc_instance_config
	PSCInstanceConfig *Instance_PSCInstanceConfig `json:"pscInstanceConfig,omitempty"`

	// Optional. Instance-level network configuration.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.network_config
	NetworkConfig *Instance_InstanceNetworkConfig `json:"networkConfig,omitempty"`

	// Optional. Deprecated and unused. This field will be removed in the near
	//  future.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.gemini_config
	GeminiConfig *GeminiInstanceConfig `json:"geminiConfig,omitempty"`

	// Optional. Specifies whether an instance needs to spin up. Once the instance
	//  is active, the activation policy can be updated to the `NEVER` to stop the
	//  instance. Likewise, the activation policy can be updated to `ALWAYS` to
	//  start the instance.
	//  There are restrictions around when an instance can/cannot be activated (for
	//  example, a read pool instance should be stopped before stopping primary
	//  etc.). Please refer to the API documentation for more details.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.activation_policy
	ActivationPolicy *string `json:"activationPolicy,omitempty"`

	// Optional. The configuration for Managed Connection Pool (MCP).
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.connection_pool_config
	ConnectionPoolConfig *Instance_ConnectionPoolConfig `json:"connectionPoolConfig,omitempty"`
}

// AlloyDBInstanceStatus defines the config connector machine state of AlloyDBInstance
type AlloyDBInstanceStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the AlloyDBInstance resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *AlloyDBInstanceObservedState `json:"observedState,omitempty"`
}

// AlloyDBInstanceObservedState is the state of the AlloyDBInstance resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.alloydb.v1beta.Instance
type AlloyDBInstanceObservedState struct {
	// Output only. The system-generated UID of the resource. The UID is assigned
	//  when the resource is created, and it is retained until it is deleted.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. Create time stamp
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Update time stamp
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. Delete time stamp
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.delete_time
	DeleteTime *string `json:"deleteTime,omitempty"`

	// Output only. The current serving state of the instance.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.state
	State *string `json:"state,omitempty"`

	// Output only. This is set for the read-write VM of the PRIMARY instance
	//  only.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.writable_node
	WritableNode *Instance_NodeObservedState `json:"writableNode,omitempty"`

	// Output only. List of available read-only VMs in this instance, including
	//  the standby for a PRIMARY instance.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.nodes
	Nodes []Instance_NodeObservedState `json:"nodes,omitempty"`

	// Configuration for observability.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.observability_config
	ObservabilityConfig *Instance_ObservabilityInstanceConfigObservedState `json:"observabilityConfig,omitempty"`

	// Output only. The IP address for the Instance.
	//  This is the connection endpoint for an end-user application.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.ip_address
	IPAddress *string `json:"ipAddress,omitempty"`

	// Output only. The public IP addresses for the Instance. This is available
	//  ONLY when enable_public_ip is set. This is the connection endpoint for an
	//  end-user application.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.public_ip_address
	PublicIPAddress *string `json:"publicIPAddress,omitempty"`

	// Output only. Reconciling (https://google.aip.dev/128#reconciliation).
	//  Set to true if the current state of Instance does not match the user's
	//  intended state, and the service is actively updating the resource to
	//  reconcile them. This can happen due to user-triggered updates or
	//  system actions like failover or maintenance.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.reconciling
	Reconciling *bool `json:"reconciling,omitempty"`

	// Output only. Reserved for future use.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.satisfies_pzs
	SatisfiesPzs *bool `json:"satisfiesPzs,omitempty"`

	// Optional. The configuration for Private Service Connect (PSC) for the
	//  instance.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.psc_instance_config
	PSCInstanceConfig *Instance_PSCInstanceConfigObservedState `json:"pscInstanceConfig,omitempty"`

	// Optional. Instance-level network configuration.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.network_config
	NetworkConfig *Instance_InstanceNetworkConfigObservedState `json:"networkConfig,omitempty"`

	// Optional. Deprecated and unused. This field will be removed in the near
	//  future.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.gemini_config
	GeminiConfig *GeminiInstanceConfigObservedState `json:"geminiConfig,omitempty"`

	// Output only. All outbound public IP addresses configured for the instance.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.outbound_public_ip_addresses
	OutboundPublicIPAddresses []string `json:"outboundPublicIPAddresses,omitempty"`

	// Output only. Configuration parameters related to Gemini Cloud Assist.
	// +kcc:proto:field=google.cloud.alloydb.v1beta.Instance.gca_config
	GcaConfig *GcaInstanceConfigObservedState `json:"gcaConfig,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpalloydbinstance;gcpalloydbinstances
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// AlloyDBInstance is the Schema for the AlloyDBInstance API
// +k8s:openapi-gen=true
type AlloyDBInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   AlloyDBInstanceSpec   `json:"spec,omitempty"`
	Status AlloyDBInstanceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// AlloyDBInstanceList contains a list of AlloyDBInstance
type AlloyDBInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlloyDBInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlloyDBInstance{}, &AlloyDBInstanceList{})
}
