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

package main

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Scoring answers a different question from compareEquivalence. That one asks
// whether a change is backward compatible and returns pass/fail. This one asks
// how closely a regenerated CRD reproduces a hand-written one, and returns
// counts, so a generator can be measured rather than gated.

// Bucket is one comparable slice of the schema, scored on its own.
type Bucket struct {
	Name     string
	Matched  int
	Missing  []string // in the baseline, absent from ours
	Extra    []string // in ours, absent from the baseline
	Mismatch []string // same path, different type
}

func (b *Bucket) total() int { return b.Matched + len(b.Missing) + len(b.Mismatch) }

// Rate is the share of baseline paths we reproduced with the same type.
func (b *Bucket) Rate() float64 {
	if b.total() == 0 {
		return 1
	}
	return float64(b.Matched) / float64(b.total())
}

// RefFinding records one reference field the baseline has, and what we put in
// its place. A generated resource emits references as plain strings, so the
// interesting question is not whether the path matches but whether the field
// survived at all under a different name and type.
type RefFinding struct {
	BaselinePath string
	Ours         string // "same", "plain string at <path>", or "absent"
}

// ScoreResult is one version of one CRD, scored.
type ScoreResult struct {
	Version       string
	Spec          Bucket
	Required      Bucket
	ObservedState Bucket
	StatusOther   Bucket
	Refs          []RefFinding
}

// classify assigns a flattened path to the bucket it should be scored in.
// Required markers are separated from fields because they answer a different
// question: whether we agree on optionality, not on the field existing.
func classify(path string) string {
	switch {
	case strings.Contains(path, ":required."):
		return "required"
	case strings.Contains(path, ":validation["):
		return "validation"
	case strings.HasPrefix(path, "status.observedState"):
		return "observedState"
	case isUnderStatus(path):
		return "statusOther"
	case path == "spec" || strings.HasPrefix(path, "spec"):
		return "spec"
	default:
		return "other"
	}
}

// scorePaths compares one flattened schema against another.
func scorePaths(baseline, ours map[string]string) *ScoreResult {
	r := &ScoreResult{
		Spec:          Bucket{Name: "spec"},
		Required:      Bucket{Name: "required"},
		ObservedState: Bucket{Name: "status.observedState"},
		StatusOther:   Bucket{Name: "status (other)"},
	}
	bucketFor := func(path string) *Bucket {
		switch classify(path) {
		case "required":
			return &r.Required
		case "observedState":
			return &r.ObservedState
		case "statusOther":
			return &r.StatusOther
		case "spec":
			return &r.Spec
		}
		return nil // validation and metadata are not scored
	}

	for _, path := range slices.Sorted(maps.Keys(baseline)) {
		b := bucketFor(path)
		if b == nil {
			continue
		}
		ourType, ok := ours[path]
		switch {
		case !ok:
			b.Missing = append(b.Missing, fmt.Sprintf("%s (%s)", path, baseline[path]))
		case ourType != baseline[path]:
			b.Mismatch = append(b.Mismatch, fmt.Sprintf("%s: want %s, got %s", path, baseline[path], ourType))
		default:
			b.Matched++
		}
	}
	for _, path := range slices.Sorted(maps.Keys(ours)) {
		b := bucketFor(path)
		if b == nil {
			continue
		}
		if _, ok := baseline[path]; !ok {
			b.Extra = append(b.Extra, fmt.Sprintf("%s (%s)", path, ours[path]))
		}
	}

	r.Refs = findRefs(baseline, ours)
	return r
}

// findRefs locates every reference field in the baseline and reports what we
// produced instead. A KCC reference is an object carrying an "external" child,
// named with a Ref or Refs suffix; the mechanically generated equivalent is the
// same field without the suffix, typed as a string.
func findRefs(baseline, ours map[string]string) []RefFinding {
	var out []RefFinding
	for _, path := range slices.Sorted(maps.Keys(baseline)) {
		if _, ok := baseline[path+".external"]; !ok {
			continue
		}
		f := RefFinding{BaselinePath: path}
		switch {
		case ours[path] == baseline[path]:
			f.Ours = "same"
		default:
			if plain, ok := plainCounterpart(path, ours); ok {
				f.Ours = "plain string at " + plain
			} else {
				f.Ours = "absent"
			}
		}
		out = append(out, f)
	}
	return out
}

// plainCounterpart strips a Ref or Refs suffix from the last path segment and
// reports whether we emitted a field there.
func plainCounterpart(path string, ours map[string]string) (string, bool) {
	base := strings.TrimSuffix(path, "[]")
	i := strings.LastIndex(base, ".")
	if i < 0 {
		return "", false
	}
	head, last := base[:i+1], base[i+1:]
	for _, suffix := range []string{"Refs", "Ref"} {
		if !strings.HasSuffix(last, suffix) {
			continue
		}
		stem := strings.TrimSuffix(last, suffix)
		if stem == "" {
			continue
		}
		for _, cand := range []string{head + stem, head + stem + "[]", head + stem + "s", head + stem + "s[]"} {
			if _, ok := ours[cand]; ok {
				return cand, true
			}
		}
	}
	return "", false
}
