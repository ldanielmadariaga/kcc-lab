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

The report answers two questions, and they are independent. Reading them as one
is what made an earlier version of this output self-contradictory.

  why does the field differ?     five classes, the rows
  was anybody told about it?     three columns

A field can be "absent" and flagged at the same time. That is not a contradiction,
it is the queue doing its job: we did not produce the field, and the queue says so,
so a human will see it. The columns are:

  field-flagged     a queue entry names this exact field
  section-flagged   a resource-level entry names the section it belongs to,
                    e.g. "empty-observedstate" for a resource whose ObservedState
                    came out with no fields at all. Specific enough to act on.
  unflagged         nothing says anything

Unflagged is the number to drive to zero. A field a human must decide is a fine
outcome, as long as somebody is told about it.

Three of the five classes are the target:

  reference-shape   we emit a plain string, KCC has a Ref object
  moved             we emit it in Spec, KCC has it in status.observedState
  absent            it appears nowhere in our output

The other two are differences we accept, reported below the subtotal and excluded
from it:

  renamed                  same field, different name. Two shapes: an acronym
                           cased differently (bootDiskMIB/bootDiskMiB), and a
                           reference we emit un-suffixed (secretValue against
                           upstream's secretValueRef). Both are real CRD
                           differences; neither is a detection gap
  intentionally-different  google.protobuf.Value arms, which we map to
                           apiextensionsv1.JSON on purpose. Whether to keep doing
                           that is an open decision, deliberately deferred.

A note on a number this tool does not report: comparing +kcc:proto:field
annotations instead of CRD paths says the generator emits 16295 of the 16438
proto fields KCC declares. That is a useful check on whether generation is
healthy, and it is not this measurement -- using it to shrink the CRD gap
explains away fields users actually cannot set.

Four details decide whether the figure means anything, each of which produced a
wrong answer before it was fixed:

  * Report roots, not paths. When a parent is missing its whole subtree is
    reported missing too, which inflates every count several-fold. 298 missing
    observedState paths were 156 real defects.

  * Count a repeated field once. The scorer reports "foo" and "foo[]" separately;
    they are one field. 28 of the first published 533 were this.

  * Count a reference site once, whether or not it is suffixed. A missing "fooRef"
    hides its own .external/.name/.namespace/.kind children, so roots collapses it
    to one. Upstream does not always add the suffix -- producerAcceptLists[] is a
    reference too -- and without collapsing those children by hand an unsuffixed
    reference costs four where a suffixed one costs one.

  * Pair names across the reference rename. The queue names a field as we
    generated it (".spec.pipelineJob"); the baseline names it as upstream has it
    (".spec.pipelineJobRef.external"). Comparing literally scores 0% when the
    true figure is 5%.

  * Ignore the blanket queue entry. "untriaged-bulk-generation" names no field
    and no section, and covers everything trivially, which would report 100% on
    day one.

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

# Upstream does not always suffix a repeated reference. ComputeNetworkAttachment's
# producerAcceptLists[] and NetworkManagementConnectivityTest's relatedProjects[]
# are references with the same four children and no Ref in the name, so the suffix
# rule alone files them as "absent" and counts each child separately.
REF_LIST_CHILD = re.compile(r"\[\]\.(external|name|namespace|kind)$")


def is_reference_path(path, arrays=()):
    return bool(REF_SEGMENT.search(path)) or path in arrays

BUCKETS = ("spec-reference", "spec-other", "observedState")

# The section roots. A path whose parent is one of these is a top-level field.
SECTIONS = ("spec", "status.observedState")


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


def ref_arrays(paths):
    """Array paths that are really references, judged by a missing ".external".

    "external" is the discriminator rather than the whole child set: it is the
    marker KCC puts on a reference and nothing else in a CRD has it, whereas a
    plain repeated message can perfectly well have a "name" field of its own.
    """
    return {p[: -len("[].external")] + "[]"
            for p in paths if p.endswith("[].external")}


def roots(paths):
    """Reduce a missing-field list to one entry per real defect.

    Three reductions, all of which were inflating the published figure:

      * a path whose parent is also missing -- the parent explains it;
      * "foo" alongside "foo[]" -- the scorer emits both for a repeated field,
        and they are one field;
      * the external/name/namespace/kind children of an unsuffixed repeated
        reference -- one reference site, kept as the array itself so that it
        costs the same as a suffixed "fooRef" would.
    """
    present = set(paths)
    arrays = ref_arrays(paths)
    out = []
    seen = set()

    def key(path):
        return path.rstrip("[]")

    for p in paths:
        # "foo" and "foo[]" are the same field; keep whichever comes first.
        if key(p) in seen:
            continue
        parent = p.rsplit(".", 1)[0] if "." in p else ""
        if parent and (parent in present or parent + "[]" in present):
            continue
        if REF_LIST_CHILD.search(p) and parent in arrays:
            # Collapse onto the array, which is the reference site.
            p = parent
            if key(p) in seen:
                continue
        seen.add(key(p))
        out.append(p)
    return out


# The arms of google.protobuf.Value. We map that message to apiextensionsv1.JSON
# deliberately, so a baseline CRD modelling it as a union struct differs from ours
# by choice. Still a CRD difference; not a generation failure.
VALUE_ARMS = {"nullValue", "numberValue", "boolValue", "stringValue",
              "structValue", "listValue", "values"}

# Ordered so the report reads target classes first, then the two we accept.
TARGET_CLASSES = ("reference-shape", "moved", "absent")
ACCEPTED_CLASSES = ("renamed", "intentionally-different")
CLASSES = TARGET_CLASSES + ACCEPTED_CLASSES


def strip_section(path):
    """Drop the leading spec. / status.observedState. so the two can be compared."""
    for prefix in ("status.observedState.", "spec."):
        if path.startswith(prefix):
            return path[len(prefix):]
    return path


def classify(path, section, extras, arrays=()):
    """Say why a baseline field is not in our CRD.

    extras maps section -> the paths we emit that the baseline does not have, which
    is what makes "moved" and "renamed" visible: a field we put somewhere else shows
    up as missing in one place and extra in another.
    """
    if is_reference_path(path, arrays):
        # A reference we did produce, under the name upstream did not use.
        # ConnectorsConnection is the whole of this case: we emit a ref object at
        # spec.configVariables[].secretValue where upstream has secretValueRef.
        # The user's YAML still breaks, so it is a real CRD difference -- but it
        # is a rename, not a missing reference, and reporting it as undetected
        # sends someone to build detection for a field we already reference.
        stem = REF_SUFFIX.sub(lambda m: m.group(1) or "", path)
        for cand in (stem, stem + "[]"):
            if cand + ".external" in extras.get(section, ()):
                return "renamed"
        return "reference-shape"
    leaf = path.rsplit(".", 1)[-1].rstrip("[]")
    if leaf in VALUE_ARMS:
        return "intentionally-different"

    rel = strip_section(path)
    other = "spec" if section == "status.observedState" else "status.observedState"
    if any(strip_section(e) == rel for e in extras.get(other, ())):
        return "moved"

    # A proto map. Upstream renders one as a list with the key promoted into an
    # element -- the Terraform inheritance, since TypeMap holds only primitives --
    # while we emit a Go map, which crd-mcp-server writes with a .KEY segment. So
    # a missing "foo[]" against an emitted "foo.KEY" is one field in two shapes.
    if path.endswith("[]"):
        as_map = path[:-2] + ".KEY"
        if any(e == as_map or e.startswith(as_map + ".") for e in extras.get(section, ())):
            return "intentionally-different"

    # Same parent, same name but for case: an acronym the generator cased
    # differently, e.g. we write bootDiskMIB where KCC writes bootDiskMiB.
    #
    # Suffix as well as equality, because upstream also drops a leading qualifier
    # the proto carried: cloud_sql_instance is cloudSQLInstance to us and
    # sqlInstance upstream. Anchored at the end and length-guarded so "instance"
    # does not swallow every field ending in it.
    parent = path.rsplit(".", 1)[0] if "." in path else ""
    for e in extras.get(section, ()):
        e_parent = e.rsplit(".", 1)[0] if "." in e else ""
        if e_parent != parent:
            continue
        ours = e.rsplit(".", 1)[-1].lower()
        theirs = leaf.lower()
        if ours == theirs:
            return "renamed"
        if len(theirs) > 4 and ours.endswith(theirs):
            return "renamed"
    return "absent"


def bucket_of(section, path):
    if section == "status.observedState":
        return "observedState"
    return "spec-reference" if REF_SEGMENT.search(path) else "spec-other"


# A resource-level queue entry normally names nothing and so flags nothing. These
# reasons are the exception: each names a whole section of the resource, which is
# specific enough to act on. "your ObservedState came out empty" tells a human
# exactly what to go and do, and does it better than nineteen separate lines
# would. The blanket "untriaged-bulk-generation" is deliberately not here.
SECTION_REASONS = {
    "empty-observedstate": "status.observedState",
}


def queue_entries():
    """Judgement entries, keyed by kind.

    Returns (fields, sections, queued):

      fields   kind -> field paths named explicitly
      sections kind -> the section names a resource-level entry covers
      queued   every kind the queue mentions at all

    queued exists to separate two things that look identical in the output. A
    field can be unflagged because a detector looked and missed it, or because
    nothing ever ran for that resource. 15 of the 189 measured resources have no
    queue entry of any kind: their types file already existed when
    --prepopulate-spec ran, so AddTypeFile skipped and no queue was written, and
    scripts/queue-hints then skips them by design because it will not queue a
    resource nobody is generating. Both produce silence, and only one is a
    detection problem.
    """
    fields = defaultdict(set)
    sections = defaultdict(set)
    queued = set()
    for f in glob.glob("apis/*/needs_judgement_call.txt"):
        for line in open(f):
            m = re.match(r"kind=(\S+) ", line)
            if m:
                queued.add(m.group(1))
            m = re.match(r'kind=(\S+) group=\S+: field "([^"]+)"', line)
            if m:
                fields[m.group(1)].add(m.group(2).lstrip("."))
                continue
            m = re.match(r"kind=(\S+) group=\S+: resource reason=([a-zA-Z0-9-]+)", line)
            if m and m.group(2) in SECTION_REASONS:
                sections[m.group(1)].add(SECTION_REASONS[m.group(2)])
    return fields, sections, queued


# When upstream turns a field into a reference it appends "Ref" -- and it also
# drops a trailing noun that the reference makes redundant. "kmsKeyName" becomes
# "kmsKeyRef", not "kmsKeyNameRef"; so do cryptoKeyName, bucketName and
# cloudStorageObjectPath. Stripping "Ref" alone leaves "kmsKey", which matches no
# queue entry, so 20 correctly hinted fields read as unflagged. Only Name and Path
# occur in the corpus today; the rest are the same rename and cost nothing, since
# a candidate only credits an entry that exists at that exact path.
DROPPED_SUFFIXES = ("", "Name", "Path", "Id", "ID", "Email", "Uri", "URI",
                    "Names", "Paths", "Ids", "IDs", "Emails", "Uris", "URIs")


def queue_candidates(path):
    """The names a queue entry might use for this baseline field path."""
    out = {path.lstrip(".")}
    plain = REF_SUFFIX.sub(r"\1", REF_CHILD.sub("", path)).lstrip(".")
    # Re-attach any array marker after the suffix: bucketRefs[] came from
    # bucketName[], not bucket[]Name.
    stem, arr = (plain[:-2], "[]") if plain.endswith("[]") else (plain, "")
    for suffix in DROPPED_SUFFIXES:
        out.add(stem + suffix + arr)
    return out


# Markers the generator writes into a types file when it made a call it cannot
# justify from the proto. A flagged field ought to carry one: the queue is a work
# list somebody clears, the types file is what a reader opens.
TYPES_MARKERS = ("+kcc:guess=",)


def types_markers(path):
    """The JSON names in one types file that carry a generator note.

    Read from the Go source rather than the CRD because that is where the note is
    written; the CRD description inherits it, but only for fields that reached the
    CRD at all.
    """
    out = set()
    try:
        src = open(path).read()
    except OSError:
        return out
    pending = False
    for line in src.split("\n"):
        stripped = line.strip()
        if stripped.startswith("//"):
            if any(m in stripped for m in TYPES_MARKERS):
                pending = True
            continue
        m = re.search(r'json:"([A-Za-z0-9_]+)', line)
        if m:
            if pending:
                out.add(m.group(1))
            pending = False
        elif stripped and not stripped.startswith("//"):
            pending = False
    return out


def explained(entries, path):
    """Does any queue entry name this field, allowing for the reference rename?"""
    for want in queue_candidates(path):
        for e in entries:
            base = e.rstrip("[]")
            if want == e or want.startswith(base + ".") or want.startswith(base + "[]"):
                return True
    return False


def names_something_nearby(entries, path):
    """Does the queue name any field at or under this field's parent?

    Only meaningful for references, and only as an admission of uncertainty.
    Queue entries name a field as *we* generated it; the baseline names it as
    upstream renamed it. Stripping "Ref" and a trailing noun pairs kmsKeyName with
    kmsKeyRef, but nothing pairs cryptoKeyName with kmsKeyRef, or vpc with
    networkRef -- both of which are flagged today and both of which read as
    unflagged.

    So this is a third answer, not a second: the queue said something about this
    part of the resource, and we cannot tell whether it was about this field.
    Positional pairing was tried instead and rejected. Matching on the parent
    alone credits billingAccountRef to provisionedResourcesParent; restricting to
    an unambiguous one-missing-one-extra parent still mispairs clusterRef with
    topics, roughly 3 wrong in 15. A lossy guess inside the metric is worse than
    an honest column, and this number has already had to be corrected twice.

    The real fix is at the source: if a queue entry recorded the reference target
    type, this bucket could be resolved by target rather than by name. See
    greenfield-reference-generation.md.

    Returns "nested", "top-level" or "" -- the first two are both uncertain, but
    not equally so, and reporting them as one number flatters the result. For a
    reference inside a sub-message, a queue entry in that same sub-message is
    probably about it. For a top-level reference the parent is "spec" itself, so
    any entry anywhere in the resource qualifies, which is barely evidence at all.
    Half the bucket is the weak kind.
    """
    parent = path.rsplit(".", 1)[0] if "." in path else ""
    if not parent:
        return ""
    parent = parent.lstrip(".")
    want = name_words(path.rsplit(".", 1)[-1])
    for e in entries:
        if not (e == parent or e.startswith(parent + ".") or e.startswith(parent + "[]")):
            continue
        # Proximity alone credited billingAccountRef to spec.location, because at
        # the top level the parent is the whole section. Require a shared word.
        if name_words(e.rsplit(".", 1)[-1]) & want:
            return "nested"
    return ""


def name_words(leaf):
    """The lowercase words of a field name, minus a Ref suffix, singularised.

    Singularising is what pairs projectRefs with projects and agentRefs with
    agents, which an exact match misses over one letter.
    """
    leaf = REF_SUFFIX.sub(lambda m: m.group(1) or "", leaf.rstrip("[]"))
    words = re.findall(r"[A-Z]+(?![a-z])|[A-Z][a-z]*|[a-z]+", leaf)
    return {w[:-1].lower() if len(w) > 3 and w.endswith("s") else w.lower()
            for w in words}


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

    q, qsections, queued_kinds = queue_entries()
    matched = Counter()
    gap = {c: Counter() for c in CLASSES}   # field-flagged / section-flagged / unflagged
    unflagged_list = defaultdict(list)
    unsure_list = defaultdict(list)
    never_queued = []
    n = skipped = 0

    for line in open(args.resources):
        if not line.strip():
            continue
        kind, svc, types_path, crd = line.rstrip("\n").split("\t")[:4]
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
        markers = types_markers(types_path)
        m, missing, extra = parse_score(text)
        matched.update(m)
        for section, paths in missing.items():
            arrays = ref_arrays(paths)
            for p in roots(paths):
                c = classify(p, section, extra, arrays)
                entries = q.get(kind, ())
                if explained(entries, p):
                    gap[c]["field"] += 1
                    # Does the generated source say anything about it too?
                    leaf = strip_section(p).rsplit(".", 1)[-1].rstrip("[]")
                    leaf = REF_SUFFIX.sub(lambda m: m.group(1) or "", leaf)
                    if leaf in markers:
                        gap[c]["in_types"] += 1
                elif section in qsections.get(kind, ()):
                    gap[c]["section"] += 1
                elif c == "reference-shape" and names_something_nearby(entries, p):
                    # Cannot tell: the queue spoke about this parent, but upstream
                    # renamed the field so the two cannot be paired by name. Split
                    # by how much the parent narrows it down.
                    where = names_something_nearby(entries, p)
                    gap[c]["unsure" if where == "nested" else "weak"] += 1
                    unsure_list[c if where == "nested" else "weak"].append((kind, p))
                else:
                    gap[c]["unflagged"] += 1
                    unflagged_list[c].append((kind, p))
                    if kind not in queued_kinds:
                        never_queued.append((kind, p))

    print(f"resources scored: {n}" + (f"   skipped: {skipped}" if skipped else ""))
    print(f"baseline: {args.ref}\n")
    produced = matched["spec"] + matched["required"] + matched["status.observedState"]
    missing_total = sum(sum(gap[c].values()) for c in CLASSES)
    surface = produced + missing_total

    # The headline: three states, one field in exactly one of them.
    flagged_q = sum(gap[c]["field"] + gap[c]["section"] for c in CLASSES)
    flagged_both = sum(gap[c]["in_types"] for c in CLASSES)
    missed = missing_total - flagged_q

    print(f"Against KCC master at {args.ref}, every field in its CRDs is one of three things.\n")
    print(f"  {'1. implemented':34s} {produced:6d}   ({100 * produced / surface:.1f}%)")
    print("        in our types and CRD, at the same path")
    print(f"  {'2. flagged':34s} {flagged_q:6d}   ({100 * flagged_q / surface:.1f}%)")
    print(f"        named in needs_judgement_call.txt{'':12s}{flagged_q:6d}")
    print(f"        ...and also in the types file{'':15s}{flagged_both:6d}")
    print(f"  {'3. missed':34s} {missed:6d}   ({100 * missed / surface:.1f}%)")
    print("        in neither. Nobody would find out.")
    print(f"  {'':34s} {'-' * 6}")
    print(f"  {'fields in KCC master CRDs':34s} {surface:6d}")

    emitted_elsewhere = sum(sum(gap[c].values()) for c in ACCEPTED_CLASSES)
    unpairable = sum(gap[c]["unsure"] + gap[c]["weak"] for c in CLASSES)
    if emitted_elsewhere or unpairable:
        print("\n  Two things inside \"missed\" that are not quite missing:")
        if emitted_elsewhere:
            print(f"    {emitted_elsewhere:4d}  we do emit, but renamed or modelled differently -- a mis-cased")
            print("          acronym, a reference without the Ref suffix, a Value union we map")
            print("          to JSON on purpose. Broken for a user, but not absent.")
        if unpairable:
            print(f"    {unpairable:4d}  references the queue may well name, under a name we cannot pair")
            print("          to upstream's. DatastreamPrivateConnection's vpc is upstream's")
            print("          networkRef; nothing matches those by name. Counted as missed,")
            print("          which overstates the gap rather than flattering it.")

    if flagged_both < flagged_q:
        print(f"\nThe second line of state 2 is the one to watch. {flagged_q - flagged_both} fields are named in")
        print("the queue and nowhere in the generated source, so a person reading the type")
        print("sees a plain string with nothing to suggest it is unfinished. The queue is a")
        print("work list somebody clears; the types file is what a reader opens.")

    print(f"\n\nBelow: why each of the {missing_total} we did not implement differs, which routes the")
    print("fix rather than excusing it. Columns say how we know about it.\n")

    def row(label, counters):
        f, sec = counters["field"], counters["section"]
        total = sum(counters.values()) - counters["in_types"]
        print(f"  {label:24s} {total:8d} {f:9d} {sec:11d} {total - f - sec:9d}")

    print(f"  {'why it differs':24s} {'we miss':>8s} {'by field':>9s} "
          f"{'by section':>11s} {'missed':>9s}")
    for c in TARGET_CLASSES:
        row(c, gap[c])
    target = Counter()
    for c in TARGET_CLASSES:
        target.update(gap[c])
    print("  " + "-" * 63)
    row("subtotal, the target", target)

    accepted = Counter()
    for c in ACCEPTED_CLASSES:
        accepted.update(gap[c])
    if sum(accepted.values()):
        print("\n  differences we accept, not counted above:")
        for c in ACCEPTED_CLASSES:
            row(c, gap[c])

    print("\nThe target is state 3 at zero: a field a human must decide is a fine")
    print("outcome, a field nobody was told about is not. Watch \"implemented\"")
    print("alongside it -- flagging fields by no longer emitting them improves this")
    print("report and takes working fields away.")

    if args.list_silent:
        for c in CLASSES:
            if not unflagged_list[c]:
                continue
            print(f"\n### missed without flagging, {c} ({len(unflagged_list[c])})")
            for kind, p in sorted(unflagged_list[c]):
                print(f"  {kind}\t{p}")
        for c in list(CLASSES) + ["weak"]:
            if not unsure_list[c]:
                continue
            print(f"\n### unsure, {c} ({len(unsure_list[c])})")
            for kind, p in sorted(unsure_list[c]):
                print(f"  {kind}\t{p}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
