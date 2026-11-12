#!/usr/bin/env bash
# Consistent logical backup of the profile's PostgreSQL (pg_dump custom format) plus an evidence file
# usable as BACKUP_EVIDENCE for scripts/migrate-up.sh (docs/SHIP.md). Read-only on the database.
# Usage: scripts/backup.sh <profile>      Env: BACKUP_DIR (default: backups/, git-ignored)
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
profile="${1-}"
case "$profile" in
  local-dev | staging-test | production-live) ;;
  *) printf 'REFUSED: unknown profile %q (local-dev | staging-test | production-live)\n' "$profile" >&2; exit 2 ;;
esac
dir="${BACKUP_DIR:-$root/backups}"
umask 077
mkdir -p "$dir"
stamp="$(date -u +%Y%m%dT%H%M%SZ)"
dump="$dir/$profile-$stamp.dump"
tmp="$(mktemp "$dir/.backup.XXXXXX")"
trap 'rm -f "$tmp"' EXIT

compose() { "$root/scripts/compose.sh" "$profile" "$@"; }
psql_value() { compose exec -T postgres sh -c "psql -X -At -U \"\$POSTGRES_USER\" -d \"\$POSTGRES_DB\" -c \"$1\""; }

compose exec -T postgres sh -c 'pg_dump -Fc -U "$POSTGRES_USER" -d "$POSTGRES_DB"' >"$tmp" ||
  { printf 'REFUSED: pg_dump failed; nothing written\n' >&2; exit 1; }
[ -s "$tmp" ] || { printf 'REFUSED: pg_dump produced an empty dump; nothing written\n' >&2; exit 1; }
[ "$(head -c 5 "$tmp")" = "PGDMP" ] || { printf 'REFUSED: output is not a pg_dump archive; nothing written\n' >&2; exit 1; }

tables="$(psql_value "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'")"
version="$(psql_value 'SELECT coalesce(max(version_id),0) FROM goose_db_version' 2>/dev/null || echo none)"
sha="$(sha256sum "$tmp" | cut -d' ' -f1)"

mv "$tmp" "$dump"
trap - EXIT
evidence="$dump.evidence.txt"
{
  printf 'backup: %s\n' "$(basename "$dump")"
  printf 'profile: %s\n' "$profile"
  printf 'created_utc: %s\n' "$stamp"
  printf 'sha256: %s\n' "$sha"
  printf 'bytes: %s\n' "$(stat -c %s "$dump")"
  printf 'tables: %s\n' "$tables"
  printf 'migration_version: %s\n' "$version"
} >"$evidence"
printf 'backup: %s\nBACKUP_EVIDENCE=%s\n' "$dump" "$evidence"
