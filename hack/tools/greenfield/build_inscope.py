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

"""Derive the experiment corpus instead of maintaining it by hand.

inscope.tsv lists the resources the replication experiment deletes and
regenerates. It was hand-maintained and drifted: 39 greenfield kinds that had an
upstream CRD to compare against were simply absent, so the corpus was about 17%
under-scoped and every percentage was measured on an incomplete population.

A resource is in scope when all four hold:

  1. static_config.go does NOT list Terraform or DCL support for it. A kind that
     supports either is brownfield: its schema is years of hand-curation rather
     than something a generator was ever meant to produce, and scoring it
     measures migration debt, not generation.

     Absence from static_config.go is not disqualifying. Plenty of greenfield
     resources are implemented in apis/ before they are registered -- roughly
     90 of them -- and requiring an entry drops 96 resources that are exactly
     what the experiment is about.
  2. Some apis/<service>/generate.sh declares it with --resource.
  3. That invocation does not pass --skip-scaffold-files, which means the types
     file is hand-written and regeneration would not produce it.
  4. It had a CRD at the baseline ref. Without one there is nothing to compare
     against, so it can be generated but not scored.

Usage:
  python3 hack/tools/greenfield/build_inscope.py [--ref c1df0b9326] [--check]
"""

import argparse
import glob
import os
import re
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(
    os.path.dirname(os.path.abspath(__file__)))))
OUT = "docs/ai/experiments/data-greenfield/inscope.tsv"


def brownfield():
    """Kinds static_config.go says are backed by Terraform or DCL.

    These are the ones to exclude. Registered-as-Direct-only and
    not-registered-at-all are both in scope; only a TF or DCL controller makes a
    resource brownfield.
    """
    src = open(os.path.join(ROOT, "pkg/controller/resourceconfig/static_config.go")).read()
    out = {}
    pat = (r'Group:\s*"([\w.]+)\.cnrm\.cloud\.google\.com",\s*Kind:\s*"(\w+)"\}:\s*'
           r'\{DefaultController:[^,]+,\s*SupportedControllers:\s*\[\]k8s\.ReconcilerType\{([^}]*)\}')
    for m in re.finditer(pat, src):
        ctrls = set(re.findall(r"ReconcilerType(\w+)", m.group(3)))
        if ctrls - {"Direct"}:
            # Keyed lowercase: generate.sh and static_config.go disagree on
            # acronym casing, and ComputeTargetTcpProxy against
            # ComputeTargetTCPProxy was enough to slip a Terraform-backed
            # resource into the corpus.
            out[m.group(2).lower()] = m.group(1)
    return out


def scaffolded():
    """Kind -> service, for kinds a generate.sh actually scaffolds.

    Attributed per generate-types invocation, because --skip-scaffold-files is
    set on the invocation and a service can have several.
    """
    out = {}
    for f in glob.glob(os.path.join(ROOT, "apis/*/generate.sh")):
        svc = os.path.basename(os.path.dirname(f))
        for block in open(f).read().split("generate-types")[1:]:
            block = block.split("generate-mapper")[0]
            if "--skip-scaffold-files" in block:
                continue
            for kind in re.findall(r"--resource\s+(\w+):", block):
                out[kind] = svc
    return out


def baseline_crds(ref):
    """CRD path by plural, as of the baseline ref."""
    listing = subprocess.run(["git", "ls-tree", "-r", "--name-only", ref,
                              "config/crds/resources/"],
                             capture_output=True, text=True, cwd=ROOT).stdout
    out = {}
    for line in listing.split("\n"):
        m = re.search(r"_([a-z0-9]+)\.[a-z0-9]+\.cnrm\.cloud\.google\.com\.yaml$", line)
        if m:
            out[m.group(1)] = line
    return out


def plurals_of(kind):
    """The CRD plural forms to try, in order.

    Kubernetes pluralises the Kind, so the -y to -ies rule matters: fifteen
    resources (BigQueryDataPolicy, DataCatalogEntry, ComputeTargetTcpProxy and
    friends) were dropped from the corpus by a matcher that only tried "+s" and
    "+es".
    """
    k = kind.lower()
    out = [k, k + "s", k + "es"]
    if k.endswith("y") and not k.endswith(("ay", "ey", "iy", "oy", "uy")):
        out.insert(1, k[:-1] + "ies")
    if k.endswith(("s", "x", "z", "ch", "sh")):
        out.insert(1, k + "es")
    return out


def types_path(kind, service):
    """The scaffolded types file, whichever version directory holds it."""
    for version in ("v1alpha1", "v1beta1"):
        p = f"apis/{service}/{version}/{kind.lower()}_types.go"
        if os.path.exists(os.path.join(ROOT, p)):
            return p
    return f"apis/{service}/v1alpha1/{kind.lower()}_types.go"


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--ref", default="c1df0b9326")
    ap.add_argument("--check", action="store_true",
                    help="report the diff against the committed file, write nothing")
    args = ap.parse_args()

    brown, scaf, crds = brownfield(), scaffolded(), baseline_crds(args.ref)
    rows, no_baseline = [], []
    for kind, service in sorted(scaf.items()):
        if kind.lower() in brown:
            continue
        crd = None
        for cand in plurals_of(kind):
            if cand in crds:
                crd = crds[cand]
                break
        if crd is None:
            no_baseline.append(kind)
            continue
        rows.append((kind, service, types_path(kind, service), crd))

    path = os.path.join(ROOT, OUT)
    existing = {l.split("\t")[0] for l in open(path) if l.strip()} if os.path.exists(path) else set()
    now = {r[0] for r in rows}
    print(f"in scope: {len(rows)}   (was {len(existing)})")
    added, dropped = sorted(now - existing), sorted(existing - now)
    if added:
        print(f"  + {len(added)} added: {', '.join(added[:10])}{' …' if len(added) > 10 else ''}")
    if dropped:
        print(f"  - {len(dropped)} dropped: {', '.join(dropped[:10])}{' …' if len(dropped) > 10 else ''}")
    print(f"  {len(no_baseline)} greenfield kinds have no baseline CRD and cannot be scored: "
          f"{', '.join(no_baseline[:6])}")

    if args.check:
        return 1 if (added or dropped) else 0
    with open(path, "w") as fh:
        for r in rows:
            fh.write("\t".join(r) + "\n")
    print(f"\nwrote {OUT}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
