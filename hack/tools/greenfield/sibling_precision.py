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

"""Score the sibling-resource rule against what upstream actually modelled.

The rule marks a string field whose name matches a resource the same service
declares. This asks, for every field it marked, what upstream did with that
field: turn it into a reference, or leave it a plain string.

Precision is the only honest measure here. Recall would need a list of every
reference upstream has and a way to pair each one to a field we generate, which
is the same name-pairing problem silence_report.py documents at length. So this
answers the narrower question the rule has to get right: when it speaks, is it
worth a reviewer's time?

A field upstream never modelled is excluded rather than counted as wrong. The
rule cannot be wrong about a field nobody had an opinion on, and including them
would score the corpus rather than the rule.

Usage:
  python3 hack/tools/greenfield/sibling_precision.py [--ref c1df0b9326] [--list]
"""

import argparse
import collections
import glob
import os
import re
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(
    os.path.dirname(os.path.abspath(__file__)))))

MARKER = re.compile(r"\+kcc:guess=possible-reference target=(\S+)")
JSON_TAG = re.compile(r'json:"([A-Za-z0-9_]+)')


def marked_fields():
    """service -> {json name: target Kind} for every field the rule marked."""
    out = collections.defaultdict(dict)
    for path in glob.glob(os.path.join(ROOT, "apis", "*", "v1*", "*.go")):
        service = os.path.basename(os.path.dirname(os.path.dirname(path)))
        pending = None
        # Skip /* ... */ blocks, which dump what the generator would have written
        # for a hand-declared type. Nothing in them reaches the CRD, so scoring
        # them would measure output that does not exist.
        depth = 0
        for line in open(path, errors="ignore"):
            stripped = line.strip()
            if stripped.startswith("/*"):
                depth += 1
                continue
            if stripped.startswith("*/"):
                depth = max(0, depth - 1)
                continue
            if depth:
                continue
            if stripped.startswith("//"):
                m = MARKER.search(stripped)
                if m:
                    pending = m.group(1)
                continue
            m = JSON_TAG.search(line)
            if m:
                if pending:
                    out[service][m.group(1)] = pending
                pending = None
            elif stripped and not stripped.startswith("//"):
                pending = None
    return out


def baseline_crds(ref):
    """service -> the text of every upstream CRD for it, concatenated.

    Read from git rather than the working tree: the working tree is the
    regenerated output, which is the thing being scored.
    """
    listing = subprocess.run(["git", "ls-tree", "-r", "--name-only", ref,
                              "config/crds/resources/"],
                             capture_output=True, text=True, cwd=ROOT)
    by_service = collections.defaultdict(list)
    for f in listing.stdout.split("\n"):
        if not f.endswith(".yaml"):
            continue
        # apiextensions.k8s.io_v1_customresourcedefinition_<plural>.<group>.yaml
        # The group is separated by a dot, not the underscore that precedes the
        # plural -- the whole listing silently produced nothing when this said "_".
        m = re.search(r"\.([a-z0-9]+)\.cnrm\.cloud\.google\.com\.yaml$", f)
        if m:
            by_service[m.group(1)].append(f)
    out = {}
    for svc, files in by_service.items():
        blobs = []
        for f in files:
            r = subprocess.run(["git", "show", f"{ref}:{f}"],
                               capture_output=True, text=True, cwd=ROOT)
            if r.returncode == 0:
                blobs.append(r.stdout)
        out[svc] = "\n".join(blobs)
    return out


def verdict(name, crd_text):
    """What upstream did with a field of this name: 'ref', 'plain' or None.

    A KCC reference is an object carrying "external", and nothing else in a CRD
    has that, so the property name plus a nearby "external" is what separates a
    reference from a string that merely shares the name.
    """
    ref_names = {name + "Ref", name + "Refs", name}
    # Property blocks are indented YAML; find each occurrence of the name as a
    # property key and look at what follows it.
    plain = False
    for m in re.finditer(r"^(\s+)(%s):\s*$" % "|".join(re.escape(n) for n in ref_names),
                         crd_text, re.M):
        indent, key = m.group(1), m.group(2)
        block = crd_text[m.end(): m.end() + 4000]
        # Stop at the next key at the same or shallower indent.
        end = re.search(r"^\s{0,%d}\S" % (len(indent) - 1), block, re.M)
        if end:
            block = block[: end.start()]
        if "external:" in block:
            return "ref"
        if key == name:
            plain = True
    if plain:
        return "plain"
    return None


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("--ref", default="c1df0b9326", help="baseline git ref")
    ap.add_argument("--list", action="store_true")
    args = ap.parse_args()

    marks = marked_fields()
    if not marks:
        print("no +kcc:guess=possible-reference markers found -- has the tree "
              "been regenerated?")
        return 1
    crds = baseline_crds(args.ref)

    rows = []
    counts = collections.Counter()
    for svc, fields in sorted(marks.items()):
        text = crds.get(svc, "")
        for name, target in sorted(fields.items()):
            v = verdict(name, text) if text else None
            counts[v or "no opinion"] += 1
            rows.append((svc, name, target, v or "no opinion"))

    judged = counts["ref"] + counts["plain"]
    total = sum(counts.values())
    print(f"fields marked by the sibling rule: {total}")
    print(f"  upstream made it a reference     {counts['ref']:5d}")
    print(f"  upstream kept it a plain string  {counts['plain']:5d}")
    print(f"  upstream never modelled it       {counts['no opinion']:5d}"
          "   (excluded: the rule cannot be wrong about a field nobody modelled)")
    if judged:
        print(f"\nprecision over fields upstream modelled: "
              f"{100 * counts['ref'] / judged:.0f}%  ({counts['ref']}/{judged})")
    if args.list:
        print()
        for svc, name, target, v in sorted(rows, key=lambda r: (r[3], r[0], r[1])):
            print(f"  {v:12s} {svc:24s} {name:32s} -> {target}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
