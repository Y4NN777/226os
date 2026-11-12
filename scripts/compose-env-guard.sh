#!/usr/bin/env bash
# Compose environment guard (ADR-0006): the APP_ENV in an env file must equal the
# Compose profile, otherwise exit non-zero before anything is started.
# A profile selects services; it is not a safety boundary, so every wrapper
# (scripts/compose.sh, Make up-/down- targets) calls this first.
#
# Usage: scripts/compose-env-guard.sh <profile> <env-file>
# Exit 0 when valid, 2 when refused. Read-only; the env file is parsed, never sourced.
set -euo pipefail

refuse() { printf 'REFUSED: %s\n' "$*" >&2; exit 2; }

profile="${1-}"
envfile="${2-}"

case "$profile" in
  local-dev | staging-test | production-live) ;;
  *) refuse "profile must be local-dev, staging-test or production-live, got '${profile}'" ;;
esac
[ -n "$envfile" ] && [ -f "$envfile" ] || refuse "env file not found: '${envfile}'"

# An ambient APP_ENV that disagrees would be silently preferred by some tools: refuse.
if [ -n "${APP_ENV+set}" ] && [ "$APP_ENV" != "$profile" ]; then
  refuse "ambient APP_ENV='${APP_ENV}' conflicts with profile '${profile}'"
fi

count=0
value=""
public_host=""
acme_email=""
while IFS= read -r line || [ -n "$line" ]; do
  case "$line" in
    '' | '#'*) continue ;;
  esac
  if [ "$profile" != local-dev ] && [[ "$line" == *CHANGE_ME* ]]; then
    refuse "placeholder CHANGE_ME left in ${envfile} for profile '${profile}'"
  fi
  if [[ "$line" =~ ^APP_ENV=(.*)$ ]]; then
    count=$((count + 1))
    value="${BASH_REMATCH[1]}"
    if [[ "$value" =~ ^\"(.*)\"$ || "$value" =~ ^\'(.*)\'$ ]]; then value="${BASH_REMATCH[1]}"; fi
  fi
  if [[ "$line" =~ ^PUBLIC_HOST=(.*)$ ]]; then public_host="${BASH_REMATCH[1]}"; fi
  if [[ "$line" =~ ^ACME_EMAIL=(.*)$ ]]; then acme_email="${BASH_REMATCH[1]}"; fi
done <"$envfile"

[ "$count" -eq 1 ] || refuse "APP_ENV must be defined exactly once in ${envfile} (found ${count})"
[ "$value" = "$profile" ] || refuse "APP_ENV='${value}' in ${envfile} does not match profile '${profile}'"
# The public edge needs a real DNS name and an ACME contact (Compose cannot require them per profile).
if [ "$profile" != local-dev ]; then
  host_re='^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]([a-z0-9-]{0,61}[a-z0-9])?$'
  [[ "$public_host" =~ $host_re && "$public_host" != localhost ]] ||
    refuse "PUBLIC_HOST must be a DNS name (no scheme, port or localhost) for profile '${profile}'"
  [[ "$acme_email" =~ ^[^@[:space:]]+@[^@[:space:]]+\.[^@[:space:]]+$ ]] ||
    refuse "ACME_EMAIL must be set to a contact address for profile '${profile}'"
fi
printf 'compose-env-guard: OK (profile=%s)\n' "$profile" >&2
