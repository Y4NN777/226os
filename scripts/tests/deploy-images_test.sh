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
  "image inspect") grep -F "$3 " "$FAKE_DOCKER_IDS" | head -n 1 | cut -d' ' -f2; exit 0 ;;
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

run() { : >"$tmp/log"; rc=0; out="$(env PATH="$tmp/bin:$PATH" FAKE_DOCKER_LOG="$tmp/log" FAKE_DOCKER_MAP="$tmp/map" FAKE_DOCKER_IDS="$tmp/ids" FAKE_DOCKER_CFG="$tmp/cfg" "$@" 2>&1)" || rc=$?; }
# local image ids, and the config digest the registry holds for the existing tags
cat >"$tmp/ids" <<EOM
226os-api:localtag sha256:$(printf '1%.0s' {1..64})
226os-web:localtag sha256:$(printf '2%.0s' {1..64})
226os-migrate:localtag sha256:$(printf '3%.0s' {1..64})
EOM
: >"$tmp/cfg"

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

run "$images" tag tested-t1 tested-tree
if [ "$rc" -eq 0 ] && grep -q "buildx imagetools create --tag $ns/226os-api:tested-tree $ns/226os-api@$d_api" "$tmp/log" &&
  grep -q "create --tag $ns/226os-migrate:tested-tree $ns/226os-migrate@$d_mig" "$tmp/log"; then ok "tag re-tags the digests, it never rebuilds"; else ko "tag (rc=$rc out=$out log=$(cat "$tmp/log"))"; fi

# abc123: api and web exist, with other content than the local images: refused, nothing pushed.
printf '%s sha256:%s\n%s sha256:%s\n' "$ns/226os-api:abc123" "$(printf '9%.0s' {1..64})" "$ns/226os-web:abc123" "$(printf '9%.0s' {1..64})" >"$tmp/cfg"
run "$images" push abc123 localtag
if [ "$rc" -ne 0 ] && grep -q 'different content' <<<"$out" && ! grep -q '^push' "$tmp/log"; then ok "push refuses an existing tag that holds other content (immutable tags)"; else ko "existing tag (rc=$rc out=$out)"; fi

# Partial push: api already holds exactly the local image, web and migrate are missing: skip api, push the rest.
: >"$tmp/map"
printf '%s %s\n' "$ns/226os-api:part1" "$d_api" >"$tmp/map"
printf '%s sha256:%s\n' "$ns/226os-api:part1" "$(printf '1%.0s' {1..64})" >"$tmp/cfg"
run "$images" push part1 localtag
if [ "$rc" -eq 0 ] && grep -q 'already holds this image' <<<"$out" && ! grep -q "push $ns/226os-api:part1" "$tmp/log" &&
  grep -qx "push $ns/226os-web:part1" "$tmp/log" && grep -qx "push $ns/226os-migrate:part1" "$tmp/log"; then
  ok "a partial push is completed: the tag with the same image is skipped, the others are pushed"
else ko "partial push (rc=$rc out=$out log=$(cat "$tmp/log"))"; fi
cat >"$tmp/map" <<EOM
$ns/226os-api:tested-t1 $d_api
$ns/226os-web:tested-t1 $d_web
$ns/226os-migrate:tested-t1 $d_mig
EOM
: >"$tmp/cfg"
run "$images" push fresh1 localtag
if [ "$rc" -eq 0 ] && grep -qx "tag 226os-api:localtag $ns/226os-api:fresh1" "$tmp/log" && grep -qx "push $ns/226os-web:fresh1" "$tmp/log"; then
  ok "push tags the local image and pushes it under the commit tag"
else ko "push (rc=$rc log=$(cat "$tmp/log"))"; fi

run "$images"
[ "$rc" -eq 64 ] && ok "no command is a usage error" || ko "usage (rc=$rc)"

finish
