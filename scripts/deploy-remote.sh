#!/usr/bin/env bash
# Runs ON THE SERVER, from the release directory the workflow just unpacked (/opt/226os/app), as the `deploy` user.
# It only sequences the existing scripts; their guards stay in charge (compose-env-guard, migrate-up, backup, smoke).
#
# Usage: scripts/deploy-remote.sh <staging-test | production-live> <https://public-host>
#
#   0. the server is tied to ONE profile (state/profile, written by the first deploy): another profile is refused
#   1. pull the images (fails fast on registry access, before the database is touched). With REGISTRY_TOKEN_STDIN=1
#      the registry token is read from stdin, used with a private DOCKER_CONFIG directory and deleted right after
#      the pull: no registry login stays on the host.
#   2. start PostgreSQL only (a fresh server has no database to back up yet)
#   3. scripts/backup.sh, the dump is ENCRYPTED with age to secrets/<profile>.recipients (public keys) and the
#      plaintext removed; then scripts/migrate-up.sh with MIGRATE_CONFIRM and that backup's BACKUP_EVIDENCE
#   4. up -d --wait for the whole profile (images by digest, from the env file); Traefik is recreated when its
#      one bind-mounted file (deploy/traefik/dynamic.edge.yml) changed, because Compose cannot see a content change
#   5. staging-test only: `make seed` (demo fixtures, idempotent). production-live never gets seeds.
#   6. scripts/smoke.sh against the public URL, retried (a first certificate takes a moment); failure fails the deploy
# A failure prints the stage, whether the migrations were applied, the backup file and the rollback hint.
#
# Env: DEPLOY_ROOT (default /opt/226os; holds env/.env.<profile> mode 600, backups/, state/), SMOKE_ATTEMPTS (8),
# SMOKE_DELAY seconds (15), KEEP_BACKUPS (10). The env file is passed to every script through COMPOSE_ENV_FILE:
# the server holds no age private key and no secrets/ directory (only the public recipients file).
set -eEuo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
usage() { printf 'usage: %s <staging-test | production-live> <https://public-host>\n' "$0" >&2; exit 64; }
refuse() { printf 'deploy-remote: REFUSED: %s\n' "$*" >&2; exit 2; }
step() {
  stage="$*"
  printf '\n== %s ==\n' "$*"
}

[ $# -eq 2 ] || usage
profile="$1"
base="$2"
case "$profile" in staging-test | production-live) ;; *) refuse "profile must be staging-test or production-live (got '$profile')" ;; esac
[[ "$base" =~ ^https://[a-z0-9.-]+$ ]] || refuse "the base URL must be https://<host>"

deploy_root="${DEPLOY_ROOT:-/opt/226os}"
envfile="$deploy_root/env/.env.$profile"
[ -f "$envfile" ] && [ ! -L "$envfile" ] || refuse "env file $envfile not found (the workflow copies it before this script)"
[ "$(stat -c %a "$envfile")" = 600 ] || refuse "env file $envfile must be mode 600"
recipients="$root/secrets/$profile.recipients"
[ -s "$recipients" ] || refuse "$recipients is missing: backups cannot be encrypted (the workflow ships it)"
backups="$deploy_root/backups"
state="$deploy_root/state"
mkdir -p "$backups" "$state"
chmod 700 "$backups"

# 0. One server, one profile: a staging deploy can never land on the production server, or the reverse.
if [ -f "$state/profile" ]; then
  bound="$(cat "$state/profile")"
  [ "$bound" = "$profile" ] || refuse "this server is bound to profile '$bound', not '$profile' (state/profile): wrong DEPLOY_HOST for this environment?"
else
  printf '%s\n' "$profile" >"$state/profile"
fi

stage="start"
migrated="no"
backup_file="none yet"
on_error() {
  local rc=$?
  [ "$BASHPID" = "$$" ] || exit "$rc"
  report "$rc"
  exit "$rc"
}
report() {
  {
    printf '\ndeploy-remote: FAILED (exit %s) at stage: %s\n' "$1" "$stage"
    case "$migrated" in
      yes) printf '  migrations: APPLIED (the database is on the new schema)\n' ;;
      no) [ "$stage" = "migrate (forward-only, gated)" ] || printf '  migrations: not applied by this run\n' ;;
    esac
    [ "$stage" != "migrate (forward-only, gated)" ] || printf '  migrations: UNKNOWN (the migration job failed: check scripts/compose.sh %s --profile migrate-gated logs, and the goose version)\n' "$profile"
    printf '  backup (age-encrypted): %s\n' "$backup_file"
    printf '  rollback: redeploy the previous image digests; an application rollback does not undo a migration,\n'
    printf '            use the tested restore plan (docs/SHIP.md section 5, scripts/restore-check.sh).\n'
  } >&2
}
trap on_error ERR

export COMPOSE_ENV_FILE="$envfile"
cd "$root"
compose() { scripts/compose.sh "$profile" "$@"; }

step "pull the images"
if [ "${REGISTRY_TOKEN_STDIN-}" = 1 ]; then
  docker_cfg="$(mktemp -d "${TMPDIR:-/tmp}/docker-config.XXXXXX")"
  trap 'rm -rf "$docker_cfg"' EXIT
  DOCKER_CONFIG="$docker_cfg" docker login ghcr.io --username github-actions --password-stdin
  DOCKER_CONFIG="$docker_cfg" compose --profile migrate-gated pull api web migrate-gated
  rm -rf "$docker_cfg"
  trap - EXIT
else
  compose --profile migrate-gated pull api web migrate-gated
fi

step "start PostgreSQL"
compose up -d --wait postgres

step "backup before migration"
backup_out="$(BACKUP_DIR="$backups" scripts/backup.sh "$profile")"
printf '%s\n' "$backup_out"
dump="$(sed -n 's/^backup: //p' <<<"$backup_out" | tail -n 1)"
evidence="$(sed -n 's/^BACKUP_EVIDENCE=//p' <<<"$backup_out" | tail -n 1)"
[ -n "$evidence" ] || refuse "scripts/backup.sh gave no BACKUP_EVIDENCE"
[ -f "$dump" ] || refuse "scripts/backup.sh gave no dump file"
# The dump holds every user's data: encrypt it to the public age recipients, then remove the plaintext. The evidence
# file only holds the metadata (sizes, counts, the hash of the plaintext).
backup_file="$dump.age"
age -R "$recipients" -o "$backup_file" "$dump"
[ -s "$backup_file" ] || refuse "the encrypted backup is empty; the plaintext dump was kept at $dump"
rm -f -- "$dump"
chmod 600 "$backup_file"

step "migrate (forward-only, gated)"
MIGRATE_CONFIRM="$profile" BACKUP_EVIDENCE="$evidence" scripts/migrate-up.sh "$profile"
migrated="yes"

step "start the profile by digest"
compose up -d --wait
new_sum="$(sha256sum deploy/traefik/dynamic.edge.yml | cut -d' ' -f1)"
old_sum="$(cat "$state/dynamic.edge.sha256" 2>/dev/null || echo none)"
if [ "$new_sum" != "$old_sum" ]; then
  step "recreate Traefik (its routes file changed)"
  compose up -d --wait --force-recreate traefik
fi
printf '%s\n' "$new_sum" >"$state/dynamic.edge.sha256"

if [ "$profile" = staging-test ]; then
  step "demo seeds (staging-test only)"
  APP_ENV=staging-test make seed
fi

step "smoke"
attempts="${SMOKE_ATTEMPTS:-8}"
delay="${SMOKE_DELAY:-15}"
i=1
until scripts/smoke.sh "$profile" "$base"; do
  if [ "$i" -ge "$attempts" ]; then
    printf 'deploy-remote: FAIL: smoke did not pass after %s attempts\n' "$attempts" >&2
    report 1
    exit 1
  fi
  printf 'deploy-remote: smoke attempt %s of %s failed, retrying in %ss\n' "$i" "$attempts" "$delay" >&2
  i=$((i + 1))
  sleep "$delay"
done

# Retention: keep the newest encrypted dumps (and their evidence files); never touches anything else.
keep="${KEEP_BACKUPS:-10}"
# shellcheck disable=SC2012 # names are generated by backup.sh (profile-timestamp.dump), no odd characters
{ ls -1t "$backups"/"$profile"-*.dump.age 2>/dev/null || true; } | tail -n +"$((keep + 1))" | while IFS= read -r old; do rm -f -- "$old" "${old%.age}.evidence.txt"; done
printf '\ndeploy-remote: %s deployed\n' "$profile"
