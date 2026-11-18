#!/usr/bin/env bash
# Tests for scripts/e2e-junit-check.sh: the junit XML of the E2E job is published from a public repository, so it must
# not hold the E2E account. The check compares the file with E2E_USER and E2E_PASSWORD from the environment, never
# prints them, and deletes a leaking file so that no later step can upload it.
# Usage: scripts/tests/e2e-junit-check_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
chk="$root/scripts/e2e-junit-check.sh"

[ -x "$chk" ] || { ko "scripts/e2e-junit-check.sh is missing or not executable"; finish; exit $?; }

tmp="$(mktemp -d)"
bs='\' # a backslash, so that no fixture below holds a literal backslash-u sequence
trap 'rm -rf "$tmp"' EXIT
run() { rc=0; out="$(env -i PATH="$PATH" E2E_USER="${U-ci@example.test}" E2E_PASSWORD="${P-sUp3r-S3cret}" "$chk" "$@" 2>&1)" || rc=$?; }

printf '<testsuites><testcase name="sign in"/></testsuites>\n' >"$tmp/clean.xml"
run "$tmp/clean.xml"
[ "$rc" -eq 0 ] && [ -f "$tmp/clean.xml" ] && ok "a clean report passes and is kept" || ko "clean (rc=$rc out=$out)"

printf '<failure message="expected sUp3r-S3cret to be visible"/>\n' >"$tmp/pw.xml"
run "$tmp/pw.xml"
if [ "$rc" -ne 0 ] && grep -q '::error' <<<"$out" && ! grep -q 'sUp3r-S3cret' <<<"$out" && [ ! -e "$tmp/pw.xml" ]; then ok "a report holding the password fails, the value is not printed, the file is deleted"; else ko "password (rc=$rc out=$out)"; fi

printf '<testcase name="login as ci@example.test"/>\n' >"$tmp/user.xml"
run "$tmp/user.xml"
if [ "$rc" -ne 0 ] && ! grep -q 'ci@example.test' <<<"$out" && [ ! -e "$tmp/user.xml" ]; then ok "a report holding the user fails and is deleted"; else ko "user (rc=$rc out=$out)"; fi

printf 'a & b *.* [x] \\ pa$$\n' >"$tmp/special.xml"
U=ok@x.test P='*.*' run "$tmp/special.xml"
[ "$rc" -ne 0 ] && ok "special characters in a value are matched literally (no pattern, no regex)" || ko "special (rc=$rc)"
printf 'abcdef\n' >"$tmp/nopattern.xml"
U=ok@x.test P='a.c' run "$tmp/nopattern.xml"
[ "$rc" -eq 0 ] && ok "a value is not a regular expression" || ko "regex (rc=$rc)"

printf '<failure message="typed p&amp;ss&lt;&quot;&apos;&gt;"/>\n' >"$tmp/escaped.xml"
U=ok@x.test P="p&ss<\"'>" run "$tmp/escaped.xml"
[ "$rc" -ne 0 ] && [ ! -e "$tmp/escaped.xml" ] && ok "a value written XML-escaped in the report is caught" || ko "escaped (rc=$rc)"
printf '%s\n' '<failure message="Expected: &quot;pa\&quot;ss\\w0rd&quot;"/>' >"$tmp/js.xml"
U=ok@x.test P='pa"ss\w0rd' run "$tmp/js.xml"
[ "$rc" -ne 0 ] && [ ! -e "$tmp/js.xml" ] && ok "a value written JS/JSON-escaped (\\\" and \\\\\\\\) inside the XML is caught" || ko "js-escaped (rc=$rc out=$out)"
printf '%s\n' '<failure message="Expected: &quot;mé\&quot;x&quot;"/>' >"$tmp/jsutf8.xml"
U=ok@x.test P='mé"x' run "$tmp/jsutf8.xml"
[ "$rc" -ne 0 ] && [ ! -e "$tmp/jsutf8.xml" ] && ok "a JS-escaped value with a non-ASCII character (kept literal by the runner) is caught" || ko "js-escaped non-ASCII (rc=$rc out=$out)"
printf '%s\n' "<failure message=\"Expected: &quot;m${bs}u00e9${bs}&quot;x&quot;\"/>" >"$tmp/jsascii.xml"
U=ok@x.test P='mé"x' run "$tmp/jsascii.xml"
[ "$rc" -ne 0 ] && [ ! -e "$tmp/jsascii.xml" ] && ok "the same value with the non-ASCII character written \\\\uXXXX is still caught" || ko "js-escaped ascii (rc=$rc out=$out)"
# Skeptic N4: other printings of a value. Uppercase \uXXXX, a JSON "\/" escape, a quote escaped while the backslash stays
# raw, and URL encoding (a value inside a logged URL).
printf '%s\n' "<failure message=\"Expected: m${bs}u00E9x\"/>" >"$tmp/upper.xml"
U=ok@x.test P='méx' run "$tmp/upper.xml"
[ "$rc" -ne 0 ] && [ ! -e "$tmp/upper.xml" ] && ok "a value with uppercase \\\\uXXXX escapes is caught" || ko "uppercase uXXXX (rc=$rc out=$out)"
printf '%s\n' '<failure message="Expected: a\/b9z"/>' >"$tmp/slash.xml"
U=ok@x.test P='a/b9z' run "$tmp/slash.xml"
[ "$rc" -ne 0 ] && [ ! -e "$tmp/slash.xml" ] && ok "a value with a JSON \\\\/ escape is caught" || ko "slash escape (rc=$rc out=$out)"
printf '%s\n' '<failure message="Expected: pa\&quot;ss\w0rd"/>' >"$tmp/rawq.xml"
U=ok@x.test P='pa"ss\w0rd' run "$tmp/rawq.xml"
[ "$rc" -ne 0 ] && [ ! -e "$tmp/rawq.xml" ] && ok "a value whose quote is escaped while its backslash stays raw is caught" || ko "raw quote form (rc=$rc out=$out)"
printf '%s\n' '<failure message="GET /login?u=ci%40example.test&amp;p=p%26ss%20w"/>' >"$tmp/url.xml"
U='ci@example.test' P='p&ss w' run "$tmp/url.xml"
[ "$rc" -ne 0 ] && [ ! -e "$tmp/url.xml" ] && ! grep -q 'example.test' <<<"$out" && ok "a URL-encoded value is caught and not printed" || ko "url-encoded (rc=$rc out=$out)"
printf '<failure message="typed p&#38;ss&#60;&#34;&#39;&#62;"/>\n' >"$tmp/numeric.xml"
U=ok@x.test P="p&ss<\"'>" run "$tmp/numeric.xml"
[ "$rc" -ne 0 ] && ok "a value written with numeric XML entities is caught" || ko "numeric (rc=$rc)"

run "$tmp/absent.xml"
[ "$rc" -eq 0 ] && grep -q 'no junit' <<<"$out" && ok "no report is not an error (nothing to upload)" || ko "absent (rc=$rc out=$out)"

printf 'x\n' >"$tmp/x.xml"
U='' run "$tmp/x.xml"
[ "$rc" -ne 0 ] && ok "an empty E2E_USER fails closed (an empty value would match everything)" || ko "empty user (rc=$rc)"
P='' run "$tmp/x.xml"
[ "$rc" -ne 0 ] && ok "an empty E2E_PASSWORD fails closed" || ko "empty password (rc=$rc)"
run
[ "$rc" -eq 64 ] && ok "no argument is a usage error" || ko "usage (rc=$rc)"

finish
