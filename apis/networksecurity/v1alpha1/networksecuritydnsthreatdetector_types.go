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

var NetworkSecurityDNSThreatDetectorGVK = GroupVersion.WithKind("NetworkSecurityDNSThreatDetector")

// NetworkSecurityDNSThreatDetectorSpec defines the desired state of NetworkSecurityDNSThreatDetector
// +kcc:spec:proto=google.cloud.networksecurity.v1.DnsThreatDetector
type NetworkSecurityDNSThreatDetectorSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/dnsThreatDetectors/{dns_threat_detector}
	Location *string `json:"location"`

	// The NetworkSecurityDNSThreatDetector name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. Any labels associated with the DnsThreatDetector, listed as key
	//  value pairs.
	// +kcc:proto:field=google.cloud.networksecurity.v1.DnsThreatDetector.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Optional. A list of network resource names which aren't monitored by this
	//  DnsThreatDetector.
	//
	//  Example:
	//  `projects/PROJECT_ID/global/networks/NETWORK_NAME`.
	// +kcc:proto:field=google.cloud.networksecurity.v1.DnsThreatDetector.excluded_networks
	ExcludedNetworks []string `json:"excludedNetworks,omitempty"`

	// Required. The provider used for DNS threat analysis.
	// +kcc:proto:field=google.cloud.networksecurity.v1.DnsThreatDetector.provider
	// +required
	Provider *string `json:"provider,omitempty"`
}

// NetworkSecurityDNSThreatDetectorStatus defines the config connector machine state of NetworkSecurityDNSThreatDetector
type NetworkSecurityDNSThreatDetectorStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the NetworkSecurityDNSThreatDetector resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *NetworkSecurityDNSThreatDetectorObservedState `json:"observedState,omitempty"`
}

// NetworkSecurityDNSThreatDetectorObservedState is the state of the NetworkSecurityDNSThreatDetector resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.networksecurity.v1.DnsThreatDetector
type NetworkSecurityDNSThreatDetectorObservedState struct {
	// Output only. Create time stamp.
	// +kcc:proto:field=google.cloud.networksecurity.v1.DnsThreatDetector.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. Update time stamp.
	// +kcc:proto:field=google.cloud.networksecurity.v1.DnsThreatDetector.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpnetworksecuritydnsthreatdetector;gcpnetworksecuritydnsthreatdetectors
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// NetworkSecurityDNSThreatDetector is the Schema for the NetworkSecurityDNSThreatDetector API
// +k8s:openapi-gen=true
type NetworkSecurityDNSThreatDetector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   NetworkSecurityDNSThreatDetectorSpec   `json:"spec,omitempty"`
	Status NetworkSecurityDNSThreatDetectorStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkSecurityDNSThreatDetectorList contains a list of NetworkSecurityDNSThreatDetector
type NetworkSecurityDNSThreatDetectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkSecurityDNSThreatDetector `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkSecurityDNSThreatDetector{}, &NetworkSecurityDNSThreatDetectorList{})
}
