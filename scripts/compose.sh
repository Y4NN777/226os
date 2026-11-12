#!/usr/bin/env bash
# Single wrapper around `docker compose` for the profile-based tree (ADR-0006).
# Runs scripts/compose-env-guard.sh first, so a profile never starts with a mismatched APP_ENV.
#
# Usage: scripts/compose.sh <profile> <docker compose args...>
#   e.g. scripts/compose.sh local-dev up -d --build
# Env:   COMPOSE_ENV_FILE  env file to use (default: .env.<profile>)
# `down` / `rm` with -v or --volumes is refused (volumes hold the database).
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
profile="${1-}"
[ $# -ge 2 ] || { printf 'usage: %s <profile> <docker compose args...>\n' "$0" >&2; exit 64; }
shift

envfile="${COMPOSE_ENV_FILE:-$root/.env.$profile}"
"$root/scripts/compose-env-guard.sh" "$profile" "$envfile"

# Only for the commands that can remove volumes (`-v` means something else after `exec ... psql`).
case "$1" in
  down | rm)
    for a in "$@"; do
      case "$a" in
        -v | --volumes) printf 'REFUSED: compose.sh never removes volumes (%s %s)\n' "$1" "$a" >&2; exit 2 ;;
      esac
    done
    ;;
esac

# Ambient COMPOSE_PROFILES / APP_ENV must not influence selection: the wrapper is the source of truth.
unset COMPOSE_PROFILES APP_ENV
exec docker compose --project-directory "$root" --env-file "$envfile" --profile "$profile" "$@"
