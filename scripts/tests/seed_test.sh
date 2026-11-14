#!/usr/bin/env bash
# RED/GREEN evidence for the `seed` target (mk/50-ops.mk): it refuses every environment except
# APP_ENV=local-dev before any Docker/Compose call, then runs db/seeds/*.sql in lexical order through one
# psql transaction. A fake `docker` records argv and stdin, so no database is needed. The fixture SQL itself
# is proven on real PostgreSQL by the pgtest integration tests.
# Usage: scripts/tests/seed_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"

tmp="$(mktemp -d)"
seeds_dir="$root/db/seeds"
seed_file="$seeds_dir/00099_seed_test.sql"
created_seeds_dir=0

cleanup() {
  rm -f "$seed_file"
  # Only remove db/seeds when this test created it and nothing else was left in it.
  if [ "$created_seeds_dir" -eq 1 ]; then rmdir "$seeds_dir" 2>/dev/null || true; fi
  rm -rf "$tmp"
}
trap cleanup EXIT

mkdir "$tmp/bin"
cat >"$tmp/bin/docker" <<'FAKE'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$LOG"
cat >"$STDIN_LOG"
FAKE
chmod +x "$tmp/bin/docker"
printf 'APP_ENV=local-dev\n' >"$tmp/env.local-dev"

# run_seed DIR [VAR=value ...] -- : run `make -C DIR seed` under env -i; sets rc and out.
run_seed() {
  : >"$tmp/docker.log"
  : >"$tmp/stdin.log"
  local dir="$1"
  shift
  local envs=()
  while [ "$1" != "--" ]; do envs+=("$1"); shift; done
  shift
  rc=0
  out="$(env -i PATH="$tmp/bin:$PATH" HOME="$tmp" LOG="$tmp/docker.log" STDIN_LOG="$tmp/stdin.log" \
    "${envs[@]}" make -C "$dir" seed 2>&1)" || rc=$?
}

refused() {
  if [ "$rc" -eq 2 ] && grep -q '^REFUSED' <<<"$out" && [ ! -s "$tmp/docker.log" ]; then
    ok "$1"
  else
    ko "$1 (rc=$rc out=$out docker=$(cat "$tmp/docker.log" 2>/dev/null))"
  fi
}

run_seed "$root" --
refused "APP_ENV unset is refused"
run_seed "$root" APP_ENV= --
refused "APP_ENV empty is refused"
run_seed "$root" APP_ENV=staging-test --
refused "APP_ENV=staging-test is refused"
run_seed "$root" APP_ENV=production-live --
refused "APP_ENV=production-live is refused"
run_seed "$root" "APP_ENV=local-dev " --
refused "APP_ENV=local-dev with a trailing space is refused"
run_seed "$root" APP_ENV=LOCAL-DEV --
refused "APP_ENV=LOCAL-DEV (uppercase) is refused"
run_seed "$root" APP_ENV=local-dev COMPOSE_PROFILES=staging-test --
refused "COMPOSE_PROFILES=staging-test with APP_ENV=local-dev is refused"

# A scratch copy of the Makefile and mk/ has no db/seeds at all, so the missing-seeds guard is proven even
# once the repository ships real fixtures. Nothing reaches Docker, so scripts/ is not needed there.
scratch="$tmp/no-seeds-root"
mkdir -p "$scratch/mk"
cp "$root/Makefile" "$scratch/Makefile"
cp "$root"/mk/*.mk "$scratch/mk/"
run_seed "$scratch" APP_ENV=local-dev COMPOSE_ENV_FILE="$tmp/env.local-dev" --
refused "a missing db/seeds is refused"

# Positive path: a temporary SQL file, a fake docker that captures argv and stdin.
if [ ! -d "$seeds_dir" ]; then
  mkdir "$seeds_dir"
  created_seeds_dir=1
fi
printf 'SELECT 1;\n' >"$seed_file"

run_seed "$root" APP_ENV=local-dev COMPOSE_ENV_FILE="$tmp/env.local-dev" --
if [ "$rc" -eq 0 ] && grep -q 'psql -X -v ON_ERROR_STOP=1 -1' "$tmp/docker.log"; then
  ok "local-dev runs psql in one transaction with ON_ERROR_STOP"
else
  ko "local-dev run (rc=$rc out=$out docker=$(cat "$tmp/docker.log" 2>/dev/null))"
fi
if grep -q 'SELECT 1;' "$tmp/stdin.log"; then
  ok "the seed SQL is passed to psql on stdin"
else
  ko "the seed SQL was not passed on stdin (stdin=$(cat "$tmp/stdin.log" 2>/dev/null))"
fi

finish
