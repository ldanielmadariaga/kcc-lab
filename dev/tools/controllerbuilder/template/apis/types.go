// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apis

type APIArgs struct {
	Group           string
	Version         string
	Kind            string
	ProtoResource   string
	PackageProtoTag string
	KindProtoTag    string

	// ProtoMessageName is the last component of the proto message name, e.g. for google.cloud.v1.Foo, it will be "Foo"
	ProtoMessageName string
	// ProtoMessageFullName is the fully qualified proto message name, e.g. google.cloud.v1.Foo
	ProtoMessageFullName string

	// Collection is the resource's collection segment as the API spells it, taken
	// from google.api.resource, e.g. "lbTrafficExtensions". Empty when the proto
	// declares no pattern, in which case templates fall back to guessing.
	Collection string
	// ParentRefFields holds every part of the resource's name below the root:
	// refs to parent resources, plain strings where no ref type exists, and the
	// location. Rendered by the scaffolder, since only it can tell which ref
	// types are available.
	ParentRefFields string
	// RootRefType, RootRefField and RootRefDescription name the ref for the root
	// of the resource's name -- ProjectRef for nearly everything, OrganizationRef
	// or FolderRef for a resource rooted outside a project, which has no project
	// to point at.
	RootRefType        string
	RootRefField       string
	RootRefDescription string
	// ParentStyle is the shape of the resource's parent: "project_location",
	// "project", "organization", "folder", "other" or "unknown".
	ParentStyle string
	// ResourcePattern is the declared pattern, carried through verbatim so the
	// generated file can show a human what the real name looks like.
	ResourcePattern string
	// SpecFields is pre-rendered Go source for the body of the Spec struct, one
	// field per proto field. Empty means the old three-field stub, which is what
	// every service gets until it opts in.
	SpecFields string
	// ObservedStateFields is the same for the ObservedState struct. Empty leaves
	// it empty, which is what a proto with no OUTPUT_ONLY field should produce.
	ObservedStateFields string
	// ExtraImports are complete aliased import lines the rendered fields need
	// beyond the three below, e.g. common "github.com/.../apis/common" when a
	// field is a google.rpc.Status. Aliased because the qualifier a field uses
	// is often not the import path's last segment.
	ExtraImports []string
}

// Location, emitted only when the proto's resource pattern makes the parent
// project+location, is a pointer: 100 of the 129 existing resources that
// declare one use that form, and the value form does not compile against
// hand-written identity files that dereference it. It carries no omitempty, so
// the field stays required, as upstream has it.
//
// Its comment in the template is the one-line "The location of this resource."
// and nothing more. Whatever is written there becomes the field's CRD
// description, which users read; a rationale for the generator's own choice
// does not belong in a published API schema.
const TypesTemplate = `
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

package {{ .Version }}

import (
	refsv1beta1 "github.com/GoogleCloudPlatform/k8s-config-connector/apis/refs/v1beta1"
	"github.com/GoogleCloudPlatform/k8s-config-connector/pkg/apis/k8s/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
{{- range .ExtraImports }}
	{{ . }}
{{- end }}
)

var {{ .Kind }}GVK = GroupVersion.WithKind("{{ .Kind }}")

// {{ .Kind }}Spec defines the desired state of {{ .Kind }}
{{- if .KindProtoTag }}
// +kcc:spec:proto={{ .KindProtoTag }}
{{- end }}
type {{ .Kind }}Spec struct {
	// {{ .RootRefDescription }}
	{{ .RootRefType }} *refsv1beta1.{{ .RootRefType }} ` + "`" + `json:"{{ .RootRefField }}"` + "`" + `
{{ .ParentRefFields }}

	// The {{ .Kind }} name. If not given, the metadata.name will be used.
	ResourceID *string ` + "`" + `json:"resourceID,omitempty"` + "`" + `
{{ .SpecFields }}}

// {{ .Kind }}Status defines the config connector machine state of {{ .Kind }}
type {{ .Kind }}Status struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition ` + "`" + `json:"conditions,omitempty"` + "`" + ` 

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 ` + "`" + `json:"observedGeneration,omitempty"` + "`" + `

	// A unique specifier for the {{ .Kind }} resource in GCP.
	ExternalRef *string ` + "`" + `json:"externalRef,omitempty"` + "`" + `

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *{{ .Kind }}ObservedState ` + "`" + `json:"observedState,omitempty"` + "`" + `
}

// {{ .Kind }}ObservedState is the state of the {{ .Kind }} resource as most recently observed in GCP.
{{- if .KindProtoTag }}
// +kcc:observedstate:proto={{ .KindProtoTag }}
{{- end }}
type {{ .Kind }}ObservedState struct {
{{ .ObservedStateFields }}}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcp{{ .Kind | ToLower }};gcp{{ .Kind | ToLower }}s
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// {{ .Kind }} is the Schema for the {{ .Kind }} API
// +k8s:openapi-gen=true
type {{ .Kind }} struct {
	metav1.TypeMeta   ` + "`" + `json:",inline"` + "`" + `
	metav1.ObjectMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `

	// +required
	Spec   {{ .Kind }}Spec   ` + "`" + `json:"spec,omitempty"` + "`" + `
	Status {{ .Kind }}Status ` + "`" + `json:"status,omitempty"` + "`" + `
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// {{ .Kind }}List contains a list of {{ .Kind }}
type {{ .Kind }}List struct {
	metav1.TypeMeta ` + "`" + `json:",inline"` + "`" + `
	metav1.ListMeta ` + "`" + `json:"metadata,omitempty"` + "`" + `
	Items           []{{ .Kind }} ` + "`" + `json:"items"` + "`" + `
}

func init() {
	SchemeBuilder.Register(&{{ .Kind }}{}, &{{ .Kind }}List{})
}
`
