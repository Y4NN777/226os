#!/usr/bin/env bash
# Tests for scripts/deploy-verify-tested.sh (production trusts a `tested-<tree>` digest only when a signed attestation
# made by the deploy workflow on develop says that exact tree was tested on staging) with a fake `gh`, plus static checks
# of the workflow: only `promote` may attest, the predicate type is the same on both sides, nothing is interpolated.
# Usage: scripts/tests/deploy-verify-tested_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
verify="$root/scripts/deploy-verify-tested.sh"
wf="$root/.github/workflows/deploy.yml"
ptype='https://226os.tech/attestations/tested-on-staging/v1'

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin"
ns=ghcr.io/y4nn777
d_api="sha256:$(printf 'a%.0s' {1..64})"
d_web="sha256:$(printf 'b%.0s' {1..64})"
d_mig="sha256:$(printf 'c%.0s' {1..64})"
cat >"$tmp/images.env" <<EOF
API_IMAGE=$ns/226os-api@$d_api
WEB_IMAGE=$ns/226os-web@$d_web
MIGRATE_IMAGE=$ns/226os-migrate@$d_mig
EOF

# The fake gh logs its argv and prints $FAKE_GH_OUT/<name>.json for the image asked (name = api, web or migrate),
# or fails with FAKE_GH_RC when set. A missing file means "no attestation": gh exits 1 with nothing on stdout.
cat >"$tmp/bin/gh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$FAKE_GH_LOG"
[ -z "${FAKE_GH_RC:-}" ] || { echo "gh: boom" >&2; exit "$FAKE_GH_RC"; }
case "$1 $2" in "attestation verify") ;; *) echo "unexpected gh call" >&2; exit 99 ;; esac
name="${3#oci://*/226os-}"
name="${name%%@*}"
[ -f "$FAKE_GH_OUT/$name.json" ] || { echo "no attestations found" >&2; exit 1; }
cat "$FAKE_GH_OUT/$name.json"
EOF
chmod +x "$tmp/bin/gh"

# att TREE: what `gh attestation verify --format json` prints for one verified attestation of that tree.
att() {
  printf '[{"attestation":{},"verificationResult":{"statement":{"predicateType":"%s","predicate":{"tree":"%s","sha":"s1"}}}}]\n' "$ptype" "$1"
}
setup_all() { rm -rf "$tmp/out"; mkdir "$tmp/out"; for n in api web migrate; do att "$1" >"$tmp/out/$n.json"; done; }

run() { : >"$tmp/log"; rc=0; out="$(env PATH="$tmp/bin:$PATH" FAKE_GH_LOG="$tmp/log" FAKE_GH_OUT="$tmp/out" "$@" 2>&1)" || rc=$?; }

tree=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
other=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb

setup_all "$tree"
run "$verify" "$tree" "$tmp/images.env"
if [ "$rc" -eq 0 ] && [ "$(grep -c '^attestation verify' "$tmp/log")" -eq 3 ]; then ok "valid attestations of the right tree: the three images are accepted"; else ko "valid (rc=$rc out=$out)"; fi

for flag in "--repo Y4NN777/226os" "--signer-workflow Y4NN777/226os/.github/workflows/deploy.yml" "--source-ref refs/heads/develop" "--predicate-type $ptype" "--format json"; do
  [ "$(grep -c -- "$flag" "$tmp/log")" -eq 3 ] && ok "every verification passes $flag" || ko "flag $flag missing in $(cat "$tmp/log")"
done
grep -q "oci://$ns/226os-api@$d_api" "$tmp/log" && grep -q "oci://$ns/226os-migrate@$d_mig" "$tmp/log" && ok "images are verified by digest" || ko "digest ref"

setup_all "$other"
run "$verify" "$tree" "$tmp/images.env"
[ "$rc" -ne 0 ] && grep -q '::error' <<<"$out" && ok "a digest attested for ANOTHER tree is refused (the forged tag case)" || ko "other tree (rc=$rc out=$out)"

setup_all "$tree"
att "$other" >"$tmp/out/web.json"
run "$verify" "$tree" "$tmp/images.env"
[ "$rc" -ne 0 ] && ok "one image attested for another tree refuses the whole set" || ko "mixed (rc=$rc)"

# several attestations for the digest (a re-run): one of them with the right tree is enough
setup_all "$tree"
{ printf '['; att "$other" | sed 's/^\[//; s/\]$//'; printf ','; att "$tree" | sed 's/^\[//; s/\]$//'; printf ']\n'; } >"$tmp/out/api.json"
run "$verify" "$tree" "$tmp/images.env"
[ "$rc" -eq 0 ] && ok "a re-run with several attestations is accepted when one matches" || ko "several (rc=$rc out=$out)"

rm -f "$tmp/out/migrate.json"
run "$verify" "$tree" "$tmp/images.env"
[ "$rc" -ne 0 ] && grep -q '::error' <<<"$out" && ok "no attestation for an image: refused" || ko "none (rc=$rc out=$out)"

setup_all "$tree"
run FAKE_GH_RC=4 "$verify" "$tree" "$tmp/images.env"
[ "$rc" -ne 0 ] && ok "a gh error is refused (fail closed)" || ko "gh error (rc=$rc)"

echo '[]' >"$tmp/out/api.json"
run "$verify" "$tree" "$tmp/images.env"
[ "$rc" -ne 0 ] && ok "an empty verified list is refused" || ko "empty (rc=$rc)"

echo 'not json' >"$tmp/out/api.json"
run "$verify" "$tree" "$tmp/images.env"
[ "$rc" -ne 0 ] && ok "unparsable gh output is refused" || ko "garbage (rc=$rc)"

# a verified attestation of another predicate type, or without a tree, never counts
setup_all "$tree"
printf '[{"verificationResult":{"statement":{"predicateType":"https://slsa.dev/provenance/v1","predicate":{"tree":"%s"}}}}]\n' "$tree" >"$tmp/out/api.json"
run "$verify" "$tree" "$tmp/images.env"
[ "$rc" -ne 0 ] && ok "an attestation of another predicate type is refused" || ko "predicate type (rc=$rc)"
printf '[{"verificationResult":{"statement":{"predicateType":"%s","predicate":{"sha":"x"}}}}]\n' "$ptype" >"$tmp/out/api.json"
run "$verify" "$tree" "$tmp/images.env"
[ "$rc" -ne 0 ] && ok "an attestation without a tree is refused" || ko "no tree (rc=$rc)"

# inputs: a bad tree, an image that is not a digest reference, a missing image
setup_all "$tree"
run "$verify" 'x;y' "$tmp/images.env"
[ "$rc" -ne 0 ] && ! [ -s "$tmp/log" ] && ok "a tree with shell characters is refused before any call" || ko "bad tree (rc=$rc)"
printf 'API_IMAGE=%s/226os-api:tested-t1\nWEB_IMAGE=%s/226os-web@%s\nMIGRATE_IMAGE=%s/226os-migrate@%s\n' "$ns" "$ns" "$d_web" "$ns" "$d_mig" >"$tmp/bad.env"
run "$verify" "$tree" "$tmp/bad.env"
[ "$rc" -ne 0 ] && ok "an image reference by tag (not by digest) is refused" || ko "tag ref (rc=$rc)"
grep -v '^WEB_IMAGE=' "$tmp/images.env" >"$tmp/short.env"
run "$verify" "$tree" "$tmp/short.env"
[ "$rc" -ne 0 ] && ok "a missing image line is refused" || ko "missing line (rc=$rc)"
run "$verify"
[ "$rc" -eq 64 ] && ok "no argument is a usage error" || ko "usage (rc=$rc)"

# --- static checks of the workflow --------------------------------------------------------------------------------
# job -> (permissions, step uses/run) read with PyYAML (PyYAML parses `on:` as True, irrelevant here).
perm="$(python3 -I - "$wf" <<'PY'
import sys, yaml
w = yaml.safe_load(open(sys.argv[1]))
for name, job in w["jobs"].items():
    p = job.get("permissions") or {}
    print(name, " ".join(f"{k}={v}" for k, v in sorted(p.items())))
PY
)"
writers="$(awk '/(id-token|attestations)=write/ {print $1}' <<<"$perm" | tr '\n' ' ')"
[ "$writers" = "promote " ] && ok "only promote has id-token or attestations write" || ko "attest writers: $writers"
grep -E '^promote ' <<<"$perm" | grep -q 'id-token=write' && grep -E '^promote ' <<<"$perm" | grep -q 'attestations=write' &&
  grep -E '^promote ' <<<"$perm" | grep -q 'packages=write' && grep -E '^promote ' <<<"$perm" | grep -q 'contents=read' &&
  ok "promote keeps packages write and contents read, and gains both attestation permissions" || ko "promote permissions: $perm"
if python3 -I -c 'import sys,yaml;sys.exit(0 if yaml.safe_load(open(sys.argv[1]))["permissions"]=={"contents":"read"} else 1)' "$wf"; then
  ok "top-level permissions stay contents: read"
else ko "top-level permissions widened"; fi

pin='actions/attest@1e69f48acb82d1966a394da916b4c1698aa569d6 # v4.2.2'
[ "$(grep -c 'uses: actions/attest@' "$wf")" -eq 3 ] && [ "$(grep -c "uses: $pin" "$wf")" -eq 3 ] && ok "three attest steps (api, web, migrate), SHA-pinned" || ko "attest pins"
[ "$(grep -c "predicate-type: $ptype" "$wf")" -eq 3 ] && ok "the attestations use the tested-on-staging predicate type" || ko "predicate type in the workflow"
grep -q "$ptype" "$verify" && ok "the verifier checks the same predicate type" || ko "predicate type in the verifier"
# `gh attestation verify` reads the attestations API, not the registry: nothing is pushed to the registry, and the
# storage record (default true) is switched off explicitly.
! grep -q 'push-to-registry' "$wf" && [ "$(grep -c 'create-storage-record: false' "$wf")" -eq 3 ] && ok "attestations are not pushed to the registry; the storage record is off" || ko "push-to-registry / create-storage-record"
grep -q 'scripts/deploy-verify-tested.sh' "$wf" && ok "resolve calls the verifier" || ko "resolve does not call the verifier"
if grep 'deploy-verify-tested' "$wf" | grep -q '\${{'; then ko "an expression is interpolated on the verify call"; else ok "no expression interpolated on the verify call"; fi

# --- the digests production trusts are the ones THIS run built, deployed and tested ---------------------------------
steps="$(python3 -I - "$wf" <<'PY'
import json, sys, yaml
w = yaml.safe_load(open(sys.argv[1]))
print(json.dumps({n: {"needs": j.get("needs"), "outputs": sorted((j.get("outputs") or {}).keys()), "steps": j.get("steps", [])} for n, j in w["jobs"].items()}))
PY
)"
sq() { jq -r "$1" <<<"$steps"; }
[ "$(sq '.build.outputs | join(",")')" = api_image,migrate_image,web_image ] && ok "build outputs the digests of the images it pushed" || ko "build outputs: $(sq '.build.outputs | join(",")')"
[ "$(sq '.promote.needs | sort | join(",")')" = deploy,e2e,plan,resolve ] && ok "promote needs plan, resolve, deploy and e2e" || ko "promote needs: $(sq '.promote.needs | tojson')"
if sq '.promote.steps | map(.run // "") | join("\n")' | grep -q 'deploy-images.sh resolve'; then ko "promote re-reads a tag with deploy-images.sh resolve"; else ok "promote never re-reads a tag (no deploy-images.sh resolve)"; fi
penv="$(sq '.promote.steps | map(.env // {} | tojson) | join("\n")')"
grep -q 'needs.resolve.outputs.api_image' <<<"$penv" && grep -q 'needs.resolve.outputs.web_image' <<<"$penv" && grep -q 'needs.resolve.outputs.migrate_image' <<<"$penv" && ok "promote tags and attests the digests of resolve (the ones deployed and tested)" || ko "promote does not use needs.resolve.outputs"
sq '.promote.steps | map(.run // "") | join("\n")' | grep -qE 'deploy-images.sh tag "tested-\$TREE" ' && ok "promote tags from a digests file, not from a tag" || ko "promote tag call"
# staging: the digests come from build and no step reads the <sha> tag back. Production: tested tag plus the verifier.
sq '.resolve.steps | map(select((.["if"] // "") | contains("staging"))) | map((.run // ""), (.env // {} | tojson)) | join("\n")' >"$tmp/staging-resolve.txt"
if grep -q 'needs.build.outputs.api_image' "$tmp/staging-resolve.txt" && ! grep -q 'deploy-images.sh resolve' "$tmp/staging-resolve.txt"; then ok "staging resolve takes the digests from build and does not re-read the tag"; else ko "staging resolve: $(cat "$tmp/staging-resolve.txt")"; fi
sq '.resolve.steps | map(select((.["if"] // "") | contains("production"))) | map(.run // "") | join("\n")' >"$tmp/prod-resolve.txt"
if grep -q 'deploy-images.sh resolve "tested-' "$tmp/prod-resolve.txt" && grep -q 'scripts/deploy-verify-tested.sh' "$tmp/prod-resolve.txt"; then ok "production resolve reads tested-<tree> and runs the verifier"; else ko "production resolve"; fi
finish
