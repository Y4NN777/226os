#!/usr/bin/env bash
# Tests for scripts/compose-env-guard.sh: APP_ENV must equal the Compose profile.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
script="$here/../compose-env-guard.sh"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# case NAME EXPECT(ok|refuse) PROFILE ENVFILE-CONTENT
case_() {
  local name="$1" expect="$2" profile="$3" content="$4" rc=0 out
  printf '%b' "$content" >"$tmp/envfile"
  out="$(env -i PATH="$PATH" bash "$script" "$profile" "$tmp/envfile" 2>&1)" || rc=$?
  if [ "$expect" = ok ]; then
    [ "$rc" -eq 0 ] && ok "$name" || ko "$name (rc=$rc out=$out)"
  else
    if [ "$rc" -ne 0 ] && grep -q REFUSED <<<"$out"; then ok "$name"; else ko "$name (rc=$rc out=$out)"; fi
  fi
}

case_ "local-dev matches" ok local-dev 'APP_ENV=local-dev\n'
case_ "staging-test matches" ok staging-test 'APP_ENV=staging-test\nPUBLIC_HOST=staging.example.test\nACME_EMAIL=ops@example.test\n'
case_ "production-live matches" ok production-live 'APP_ENV=production-live\nPUBLIC_HOST=example.test\nACME_EMAIL=ops@example.test\n'
case_ "quoted value matches" ok local-dev 'APP_ENV="local-dev"\n'
case_ "comments and blank lines ignored" ok local-dev '# APP_ENV=production-live\n\nAPP_ENV=local-dev\n'
case_ "staging without PUBLIC_HOST" refuse staging-test 'APP_ENV=staging-test\nACME_EMAIL=ops@example.test\n'
case_ "staging without ACME_EMAIL" refuse staging-test 'APP_ENV=staging-test\nPUBLIC_HOST=staging.example.test\n'
case_ "PUBLIC_HOST not a host name" refuse production-live 'APP_ENV=production-live\nPUBLIC_HOST=a.test`x\nACME_EMAIL=ops@example.test\n'
case_ "PUBLIC_HOST with scheme" refuse production-live 'APP_ENV=production-live\nPUBLIC_HOST=https://example.test\nACME_EMAIL=ops@example.test\n'
case_ "PUBLIC_HOST localhost in production" refuse production-live 'APP_ENV=production-live\nPUBLIC_HOST=localhost\nACME_EMAIL=ops@example.test\n'
case_ "mismatch staging file / production profile" refuse production-live 'APP_ENV=staging-test\n'
case_ "mismatch production file / local profile" refuse local-dev 'APP_ENV=production-live\n'
case_ "APP_ENV missing" refuse local-dev 'FOO=bar\n'
case_ "APP_ENV empty" refuse local-dev 'APP_ENV=\n'
case_ "APP_ENV duplicated" refuse local-dev 'APP_ENV=local-dev\nAPP_ENV=local-dev\n'
case_ "APP_ENV with trailing garbage" refuse local-dev 'APP_ENV=local-dev x\n'
case_ "unknown profile" refuse prod 'APP_ENV=prod\n'
case_ "empty profile" refuse '' 'APP_ENV=local-dev\n'
case_ "placeholder secret in staging" refuse staging-test 'APP_ENV=staging-test\nPUBLIC_HOST=s.example.test\nACME_EMAIL=a@example.test\nPOSTGRES_PASSWORD=CHANGE_ME\n'
case_ "placeholder secret in production" refuse production-live 'APP_ENV=production-live\nPUBLIC_HOST=s.example.test\nACME_EMAIL=a@example.test\nDATABASE_URL=postgres://a:CHANGE_ME@postgres/x\n'
case_ "placeholder allowed in local-dev template-derived file" ok local-dev 'APP_ENV=local-dev\nPOSTGRES_PASSWORD=CHANGE_ME\n'

rc=0
out="$(env -i PATH="$PATH" bash "$script" local-dev "$tmp/does-not-exist" 2>&1)" || rc=$?
if [ "$rc" -ne 0 ] && grep -q REFUSED <<<"$out"; then ok "missing env file"; else ko "missing env file (rc=$rc)"; fi

# Ambient APP_ENV in the caller's shell must not override the file.
printf 'APP_ENV=local-dev\n' >"$tmp/envfile"
rc=0
out="$(env -i PATH="$PATH" APP_ENV=production-live bash "$script" local-dev "$tmp/envfile" 2>&1)" || rc=$?
if [ "$rc" -ne 0 ] && grep -q REFUSED <<<"$out"; then ok "ambient APP_ENV conflicting with profile refused"; else ko "ambient APP_ENV (rc=$rc)"; fi

finish
