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
trap 'rm -rf "$tmp"' EXIT

# scratch_root DIR: a copy of the Makefile, mk/ and the two Compose wrapper scripts, so a run never reads
# or writes the repository's db/seeds (a hard-killed run would otherwise leave a stray seed file behind).
scratch_root() {
  mkdir -p "$1/mk" "$1/scripts"
  cp "$root/Makefile" "$1/Makefile"
  cp "$root"/mk/*.mk "$1/mk/"
  cp "$root/scripts/compose.sh" "$root/scripts/compose-env-guard.sh" "$1/scripts/"
}

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

# The case-sensitivity message is only for a local-dev spelled with other letter cases.
run_seed "$root" APP_ENV=Local-Dev --
refused "APP_ENV=Local-Dev (mixed case) is refused"
grep -q 'case-sensitive' <<<"$out" && ok "a mixed-case local-dev gets the case-sensitive message" \
  || ko "a mixed-case local-dev gets the case-sensitive message (out=$out)"
run_seed "$root" APP_ENV=123 --
refused "APP_ENV=123 is refused"
if grep -q 'only runs with APP_ENV=local-dev' <<<"$out" && ! grep -q 'case-sensitive' <<<"$out"; then
  ok "a value with no letters gets the generic message"
else ko "a value with no letters gets the generic message (out=$out)"; fi

# A scratch root with no db/seeds at all, so the missing-seeds guard is proven even once the repository
# ships real fixtures.
scratch="$tmp/no-seeds-root"
scratch_root "$scratch"
run_seed "$scratch" APP_ENV=local-dev COMPOSE_ENV_FILE="$tmp/env.local-dev" --
refused "a missing db/seeds is refused"

# Positive path, in a scratch root with its own db/seeds: two files written out of order prove the lexical
# order, and a fake docker captures argv and stdin.
scratch="$tmp/seeds-root"
scratch_root "$scratch"
mkdir "$scratch/db" "$scratch/db/seeds"
printf 'SELECT 2;\n' >"$scratch/db/seeds/00002_b.sql"
printf 'SELECT 1;\n' >"$scratch/db/seeds/00001_a.sql"

run_seed "$scratch" APP_ENV=local-dev COMPOSE_ENV_FILE="$tmp/env.local-dev" --
if [ "$rc" -eq 0 ] && grep -q 'psql -X -v ON_ERROR_STOP=1 -1' "$tmp/docker.log"; then
  ok "local-dev runs psql in one transaction with ON_ERROR_STOP"
else
  ko "local-dev run (rc=$rc out=$out docker=$(cat "$tmp/docker.log" 2>/dev/null))"
fi
if [ "$(cat "$tmp/stdin.log" 2>/dev/null)" = $'SELECT 1;\nSELECT 2;' ]; then
  ok "the seed SQL is passed to psql on stdin, in lexical file order"
else
  ko "the seed SQL was not passed on stdin in lexical order (stdin=$(cat "$tmp/stdin.log" 2>/dev/null))"
fi
if [ -z "$(find "$root/db/seeds" -newer "$tmp/env.local-dev" -name '*.sql' 2>/dev/null)" ] && [ ! -e "$root/db/seeds/00099_seed_test.sql" ]; then
  ok "the repository db/seeds is never written by this test"
else
  ko "the repository db/seeds was written by this test"
fi

finish
