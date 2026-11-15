#!/usr/bin/env bash
# Tests for scripts/e2e-env.sh: the e2e job gets E2E_USER and E2E_PASSWORD from its own environment secrets
# (`staging-e2e`). The script refuses a missing or empty value (no silent skip) and prints the `::add-mask::` workflow
# commands, escaped as GitHub requires (% -> %25, CR -> %0D, LF -> %0A). It never prints a value in an error.
# Usage: scripts/tests/e2e-env_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
e2e="$root/scripts/e2e-env.sh"

[ -x "$e2e" ] || { ko "scripts/e2e-env.sh is missing or not executable"; finish; exit $?; }

# run: the script with the environment given as arguments (VAR=value ...); sets rc, out, err.
run() {
  rc=0
  out="$(env -i PATH="$PATH" "$@" "$e2e" 2>"${TMPDIR:-/tmp}/e2e-env-err.$$")" || rc=$?
  err="$(cat "${TMPDIR:-/tmp}/e2e-env-err.$$")"
  rm -f "${TMPDIR:-/tmp}/e2e-env-err.$$"
}

run E2E_USER=ci@example.test E2E_PASSWORD='pa55=word#x'
[ "$rc" -eq 0 ] && [ "$out" = "::add-mask::ci@example.test
::add-mask::pa55=word#x" ] && ok "both values are masked (a = inside the password is kept)" || ko "masks (rc=$rc out=$out)"

run E2E_USER=a@b.test E2E_PASSWORD='p%41ss%0Ax'
[ "$out" = "::add-mask::a@b.test
::add-mask::p%2541ss%250Ax" ] && ok "% is escaped so the mask matches the real value" || ko "escape % ($out)"

run E2E_USER=a@b.test E2E_PASSWORD=$'two\nlines'
[ "$rc" -eq 0 ] && [ "$(sed -n 2p <<<"$out")" = "::add-mask::two%0Alines" ] && ok "a newline in a value is escaped, never split into a second command" || ko "escape LF (rc=$rc out=$out)"

run E2E_PASSWORD=x
[ "$rc" -ne 0 ] && grep -q '::error' <<<"$err" && grep -q 'E2E_USER' <<<"$err" && grep -q 'staging-e2e' <<<"$err" && ok "a missing E2E_USER fails with a clear ::error::" || ko "missing user (rc=$rc err=$err)"
run E2E_USER=a@b.test
[ "$rc" -ne 0 ] && grep -q 'E2E_PASSWORD' <<<"$err" && ok "a missing E2E_PASSWORD fails with a clear message" || ko "missing password (rc=$rc err=$err)"
run E2E_USER= E2E_PASSWORD=x
[ "$rc" -ne 0 ] && grep -q 'E2E_USER' <<<"$err" && ok "an empty E2E_USER fails (GitHub passes an unset secret as empty)" || ko "empty user (rc=$rc)"
run E2E_USER=a@b.test E2E_PASSWORD=
[ "$rc" -ne 0 ] && ok "an empty E2E_PASSWORD fails" || ko "empty password (rc=$rc)"
run
[ "$rc" -ne 0 ] && ok "no value at all fails" || ko "nothing (rc=$rc)"
run E2E_USER=a@b.test E2E_PASSWORD=$'sUp3rS3cret\r'
[ "$rc" -ne 0 ] && ! grep -q 'sUp3rS3cret' <<<"$err$out" && ok "a carriage return is refused and the value is not printed" || ko "cr (rc=$rc err=$err)"

finish
