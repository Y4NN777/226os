#!/usr/bin/env bash
# Tests for scripts/deploy-env.sh: the decrypted env (stdin) becomes the deploy env (stdout) with the three images set
# by digest, existing *_IMAGE lines replaced (not duplicated), and every mismatch refused before anything is sent to a
# server. Values are never printed in errors.
# Usage: scripts/tests/deploy-env_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
denv="$root/scripts/deploy-env.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
ns=ghcr.io/y4nn777
da="sha256:$(printf 'a%.0s' {1..64})"
dw="sha256:$(printf 'b%.0s' {1..64})"
dm="sha256:$(printf 'c%.0s' {1..64})"
export API_IMAGE="$ns/226os-api@$da" WEB_IMAGE="$ns/226os-web@$dw" MIGRATE_IMAGE="$ns/226os-migrate@$dm"

mkenv() { # mkenv PROFILE HOST
  cat >"$tmp/in.env" <<EOF
APP_ENV=$1
API_IMAGE=226os-api:dev
WEB_IMAGE=226os-web:dev
MIGRATE_IMAGE=226os-migrate:dev
POSTGRES_PASSWORD=hunter2-not-a-real-secret
PUBLIC_HOST=$2
ACME_EMAIL=ops@example.test
EOF
}
run() { rc=0; out="$("$denv" "$@" <"$tmp/in.env" 2>"$tmp/err")" || rc=$?; err="$(cat "$tmp/err")"; }

mkenv staging-test staging.example.test
run staging-test staging.example.test
if [ "$rc" -eq 0 ] && grep -qx "API_IMAGE=$API_IMAGE" <<<"$out" && grep -qx "WEB_IMAGE=$WEB_IMAGE" <<<"$out" &&
  grep -qx "MIGRATE_IMAGE=$MIGRATE_IMAGE" <<<"$out" && grep -qx 'POSTGRES_PASSWORD=hunter2-not-a-real-secret' <<<"$out"; then
  ok "the images are set by digest and the other values pass through"
else ko "staging env (rc=$rc err=$err)"; fi
[ "$(grep -c '_IMAGE=' <<<"$out")" -eq 3 ] && ! grep -q ':dev' <<<"$out" && ok "the placeholder image lines are replaced, not duplicated" || ko "duplicates"

mkenv production-live staging.example.test
run production-live prod.example.test
[ "$rc" -ne 0 ] && grep -q 'PUBLIC_HOST' <<<"$err" && ! grep -q 'example.test' <<<"$err" && ok "a PUBLIC_HOST that differs from the environment variable is refused, values not printed" || ko "host mismatch (rc=$rc err=$err)"

mkenv production-live prod.example.test
run staging-test prod.example.test
[ "$rc" -ne 0 ] && ok "an env file whose APP_ENV is another profile is refused" || ko "profile mismatch (rc=$rc)"

mkenv staging-test staging.example.test
sed -i 's/^POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=CHANGE_ME/' "$tmp/in.env"
run staging-test staging.example.test
[ "$rc" -ne 0 ] && ok "a CHANGE_ME placeholder is refused" || ko "placeholder (rc=$rc)"

mkenv staging-test staging.example.test
run local-dev staging.example.test
[ "$rc" -eq 64 ] && ok "local-dev is not a deploy profile" || ko "local-dev (rc=$rc)"

mkenv staging-test staging.example.test
API_IMAGE="$ns/226os-api:latest" run staging-test staging.example.test
[ "$rc" -ne 0 ] && grep -q 'API_IMAGE' <<<"$err" && ok "an image by tag is refused" || ko "tag image (rc=$rc err=$err)"

API_IMAGE="$ns/226os-web@$da" run staging-test staging.example.test
[ "$rc" -ne 0 ] && ok "an image of the wrong name in a variable is refused" || ko "wrong name (rc=$rc)"

mkenv staging-test staging.example.test
env -u MIGRATE_IMAGE "$denv" staging-test staging.example.test <"$tmp/in.env" >/dev/null 2>&1 && ko "missing image variable passed" || ok "a missing image variable is refused"

run staging-test
[ "$rc" -eq 64 ] && ok "a missing host argument is a usage error" || ko "usage (rc=$rc)"

finish
