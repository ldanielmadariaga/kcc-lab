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
TARGET_CLASSES = ("reference-shape", "moved", "absent", "reference-not-detected")
ACCEPTED_CLASSES = ("renamed", "intentionally-different")
# Classes where the field IS in our output, just not where or how upstream has
# it. They stay under "missed" -- a user's YAML still breaks -- but they are a
# detection or placement problem, not a generation one, and the headline has to
# say so or it sends people to build generation for fields we already generate.
EMITTED_CLASSES = ("moved", "reference-not-detected")
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
        # We produce the field, as a plain string, and never spotted that it is a
        # reference. That is a different problem from not producing it at all:
        # nothing has to be generated, only detected. Rolling the two together
        # made 19 fields read as "in neither our types nor the queue" when they
        # were sitting in the types file as strings.
        #
        # queue_candidates gives every spelling the trailing-noun rename can
        # produce, so billingAccountRef finds our billingAccount and dataStoreRefs
        # finds our dataStoreIDs.
        emitted = {e.rstrip("[]") for e in extras.get(section, ())}
        if {c.rstrip("[]") for c in queue_candidates(path)} & emitted:
            return "reference-not-detected"
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
      leaves   kind -> leaf names its service flagged in a comment, for the
               shared-nested-message findings that name no Kind

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
    leaves = defaultdict(set)
    for f in glob.glob("apis/*/needs_judgement_call.txt"):
        kinds_here, leaves_here = set(), set()
        for line in open(f):
            # Findings against a shared nested message are written as comments,
            # because a nested type belongs to no single Kind and every
            # non-comment line here suppresses [refs] for the Kind it names. They
            # carry the proto message, not the KRM path, so they can only be
            # matched on the field's leaf name -- looser than the path matching
            # above, and it credits every Kind in the service for a finding
            # recorded against the service. Without it 7 fields read as unflagged
            # while the queue named them.
            m = re.match(r"#\s*possible-reference-by-sibling:\s*\S+\.(\w+)\s", line)
            if m:
                leaves_here.add(m.group(1).lower())
                continue
            m = re.match(r"kind=(\S+) ", line)
            if m:
                queued.add(m.group(1))
                kinds_here.add(m.group(1))
            m = re.match(r'kind=(\S+) group=\S+: field "([^"]+)"', line)
            if m:
                fields[m.group(1)].add(m.group(2).lstrip("."))
                continue
            m = re.match(r"kind=(\S+) group=\S+: resource reason=([a-zA-Z0-9-]+)", line)
            if m and m.group(2) in SECTION_REASONS:
                sections[m.group(1)].add(SECTION_REASONS[m.group(2)])
        for k in kinds_here:
            leaves[k] |= leaves_here
    return fields, sections, queued, leaves


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


def _leaf_flagged(leaves, path):
    """Did a shared-nested-message comment name this field, by leaf name?

    Matched after the exact-path checks, so it only ever rescues a field nothing
    more precise accounted for.
    """
    if not leaves:
        return False
    return any(c.rsplit(".", 1)[-1].rstrip("[]").lower() in leaves
               for c in queue_candidates(path))


def we_emit_it(path, section, extra):
    """Is this baseline field in our output at all, under any name we can pair?

    This is the axis the report leads with, because it decides what kind of work
    the difference needs. A field we emit in the wrong section or as a plain
    string needs detecting or moving; a field we emit nowhere needs generating.
    Rolling the two together made "missed" read as absence when over half of it
    was neither.

    Same-section pairing uses queue_candidates, which already knows the trailing
    noun upstream drops when it makes a reference, so billingAccountRef finds our
    billingAccount and dataStoreRefs finds our dataStoreIDs. The cross-section
    case is what "moved" means: the same relative path in the other half.
    """
    cands = {c.rstrip("[]") for c in queue_candidates(path)}
    if cands & {e.rstrip("[]") for e in extra.get(section, ())}:
        return True
    other = "spec" if section == "status.observedState" else "status.observedState"
    rel = {strip_section(c).rstrip("[]") for c in cands}
    return bool(rel & {strip_section(e).rstrip("[]") for e in extra.get(other, ())})


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

    q, qsections, queued_kinds, qleaves = queue_entries()
    matched = Counter()
    gap = {c: Counter() for c in CLASSES}   # field / section / unsure / weak / unflagged
    # Tracked apart from gap because it is a SUB-count of "field" -- a flagged
    # field that also carries a marker -- not a bucket of its own. Summing it
    # into the totals double-counted 16 fields and put "truly missed" 19 out.
    in_types_total = 0
    # The primary axis: (do we emit it, did we say anything). Counted alongside
    # gap rather than derived from it, because gap is keyed by *why* a field
    # differs and that is a different question from *whether we produced it*.
    shape = Counter()
    shape_cls = defaultdict(Counter)
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
                # Three classes are only ever returned when classify() already
                # found our counterpart in extras: "moved" needs the same path in
                # the other section, "renamed" needs a matching name at the same
                # parent, "reference-not-detected" needs a plain field at the
                # de-suffixed stem. Trusting the classifier there keeps the two
                # from disagreeing -- name-pairing alone called 29 renamed fields
                # unproduced, which put them in "missing" while the class column
                # said we emit them.
                emitted = (c in ("moved", "renamed", "reference-not-detected")
                           or we_emit_it(p, section, extra))
                told = bool(explained(entries, p)
                            or section in qsections.get(kind, ())
                            or _leaf_flagged(qleaves.get(kind, ()), p))
                shape[(emitted, told)] += 1
                shape_cls[c][(emitted, told)] += 1
                if explained(entries, p):
                    gap[c]["field"] += 1
                    # Does the generated source say anything about it too?
                    leaf = strip_section(p).rsplit(".", 1)[-1].rstrip("[]")
                    leaf = REF_SUFFIX.sub(lambda m: m.group(1) or "", leaf)
                    if leaf in markers:
                        in_types_total += 1
                elif section in qsections.get(kind, ()):
                    gap[c]["section"] += 1
                elif _leaf_flagged(qleaves.get(kind, ()), p):
                    gap[c]["field"] += 1
                elif c == "reference-shape" and names_something_nearby(entries, p):
                    # Cannot tell: the queue spoke about this parent, but upstream
                    # renamed the field so the two cannot be paired by name. Split
                    # by how much the parent narrows it down.
                    where = names_something_nearby(entries, p)
                    gap[c]["unsure" if where == "nested" else "weak"] += 1
                    unsure_list[c if where == "nested" else "weak"].append((kind, p))
                else:
                    gap[c]["unflagged"] += 1
                    unflagged_list[c].append((kind, p, emitted))
                    if kind not in queued_kinds:
                        never_queued.append((kind, p))

    print(f"resources scored: {n}" + (f"   skipped: {skipped}" if skipped else ""))
    print(f"baseline: {args.ref}\n")
    produced = matched["spec"] + matched["required"] + matched["status.observedState"]
    missing_total = sum(sum(gap[c].values()) for c in CLASSES)
    surface = produced + missing_total

    # The headline splits on what we produced, not on whether we mentioned it.
    # Those are close to independent -- 280 of the 339 discrepancies are flagged
    # and only 39 of the 286 absences are -- so leading with "flagged" hid the
    # distinction that decides what work a difference actually needs.
    discrepancy = shape[(True, True)] + shape[(True, False)]
    absent      = shape[(False, True)] + shape[(False, False)]
    disc_told   = shape[(True, True)]
    abs_told    = shape[(False, True)]
    # Fields we decline to produce on purpose: the google.protobuf.Value union
    # arms, which we map whole to apiextensionsv1.JSON so the individual arms
    # cannot exist as fields, plus casing renames. Held out because a headline
    # that counts them as absences overstates the gap.
    by_design = sum(shape_cls[c][(False, True)] + shape_cls[c][(False, False)]
                    for c in ACCEPTED_CLASSES)
    real_gap  = absent - by_design
    flagged_both = in_types_total

    pct = lambda x: f"({100 * x / surface:.1f}%)"
    print(f"Against KCC master at {args.ref}, every field in its CRDs is one of three things.\n")
    print(f"  {'1. implemented':34s} {produced:6d}   {pct(produced)}")
    print("        in our types and CRD, at the same path")
    print(f"  {'2. discrepancy':34s} {discrepancy:6d}   {pct(discrepancy)}")
    print("        we produce it, but not as upstream has it")
    print(f"        flagged for a second pass{'':9s}{disc_told:6d}   ({100 * disc_told / discrepancy:.0f}% of them)")
    print(f"        nothing says so{'':19s}{discrepancy - disc_told:6d}")
    print(f"  {'3. missing':34s} {absent:6d}   {pct(absent)}")
    print("        we produce nothing at all")
    print(f"        a gap to close{'':20s}{real_gap:6d}")
    print(f"        we model it differently on purpose{'':0s}{by_design:6d}")
    print(f"        flagged for a second pass{'':9s}{abs_told:6d}   ({100 * abs_told / absent:.0f}% of them)")
    print(f"  {'':34s} {'-' * 6}")
    print(f"  {'fields in KCC master CRDs':34s} {surface:6d}")

    if flagged_both < disc_told + abs_told:
        print(f"\nOf the {disc_told + abs_told} flagged, {flagged_both} also carry a marker in the types")
        print("file. The rest are named in the queue and nowhere in the generated source, so a")
        print("person reading the type sees a plain string with nothing to suggest it is")
        print("unfinished. The queue is a work list somebody clears; the types file is what a")
        print("reader opens.")

    unpairable = sum(gap[c]["unsure"] + gap[c]["weak"] for c in CLASSES)
    if unpairable:
        print(f"\n  {unpairable} of the above are references upstream renamed rather than suffixed:")
        print("  DatastreamPrivateConnection's vpc is upstream's networkRef, and no name")
        print("  match bridges those. Counted as unflagged, which overstates the gap rather")
        print("  than flattering it.")

    wrong_section = gap["moved"]["unflagged"]
    undetected_ref = gap["reference-not-detected"]["unflagged"]
    print(f"\n\nBelow: why each of the {missing_total} we did not implement differs, which routes the")
    print("fix rather than excusing it. Columns say how we know about it.\n")

    def row(label, classes):
        d = Counter()
        for c in classes:
            d.update(shape_cls[c])
        prod = d[(True, True)] + d[(True, False)]
        gone = d[(False, True)] + d[(False, False)]
        told = d[(True, True)] + d[(False, True)]
        print(f"  {label:24s} {prod + gone:7d} {prod:10d} {gone:14d} {told:9d}")

    print(f"  {'why it differs':24s} {'total':>7s} {'we emit it':>10s} "
          f"{'we emit nothing':>14s} {'flagged':>9s}")
    for c in TARGET_CLASSES:
        row(c, [c])
    print("  " + "-" * 66)
    row("subtotal, the target", TARGET_CLASSES)

    if any(sum(shape_cls[c].values()) for c in ACCEPTED_CLASSES):
        print("\n  differences we accept, not counted above:")
        for c in ACCEPTED_CLASSES:
            row(c, [c])

    print("\nThe target is the second line of state 2 and the first of state 3: a field")
    print("a human must decide is a fine outcome, a field nobody was told about is not.")
    print("Watch \"implemented\" alongside them -- flagging fields by no longer emitting")
    print("them improves this report and takes working fields away.")

    if args.list_silent:
        # Grouped by whether we emit the field, because that is what decides the
        # work: a discrepancy nobody flagged needs detecting or moving, a silent
        # absence needs generating. The class is the second key.
        for group, label in ((True, "discrepancy, nothing says so"),
                             (False, "missing, nothing says so")):
            for c in CLASSES:
                rows = [(k, path) for k, path, em in unflagged_list[c] if em == group]
                if not rows:
                    continue
                print(f"\n### {label}, {c} ({len(rows)})")
                for kind, path in sorted(rows):
                    print(f"  {kind}\t{path}")
        for c in list(CLASSES) + ["weak"]:
            if not unsure_list[c]:
                continue
            print(f"\n### unsure, {c} ({len(unsure_list[c])})")
            for kind, p in sorted(unsure_list[c]):
                print(f"  {kind}\t{p}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
