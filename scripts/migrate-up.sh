#!/usr/bin/env bash
# Forward-only goose migration through the Compose `migrate` jobs (ADR-0005, ADR-0006).
#   local-dev:                 runs the `migrate` service.
#   staging-test / production: refused unless MIGRATE_CONFIRM equals the profile AND
#                              BACKUP_EVIDENCE is the path of an existing, non-empty,
#                              non-symlink regular file (the recorded backup evidence); then runs
#                              the gated `migrate-gated` service. Never run production from a laptop.
# Usage: scripts/migrate-up.sh <profile>
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
profile="${1-}"
refuse() { printf 'REFUSED: migrate-up %s\n' "$*" >&2; exit 2; }

case "$profile" in
  local-dev)
    exec "$root/scripts/compose.sh" local-dev run --rm migrate
    ;;
  staging-test | production-live)
    [ "${MIGRATE_CONFIRM-}" = "$profile" ] || refuse "outside local-dev needs MIGRATE_CONFIRM=${profile}"
    evidence="${BACKUP_EVIDENCE-}"
    if [ -z "$evidence" ] || [ -L "$evidence" ] || [ ! -f "$evidence" ] || [ ! -s "$evidence" ]; then
      refuse "outside local-dev needs BACKUP_EVIDENCE=<path to an existing non-empty regular file recording the backup>"
    fi
    printf 'migrate-up: profile=%s backup evidence=%s\n' "$profile" "$evidence"
    # --profile migrate-gated is added on top of the wrapper's profile.
    exec "$root/scripts/compose.sh" "$profile" --profile migrate-gated run --rm migrate-gated
    ;;
  *) refuse "unknown profile '${profile}'" ;;
esac
