#!/usr/bin/env bash
# Tests for scripts/deploy-remote.sh (the part that runs on the server) in a scratch tree where compose.sh,
# backup.sh, migrate-up.sh and smoke.sh are stubs that log their calls and `make` is a fake: the order of the steps,
# the BACKUP_EVIDENCE hand-over, seeds on staging only, the smoke retries and the Traefik recreate rule.
# Usage: scripts/tests/deploy-remote_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
app="$tmp/app"
dr="$tmp/226os"
mkdir -p "$app/scripts" "$app/deploy/traefik" "$tmp/bin" "$dr/env" "$dr/state" "$dr/backups"
cp "$root/scripts/deploy-remote.sh" "$app/scripts/" 2>/dev/null || true
printf 'routes: v1\n' >"$app/deploy/traefik/dynamic.edge.yml"
log="$tmp/calls.log"

stub() { # stub NAME BODY
  printf '#!/usr/bin/env bash\n%s\n' "$2" >"$app/scripts/$1"
  chmod +x "$app/scripts/$1"
}
stub compose.sh 'printf "compose %s (env=%s)\n" "$*" "${COMPOSE_ENV_FILE-unset}" >>"$CALLS"; exit "${COMPOSE_RC:-0}"'
stub backup.sh 'printf "backup %s (dir=%s)\n" "$*" "${BACKUP_DIR-unset}" >>"$CALLS"; printf "backup: x.dump\nBACKUP_EVIDENCE=%s/x.dump.evidence.txt\n" "$BACKUP_DIR"; echo evidence >"$BACKUP_DIR/x.dump.evidence.txt"'
stub migrate-up.sh 'printf "migrate %s (confirm=%s evidence=%s)\n" "$*" "${MIGRATE_CONFIRM-}" "${BACKUP_EVIDENCE-}" >>"$CALLS"'
stub smoke.sh 'n=$(cat "$SMOKE_COUNT" 2>/dev/null || echo 0); n=$((n+1)); echo "$n" >"$SMOKE_COUNT"; printf "smoke %s\n" "$*" >>"$CALLS"; [ "$n" -ge "${SMOKE_OK_AT:-1}" ]'
printf '#!/usr/bin/env bash\nprintf "make %%s (APP_ENV=%%s)\\n" "$*" "${APP_ENV-}" >>"$CALLS"\n' >"$tmp/bin/make"
chmod +x "$tmp/bin/make"
# sleep must not slow the tests down
printf '#!/usr/bin/env bash\nexit 0\n' >"$tmp/bin/sleep"
chmod +x "$tmp/bin/sleep"

deploy="$app/scripts/deploy-remote.sh"
run() {
  : >"$log"
  rm -f "$tmp/smoke.count"
  rc=0
  out="$(cd "$app" && env PATH="$tmp/bin:$PATH" CALLS="$log" SMOKE_COUNT="$tmp/smoke.count" DEPLOY_ROOT="$dr" "$@" 2>&1)" || rc=$?
}
n_of() { grep -n "$1" "$log" | head -n 1 | cut -d: -f1; }
printf 'APP_ENV=staging-test\n' >"$dr/env/.env.staging-test"
chmod 600 "$dr/env/.env.staging-test"
printf 'APP_ENV=production-live\n' >"$dr/env/.env.production-live"
chmod 600 "$dr/env/.env.production-live"

run "$deploy" staging-test https://staging.example.test
if [ "$rc" -eq 0 ]; then
  p="$(n_of 'pull')" u1="$(n_of 'up -d --wait postgres')" b="$(n_of '^backup')" m="$(n_of '^migrate')" u2="$(n_of 'up -d --wait (env')" s="$(n_of 'make seed')" k="$(n_of '^smoke')"
  if [ -n "$p" ] && [ -n "$u1" ] && [ -n "$b" ] && [ -n "$m" ] && [ -n "$u2" ] && [ -n "$s" ] && [ -n "$k" ] && [ "$p" -lt "$u1" ] && [ "$u1" -lt "$b" ] && [ "$b" -lt "$m" ] && [ "$m" -lt "$u2" ] && [ "$u2" -lt "$s" ] && [ "$s" -lt "$k" ]; then
    ok "staging: pull, postgres up, backup, migrate, up, seed, smoke in that order"
  else ko "staging order ($(tr '\n' '|' <"$log"))"; fi
else ko "staging run (rc=$rc out=$out)"; fi
grep -q "migrate staging-test (confirm=staging-test evidence=$dr/backups/x.dump.evidence.txt)" "$log" && ok "migrate-up gets MIGRATE_CONFIRM and the BACKUP_EVIDENCE of this backup" || ko "migrate env ($(cat "$log"))"
grep -q "dir=$dr/backups" "$log" && grep -q "env=$dr/env/.env.staging-test" "$log" && ok "backups live outside the app directory and compose reads the server env file" || ko "paths ($(cat "$log"))"
grep -q 'make seed (APP_ENV=staging-test)' "$log" && ok "staging seeds with APP_ENV=staging-test" || ko "seed env"
grep -q 'smoke staging-test https://staging.example.test' "$log" && ok "the smoke targets the public URL" || ko "smoke args"

run "$deploy" production-live https://prod.example.test
if [ "$rc" -eq 0 ] && ! grep -q 'make' "$log" && grep -q '^backup' "$log" && grep -q '^migrate' "$log" && grep -q '^smoke production-live' "$log"; then
  ok "production: backup, migrate, up, smoke, and never a seed"
else ko "production (rc=$rc log=$(tr '\n' '|' <"$log"))"; fi

run "$deploy" local-dev https://x.example.test
[ "$rc" -ne 0 ] && [ ! -s "$log" ] && ok "local-dev is refused before anything runs" || ko "local-dev (rc=$rc)"

run "$deploy" staging-test http://staging.example.test
[ "$rc" -ne 0 ] && [ ! -s "$log" ] && ok "a non-HTTPS base URL is refused before anything runs" || ko "http url (rc=$rc)"

run "$deploy" staging-test
[ "$rc" -eq 64 ] && ok "a missing argument is a usage error" || ko "usage (rc=$rc)"

mv "$dr/env/.env.staging-test" "$dr/env/gone"
run "$deploy" staging-test https://staging.example.test
[ "$rc" -ne 0 ] && [ ! -s "$log" ] && ok "a missing env file is refused before anything runs" || ko "missing env (rc=$rc)"
mv "$dr/env/gone" "$dr/env/.env.staging-test"
chmod 644 "$dr/env/.env.staging-test"
run "$deploy" staging-test https://staging.example.test
[ "$rc" -ne 0 ] && grep -q '600' <<<"$out" && [ ! -s "$log" ] && ok "an env file readable by others is refused" || ko "env mode (rc=$rc)"
chmod 600 "$dr/env/.env.staging-test"

# A failing step stops the deploy: no seed, no smoke after a failed migration.
stub migrate-up.sh 'echo "migrate" >>"$CALLS"; exit 2'
run "$deploy" staging-test https://staging.example.test
[ "$rc" -ne 0 ] && ! grep -q 'make seed' "$log" && ! grep -q '^smoke' "$log" && ok "a failed migration stops the deploy before up, seed and smoke" || ko "migrate failure (rc=$rc log=$(tr '\n' '|' <"$log"))"
stub migrate-up.sh 'printf "migrate %s (confirm=%s evidence=%s)\n" "$*" "${MIGRATE_CONFIRM-}" "${BACKUP_EVIDENCE-}" >>"$CALLS"'

# Smoke retries (the certificate may need a moment on a first deploy), then fails closed.
run SMOKE_OK_AT=3 SMOKE_ATTEMPTS=5 "$deploy" staging-test https://staging.example.test
[ "$rc" -eq 0 ] && [ "$(grep -c '^smoke' "$log")" -eq 3 ] && ok "the smoke is retried until it passes" || ko "smoke retry (rc=$rc n=$(grep -c '^smoke' "$log"))"
run SMOKE_OK_AT=99 SMOKE_ATTEMPTS=3 "$deploy" staging-test https://staging.example.test
[ "$rc" -ne 0 ] && [ "$(grep -c '^smoke' "$log")" -eq 3 ] && grep -q 'FAIL' <<<"$out" && ok "a smoke that never passes fails the deploy" || ko "smoke fail (rc=$rc)"

# Traefik reads one bind-mounted file: when its content changes the container is recreated.
rm -f "$dr/state/dynamic.edge.sha256"
run "$deploy" staging-test https://staging.example.test
run "$deploy" staging-test https://staging.example.test
! grep -q 'force-recreate traefik' "$log" && ok "an unchanged Traefik file does not recreate Traefik" || ko "recreate without change"
printf 'routes: v2\n' >"$app/deploy/traefik/dynamic.edge.yml"
run "$deploy" staging-test https://staging.example.test
grep -q 'up -d --wait --force-recreate traefik' "$log" && ok "a changed Traefik file recreates Traefik" || ko "no recreate after change ($(tr '\n' '|' <"$log"))"

finish
