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

var CertificateManagerDNSAuthorizationGVK = GroupVersion.WithKind("CertificateManagerDNSAuthorization")

// CertificateManagerDNSAuthorizationSpec defines the desired state of CertificateManagerDNSAuthorization
// +kcc:spec:proto=google.cloud.certificatemanager.v1.DnsAuthorization
type CertificateManagerDNSAuthorizationSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/dnsAuthorizations/{dns_authorization}
	Location *string `json:"location"`

	// The CertificateManagerDNSAuthorization name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Set of labels associated with a DnsAuthorization.
	// +kcc:proto:field=google.cloud.certificatemanager.v1.DnsAuthorization.labels
	Labels map[string]string `json:"labels,omitempty"`

	// One or more paragraphs of text description of a DnsAuthorization.
	// +kcc:proto:field=google.cloud.certificatemanager.v1.DnsAuthorization.description
	Description *string `json:"description,omitempty"`

	// Required. Immutable. A domain that is being authorized. A DnsAuthorization
	//  resource covers a single domain and its wildcard, e.g. authorization for
	//  `example.com` can be used to issue certificates for `example.com` and
	//  `*.example.com`.
	// +kcc:proto:field=google.cloud.certificatemanager.v1.DnsAuthorization.domain
	// +required
	Domain *string `json:"domain,omitempty"`

	// Immutable. Type of DnsAuthorization. If unset during resource creation the
	//  following default will be used:
	//  - in location global: FIXED_RECORD.
	// +kcc:proto:field=google.cloud.certificatemanager.v1.DnsAuthorization.type
	Type *string `json:"type,omitempty"`
}

// CertificateManagerDNSAuthorizationStatus defines the config connector machine state of CertificateManagerDNSAuthorization
type CertificateManagerDNSAuthorizationStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the CertificateManagerDNSAuthorization resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *CertificateManagerDNSAuthorizationObservedState `json:"observedState,omitempty"`
}

// CertificateManagerDNSAuthorizationObservedState is the state of the CertificateManagerDNSAuthorization resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.certificatemanager.v1.DnsAuthorization
type CertificateManagerDNSAuthorizationObservedState struct {
	// Output only. The creation timestamp of a DnsAuthorization.
	// +kcc:proto:field=google.cloud.certificatemanager.v1.DnsAuthorization.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The last update timestamp of a DnsAuthorization.
	// +kcc:proto:field=google.cloud.certificatemanager.v1.DnsAuthorization.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. DNS Resource Record that needs to be added to DNS
	//  configuration.
	// +kcc:proto:field=google.cloud.certificatemanager.v1.DnsAuthorization.dns_resource_record
	DNSResourceRecord *DNSAuthorization_DNSResourceRecordObservedState `json:"dnsResourceRecord,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcertificatemanagerdnsauthorization;gcpcertificatemanagerdnsauthorizations
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// CertificateManagerDNSAuthorization is the Schema for the CertificateManagerDNSAuthorization API
// +k8s:openapi-gen=true
type CertificateManagerDNSAuthorization struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   CertificateManagerDNSAuthorizationSpec   `json:"spec,omitempty"`
	Status CertificateManagerDNSAuthorizationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// CertificateManagerDNSAuthorizationList contains a list of CertificateManagerDNSAuthorization
type CertificateManagerDNSAuthorizationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertificateManagerDNSAuthorization `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertificateManagerDNSAuthorization{}, &CertificateManagerDNSAuthorizationList{})
}
