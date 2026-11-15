#!/usr/bin/env bash
# Deploy planner (docs/SHIP.md section 2): decides from the triggering event WHAT the deploy workflow does, before
# any image is built or any secret is read. Fail closed: an unexpected event, ref or SHA is an error, not a skip.
# A skip (proceed=false) is only ever the answer to a well-formed event that must not deploy.
#
#   workflow_run (Quality completed)   -> staging, only for a successful push run of this repository on develop,
#                                         and only for the develop tip (an older commit is superseded)
#   push to main                       -> production, for the tree of that commit
#
# Environment (set by the workflow through `env:`, never inlined in a script):
#   EVENT_NAME, THIS_REPO; workflow_run: RUN_CONCLUSION, RUN_EVENT, RUN_BRANCH, RUN_REPO, RUN_SHA;
#   push: PUSH_REF, PUSH_SHA. Needs a checkout with origin/develop fetched (staging only).
# Output (GITHUB_OUTPUT, or stdout when unset): proceed, reason, target, profile, sha, tree.
set -euo pipefail

die() { printf 'deploy-plan: REFUSED: %s\n' "$*" >&2; exit 2; }
out="${GITHUB_OUTPUT:-/dev/stdout}"
emit() { printf '%s=%s\n' "$1" "$2" >>"$out"; }
skip() {
  printf 'deploy-plan: nothing to deploy: %s\n' "$1" >&2
  emit proceed false
  emit reason "$1"
  exit 0
}
sha_re='^[0-9a-f]{40}$'

case "${EVENT_NAME-}" in
  workflow_run)
    [ -n "${THIS_REPO-}" ] || die "THIS_REPO is not set"
    [ "${RUN_CONCLUSION-}" = success ] || skip "the Quality run did not succeed (${RUN_CONCLUSION:-unknown})"
    [ "${RUN_EVENT-}" = push ] || skip "the Quality run was not a push"
    [ "${RUN_BRANCH-}" = develop ] || skip "the Quality run was not on develop"
    [ "${RUN_REPO-}" = "$THIS_REPO" ] || skip "the Quality run comes from another repository"
    sha="${RUN_SHA-}"
    [[ "$sha" =~ $sha_re ]] || die "RUN_SHA is not a 40-character commit SHA"
    git cat-file -e "$sha^{commit}" 2>/dev/null || die "commit $sha is not in this checkout"
    tip="$(git rev-parse --verify "refs/remotes/origin/develop^{commit}" 2>/dev/null)" || die "origin/develop is not fetched"
    git merge-base --is-ancestor "$sha" "$tip" || die "commit $sha is not on develop"
    [ "$sha" = "$tip" ] || skip "superseded: $sha is behind the develop tip $tip, which has its own deploy"
    emit target staging
    emit profile staging-test
    ;;
  push)
    [ "${PUSH_REF-}" = refs/heads/main ] || die "a push deploys only from refs/heads/main (got '${PUSH_REF-}')"
    sha="${PUSH_SHA-}"
    [[ "$sha" =~ $sha_re ]] || die "PUSH_SHA is not a 40-character commit SHA"
    git cat-file -e "$sha^{commit}" 2>/dev/null || die "commit $sha is not in this checkout"
    emit target production
    emit profile production-live
    ;;
  *) die "unexpected event '${EVENT_NAME-}' (workflow_run or push expected)" ;;
esac

emit proceed true
emit sha "$sha"
emit tree "$(git rev-parse "$sha^{tree}")"
printf 'deploy-plan: %s for %s\n' "${EVENT_NAME}" "$sha" >&2
