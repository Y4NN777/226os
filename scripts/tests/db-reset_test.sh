#!/usr/bin/env bash
# Tests for scripts/db-reset.sh (ADR-0006: runs only with APP_ENV=local-dev and a local Compose DB).
# Refusal cases use an injected runner that records any invocation: it must never be called.
# Usage: scripts/tests/db-reset_test.sh            (unit cases, no Docker)
#        DB_RESET_TEST_DOCKER=1 scripts/tests/db-reset_test.sh   (adds one real test on an ephemeral postgres)
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
script="$here/../db-reset.sh"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
marker="$tmp/runner-called"
runner="$tmp/runner.sh"
cat >"$runner" <<RUNNER
#!/usr/bin/env bash
cat >>"$tmp/runner-stdin"
echo called >"$marker"
RUNNER
chmod +x "$runner"

good_url='postgres://app:s3cret@postgres:5432/app226'

# run_case NAME EXPECT(refuse|run) [VAR=value ...]  (runs under env -i: nothing leaks from the caller)
run_case() {
  local name="$1" expect="$2"
  shift 2
  rm -f "$marker" "$tmp/runner-stdin"
  local rc=0 out
  out="$(env -i PATH="$PATH" DB_RESET_RUNNER="$runner" DB_RESET_ALLOW_TEST_RUNNER=1 "$@" bash "$script" 2>&1)" || rc=$?
  if [ "$expect" = refuse ]; then
    if [ "$rc" -ne 0 ] && [ ! -e "$marker" ] && grep -q 'REFUSED' <<<"$out"; then ok "$name"; else ko "$name (rc=$rc, runner_called=$([ -e "$marker" ] && echo yes || echo no), out=$out)"; fi
  else
    if [ "$rc" -eq 0 ] && [ -e "$marker" ]; then ok "$name"; else ko "$name (rc=$rc, out=$out)"; fi
  fi
}

run_case "APP_ENV unset" refuse DATABASE_URL="$good_url"
run_case "APP_ENV empty" refuse APP_ENV= DATABASE_URL="$good_url"
run_case "APP_ENV=staging-test" refuse APP_ENV=staging-test DATABASE_URL="$good_url"
run_case "APP_ENV=production-live" refuse APP_ENV=production-live DATABASE_URL="$good_url"
run_case "APP_ENV=local-dev with trailing space" refuse "APP_ENV=local-dev " DATABASE_URL="$good_url"
run_case "APP_ENV=LOCAL-DEV (case)" refuse APP_ENV=LOCAL-DEV DATABASE_URL="$good_url"
run_case "DATABASE_URL unset" refuse APP_ENV=local-dev
run_case "DATABASE_URL empty" refuse APP_ENV=local-dev DATABASE_URL=
run_case "remote host" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@db.prod.internal:5432/app226'
run_case "public IP host" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@203.0.113.7:5432/app226'
run_case "postgres.example.com" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres.example.com:5432/app226'
run_case "localhost.evil" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@localhost.evil:5432/app226'
run_case "127.0.0.1.evil" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@127.0.0.1.evil:5432/app226'
run_case "userinfo trick user=localhost host=evil" refuse APP_ENV=local-dev DATABASE_URL='postgres://localhost@evil.example:5432/app226'
run_case "userinfo trick double @" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@evil.example@localhost:5432/app226'
run_case "userinfo trick password contains @host" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:pw@postgres@evil.example/app226'
run_case "query host override" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres:5432/app226?host=evil.example'
run_case "query hostaddr override" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres/app226?hostaddr=203.0.113.7'
run_case "multi-host URL" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres,evil.example:5432/app226'
run_case "fragment trick" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@evil.example/app226#@postgres'
run_case "wrong scheme" refuse APP_ENV=local-dev DATABASE_URL='mysql://app:x@postgres:5432/app226'
run_case "no host" refuse APP_ENV=local-dev DATABASE_URL='postgres:///app226'
run_case "newline in URL" refuse APP_ENV=local-dev DATABASE_URL=$'postgres://app:x@postgres:5432/app226\n@evil.example'
run_case "db name is postgres" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres:5432/postgres'
run_case "db name with SQL injection" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres:5432/a";drop'
run_case "COMPOSE_PROFILES=production-live" refuse APP_ENV=local-dev COMPOSE_PROFILES=production-live DATABASE_URL="$good_url"
run_case "COMPOSE_PROFILES mixed" refuse APP_ENV=local-dev COMPOSE_PROFILES=local-dev,staging-test DATABASE_URL="$good_url"
run_case "valid: postgres service" run APP_ENV=local-dev DATABASE_URL="$good_url"
run_case "valid: 127.0.0.1" run APP_ENV=local-dev DATABASE_URL='postgres://app:x@127.0.0.1:5432/app226'
run_case "valid: localhost + sslmode=disable" run APP_ENV=local-dev DATABASE_URL='postgres://app:x@localhost/app226?sslmode=disable'
run_case "valid: COMPOSE_PROFILES=local-dev" run APP_ENV=local-dev COMPOSE_PROFILES=local-dev DATABASE_URL="$good_url"

run_case "port 99999" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres:99999/app226'
run_case "port 5433" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres:5433/app226'
run_case "port 0" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres:0/app226'
run_case "port empty after colon" refuse APP_ENV=local-dev DATABASE_URL='postgres://app:x@postgres:/app226'

# An ambient runner without the explicit test flag is refused and never executed.
rm -f "$marker"
out="$(env -i PATH="$PATH" DB_RESET_RUNNER="$runner" APP_ENV=local-dev DATABASE_URL="$good_url" bash "$script" 2>&1)" && rc=0 || rc=$?
if [ "$rc" -ne 0 ] && [ ! -e "$marker" ] && grep -q REFUSED <<<"$out"; then ok "ambient DB_RESET_RUNNER without flag refused, not executed"; else ko "ambient runner (rc=$rc out=$out)"; fi
# ... and also when the flag has a value other than 1.
rm -f "$marker"
out="$(env -i PATH="$PATH" DB_RESET_RUNNER="$runner" DB_RESET_ALLOW_TEST_RUNNER=yes APP_ENV=local-dev DATABASE_URL="$good_url" bash "$script" 2>&1)" && rc=0 || rc=$?
if [ "$rc" -ne 0 ] && [ ! -e "$marker" ]; then ok "runner flag must be exactly 1"; else ko "runner flag value (rc=$rc)"; fi
# With the flag it is executed (also covered by the valid cases above).
rm -f "$marker"
env -i PATH="$PATH" DB_RESET_RUNNER="$runner" DB_RESET_ALLOW_TEST_RUNNER=1 APP_ENV=local-dev DATABASE_URL="$good_url" bash "$script" >/dev/null 2>&1 || true
[ -e "$marker" ] && ok "runner executed with DB_RESET_ALLOW_TEST_RUNNER=1" || ko "runner not executed with flag"

# DRY_RUN prints the SQL and never calls the runner.
rm -f "$marker"
out="$(env -i PATH="$PATH" DB_RESET_RUNNER="$runner" DB_RESET_ALLOW_TEST_RUNNER=1 APP_ENV=local-dev DATABASE_URL="$good_url" DRY_RUN=1 bash "$script" 2>&1)" && rc=0 || rc=$?
if [ "$rc" -eq 0 ] && [ ! -e "$marker" ] && grep -q 'DROP DATABASE' <<<"$out"; then ok "DRY_RUN prints SQL, runner not called"; else ko "DRY_RUN (rc=$rc out=$out)"; fi

# The password must never be echoed.
out="$(env -i PATH="$PATH" DB_RESET_RUNNER="$runner" DB_RESET_ALLOW_TEST_RUNNER=1 APP_ENV=local-dev DATABASE_URL="$good_url" DRY_RUN=1 bash "$script" 2>&1)" || true
if grep -q 's3cret' <<<"$out"; then ko "password leaked in output"; else ok "password not echoed"; fi

# Refusal output must not leak the password either.
out="$(env -i PATH="$PATH" DB_RESET_RUNNER="$runner" APP_ENV=production-live DATABASE_URL="$good_url" bash "$script" 2>&1)" || true
if grep -q 's3cret' <<<"$out"; then ko "password leaked in refusal"; else ok "password not echoed on refusal"; fi

# Optional real positive test on an ephemeral container of the pinned image.
if [ "${DB_RESET_TEST_DOCKER:-0}" = 1 ]; then
  image="$(grep -Eo 'postgres:18@sha256:[0-9a-f]{64}' "$here/../../compose.yaml" | head -n1)"
  name="db-reset-test-$$"
  docker run -d --rm --name "$name" -e POSTGRES_PASSWORD=pw -e POSTGRES_USER=app "$image" >/dev/null
  trap 'docker rm -f "$name" >/dev/null 2>&1 || true; rm -rf "$tmp"' EXIT
  for _ in $(seq 1 60); do docker exec "$name" pg_isready -U app -h 127.0.0.1 >/dev/null 2>&1 && break; sleep 1; done
  sleep 2
  docker exec "$name" psql -U app -d postgres -c 'CREATE DATABASE app226' >/dev/null
  docker exec "$name" psql -U app -d app226 -c 'CREATE TABLE canary(id int)' >/dev/null
  if env -i PATH="$PATH" APP_ENV=local-dev DATABASE_URL='postgres://app:pw@postgres:5432/app226' DB_RESET_ALLOW_TEST_RUNNER=1 \
    DB_RESET_RUNNER="docker exec -i $name psql -v ON_ERROR_STOP=1 -U app -d postgres" bash "$script" >/dev/null 2>&1 &&
    [ "$(docker exec "$name" psql -U app -d app226 -Atc "select count(*) from pg_tables where tablename='canary'")" = 0 ]; then
    ok "real reset drops and recreates the database (ephemeral postgres)"
  else
    ko "real reset on ephemeral postgres"
  fi
fi

finish
