#!/usr/bin/env bash
# Tests for scripts/deploy-check-bound.sh: the server/profile binding is checked before any copy.
# Usage: scripts/tests/deploy-check-bound_test.sh
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
chk="$root/scripts/deploy-check-bound.sh"
run() { rc=0; out="$("$chk" "$@" 2>&1)" || rc=$?; }

run staging-test ""
[ "$rc" -eq 0 ] && grep -q 'first deploy' <<<"$out" && ok "an absent state file is a first deploy" || ko "first deploy (rc=$rc)"
run staging-test staging-test
[ "$rc" -eq 0 ] && ok "a matching binding passes" || ko "match (rc=$rc)"
run production-live staging-test
[ "$rc" -eq 2 ] && grep -q 'Nothing was copied' <<<"$out" && ok "a server bound to another profile is refused before any copy" || ko "mismatch (rc=$rc)"
run local-dev ""
[ "$rc" -eq 2 ] && ok "an unknown profile is refused" || ko "unknown profile (rc=$rc)"
run staging-test
[ "$rc" -eq 64 ] && ok "a missing argument is a usage error" || ko "usage (rc=$rc)"
grep -n 'deploy-check-bound.sh' "$root/.github/workflows/deploy.yml" | head -1 | grep -q . && ok "the workflow runs the check" || ko "workflow does not call the check"
awk '/deploy-check-bound.sh/ {c=NR} /Copy the release/ {r=NR} END {exit !(c && r && c<r)}' "$root/.github/workflows/deploy.yml" && ok "the check comes before the release copy" || ko "check order"
finish
