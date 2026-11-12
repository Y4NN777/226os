#!/usr/bin/env bash
# Drop and recreate the application database. LOCAL DEVELOPMENT ONLY (ADR-0006).
#
# Runs only when ALL hold, otherwise exits 2 BEFORE touching anything:
#   - APP_ENV is exactly "local-dev"
#   - COMPOSE_PROFILES, if set, is exactly "local-dev"
#   - DATABASE_URL is a strict postgres:// URL (no multi-host, no userinfo tricks,
#     only an empty query or sslmode=disable) whose host is one of the local ones:
#     postgres (the Compose service), localhost, 127.0.0.1
#   - the port is absent or exactly 5432
#   - the database name is a plain lower-case identifier and not postgres/template*
# Unset, empty or ambiguous variables fail closed.
#
# Env:     DRY_RUN=1          print the SQL, execute nothing
#          DB_RESET_RUNNER    TEST HOOK: command that receives the SQL on stdin. It is honoured only
#                             together with DB_RESET_ALLOW_TEST_RUNNER=1; set without that flag the
#                             script REFUSES (an ambient variable must never replace the default).
#                             Default runner: psql inside the local Compose "postgres" service via
#                             scripts/compose.sh.
# Usage:   APP_ENV=local-dev DATABASE_URL=... scripts/db-reset.sh
set -euo pipefail

refuse() { printf 'REFUSED: db-reset %s\n' "$*" >&2; exit 2; }

[ -n "${APP_ENV-}" ] || refuse "needs APP_ENV=local-dev; APP_ENV is unset or empty"
[ "$APP_ENV" = "local-dev" ] || refuse "only runs with APP_ENV=local-dev; got a different environment ('${APP_ENV}')"

if [ -n "${COMPOSE_PROFILES+set}" ] && [ "$COMPOSE_PROFILES" != "local-dev" ]; then
  refuse "COMPOSE_PROFILES must be unset or exactly local-dev; got '${COMPOSE_PROFILES}'"
fi

if [ -n "${DB_RESET_RUNNER+set}" ] && [ "${DB_RESET_ALLOW_TEST_RUNNER-}" != 1 ]; then
  refuse "DB_RESET_RUNNER is set without DB_RESET_ALLOW_TEST_RUNNER=1; refusing to use a custom runner"
fi

url="${DATABASE_URL-}"
[ -n "$url" ] || refuse "DATABASE_URL is unset or empty"

# user[:password]@host[:port]/db[?sslmode=disable]; no @ / ? # inside any component.
re='^postgres(ql)?://([^:@/?#[:space:]]+)(:[^@/?#[:space:]]*)?@([A-Za-z0-9.-]+)(:[0-9]{1,5})?/([^/?#@[:space:]]+)(\?sslmode=disable)?$'
if [[ ! "$url" =~ $re ]]; then
  refuse "DATABASE_URL is not a strict single-host postgres URL (credentials hidden)"
fi
user="${BASH_REMATCH[2]}"
host="${BASH_REMATCH[4]}"
port="${BASH_REMATCH[5]#:}"
dbname="${BASH_REMATCH[6]}"

if [ -n "$port" ] && [ "$port" != 5432 ]; then
  refuse "port '${port}' is not allowed (only 5432 or none)"
fi

case "$host" in
  postgres | localhost | 127.0.0.1) ;;
  *) refuse "host '${host}' is not the local Compose database (allowed: postgres, localhost, 127.0.0.1)" ;;
esac

if [[ ! "$dbname" =~ ^[a-z][a-z0-9_]{0,62}$ ]] || [[ "$dbname" == postgres || "$dbname" == template* ]]; then
  refuse "database name '${dbname}' is not an allowed application database name"
fi
if [[ ! "$user" =~ ^[A-Za-z_][A-Za-z0-9_]{0,62}$ ]]; then
  refuse "database user is not a plain identifier"
fi

sql="DROP DATABASE IF EXISTS \"${dbname}\" WITH (FORCE);
CREATE DATABASE \"${dbname}\" OWNER \"${user}\";"

if [ "${DRY_RUN:-0}" = 1 ]; then
  printf 'DRY_RUN: would run against local database host=%s db=%s:\n%s\n' "$host" "$dbname" "$sql"
  exit 0
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
printf 'db-reset: dropping and recreating "%s" on local host %s\n' "$dbname" "$host"
if [ -n "${DB_RESET_RUNNER-}" ]; then # only reachable with DB_RESET_ALLOW_TEST_RUNNER=1
  # shellcheck disable=SC2086  # the runner is an intentionally word-split command line
  printf '%s\n' "$sql" | $DB_RESET_RUNNER
else
  printf '%s\n' "$sql" |
    "$root/scripts/compose.sh" local-dev exec -T postgres psql -v ON_ERROR_STOP=1 -U "$user" -d postgres
fi
echo "db-reset: done (run migrations next)"
