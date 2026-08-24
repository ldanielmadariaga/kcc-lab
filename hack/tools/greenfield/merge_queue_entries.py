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

"""Merge judgement-queue lines from a scratch generation into the real queue.

Companion to regenerate_into_scratch.py. That writes a whole tree somewhere
harmless; this takes only the queue lines you name and appends them to
apis/<service>/needs_judgement_call.txt, leaving every types file alone. It is
how a change to what the generator records reaches the corpus without a
destructive regeneration.

Two guards, both of which matter:

  * only the reasons named on the command line are copied, so a run cannot
    quietly import unrelated changes from the scratch generator;
  * only a Kind that already has a resource-level entry is touched. A queued
    resource contributes no [refs] findings, so queueing one that nobody is
    generating makes its existing missingrefs.txt entries read as fixed and get
    pruned -- they then return as fresh violations later. scripts/queue-hints
    gates on the same rule for the same reason.

Usage:
  python3 hack/tools/greenfield/merge_queue_entries.py <scratch-dir> \\
      --reasons parent-segment-omitted,identity-field-omitted [--apply]

Without --apply it reports what it would write.
"""

import argparse
import collections
import glob
import os
import re

# .../hack/tools/greenfield/<this file> -> the repo root is four levels up.
ROOT = os.path.dirname(os.path.dirname(os.path.dirname(
    os.path.dirname(os.path.abspath(__file__)))))
LINE = re.compile(r'kind=(\S+) group=(\S+): field "([^"]+)" reason=([a-z-]+)')


def parse(path):
    """Return (kind -> {(field, reason)}, kinds mentioned at all)."""
    entries = collections.defaultdict(set)
    kinds = set()
    if not os.path.exists(path):
        return entries, kinds
    for line in open(path):
        m = re.match(r"kind=(\S+) ", line)
        if m:
            kinds.add(m.group(1))
        m = LINE.match(line)
        if m:
            entries[m.group(1)].add((m.group(3), m.group(4)))
    return entries, kinds


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("scratch", help="directory written by regenerate_into_scratch.py")
    ap.add_argument("--reasons", required=True, help="comma-separated reason= values to copy")
    ap.add_argument("--apply", action="store_true", help="write, rather than report")
    args = ap.parse_args()
    reasons = tuple(r.strip() for r in args.reasons.split(",") if r.strip())

    added = collections.Counter()
    for src in sorted(glob.glob(os.path.join(args.scratch, "*", "needs_judgement_call.txt"))):
        service = os.path.basename(os.path.dirname(src))
        real = os.path.join(ROOT, "apis", service, "needs_judgement_call.txt")
        if not os.path.exists(real):
            continue
        have, queued = parse(real)
        new = []
        for line in open(src):
            m = LINE.match(line)
            if not m or m.group(4) not in reasons:
                continue
            kind, field, reason = m.group(1), m.group(3), m.group(4)
            if kind not in queued:
                continue
            if (field, reason) in have[kind]:
                continue
            # A field already named for one of these reasons is covered.
            if any(f == field and r in reasons for f, r in have[kind]):
                continue
            have[kind].add((field, reason))
            new.append(line.rstrip("\n"))
        if not new:
            continue
        added[service] = len(new)
        if args.apply:
            body = open(real).read()
            if not body.endswith("\n"):
                body += "\n"
            open(real, "w").write(body + "\n".join(new) + "\n")

    print(f"{sum(added.values())} line(s) across {len(added)} service(s)"
          + ("" if args.apply else "  [report only, pass --apply]"))
    for service, n in added.most_common(15):
        print(f"  {service}: {n}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
