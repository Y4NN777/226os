#!/usr/bin/env bash
# RED/GREEN evidence for scripts/check-secrets-encrypted.sh (ADR-0009): everything under secrets/ must be
# SOPS-encrypted (every value ENC[...], SOPS metadata present, no private age key, no stray files).
# Pure bash: needs neither sops nor age, so it runs anywhere, including CI.
# Usage: scripts/tests/check-secrets-encrypted_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
check="$root/scripts/check-secrets-encrypted.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

[ -x "$check" ] || { ko "scripts/check-secrets-encrypted.sh is missing or not executable"; finish; exit 1; }

# Fixtures are built at run time: a literal private-key-shaped string would trip the secret scanner.
priv="AGE-SECRET-$(printf 'KEY')-1$(printf 'Q%.0s' $(seq 1 58))"
recip="age1$(printf 'q%.0s' $(seq 1 58))"
enc='ENC[AES256_GCM,data:AAAA,iv:BBBB,tag:CCCC,type:str]'

valid_file() { # DIR NAME
  mkdir -p "$1"
  cat >"$1/$2" <<EOF
POSTGRES_PASSWORD=$enc
CURSOR_KEY=$enc
sops_age__list_0__map_enc=-----BEGIN AGE ENCRYPTED FILE-----\nYWdlLWVuY3J5cHRpb24=\n-----END AGE ENCRYPTED FILE-----\n
sops_age__list_0__map_recipient=$recip
sops_lastmodified=2026-11-12T17:00:00Z
sops_mac=$enc
sops_unencrypted_suffix=_unencrypted
sops_version=3.13.3
EOF
}

expect() { # NAME EXPECT(pass|fail) NEEDLE DIR
  local name="$1" want="$2" needle="$3" dir="$4" rc=0 out
  out="$("$check" "$dir" 2>&1)" || rc=$?
  if [ "$want" = pass ] && [ "$rc" -ne 0 ]; then ko "$name: should pass (rc=$rc, out=$out)"; return; fi
  if [ "$want" = fail ] && [ "$rc" -eq 0 ]; then ko "$name: should fail but passed (out=$out)"; return; fi
  if [ -n "$needle" ] && ! grep -qF -- "$needle" <<<"$out"; then ko "$name: output lacks '$needle' (out=$out)"; return; fi
  ok "$name"
}

expect "no secrets directory yet: nothing to check" pass "" "$tmp/absent"

d="$tmp/valid"
valid_file "$d" staging-test.enc.env
printf '# the staging recipients\n%s\n' "$recip" >"$d/staging-test.recipients"
printf '# Secrets\nEncrypted with SOPS and age (ADR-0009).\n' >"$d/README.md"
expect "a fully encrypted file, its recipients file and a README pass" pass "" "$d"

d="$tmp/emptyvalue"
valid_file "$d" staging-test.enc.env
printf 'SMTP_HOST=\n' >>"$d/staging-test.enc.env"
expect "an empty value is allowed (sops leaves it empty: there is nothing to hide)" pass "" "$d"

d="$tmp/plain"
valid_file "$d" staging-test.enc.env
printf 'SMTP_PASSWORD=plaintext-value\n' >>"$d/staging-test.enc.env"
expect "a plaintext value is refused and named" fail "SMTP_PASSWORD" "$d"

d="$tmp/nomac"
valid_file "$d" staging-test.enc.env
grep -v '^sops_mac=' "$d/staging-test.enc.env" >"$d/x" && mv "$d/x" "$d/staging-test.enc.env"
expect "a file without sops_mac is refused" fail "sops_mac" "$d"

d="$tmp/norecipient"
valid_file "$d" staging-test.enc.env
grep -v '^sops_age__list_0__map_recipient=' "$d/staging-test.enc.env" >"$d/x" && mv "$d/x" "$d/staging-test.enc.env"
expect "a file without an age recipient is refused" fail "recipient" "$d"

d="$tmp/privkey"
valid_file "$d" staging-test.enc.env
printf '%s\n' "$priv" >"$d/staging-test.recipients"
expect "a private age key in a recipients file is refused" fail "private" "$d"

d="$tmp/privkey2"
valid_file "$d" staging-test.enc.env
printf 'NOTE=%s\n' "$priv" >>"$d/staging-test.enc.env"
expect "a private age key anywhere in an encrypted file is refused" fail "private" "$d"

d="$tmp/comment"
valid_file "$d" staging-test.enc.env
printf '# the password is hunter2\n' >>"$d/staging-test.enc.env"
expect "a plaintext comment in an encrypted file is refused" fail "comment" "$d"

d="$tmp/stray"
valid_file "$d" staging-test.enc.env
printf 'scratch\n' >"$d/notes.txt"
expect "a stray file is refused" fail "notes.txt" "$d"

d="$tmp/plainenvname"
valid_file "$d" staging-test.enc.env
printf 'A=1\n' >"$d/staging-test.env"
expect "a plain .env file under secrets/ is refused" fail "staging-test.env" "$d"

d="$tmp/badenc"
valid_file "$d" staging-test.enc.env
printf 'TOKEN=ENC[not-a-real-envelope]\n' >>"$d/staging-test.enc.env"
expect "a value that only looks like an ENC envelope is refused" fail "TOKEN" "$d"

d="$tmp/badrecipient"
valid_file "$d" staging-test.enc.env
printf 'not-an-age-key\n' >"$d/staging-test.recipients"
expect "a malformed recipients line is refused" fail "recipients" "$d"

finish
