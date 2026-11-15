#!/usr/bin/env bash
# Tests for scripts/deploy-plan.sh: which event deploys what, and when nothing must deploy. A scratch git
# repository stands in for the checkout; origin/develop is a local ref. Nothing touches a server or the network.
# Usage: scripts/tests/deploy-plan_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
plan="$root/scripts/deploy-plan.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
repo="$tmp/repo"
git init -q "$repo"
git -C "$repo" config user.email t@example.test
git -C "$repo" config user.name t
git -C "$repo" commit -q --allow-empty -m one
sha1="$(git -C "$repo" rev-parse HEAD)"
git -C "$repo" commit -q --allow-empty -m two
sha2="$(git -C "$repo" rev-parse HEAD)"
git -C "$repo" update-ref refs/remotes/origin/develop "$sha2"
git -C "$repo" update-ref refs/remotes/origin/main "$sha2"
tree="$(git -C "$repo" rev-parse "$sha2^{tree}")"
git -C "$repo" checkout -q --detach "$sha1"
git -C "$repo" commit -q --allow-empty -m side
side="$(git -C "$repo" rev-parse HEAD)"

# run VAR=value...: the plan script in the scratch repo with only the given variables; outputs in $tmp/out.
run() {
  : >"$tmp/out"
  rc=0
  msg="$(cd "$repo" && env -i PATH="$PATH" HOME="$tmp" GITHUB_OUTPUT="$tmp/out" THIS_REPO=o/r "$@" "$plan" 2>&1)" || rc=$?
}
val() { grep -m1 "^$1=" "$tmp/out" | cut -d= -f2- || true; }

wr() { printf 'EVENT_NAME=workflow_run RUN_CONCLUSION=success RUN_EVENT=push RUN_BRANCH=develop RUN_REPO=o/r RUN_SHA=%s' "$1"; }

# shellcheck disable=SC2046 # the variable list is intentionally word-split
run $(wr "$sha2")
if [ "$rc" -eq 0 ] && [ "$(val proceed)" = true ] && [ "$(val target)" = staging ] && [ "$(val profile)" = staging-test ] &&
  [ "$(val sha)" = "$sha2" ] && [ "$(val tree)" = "$tree" ]; then ok "a green Quality run on the develop tip deploys to staging-test"; else ko "staging (rc=$rc $msg)"; fi

# shellcheck disable=SC2046
run $(wr "$sha1")
[ "$rc" -eq 0 ] && [ "$(val proceed)" = false ] && grep -q superseded "$tmp/out" && ok "an older develop commit is superseded, not deployed" || ko "superseded (rc=$rc $(cat "$tmp/out"))"

# shellcheck disable=SC2046
run $(wr "$side")
[ "$rc" -ne 0 ] && ok "a commit that is not on develop fails closed" || ko "off-develop commit (rc=$rc)"

# shellcheck disable=SC2046
run $(wr "$sha2") RUN_CONCLUSION=failure
[ "$rc" -eq 0 ] && [ "$(val proceed)" = false ] && ok "a failed Quality run deploys nothing" || ko "failed quality (rc=$rc)"

# shellcheck disable=SC2046
run $(wr "$sha2") RUN_EVENT=pull_request
[ "$rc" -eq 0 ] && [ "$(val proceed)" = false ] && ok "a pull request run never deploys" || ko "pull_request run (rc=$rc)"

# shellcheck disable=SC2046
run $(wr "$sha2") RUN_REPO=fork/r
[ "$rc" -eq 0 ] && [ "$(val proceed)" = false ] && ok "a run from another repository never deploys" || ko "fork run (rc=$rc)"

# shellcheck disable=SC2046
run $(wr "$sha2") RUN_BRANCH=feature/x
[ "$rc" -eq 0 ] && [ "$(val proceed)" = false ] && ok "a run on another branch never deploys" || ko "other branch (rc=$rc)"

run EVENT_NAME=push PUSH_REF=refs/heads/main PUSH_SHA="$sha2"
if [ "$rc" -eq 0 ] && [ "$(val proceed)" = true ] && [ "$(val target)" = production ] && [ "$(val profile)" = production-live ] &&
  [ "$(val tree)" = "$tree" ] && [ "$(val sha)" = "$sha2" ]; then ok "a push to main targets production-live with the tree of that commit"; else ko "production (rc=$rc $msg)"; fi

run EVENT_NAME=push PUSH_REF=refs/heads/feature/x PUSH_SHA="$sha2"
[ "$rc" -ne 0 ] && ok "a push to another branch is refused" || ko "push other branch (rc=$rc)"

run EVENT_NAME=workflow_dispatch
[ "$rc" -ne 0 ] && ok "an unexpected event is refused" || ko "unexpected event (rc=$rc)"

# Emergency path: a manual dispatch on main, explicit opt-in and a reason; production without the staging digests.
ed() { printf 'EVENT_NAME=workflow_dispatch DISPATCH_REF=refs/heads/main DISPATCH_SHA=%s CONFIRM=EMERGENCY' "$1"; }

# shellcheck disable=SC2046
run $(ed "$sha2") 'EMERGENCY_REASON=hotfix login broken'
if [ "$rc" -eq 0 ] && [ "$(val proceed)" = true ] && [ "$(val target)" = production ] && [ "$(val profile)" = production-live ] &&
  [ "$(val emergency)" = true ] && [ "$(val sha)" = "$sha2" ] && [ "$(val tree)" = "$tree" ]; then
  ok "an emergency dispatch on main with a reason targets production-live, flagged emergency"
else ko "emergency (rc=$rc $msg)"; fi

# shellcheck disable=SC2046
run $(ed "$sha2") CONFIRM=yes 'EMERGENCY_REASON=x'
[ "$rc" -ne 0 ] && ok "a dispatch without CONFIRM=EMERGENCY is refused" || ko "dispatch without opt-in (rc=$rc)"

# shellcheck disable=SC2046
run $(ed "$sha2") DISPATCH_REF=refs/heads/develop 'EMERGENCY_REASON=x'
[ "$rc" -ne 0 ] && ok "an emergency dispatch from another ref than main is refused" || ko "dispatch from develop (rc=$rc)"

# shellcheck disable=SC2046
run $(ed "$sha2")
[ "$rc" -ne 0 ] && ok "an emergency dispatch without a reason is refused" || ko "no reason (rc=$rc)"

# shellcheck disable=SC2046
run $(ed "$sha2") "EMERGENCY_REASON=$(printf 'x%.0s' $(seq 201))"
[ "$rc" -ne 0 ] && ok "a reason longer than 200 characters is refused" || ko "long reason (rc=$rc)"

# shellcheck disable=SC2046
run $(ed "$sha2") "EMERGENCY_REASON=$(printf 'a\nb')"
[ "$rc" -ne 0 ] && ok "a reason with a control character is refused" || ko "control char reason (rc=$rc)"

run EVENT_NAME=push PUSH_REF=refs/heads/main PUSH_SHA="$sha2"
[ "$(val emergency)" = false ] && ok "a push to main is never an emergency deploy" || ko "push emergency flag ($(val emergency))"

# shellcheck disable=SC2046
run $(wr "$sha2")
[ "$(val emergency)" = false ] && ok "a staging deploy is never an emergency deploy" || ko "staging emergency flag ($(val emergency))"

# shellcheck disable=SC2046
run $(wr "$sha2") 'RUN_SHA=abc; echo injected'
[ "$rc" -ne 0 ] && ok "a malformed SHA is refused" || ko "malformed sha (rc=$rc)"

finish
