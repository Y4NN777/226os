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
mkdir -p "$app/scripts" "$app/secrets" "$app/deploy/traefik" "$tmp/bin" "$dr/env" "$dr/state" "$dr/backups"
cp "$root/scripts/deploy-remote.sh" "$app/scripts/" 2>/dev/null || true
printf 'routes: v1\n' >"$app/deploy/traefik/dynamic.edge.yml"
for pr in staging-test production-live; do printf 'age1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq\n' >"$app/secrets/$pr.recipients"; done
log="$tmp/calls.log"

stub() { # stub NAME BODY
  printf '#!/usr/bin/env bash\n%s\n' "$2" >"$app/scripts/$1"
  chmod +x "$app/scripts/$1"
}
stub compose.sh 'printf "compose %s (env=%s)\n" "$*" "${COMPOSE_ENV_FILE-unset}" >>"$CALLS"; exit "${COMPOSE_RC:-0}"'
stub backup.sh 'printf "backup %s (dir=%s)\n" "$*" "${BACKUP_DIR-unset}" >>"$CALLS"; echo PGDMP >"$BACKUP_DIR/$1-t.dump"; printf "backup: %s/%s-t.dump\nBACKUP_EVIDENCE=%s/%s-t.dump.evidence.txt\n" "$BACKUP_DIR" "$1" "$BACKUP_DIR" "$1"; echo evidence >"$BACKUP_DIR/$1-t.dump.evidence.txt"'
stub migrate-up.sh 'printf "migrate %s (confirm=%s evidence=%s)\n" "$*" "${MIGRATE_CONFIRM-}" "${BACKUP_EVIDENCE-}" >>"$CALLS"'
stub smoke.sh 'n=$(cat "$SMOKE_COUNT" 2>/dev/null || echo 0); n=$((n+1)); echo "$n" >"$SMOKE_COUNT"; printf "smoke %s\n" "$*" >>"$CALLS"; [ "$n" -ge "${SMOKE_OK_AT:-1}" ]'
printf '#!/usr/bin/env bash\nprintf "make %%s (APP_ENV=%%s)\\n" "$*" "${APP_ENV-}" >>"$CALLS"\n' >"$tmp/bin/make"
chmod +x "$tmp/bin/make"
# fake age: age -R recipients -o out in
cat >"$tmp/bin/age" <<'FAKEAGE'
#!/usr/bin/env bash
printf 'age %s\n' "$*" >>"$CALLS"
[ "$1" = -R ] && [ "$3" = -o ] || exit 64
printf 'ENCRYPTED\n' >"$4"
FAKEAGE
chmod +x "$tmp/bin/age"
# fake docker: the token goes to a file, never to the log; records the DOCKER_CONFIG the login ran with
cat >"$tmp/bin/docker" <<'FAKEDOCKER'
#!/usr/bin/env bash
if [ "$1" = login ]; then
  cat >"$TOKEN_SEEN"
  printf 'docker login cfg=%s\n' "${DOCKER_CONFIG-unset}" >>"$CALLS"
  mkdir -p "$DOCKER_CONFIG"
  echo auth >"$DOCKER_CONFIG/config.json"
fi
FAKEDOCKER
chmod +x "$tmp/bin/docker"
# sleep must not slow the tests down
printf '#!/usr/bin/env bash\nexit 0\n' >"$tmp/bin/sleep"
chmod +x "$tmp/bin/sleep"

deploy="$app/scripts/deploy-remote.sh"
run() {
  : >"$log"
  rm -f "$tmp/smoke.count"
  [ "${KEEP_STATE-}" = 1 ] || rm -f "$dr/state/profile"
  rc=0
  out="$(cd "$app" && env PATH="$tmp/bin:$PATH" CALLS="$log" TOKEN_SEEN="$tmp/token" TMPDIR="$tmp" SMOKE_COUNT="$tmp/smoke.count" DEPLOY_ROOT="$dr" "$@" 2>&1)" || rc=$?
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
grep -q "migrate staging-test (confirm=staging-test evidence=$dr/backups/staging-test-t.dump.evidence.txt)" "$log" && ok "migrate-up gets MIGRATE_CONFIRM and the BACKUP_EVIDENCE of this backup" || ko "migrate env ($(cat "$log"))"
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

# --- findings of the security review -------------------------------------------------------------------------------
# M3: the dump is encrypted to the recipients and the plaintext is gone; the evidence stays for migrate-up.
run "$deploy" staging-test https://staging.example.test
if [ "$rc" -eq 0 ] && grep -q "age -R $app/secrets/staging-test.recipients -o $dr/backups/staging-test-t.dump.age $dr/backups/staging-test-t.dump" "$log" &&
  [ -f "$dr/backups/staging-test-t.dump.age" ] && [ ! -e "$dr/backups/staging-test-t.dump" ] && [ -f "$dr/backups/staging-test-t.dump.evidence.txt" ]; then
  ok "the backup is encrypted with age to secrets/<profile>.recipients and the plaintext dump is removed"
else ko "backup encryption (rc=$rc out=$out)"; fi

stub migrate-up.sh 'echo migrate >>"$CALLS"; exit 3'
run "$deploy" staging-test https://staging.example.test
if [ "$rc" -eq 3 ] && grep -q 'FAILED (exit 3) at stage: migrate' <<<"$out" && grep -q 'migrations: UNKNOWN' <<<"$out" &&
  grep -q 'staging-test-t.dump.age' <<<"$out" && grep -q 'docs/SHIP.md section 5' <<<"$out"; then
  ok "a failed migration reports the stage, the migration state, the backup and the rollback hint"
else ko "partial deploy report (rc=$rc out=$out)"; fi
stub migrate-up.sh 'printf "migrate %s (confirm=%s evidence=%s)\n" "$*" "${MIGRATE_CONFIRM-}" "${BACKUP_EVIDENCE-}" >>"$CALLS"'
stub smoke.sh 'echo smoke >>"$CALLS"; exit 1'
run SMOKE_ATTEMPTS=1 "$deploy" staging-test https://staging.example.test
[ "$rc" -eq 1 ] && grep -q 'migrations: APPLIED' <<<"$out" && grep -q 'stage: smoke' <<<"$out" && ok "a failed smoke reports that migrations were applied" || ko "smoke report (rc=$rc out=$out)"
stub smoke.sh 'n=$(cat "$SMOKE_COUNT" 2>/dev/null || echo 0); n=$((n+1)); echo "$n" >"$SMOKE_COUNT"; printf "smoke %s\n" "$*" >>"$CALLS"; [ "$n" -ge "${SMOKE_OK_AT:-1}" ]'

# M2: a server is bound to the profile of its first deploy.
run "$deploy" staging-test https://staging.example.test
[ "$(cat "$dr/state/profile")" = staging-test ] && ok "the first deploy records the profile in state/profile" || ko "profile state"
KEEP_STATE=1 run "$deploy" production-live https://prod.example.test
[ "$rc" -eq 2 ] && grep -q "bound to profile 'staging-test'" <<<"$out" && [ ! -s "$log" ] && ok "a server bound to staging-test refuses a production-live deploy before anything runs" || ko "profile binding (rc=$rc out=$out)"

# M1: the registry token is read from stdin, used with a private DOCKER_CONFIG that is deleted before the deploy goes on.
: >"$tmp/token"
rc=0
: >"$log"
rm -f "$dr/state/profile"
out="$(cd "$app" && printf 'sekret-token' | env PATH="$tmp/bin:$PATH" CALLS="$log" TOKEN_SEEN="$tmp/token" TMPDIR="$tmp" SMOKE_COUNT="$tmp/smoke.count" DEPLOY_ROOT="$dr" REGISTRY_TOKEN_STDIN=1 "$deploy" staging-test https://staging.example.test 2>&1)" || rc=$?
cfg="$(sed -n 's/^docker login cfg=//p' "$log")"
if [ "$rc" -eq 0 ] && [ "$(cat "$tmp/token")" = sekret-token ] && [ -n "$cfg" ] && [ ! -e "$cfg" ] && grep -q "compose staging-test --profile migrate-gated pull" "$log" && ! grep -q sekret "$log"; then
  ok "the registry login uses a private DOCKER_CONFIG that is removed after the pull, token never logged"
else ko "registry login (rc=$rc cfg=$cfg out=$out)"; fi
[ "$(grep -n 'docker login' "$log" | cut -d: -f1)" -lt "$(grep -n 'pull' "$log" | cut -d: -f1)" ] && ok "login happens before the pull" || ko "login order"

finish
