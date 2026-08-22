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

"""Report how much of the gap between generated CRDs and a baseline is silent.

A field we do not generate is acceptable. A field we do not generate and do not
mention is not: the resource looks finished, passes its checks, and is quietly
missing something. This tool splits every missing baseline field three ways.

  generated  the field is in our CRD
  explained  it is not, and apis/<svc>/needs_judgement_call.txt names it
  silent     it is not, and nothing says so

The number to drive to zero is silent. "Generated" is not the target; a field a
human must decide is a fine outcome as long as somebody is told.

Three details decide whether the figure means anything, each of which produced a
wrong answer before it was fixed:

  * Report roots, not paths. When a parent is missing its whole subtree is
    reported missing too, which inflates every count several-fold. 298 missing
    observedState paths were 156 real defects.

  * Pair names across the reference rename. The queue names a field as we
    generated it (".spec.pipelineJob"); the baseline names it as upstream has it
    (".spec.pipelineJobRef.external"). Comparing literally scores 0% when the
    true figure is 5%.

  * Ignore the blanket queue entry. "untriaged-bulk-generation" names no field
    and covers everything trivially, which would report 100% on day one.

Usage:
  python3 hack/tools/greenfield/silence_report.py \
      --resources docs/ai/experiments/data-greenfield/inscope.tsv \
      --ref c1df0b9326 [--verbose-dir DIR] [--only FILE] [--list-silent]

--resources is a TSV of kind, service, types path, CRD path. --verbose-dir
caches `crd-mcp-server score --verbose` output; scoring 200 resources takes a
few minutes, so reuse it while iterating on the generator.
"""

import argparse
import glob
import os
import re
import subprocess
import sys
from collections import Counter, defaultdict

# A reference path carries a <thing>Ref or <thing>Refs segment, optionally an
# array. That covers the object and each of its .external/.name/.namespace/.kind
# children, which is why an unreproduced reference costs four or five paths.
REF_SEGMENT = re.compile(r"(^|\.)[A-Za-z0-9_]+Refs?(\[\])?(\.|$)")
REF_CHILD = re.compile(r"\.(external|name|namespace|kind)$")
REF_SUFFIX = re.compile(r"Refs?(\[\])?$")

BUCKETS = ("spec-reference", "spec-other", "observedState")


def score_resource(binary, crd, ref):
    out = subprocess.run(
        [binary, "score", "--file", crd, "--ref", ref, "--verbose", "--limit", "900"],
        capture_output=True, text=True)
    return out.stdout if out.returncode == 0 else None


def parse_score(text):
    """Return (matched-per-bucket, missing paths) from one score report."""
    matched = Counter()
    missing = {"spec": [], "status.observedState": []}
    section = None
    for line in text.split("\n"):
        m = re.match(r"\s+(spec|required|status\.observedState)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s", line)
        if m:
            matched[m.group(1)] += int(m.group(2))
            continue
        m = re.match(r"\s+(spec|status\.observedState) missing \((\d+)\):", line)
        if m:
            section = m.group(1)
            continue
        if re.match(r"\s+\S+ (missing|extra|mismatch)", line):
            section = None
            continue
        if section and line.startswith("      "):
            missing[section].append(line.strip().split(" ")[0])
    return matched, missing


def roots(paths):
    """Drop any path whose parent is also missing; the parent explains it."""
    present = set(paths)
    out = []
    for p in paths:
        parent = p.rsplit(".", 1)[0] if "." in p else ""
        if parent and (parent in present or parent + "[]" in present):
            continue
        out.append(p)
    return out


def bucket_of(section, path):
    if section == "status.observedState":
        return "observedState"
    return "spec-reference" if REF_SEGMENT.search(path) else "spec-other"


def queue_entries():
    """Field-level judgement entries, keyed by kind. Blanket entries excluded."""
    q = defaultdict(set)
    for f in glob.glob("apis/*/needs_judgement_call.txt"):
        for line in open(f):
            m = re.match(r'kind=(\S+) group=\S+: field "([^"]+)"', line)
            if m:
                q[m.group(1)].add(m.group(2).lstrip("."))
    return q


def explained(entries, path):
    """Does any queue entry name this field, allowing for the reference rename?"""
    plain = REF_SUFFIX.sub(r"\1", REF_CHILD.sub("", path))
    for want in {path.lstrip("."), plain.lstrip(".")}:
        for e in entries:
            base = e.rstrip("[]")
            if want == e or want.startswith(base + ".") or want.startswith(base + "[]"):
                return True
    return False


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--resources", required=True, help="TSV: kind, service, types path, CRD path")
    ap.add_argument("--ref", default="c1df0b9326", help="baseline git ref")
    ap.add_argument("--verbose-dir", help="cache dir for score --verbose output")
    ap.add_argument("--only", help="file of kinds to restrict to, one per line")
    ap.add_argument("--binary", default="./bin/crd-mcp-server")
    ap.add_argument("--list-silent", action="store_true", help="print every silent root")
    args = ap.parse_args()

    only = None
    if args.only:
        only = {l.strip() for l in open(args.only) if l.strip()}
    if args.verbose_dir:
        os.makedirs(args.verbose_dir, exist_ok=True)

    q = queue_entries()
    matched = Counter()
    gap = {b: Counter() for b in BUCKETS}     # explained / silent
    silent_list = defaultdict(list)
    n = skipped = 0

    for line in open(args.resources):
        if not line.strip():
            continue
        kind, svc, _types, crd = line.rstrip("\n").split("\t")[:4]
        if only is not None and kind not in only:
            continue
        cached = os.path.join(args.verbose_dir, kind + ".txt") if args.verbose_dir else None
        text = None
        if cached and os.path.exists(cached):
            text = open(cached).read()
        else:
            if not os.path.exists(crd):
                skipped += 1
                continue
            text = score_resource(args.binary, crd, args.ref)
            if text is None:
                skipped += 1
                continue
            if cached:
                open(cached, "w").write(text)
        n += 1
        m, missing = parse_score(text)
        matched.update(m)
        for section, paths in missing.items():
            for p in roots(paths):
                b = bucket_of(section, p)
                if explained(q.get(kind, ()), p):
                    gap[b]["explained"] += 1
                else:
                    gap[b]["silent"] += 1
                    silent_list[b].append((kind, p))

    print(f"resources scored: {n}" + (f"   skipped: {skipped}" if skipped else ""))
    print(f"baseline: {args.ref}\n")
    # Fields we do produce, per section. There is no separate figure for
    # references within spec, so this is reported by section rather than by
    # bucket -- attributing it to one bucket would read as if it belonged there.
    print("generated fields matching the baseline")
    print(f"  spec                 {matched['spec']}")
    print(f"  required             {matched['required']}")
    print(f"  status.observedState {matched['status.observedState']}")

    print(f"\nmissing, as roots (a missing parent explains its own subtree)\n")
    print(f"{'bucket':18s} {'explained':>10s} {'silent':>8s} {'silent %':>9s}")
    tot_e = tot_s = 0
    for b in BUCKETS:
        e, s = gap[b]["explained"], gap[b]["silent"]
        tot_e += e
        tot_s += s
        pct = f"{100*s/(e+s):8.0f}%" if e + s else " " * 9
        print(f"  {b:16s} {e:10d} {s:8d} {pct}")
    total = tot_e + tot_s
    pct = f"{100*tot_s/total:8.0f}%" if total else ""
    print(f"  {'TOTAL':16s} {tot_e:10d} {tot_s:8d} {pct}")

    if args.list_silent:
        for b in BUCKETS:
            if not silent_list[b]:
                continue
            print(f"\n### silent, {b} ({len(silent_list[b])})")
            for kind, p in sorted(silent_list[b]):
                print(f"  {kind}\t{p}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
