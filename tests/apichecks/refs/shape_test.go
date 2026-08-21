package refs

import "testing"

// A reference is identified by shape, not by name. Upstream models
// NetworkManagementConnectivityTest's destination.sqlInstance as a full
// reference while calling it "sqlInstance"; judged by name its external, name
// and namespace children read as three ordinary missing fields, which is how 25
// reference children were miscounted as unexplained drops.
func TestHasReferenceShape(t *testing.T) {
	tests := []struct {
		name  string
		props []string
		want  bool
	}{
		{name: "a reference", props: []string{"external", "name", "namespace"}, want: true},
		{name: "order does not matter", props: []string{"namespace", "external", "name"}, want: true},
		{name: "no external is not a reference", props: []string{"name", "namespace"}, want: false},
		{name: "external alone is not enough", props: []string{"external"}, want: false},
		{name: "an ordinary message", props: []string{"displayName", "labels"}, want: false},
		{name: "empty object", props: nil, want: false},
	}
	for _, tt := range tests {
		if got := HasReferenceShape(tt.props); got != tt.want {
			t.Errorf("%s: HasReferenceShape(%v) = %v, want %v", tt.name, tt.props, got, tt.want)
		}
	}
}
