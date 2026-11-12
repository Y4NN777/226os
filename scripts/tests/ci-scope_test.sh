#!/usr/bin/env bash
# Evidence for scripts/ci-scope.sh (ADR-0010): `docs` only when EVERY changed path is documentation that no gate reads;
# anything unknown, empty or mixed is `full`.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
sc="$here/../ci-scope.sh"
[ -x "$sc" ] || { ko "scripts/ci-scope.sh missing or not executable"; finish; exit 1; }

# scope WANT NAME PATH... : feed the paths on stdin, NUL-separated like `git diff --name-only --no-renames -z`, and compare
scope() {
  local want="$1" name="$2"
  shift 2
  local got
  got="$( { [ $# -eq 0 ] || printf '%s\0' "$@"; } | "$sc" 2>&1)" || got="ERROR($got)"
  if [ "$got" = "$want" ]; then ok "$name -> $want"; else ko "$name: wanted '$want', got '$got'"; fi
}

scope docs "README and a doc page" README.md docs/GIT_WORKFLOW.md
scope docs "ROADMAP, TODO, AI_USAGE, ADR" ROADMAP.md TODO.md AI_USAGE.md docs/decisions/0010-x.md
scope docs "licence files and brand images" LICENSE NOTICE docs/brand/226os-logo.png
scope docs "a design prototype page" prototypes/html/index.html
scope full "Go code" src/api/internal/app/app.go
scope full "one code file among docs" README.md src/web/src/main.tsx
scope full "the OpenAPI contract" api/openapi.yaml
scope full "a migration" db/migrations/00011_x.sql
scope full "a script" scripts/backup.sh
scope full "a Make file" mk/20-gates.mk
scope full "the workflow itself" .github/workflows/quality.yml
scope full "compose" compose.yaml
scope full "lockfiles" pnpm-lock.yaml
scope full "go.mod" go.mod
scope full "the coverage scope document is read by a gate" docs/quality/COVERAGE_SCOPE.md
scope full "secrets and their README are checked by a gate" secrets/README.md
scope full "an encrypted secrets file" secrets/staging-test.enc.env
scope full "a markdown file inside the code tree is not assumed to be docs" src/web/NOTES.md
scope full "an unknown top-level file" Dockerfile
scope full "a path that only looks like docs" docs/../scripts/x.sh
scope full "no changed path at all (unknown, not docs)"
scope full "an empty path entry" ""
scope full "a rename out of the code tree into docs lists both paths (git diff --no-renames)" src/api/x.go docs/x.md
scope full "a migration moved into docs" db/migrations/00011_x.sql docs/00011_x.md
scope docs "a non-ASCII documentation path" "docs/café.md"
scope docs "a path with a newline in its name" $'docs/a\nb.md'
scope full "a Go file under docs is still compiled by the gates" docs/example/main.go
scope full "a script under docs" docs/tools/run.sh
scope full "a Go test file under prototypes" prototypes/x_test.go
scope docs "an image and a text file under docs" docs/brand/a.png docs/notes.txt

rc=0
out="$(printf 'README.md\0' | "$sc" --bogus 2>&1)" || rc=$?
[ "$rc" -ne 0 ] && ok "an unknown option is refused" || ko "unknown option accepted: $out"

finish
