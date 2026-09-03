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

var ContentWarehouseDocumentGVK = GroupVersion.WithKind("ContentWarehouseDocument")

// ContentWarehouseDocumentSpec defines the desired state of ContentWarehouseDocument
// +kcc:spec:proto=google.cloud.contentwarehouse.v1.Document
type ContentWarehouseDocumentSpec struct {
	// The project that this resource belongs to.
	ProjectRef *refsv1beta1.ProjectRef `json:"projectRef"`

	// The location of this resource.
	// +kcc:guess=parent-location pattern=projects/{project}/locations/{location}/documents/{document}
	Location *string `json:"location"`

	// The ContentWarehouseDocument name. If not given, the metadata.name will be used.
	ResourceID *string `json:"resourceID,omitempty"`
	// The reference ID set by customers. Must be unique per project and location.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.reference_id
	ReferenceID *string `json:"referenceID,omitempty"`

	// Required. Display name of the document given by the user. This name will be
	//  displayed in the UI. Customer can populate this field with the name of the
	//  document. This differs from the 'title' field as 'title' is optional and
	//  stores the top heading in the document.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.display_name
	// +required
	DisplayName *string `json:"displayName,omitempty"`

	// Title that describes the document.
	//  This can be the top heading or text that describes the document.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.title
	Title *string `json:"title,omitempty"`

	// Uri to display the document, for example, in the UI.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.display_uri
	DisplayURI *string `json:"displayURI,omitempty"`

	// The Document schema name.
	//  Format:
	//  projects/{project_number}/locations/{location}/documentSchemas/{document_schema_id}.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.document_schema_name
	DocumentSchemaName *string `json:"documentSchemaName,omitempty"`

	// Other document format, such as PPTX, XLXS
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.plain_text
	PlainText *string `json:"plainText,omitempty"`

	// Document AI format to save the structured content, including OCR.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.cloud_ai_document
	CloudAiDocument *Document `json:"cloudAiDocument,omitempty"`

	// A path linked to structured content file.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.structured_content_uri
	StructuredContentURI *string `json:"structuredContentURI,omitempty"`

	// Raw document file in Cloud Storage path.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.raw_document_path
	RawDocumentPath *string `json:"rawDocumentPath,omitempty"`

	// Raw document content.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.inline_raw_document
	InlineRawDocument []byte `json:"inlineRawDocument,omitempty"`

	// List of values that are user supplied metadata.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.properties
	Properties []Property `json:"properties,omitempty"`

	// This is used when DocAI was not used to load the document and parsing/
	//  extracting is needed for the inline_raw_document.  For example, if
	//  inline_raw_document is the byte representation of a PDF file, then
	//  this should be set to: RAW_DOCUMENT_FILE_TYPE_PDF.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.raw_document_file_type
	RawDocumentFileType *string `json:"rawDocumentFileType,omitempty"`

	// If true, makes the document visible to asynchronous policies and rules.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.async_enabled
	AsyncEnabled *bool `json:"asyncEnabled,omitempty"`

	// Indicates the category (image, audio, video etc.) of the original content.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.content_category
	ContentCategory *string `json:"contentCategory,omitempty"`

	// If true, text extraction will not be performed.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.text_extraction_disabled
	TextExtractionDisabled *bool `json:"textExtractionDisabled,omitempty"`

	// If true, text extraction will be performed.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.text_extraction_enabled
	TextExtractionEnabled *bool `json:"textExtractionEnabled,omitempty"`

	// The user who creates the document.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.creator
	Creator *string `json:"creator,omitempty"`

	// The user who lastly updates the document.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.updater
	Updater *string `json:"updater,omitempty"`
}

// ContentWarehouseDocumentStatus defines the config connector machine state of ContentWarehouseDocument
type ContentWarehouseDocumentStatus struct {
	/* Conditions represent the latest available observations of the
	   object's current state. */
	Conditions []v1alpha1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the generation of the resource that was most recently observed by the Config Connector controller. If this is equal to metadata.generation, then that means that the current reported status reflects the most recent desired state of the resource.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`

	// A unique specifier for the ContentWarehouseDocument resource in GCP.
	ExternalRef *string `json:"externalRef,omitempty"`

	// ObservedState is the state of the resource as most recently observed in GCP.
	ObservedState *ContentWarehouseDocumentObservedState `json:"observedState,omitempty"`
}

// ContentWarehouseDocumentObservedState is the state of the ContentWarehouseDocument resource as most recently observed in GCP.
// +kcc:observedstate:proto=google.cloud.contentwarehouse.v1.Document
type ContentWarehouseDocumentObservedState struct {
	// Output only. The time when the document is last updated.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.update_time
	UpdateTime *string `json:"updateTime,omitempty"`

	// Output only. The time when the document is created.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.create_time
	CreateTime *string `json:"createTime,omitempty"`

	// Output only. If linked to a Collection with RetentionPolicy, the date when
	//  the document becomes mutable.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.disposition_time
	DispositionTime *string `json:"dispositionTime,omitempty"`

	// Output only. Indicates if the document has a legal hold on it.
	// +kcc:proto:field=google.cloud.contentwarehouse.v1.Document.legal_hold
	LegalHold *bool `json:"legalHold,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories=gcp,shortName=gcpcontentwarehousedocument;gcpcontentwarehousedocuments
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/managed-by-kcc=true"
// +kubebuilder:metadata:labels="cnrm.cloud.google.com/system=true"
// +kubebuilder:printcolumn:name="Age",JSONPath=".metadata.creationTimestamp",type="date"
// +kubebuilder:printcolumn:name="Ready",JSONPath=".status.conditions[?(@.type=='Ready')].status",type="string",description="When 'True', the most recent reconcile of the resource succeeded"
// +kubebuilder:printcolumn:name="Status",JSONPath=".status.conditions[?(@.type=='Ready')].reason",type="string",description="The reason for the value in 'Ready'"
// +kubebuilder:printcolumn:name="Status Age",JSONPath=".status.conditions[?(@.type=='Ready')].lastTransitionTime",type="date",description="The last transition time for the value in 'Status'"

// ContentWarehouseDocument is the Schema for the ContentWarehouseDocument API
// +k8s:openapi-gen=true
type ContentWarehouseDocument struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +required
	Spec   ContentWarehouseDocumentSpec   `json:"spec,omitempty"`
	Status ContentWarehouseDocumentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ContentWarehouseDocumentList contains a list of ContentWarehouseDocument
type ContentWarehouseDocumentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContentWarehouseDocument `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContentWarehouseDocument{}, &ContentWarehouseDocumentList{})
}
