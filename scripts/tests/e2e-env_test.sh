#!/usr/bin/env bash
# Tests for scripts/e2e-env.sh: from a decrypted staging-test env (stdin) keep only E2E_USER and E2E_PASSWORD.
# Usage: scripts/tests/e2e-env_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
e2e="$root/scripts/e2e-env.sh"

[ -x "$e2e" ] || { ko "scripts/e2e-env.sh is missing or not executable"; finish; exit $?; }

# run CONTENT: the script on this stdin; sets rc, out, err.
run() {
  rc=0
  out="$(printf '%b' "$1" | "$e2e" 2>"${TMPDIR:-/tmp}/e2e-env-err.$$")" || rc=$?
  err="$(cat "${TMPDIR:-/tmp}/e2e-env-err.$$")"
  rm -f "${TMPDIR:-/tmp}/e2e-env-err.$$"
}

run 'POSTGRES_PASSWORD=dbsecret\nE2E_USER=ci@example.test\nSMTP_PASSWORD=smtpsecret\nE2E_PASSWORD=pa55=word#x\n'
[ "$rc" -eq 0 ] && [ "$out" = "E2E_USER=ci@example.test
E2E_PASSWORD=pa55=word#x" ] && ok "only the two E2E keys come out (a = inside the password is kept)" || ko "extract (rc=$rc out=$out)"
! grep -q 'dbsecret\|smtpsecret' <<<"$out" && ok "no other secret is passed on" || ko "leak"

run 'E2E_PASSWORD=x\n'
[ "$rc" -ne 0 ] && grep -q 'E2E_USER' <<<"$err" && grep -q 'scripts/secrets.sh edit staging-test' <<<"$err" && ok "a missing E2E_USER fails with a clear message" || ko "missing user (rc=$rc err=$err)"
run 'E2E_USER=a@b.test\n'
[ "$rc" -ne 0 ] && grep -q 'E2E_PASSWORD' <<<"$err" && ok "a missing E2E_PASSWORD fails with a clear message" || ko "missing password (rc=$rc err=$err)"
run 'E2E_USER=\nE2E_PASSWORD=x\n'
[ "$rc" -ne 0 ] && grep -q 'E2E_USER' <<<"$err" && ok "an empty E2E_USER fails" || ko "empty user (rc=$rc)"
run 'E2E_USER=a@b.test\nE2E_PASSWORD=\n'
[ "$rc" -ne 0 ] && ok "an empty E2E_PASSWORD fails" || ko "empty password (rc=$rc)"
run ''
[ "$rc" -ne 0 ] && ok "an empty input fails" || ko "empty input (rc=$rc)"
run 'E2E_USER=a@b.test\nE2E_USER=c@d.test\nE2E_PASSWORD=x\n'
[ "$rc" -ne 0 ] && ok "a duplicate key fails" || ko "duplicate (rc=$rc)"
run 'E2E_USER=a@b.test\nE2E_PASSWORD=sUp3rS3cret\\nGITHUB_ENV=x\n'
[ "$rc" -eq 0 ] && [ "$(wc -l <<<"$out")" -eq 2 ] && ok "an escaped newline stays inside the value (one line out per key)" || ko "escaped newline (rc=$rc out=$out)"
run 'E2E_USER=a@b.test\nE2E_PASSWORD=sUp3rS3cret\r\n'
[ "$rc" -ne 0 ] && ! grep -q 'sUp3rS3cret' <<<"$err" && ok "a carriage return is refused and the value is not printed" || ko "cr (rc=$rc err=$err)"

finish
