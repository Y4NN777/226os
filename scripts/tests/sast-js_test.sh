#!/usr/bin/env bash
# RED/GREEN for scripts/sast-js.sh. Known-bad samples live only in a temp dir outside the repository
# (never committed); the real tree must be clean. Needs Docker and the pinned Semgrep image.
# Usage: scripts/tests/sast-js_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
gate="$root/scripts/sast-js.sh"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
export REPORTS_DIR="$tmp/reports"

# bad NAME FILE-CONTENT: the sample must make the gate fail.
bad() {
  local name="$1" d="$tmp/$1" rc=0 out
  mkdir -p "$d"
  printf '%s\n' "$2" >"$d/sample.tsx"
  out="$("$gate" "$d" 2>&1)" || rc=$?
  if [ "$rc" -ne 0 ]; then ok "RED: $name is rejected"; else ko "RED: $name passed (out=$out)"; fi
}
bad dangerouslySetInnerHTML 'export const A = ({ html }: { html: string }) => <div dangerouslySetInnerHTML={{ __html: html }} />;'
bad eval 'export const f = (s: string) => eval(s);'
bad new-Function 'export const f = (s: string) => new Function(s)();'
bad innerHTML 'export const f = (e: HTMLElement, s: string) => { e.innerHTML = s; };'
bad document.write 'export const f = (s: string) => { document.write(s); };'
bad string-timer 'export const f = () => { setTimeout("alert(1)", 10); };'
bad token-in-localStorage 'export const f = (t: string) => { localStorage.setItem("authToken", t); };'
bad target-blank 'export const A = () => <a href="https://example.com" target="_blank">x</a>;'

mkdir -p "$tmp/good"
printf '%s\n' 'export const A = ({ t }: { t: string }) => <a href="https://example.com" target="_blank" rel="noopener noreferrer">{t}</a>;' >"$tmp/good/ok.tsx"
rc=0
out="$("$gate" "$tmp/good" 2>&1)" || rc=$?
if [ "$rc" -eq 0 ]; then ok "GREEN: a clean sample passes"; else ko "GREEN sample (rc=$rc out=$out)"; fi

mkdir -p "$tmp/empty"
rc=0
"$gate" "$tmp/empty" >/dev/null 2>&1 || rc=$?
if [ "$rc" -ne 0 ]; then ok "a directory with nothing to scan fails (no silent pass)"; else ko "empty dir passed"; fi

rc=0
out="$(REPORTS_DIR="$tmp/reports-real" "$gate" 2>&1)" || rc=$?
if [ "$rc" -eq 0 ]; then ok "GREEN: the real apps/web/src tree has no findings"; else ko "real tree (rc=$rc out=$out)"; fi
finish
