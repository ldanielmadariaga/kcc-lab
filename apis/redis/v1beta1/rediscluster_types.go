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

var RedisClusterGVK = GroupVersion.WithKind("RedisCluster")

// RedisClusterSpec defines the desired state of RedisCluster
// +kcc:spec:proto=google.cloud.redis.cluster.v1.Cluster
type RedisClusterSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/clusters/{cluster}
	Location *string `json:"location"`

	// The RedisCluster name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Backups stored in Cloud Storage buckets.
	//  The Cloud Storage buckets need to be the same region as the clusters.
	//  Read permission is required to import from the provided Cloud Storage
	//  objects.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.gcs_source
	GCSSource *Cluster_GCSBackupSource `json:"gcsSource,omitempty"`

	// Optional. Backups generated and managed by memorystore service.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.managed_backup_source
	ManagedBackupSource *Cluster_ManagedBackupSource `json:"managedBackupSource,omitempty"`

	// Optional. The number of replica nodes per shard.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.replica_count
	ReplicaCount *int32 `json:"replicaCount,omitempty"`

	// Optional. The authorization mode of the Redis cluster.
	//  If not provided, auth feature is disabled for the cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.authorization_mode
	AuthorizationMode *string `json:"authorizationMode,omitempty"`

	// Optional. The in-transit encryption for the Redis cluster.
	//  If not provided, encryption  is disabled for the cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.transit_encryption_mode
	TransitEncryptionMode *string `json:"transitEncryptionMode,omitempty"`

	// Optional. Number of shards for the Redis cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.shard_count
	ShardCount *int32 `json:"shardCount,omitempty"`

	// Optional. Each PscConfig configures the consumer network where IPs will
	//  be designated to the cluster for client access through Private Service
	//  Connect Automation. Currently, only one PscConfig is supported.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.psc_configs
	PSCConfigs []PSCConfig `json:"pscConfigs,omitempty"`

	// Optional. The type of a redis node in the cluster. NodeType determines the
	//  underlying machine-type of a redis node.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.node_type
	NodeType *string `json:"nodeType,omitempty"`

	// Optional. Persistence config (RDB, AOF) for the cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.persistence_config
	PersistenceConfig *ClusterPersistenceConfig `json:"persistenceConfig,omitempty"`

	// Optional. Key/Value pairs of customer overrides for mutable Redis Configs
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.redis_configs
	RedisConfigs map[string]string `json:"redisConfigs,omitempty"`

	// Optional. This config will be used to determine how the customer wants us
	//  to distribute cluster resources within the region.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.zone_distribution_config
	ZoneDistributionConfig *ZoneDistributionConfig `json:"zoneDistributionConfig,omitempty"`

	// Optional. Cross cluster replication config.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.cross_cluster_replication_config
	CrossClusterReplicationConfig *CrossClusterReplicationConfig `json:"crossClusterReplicationConfig,omitempty"`

	// Optional. The delete operation will fail when the value is set to true.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.deletion_protection_enabled
	DeletionProtectionEnabled *bool `json:"deletionProtectionEnabled,omitempty"`

	// Optional. ClusterMaintenancePolicy determines when to allow or deny
	//  updates.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.maintenance_policy
	MaintenancePolicy *ClusterMaintenancePolicy `json:"maintenancePolicy,omitempty"`

	// Optional. A list of cluster enpoints.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.cluster_endpoints
	ClusterEndpoints []ClusterEndpoint `json:"clusterEndpoints,omitempty"`

	// Optional. The KMS key used to encrypt the at-rest data of the cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.kms_key
	KMSKey *string `json:"kmsKey,omitempty"`

	// Optional. The automated backup config for the cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.automated_backup_config
	AutomatedBackupConfig *AutomatedBackupConfig `json:"automatedBackupConfig,omitempty"`
}

// RedisClusterStatus defines the config connector machine state of RedisCluster
type RedisClusterStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the RedisCluster resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *RedisClusterObservedState `json:"observedState,omitempty"`
}

// RedisClusterObservedState is the state of the RedisCluster resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.redis.cluster.v1.Cluster
type RedisClusterObservedState struct {
	// Output only. The timestamp associated with the cluster creation request.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The current state of this cluster.
	//  Can be CREATING, READY, UPDATING, DELETING and SUSPENDED
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.state
	State *string `json:"state,omitempty"`

	// Output only. System assigned, unique identifier for the cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.uid
	Uid *string `json:"uid,omitempty"`

	// Output only. Redis memory size in GB for the entire cluster rounded up to
	//  the next integer.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.size_gb
	SizeGB *int32 `json:"sizeGB,omitempty"`

	// Output only. Endpoints created on each given network, for Redis clients to
	//  connect to the cluster. Currently only one discovery endpoint is supported.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.discovery_endpoints
	DiscoveryEndpoints []DiscoveryEndpointObservedState `json:"discoveryEndpoints,omitempty"`

	// Output only. The list of PSC connections that are auto-created through
	//  service connectivity automation.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.psc_connections
	PSCConnections []PSCConnectionObservedState `json:"pscConnections,omitempty"`

	// Output only. Additional information about the current state of the cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.state_info
	StateInfo *Cluster_StateInfo `json:"stateInfo,omitempty"`

	// Output only. Precise value of redis memory size in GB for the entire
	//  cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.precise_size_gb
	PreciseSizeGB *float64 `json:"preciseSizeGB,omitempty"`

	// Optional. Cross cluster replication config.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.cross_cluster_replication_config
	CrossClusterReplicationConfig *CrossClusterReplicationConfigObservedState `json:"crossClusterReplicationConfig,omitempty"`

	// Optional. ClusterMaintenancePolicy determines when to allow or deny
	//  updates.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.maintenance_policy
	MaintenancePolicy *ClusterMaintenancePolicyObservedState `json:"maintenancePolicy,omitempty"`

	// Output only. ClusterMaintenanceSchedule Output only Published maintenance
	//  schedule.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.maintenance_schedule
	MaintenanceSchedule *ClusterMaintenanceScheduleObservedState `json:"maintenanceSchedule,omitempty"`

	// Output only. Service attachment details to configure Psc connections
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.psc_service_attachments
	PSCServiceAttachments []PSCServiceAttachmentObservedState `json:"pscServiceAttachments,omitempty"`

	// Optional. A list of cluster enpoints.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.cluster_endpoints
	ClusterEndpoints []ClusterEndpointObservedState `json:"clusterEndpoints,omitempty"`

	// Optional. Output only. The backup collection full resource name. Example:
	//  projects/{project}/locations/{location}/backupCollections/{collection}
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.backup_collection
	BackupCollection *string `json:"backupCollection,omitempty"`

	// Output only. Encryption information of the data at rest of the cluster.
	// +kcc:proto:field=google.cloud.redis.cluster.v1.Cluster.encryption_info
	EncryptionInfo *EncryptionInfoObservedState `json:"encryptionInfo,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcprediscluster;gcpredisclusters
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// RedisCluster is the Schema for the RedisCluster API
// +k8s:openapi-gen=true
type RedisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   RedisClusterSpec   `json:"spec,omitempty"`
	Status RedisClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// RedisClusterList contains a list of RedisCluster
type RedisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RedisCluster{}, &RedisClusterList{})
}
