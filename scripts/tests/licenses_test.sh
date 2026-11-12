#!/usr/bin/env bash
# Tests for scripts/check-licenses.py against fixtures: allowed list, SPDX OR/AND, off-list,
# UNKNOWN and empty input all behave fail-closed.
# Usage: scripts/tests/licenses_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
checker="$root/scripts/check-licenses.py"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# case NAME EXPECT(pass|fail) GO_CSV_CONTENT JS_JSON_CONTENT
case_() {
  local name="$1" expect="$2" rc=0 out
  printf '%b' "$3" >"$tmp/go.csv"
  printf '%s' "$4" >"$tmp/js.json"
  out="$(python3 -I "$checker" --go-csv "$tmp/go.csv" --js-json "$tmp/js.json" --out "$tmp/out" --first-party github.com/Y4NN777/226os 2>&1)" || rc=$?
  if [ "$expect" = pass ]; then
    [ "$rc" -eq 0 ] && ok "$name" || ko "$name (rc=$rc out=$out)"
  else
    [ "$rc" -ne 0 ] && ok "$name" || ko "$name (rc=$rc out=$out)"
  fi
}
js_ok='{"MIT":[{"name":"a","versions":["1.0.0"]}],"Apache-2.0":[{"name":"b","versions":["2.0.0"]}]}'
go_ok='github.com/x/y,https://x,MIT\ngolang.org/x/text,https://x,BSD-3-Clause\n'

case_ "all allowed licences pass" pass "$go_ok" "$js_ok"
case_ "first-party module is not checked" pass 'github.com/Y4NN777/226os/apps/api,https://x,Unknown\nx/y,u,MIT\n' "$js_ok"
case_ "BSD-2, ISC, Unlicense pass" pass 'a/b,u,BSD-2-Clause\nc/d,u,ISC\ne/f,u,Unlicense\n' '{"ISC":[{"name":"i","versions":["1"]}],"Unlicense":[{"name":"u","versions":["1"]}]}'
case_ "OR expression with one allowed alternative passes" pass "$go_ok" '{"(MIT OR GPL-3.0-only)":[{"name":"c","versions":["1"]}]}'
case_ "parenthesised OR of allowed passes" pass "$go_ok" '{"(MIT OR Apache-2.0)":[{"name":"c","versions":["1"]}]}'
case_ "synonym 'Apache 2.0' passes" pass "$go_ok" '{"Apache 2.0":[{"name":"c","versions":["1"]}]}'
case_ "AND with an off-list part fails" fail "$go_ok" '{"(MIT AND GPL-3.0-only)":[{"name":"c","versions":["1"]}]}'
case_ "off-list JS licence fails" fail "$go_ok" '{"GPL-3.0":[{"name":"c","versions":["1"]}]}'
case_ "OR of two off-list licences fails" fail "$go_ok" '{"(GPL-2.0 OR LGPL-3.0)":[{"name":"c","versions":["1"]}]}'
case_ "off-list Go licence fails" fail 'a/b,u,MPL-2.0\n' "$js_ok"
case_ "Go Unknown licence fails" fail 'a/b,u,Unknown\n' "$js_ok"
case_ "JS UNKNOWN licence fails" fail "$go_ok" '{"UNKNOWN":[{"name":"c","versions":["1"]}]}'
case_ "JS empty licence fails" fail "$go_ok" '{"":[{"name":"c","versions":["1"]}]}'
case_ "MIT* (non-SPDX wildcard) fails" fail "$go_ok" '{"MIT*":[{"name":"c","versions":["1"]}]}'
case_ "empty Go inventory fails (nothing proved)" fail '' "$js_ok"
case_ "empty JS inventory fails (nothing proved)" fail "$go_ok" '{}'
case_ "malformed JSON fails" fail "$go_ok" 'not json'

rc=0
python3 -I "$checker" --go-csv "$tmp/go.csv" --js-json "$tmp/js.json" --out "$tmp/out2" --first-party github.com/Y4NN777/226os >/dev/null 2>&1 || rc=$?
[ -s "$tmp/out2/licenses-summary.txt" ] && ok "writes an inventory artifact even on failure" || ko "no inventory artifact"
finish
