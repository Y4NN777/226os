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
iv="$(printf 'B%.0s' $(seq 1 43))="
tag="$(printf 'C%.0s' $(seq 1 22))=="
enc="ENC[AES256_GCM,data:AAAA,iv:$iv,tag:$tag,type:str]"

valid_file() { # DIR NAME
  mkdir -p "$1"
  printf '# staging recipients\n%s\n' "$recip" >"$1/${2%.enc.env}.recipients"
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

expect "a secrets directory that does not exist is a failure, never a vacuous pass" fail "not found" "$tmp/absent"

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

# --- findings of the independent review ---
d="$tmp/sopsjunk"
valid_file "$d" staging-test.enc.env
printf 'sops_foo=Xk9fQ2mZp7LwR4vT8nBy\n' >>"$d/staging-test.enc.env"
expect "an unknown sops_ key is refused (it could carry plaintext)" fail "sops_foo" "$d"

d="$tmp/sopsversion"
valid_file "$d" staging-test.enc.env
sed -i 's/^sops_version=.*/sops_version=3.13.3 password=hunter2/' "$d/staging-test.enc.env"
expect "a sops_version value that is not a version is refused" fail "sops_version" "$d"

d="$tmp/sopslast"
valid_file "$d" staging-test.enc.env
sed -i 's/^sops_lastmodified=.*/sops_lastmodified=hunter2/' "$d/staging-test.enc.env"
expect "a sops_lastmodified value that is not a timestamp is refused" fail "sops_lastmodified" "$d"

d="$tmp/keysmuggle"
valid_file "$d" staging-test.enc.env
printf 'DB_PASSWORD_IS_hunter2=\n' >>"$d/staging-test.enc.env"
expect "a key name carrying data is refused (valid names are UPPER_SNAKE_CASE)" fail "key name" "$d"

d="$tmp/keyspace"
valid_file "$d" staging-test.enc.env
printf ' B=\n' >>"$d/staging-test.enc.env"
expect "a key name with a leading space is refused" fail "key name" "$d"

d="$tmp/keyempty"
valid_file "$d" staging-test.enc.env
printf '=\n' >>"$d/staging-test.enc.env"
expect "an empty key name is refused" fail "key name" "$d"

d="$tmp/shortiv"
valid_file "$d" staging-test.enc.env
printf 'B=ENC[AES256_GCM,data:AAAA,iv:hunter2,tag:%s,type:str]\n' "$tag" >>"$d/staging-test.enc.env"
expect "an envelope with an iv of the wrong shape is refused" fail "B" "$d"

d="$tmp/emptydata"
valid_file "$d" staging-test.enc.env
printf 'B=ENC[AES256_GCM,data:,iv:%s,tag:%s,type:str]\n' "$iv" "$tag" >>"$d/staging-test.enc.env"
expect "an envelope with empty data is refused" fail "B" "$d"

d="$tmp/locale"
valid_file "$d" staging-test.enc.env
printf 'B=ENC[AES256_GCM,data:\xc3\xa9,iv:%s,tag:%s,type:str]\n' "$iv" "$tag" >>"$d/staging-test.enc.env"
rc=0; out="$(env LC_ALL=en_US.UTF-8 "$check" "$d" 2>&1)" || rc=$?
if [ "$rc" -ne 0 ]; then ok "non-ASCII data is refused whatever the caller's locale"; else ko "the verdict depends on the locale (a UTF-8 locale passed non-ASCII data): $out"; fi

d="$tmp/revoked"
valid_file "$d" staging-test.enc.env
printf 'sops_age__list_1__map_recipient=age1%s\n' "$(printf 'z%.0s' $(seq 1 58))" >>"$d/staging-test.enc.env"
expect "a recipient still in the encrypted file but removed from .recipients is refused (run rotate)" fail "rotate" "$d"

d="$tmp/addedrecipient"
valid_file "$d" staging-test.enc.env
printf '%s\n' "age1$(printf 'z%.0s' $(seq 1 58))" >>"$d/staging-test.recipients"
expect "a recipient in .recipients that the encrypted file does not have is refused (run rotate)" fail "rotate" "$d"

d="$tmp/norecipientsfile"
valid_file "$d" staging-test.enc.env
rm "$d/staging-test.recipients"
expect "an encrypted file without its .recipients file is refused" fail "staging-test.recipients" "$d"

d="$tmp/symlink"
valid_file "$d" staging-test.enc.env
ln -s /etc/hostname "$d/README.md"
expect "a symlink is refused" fail "symlink" "$d"

d="$tmp/readmekv"
valid_file "$d" staging-test.enc.env
printf 'DB_PASSWORD=hunter2verysecret\n' >"$d/README.md"
expect "a README.md that looks like KEY=VALUE is refused" fail "README.md" "$d"

d="$tmp/longcomment"
valid_file "$d" staging-test.enc.env
printf '# %s\n' "$(printf 'x%.0s' $(seq 1 120))" >>"$d/staging-test.recipients"
expect "a long free-text comment in .recipients is refused" fail "comment" "$d"

d="$tmp/huge"
valid_file "$d" staging-test.enc.env
head -c 300000 /dev/zero | tr '\0' 'A' | sed 's/^/BIG=/' >>"$d/staging-test.enc.env"
expect "an oversized file is refused (and does not stall the gate)" fail "too large" "$d"

finish
