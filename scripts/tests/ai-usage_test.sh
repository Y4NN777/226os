#!/usr/bin/env bash
# Tests for scripts/ai-usage.py: one AI-use record per file in docs/ai-usage/, AI_USAGE.md generated from them between two
# markers. Build is deterministic and idempotent, never touches what is outside the markers, refuses an invalid record, and
# check fails when AI_USAGE.md is stale. This removes the append-at-the-end conflict every pull request used to have.
# Usage: scripts/tests/ai-usage_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/ai-usage.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
START='<!-- ai-usage:generated:start'
END='<!-- ai-usage:generated:end -->'

# record NAME TITLE DATE: writes docs/ai-usage/NAME.md with a valid record.
record() {
  mkdir -p "$tmp/t/docs/ai-usage"
  printf '## %s\n\n```yaml\ndate: "%s"\ncontributor: "someone"\ntool_or_model: "x"\n```\n' "$2" "$3" >"$tmp/t/docs/ai-usage/$1.md"
}
fresh() {
  rm -rf "$tmp/t"
  mkdir -p "$tmp/t"
  printf '# AI use\n\n## Policy\n\nOld records live here, untouched.\n\n## Old record\n\n```yaml\ndate: "2026-11-01"\n```\n' >"$tmp/t/AI_USAGE.md"
}
run() { rc=0; out="$(python3 -I "$tool" "$@" --root "$tmp/t" 2>&1)" || rc=$?; }

fresh
record 2026-11-13-beta "Beta record" 2026-11-13
record 2026-11-12-alpha "Alpha record" 2026-11-12
run build
[ "$rc" -eq 0 ] && grep -q "$START" "$tmp/t/AI_USAGE.md" && grep -q "$END" "$tmp/t/AI_USAGE.md" && ok "build adds the generated block between two markers" || ko "markers (rc=$rc out=$out)"
a="$(grep -n '## Alpha record' "$tmp/t/AI_USAGE.md" | cut -d: -f1)"; b="$(grep -n '## Beta record' "$tmp/t/AI_USAGE.md" | cut -d: -f1)"
[ -n "$a" ] && [ -n "$b" ] && [ "$a" -lt "$b" ] && ok "records are ordered by file name (date first), not by the order they were written" || ko "order (a=$a b=$b)"

cp "$tmp/t/AI_USAGE.md" "$tmp/first.md"
run build
cmp -s "$tmp/first.md" "$tmp/t/AI_USAGE.md" && ok "build is idempotent (a second run changes nothing)" || ko "idempotent"

head -n "$(($(grep -n "$START" "$tmp/t/AI_USAGE.md" | cut -d: -f1) - 1))" "$tmp/t/AI_USAGE.md" >"$tmp/before.md"
printf '# AI use\n\n## Policy\n\nOld records live here, untouched.\n\n## Old record\n\n```yaml\ndate: "2026-11-01"\n```\n' >"$tmp/orig.md"
grep -q 'Old records live here, untouched.' "$tmp/before.md" && grep -q '## Old record' "$tmp/before.md" && ok "what is outside the markers is kept" || ko "outside kept"

run check
[ "$rc" -eq 0 ] && ok "check passes when AI_USAGE.md is up to date" || ko "check fresh (rc=$rc out=$out)"

record 2026-11-13-gamma "Gamma record" 2026-11-13
run check
[ "$rc" -eq 1 ] && grep -q 'stale' <<<"$out" && ok "check fails when a record file was added and the file was not rebuilt" || ko "check stale (rc=$rc out=$out)"
run build
run check
[ "$rc" -eq 0 ] && grep -q '## Gamma record' "$tmp/t/AI_USAGE.md" && ok "build brings it up to date" || ko "rebuild (rc=$rc out=$out)"

python3 -I -c 'import sys; p=sys.argv[1]; s=open(p).read(); open(p,"w").write(s.replace("## Beta record","## Beta record EDITED BY HAND"))' "$tmp/t/AI_USAGE.md"
run check
[ "$rc" -eq 1 ] && ok "a hand edit inside the generated block is detected" || ko "hand edit (rc=$rc out=$out)"
run build
! grep -q 'EDITED BY HAND' "$tmp/t/AI_USAGE.md" && ok "build overwrites a hand edit inside the block" || ko "overwrite"

printf '# Notes\n\nnot a record\n' >"$tmp/t/docs/ai-usage/README.md"
printf 'x' >"$tmp/t/docs/ai-usage/notes.txt"
printf '## Bad name\n\n```yaml\ndate: "2026-11-13"\ncontributor: "x"\n```\n' >"$tmp/t/docs/ai-usage/draft.md"
run build
[ "$rc" -eq 0 ] && ! grep -q 'Bad name' "$tmp/t/AI_USAGE.md" && ok "files that are not named YYYY-MM-DD-slug.md are not records (README, notes, drafts)" || ko "ignored files (rc=$rc out=$out)"

fresh
printf 'no heading here\n```yaml\ndate: "2026-11-13"\ncontributor: "x"\n```\n' >"$tmp/t/bad.md"
mkdir -p "$tmp/t/docs/ai-usage" && cp "$tmp/t/bad.md" "$tmp/t/docs/ai-usage/2026-11-13-bad.md"
run build
[ "$rc" -eq 2 ] && grep -q '2026-11-13-bad.md' <<<"$out" && ok "a record without a ## heading is refused, naming the file (fails closed)" || ko "no heading (rc=$rc out=$out)"

fresh
mkdir -p "$tmp/t/docs/ai-usage"
printf '## No date\n\n```yaml\ncontributor: "x"\n```\n' >"$tmp/t/docs/ai-usage/2026-11-13-nodate.md"
run build
[ "$rc" -eq 2 ] && grep -q 'date' <<<"$out" && ok "a record without a date is refused" || ko "no date (rc=$rc out=$out)"

fresh
mkdir -p "$tmp/t/docs/ai-usage"
printf '## No contributor\n\n```yaml\ndate: "2026-11-13"\n```\n' >"$tmp/t/docs/ai-usage/2026-11-13-nocontrib.md"
run build
[ "$rc" -eq 2 ] && grep -q 'contributor' <<<"$out" && ok "a record without a contributor is refused" || ko "no contributor (rc=$rc out=$out)"

fresh
run build
[ "$rc" -eq 0 ] && grep -q "$START" "$tmp/t/AI_USAGE.md" && ok "no record files: build still adds an empty block" || ko "empty (rc=$rc out=$out)"
run check
[ "$rc" -eq 0 ] && ok "and check passes on it" || ko "empty check (rc=$rc out=$out)"

rm -rf "$tmp/t"
mkdir -p "$tmp/t"
run build
[ "$rc" -eq 2 ] && ok "no AI_USAGE.md at all is a usage error, not a silent success" || ko "no file (rc=$rc out=$out)"

finish
