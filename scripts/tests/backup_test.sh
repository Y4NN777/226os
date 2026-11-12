#!/usr/bin/env bash
# RED/GREEN evidence for scripts/backup.sh and scripts/restore-check.sh (docs/SHIP.md: consistent backup before a
# migration, restore verified). A fake `docker` answers pg_dump, pg_restore and psql, so no database is needed.
# Usage: scripts/tests/backup_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
backup="$root/scripts/backup.sh"
restore="$root/scripts/restore-check.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin" "$tmp/out"
printf 'APP_ENV=local-dev\n' >"$tmp/env.local-dev"
printf 'APP_ENV=production-live\nPUBLIC_HOST=p.example.test\nACME_EMAIL=a@example.test\n' >"$tmp/env.production-live"

[ -x "$backup" ] && [ -x "$restore" ] || { ko "scripts/backup.sh or scripts/restore-check.sh is missing or not executable"; finish; exit 1; }

# Fake docker: logs argv; FAKE_DUMP is what pg_dump prints; psql answers a table count and a migration version
# (FAKE_SRC_* for the source database, FAKE_RST_* when the call targets a restore_check_ scratch database);
# pg_restore consumes stdin; FAKE_FAIL=<word> makes any call whose argv contains the word fail.
cat >"$tmp/bin/docker" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$LOG"
if [ -n "${FAKE_FAIL:-}" ] && [[ "$*" == *"$FAKE_FAIL"* ]]; then echo "docker: simulated failure" >&2; exit 1; fi
case "$*" in
  *pg_dump*) printf '%b' "${FAKE_DUMP-PGDMPfake-archive-bytes}" ;;
  *pg_restore*) cat >/dev/null ;;
  *"CREATE DATABASE"*|*"DROP DATABASE"*) ;;
  *information_schema.tables*)
    if [[ "$*" == *restore_check_* ]]; then printf '%s\n' "${FAKE_RST_TABLES:-12}"; else printf '%s\n' "${FAKE_SRC_TABLES:-12}"; fi ;;
  *goose_db_version*)
    if [[ "$*" == *restore_check_* ]]; then printf '%s\n' "${FAKE_RST_VERSION:-10}"; else printf '%s\n' "${FAKE_SRC_VERSION:-10}"; fi ;;
esac
EOF
chmod +x "$tmp/bin/docker"

run() { # ENV... -- CMD ARGS
  : >"$tmp/docker.log"
  local envs=()
  while [ "$1" != "--" ]; do envs+=("$1"); shift; done
  shift
  rc=0
  out="$(env -i PATH="$tmp/bin:$PATH" HOME="$tmp" LOG="$tmp/docker.log" BACKUP_DIR="$tmp/out" COMPOSE_ENV_FILE="$tmp/env.local-dev" "${envs[@]}" "$@" 2>&1)" || rc=$?
}

# --- backup.sh ---
run -- "$backup" local-dev
dump="$(ls "$tmp"/out/local-dev-*.dump 2>/dev/null | head -1 || true)"
if [ "$rc" -eq 0 ] && [ -n "$dump" ] && [ -f "$dump.evidence.txt" ]; then ok "backup writes the dump and its evidence file"; else ko "backup (rc=$rc out=$out)"; fi
[ -n "$dump" ] && [ "$(stat -c %a "$dump")" = 600 ] && ok "the dump is mode 600" || ko "dump mode"
if [ -n "$dump" ] && grep -q "sha256: $(sha256sum "$dump" | cut -d' ' -f1)" "$dump.evidence.txt"; then ok "the evidence records the real sha256"; else ko "evidence sha256"; fi
grep -q 'profile: local-dev' "$dump.evidence.txt" && grep -q 'tables: 12' "$dump.evidence.txt" && grep -q 'migration_version: 10' "$dump.evidence.txt" && ok "the evidence records profile, table count and migration version" || ko "evidence content: $(cat "$dump.evidence.txt" 2>/dev/null)"
grep -q "$dump.evidence.txt" <<<"$out" && ok "the evidence path is printed for BACKUP_EVIDENCE" || ko "evidence path not printed: $out"
if grep -q 'pg_dump -Fc' "$tmp/docker.log" || grep -q 'pg_dump' "$tmp/docker.log"; then ok "pg_dump custom format is used"; else ko "no pg_dump call"; fi
grep -q -- '-Fc' "$tmp/docker.log" && ok "custom archive format (-Fc)" || ko "no -Fc"

run -- "$backup" staging
[ "$rc" -ne 0 ] && grep -qi 'profile' <<<"$out" && ok "an unknown profile is refused" || ko "bad profile (rc=$rc)"
[ ! -s "$tmp/docker.log" ] && ok "a refused profile never reaches docker" || ko "docker reached"

rm -f "$tmp"/out/*
run FAKE_DUMP= -- "$backup" local-dev
if [ "$rc" -ne 0 ] && [ -z "$(ls -A "$tmp/out")" ] && grep -qi 'empty' <<<"$out"; then ok "an empty dump is refused and nothing is left behind"; else ko "empty dump (rc=$rc, left: $(ls -A "$tmp/out"))"; fi

rm -f "$tmp"/out/*
run FAKE_DUMP="not a pg archive" -- "$backup" local-dev
if [ "$rc" -ne 0 ] && [ -z "$(ls -A "$tmp/out")" ] && grep -qi 'archive' <<<"$out"; then ok "output that is not a pg_dump archive is refused, nothing left"; else ko "non-archive (rc=$rc, left: $(ls -A "$tmp/out"))"; fi

rm -f "$tmp"/out/*
run FAKE_FAIL=pg_dump -- "$backup" local-dev
if [ "$rc" -ne 0 ] && [ -z "$(ls -A "$tmp/out")" ]; then ok "a failing pg_dump leaves nothing behind"; else ko "pg_dump failure (rc=$rc, left: $(ls -A "$tmp/out"))"; fi

# --- restore-check.sh ---
rm -f "$tmp"/out/*
run -- "$backup" local-dev
dump="$(ls "$tmp"/out/local-dev-*.dump | head -1)"

run -- "$restore" local-dev "$dump"
if [ "$rc" -eq 0 ] && grep -q 'restore check: PASS' <<<"$out"; then ok "restore check passes when the restored database matches the source"; else ko "restore PASS (rc=$rc out=$out)"; fi
grep -q 'CREATE DATABASE "restore_check_' "$tmp/docker.log" && ok "the restore goes into a scratch database" || ko "no scratch database created"
grep -q 'DROP DATABASE "restore_check_' "$tmp/docker.log" && ok "the scratch database is dropped afterwards" || ko "scratch database not dropped"
if grep 'DROP DATABASE' "$tmp/docker.log" | grep -qv 'restore_check_'; then ko "a DROP DATABASE targeted something other than the scratch database"; else ok "only restore_check_* databases are ever dropped"; fi

run FAKE_RST_TABLES=11 -- "$restore" local-dev "$dump"
[ "$rc" -ne 0 ] && grep -q 'restore check: FAIL' <<<"$out" && ok "a table count mismatch fails the restore check" || ko "table mismatch (rc=$rc out=$out)"
grep -q 'DROP DATABASE "restore_check_' "$tmp/docker.log" && ok "the scratch database is dropped even when the check fails" || ko "scratch kept after failure"

run FAKE_RST_VERSION=9 -- "$restore" local-dev "$dump"
[ "$rc" -ne 0 ] && grep -q 'migration' <<<"$out" && ok "a migration version mismatch fails the restore check" || ko "version mismatch (rc=$rc out=$out)"

run -- "$restore" local-dev "$tmp/nope.dump"
[ "$rc" -ne 0 ] && ok "a missing dump is refused" || ko "missing dump"
printf 'junk' >"$tmp/junk.dump"
run -- "$restore" local-dev "$tmp/junk.dump"
[ "$rc" -ne 0 ] && grep -qi 'archive' <<<"$out" && ok "a file that is not a pg_dump archive is refused" || ko "junk dump (rc=$rc)"
ln -s "$dump" "$tmp/link.dump"
run -- "$restore" local-dev "$tmp/link.dump"
[ "$rc" -ne 0 ] && ok "a symlink dump is refused" || ko "symlink dump"

run COMPOSE_ENV_FILE="$tmp/env.production-live" -- "$restore" production-live "$dump"
[ "$rc" -ne 0 ] && grep -q 'RESTORE_CHECK_CONFIRM' <<<"$out" && ok "restore check on production needs an explicit confirmation" || ko "production confirmation (rc=$rc out=$out)"
[ ! -s "$tmp/docker.log" ] && ok "refused before docker is reached" || ko "docker reached for unconfirmed production"

finish
