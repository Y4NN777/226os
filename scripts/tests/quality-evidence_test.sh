#!/usr/bin/env bash
# Evidence for scripts/check-quality-evidence.sh (ADR-0010): the fan-in check behind the parallel CI groups.
# It must fail unless every expected step has a PASS for the tested commit in some group summary.
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
chk="$here/../check-quality-evidence.sh"
[ -x "$chk" ] || { ko "scripts/check-quality-evidence.sh missing or not executable"; finish; exit 1; }

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
sha=1111111111111111111111111111111111111111
other=2222222222222222222222222222222222222222

# summary DIR LABEL SHA RESULT... : write a driver-style summary; each RESULT is "PASS step" / "FAIL step" / "NOT RUN step"
summary() {
  local dir="$1" label="$2" commit="$3"
  shift 3
  mkdir -p "$dir"
  {
    printf '%s summary for commit %s\n' "$label" "$commit"
    local r
    for r in "$@"; do
      case "$r" in
        "NOT RUN "*) printf '  %-8s %s (1s)\n' "NOT RUN" "${r#NOT RUN }" ;;
        *) printf '  %-8s %s (1s)\n' "${r%% *}" "${r#* }" ;;
      esac
    done
    if printf '%s\n' "$@" | grep -qE '^(FAIL|NOT RUN)'; then printf '%s: FAILED (1 step(s) failed)\n' "$label"; else printf '%s: PASS\n' "$label"; fi
  } >"$dir/ci-$label-summary.txt"
}
run() { rc=0; out="$("$chk" "$@" 2>&1)" || rc=$?; }

good="$tmp/good"
summary "$good/a" lint-go "$sha" "PASS doctor" "PASS lint-go"
summary "$good/b" unit "$sha" "PASS doctor" "PASS test-unit" "PASS coverage"
run "$good" "$sha" doctor lint-go test-unit coverage
if [ "$rc" -eq 0 ] && grep -q 'quality evidence: PASS' <<<"$out"; then ok "every expected step has a PASS for the commit"; else ko "all good (rc=$rc out=$out)"; fi
grep -q 'doctor' <<<"$out" && ok "the steps are listed" || ko "steps not listed: $out"

run "$good" "$sha" doctor lint-go test-unit coverage typecheck
if [ "$rc" -ne 0 ] && grep -q 'typecheck' <<<"$out" && grep -qi 'missing' <<<"$out"; then ok "an expected step with no PASS (typecheck) fails"; else ko "missing step (rc=$rc out=$out)"; fi

run "$good" "$other" doctor lint-go
if [ "$rc" -ne 0 ] && grep -qi 'commit' <<<"$out"; then ok "a summary for another commit fails"; else ko "commit mismatch (rc=$rc out=$out)"; fi

bad="$tmp/bad"
summary "$bad/a" lint-go "$sha" "PASS doctor" "FAIL lint-go"
summary "$bad/b" unit "$sha" "PASS doctor" "PASS test-unit"
run "$bad" "$sha" doctor lint-go test-unit
if [ "$rc" -ne 0 ] && grep -q 'FAIL' <<<"$out"; then ok "a FAIL line fails the check even if other groups passed"; else ko "FAIL line (rc=$rc out=$out)"; fi

nr="$tmp/notrun"
summary "$nr/a" lint-go "$sha" "PASS doctor" "NOT RUN lint-go"
run "$nr" "$sha" doctor lint-go
[ "$rc" -ne 0 ] && grep -q 'NOT RUN' <<<"$out" && ok "NOT RUN is not a pass" || ko "NOT RUN (rc=$rc out=$out)"

trunc="$tmp/trunc"
mkdir -p "$trunc"
printf 'ci-lint-go summary for commit %s\n  PASS     doctor (1s)\n  PASS     lint-go (1s)\n' "$sha" >"$trunc/ci-lint-go-summary.txt"
run "$trunc" "$sha" doctor lint-go
[ "$rc" -ne 0 ] && grep -qi 'final' <<<"$out" && ok "a summary without its final PASS line (cut off) fails" || ko "truncated (rc=$rc out=$out)"

mkdir -p "$tmp/empty"
run "$tmp/empty" "$sha" doctor
[ "$rc" -ne 0 ] && grep -qi 'no summary' <<<"$out" && ok "no summary at all fails" || ko "empty dir (rc=$rc out=$out)"
run "$tmp/does-not-exist" "$sha" doctor
[ "$rc" -ne 0 ] && ok "a missing evidence directory fails" || ko "missing dir"

run "$good" "$sha"
[ "$rc" -ne 0 ] && grep -qi 'usage' <<<"$out" && ok "no expected step is a usage error, not a pass" || ko "no steps (rc=$rc out=$out)"
run "$good"
[ "$rc" -ne 0 ] && ok "missing arguments are refused" || ko "missing args"

extra="$tmp/extra"
summary "$extra/a" lint-go "$sha" "PASS doctor" "PASS lint-go" "PASS something-else"
run "$extra" "$sha" doctor lint-go
[ "$rc" -eq 0 ] && ok "a step beyond the expected list does not fail the check" || ko "extra step (rc=$rc out=$out)"

# A step that FAILED in one group and PASSED in another group is still a failure (every summary must be clean).
mixed="$tmp/mixed"
summary "$mixed/a" lint-go "$sha" "PASS doctor" "FAIL lint-go"
summary "$mixed/b" lint-go2 "$sha" "PASS doctor" "PASS lint-go"
run "$mixed" "$sha" doctor lint-go
[ "$rc" -ne 0 ] && ok "a FAIL is not cancelled by a PASS of the same step elsewhere" || ko "mixed (rc=$rc out=$out)"

finish
