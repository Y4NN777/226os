#!/usr/bin/env bash
# Guard: a pull request into develop is refused while develop lacks commits that main has.
#
# Why: every promotion (develop -> main) creates a merge commit that exists only on main. If it is not
# merged back, GitHub shows develop as "N commits behind main" although no file differs, which reads
# as debt and confuses reviewers. docs/GIT_WORKFLOW.md makes the back-merge (main -> develop, branch
# chore/platform/sync-main-into-develop, Create a merge commit) a step after each promotion; this check
# makes it impossible to forget: other work waits until the back-merge has landed.
#
# Rules (pull requests only; the workflow passes the context as environment variables):
#   - BASE_REF is not develop: not concerned, pass.
#   - HEAD_REF is the back-merge branch: pass only if its head contains main (so it can fix the gap).
#   - any other work branch: pass only if develop contains main.
# Fails closed: a missing variable, a GitHub API error or a non-numeric answer is a failure.
#
# Environment: GITHUB_REPOSITORY (owner/name), GH_TOKEN, BASE_REF, HEAD_REF, HEAD_SHA.
set -euo pipefail

SYNC_BRANCH="chore/platform/sync-main-into-develop"

need() {
  [ -n "${!1:-}" ] || { printf 'check-develop-contains-main: BLOCKED: %s is not set\n' "$1" >&2; exit 1; }
}
need GITHUB_REPOSITORY
need BASE_REF
need HEAD_REF
need HEAD_SHA

# count LABEL ENDPOINT JQ: one non-negative integer from the GitHub compare API, or fail closed.
count() {
  local label="$1" endpoint="$2" field="$3" value
  if ! value="$(gh api "$endpoint" --jq "$field" 2>/dev/null)"; then
    printf 'check-develop-contains-main: could not read %s from the GitHub API\n' "$label" >&2
    exit 1
  fi
  case "$value" in
    '' | *[!0-9]*)
      printf 'check-develop-contains-main: could not read %s: unexpected answer\n' "$label" >&2
      exit 1
      ;;
  esac
  printf '%s' "$value"
}

[ "$BASE_REF" = develop ] || { echo "check-develop-contains-main: not concerned (base is $BASE_REF)"; exit 0; }

if [ "$HEAD_REF" = "$SYNC_BRANCH" ]; then
  behind="$(count "how far the back-merge head is behind main" "repos/$GITHUB_REPOSITORY/compare/main...$HEAD_SHA" .behind_by)"
  if [ "$behind" -ne 0 ]; then
    printf 'check-develop-contains-main: FAIL: the back-merge branch does not contain main (%s commit(s) missing): merge origin/main into it\n' "$behind" >&2
    exit 1
  fi
  echo "check-develop-contains-main: PASS (the back-merge branch contains main)"
  exit 0
fi

lacking="$(count "how far develop is behind main" "repos/$GITHUB_REPOSITORY/compare/develop...main" .ahead_by)"
if [ "$lacking" -ne 0 ]; then
  printf 'check-develop-contains-main: FAIL: develop is %s commit(s) behind main (promotion merge commits, no file difference).\n' "$lacking" >&2
  printf '  Open the back-merge first: branch %s from develop, merge origin/main with a merge commit, PR into develop.\n' "$SYNC_BRANCH" >&2
  printf '  See docs/GIT_WORKFLOW.md, promotion flow step 7. Other pull requests resume once it has landed.\n' >&2
  exit 1
fi
echo "check-develop-contains-main: PASS (develop contains main)"
