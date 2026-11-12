#!/usr/bin/env bash
# Restore drill: load a backup into a scratch database next to the real one, compare table count and
# migration version with the source, always drop the scratch database. Never touches the source database.
# Usage: scripts/restore-check.sh <profile> <dump>
# Outside local-dev it needs RESTORE_CHECK_CONFIRM=<profile> (it still never writes to the real database).
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
profile="${1-}"
dump="${2-}"
refuse() { printf 'REFUSED: restore-check %s\n' "$*" >&2; exit 2; }
case "$profile" in
  local-dev) ;;
  staging-test | production-live) [ "${RESTORE_CHECK_CONFIRM-}" = "$profile" ] || refuse "outside local-dev needs RESTORE_CHECK_CONFIRM=${profile}" ;;
  *) refuse "unknown profile '${profile}'" ;;
esac
[ -n "$dump" ] && [ -f "$dump" ] && [ ! -L "$dump" ] || refuse "dump must be an existing regular file (not a symlink)"
[ "$(head -c 5 "$dump")" = "PGDMP" ] || refuse "not a pg_dump archive: $dump"

scratch="restore_check_$(date -u +%Y%m%d%H%M%S)_$$"
compose() { "$root/scripts/compose.sh" "$profile" "$@"; }
admin() { compose exec -T postgres sh -c "psql -X -q -v ON_ERROR_STOP=1 -U \"\$POSTGRES_USER\" -d postgres -c '$1'"; }
value() { # <database-expr> <sql>
  compose exec -T postgres sh -c "psql -X -At -U \"\$POSTGRES_USER\" -d $1 -c \"$2\""
}
created=0
cleanup() { [ "$created" -eq 1 ] && admin "DROP DATABASE \"$scratch\"" >/dev/null 2>&1 || true; }
trap cleanup EXIT

admin "CREATE DATABASE \"$scratch\""
created=1
compose exec -T postgres sh -c "pg_restore -U \"\$POSTGRES_USER\" -d $scratch --no-owner --exit-on-error" <"$dump"

tabq="SELECT count(*) FROM information_schema.tables WHERE table_schema='public'"
verq='SELECT coalesce(max(version_id),0) FROM goose_db_version'
src_tables="$(value '"$POSTGRES_DB"' "$tabq")"
rst_tables="$(value "$scratch" "$tabq")"
src_version="$(value '"$POSTGRES_DB"' "$verq" 2>/dev/null || echo none)"
rst_version="$(value "$scratch" "$verq" 2>/dev/null || echo none)"

printf 'source:   tables=%s migration_version=%s\nrestored: tables=%s migration_version=%s\n' "$src_tables" "$src_version" "$rst_tables" "$rst_version"
[ "$src_tables" = "$rst_tables" ] || { printf 'restore check: FAIL (table count differs)\n' >&2; exit 1; }
[ "$src_version" = "$rst_version" ] || { printf 'restore check: FAIL (migration version differs)\n' >&2; exit 1; }
printf 'restore check: PASS (%s)\n' "$dump"
