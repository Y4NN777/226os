#!/usr/bin/env bash
# Tests for scripts/check-docs.py against scratch trees: a clean tree passes; a broken link, a private path,
# a wrong operation count, an unlisted ADR, a missing count statement and an empty tree each fail closed.
# Usage: scripts/tests/check-docs_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
checker="$root/scripts/check-docs.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# fresh DIR: a clean scratch tree with one linked page, two operations and one listed ADR.
fresh() {
  rm -rf "$tmp/t"
  mkdir -p "$tmp/t/docs/decisions" "$tmp/t/docs/quality" "$tmp/t/api"
  printf '# Home\n\nSee [guide](docs/guide.md) and [web](https://example.org).\n' >"$tmp/t/README.md"
  printf '# Guide\n' >"$tmp/t/docs/guide.md"
  printf 'paths:\n  /a:\n    get:\n      operationId: a\n  /b:\n    get:\n      operationId: b\n' >"$tmp/t/api/openapi.yaml"
  printf 'It proves all 2 operations are routed.\n' >"$tmp/t/docs/quality/COVERAGE_SCOPE.md"
  printf '# Decisions\n\n| 0001 | 0001-first.md |\n' >"$tmp/t/docs/decisions/README.md"
  printf '# First\n' >"$tmp/t/docs/decisions/0001-first.md"
}

# case NAME EXPECT(pass|fail)
run() {
  local name="$1" expect="$2" rc=0 out
  out="$(python3 -I "$checker" --root "$tmp/t" 2>&1)" || rc=$?
  if [ "$expect" = pass ]; then
    [ "$rc" -eq 0 ] && ok "$name" || ko "$name (rc=$rc out=$out)"
  else
    [ "$rc" -ne 0 ] && ok "$name" || ko "$name (rc=$rc out=$out)"
  fi
}

fresh
run "a clean tree passes" pass

fresh
printf 'See [gone](docs/missing.md).\n' >>"$tmp/t/README.md"
run "a broken relative link fails" fail

fresh
printf 'Example `[x](docs/missing.md)` and a fenced block:\n```\n[y](nope.md)\n```\n' >>"$tmp/t/README.md"
run "links inside code are ignored" pass

fresh
printf 'Run it from /home/someone/work.\n' >>"$tmp/t/README.md"
run "an absolute home path fails" fail

fresh
printf 'scratch: /tmp/claude-1000/x\n' >>"$tmp/t/docs/guide.md"
run "an agent scratch path fails" fail

fresh
printf 'paths:\n  /a:\n    get:\n      operationId: a\n' >"$tmp/t/api/openapi.yaml"
run "a wrong operation count fails" fail

fresh
printf 'No count here.\n' >"$tmp/t/docs/quality/COVERAGE_SCOPE.md"
run "a missing operation count statement fails" fail

fresh
printf '# Second\n' >"$tmp/t/docs/decisions/0002-second.md"
run "an ADR missing from the index fails" fail

fresh
rm -f "$tmp/t/docs/decisions/README.md"
run "a missing ADR index fails" fail

rm -rf "$tmp/t"
mkdir -p "$tmp/t"
run "an empty tree is BLOCKED (fails closed)" fail

finish
