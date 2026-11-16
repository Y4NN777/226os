#!/usr/bin/env bash
# RED/GREEN evidence for scripts/content/run-free-series.sh (S4 content script, B5/H19): the gate refuses without
# a valid owner username, an explicit confirmation and backup evidence outside local-dev, and never reaches docker
# when it refuses. A fake `docker` records argv and stdin, so no database is needed. The SQL itself is proven on
# real PostgreSQL by TestFreeSeriesContentScript (pgtest).
# Usage: scripts/tests/free-series_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
script="$root/scripts/content/run-free-series.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/bin"
printf 'APP_ENV=local-dev\n' >"$tmp/env.local-dev"
printf 'APP_ENV=staging-test\nPUBLIC_HOST=s.example.test\nACME_EMAIL=a@example.test\n' >"$tmp/env.staging-test"
printf 'backup: x\n' >"$tmp/evidence.txt"
: >"$tmp/empty.txt"
ln -s "$tmp/evidence.txt" "$tmp/link.txt"

[ -x "$script" ] || { ko "scripts/content/run-free-series.sh is missing or not executable"; finish; exit 1; }

cat >"$tmp/bin/docker" <<'FAKE'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$LOG"
cat >"$STDIN_LOG"
FAKE
chmod +x "$tmp/bin/docker"

run() { # ENV... -- ARGS
  : >"$tmp/docker.log"
  : >"$tmp/stdin.log"
  local envs=()
  while [ "$1" != "--" ]; do envs+=("$1"); shift; done
  shift
  rc=0
  out="$(env -i PATH="$tmp/bin:$PATH" HOME="$tmp" LOG="$tmp/docker.log" STDIN_LOG="$tmp/stdin.log" "${envs[@]}" "$script" "$@" 2>&1)" || rc=$?
}

refused() { # name
  if [ "$rc" -eq 2 ] && grep -q '^REFUSED' <<<"$out" && [ ! -s "$tmp/docker.log" ]; then ok "$1"; else ko "$1 (rc=$rc out=$out docker=$(cat "$tmp/docker.log"))"; fi
}

local_env="COMPOSE_ENV_FILE=$tmp/env.local-dev"
staging_env="COMPOSE_ENV_FILE=$tmp/env.staging-test"

run "$local_env" -- local-dev
refused "no OWNER_USERNAME is refused"
run "$local_env" OWNER_USERNAME="Bad Name" -- local-dev
refused "an owner username outside the profile rule is refused"
run "$local_env" OWNER_USERNAME="ok-name -c x=y" -- local-dev
refused "an owner username cannot inject PGOPTIONS"
run "$local_env" OWNER_USERNAME=owner -- staging
refused "an unknown profile is refused"
run "$staging_env" OWNER_USERNAME=owner -- staging-test
refused "staging-test without CONTENT_CONFIRM is refused"
run "$staging_env" OWNER_USERNAME=owner CONTENT_CONFIRM=production-live BACKUP_EVIDENCE="$tmp/evidence.txt" -- staging-test
refused "a confirmation for another profile is refused"
run "$staging_env" OWNER_USERNAME=owner CONTENT_CONFIRM=staging-test -- staging-test
refused "staging-test without BACKUP_EVIDENCE is refused"
for ev in "$tmp/missing.txt" "$tmp/empty.txt" "$tmp/link.txt"; do
  run "$staging_env" OWNER_USERNAME=owner CONTENT_CONFIRM=staging-test BACKUP_EVIDENCE="$ev" -- staging-test
  refused "backup evidence $(basename "$ev") is refused"
done

run "$local_env" OWNER_USERNAME=owner -- local-dev
if [ "$rc" -eq 0 ] && grep -q 'PGOPTIONS=-c content.owner_username=owner' "$tmp/docker.log" && grep -q 'ON_ERROR_STOP=1' "$tmp/docker.log"; then
  ok "local-dev runs psql with the owner setting and ON_ERROR_STOP"
else
  ko "local-dev run (rc=$rc out=$out docker=$(cat "$tmp/docker.log"))"
fi
cmp -s "$tmp/stdin.log" "$root/scripts/content/free-series.sql" && ok "the reviewed SQL file is what psql reads" || ko "stdin is not free-series.sql"

run "$staging_env" OWNER_USERNAME=owner CONTENT_CONFIRM=staging-test BACKUP_EVIDENCE="$tmp/evidence.txt" -- staging-test
if [ "$rc" -eq 0 ] && grep -q "backup evidence=$tmp/evidence.txt" <<<"$out" && [ -s "$tmp/docker.log" ]; then
  ok "staging-test runs with a confirmation and backup evidence"
else
  ko "staging-test run (rc=$rc out=$out)"
fi

finish
