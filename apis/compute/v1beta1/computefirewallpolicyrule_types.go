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

var ComputeFirewallPolicyRuleGVK = GroupVersion.WithKind("ComputeFirewallPolicyRule")

// ComputeFirewallPolicyRuleSpec defines the desired state of ComputeFirewallPolicyRule
// +kcc:spec:proto=google.cloud.compute.v1.FirewallPolicyRule
type ComputeFirewallPolicyRuleSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The ComputeFirewallPolicyRule name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The Action to perform when the client connection triggers the rule. Valid actions for firewall rules are: "allow", "deny", "apply_security_profile_group" and "goto_next". Valid actions for packet mirroring rules are: "mirror", "do_not_mirror" and "goto_next".
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.action
	Action *string `json:"action,omitempty"`

	// An optional description for this resource.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.description
	Description *string `json:"description,omitempty"`

	// The direction in which this rule applies.
	//  Check the Direction enum for the list of possible values.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.direction
	Direction *string `json:"direction,omitempty"`

	// Denotes whether the firewall policy rule is disabled. When set to true, the firewall policy rule is not enforced and traffic behaves as if it did not exist. If this is unspecified, the firewall policy rule will be enabled.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.disabled
	Disabled *bool `json:"disabled,omitempty"`

	// Denotes whether to enable logging for a particular rule. If logging is enabled, logs will be exported to the configured export destination in Stackdriver. Logs may be exported to BigQuery or Pub/Sub. Note: you cannot enable logging on "goto_next" rules.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.enable_logging
	EnableLogging *bool `json:"enableLogging,omitempty"`

	// A match condition that incoming traffic is evaluated against. If it evaluates to true, the corresponding 'action' is enforced.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.match
	Match *FirewallPolicyRuleMatcher `json:"match,omitempty"`

	// An integer indicating the priority of a rule in the list. The priority must be a positive value between 0 and 2147483647. Rules are evaluated from highest to lowest priority where 0 is the highest priority and 2147483647 is the lowest priority.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.priority
	Priority *int32 `json:"priority,omitempty"`

	// An optional name for the rule. This field is not a unique identifier and can be updated.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.rule_name
	RuleName *string `json:"ruleName,omitempty"`

	// [Output Only] Calculation of the complexity of a single firewall policy rule.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.rule_tuple_count
	RuleTupleCount *int32 `json:"ruleTupleCount,omitempty"`

	// A fully-qualified URL of a SecurityProfile resource instance. Example: https://networksecurity.googleapis.com/v1/projects/{project}/locations/{location}/securityProfileGroups/my-security-profile-group Must be specified if action is one of 'apply_security_profile_group' or 'mirror'. Cannot be specified for other actions.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.security_profile_group
	SecurityProfileGroup *string `json:"securityProfileGroup,omitempty"`

	// A list of network resource URLs to which this rule applies. This field allows you to control which network's VMs get this rule. If this field is left blank, all VMs within the organization will receive the rule.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.target_resources
	TargetResources []string `json:"targetResources,omitempty"`

	// A list of secure tags that controls which instances the firewall rule applies to. If targetSecureTag are specified, then the firewall rule applies only to instances in the VPC network that have one of those EFFECTIVE secure tags, if all the target_secure_tag are in INEFFECTIVE state, then this rule will be ignored. targetSecureTag may not be set at the same time as targetServiceAccounts. If neither targetServiceAccounts nor targetSecureTag are specified, the firewall rule applies to all instances on the specified network. Maximum number of target label tags allowed is 256.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.target_secure_tags
	TargetSecureTags []FirewallPolicyRuleSecureTag `json:"targetSecureTags,omitempty"`

	// A list of service accounts indicating the sets of instances that are applied with this rule.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.target_service_accounts
	TargetServiceAccounts []string `json:"targetServiceAccounts,omitempty"`

	// Boolean flag indicating if the traffic should be TLS decrypted. Can be set only if action = 'apply_security_profile_group' and cannot be set for other actions.
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.tls_inspect
	TLSInspect *bool `json:"tlsInspect,omitempty"`
}

// ComputeFirewallPolicyRuleStatus defines the config connector machine state of ComputeFirewallPolicyRule
type ComputeFirewallPolicyRuleStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the ComputeFirewallPolicyRule resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *ComputeFirewallPolicyRuleObservedState `json:"observedState,omitempty"`
}

// ComputeFirewallPolicyRuleObservedState is the state of the ComputeFirewallPolicyRule resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.compute.v1.FirewallPolicyRule
type ComputeFirewallPolicyRuleObservedState struct {
	// [Output only] Type of the resource. Returns compute#firewallPolicyRule for firewall rules and compute#packetMirroringRule for packet mirroring rules.
	// +kcc:guess=placement reason=no-field-behavior-on-message
	// +kcc:proto:field=google.cloud.compute.v1.FirewallPolicyRule.kind
	Kind *string `json:"kind,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcomputefirewallpolicyrule;gcpcomputefirewallpolicyrules
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// ComputeFirewallPolicyRule is the Schema for the ComputeFirewallPolicyRule API
// +k8s:openapi-gen=true
type ComputeFirewallPolicyRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ComputeFirewallPolicyRuleSpec   `json:"spec,omitempty"`
	Status ComputeFirewallPolicyRuleStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ComputeFirewallPolicyRuleList contains a list of ComputeFirewallPolicyRule
type ComputeFirewallPolicyRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ComputeFirewallPolicyRule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ComputeFirewallPolicyRule{}, &ComputeFirewallPolicyRuleList{})
}
