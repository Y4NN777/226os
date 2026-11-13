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

# ---------- AI_USAGE.md record dates (report mode: a WARN, never a failure; --strict-dates makes it one) ----------
# dated_tree HEAD_DATE RECORD_DATE: a git repository whose HEAD commit is dated HEAD_DATE, with one AI_USAGE record.
dated_tree() {
  fresh
  printf '# AI use\n\n```yaml\ndate: "%s"\ncontributor: "x"\n```\n' "$2" >"$tmp/t/AI_USAGE.md"
  (
    cd "$tmp/t"
    git init -q -b main
    git config user.email t@example.invalid
    git config user.name t
    git add -A
    GIT_AUTHOR_DATE="$1T12:00:00Z" GIT_COMMITTER_DATE="$1T12:00:00Z" git -c commit.gpgsign=false commit -q -m "tree"
  )
}
# dates_run [ARGS...]: runs the checker on the tree; sets rc and out.
dates_run() {
  rc=0
  out="$(python3 -I "$checker" --root "$tmp/t" "$@" 2>&1)" || rc=$?
}

dated_tree 2026-11-13 2026-11-13
dates_run
[ "$rc" -eq 0 ] && ! grep -q 'WARN' <<<"$out" && ok "a record dated the HEAD commit day passes without a warning" || ko "same day (rc=$rc out=$out)"

dated_tree 2026-11-13 2026-11-14
dates_run
[ "$rc" -eq 0 ] && ! grep -q 'WARN' <<<"$out" && ok "a record one day ahead is tolerated (timezones, rebases)" || ko "one day (rc=$rc out=$out)"

dated_tree 2026-11-13 2026-11-16
dates_run
[ "$rc" -eq 0 ] && grep -q 'WARN: ai-usage: AI_USAGE.md:4 date 2026-11-16 is after the HEAD commit date 2026-11-13' <<<"$out" && ok "a record dated days after HEAD is reported as a warning (report mode)" || ko "future (rc=$rc out=$out)"

dates_run --strict-dates
[ "$rc" -eq 1 ] && ok "--strict-dates turns that warning into a failure (the promotion switch)" || ko "strict (rc=$rc out=$out)"

dated_tree 2026-11-13 2026-11-10
dates_run --strict-dates
[ "$rc" -eq 0 ] && ok "an older record is fine even in strict mode" || ko "older strict (rc=$rc out=$out)"

fresh
printf '# AI use\n\n```yaml\ndate: "2030-01-01"\n```\n' >"$tmp/t/AI_USAGE.md"
dates_run --strict-dates
[ "$rc" -eq 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "outside a git repository the date check is NOT RUN, not a failure" || ko "no git (rc=$rc out=$out)"

finish
