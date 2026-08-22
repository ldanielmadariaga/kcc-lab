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

"""Check that every field KCC master has is either produced or flagged.

The target is CRD compatibility with k8s-config-connector master. If master's CRD
has spec.pipelineJobRef.external and we produce spec.pipelineJob, a user's YAML
breaks -- it does not help that both come from the same proto field. Renamed,
moved and reference-shaped fields are all fields KCC has and we do not.

Every field in the baseline CRD is one of:

  produced   we emit it too
  flagged    we do not, and apis/<svc>/needs_judgement_call.txt names it
  unflagged  we do not, and nothing says so

Unflagged is the number to drive to zero. A field a human must decide is a fine
outcome, as long as somebody is told about it.

Each field we miss is also classified by *why* it differs, to route the fix. All
five classes are real gaps; they just need different work:

  reference-shape        we emit a plain string, KCC has a Ref object
  renamed                same field, different name (bootDiskMIB/bootDiskMiB)
  moved                  we emit it in Spec, KCC has it in status.observedState
  intentionally-different  we model it deliberately otherwise, e.g. Value as JSON
  absent                 it appears nowhere in our output

A note on a number this tool does not report: comparing +kcc:proto:field
annotations instead of CRD paths says the generator emits 16295 of the 16438
proto fields KCC declares. That is a useful check on whether generation is
healthy, and it is not this measurement -- using it to shrink the CRD gap
explains away fields users actually cannot set.

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
    """Return (matched-per-bucket, missing paths, extra paths) from one report."""
    matched = Counter()
    missing = {"spec": [], "status.observedState": []}
    extra = {"spec": [], "status.observedState": []}
    section = kind = None
    for line in text.split("\n"):
        m = re.match(r"\s+(spec|required|status\.observedState)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s", line)
        if m:
            matched[m.group(1)] += int(m.group(2))
            continue
        m = re.match(r"\s+(spec|status\.observedState) (missing|extra) \((\d+)\):", line)
        if m:
            section, kind = m.group(1), m.group(2)
            continue
        if re.match(r"\s+\S+ (missing|extra|mismatch)", line):
            section = kind = None
            continue
        if section and line.startswith("      "):
            path = line.strip().split(" ")[0]
            (missing if kind == "missing" else extra)[section].append(path)
    return matched, missing, extra


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


# The arms of google.protobuf.Value. We map that message to apiextensionsv1.JSON
# deliberately, so a baseline CRD modelling it as a union struct differs from ours
# by choice. Still a CRD difference; not a generation failure.
VALUE_ARMS = {"nullValue", "numberValue", "boolValue", "stringValue",
              "structValue", "listValue", "values"}

CLASSES = ("reference-shape", "renamed", "moved", "intentionally-different", "absent")


def strip_section(path):
    """Drop the leading spec. / status.observedState. so the two can be compared."""
    for prefix in ("status.observedState.", "spec."):
        if path.startswith(prefix):
            return path[len(prefix):]
    return path


def classify(path, section, extras):
    """Say why a baseline field is not in our CRD.

    extras maps section -> the paths we emit that the baseline does not have, which
    is what makes "moved" and "renamed" visible: a field we put somewhere else shows
    up as missing in one place and extra in another.
    """
    if REF_SEGMENT.search(path):
        return "reference-shape"
    leaf = path.rsplit(".", 1)[-1].rstrip("[]")
    if leaf in VALUE_ARMS:
        return "intentionally-different"

    rel = strip_section(path)
    other = "spec" if section == "status.observedState" else "status.observedState"
    if any(strip_section(e) == rel for e in extras.get(other, ())):
        return "moved"

    # Same parent, same name but for case: an acronym the generator cased
    # differently, e.g. we write bootDiskMIB where KCC writes bootDiskMiB.
    parent = path.rsplit(".", 1)[0] if "." in path else ""
    for e in extras.get(section, ()):
        e_parent = e.rsplit(".", 1)[0] if "." in e else ""
        if e_parent == parent and e.rsplit(".", 1)[-1].lower() == leaf.lower():
            return "renamed"
    return "absent"


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
    ap.add_argument("--list-silent", action="store_true", help="print every field missed without a flag")
    args = ap.parse_args()

    only = None
    if args.only:
        only = {l.strip() for l in open(args.only) if l.strip()}
    if args.verbose_dir:
        os.makedirs(args.verbose_dir, exist_ok=True)

    q = queue_entries()
    matched = Counter()
    gap = {c: Counter() for c in CLASSES}     # flagged / unflagged, per class
    unflagged_list = defaultdict(list)
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
        m, missing, extra = parse_score(text)
        matched.update(m)
        for section, paths in missing.items():
            for p in roots(paths):
                c = classify(p, section, extra)
                if explained(q.get(kind, ()), p):
                    gap[c]["flagged"] += 1
                else:
                    gap[c]["unflagged"] += 1
                    unflagged_list[c].append((kind, p))

    print(f"resources scored: {n}" + (f"   skipped: {skipped}" if skipped else ""))
    print(f"baseline: {args.ref}\n")
    produced = matched["spec"] + matched["required"] + matched["status.observedState"]
    missing_total = sum(gap[c]["flagged"] + gap[c]["unflagged"] for c in CLASSES)
    flagged_total = sum(gap[c]["flagged"] for c in CLASSES)
    unflagged_total = missing_total - flagged_total
    surface = produced + missing_total

    print(f"fields in KCC master's CRDs      {surface}")
    print(f"  we produce                     {produced}"
          f"   ({100 * produced / surface:.1f}%)")
    print(f"  we miss                        {missing_total}"
          f"   ({100 * missing_total / surface:.1f}%)")
    print()
    print(f"Of the {missing_total} we miss, how many does the queue name?\n")
    print(f"{'why it differs':26s} {'we miss':>8s} {'flagged':>8s} {'unflagged':>10s}")
    for c in CLASSES:
        f, u = gap[c]["flagged"], gap[c]["unflagged"]
        if f + u == 0:
            continue
        print(f"  {c:24s} {f + u:8d} {f:8d} {u:10d}")
    print(f"  {'TOTAL':24s} {missing_total:8d} {flagged_total:8d} {unflagged_total:10d}")
    print(f"\n{unflagged_total} fields we miss without flagging. That is the number to drive to")
    print(f"zero -- a share of the {missing_total} we miss, not of the {surface} KCC has.")
    print("Watch \"we produce\" alongside it: a change that flags fields by no longer")
    print("producing them improves this report and takes working fields away.")

    if args.list_silent:
        for c in CLASSES:
            if not unflagged_list[c]:
                continue
            print(f"\n### missed without flagging, {c} ({len(unflagged_list[c])})")
            for kind, p in sorted(unflagged_list[c]):
                print(f"  {kind}\t{p}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
