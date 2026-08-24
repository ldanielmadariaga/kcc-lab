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

"""Re-run generate-types for every service into a throwaway tree.

Why this exists: a change to what the generator writes into
apis/<service>/needs_judgement_call.txt normally only takes effect on a real
regeneration, and a real regeneration overwrites the types files the whole
coverage measurement is taken against. --output-api redirects everything,
including the queue path, so each service can be generated in isolation -- about
1.4s each -- and only the queue lines merged back. No types file is touched.

Pair with merge_queue_entries.py, which does the merging.

Two things to expect in the log:

  * every invocation reports a failure at the prune step ("no packages found"),
    because a bare scratch tree has no Go packages to prune. Harmless -- the
    queue is written first -- and --prune-unused-types=false is passed to avoid
    it.
  * about a dozen services fail earlier with "proto: not found". Their
    generate.sh supplies an overlay or a specific --proto-source-path that this
    re-invocation does not reconstruct. Those services simply produce nothing.

Usage:
  python3 hack/tools/greenfield/regenerate_into_scratch.py <output-dir>
"""

import glob
import os
import shlex
import subprocess
import sys

# .../hack/tools/greenfield/<this file> -> the repo root is four levels up.
ROOT = os.path.dirname(os.path.dirname(os.path.dirname(
    os.path.dirname(os.path.abspath(__file__)))))


def invocations(path):
    """Yield the --service/--api-version/--resource flags of each generate-types
    call in a generate.sh, with line continuations joined."""
    body = open(path).read().replace("\\\n", " ")
    for line in body.split("\n"):
        if "generate-types" not in line or line.strip().startswith("#"):
            continue
        try:
            toks = shlex.split(line)
        except ValueError:
            continue
        if "generate-types" not in toks:
            continue
        args, j = [], toks.index("generate-types") + 1
        while j < len(toks):
            if toks[j] in ("--service", "--api-version", "--resource") and j + 1 < len(toks):
                args += [toks[j], toks[j + 1]]
                j += 2
                continue
            j += 1
        if "--service" in args and "--resource" in args:
            yield args


def main():
    if len(sys.argv) != 2:
        print(__doc__)
        return 2
    out = sys.argv[1]
    binary = os.path.join(ROOT, "bin", "controllerbuilder")
    if not os.path.exists(binary):
        print(f"build it first: go build -o {binary} ./dev/tools/controllerbuilder")
        return 1
    os.makedirs(out, exist_ok=True)

    ok = failed = 0
    for gs in sorted(glob.glob(os.path.join(ROOT, "apis", "*", "generate.sh"))):
        service = os.path.basename(os.path.dirname(gs))
        for args in invocations(gs):
            cmd = [binary, "generate-types", "--prepopulate-spec",
                   "--prune-unused-types=false", "--output-api", out] + args
            r = subprocess.run(cmd, capture_output=True, text=True, timeout=600)
            if r.returncode == 0:
                ok += 1
            else:
                failed += 1
                last = (r.stderr.strip().splitlines() or ["?"])[-1]
                print(f"FAIL {service}: {last[:120]}")
    print(f"invocations ok={ok} failed={failed}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
