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
printf '# template\nAPP_ENV=local-dev\nPOSTGRES_PASSWORD=CHANGE_ME\nSMTP_HOST=\nSMTP_PASSWORD=CHANGE_ME\n\nCURSOR_KEY=CHANGE_ME\n' >"$tmp/template.env"

run "no command prints the usage and fails" fail "usage" "${common[@]}" --
run "an unknown command fails" fail "usage" "${common[@]}" -- frobnicate
run "an unknown profile is refused" fail "profile" "${common[@]}" -- init staging
run "a path-like profile is refused" fail "profile" "${common[@]}" -- init ../../etc/passwd
run "init without a recipients file explains how to get one" fail "recipients" "${common[@]}" -- init staging-test
printf '# empty\n' >"$sd/staging-test.recipients"
run "init with an empty recipients file is refused" fail "recipient" "${common[@]}" -- init staging-test

priv="AGE-SECRET-$(printf 'KEY')-1$(printf 'Q%.0s' $(seq 1 58))"
printf '# staging\n%s\n' "$priv" >"$sd/staging-test.recipients"
rc=0; out="$(env "${common[@]}" "$script" init staging-test 2>&1)" || rc=$?
if [ "$rc" -ne 0 ] && ! grep -qF -- "$priv" <<<"$out"; then ok "a private key pasted into a recipients file is refused without being echoed"; else ko "private key echoed or accepted (rc=$rc): $out"; fi
printf '# empty\n' >"$sd/staging-test.recipients"

mkdir -p "$tmp/none"
printf 'existing\n' >"$tmp/keys.txt"
run "keygen refuses to overwrite an existing key file" fail "exists" SECRETS_DIR="$sd" SOPS_AGE_KEY_FILE="$tmp/keys.txt" -- keygen

if command -v age-keygen >/dev/null 2>&1; then
  ln -s "$tmp/does-not-exist/keys.txt" "$tmp/dangling.txt"
  rc=0; out="$(env SECRETS_DIR="$sd" SOPS_AGE_KEY_FILE="$tmp/dangling.txt" "$script" keygen 2>&1)" || rc=$?
  if [ "$rc" -ne 0 ] && [ -n "$out" ]; then ok "keygen on a dangling symlink fails with a message, not silently"; else ko "keygen dangling symlink: rc=$rc out='$out'"; fi
  printf 'not an age key\n' >"$tmp/garbage.txt"
  rc=0; out="$(env SECRETS_DIR="$sd" SOPS_AGE_KEY_FILE="$tmp/garbage.txt" "$script" keygen 2>&1)" || rc=$?
  if [ "$rc" -ne 0 ] && grep -q 'not a valid age key' <<<"$out"; then ok "keygen on an existing garbage key file says so"; else ko "keygen garbage key: rc=$rc out='$out'"; fi
fi

# CI has neither sops nor age: the checks that need no tool (profile, recipients file, existing files) must
# answer BEFORE the tool check. Emulated here with a PATH of symlinks that leaves sops and age out.
nb="$tmp/notools-bin"
mkdir -p "$nb"
for c in bash env sed grep mktemp rm mkdir head dirname basename cat sort tr mv chmod find wc stat ln cp tail awk readlink seq; do
  src="$(command -v "$c" 2>/dev/null || true)"
  [ -n "$src" ] && ln -sf "$src" "$nb/$c"
done
if PATH="$nb" command -v sops >/dev/null 2>&1 || PATH="$nb" command -v age-keygen >/dev/null 2>&1; then
  ko "the no-tools PATH still finds sops or age-keygen"
fi
nt() { # NAME EXPECT NEEDLE ARGS... : run secrets.sh with sops and age hidden
  local name="$1" want="$2" needle="$3"
  shift 3
  local rc=0 out
  out="$(env -i PATH="$nb" HOME="$tmp" SECRETS_DIR="$sd" SECRETS_OUT_DIR="$tmp/out" SECRETS_ENV_TEMPLATE="$tmp/template.env" \
    SOPS_AGE_KEY_FILE="$tmp/keys.txt" "$script" "$@" 2>&1)" || rc=$?
  if [ "$want" = fail ] && [ "$rc" -eq 0 ]; then ko "$name: should fail"; return; fi
  if ! grep -qF -- "$needle" <<<"$out"; then ko "$name: output lacks '$needle' (rc=$rc, out=$out)"; return; fi
  ok "$name (no sops, no age)"
}
rm -f "$sd/nofile.recipients"
nt "init without a recipients file says so, not 'sops is not installed'" fail "recipients" init production-live
nt "init with an empty recipients file says so" fail "recipient" init staging-test
nt "keygen on an existing key file refuses before needing age-keygen" fail "exists" keygen
nt "decrypt of a missing file says it does not exist" fail "does not exist" decrypt local-dev
nt "edit of a missing file says to init first" fail "init" edit local-dev
nt "rotate of a missing file says it does not exist" fail "does not exist" rotate local-dev
nt "a bad profile is refused" fail "profile" decrypt nope
printf '%s\n' "age1$(printf 'q%.0s' $(seq 1 58))" >"$sd/local-dev.recipients"
nt "with everything valid but no tool, the error names sops" fail "sops is not installed" init local-dev
rm -f "$sd/local-dev.recipients"

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

  run "init encrypts the template for the recipients (a KMS or PGP key group from the environment is ignored)" pass "" "${keyenv[@]}" SOPS_KMS_ARN=arn:aws:kms:eu-west-1:111122223333:key/not-real SOPS_PGP_FP=0000000000000000000000000000000000000000 -- init staging-test
  if grep -q '^sops_kms\|^sops_pgp' "$sd/staging-test.enc.env"; then ko "a KMS or PGP key group leaked into the encrypted file"; else ok "only age recipients in the encrypted file"; fi
  if [ "$(find "$sd" -mindepth 1 ! -name 'staging-test.*' | wc -l)" -eq 0 ]; then ok "no staging or temporary file is left in secrets/ after init"; else ko "leftovers in secrets/: $(find "$sd" -mindepth 1 | tr '\n' ' ')"; fi
  [ -f "$sd/staging-test.enc.env" ] && ok "the encrypted file exists" || ko "no secrets/staging-test.enc.env"
  if grep -q 'CHANGE_ME' "$sd/staging-test.enc.env"; then ko "the encrypted file still contains a plaintext value"; else ok "no plaintext value in the encrypted file"; fi
  if "$root/scripts/check-secrets-encrypted.sh" "$sd" >/dev/null 2>&1; then ok "the real sops output passes check-secrets-encrypted.sh"; else ko "check-secrets-encrypted.sh rejects real sops output"; fi
  run "init refuses to overwrite an existing secrets file" fail "exists" "${keyenv[@]}" -- init staging-test

  run "decrypt writes the profile env file" pass "" "${keyenv[@]}" -- decrypt staging-test
  if [ "$(stat -c %a "$tmp/out/.env.staging-test")" = 600 ]; then ok "the decrypted file is mode 600"; else ko "decrypted file mode is $(stat -c %a "$tmp/out/.env.staging-test"), want 600"; fi
  if grep -qx "APP_ENV=staging-test" "$tmp/out/.env.staging-test" && grep -qE "^POSTGRES_PASSWORD=[0-9a-f]{32}$" "$tmp/out/.env.staging-test" && grep -qx "SMTP_PASSWORD=CHANGE_ME" "$tmp/out/.env.staging-test"; then ok "staging: APP_ENV follows the profile, the generated secrets are filled, the external ones (SMTP_PASSWORD) stay placeholders to edit"; else ko "staging generated/external values wrong"; fi
  run "decrypt refuses to overwrite an existing env file" fail "exists" "${keyenv[@]}" -- decrypt staging-test
  run "decrypt --force overwrites it" pass "" "${keyenv[@]}" -- decrypt staging-test --force

  run "decrypt with the wrong key fails and shows what sops said" fail "sops said" SECRETS_DIR="$sd" SECRETS_OUT_DIR="$tmp/out2" SOPS_AGE_KEY_FILE="$k2" -- decrypt staging-test
  if [ -z "$(find "$tmp/out2" -mindepth 1 2>/dev/null)" ]; then ok "a failed decrypt leaves no file behind (not even a partial one)"; else ko "leftovers after a failed decrypt: $(find "$tmp/out2" -mindepth 1 | tr '\n' ' ')"; fi

  printf '# staging recipients\n%s\n%s\n' "$pub1" "$pub2" >"$sd/staging-test.recipients"
  run "rotate re-encrypts for the new recipient list" pass "" "${keyenv[@]}" -- rotate staging-test
  if [ "$(find "$sd" -mindepth 1 ! -name 'staging-test.*' | wc -l)" -eq 0 ]; then ok "no temporary file is left in secrets/ after rotate"; else ko "leftovers after rotate: $(find "$sd" -mindepth 1 | tr '\n' ' ')"; fi
  mkdir -p "$tmp/out3"
  run "the added recipient can now decrypt" pass "" SECRETS_DIR="$sd" SECRETS_OUT_DIR="$tmp/out3" SOPS_AGE_KEY_FILE="$k2" -- decrypt staging-test
  if "$root/scripts/check-secrets-encrypted.sh" "$sd" >/dev/null 2>&1; then ok "the rotated file still passes the encryption check"; else ko "rotated file fails the encryption check"; fi

  # local-dev only: init generates the three secrets, so nobody has to invent them (staging and production stay manual).
  printf '%s\n' "$pub1" >"$sd/local-dev.recipients"
  printf 'APP_ENV=local-dev\nPOSTGRES_USER=u\nPOSTGRES_PASSWORD=CHANGE_ME\nDATABASE_URL=postgres://u:CHANGE_ME@postgres:5432/d?sslmode=disable\nCURSOR_KEY=CHANGE_ME\nSMTP_HOST=\n' >"$tmp/template-ld.env"
  ldenv=(SECRETS_DIR="$sd" SECRETS_OUT_DIR="$tmp/out-ld" SECRETS_ENV_TEMPLATE="$tmp/template-ld.env" SOPS_AGE_KEY_FILE="$k1")
  mkdir -p "$tmp/out-ld"
  run "init local-dev generates its secrets" pass "" "${ldenv[@]}" -- init local-dev
  run "decrypt of the generated local-dev file" pass "" "${ldenv[@]}" -- decrypt local-dev
  ld="$tmp/out-ld/.env.local-dev"
  if grep -q CHANGE_ME "$ld"; then ko "a CHANGE_ME placeholder is left in the generated local-dev file"; else ok "no placeholder left in local-dev"; fi
  pw="$(sed -n 's/^POSTGRES_PASSWORD=//p' "$ld")"
  if [[ "$pw" =~ ^[0-9a-f]{32}$ ]]; then ok "POSTGRES_PASSWORD is 32 random hex characters"; else ko "POSTGRES_PASSWORD has an unexpected shape"; fi
  if grep -qxF "DATABASE_URL=postgres://u:${pw}@postgres:5432/d?sslmode=disable" "$ld"; then ok "DATABASE_URL carries the same password"; else ko "DATABASE_URL does not match POSTGRES_PASSWORD"; fi
  ck="$(sed -n 's/^CURSOR_KEY=//p' "$ld")"
  if [ "$(printf '%s' "$ck" | base64 -d 2>/dev/null | wc -c)" -eq 32 ]; then ok "CURSOR_KEY is 32 random bytes, base64"; else ko "CURSOR_KEY is not 32 base64 bytes"; fi
  grep -qx 'APP_ENV=local-dev' "$ld" && grep -qx 'SMTP_HOST=' "$ld" && ok "the non-secret values are untouched" || ko "non-secret values changed"
  "$root/scripts/check-secrets-encrypted.sh" "$sd" >/dev/null 2>&1 && ok "the generated file passes the encryption check" || ko "generated file fails the encryption check"
  rm "$sd/local-dev.enc.env" "$ld"
  run "a second init generates different values" pass "" "${ldenv[@]}" -- init local-dev
  run "decrypt of the second file" pass "" "${ldenv[@]}" -- decrypt local-dev
  [ "$(sed -n 's/^POSTGRES_PASSWORD=//p' "$ld")" != "$pw" ] && ok "values differ between two inits" || ko "the same password was generated twice"
  rm "$sd/local-dev.enc.env" "$sd/local-dev.recipients"
else
  notrun=1
  printf 'NOT RUN - the encrypt, decrypt and rotate round trip needs sops and age-keygen on PATH (not found)\n'
  [ -z "${GITHUB_ACTIONS:-}" ] || printf '::warning title=NOT RUN::secrets round trip needs sops and age; only the check script was tested here\n'
fi

finish
[ "$notrun" -eq 0 ] || printf 'note: the round trip above was NOT RUN on this machine\n'
