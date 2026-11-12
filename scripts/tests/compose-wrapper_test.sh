#!/usr/bin/env bash
# Tests for scripts/compose.sh and scripts/migrate-up.sh using a fake `docker` that records its argv.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin"
printf '#!/usr/bin/env bash\nprintf "%%s\\n" "$*" >>"%s/docker.log"\n' "$tmp" >"$tmp/bin/docker"
chmod +x "$tmp/bin/docker"
printf 'APP_ENV=local-dev\n' >"$tmp/env.local-dev"
printf 'APP_ENV=staging-test\nPUBLIC_HOST=s.example.test\nACME_EMAIL=a@example.test\n' >"$tmp/env.staging-test"

run() { : >"$tmp/docker.log"; rc=0; out="$(env -i PATH="$tmp/bin:$PATH" "$@" 2>&1)" || rc=$?; }

run COMPOSE_ENV_FILE="$tmp/env.local-dev" "$root/scripts/compose.sh" local-dev up -d
[ "$rc" -eq 0 ] && grep -q -- '--profile local-dev up -d' "$tmp/docker.log" && ok "up passes through with --profile" || ko "up ($rc $out)"

run COMPOSE_ENV_FILE="$tmp/env.local-dev" "$root/scripts/compose.sh" local-dev down -v
[ "$rc" -ne 0 ] && [ ! -s "$tmp/docker.log" ] && grep -q REFUSED <<<"$out" && ok "down -v refused" || ko "down -v ($rc)"

run COMPOSE_ENV_FILE="$tmp/env.local-dev" "$root/scripts/compose.sh" local-dev down --volumes
[ "$rc" -ne 0 ] && [ ! -s "$tmp/docker.log" ] && ok "down --volumes refused" || ko "down --volumes ($rc)"

run COMPOSE_ENV_FILE="$tmp/env.local-dev" "$root/scripts/compose.sh" local-dev exec -T postgres psql -v ON_ERROR_STOP=1 -U app
[ "$rc" -eq 0 ] && grep -q 'psql -v ON_ERROR_STOP=1' "$tmp/docker.log" && ok "psql -v is not mistaken for volumes" || ko "psql -v ($rc $out)"

run COMPOSE_ENV_FILE="$tmp/env.local-dev" "$root/scripts/compose.sh" staging-test up -d
[ "$rc" -ne 0 ] && [ ! -s "$tmp/docker.log" ] && ok "profile/APP_ENV mismatch stops before docker" || ko "mismatch ($rc)"

run COMPOSE_PROFILES=production-live COMPOSE_ENV_FILE="$tmp/env.local-dev" "$root/scripts/compose.sh" local-dev up -d
[ "$rc" -eq 0 ] && ! grep -q production-live "$tmp/docker.log" && ok "ambient COMPOSE_PROFILES ignored" || ko "ambient profiles ($rc)"

run "$root/scripts/compose.sh" local-dev
[ "$rc" -ne 0 ] && ok "no compose args refused" || ko "no args"

# migrate-up gating (docker never reached when refused)
for p in staging-test production-live; do
  run COMPOSE_ENV_FILE="$tmp/env.staging-test" "$root/scripts/migrate-up.sh" "$p"
  [ "$rc" -ne 0 ] && [ ! -s "$tmp/docker.log" ] && ok "migrate-up $p without confirmation refused" || ko "migrate $p ($rc)"
  run COMPOSE_ENV_FILE="$tmp/env.staging-test" MIGRATE_CONFIRM=local-dev BACKUP_EVIDENCE="$tmp/env.local-dev" "$root/scripts/migrate-up.sh" "$p"
  [ "$rc" -ne 0 ] && [ ! -s "$tmp/docker.log" ] && ok "migrate-up $p wrong confirmation refused" || ko "migrate wrong $p ($rc)"
done
run COMPOSE_ENV_FILE="$tmp/env.staging-test" MIGRATE_CONFIRM=staging-test "$root/scripts/migrate-up.sh" staging-test
[ "$rc" -ne 0 ] && [ ! -s "$tmp/docker.log" ] && ok "migrate-up without backup evidence refused" || ko "no backup ($rc)"
printf 'pg_dump evidence\n' >"$tmp/backup.evidence"
: >"$tmp/empty.evidence"
mkdir "$tmp/evidence-dir"
for ev in "backup-2026-11-12" "$tmp/nonexistent" "$tmp/empty.evidence" "$tmp/evidence-dir" ""; do
  run COMPOSE_ENV_FILE="$tmp/env.staging-test" MIGRATE_CONFIRM=staging-test BACKUP_EVIDENCE="$ev" "$root/scripts/migrate-up.sh" staging-test
  [ "$rc" -eq 2 ] && [ ! -s "$tmp/docker.log" ] && grep -q REFUSED <<<"$out" && ok "BACKUP_EVIDENCE='${ev##*/}' refused" || ko "evidence '$ev' ($rc $out)"
done
ln -s "$tmp/backup.evidence" "$tmp/link.evidence"
run COMPOSE_ENV_FILE="$tmp/env.staging-test" MIGRATE_CONFIRM=staging-test BACKUP_EVIDENCE="$tmp/link.evidence" "$root/scripts/migrate-up.sh" staging-test
[ "$rc" -eq 2 ] && [ ! -s "$tmp/docker.log" ] && ok "symlink evidence refused" || ko "symlink evidence ($rc)"
run COMPOSE_ENV_FILE="$tmp/env.staging-test" MIGRATE_CONFIRM=production-live BACKUP_EVIDENCE="$tmp/backup.evidence" "$root/scripts/migrate-up.sh" staging-test
[ "$rc" -eq 2 ] && [ ! -s "$tmp/docker.log" ] && ok "valid evidence but wrong MIGRATE_CONFIRM refused" || ko "wrong confirm with evidence ($rc)"
run COMPOSE_ENV_FILE="$tmp/env.staging-test" MIGRATE_CONFIRM=staging-test BACKUP_EVIDENCE="$tmp/backup.evidence" "$root/scripts/migrate-up.sh" staging-test
[ "$rc" -eq 0 ] && grep -q 'migrate-gated run --rm migrate-gated' "$tmp/docker.log" && ok "migrate-up staging with confirmation and valid evidence file runs gated job" || ko "gated run ($rc $out)"
run "$root/scripts/migrate-up.sh" prod
[ "$rc" -ne 0 ] && ok "migrate-up unknown profile refused" || ko "unknown profile"
finish
