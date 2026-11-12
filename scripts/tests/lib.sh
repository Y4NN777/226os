# Minimal assertion helpers for plain-bash script tests (sourced, not executed).
# shellcheck shell=bash
pass=0
failn=0
ok() { pass=$((pass + 1)); printf 'ok   - %s\n' "$1"; }
ko() { failn=$((failn + 1)); printf 'FAIL - %s\n' "$1" >&2; }
finish() {
  printf '%d passed, %d failed\n' "$pass" "$failn"
  [ "$failn" -eq 0 ]
}
