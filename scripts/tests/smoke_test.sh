#!/usr/bin/env bash
# Tests for scripts/smoke.sh with a fake curl and a fake compose wrapper: every check reports PASS or FAIL,
# a failing check fails the run, a missing base URL is BLOCKED, and /readyz is never expected at the edge (B2).
# Usage: scripts/tests/smoke_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
smoke="$root/scripts/smoke.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/bin"
map="$tmp/map"

# Fake curl: the last argument is the URL; the -w format is answered from the map
# (lines: URL|http_code|content_type|redirect_url). An unknown URL behaves like a refused connection.
cat >"$tmp/bin/curl" <<'EOF'
#!/usr/bin/env bash
fmt="" url=""
while [ $# -gt 0 ]; do
  case "$1" in
    -w) fmt="$2"; shift 2 ;;
    --max-time) shift 2 ;;
    -*) shift ;;
    *) url="$1"; shift ;;
  esac
done
line="$(grep -F "$url|" "$FAKE_CURL_MAP" | head -n 1 || true)"
[ -n "$line" ] || { printf '000'; exit 7; }
IFS='|' read -r _ code ctype redirect <<<"$line"
out="${fmt//%\{http_code\}/$code}"
out="${out//%\{content_type\}/$ctype}"
out="${out//%\{redirect_url\}/$redirect}"
printf '%s' "$out"
EOF
chmod +x "$tmp/bin/curl"

# Fake compose wrapper: exits with FAKE_COMPOSE_RC and records its arguments.
cat >"$tmp/compose" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >>"$FAKE_COMPOSE_LOG"
exit "${FAKE_COMPOSE_RC:-0}"
EOF
chmod +x "$tmp/compose"

run() {
  rc=0
  : >"$tmp/compose.log"
  out="$(PATH="$tmp/bin:$PATH" FAKE_CURL_MAP="$map" FAKE_COMPOSE_LOG="$tmp/compose.log" SMOKE_COMPOSE="$tmp/compose" "$smoke" "$@" 2>&1)" || rc=$?
}

good_https() {
  cat >"$map" <<'EOM'
https://staging.example.test/healthz|200|application/json|
http://staging.example.test/|308|text/html|https://staging.example.test/
https://staging.example.test/readyz|200|text/html|
EOM
}

good_https
FAKE_COMPOSE_RC=0 run staging-test https://staging.example.test
if [ "$rc" -eq 0 ] && grep -q 'PASS' <<<"$out" && ! grep -q 'FAIL' <<<"$out" && grep -q 'smoke staging-test: PASS' <<<"$out"; then
  ok "a healthy staging passes every check"
else ko "healthy staging (rc=$rc out=$out)"; fi
grep -q 'staging-test exec -T api /healthcheck /readyz' "$tmp/compose.log" && ok "/readyz is checked inside the container with /healthcheck" || ko "inside check ($(cat "$tmp/compose.log"))"

good_https
FAKE_COMPOSE_RC=1 run staging-test https://staging.example.test
[ "$rc" -eq 1 ] && grep -q 'FAIL.*readyz inside' <<<"$out" && ok "a failing /readyz inside the container fails the run" || ko "inside failure (rc=$rc out=$out)"

sed -i 's#^https://staging.example.test/healthz|200#https://staging.example.test/healthz|503#' "$map"
FAKE_COMPOSE_RC=0 run staging-test https://staging.example.test
[ "$rc" -eq 1 ] && grep -q 'FAIL.*healthz' <<<"$out" && ok "an unhealthy /healthz fails the run" || ko "healthz 503 (rc=$rc out=$out)"

good_https
sed -i 's#^http://staging.example.test/|308#http://staging.example.test/|200#' "$map"
FAKE_COMPOSE_RC=0 run staging-test https://staging.example.test
[ "$rc" -eq 1 ] && grep -q 'FAIL.*redirect' <<<"$out" && ok "no HTTP to HTTPS redirect fails the run" || ko "redirect (rc=$rc out=$out)"

good_https
sed -i 's#^https://staging.example.test/readyz|200|text/html#https://staging.example.test/readyz|200|application/json#' "$map"
FAKE_COMPOSE_RC=0 run staging-test https://staging.example.test
[ "$rc" -eq 1 ] && grep -q 'FAIL.*readyz.*edge' <<<"$out" && ok "an API answer for /readyz at the edge fails (B2: not routed)" || ko "readyz at the edge (rc=$rc out=$out)"

: >"$map"
FAKE_COMPOSE_RC=0 run staging-test https://staging.example.test
[ "$rc" -eq 1 ] && grep -q 'FAIL.*healthz' <<<"$out" && ok "an unreachable host fails the run" || ko "unreachable (rc=$rc out=$out)"

run staging-test
[ "$rc" -eq 2 ] && grep -q 'BLOCKED' <<<"$out" && ok "a remote profile without a base URL is BLOCKED" || ko "no base url (rc=$rc out=$out)"

run staging-test http://staging.example.test
[ "$rc" -eq 2 ] && grep -q 'https' <<<"$out" && ok "a remote profile refuses a non-HTTPS base URL" || ko "http base (rc=$rc out=$out)"

run nonsense
[ "$rc" -eq 64 ] && ok "an unknown profile is a usage error" || ko "unknown profile (rc=$rc)"

run
[ "$rc" -eq 64 ] && ok "no argument is a usage error" || ko "no arguments (rc=$rc)"

cat >"$map" <<'EOM'
http://127.0.0.1:8088/healthz|200|application/json|
http://127.0.0.1:8088/readyz|200|text/html|
EOM
FAKE_COMPOSE_RC=0 run local-dev
[ "$rc" -eq 0 ] && grep -q 'NOT RUN.*redirect' <<<"$out" && grep -q 'smoke local-dev: PASS' <<<"$out" \
  && ok "local-dev uses the loopback port and reports the TLS and redirect checks as NOT RUN" || ko "local-dev (rc=$rc out=$out)"

finish
