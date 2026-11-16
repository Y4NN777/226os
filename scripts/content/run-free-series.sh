#!/usr/bin/env bash
# Runs the one-time free Learn series script (scripts/content/free-series.sql, S4, B5/H19) on a profile.
#   local-dev:                 runs it directly (rehearsal).
#   staging-test / production: refused unless CONTENT_CONFIRM equals the profile AND BACKUP_EVIDENCE is the
#                              path of an existing, non-empty, non-symlink regular file (scripts/backup.sh
#                              prints it), as for migrate-up. Run on staging-test first, then production-live.
# OWNER_USERNAME is the username of the real account the content is attributed to (looked up in SQL).
# The SQL is one DO block: it writes everything or nothing, and refuses a second run.
# Usage: OWNER_USERNAME=<username> scripts/content/run-free-series.sh <profile>
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
profile="${1-}"
refuse() { printf 'REFUSED: free-series %s\n' "$*" >&2; exit 2; }

owner="${OWNER_USERNAME-}"
# Same rule as profiles_username_ck; it also keeps the value safe inside PGOPTIONS.
[[ "$owner" =~ ^[a-z0-9][a-z0-9_-]{2,29}$ ]] || refuse "needs OWNER_USERNAME=<username of the owner's real account>"

case "$profile" in
  local-dev) ;;
  staging-test | production-live)
    [ "${CONTENT_CONFIRM-}" = "$profile" ] || refuse "outside local-dev needs CONTENT_CONFIRM=${profile}"
    evidence="${BACKUP_EVIDENCE-}"
    if [ -z "$evidence" ] || [ -L "$evidence" ] || [ ! -f "$evidence" ] || [ ! -s "$evidence" ]; then
      refuse "outside local-dev needs BACKUP_EVIDENCE=<path to an existing non-empty regular file recording the backup>"
    fi
    printf 'free-series: profile=%s owner=%s backup evidence=%s\n' "$profile" "$owner" "$evidence"
    ;;
  *) refuse "unknown profile '${profile}'" ;;
esac

exec "$root/scripts/compose.sh" "$profile" exec -T -e "PGOPTIONS=-c content.owner_username=${owner}" postgres \
  sh -c 'psql -X -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$POSTGRES_DB"' <"$root/scripts/content/free-series.sql"
