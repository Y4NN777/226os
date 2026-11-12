#!/usr/bin/env bash
# Reproducible sqlc generation and drift check (ADR-0005, ADR-0006 `generate`).
# Same semantics as scripts/check-generated.sh: regenerate, then fail when a generated
# file differs from the index or is missing/untracked. In CI the index equals HEAD, so
# this is the "generate && git diff --exit-code" gate; locally `git add` generated files first.
# sqlc is pinned through `go tool` (go.mod `tool` line). Exit 0 = in sync.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

config=db/sqlc.yaml
out=src/api/internal/platform/dbgen

go tool sqlc generate -f "$config"

fail=0
if [ -z "$(git ls-files -- "$out")" ]; then
  printf 'check-sqlc: no generated file is tracked under %s\n' "$out" >&2
  fail=1
fi
if ! git diff --exit-code --stat -- "$out"; then
  printf 'check-sqlc: generated files differ from the index; commit the regenerated output\n' >&2
  fail=1
fi
untracked="$(git ls-files --others --exclude-standard -- "$out")"
if [ -n "$untracked" ]; then
  printf 'check-sqlc: untracked generated files:\n%s\n' "$untracked" >&2
  fail=1
fi
# A tracked file that sqlc no longer generates would linger silently: it shows as deleted
# only if sqlc removed it, so also list tracked files that are missing on disk.
missing="$(git ls-files --deleted -- "$out")"
if [ -n "$missing" ]; then
  printf 'check-sqlc: tracked generated files missing after generation:\n%s\n' "$missing" >&2
  fail=1
fi

[ "$fail" -eq 0 ] && printf 'check-sqlc: OK (generated files in sync with %s)\n' "$config"
exit "$fail"
