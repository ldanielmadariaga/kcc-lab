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

var NetworkSecurityAuthzPolicyGVK = GroupVersion.WithKind("NetworkSecurityAuthzPolicy")

// NetworkSecurityAuthzPolicySpec defines the desired state of NetworkSecurityAuthzPolicy
// +kcc:spec:proto=google.cloud.networksecurity.v1.AuthzPolicy
type NetworkSecurityAuthzPolicySpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/authzPolicies/{authz_policy}
	Location *string `json:"location"`

	// The NetworkSecurityAuthzPolicy name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. A human-readable description of the resource.
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.description
	Description *string `json:"description,omitempty"`

	// Optional. Set of labels associated with the `AuthzPolicy` resource.
	//
	//  The format must comply with [the following
	//  requirements](/compute/docs/labeling-resources#requirements).
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.labels
	Labels map[string]string `json:"labels,omitempty"`

	// Required. Specifies the set of resources to which this policy should be
	//  applied to.
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.target
	// +required
	Target *AuthzPolicy_Target `json:"target,omitempty"`

	// Optional. A list of authorization HTTP rules to match against the incoming
	//  request. A policy match occurs when at least one HTTP rule matches the
	//  request or when no HTTP rules are specified in the policy.
	//  At least one HTTP Rule is required for Allow or Deny Action. Limited
	//  to 5 rules.
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.http_rules
	HTTPRules []AuthzPolicy_AuthzRule `json:"httpRules,omitempty"`

	// Optional. A list of authorization network rules to match against the
	//  incoming request. A policy match occurs when at least one network rule
	//  matches the request.
	//  At least one network rule is required for Allow or Deny Action if no HTTP
	//  rules are provided. Network rules are mutually exclusive with HTTP rules.
	//  Limited to 5 rules.
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.network_rules
	NetworkRules []AuthzPolicy_AuthzRule `json:"networkRules,omitempty"`

	// Required. Can be one of `ALLOW`, `DENY`, `CUSTOM`.
	//
	//  When the action is `CUSTOM`, `customProvider` must be specified.
	//
	//  When the action is `ALLOW`, only requests matching the policy will
	//  be allowed.
	//
	//  When the action is `DENY`, only requests matching the policy will be
	//  denied.
	//
	//  When a request arrives, the policies are evaluated in the following order:
	//
	//  1. If there is a `CUSTOM` policy that matches the request, the `CUSTOM`
	//  policy is evaluated using the custom authorization providers and the
	//  request is denied if the provider rejects the request.
	//
	//  2. If there are any `DENY` policies that match the request, the request
	//  is denied.
	//
	//  3. If there are no `ALLOW` policies for the resource or if any of the
	//  `ALLOW` policies match the request, the request is allowed.
	//
	//  4. Else the request is denied by default if none of the configured
	//  AuthzPolicies with `ALLOW` action match the request.
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.action
	// +required
	Action *string `json:"action,omitempty"`

	// Optional. Required if the action is `CUSTOM`. Allows delegating
	//  authorization decisions to Cloud IAP or to Service Extensions. One of
	//  `cloudIap` or `authzExtension` must be specified.
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.custom_provider
	CustomProvider *AuthzPolicy_CustomProvider `json:"customProvider,omitempty"`

	// Optional. Immutable. Defines the type of authorization being performed.
	//  If not specified, `REQUEST_AUTHZ` is applied. This field cannot be changed
	//  once AuthzPolicy is created.
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.policy_profile
	PolicyProfile *string `json:"policyProfile,omitempty"`
}

// NetworkSecurityAuthzPolicyStatus defines the config connector machine state of NetworkSecurityAuthzPolicy
type NetworkSecurityAuthzPolicyStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the NetworkSecurityAuthzPolicy resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *NetworkSecurityAuthzPolicyObservedState `json:"observedState,omitempty"`
}

// NetworkSecurityAuthzPolicyObservedState is the state of the NetworkSecurityAuthzPolicy resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.networksecurity.v1.AuthzPolicy
type NetworkSecurityAuthzPolicyObservedState struct {
	// Output only. The timestamp when the resource was created.
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The timestamp when the resource was updated.
	// +kcc:proto:field=google.cloud.networksecurity.v1.AuthzPolicy.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpnetworksecurityauthzpolicy;gcpnetworksecurityauthzpolicys
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// NetworkSecurityAuthzPolicy is the Schema for the NetworkSecurityAuthzPolicy API
// +k8s:openapi-gen=true
type NetworkSecurityAuthzPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   NetworkSecurityAuthzPolicySpec   `json:"spec,omitempty"`
	Status NetworkSecurityAuthzPolicyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// NetworkSecurityAuthzPolicyList contains a list of NetworkSecurityAuthzPolicy
type NetworkSecurityAuthzPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkSecurityAuthzPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkSecurityAuthzPolicy{}, &NetworkSecurityAuthzPolicyList{})
}
