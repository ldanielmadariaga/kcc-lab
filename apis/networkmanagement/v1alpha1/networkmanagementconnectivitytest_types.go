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

var NetworkManagementConnectivityTestGVK = GroupVersion.WithKind("NetworkManagementConnectivityTest")

// NetworkManagementConnectivityTestSpec defines the desired state of NetworkManagementConnectivityTest
// +kcc:spec:proto=google.cloud.networkmanagement.v1.ConnectivityTest
type NetworkManagementConnectivityTestSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The NetworkManagementConnectivityTest name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The user-supplied description of the Connectivity Test.
	//  Maximum of 512 characters.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.description
	Description *string `json:"description,omitempty"`

	// Required. Source specification of the Connectivity Test.
	//
	//  You can use a combination of source IP address, URI of a supported
	//  endpoint, project ID, or VPC network to identify the source location.
	//
	//  Reachability analysis might proceed even if the source location is
	//  ambiguous. However, the test result might include endpoints or use a source
	//  that you don't intend to test.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.source
	// +required
	Source *Endpoint `json:"source,omitempty"`

	// Required. Destination specification of the Connectivity Test.
	//
	//  You can use a combination of destination IP address, URI of a supported
	//  endpoint, project ID, or VPC network to identify the destination location.
	//
	//  Reachability analysis proceeds even if the destination location is
	//  ambiguous. However, the test result might include endpoints or use a
	//  destination that you don't intend to test.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.destination
	// +required
	Destination *Endpoint `json:"destination,omitempty"`

	// IP Protocol of the test. When not provided, "TCP" is assumed.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.protocol
	Protocol *string `json:"protocol,omitempty"`

	// Other projects that may be relevant for reachability analysis.
	//  This is applicable to scenarios where a test can cross project boundaries.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.related_projects
	RelatedProjects []string `json:"relatedProjects,omitempty"`

	// Resource labels to represent user-provided metadata.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Whether run analysis for the return path from destination to source.
	//  Default value is false.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.round_trip
	RoundTrip *bool `json:"roundTrip,omitempty"`

	// Whether the analysis should skip firewall checking. Default value is false.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.bypass_firewall_checks
	BypassFirewallChecks *bool `json:"bypassFirewallChecks,omitempty"`
}

// NetworkManagementConnectivityTestStatus defines the config connector machine state of NetworkManagementConnectivityTest
type NetworkManagementConnectivityTestStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the NetworkManagementConnectivityTest resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *NetworkManagementConnectivityTestObservedState `json:"observedState,omitempty"`
}

// NetworkManagementConnectivityTestObservedState is the state of the NetworkManagementConnectivityTest resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.networkmanagement.v1.ConnectivityTest
type NetworkManagementConnectivityTestObservedState struct {
	// Required. Source specification of the Connectivity Test.
	//
	//  You can use a combination of source IP address, URI of a supported
	//  endpoint, project ID, or VPC network to identify the source location.
	//
	//  Reachability analysis might proceed even if the source location is
	//  ambiguous. However, the test result might include endpoints or use a source
	//  that you don't intend to test.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.source
	Source *EndpointObservedState `json:"source,omitempty"`

	// Required. Destination specification of the Connectivity Test.
	//
	//  You can use a combination of destination IP address, URI of a supported
	//  endpoint, project ID, or VPC network to identify the destination location.
	//
	//  Reachability analysis proceeds even if the destination location is
	//  ambiguous. However, the test result might include endpoints or use a
	//  destination that you don't intend to test.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.destination
	Destination *EndpointObservedState `json:"destination,omitempty"`

	// Output only. The display name of a Connectivity Test.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Output only. The time the test was created.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The time the test's configuration was updated.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The reachability details of this test from the latest run.
	//  The details are updated when creating a new test, updating an
	//  existing test, or triggering a one-time rerun of an existing test.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.reachability_details
	ReachabilityDetails *ReachabilityDetailsObservedState `json:"reachabilityDetails,omitempty"`

	// Output only. The probing details of this test from the latest run, present
	//  for applicable tests only. The details are updated when creating a new
	//  test, updating an existing test, or triggering a one-time rerun of an
	//  existing test.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.probing_details
	ProbingDetails *ProbingDetails `json:"probingDetails,omitempty"`

	// Output only. The reachability details of this test from the latest run for
	//  the return path. The details are updated when creating a new test,
	//  updating an existing test, or triggering a one-time rerun of an existing
	//  test.
	// +kcc:proto:field=google.cloud.networkmanagement.v1.ConnectivityTest.return_reachability_details
	ReturnReachabilityDetails *ReachabilityDetailsObservedState `json:"returnReachabilityDetails,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpnetworkmanagementconnectivitytest;gcpnetworkmanagementconnectivitytests
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// NetworkManagementConnectivityTest is the Schema for the NetworkManagementConnectivityTest API
// +k8s:openapi-gen=true
type NetworkManagementConnectivityTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   NetworkManagementConnectivityTestSpec   `json:"spec,omitempty"`
	Status NetworkManagementConnectivityTestStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkManagementConnectivityTestList contains a list of NetworkManagementConnectivityTest
type NetworkManagementConnectivityTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkManagementConnectivityTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkManagementConnectivityTest{}, &NetworkManagementConnectivityTestList{})
}
