#!/usr/bin/env bash
# Sequential driver behind `make quality` / `make quality-full` (ADR-0006).
# Runs every step through make, records each exit code, prints a final summary and exits
# non-zero if ANY step failed. A failure is never masked: the default is to keep going so
# the summary lists every failing gate; QUALITY_FAIL_FAST=1 stops at the first failure and
# reports the remaining steps as NOT RUN. A run with no steps is a failure (nothing proved).
#
# Usage: scripts/quality.sh <make-command> <label> <step>...
# Env:   QUALITY_FAIL_FAST=1   stop at the first failing step
#        REPORTS_DIR           where logs and the summary go (default: reports)
# Evidence: $REPORTS_DIR/logs/<step>.log per step and $REPORTS_DIR/<label>-summary.txt.
set -euo pipefail

[ $# -ge 2 ] || { printf 'usage: %s <make-command> <label> <step>...\n' "$0" >&2; exit 64; }
make_cmd="$1"
label="$2"
shift 2
steps=("$@")

reports="${REPORTS_DIR:-reports}"
mkdir -p "$reports/logs"
summary="$reports/$label-summary.txt"

if [ "${#steps[@]}" -eq 0 ]; then
  printf '%s: FAILED (no steps configured; nothing was run, which is not a PASS)\n' "$label" >&2
  exit 1
fi

sha="$(git rev-parse HEAD 2>/dev/null)" || sha=unknown
results=()
failed=0
stopped=0
for step in "${steps[@]}"; do
  if [ "$stopped" -eq 1 ]; then
    results+=("NOT RUN  $step")
    continue
  fi
  printf '\n==== %s: step %s ====\n' "$label" "$step"
  start="$(date +%s)"
  rc=0
  # pipefail makes the `if` reflect the make exit status, not tee's.
  if "$make_cmd" --no-print-directory "$step" 2>&1 | tee "$reports/logs/$step.log"; then
    rc=0
  else
    rc=1
  fi
  secs=$(($(date +%s) - start))
  if [ "$rc" -eq 0 ]; then
    results+=("PASS     $step (${secs}s)")
  else
    results+=("FAIL     $step (${secs}s)")
    failed=$((failed + 1))
    [ "${QUALITY_FAIL_FAST:-0}" = 1 ] && stopped=1
  fi
done

{
  printf '%s summary for commit %s\n' "$label" "$sha"
  printf '  %s\n' "${results[@]}"
  if [ "$failed" -eq 0 ]; then
    printf '%s: PASS\n' "$label"
  else
    printf '%s: FAILED (%d step(s) failed)\n' "$label" "$failed"
  fi
} | tee "$summary"

[ "$failed" -eq 0 ]
