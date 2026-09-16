// Copyright 2024 Google LLC
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

package direct

import (
	"fmt"
	"testing"

	"github.com/googleapis/gax-go/v2/apierror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestStringDuration_FromProto(t *testing.T) {
	mapctx := &MapContext{}
	d := &durationpb.Duration{Seconds: 34312, Nanos: 20}
	krm := StringDuration_FromProto(mapctx, d)
	if *krm != "9h31m52.00000002s" {
		t.Fatalf("google.protobuf.Duration -> string, expect \"9h31m52.00000002s\", got %s", *krm)
	}
	if mapctx.Err() != nil {
		t.Fatalf("google.protobuf.Duration -> string error: %s", mapctx.Err())
	}
}

func TestStringDuration_ToProto(t *testing.T) {
	mapctx := &MapContext{}
	s := "1h1m"
	d := StringDuration_ToProto(mapctx, &s)
	if d.Seconds != 3660 || d.Nanos != 0 {
		t.Fatalf("string -> google.protobuf.Duration, expect \"seconds:3660 nanos:00\", got %s", d)
	}
	if mapctx.Err() != nil {
		t.Fatalf("google.protobuf.Duration -> String error: %s", mapctx.Err())
	}
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "unrelated error",
			err:  fmt.Errorf("something went wrong"),
			want: false,
		},
		{
			name: "gRPC AlreadyExists",
			err:  status.Error(codes.AlreadyExists, "resource already exists"),
			want: true,
		},
		{
			name: "gRPC NotFound",
			err:  status.Error(codes.NotFound, "not found"),
			want: false,
		},
		{
			name: "gRPC AlreadyExists wrapped with apierror",
			err: func() error {
				grpcErr := status.Error(codes.AlreadyExists, "resource already exists")
				apiErr, _ := apierror.ParseError(grpcErr, true)
				return apiErr
			}(),
			want: true,
		},
		{
			name: "gRPC NotFound wrapped with apierror",
			err: func() error {
				grpcErr := status.Error(codes.NotFound, "not found")
				apiErr, _ := apierror.ParseError(grpcErr, true)
				return apiErr
			}(),
			want: false,
		},
		{
			name: "wrapped gRPC AlreadyExists",
			err:  fmt.Errorf("creating resource: %w", status.Error(codes.AlreadyExists, "already exists")),
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsAlreadyExists(tc.err)
			if got != tc.want {
				t.Errorf("IsAlreadyExists(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "unrelated error",
			err:  fmt.Errorf("something went wrong"),
			want: false,
		},
		{
			name: "gRPC NotFound",
			err:  status.Error(codes.NotFound, "not found"),
			want: true,
		},
		{
			name: "gRPC AlreadyExists",
			err:  status.Error(codes.AlreadyExists, "already exists"),
			want: false,
		},
		{
			name: "gRPC NotFound wrapped with apierror",
			err: func() error {
				grpcErr := status.Error(codes.NotFound, "not found")
				apiErr, _ := apierror.ParseError(grpcErr, true)
				return apiErr
			}(),
			want: true,
		},
		{
			name: "gRPC AlreadyExists wrapped with apierror",
			err: func() error {
				grpcErr := status.Error(codes.AlreadyExists, "already exists")
				apiErr, _ := apierror.ParseError(grpcErr, true)
				return apiErr
			}(),
			want: false,
		},
		{
			name: "wrapped gRPC NotFound",
			err:  fmt.Errorf("getting resource: %w", status.Error(codes.NotFound, "not found")),
			want: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsNotFound(tc.err)
			if got != tc.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

// Value and ListValue round-trip through JSON, which is the whole reason they
// map to apiextensionsv1.JSON rather than a generated struct: the struct form
// is recursive (a Value may hold a ListValue of Values) and controller-gen
// cannot build a terminating CRD schema for it.
//
// The cases below cover each arm of the Value union, including the two nested
// ones that make the recursion, so a regression that flattens or drops a nested
// value fails here rather than in a CRD somewhere.
func TestValue_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		in   any
	}{
		{name: "null", in: nil},
		{name: "number", in: float64(42)},
		{name: "string", in: "hello"},
		{name: "bool", in: true},
		{name: "struct", in: map[string]any{"a": "b", "n": float64(1)}},
		{name: "list", in: []any{"a", float64(1), true, nil}},
		{name: "nested", in: map[string]any{
			"list": []any{map[string]any{"deep": []any{float64(1)}}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb, err := structpb.NewValue(tt.in)
			if err != nil {
				t.Fatalf("building structpb.Value: %v", err)
			}

			mapCtx := &MapContext{}
			krm := Value_FromProto(mapCtx, pb)
			if err := mapCtx.Err(); err != nil {
				t.Fatalf("Value_FromProto: %v", err)
			}
			if krm == nil {
				t.Fatal("Value_FromProto returned nil")
			}

			back := Value_ToProto(mapCtx, krm)
			if err := mapCtx.Err(); err != nil {
				t.Fatalf("Value_ToProto: %v", err)
			}
			if !proto.Equal(pb, back) {
				t.Errorf("round trip changed the value.\n got: %v\nwant: %v", back, pb)
			}
		})
	}
}

// A nil input is not an error, it is an absent field, and both directions have
// to say so rather than producing an empty JSON document or an empty message.
func TestValue_Nil(t *testing.T) {
	mapCtx := &MapContext{}
	if got := Value_FromProto(mapCtx, nil); got != nil {
		t.Errorf("Value_FromProto(nil) = %v, want nil", got)
	}
	if got := Value_ToProto(mapCtx, nil); got != nil {
		t.Errorf("Value_ToProto(nil) = %v, want nil", got)
	}
	if got := ListValue_FromProto(mapCtx, nil); got != nil {
		t.Errorf("ListValue_FromProto(nil) = %v, want nil", got)
	}
	if got := ListValue_ToProto(mapCtx, nil); got != nil {
		t.Errorf("ListValue_ToProto(nil) = %v, want nil", got)
	}
	if err := mapCtx.Err(); err != nil {
		t.Errorf("nil inputs recorded an error: %v", err)
	}
}

func TestListValue_RoundTrip(t *testing.T) {
	pb, err := structpb.NewList([]any{
		"a", float64(1), true, nil,
		map[string]any{"k": "v"},
		[]any{float64(1), float64(2)},
	})
	if err != nil {
		t.Fatalf("building structpb.ListValue: %v", err)
	}

	mapCtx := &MapContext{}
	krm := ListValue_FromProto(mapCtx, pb)
	if err := mapCtx.Err(); err != nil {
		t.Fatalf("ListValue_FromProto: %v", err)
	}
	back := ListValue_ToProto(mapCtx, krm)
	if err := mapCtx.Err(); err != nil {
		t.Fatalf("ListValue_ToProto: %v", err)
	}
	if !proto.Equal(pb, back) {
		t.Errorf("round trip changed the list.\n got: %v\nwant: %v", back, pb)
	}
}

// An empty list is distinct from an absent one, and "[]" has to survive as an
// empty ListValue rather than collapsing to nil.
func TestListValue_Empty(t *testing.T) {
	mapCtx := &MapContext{}
	krm := ListValue_FromProto(mapCtx, &structpb.ListValue{})
	if krm == nil || string(krm.Raw) != "[]" {
		t.Fatalf("ListValue_FromProto(empty) = %v, want []", krm)
	}
	back := ListValue_ToProto(mapCtx, krm)
	if back == nil || len(back.Values) != 0 {
		t.Errorf("ListValue_ToProto([]) = %v, want an empty ListValue", back)
	}
	if err := mapCtx.Err(); err != nil {
		t.Errorf("empty list recorded an error: %v", err)
	}
}
