#!/usr/bin/env bash
# The tracker check fails when a slice file changes and TODO.md does not, and passes
# when TODO.md changes with it, when the diff is not a slice, and for an untracked migration.
# Usage: scripts/tests/check-todo-updated_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
checker="$root/scripts/check-todo-updated.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
repo="$tmp/repo"

gitc() {
  git -C "$repo" -c user.name=todo-check -c user.email=todo-check@example.com "$@"
}

seed() {
  rm -rf "$repo"
  mkdir -p "$repo/src" "$repo/docs"
  gitc init -q
  printf '# tracker\n- [ ] item\n' >"$repo/TODO.md"
  printf 'package p\n' >"$repo/src/main.go"
  printf '# note\n' >"$repo/docs/note.md"
  gitc add TODO.md src/main.go docs/note.md
  gitc commit -q -m base
  gitc rev-parse HEAD
}

run() {
  local name="$1" expect="$2" rc=0 out base
  base="$(gitc rev-parse HEAD)"
  out="$(TODO_BASE_REF="$base" "$checker" "$repo" 2>&1)" || rc=$?
  if [ "$expect" = pass ]; then
    [ "$rc" -eq 0 ] && ok "$name" || ko "$name (rc=$rc out=$out)"
  else
    [ "$rc" -ne 0 ] && ok "$name" || ko "$name (rc=$rc out=$out)"
  fi
}

seed >/dev/null
printf 'package p\nfunc F() {}\n' >"$repo/src/main.go"
run "a slice file without TODO.md fails" fail

printf '# tracker\n- [x] item\n' >"$repo/TODO.md"
run "TODO.md moving with the slice passes" pass

gitc checkout -q -- TODO.md src/main.go
printf '# edited\n' >"$repo/docs/note.md"
run "a docs-only change passes" pass

gitc checkout -q -- docs/note.md
mkdir -p "$repo/db/migrations"
printf '%s\n' '-- new' >"$repo/db/migrations/00011_new.sql"
run "an untracked migration without TODO.md fails" fail

printf '# tracker\n- [x] item\n' >"$repo/TODO.md"
run "an untracked migration with TODO.md passes" pass

gitc checkout -q -- TODO.md
rm -rf "$repo/db"
mkdir -p "$repo/.claude/hooks"
printf '# hook\n' >"$repo/.claude/hooks/domain_skills.py"
run "a harness-only change passes" pass

finish
