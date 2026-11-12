#!/usr/bin/env bash
# RED/GREEN evidence for scripts/secrets.sh (ADR-0009): the one way to create, edit, decrypt and rotate the
# SOPS+age secrets files. Argument handling and refusals are tested everywhere. The real round trip
# (init, decrypt, rotate with another recipient) needs sops and age; without them it reports
# "NOT RUN" explicitly, it is never counted as passed.
# Usage: scripts/tests/secrets_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
script="$root/scripts/secrets.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
notrun=0

[ -x "$script" ] || { ko "scripts/secrets.sh is missing or not executable"; finish; exit 1; }

# run NAME EXPECT(pass|fail) NEEDLE ENV... -- ARGS...
run() {
  local name="$1" want="$2" needle="$3"
  shift 3
  local envs=()
  while [ "$1" != "--" ]; do envs+=("$1"); shift; done
  shift
  local rc=0 out
  out="$(env "${envs[@]}" "$script" "$@" 2>&1)" || rc=$?
  if [ "$want" = pass ] && [ "$rc" -ne 0 ]; then ko "$name: should pass (rc=$rc, out=$out)"; return; fi
  if [ "$want" = fail ] && [ "$rc" -eq 0 ]; then ko "$name: should fail but passed (out=$out)"; return; fi
  if [ -n "$needle" ] && ! grep -qF -- "$needle" <<<"$out"; then ko "$name: output lacks '$needle' (out=$out)"; return; fi
  ok "$name"
}

sd="$tmp/secrets"
mkdir -p "$sd"
common=(SECRETS_DIR="$sd" SECRETS_OUT_DIR="$tmp/out" SECRETS_ENV_TEMPLATE="$tmp/template.env" SOPS_AGE_KEY_FILE="$tmp/none/keys.txt")
mkdir -p "$tmp/out"
printf '# template\nPOSTGRES_PASSWORD=CHANGE_ME\nSMTP_HOST=\n\nCURSOR_KEY=CHANGE_ME\n' >"$tmp/template.env"

run "no command prints the usage and fails" fail "usage" "${common[@]}" --
run "an unknown command fails" fail "usage" "${common[@]}" -- frobnicate
run "an unknown profile is refused" fail "profile" "${common[@]}" -- init staging
run "a path-like profile is refused" fail "profile" "${common[@]}" -- init ../../etc/passwd
run "init without a recipients file explains how to get one" fail "recipients" "${common[@]}" -- init staging-test
printf '# empty\n' >"$sd/staging-test.recipients"
run "init with an empty recipients file is refused" fail "recipient" "${common[@]}" -- init staging-test

mkdir -p "$tmp/none"
printf 'existing\n' >"$tmp/keys.txt"
run "keygen refuses to overwrite an existing key file" fail "already exists" SECRETS_DIR="$sd" SOPS_AGE_KEY_FILE="$tmp/keys.txt" -- keygen

# The tools themselves: sops and age are optional on a machine that does not touch secrets.
if command -v sops >/dev/null 2>&1 && command -v age-keygen >/dev/null 2>&1; then
  k1="$tmp/k1.txt"
  k2="$tmp/k2.txt"
  age-keygen -o "$k1" >/dev/null 2>&1
  age-keygen -o "$k2" >/dev/null 2>&1
  pub1="$(age-keygen -y "$k1")"
  pub2="$(age-keygen -y "$k2")"
  printf '# staging recipients\n%s\n' "$pub1" >"$sd/staging-test.recipients"
  keyenv=(SECRETS_DIR="$sd" SECRETS_OUT_DIR="$tmp/out" SECRETS_ENV_TEMPLATE="$tmp/template.env" SOPS_AGE_KEY_FILE="$k1")

  run "init encrypts the template for the recipients" pass "" "${keyenv[@]}" -- init staging-test
  [ -f "$sd/staging-test.enc.env" ] && ok "the encrypted file exists" || ko "no secrets/staging-test.enc.env"
  if grep -q 'CHANGE_ME' "$sd/staging-test.enc.env"; then ko "the encrypted file still contains a plaintext value"; else ok "no plaintext value in the encrypted file"; fi
  if "$root/scripts/check-secrets-encrypted.sh" "$sd" >/dev/null 2>&1; then ok "the real sops output passes check-secrets-encrypted.sh"; else ko "check-secrets-encrypted.sh rejects real sops output"; fi
  run "init refuses to overwrite an existing secrets file" fail "exists" "${keyenv[@]}" -- init staging-test

  run "decrypt writes the profile env file" pass "" "${keyenv[@]}" -- decrypt staging-test
  if [ "$(stat -c %a "$tmp/out/.env.staging-test")" = 600 ]; then ok "the decrypted file is mode 600"; else ko "decrypted file mode is $(stat -c %a "$tmp/out/.env.staging-test"), want 600"; fi
  if grep -qx 'POSTGRES_PASSWORD=CHANGE_ME' "$tmp/out/.env.staging-test"; then ok "decrypt restores the values"; else ko "decrypted content differs"; fi
  run "decrypt refuses to overwrite an existing env file" fail "exists" "${keyenv[@]}" -- decrypt staging-test
  run "decrypt --force overwrites it" pass "" "${keyenv[@]}" -- decrypt staging-test --force

  run "decrypt with the wrong key fails and writes nothing new" fail "" SECRETS_DIR="$sd" SECRETS_OUT_DIR="$tmp/out2" SOPS_AGE_KEY_FILE="$k2" -- decrypt staging-test

  printf '# staging recipients\n%s\n%s\n' "$pub1" "$pub2" >"$sd/staging-test.recipients"
  run "rotate re-encrypts for the new recipient list" pass "" "${keyenv[@]}" -- rotate staging-test
  mkdir -p "$tmp/out3"
  run "the added recipient can now decrypt" pass "" SECRETS_DIR="$sd" SECRETS_OUT_DIR="$tmp/out3" SOPS_AGE_KEY_FILE="$k2" -- decrypt staging-test
  if "$root/scripts/check-secrets-encrypted.sh" "$sd" >/dev/null 2>&1; then ok "the rotated file still passes the encryption check"; else ko "rotated file fails the encryption check"; fi
else
  notrun=1
  printf 'NOT RUN - the encrypt, decrypt and rotate round trip needs sops and age-keygen on PATH (not found)\n'
fi

finish
[ "$notrun" -eq 0 ] || printf 'note: the round trip above was NOT RUN on this machine\n'
