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
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var DataLineageProcessGVK = GroupVersion.WithKind("DataLineageProcess")

// DataLineageProcessSpec defines the desired state of DataLineageProcess
// +kcc:spec:proto=google.cloud.datacatalog.lineage.v1.Process
type DataLineageProcessSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	Location *string `json:"location"`

	// The DataLineageProcess name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// Optional. A human-readable name you can set to display in a user interface.
	//  Must be not longer than 200 characters and only contain UTF-8 letters
	//  or numbers, spaces or characters like `_-:&.`
	// +kcc:proto:field=google.cloud.datacatalog.lineage.v1.Process.display_name
	DisplayName *string `json:"displayName,omitempty"`

	// Optional. The attributes of the process. Should only be used for the
	//  purpose of non-semantic management (classifying, describing or labeling the
	//  process).
	//
	//  Up to 100 attributes are allowed.
	// +kcc:proto:field=google.cloud.datacatalog.lineage.v1.Process.attributes
	Attributes map[string]apiextensionsv1.JSON `json:"attributes,omitempty"`

	// Optional. The origin of this process and its runs and lineage events.
	// +kcc:proto:field=google.cloud.datacatalog.lineage.v1.Process.origin
	Origin *Origin `json:"origin,omitempty"`
}

// DataLineageProcessStatus defines the config connector machine state of DataLineageProcess
type DataLineageProcessStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the DataLineageProcess resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *DataLineageProcessObservedState `json:"observedState,omitempty"`
}

// DataLineageProcessObservedState is the state of the DataLineageProcess resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.datacatalog.lineage.v1.Process
type DataLineageProcessObservedState struct {
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpdatalineageprocess;gcpdatalineageprocesss
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// DataLineageProcess is the Schema for the DataLineageProcess API
// +k8s:openapi-gen=true
type DataLineageProcess struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   DataLineageProcessSpec   `json:"spec,omitempty"`
	Status DataLineageProcessStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// DataLineageProcessList contains a list of DataLineageProcess
type DataLineageProcessList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DataLineageProcess `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DataLineageProcess{}, &DataLineageProcessList{})
}
