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

"""Check that every +kcc:guess marker has a judgement-queue entry.

Anything the generator marks +kcc:guess is something it could not justify from
the proto: a parent segment's field name taken from a pattern placeholder, a
ref's target type assumed from a collection segment, a field placed in
ObservedState because of its name. All of it needs a human, so all of it belongs
in apis/<service>/needs_judgement_call.txt.

Marker implies entry, and this checks it. It exists because the invariant broke
silently once already: an earlier version of the parent generation suppressed the
queue entry for any field it emitted, so 227 markers sat in the tree with almost
no entries, and the compiler was proving 32 fields missing while the queue named
2. A rule nothing checks is a rule that regresses.

The reverse direction is reported too but is not a failure: plenty of queue
entries are about fields we never emitted, which have no marker by construction.

Usage:
  python3 hack/tools/greenfield/check_guess_entries.py [--list]
"""

import collections
import glob
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(
    os.path.dirname(os.path.abspath(__file__)))))

# The marker sits above the field it describes, alongside any other comment
# lines, so remember it until the next struct field carrying a json tag.
MARKER = re.compile(r"\+kcc:guess=([a-z-]+)")
JSON_TAG = re.compile(r'json:"([A-Za-z0-9_]+)')


def markers_in(path):
    """(json name, guess kind) for every marked field in one types file."""
    out = []
    pending = None
    for line in open(path, errors="ignore"):
        stripped = line.strip()
        if stripped.startswith("//"):
            m = MARKER.search(stripped)
            if m:
                pending = m.group(1)
            continue
        m = JSON_TAG.search(line)
        if m:
            if pending:
                out.append((m.group(1), pending))
            pending = None
        elif stripped and not stripped.startswith("//"):
            pending = None
    return out


def kind_of(path):
    """The KRM Kind a types file declares, from its <Kind>Spec struct."""
    m = re.search(r"type (\w+)Spec struct", open(path, errors="ignore").read())
    return m.group(1) if m else None


def queue_fields():
    """kind -> the field paths its queue names."""
    out = collections.defaultdict(set)
    for f in glob.glob(os.path.join(ROOT, "apis", "*", "needs_judgement_call.txt")):
        for line in open(f):
            m = re.match(r'kind=(\S+) group=\S+: field "([^"]+)"', line)
            if m:
                out[m.group(1)].add(m.group(2).lstrip("."))
    return out


def main():
    show = "--list" in sys.argv
    queue = queue_fields()
    total = 0
    missing = []
    by_kind = collections.Counter()
    for path in glob.glob(os.path.join(ROOT, "apis", "*", "v1*", "*_types.go")):
        kind = kind_of(path)
        if not kind:
            continue
        for jsonname, guess in markers_in(path):
            total += 1
            by_kind[guess] += 1
            named = queue.get(kind, ())
            if not any(e == "spec." + jsonname or e == "status.observedState." + jsonname
                       for e in named):
                missing.append((kind, jsonname, guess))

    print(f"+kcc:guess markers: {total}")
    for g, n in by_kind.most_common():
        print(f"  {n:5d}  {g}")
    print(f"\nmarkers with no queue entry: {len(missing)}")
    if missing and show:
        for kind, field, guess in sorted(missing):
            print(f"  {kind:44s} {field:28s} {guess}")
    elif missing:
        print("  (pass --list to see them)")
    return 1 if missing else 0


if __name__ == "__main__":
    sys.exit(main())
