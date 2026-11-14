#!/usr/bin/env bash
# Single wrapper around `docker compose` for the profile-based tree (ADR-0006).
# Runs scripts/compose-env-guard.sh first, so a profile never starts with a mismatched APP_ENV.
#
# Usage: scripts/compose.sh <profile> <docker compose args...>
#   e.g. scripts/compose.sh local-dev up -d --build
# Env source, first match wins (ADR-0009):
#   1. COMPOSE_ENV_FILE                    an explicit env file
#   2. secrets/<profile>.enc.env           the SOPS-encrypted secrets in Git: decrypted at RUNTIME into a private
#                                          RAM-backed temp file for this command only, removed on exit, so no
#                                          plaintext .env has to sit on disk (needs sops and a private age key)
#   3. .env.<profile>                      a plain local env file (git-ignored)
# Overrides for tests: SECRETS_DIR.
# `down` / `rm` with -v or --volumes is refused (volumes hold the database).
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
profile="${1-}"
[ $# -ge 2 ] || { printf 'usage: %s <profile> <docker compose args...>\n' "$0" >&2; exit 64; }
shift

secrets_dir="${SECRETS_DIR:-$root/secrets}"
envfile="${COMPOSE_ENV_FILE:-}"
if [ -z "$envfile" ]; then
  if [ -f "$secrets_dir/$profile.enc.env" ]; then
    base="${TMPDIR:-/tmp}"
    [ -d /dev/shm ] && [ -w /dev/shm ] && base=/dev/shm
    tmpd="$(mktemp -d "$base/compose-env.XXXXXX")"
    trap 'rm -rf "$tmpd"' EXIT
    SECRETS_DIR="$secrets_dir" SECRETS_OUT_DIR="$tmpd" "$root/scripts/secrets.sh" decrypt "$profile" >/dev/null
    envfile="$tmpd/.env.$profile"
    [ ! -f "$root/.env.$profile" ] || printf 'compose.sh: using secrets/%s.enc.env; the plain .env.%s is ignored\n' "$profile" "$profile" >&2
  else
    envfile="$root/.env.$profile"
  fi
fi
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
# GITHUB_TOKEN too: the profile's env file is its only source (the ambient name is used by gh and other
# tools, and Compose lets the shell override --env-file), so a personal token never reaches the api container.
unset COMPOSE_PROFILES APP_ENV GITHUB_TOKEN
# No `exec`: the trap above must run to remove the decrypted temp file.
rc=0
docker compose --project-directory "$root" --env-file "$envfile" --profile "$profile" "$@" || rc=$?
exit "$rc"
