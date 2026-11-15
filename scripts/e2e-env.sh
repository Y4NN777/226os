#!/usr/bin/env bash
# The e2e job of deploy.yml gets the E2E account from the secrets of its own GitHub environment (`staging-e2e`), as
# E2E_USER and E2E_PASSWORD in the environment of this step. This script checks them and prints the `::add-mask::`
# workflow commands for the two values, escaped as GitHub requires (% -> %25, CR -> %0D, LF -> %0A), so the mask matches
# the real value even when it holds a %. Run it BEFORE anything uses the values.
# Fails closed (exit 2) with an ::error:: line when a value is missing or empty (GitHub passes an unset secret as an
# empty string) or holds a carriage return. Errors name the keys, never the values.
#
# Usage: E2E_USER=... E2E_PASSWORD=... scripts/e2e-env.sh
set -euo pipefail

die() { printf '::error title=E2E account not configured::%s\n' "$*" >&2; exit 2; }
hint='the owner sets it with: gh secret set <NAME> --env staging-e2e (docs/runbooks/deploy-setup.md section 7)'

for k in E2E_USER E2E_PASSWORD; do
  v="${!k-}"
  [ -n "$v" ] || die "$k is missing or empty in the staging-e2e environment: $hint"
  case "$v" in *$'\r'*) die "$k contains a carriage return" ;; esac
done
for k in E2E_USER E2E_PASSWORD; do
  v="${!k}"
  v="${v//%/%25}"
  v="${v//$'\r'/%0D}"
  v="${v//$'\n'/%0A}"
  printf '::add-mask::%s\n' "$v"
done
