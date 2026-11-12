#!/usr/bin/env bash
# Evidence for scripts/worktree.sh: new/done on a real throwaway git repo, never forcing anything.
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
wt="$here/../worktree.sh"
[ -x "$wt" ] || { ko "scripts/worktree.sh missing or not executable"; finish; exit 1; }

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
repo="$tmp/repo"
git init -q -b develop "$repo"
git -C "$repo" config user.name t
git -C "$repo" config user.email t@example.test
git -C "$repo" -c commit.gpgsign=false commit -q --allow-empty -m init
run() { rc=0; out="$(cd "$repo" && "$wt" "$@" 2>&1)" || rc=$?; }

run new feature/ops/demo
dir="$repo/.claude/worktrees/feature-ops-demo"
if [ "$rc" -eq 0 ] && [ -d "$dir" ] && [ "$(git -C "$dir" branch --show-current)" = feature/ops/demo ]; then ok "new creates the worktree on a new branch from develop"; else ko "new (rc=$rc out=$out)"; fi
grep -q "$dir" <<<"$out" && ok "new prints the path" || ko "path not printed"

run new 'bad name'
[ "$rc" -ne 0 ] && ok "a branch name outside the policy is refused" || ko "bad name accepted"

run new feature/ops/demo
[ "$rc" -ne 0 ] && ok "an existing worktree is refused" || ko "duplicate accepted"

printf x >"$dir/dirty.txt"
run "done" feature/ops/demo
if [ "$rc" -ne 0 ] && [ -d "$dir" ] && grep -qi 'uncommitted' <<<"$out"; then ok "done refuses a dirty worktree and keeps it"; else ko "dirty (rc=$rc out=$out)"; fi
rm "$dir/dirty.txt"

git -C "$dir" -c commit.gpgsign=false commit -q --allow-empty -m work
run "done" feature/ops/demo
if [ "$rc" -ne 0 ] && [ -d "$dir" ] && grep -qi 'not merged' <<<"$out"; then ok "done refuses a branch that is not merged into develop"; else ko "unmerged (rc=$rc out=$out)"; fi

git -C "$repo" -c commit.gpgsign=false merge -q --no-ff -m merge feature/ops/demo
run "done" feature/ops/demo
if [ "$rc" -eq 0 ] && [ ! -d "$dir" ] && ! git -C "$repo" rev-parse -q --verify refs/heads/feature/ops/demo >/dev/null; then ok "done removes the merged worktree and its branch"; else ko "done (rc=$rc out=$out)"; fi

if grep -vE '^[0-9]*:\s*#' <(grep -nE -- '--force|(^| )-f( |$)|branch -D' "$wt") | grep -q .; then ko "worktree.sh must never force"; else ok "worktree.sh never forces"; fi
finish
