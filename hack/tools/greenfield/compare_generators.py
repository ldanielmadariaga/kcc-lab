# Copyright 2026 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Compare two generators' CRD output on the kinds both produced.

Takes two --verbose-dir caches written by silence_report.py and reports, for the
kinds present in both, what fraction of the baseline's fields each side
reproduces at the same path.

The denominator is `matched + missing + mismatch`, read from the scorer's own
per-bucket rows. That is the baseline's field count, so it is a property of
upstream rather than of either generator, and the two sides must agree on it.
The script asserts they agree within a tolerance and fails loudly otherwise,
because a comparison whose denominators disagree is measuring two things.

Why not silence_report's own denominator: that one runs missing paths through
roots(), which collapses every child of a missing parent into a single defect.
That is the right unit for a work list -- one thing to go and fix -- but it
cannot compare two generators. A side that misses one fifty-field subtree scores
one defect where a side that misses twenty scattered leaves scores twenty. On
this corpus that reversal is not hypothetical: it flips the ordering.

Usage:
  python3 hack/tools/greenfield/compare_generators.py \\
      --a /tmp/score_ours --a-label Claude \\
      --b /tmp/score_theirs --b-label Gemini \\
      [--only kinds.txt] [--split flagless.txt --split-label "flag off"]
"""

import argparse
import os
import re
import sys

ROW = re.compile(r"\s+(spec|required|status\.observedState)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s")


def read(cache, kind):
    """(matched, baseline total) for one kind, or None if not scored.

    'required' is skipped: it is a count of entries in the CRD's required list,
    not of fields, and including it double-counts fields that are also required.
    """
    p = os.path.join(cache, kind + ".txt")
    if not os.path.exists(p):
        return None
    matched = total = 0
    for line in open(p):
        m = ROW.match(line)
        if m and m.group(1) != "required":
            matched += int(m.group(2))
            total += int(m.group(2)) + int(m.group(3)) + int(m.group(5))
    return (matched, total) if total else None


def aggregate(cache, kinds):
    m = t = 0
    for k in kinds:
        r = read(cache, k)
        if r:
            m += r[0]
            t += r[1]
    return m, t


def line(label, m, t):
    return f"  {label:26s} {m:6d} / {t:6d} = {100 * m / t:5.1f}%" if t else f"  {label:26s} (nothing scored)"


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--a", required=True)
    ap.add_argument("--b", required=True)
    ap.add_argument("--a-label", default="A")
    ap.add_argument("--b-label", default="B")
    ap.add_argument("--only", help="file of kinds to restrict to, one per line")
    ap.add_argument("--split", help="file of kinds forming a second group, reported apart")
    ap.add_argument("--split-label", default="split group")
    ap.add_argument("--tolerance", type=float, default=2.0,
                    help="max %% disagreement between the two denominators (default 2)")
    ap.add_argument("--outliers", type=int, default=6)
    args = ap.parse_args()

    shared = {f[:-4] for f in os.listdir(args.a)} & {f[:-4] for f in os.listdir(args.b)}
    if args.only:
        shared &= {l.strip() for l in open(args.only) if l.strip()}
    shared = {k for k in shared if read(args.a, k) and read(args.b, k)}
    if not shared:
        print("no kinds scored by both", file=sys.stderr)
        return 1

    split = {l.strip() for l in open(args.split)} & shared if args.split else set()
    groups = [("all shared kinds", shared)]
    if split:
        groups = [("where both generate", shared - split), (args.split_label, split),
                  ("all shared kinds", shared)]

    print(f"{len(shared)} kinds scored by both\n")
    for name, ks in groups:
        am, at = aggregate(args.a, ks)
        bm, bt = aggregate(args.b, ks)
        drift = 100 * abs(at - bt) / max(at, bt, 1)
        print(f"{name}  ({len(ks)} kinds)")
        print(line(args.a_label, am, at))
        print(line(args.b_label, bm, bt))
        flag = "" if drift <= args.tolerance else "   <-- ABOVE TOLERANCE, comparison unsafe"
        print(f"  {'denominators differ by':26s} {drift:5.2f}%{flag}\n")
        if drift > args.tolerance:
            print("The two sides are not measuring the same baseline. Do not quote these numbers.",
                  file=sys.stderr)

    diffs = []
    for k in shared:
        (am, at), (bm, bt) = read(args.a, k), read(args.b, k)
        diffs.append((am / at - bm / bt, k, am, at, bm, bt))
    diffs.sort()
    print(f"largest per-kind gaps (aggregates hide a lot; the spread here is wide)")
    for d, k, am, at, bm, bt in diffs[:args.outliers]:
        print(f"  {args.b_label} ahead  {k:34s} {args.a_label} {am:4d}/{at:<4d}  {args.b_label} {bm:4d}/{bt:<4d}")
    for d, k, am, at, bm, bt in diffs[-args.outliers:][::-1]:
        print(f"  {args.a_label} ahead  {k:34s} {args.a_label} {am:4d}/{at:<4d}  {args.b_label} {bm:4d}/{bt:<4d}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
