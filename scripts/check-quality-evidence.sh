#!/usr/bin/env bash
# Fan-in check for the parallel CI groups (ADR-0010). Every group writes a driver summary
# (reports/ci-<group>-summary.txt, scripts/quality.sh); this script proves the whole of `make quality` ran:
#   - at least one summary exists;
#   - every summary is for the tested commit, has no FAIL or NOT RUN line, and ends with its own PASS line;
#   - every expected step has a PASS line in some summary.
# A PASS in one group never cancels a FAIL in another. A step beyond the expected list is allowed.
#
# Usage: scripts/check-quality-evidence.sh <evidence-dir> <commit-sha> <expected-step>...
set -euo pipefail

if [ $# -lt 3 ]; then
  printf 'usage: %s <evidence-dir> <commit-sha> <expected-step>...\n' "$0" >&2
  exit 64
fi
dir="$1"
sha="$2"
shift 2
expected=("$@")

fails=0
bad() { printf 'quality evidence: %s\n' "$*" >&2; fails=$((fails + 1)); }

[ -d "$dir" ] || { printf 'quality evidence: FAIL (no such directory: %s)\n' "$dir" >&2; exit 1; }

mapfile -t files < <(find "$dir" -type f -name 'ci-*-summary.txt' | sort)
[ "${#files[@]}" -gt 0 ] || { printf 'quality evidence: FAIL (no summary found under %s: nothing was proved)\n' "$dir" >&2; exit 1; }

passed=" "
for f in "${files[@]}"; do
  name="${f#"$dir"/}"
  head="$(head -n 1 "$f")"
  label="${head%% summary for commit *}"
  commit="${head##* summary for commit }"
  if [ "$commit" != "$sha" ]; then
    bad "$name is for commit '${commit}', not the tested commit $sha"
  fi
  while IFS= read -r line; do
    case "$line" in
      '  FAIL '*) bad "$name: ${line#  }" ;;
      '  NOT RUN '*) bad "$name: ${line#  } (NOT RUN is not a pass)" ;;
      '  PASS '*)
        step="${line#  PASS}"
        step="${step#"${step%%[![:space:]]*}"}"
        step="${step%% *}"
        passed+="$step "
        ;;
    esac
  done <"$f"
  last="$(tail -n 1 "$f")"
  [ "$last" = "$label: PASS" ] || bad "$name has no final '$label: PASS' line (cut off or failed): '$last'"
done

for s in "${expected[@]}"; do
  case "$passed" in
    *" $s "*) ;;
    *) bad "step '$s' is missing: no PASS for it in any summary" ;;
  esac
done

if [ "$fails" -gt 0 ]; then
  printf 'quality evidence: FAIL (%d problem(s)) for commit %s\n' "$fails" "$sha" >&2
  exit 1
fi
printf 'quality evidence: PASS for commit %s (%d summaries, steps: %s)\n' "$sha" "${#files[@]}" "${expected[*]}"
