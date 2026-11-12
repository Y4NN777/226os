#!/usr/bin/env bash
# Short-lived worktrees for one branch each, never forced (no --force, no branch -D), so it works under
# deny rules on forced removal and cannot lose work.
#   scripts/worktree.sh new <branch>    create .claude/worktrees/<branch with / as ->, branch cut from develop
#   scripts/worktree.sh done <branch>   remove it and delete the branch, only if clean and merged into develop
# Run from the main checkout. .claude/worktrees/ is git-ignored.
set -euo pipefail

cmd="${1-}"
branch="${2-}"
die() { printf 'REFUSED: %s\n' "$*" >&2; exit 2; }
[ -n "$branch" ] || die "usage: worktree.sh new|done <type/scope/slug>"
[[ "$branch" =~ ^(feature|fix|hotfix|refactor|test|docs|chore)/[a-z0-9-]+/[a-z0-9-]+$ ]] || die "branch '$branch' does not match type/scope/slug (docs/GIT_WORKFLOW.md)"

root="$(git rev-parse --show-toplevel)"
dir="$root/.claude/worktrees/${branch//\//-}"

case "$cmd" in
  new)
    [ ! -e "$dir" ] || die "$dir already exists"
    git rev-parse -q --verify refs/heads/develop >/dev/null || die "no local develop branch"
    git worktree add -q -b "$branch" "$dir" develop
    printf 'worktree: %s\n' "$dir"
    ;;
  done)
    [ -d "$dir" ] || die "no worktree at $dir"
    [ -z "$(git -C "$dir" status --porcelain)" ] || die "$dir has uncommitted changes; commit or discard them yourself first"
    git merge-base --is-ancestor "$branch" develop || die "branch '$branch' is not merged into develop (a merge commit on develop counts)"
    git worktree remove "$dir"
    git branch -d "$branch"
    printf 'removed: %s and branch %s\n' "$dir" "$branch"
    ;;
  *) die "unknown command '$cmd' (new | done)" ;;
esac
