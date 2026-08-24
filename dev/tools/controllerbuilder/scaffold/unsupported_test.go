package scaffold

import "testing"

// WriteField replaces a field it cannot type with a "// TODO:" comment and moves
// on, so the field never reaches the CRD. Measured on the 239-resource run: 15
// such markers in scaffolded type files and 37 in types.generated.go, together
// costing 124 CRD field paths, recorded nowhere a person would look.
func TestUnsupportedFieldReason(t *testing.T) {
	tests := []struct {
		name     string
		rendered string
		want     string
		wantOK   bool
	}{
		{
			name:     "field name stripped, since FieldPath carries it",
			rendered: "\n\t// TODO: attributes: unsupported map type with key string and value message\n\n",
			want:     "unsupported map type with key string and value message",
			wantOK:   true,
		},
		{
			name:     "marker without a field name still reports",
			rendered: "\n\t// TODO: unsupported map type\n\n",
			want:     "unsupported map type",
			wantOK:   true,
		},
		{
			name:     "an ordinary field is not a finding",
			rendered: "\t// +kcc:proto:field=pkg.Msg.name\n\tName *string `json:\"name,omitempty\"`\n",
		},
		{
			name:     "a doc comment mentioning TODO is not a marker",
			rendered: "\t// Some description with TODO in prose\n\tName *string `json:\"name,omitempty\"`\n",
		},
	}
	for _, tt := range tests {
		got, ok := unsupportedFieldReason(tt.rendered)
		if ok != tt.wantOK {
			t.Errorf("%s: ok = %v, want %v", tt.name, ok, tt.wantOK)
			continue
		}
		if ok && got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, got, tt.want)
		}
	}
}
