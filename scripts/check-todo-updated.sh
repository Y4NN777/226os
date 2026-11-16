#!/usr/bin/env bash
# The live tracker moves with the slice (TODO.md header, docs/team/KICKOFF.md).
#
# When the diff against the base changes a slice file, TODO.md must change in
# that same diff. The check does not read checkboxes and does not require every
# box to be ticked. A docs-only or harness-only diff passes.
#
# Slice files: src/, db/, api/, scripts/, mk/ and Makefile.
# Untracked slice files count. The working tree counts, not only commits.
#
# Usage: scripts/check-todo-updated.sh [repo-dir]
# Env:   TODO_BASE_REF  git ref to diff against (default: origin/develop, else origin/main)
#
# Exit 0 when the rule holds, 1 otherwise. Read-only.
set -euo pipefail

root="${1:-}"
if [ -z "$root" ]; then
  root="$(git rev-parse --show-toplevel)"
fi
root="$(cd "$root" && pwd)"

die() { printf 'check-todo-updated: %s\n' "$*" >&2; exit 1; }

git -C "$root" rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "not a git repository: $root"

base="${TODO_BASE_REF:-}"
if [ -z "$base" ]; then
  for c in origin/develop origin/main; do
    if git -C "$root" rev-parse --verify --quiet "${c}^{commit}" >/dev/null; then
      base="$c"
      break
    fi
  done
fi
[ -n "$base" ] || die "no TODO_BASE_REF and no origin/develop or origin/main; cannot tell whether the tracker moved"
git -C "$root" rev-parse --verify --quiet "${base}^{commit}" >/dev/null || die "base '${base}' does not resolve to a commit"

is_slice() {
  case "$1" in
    src/* | db/* | api/* | scripts/* | mk/* | Makefile) return 0 ;;
    *) return 1 ;;
  esac
}

# Compare with the merge base, not the tip of the base: on a branch behind it, a two-dot diff would also see the
# changes (including TODO.md) that landed on the base after the branch started.
mergebase="$(git -C "$root" merge-base "$base" HEAD)" || die "no merge base between '${base}' and HEAD"

paths=()
while IFS= read -r path; do
  [ -n "$path" ] || continue
  paths+=("$path")
done < <(git -C "$root" diff --name-only "$mergebase"; git -C "$root" ls-files --others --exclude-standard)

todo=0
slice=()
for path in "${paths[@]}"; do
  [ "$path" = "TODO.md" ] && todo=1
  if is_slice "$path"; then
    slice+=("$path")
  fi
done

if [ "${#slice[@]}" -eq 0 ]; then
  printf 'check-todo-updated: OK (no slice files against %s)\n' "$base"
  exit 0
fi
if [ "$todo" -eq 1 ]; then
  printf 'check-todo-updated: OK (TODO.md moves with %d slice file(s) against %s)\n' "${#slice[@]}" "$base"
  exit 0
fi

printf 'check-todo-updated: slice files changed against %s but TODO.md did not. Update the tracker in the same change. Boxes that are not done stay open.\n' "$base" >&2
for path in "${slice[@]}"; do
  printf '  %s\n' "$path" >&2
done
exit 1
