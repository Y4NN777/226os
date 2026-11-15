#!/usr/bin/env bash
# Usage: scripts/deploy-check-bound.sh <profile> <bound-profile-read-from-the-server>
# A server belongs to the profile of its first deploy (state/profile). An empty value is a first deploy: continue.
# Any other value than <profile> exits 2, before the workflow copies a release or an env file to that server.
set -euo pipefail
[ $# -eq 2 ] || { printf 'usage: %s <profile> <bound-profile>\n' "$0" >&2; exit 64; }
profile="$1"
bound="$2"
case "$profile" in staging-test | production-live) ;; *) printf 'deploy-check-bound: REFUSED: unknown profile\n' >&2; exit 2 ;; esac
[ -z "$bound" ] && { printf 'deploy-check-bound: no profile recorded yet: first deploy of this server\n'; exit 0; }
if [ "$bound" != "$profile" ]; then
  printf 'deploy-check-bound: REFUSED: DEPLOY_HOST belongs to profile %s, not %s: wrong host for this environment? Nothing was copied.\n' "$bound" "$profile" >&2
  exit 2
fi
printf 'deploy-check-bound: server is bound to %s\n' "$profile"
