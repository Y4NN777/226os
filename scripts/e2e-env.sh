#!/usr/bin/env bash
# From the decrypted staging-test env (stdin) keep ONLY the E2E account: prints `E2E_USER=...` and `E2E_PASSWORD=...`.
# The e2e job of deploy.yml decrypts secrets/staging-test.enc.env with SOPS_AGE_KEY into /dev/shm, pipes it here and
# loads the two lines into the job environment (after masking them). Nothing else leaves the file.
# Fails closed with a clear message when a key is missing, empty or repeated. Errors name keys, never values.
#
# Usage: scripts/e2e-env.sh < decrypted.env
set -euo pipefail

die() { printf 'e2e-env: REFUSED: %s\n' "$*" >&2; exit 2; }
hint='add it with scripts/secrets.sh edit staging-test (E2E_USER and E2E_PASSWORD, the verified staging account; docs/runbooks/deploy-setup.md section 7)'

declare -A val=()
while IFS= read -r line || [ -n "$line" ]; do
  case "$line" in
    E2E_USER=* | E2E_PASSWORD=*)
      k="${line%%=*}"
      [ -z "${val[$k]+set}" ] || die "$k is set twice in the staging-test env file"
      val[$k]="${line#*=}"
      ;;
  esac
done

for k in E2E_USER E2E_PASSWORD; do
  [ -n "${val[$k]+set}" ] || die "$k is missing from secrets/staging-test.enc.env: $hint"
  [ -n "${val[$k]}" ] || die "$k is empty in secrets/staging-test.enc.env: $hint"
  case "${val[$k]}" in *$'\r'*) die "$k contains a carriage return" ;; esac
done
printf 'E2E_USER=%s\nE2E_PASSWORD=%s\n' "${val[E2E_USER]}" "${val[E2E_PASSWORD]}"
