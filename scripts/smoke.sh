#!/usr/bin/env bash
# Post-deploy smoke test (docs/SHIP.md section 4.1 and 4.2, ROADMAP B2). Read-only: it sends GET requests and one
# `docker compose exec`, nothing else. Every check prints PASS, FAIL or NOT RUN; the run fails if any check fails.
#
# Usage: scripts/smoke.sh <profile> [base-url]
#   local-dev                    base URL defaults to http://127.0.0.1:${HTTP_PORT:-8088}
#   staging-test | production-live   base URL is required and must be https://<host>
#
# Checks
#   1. GET <base>/healthz answers 200 from the outside (for https, the certificate must verify: curl is never run
#      with -k).
#   2. GET http://<host>/ redirects to https (remote profiles only; NOT RUN for local-dev).
#   3. GET <base>/readyz is NOT answered by the API at the edge (B2: it is not routed; the web site answers it with
#      its page, so an application/json answer is a failure).
#   4. /readyz inside the container: scripts/compose.sh <profile> exec -T api /healthcheck /readyz (override the
#      wrapper with SMOKE_COMPOSE for tests). It needs the profile's secrets, like any compose command.
# Exit codes: 0 all checks passed, 1 a check failed, 2 BLOCKED (missing or unusable base URL), 64 usage.
# Application flows (sign-in, question, search...) are SHIP 4.3 to 4.5 and are NOT covered here.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
compose="${SMOKE_COMPOSE:-$root/scripts/compose.sh}"

usage() { printf 'usage: %s <local-dev | staging-test | production-live> [base-url]\n' "$0" >&2; exit 64; }

[ $# -ge 1 ] || usage
profile="$1"
base="${2:-}"

case "$profile" in
  local-dev) [ -n "$base" ] || base="http://127.0.0.1:${HTTP_PORT:-8088}" ;;
  staging-test | production-live)
    [ -n "$base" ] || { printf 'smoke %s: BLOCKED: a base URL (https://<host>) is required\n' "$profile" >&2; exit 2; }
    case "$base" in
      https://*) ;;
      *) printf 'smoke %s: BLOCKED: the base URL must be https://<host> (got %s)\n' "$profile" "$base" >&2; exit 2 ;;
    esac
    ;;
  *) usage ;;
esac
base="${base%/}"

failed=0
pass() { printf 'PASS    - %s\n' "$1"; }
fail() { printf 'FAIL    - %s\n' "$1"; failed=1; }
notrun() { printf 'NOT RUN - %s\n' "$1"; }

# fetch FORMAT URL: curl without -L and without -k; prints the -w format, or nothing when curl fails.
fetch() { curl -sS -o /dev/null -w "$1" --max-time 10 "$2" 2>/dev/null || true; }

code="$(fetch '%{http_code}' "$base/healthz")"
if [ "$code" = 200 ]; then pass "GET /healthz answers 200 from outside ($base)"; else fail "GET /healthz from outside answered '${code:-none}' ($base)"; fi

host="${base#https://}"
if [ "$profile" = local-dev ]; then
  notrun "http to https redirect (local-dev has no TLS)"
else
  read -r rcode rurl <<<"$(fetch '%{http_code} %{redirect_url}' "http://$host/")"
  case "${rcode:-}:${rurl:-}" in
    301:https://* | 302:https://* | 307:https://* | 308:https://*) pass "http://$host/ redirects to https" ;;
    *) fail "http://$host/ did not redirect to https (status '${rcode:-none}')" ;;
  esac
fi

ctype="$(fetch '%{content_type}' "$base/readyz")"
case "$ctype" in
  *application/json*) fail "GET /readyz at the edge was answered by the API ($ctype): it must not be routed (B2)" ;;
  *) pass "/readyz is not routed to the API at the edge (content type '${ctype:-none}')" ;;
esac

if "$compose" "$profile" exec -T api /healthcheck /readyz >/dev/null 2>&1; then
  pass "/readyz inside the api container (database reachable)"
else
  fail "/readyz inside the api container failed (compose exec; is the stack up and the profile's secrets readable?)"
fi

if [ "$failed" -eq 0 ]; then
  printf 'smoke %s: PASS\n' "$profile"
else
  printf 'smoke %s: FAIL\n' "$profile"
  exit 1
fi
