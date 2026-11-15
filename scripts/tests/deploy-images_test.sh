#!/usr/bin/env bash
# Tests for scripts/deploy-images.sh with a fake `docker` that records its argv and answers `buildx imagetools
# inspect` from a map: push, resolve (digests come from the registry, never from local IDs) and tag by digest.
# Usage: scripts/tests/deploy-images_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
images="$root/scripts/deploy-images.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin"
d_api="sha256:$(printf 'a%.0s' {1..64})"
d_web="sha256:$(printf 'b%.0s' {1..64})"
d_mig="sha256:$(printf 'c%.0s' {1..64})"
cat >"$tmp/bin/docker" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$FAKE_DOCKER_LOG"
case "$1 $2" in
  "image inspect")
    case "$*" in
      *RepoDigests*) grep -F "$3 " "$FAKE_DOCKER_REPO" | head -n 1 | cut -d' ' -f2; exit 0 ;;
    esac
    grep -F "$3 " "$FAKE_DOCKER_IDS" | head -n 1 | cut -d' ' -f2; exit 0 ;;
esac
case "$1 $2 $3" in
  "buildx imagetools inspect")
    if [ "$4" = --raw ]; then
      printf '{"config":{"digest":"%s"}}\n' "$(grep -F "$5 " "$FAKE_DOCKER_CFG" | head -n 1 | cut -d' ' -f2)"
      exit 0
    fi
    grep -F "$4 " "$FAKE_DOCKER_MAP" | head -n 1 | cut -d' ' -f2 | grep . || { echo "not found: $4" >&2; exit 1; }
    ;;
esac
exit "${FAKE_DOCKER_RC:-0}"
EOF
chmod +x "$tmp/bin/docker"

ns=ghcr.io/y4nn777
cat >"$tmp/map" <<EOF
$ns/226os-api:tested-t1 $d_api
$ns/226os-web:tested-t1 $d_web
$ns/226os-migrate:tested-t1 $d_mig
$ns/226os-api:abc123 $d_api
$ns/226os-web:abc123 $d_web
$ns/226os-migrate:abc123 bad-digest
EOF

run() { : >"$tmp/log"; rc=0; out="$(env PATH="$tmp/bin:$PATH" FAKE_DOCKER_LOG="$tmp/log" FAKE_DOCKER_MAP="$tmp/map" FAKE_DOCKER_IDS="$tmp/ids" FAKE_DOCKER_CFG="$tmp/cfg" FAKE_DOCKER_REPO="$tmp/repo" "$@" 2>&1)" || rc=$?; }
# local image ids, and the config digest the registry holds for the existing tags
cat >"$tmp/ids" <<EOM
226os-api:localtag sha256:$(printf '1%.0s' {1..64})
226os-web:localtag sha256:$(printf '2%.0s' {1..64})
226os-migrate:localtag sha256:$(printf '3%.0s' {1..64})
EOM
: >"$tmp/cfg"
: >"$tmp/repo"

run "$images" resolve tested-t1
if [ "$rc" -eq 0 ] && grep -qx "API_IMAGE=$ns/226os-api@$d_api" <<<"$out" && grep -qx "WEB_IMAGE=$ns/226os-web@$d_web" <<<"$out" &&
  grep -qx "MIGRATE_IMAGE=$ns/226os-migrate@$d_mig" <<<"$out"; then ok "resolve prints the three images by digest"; else ko "resolve (rc=$rc out=$out)"; fi

run "$images" resolve missing
[ "$rc" -ne 0 ] && grep -q 'no image' <<<"$out" && ok "a tag that is not in the registry fails with a clear message" || ko "missing tag (rc=$rc out=$out)"

run "$images" resolve abc123
[ "$rc" -ne 0 ] && grep -q 'digest' <<<"$out" && ok "a malformed digest from the registry is refused" || ko "bad digest (rc=$rc out=$out)"

run "$images" resolve 'x;y'
[ "$rc" -eq 64 ] && ok "a tag with shell characters is refused" || ko "bad tag (rc=$rc)"

run IMAGE_NAMESPACE='ghcr.io/Y4NN777' "$images" resolve tested-t1
[ "$rc" -eq 64 ] && ok "an upper-case namespace is refused (registries are lower case)" || ko "namespace (rc=$rc)"

# tag works from digests (a file of IMAGE=ns/226os-x@sha256:...), never from a tag: nothing is read from the registry.
cat >"$tmp/by-digest.env" <<EOM
API_IMAGE=$ns/226os-api@$d_api
WEB_IMAGE=$ns/226os-web@$d_web
MIGRATE_IMAGE=$ns/226os-migrate@$d_mig
EOM
run "$images" tag tested-tree "$tmp/by-digest.env"
if [ "$rc" -eq 0 ] && grep -q "buildx imagetools create --tag $ns/226os-api:tested-tree $ns/226os-api@$d_api" "$tmp/log" &&
  grep -q "create --tag $ns/226os-migrate:tested-tree $ns/226os-migrate@$d_mig" "$tmp/log" && ! grep -q 'imagetools inspect' "$tmp/log"; then ok "tag re-tags the given digests and never re-reads a tag"; else ko "tag (rc=$rc out=$out log=$(cat "$tmp/log"))"; fi
printf 'API_IMAGE=%s/226os-api:abc123\nWEB_IMAGE=%s/226os-web@%s\nMIGRATE_IMAGE=%s/226os-migrate@%s\n' "$ns" "$ns" "$d_web" "$ns" "$d_mig" >"$tmp/bad.env"
run "$images" tag tested-tree "$tmp/bad.env"
[ "$rc" -ne 0 ] && ! grep -q 'imagetools create' "$tmp/log" && ok "tag refuses an image given by tag, and tags nothing" || ko "tag by tag (rc=$rc log=$(cat "$tmp/log"))"
printf 'API_IMAGE=%s/226os-web@%s\nWEB_IMAGE=%s/226os-web@%s\nMIGRATE_IMAGE=%s/226os-migrate@%s\n' "$ns" "$d_api" "$ns" "$d_web" "$ns" "$d_mig" >"$tmp/bad.env"
run "$images" tag tested-tree "$tmp/bad.env"
[ "$rc" -ne 0 ] && ! grep -q 'imagetools create' "$tmp/log" && ok "tag refuses an image of another name than its key" || ko "tag wrong name (rc=$rc)"
grep -v '^WEB_IMAGE=' "$tmp/by-digest.env" >"$tmp/bad.env"
run "$images" tag tested-tree "$tmp/bad.env"
[ "$rc" -ne 0 ] && ! grep -q 'imagetools create' "$tmp/log" && ok "tag refuses a file with a missing image" || ko "tag missing (rc=$rc)"

# push prints the digests of what THIS run pushed (read from the local RepoDigests, set by the push itself), never by
# re-reading the tag. A tag that exists with other content fails; one with the same content is accepted.
d_la="sha256:$(printf '1%.0s' {1..64})"
printf '%s %s@%s\n%s %s@%s\n%s %s@%s\n' "$ns/226os-api:fresh1" "$ns/226os-api" "$d_api" "$ns/226os-web:fresh1" "$ns/226os-web" "$d_web" "$ns/226os-migrate:fresh1" "$ns/226os-migrate" "$d_mig" >"$tmp/repo"
: >"$tmp/map"
: >"$tmp/cfg"
run "$images" push fresh1 localtag
if [ "$rc" -eq 0 ] && grep -qx "tag 226os-api:localtag $ns/226os-api:fresh1" "$tmp/log" && grep -qx "push $ns/226os-web:fresh1" "$tmp/log" &&
  grep -qx "API_IMAGE=$ns/226os-api@$d_api" <<<"$out" && grep -qx "WEB_IMAGE=$ns/226os-web@$d_web" <<<"$out" && grep -qx "MIGRATE_IMAGE=$ns/226os-migrate@$d_mig" <<<"$out"; then
  ok "push tags, pushes and prints the digests it pushed"
else ko "push (rc=$rc out=$out log=$(cat "$tmp/log"))"; fi
! grep -q 'imagetools inspect.*fresh1.*--format' "$tmp/log" && ok "push never reads its digests back from the tag" || ko "push re-read the tag"

# containerd image store: the local Id is the manifest digest, not the config digest. An existing tag whose manifest
# digest equals the local Id is the same image: kept. One whose manifest and config both differ still fails (above/below).
cat >>"$tmp/ids" <<EOM
226os-api:ctrtag $d_api
226os-web:ctrtag sha256:$(printf '2%.0s' {1..64})
226os-migrate:ctrtag sha256:$(printf '3%.0s' {1..64})
EOM
printf '%s %s\n' "$ns/226os-api:ctr1" "$d_api" >"$tmp/map"
printf '%s sha256:%s\n' "$ns/226os-api:ctr1" "$(printf '8%.0s' {1..64})" >"$tmp/cfg"
printf '%s %s@%s\n%s %s@%s\n' "$ns/226os-web:ctr1" "$ns/226os-web" "$d_web" "$ns/226os-migrate:ctr1" "$ns/226os-migrate" "$d_mig" >"$tmp/repo"
run "$images" push ctr1 ctrtag
if [ "$rc" -eq 0 ] && grep -qx "API_IMAGE=$ns/226os-api@$d_api" <<<"$out" && ! grep -qx "push $ns/226os-api:ctr1" "$tmp/log"; then
  ok "containerd store: an existing tag whose manifest digest is the local Id is the same image, kept"
else ko "containerd (rc=$rc out=$out log=$(cat "$tmp/log"))"; fi

# a tag pushed by someone else BEFORE the build (other content): the build fails, nothing is pushed or printed.
printf '%s %s\n' "$ns/226os-api:abc123" "$d_web" >"$tmp/map"
printf '%s sha256:%s\n' "$ns/226os-api:abc123" "$(printf '9%.0s' {1..64})" >"$tmp/cfg"
run "$images" push abc123 localtag
if [ "$rc" -ne 0 ] && grep -q 'immutable' <<<"$out" && ! grep -q '^push' "$tmp/log" && ! grep -q '_IMAGE=' <<<"$out"; then
  ok "a pre-existing <sha> tag with other content fails the build (never kept silently)"
else ko "pre-existing tag (rc=$rc out=$out log=$(cat "$tmp/log"))"; fi

# a re-run whose image is identical to the one already pushed: accepted, with the registry digest of that tag.
printf '%s %s\n' "$ns/226os-api:same1" "$d_api" >"$tmp/map"
printf '%s %s\n' "$ns/226os-api:same1" "$d_la" >"$tmp/cfg"
printf '%s %s@%s\n%s %s@%s\n' "$ns/226os-web:same1" "$ns/226os-web" "$d_web" "$ns/226os-migrate:same1" "$ns/226os-migrate" "$d_mig" >"$tmp/repo"
run "$images" push same1 localtag
if [ "$rc" -eq 0 ] && grep -qx "API_IMAGE=$ns/226os-api@$d_api" <<<"$out" && ! grep -qx "push $ns/226os-api:same1" "$tmp/log" &&
  grep -qx "push $ns/226os-web:same1" "$tmp/log" && grep -qx "MIGRATE_IMAGE=$ns/226os-migrate@$d_mig" <<<"$out"; then
  ok "an existing tag with the same content is kept; the missing ones are pushed"
else ko "same content (rc=$rc out=$out log=$(cat "$tmp/log"))"; fi

# Emergency tag (deploy-emergency.yml): emergency-<sha>-<run id>-<attempt>, under the same immutability rule.
etag="emergency-$(printf 'a%.0s' {1..40})-123456789-1"
printf '%s %s@%s\n%s %s@%s\n%s %s@%s\n' "$ns/226os-api:$etag" "$ns/226os-api" "$d_api" "$ns/226os-web:$etag" "$ns/226os-web" "$d_web" "$ns/226os-migrate:$etag" "$ns/226os-migrate" "$d_mig" >"$tmp/repo"
: >"$tmp/map"
: >"$tmp/cfg"
run "$images" push "$etag" localtag
if [ "$rc" -eq 0 ] && grep -qx "push $ns/226os-api:$etag" "$tmp/log" && grep -qx "API_IMAGE=$ns/226os-api@$d_api" <<<"$out" &&
  grep -qx "MIGRATE_IMAGE=$ns/226os-migrate@$d_mig" <<<"$out"; then
  ok "an emergency-<sha>-<run>-<attempt> tag is pushed and its digests printed"
else ko "emergency push (rc=$rc out=$out log=$(cat "$tmp/log"))"; fi

printf '%s %s\n' "$ns/226os-api:$etag" "$d_web" >"$tmp/map"
printf '%s sha256:%s\n' "$ns/226os-api:$etag" "$(printf '9%.0s' {1..64})" >"$tmp/cfg"
run "$images" push "$etag" localtag
if [ "$rc" -ne 0 ] && grep -q 'immutable' <<<"$out" && ! grep -q '^push' "$tmp/log" && ! grep -q '_IMAGE=' <<<"$out"; then
  ok "a pre-existing emergency tag with other content fails too (same immutability rule)"
else ko "emergency pre-existing (rc=$rc out=$out log=$(cat "$tmp/log"))"; fi

# a push that leaves no registry digest locally is an error, not a guess.
: >"$tmp/map"
: >"$tmp/cfg"
: >"$tmp/repo"
run "$images" push nodigest localtag
[ "$rc" -ne 0 ] && ! grep -q '_IMAGE=' <<<"$out" && ok "no RepoDigest after a push: the build fails" || ko "no repo digest (rc=$rc out=$out)"

run "$images"
[ "$rc" -eq 64 ] && ok "no command is a usage error" || ko "usage (rc=$rc)"

finish
