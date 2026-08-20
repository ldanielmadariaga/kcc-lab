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
	"context"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

const SCORE_DESCRIPTION = `Score a CRD against the version stored at a git ref.

Reports how closely the file on disk reproduces the baseline: field coverage,
type agreement, required markers, observed state, and what became of each
reference field. Unlike "equivalent" and "compatible" this never fails; it
measures, so that a generated CRD can be compared with a hand-written one.`

type ScoreOptions struct {
	File    string
	Ref     string
	Verbose bool
	Limit   int
}

func BuildScoreCmd() *cobra.Command {
	var opts ScoreOptions
	cmd := &cobra.Command{
		Use:     "score",
		Short:   "Score a CRD against its version at a git ref",
		Long:    SCORE_DESCRIPTION,
		Example: `crd-mcp-server score --file config/crds/resources/....yaml --ref HEAD~1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunScore(cmd.Context(), &opts)
		},
		Args: cobra.ExactArgs(0),
	}
	cmd.Flags().StringVar(&opts.File, "file", opts.File, "Path to the CRD YAML file to score.")
	cmd.MarkFlagRequired("file")
	cmd.Flags().StringVar(&opts.Ref, "ref", "HEAD", "Git ref holding the baseline version.")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "List the individual differing paths.")
	cmd.Flags().IntVar(&opts.Limit, "limit", 25, "Maximum paths to list per bucket when --verbose is set.")
	return cmd
}

func RunScore(ctx context.Context, opts *ScoreOptions) error {
	ourBytes, err := os.ReadFile(opts.File)
	if err != nil {
		return fmt.Errorf("reading %q: %w", opts.File, err)
	}
	baseBytes, isNew, err := gitShow(opts.Ref, opts.File)
	if err != nil {
		return err
	}
	if isNew {
		return fmt.Errorf("%q does not exist at ref %q, so there is nothing to score against", opts.File, opts.Ref)
	}

	ourCRD, err := parseCRD(ourBytes)
	if err != nil {
		return fmt.Errorf("parsing current CRD: %w", err)
	}
	baseCRD, err := parseCRD(baseBytes)
	if err != nil {
		return fmt.Errorf("parsing baseline CRD: %w", err)
	}

	baseVersions := getVersionSchemas(baseCRD)
	ourVersions := getVersionSchemas(ourCRD)

	fmt.Printf("%s  (baseline %s)\n", baseCRD.Spec.Names.Kind, opts.Ref)
	for _, v := range slices.Sorted(maps.Keys(baseVersions)) {
		ourPaths, ok := ourVersions[v]
		if !ok {
			fmt.Printf("  [%s] version absent from the current file\n", v)
			continue
		}
		printScore(scorePaths(baseVersions[v], ourPaths), v, opts)
	}
	return nil
}

func printScore(r *ScoreResult, version string, opts *ScoreOptions) {
	fmt.Printf("\n  [%s]\n", version)
	fmt.Printf("  %-22s %8s %8s %8s %8s %8s\n", "bucket", "matched", "missing", "extra", "mismatch", "rate")
	for _, b := range []*Bucket{&r.Spec, &r.Required, &r.ObservedState, &r.StatusOther} {
		fmt.Printf("  %-22s %8d %8d %8d %8d %7.1f%%\n",
			b.Name, b.Matched, len(b.Missing), len(b.Extra), len(b.Mismatch), 100*b.Rate())
	}

	if len(r.Refs) > 0 {
		same, plain, absent := 0, 0, 0
		for _, f := range r.Refs {
			switch {
			case f.Ours == "same":
				same++
			case strings.HasPrefix(f.Ours, "plain string"):
				plain++
			default:
				absent++
			}
		}
		fmt.Printf("  references: %d in baseline -> %d reproduced, %d left as plain strings, %d absent\n",
			len(r.Refs), same, plain, absent)
		for _, f := range r.Refs {
			if f.Ours != "same" {
				fmt.Printf("      %s -> %s\n", f.BaselinePath, f.Ours)
			}
		}
	}

	if !opts.Verbose {
		return
	}
	for _, b := range []*Bucket{&r.Spec, &r.Required, &r.ObservedState, &r.StatusOther} {
		printList(b.Name+" missing", b.Missing, opts.Limit)
		printList(b.Name+" extra", b.Extra, opts.Limit)
		printList(b.Name+" mismatch", b.Mismatch, opts.Limit)
	}
}

func printList(label string, items []string, limit int) {
	if len(items) == 0 {
		return
	}
	fmt.Printf("    %s (%d):\n", label, len(items))
	for i, s := range items {
		if i >= limit {
			fmt.Printf("      ... and %d more\n", len(items)-limit)
			break
		}
		fmt.Printf("      %s\n", s)
	}
}
