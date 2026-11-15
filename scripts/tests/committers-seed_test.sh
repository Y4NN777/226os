#!/usr/bin/env bash
# Tests for scripts/content/committers-seed.py: committers.top YAML (the `users` list) to the local-dev Connect seed.
# Offline: a fixture file stands in for the source. Usage: scripts/tests/committers-seed_test.sh
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
gen="$root/scripts/content/committers-seed.py"
fx="$here/fixtures/committers-bf.yml"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

rc=0
python3 -I "$gen" "BF=$fx:10" >"$tmp/a.sql" 2>"$tmp/err" || rc=$?
[ "$rc" -eq 0 ] && ok "the generator runs on a committers.top file" || ko "run (rc=$rc $(cat "$tmp/err"))"
python3 -I "$gen" "BF=$fx:10" >"$tmp/b.sql" 2>/dev/null || true
cmp -s "$tmp/a.sql" "$tmp/b.sql" && ok "the output is deterministic" || ko "two runs differ"

grep -q "'spwoodcock'" "$tmp/a.sql" && grep -q "'enok-dev'" "$tmp/a.sql" && ok "valid logins become lower-case usernames" || ko "usernames"
grep -q "'Enok-Dev'" "$tmp/a.sql" && ok "github_username keeps the login as GitHub spells it" || ko "github login case"
grep -q "O''Brien \"Dev\"" "$tmp/a.sql" && ok "quotes in names are escaped for SQL" || ko "quote escaping"
grep -q "'Sam Rémi'" "$tmp/a.sql" && ! grep -q "u00e9" "$tmp/a.sql" && ok "\\uXXXX escapes are decoded (accented names)" || ko "unicode escape"
! grep -q "'ab'" "$tmp/a.sql" && ok "a login too short for the username rule is skipped" || ko "short login kept"
! grep -q "'demo-aicha'" "$tmp/a.sql" && ok "a login that collides with a demo seed username is skipped" || ko "collision kept"
! grep -q "faso_dev" "$tmp/a.sql" && ok "a login GitHub would not allow (underscore) is skipped" || ko "faso_dev kept"
! grep -q "notused" "$tmp/a.sql" && ok "only the all-contributions list (users) is read" || ko "other list read"
grep -q "01900000-0000-7000-8000-000000000a01" "$tmp/a.sql" && ! grep -qE "8000-000000000(b|[0-9])[0-9a-f]{2}'" "$tmp/a.sql" \
  && ok "ids are in the A01-AFF block" || ko "id block"
grep -q "'demo'" "$tmp/a.sql" && ! grep -qiE "now\(\)|uuidv7\(\)|random\(\)" "$tmp/a.sql" && ok "demo provenance, literal values only" || ko "provenance or non-literal"
grep -q "ON CONFLICT (id) DO NOTHING" "$tmp/a.sql" && grep -q "ON CONFLICT (user_id) DO NOTHING" "$tmp/a.sql" && [ "$(grep -c "WHERE NOT EXISTS (SELECT 1 FROM profiles p WHERE p.username = s.username" "$tmp/a.sql")" -eq 2 ] && ok "idempotent inserts that skip taken usernames and logins" || ko "on conflict"
grep -q "Source: committers.top" "$tmp/a.sql" && grep -q "generated 2026-11-02T19:04:24Z" "$tmp/a.sql" && ok "the header names the source and its date" || ko "header"
n="$(grep -cE "^  \(.01900000" "$tmp/a.sql")"
[ "$n" -eq 2 ] && ok "two users kept from the fixture (spwoodcock, enok-dev)" || ko "count ($n)"

python3 -I "$gen" "BF=$fx:1" >"$tmp/c.sql" 2>/dev/null
[ "$(grep -cE "^  \(.01900000" "$tmp/c.sql")" -eq 1 ] && ok "the per-country limit is applied" || ko "limit"
rc=0; python3 -I "$gen" "XX=$fx:5" >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] && ok "an unknown country code is refused" || ko "unknown country accepted"
rc=0; python3 -I "$gen" "BF=$tmp/missing.yml:5" >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] && ok "a missing source file fails" || ko "missing file accepted"
finish
