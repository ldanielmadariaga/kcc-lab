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

package codegen

import "testing"

func TestKRMMapValueType(t *testing.T) {
	for _, tc := range []struct {
		name          string
		elemType      string
		krmFieldType  string
		wantGoType    string
		wantIsPointer bool
	}{
		{
			name:         "a generated struct takes the krm alias",
			elemType:     "TargetMessage",
			krmFieldType: "map[string]TargetMessage",
			wantGoType:   "krm.TargetMessage",
		},
		{
			name:          "a pointer to a generated struct",
			elemType:      "TargetMessage",
			krmFieldType:  "map[string]*TargetMessage",
			wantGoType:    "*krm.TargetMessage",
			wantIsPointer: true,
		},
		{
			name:         "a qualified type keeps its own qualifier",
			elemType:     "apiextensionsv1.JSON",
			krmFieldType: "map[string]apiextensionsv1.JSON",
			wantGoType:   "apiextensionsv1.JSON",
		},
		{
			name:         "a built-in type has no qualifier",
			elemType:     "string",
			krmFieldType: "map[string]string",
			wantGoType:   "string",
		},
		{
			name:          "a pointer to a built-in type",
			elemType:      "int64",
			krmFieldType:  "map[string]*int64",
			wantGoType:    "*int64",
			wantIsPointer: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			goType, isPointer := krmMapValueType(tc.elemType, tc.krmFieldType, "krm")

			// Assert
			if goType != tc.wantGoType || isPointer != tc.wantIsPointer {
				t.Errorf("krmMapValueType(%q, %q) = (%q, %v), want (%q, %v)",
					tc.elemType, tc.krmFieldType, goType, isPointer, tc.wantGoType, tc.wantIsPointer)
			}
		})
	}
}
