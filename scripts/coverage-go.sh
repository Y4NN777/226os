#!/usr/bin/env bash
# Go first-party statement coverage gate (G2): exactly 100.00%.
# Exclusions come only from the `coverage-exclude-go` block of
# docs/quality/COVERAGE_SCOPE.md; the script fails closed if that file is absent.
# Usage: scripts/coverage-go.sh [output-dir]
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
scope=docs/quality/COVERAGE_SCOPE.md
out="${1:-coverage}"

if [[ ! -f "$scope" ]]; then
  echo "BLOCKED: $scope is absent" >&2
  exit 1
fi

export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.27.4}"
mkdir -p "$out"

# Exclusion patterns: lines between the ```coverage-exclude-go fence markers.
mapfile -t excludes < <(awk '/^```coverage-exclude-go$/{f=1;next} /^```$/{f=0} f && NF && $1 !~ /^#/{print $1}' "$scope")

# Every exclusion must be a full repo-relative first-party path to an existing file
# (a short entry such as main.go would otherwise match every file with that name).
for p in "${excludes[@]}"; do
  if [[ "$p" != src/* || ! -f "$p" ]]; then
    echo "BLOCKED: invalid coverage exclusion '$p' (must be an existing file under src/)" >&2
    exit 1
  fi
done

go test -race -count=1 -covermode=atomic -coverprofile="$out/go.raw.out" ./...

filter="$out/go.out"
head -n1 "$out/go.raw.out" >"$filter"
tail -n +2 "$out/go.raw.out" | while IFS= read -r line; do
  skip=0
  for p in "${excludes[@]}"; do
    [[ "$line" == *"/$p:"* ]] && skip=1 && break
  done
  ((skip)) || printf '%s\n' "$line"
done >>"$filter"

echo "excluded paths: ${excludes[*]:-none}"
total="$(go tool cover -func="$filter" | awk '/^total:/{print $3}')"
# Exact gate: count uncovered statements in the profile, never a rounded percentage.
read -r stmts uncovered < <(awk 'NR>1{n+=$2; if($3==0)u+=$2} END{print n+0, u+0}' "$filter")
echo "go first-party statement coverage: $total ($((stmts - uncovered))/$stmts statements)"
if ((stmts == 0 || uncovered > 0)); then
  echo "FAIL: Go coverage must be exactly 100.00%; $uncovered of $stmts statements uncovered" >&2
  go tool cover -func="$filter" | awk '$3 != "100.0%"' >&2
  exit 1
fi
echo "PASS"
