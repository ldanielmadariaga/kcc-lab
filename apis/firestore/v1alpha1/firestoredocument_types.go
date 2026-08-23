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

var FirestoreDocumentGVK = GroupVersion.WithKind("FirestoreDocument")

// FirestoreDocumentSpec defines the desired state of FirestoreDocument
// +kcc:spec:proto=google.firestore.v1.Document
type FirestoreDocumentSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The FirestoreDocument name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The document's fields.
	//
	//  The map keys represent field names.
	//
	//  Field names matching the regular expression `__.*__` are reserved. Reserved
	//  field names are forbidden except in certain documented contexts. The field
	//  names, represented as UTF-8, must not exceed 1,500 bytes and cannot be
	//  empty.
	//
	//  Field paths may be used in other contexts to refer to structured fields
	//  defined here. For `map_value`, the field path is represented by a
	//  dot-delimited (`.`) string of segments. Each segment is either a simple
	//  field name (defined below) or a quoted field name. For example, the
	//  structured field `"foo" : { map_value: { "x&y" : { string_value: "hello"
	//  }}}` would be represented by the field path `` foo.`x&y` ``.
	//
	//  A simple field name contains only characters `a` to `z`, `A` to `Z`,
	//  `0` to `9`, or `_`, and must not start with `0` to `9`. For example,
	//  `foo_bar_17`.
	//
	//  A quoted field name starts and ends with `` ` `` and
	//  may contain any character. Some characters, including `` ` ``, must be
	//  escaped using a `\`. For example, `` `x&y` `` represents `x&y` and
	//  `` `bak\`tik` `` represents `` bak`tik ``.
	// +kcc:proto:field=google.firestore.v1.Document.fields
	Fields map[string]Value `json:"fields,omitempty"`
}

// FirestoreDocumentStatus defines the config connector machine state of FirestoreDocument
type FirestoreDocumentStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the FirestoreDocument resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *FirestoreDocumentObservedState `json:"observedState,omitempty"`
}

// FirestoreDocumentObservedState is the state of the FirestoreDocument resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.firestore.v1.Document
type FirestoreDocumentObservedState struct {
	// Output only. The time at which the document was created.
	//
	//  This value increases monotonically when a document is deleted then
	//  recreated. It can also be compared to values from other documents and
	//  the `read_time` of a query.
	// +kcc:guess=placement reason=no-field-behavior-on-message
	// +kcc:proto:field=google.firestore.v1.Document.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. The time at which the document was last changed.
	//
	//  This value is initially set to the `create_time` then increases
	//  monotonically with each change to the document. It can also be
	//  compared to values from other documents and the `read_time` of a query.
	// +kcc:guess=placement reason=no-field-behavior-on-message
	// +kcc:proto:field=google.firestore.v1.Document.update_time
	UpdateTime *string `json:"updateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpfirestoredocument;gcpfirestoredocuments
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// FirestoreDocument is the Schema for the FirestoreDocument API
// +k8s:openapi-gen=true
type FirestoreDocument struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   FirestoreDocumentSpec   `json:"spec,omitempty"`
	Status FirestoreDocumentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// FirestoreDocumentList contains a list of FirestoreDocument
type FirestoreDocumentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FirestoreDocument `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FirestoreDocument{}, &FirestoreDocumentList{})
}
