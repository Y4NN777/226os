#!/usr/bin/env bash
# Build the env file a server receives: the decrypted profile env (stdin) with the three images set by digest (stdout).
#
# Usage: scripts/deploy-env.sh <staging-test | production-live> <public-host> < decrypted.env > deploy.env
# Env:   API_IMAGE, WEB_IMAGE, MIGRATE_IMAGE: <registry>/<owner>/226os-<api|web|migrate>@sha256:<64 hex> (by digest only).
# The existing *_IMAGE lines (the local-dev placeholders of .env.example) are replaced, not duplicated. The result must
# pass scripts/compose-env-guard.sh (APP_ENV once and equal to the profile, no CHANGE_ME, a real PUBLIC_HOST and
# ACME_EMAIL), and its PUBLIC_HOST must equal the PUBLIC_HOST variable of the GitHub environment (a wrong host for
# the environment is refused before anything is sent). Errors name variables, never print values.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
usage() { printf 'usage: %s <staging-test | production-live> <public-host>  (env on stdin, API_IMAGE WEB_IMAGE MIGRATE_IMAGE in the environment)\n' "$0" >&2; exit 64; }
refuse() { printf 'deploy-env: REFUSED: %s\n' "$*" >&2; exit 2; }

[ $# -eq 2 ] || usage
profile="$1"
host="$2"
case "$profile" in staging-test | production-live) ;; *) usage ;; esac

for pair in "API_IMAGE:api" "WEB_IMAGE:web" "MIGRATE_IMAGE:migrate"; do
  var="${pair%%:*}"
  name="${pair##*:}"
  value="${!var-}"
  [ -n "$value" ] || refuse "$var is not set"
  [[ "$value" =~ ^[a-z0-9.-]+(:[0-9]+)?(/[a-z0-9._-]+)+/226os-$name@sha256:[0-9a-f]{64}$ ]] ||
    refuse "$var must be <registry>/<owner>/226os-$name@sha256:<digest> (an image by digest, never by tag)"
done

# The plaintext env goes to RAM when there is any (/dev/shm), like scripts/compose.sh and scripts/secrets.sh.
base="${TMPDIR:-/tmp}"
[ -d /dev/shm ] && [ -w /dev/shm ] && base=/dev/shm
tmp="$(mktemp "$base/deploy-env.XXXXXX")"
trap 'rm -f "$tmp"' EXIT
chmod 600 "$tmp"
{
  grep -vE '^(API_IMAGE|WEB_IMAGE|MIGRATE_IMAGE)=' || true
  printf 'API_IMAGE=%s\nWEB_IMAGE=%s\nMIGRATE_IMAGE=%s\n' "$API_IMAGE" "$WEB_IMAGE" "$MIGRATE_IMAGE"
} >"$tmp"

"$root/scripts/compose-env-guard.sh" "$profile" "$tmp" >/dev/null || refuse "the env file of $profile does not pass scripts/compose-env-guard.sh"
grep -qxF "PUBLIC_HOST=$host" "$tmp" || refuse "PUBLIC_HOST in the env file differs from the PUBLIC_HOST variable of the environment"
cat "$tmp"
