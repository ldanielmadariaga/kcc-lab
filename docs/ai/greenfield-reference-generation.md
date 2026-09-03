# Follow-up: generate references from the proto annotation

**Not built. Deferred deliberately**, because it is generation and the current priority is finishing
detection — every field we do not generate should be named somewhere a human will look before more
generation work starts. Written down here so the research behind it is not lost.

## Today: the lists only flag, they never generate

Whatever the reference detectors conclude, the types file still gets `KmsKeyName *string`. There are
three detectors and they differ in what they can reach:

| detector | source | consumers |
|---|---|---|
| `google.api.resource_reference` | the proto — **authoritative**, names the exact target type | queue only, `possible-reference` |
| `refs.Classify` — resource-name templates in the description, `serviceAccount` suffix, GCS buckets | CRD description | `queue-hints` **and** `TestMissingRefs` |
| `refs.MatchName` / `NameRules` — name spellings | field name | `queue-hints` only |

`scaffoldRefsFile` does generate a `<Kind>Ref` type, but for the resource **itself**, so other
resources can point at it. It never converts one of the resource's own fields.

## The finding that makes this cheap

**The mapper generator already supports references.** `mappergenerator.go:290` looks for a
`<Field>Ref` in the Go struct and emits `out.XRef = &T{External: in.X}`; line 307 does the same for
`<Fields>Refs`; line 633 maps back. So the types side is the only missing piece — declare the field
as a Ref and the mapper follows. That should still be demonstrated rather than assumed, since the
whole plan rests on it.

## Scope: the authoritative detector only

`resource_reference` is not a heuristic. The proto states `type: "cloudkms.googleapis.com/CryptoKey"`,
naming the target resource. Turning a statement into a field is not guessing.

**The two heuristics stay flag-only.** Description rules and name rules are right in general and will
be wrong for outliers, which is what the queue is for.

### Sizing

37 annotation-derived entries exist across the corpus:

* 19 match a Ref type that already exists
* 6 more (`cloudkms.googleapis.com/CryptoKey` → `KMSCryptoKeyRef`) match only via an explicit table
* 12 would need a Ref type written first — out of scope, they stay flagged

So **~25 fields**. Note what that does to the score: those fields are *already flagged*, so unflagged
does not move; `produced` rises instead. The value is proving the path end to end on the subset where
we are not guessing.

### The target→type mapping cannot be derived

It looks derivable and is not. A naive transform of `cloudkms.googleapis.com/CryptoKey` yields
`CryptoKeyRef` or `CloudkmsCryptoKeyRef`; the actual type is `KMSCryptoKeyRef`. Worse, several naive
matches are actively wrong — `container.googleapis.com/Cluster` would match a bare `ClusterRef`, and
several unrelated `ClusterRef` types exist. There are 442 `Ref` types in `apis/`. The table must be
explicit, and an unmapped target must fall through to today's behaviour rather than guess.

## Sketch

1. **An explicit target table**, `dev/tools/controllerbuilder/scaffold/referencetargets.go`: target
   string → Go type name, import qualifier and path. ~20 entries. Anything absent falls through
   unchanged.
2. **Emit the Ref in `PrepopulateSpec`**, naming the field the way upstream does: append `Ref` and
   drop the trailing noun the reference makes redundant, so `kms_key_name` becomes `kmsKeyRef`, not
   `kmsKeyNameRef`.
3. **Leave a trace in the generated type**, not only in the queue — the same treatment
   `--place-server-set-fields` got, and the reason it matters is that this can produce a false
   positive that only a reader of the type will catch:

   ```go
   // REFERENCE GENERATED: the proto's google.api.resource_reference names
   // cloudkms.googleapis.com/CryptoKey as the target. Confirm the target type is right.
   ```

   Reuse the `note` parameter on `codegen.WriteField`. Also emit a queue entry, distinct from
   `possible-reference` so generated and merely-suspected references can be told apart.
4. **Register the import** — `ExtraImportsFor` scans the rendered body against
   `codegen.QualifierImports`, so adding `refsv1beta1` there is the whole import story.
5. **Opt-in per service**, `--generate-annotated-refs`, default off. Changing a field from `string`
   to an object is a breaking CRD change.

## Verification when it is built

* A mapped target resolves; an unmapped one falls back to plain string plus the existing queue entry;
  `cloudkms.googleapis.com/CryptoKey` resolves to `KMSCryptoKeyRef` and not `CryptoKeyRef`.
* Naming: `kms_key_name` → `kmsKeyRef`, `network` → `networkRef`.
* End to end in a scratch tree with `--output-api` on `VertexAIPipelineJob`, whose `network` field
  carries the annotation with target `compute.googleapis.com/Network`.
* **Run `generate-mapper` for that service and confirm the mapper picks it up with no mapper change.**
* Flag off: byte-identical output.

## One thing to fix while here

Queue entries name a field as *we* generated it, and the baseline names it as *upstream* renamed it.
When the rename is not a simple suffix drop the two cannot be paired, and 83 reference fields sit in
an "unable to tell" bucket in `silence_report.py` because of it. If a generated reference recorded
its target type in the queue entry, that bucket could be resolved by target rather than by name. See
[greenfield-detection-gaps.md](greenfield-detection-gaps.md).

## Explicitly not in scope

Generating references from `Classify` or `MatchName`. If that ever changes it needs the same trace
comment and its own measured precision argument.
