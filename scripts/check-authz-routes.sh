#!/usr/bin/env bash
# Negative-authz route-table gate (ADR-0006 `security`, G4).
#
# Passes ONLY while every route mounted by httpx.NewRouter (the chi route table, walked with
# chi.Walk) is listed in scripts/public-routes.txt. The first mounted route outside the
# allowlist fails the gate with a message that an authz negative test is required. Today only
# GET /healthz is mounted, so the gate passes and says that no protected routes exist yet.
#
# Mechanism: scripts/authz/routes_check.go.txt is injected into the httpx package as an
# external test file through `go test -overlay` (internal packages cannot be imported from
# scripts/), so no application file is edited. Exit 0 = pass, 1 = fail.
#
# Usage: scripts/check-authz-routes.sh [--router-override FILE]
#   --router-override FILE   TEST HOOK (used only by scripts/tests/authz-routes_test.sh):
#                            compile FILE instead of router.go, to prove the gate goes RED
#                            when a route is added. Never passed by the Makefile.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.27.4}"

pkg=./src/api/internal/platform/httpx
pkgdir="$root/src/api/internal/platform/httpx"
check_src="$root/scripts/authz/routes_check.go.txt"
allowlist="$root/scripts/public-routes.txt"

override=""
case "${1-}" in
  "") ;;
  --router-override)
    override="${2-}"
    [ -f "$override" ] || { printf 'check-authz-routes: override file not found: %s\n' "$override" >&2; exit 2; }
    override="$(cd "$(dirname "$override")" && pwd)/$(basename "$override")"
    ;;
  *) printf 'usage: %s [--router-override FILE]\n' "$0" >&2; exit 64 ;;
esac
[ -f "$check_src" ] && [ -f "$allowlist" ] || { printf 'check-authz-routes: BLOCKED: %s or %s is missing\n' "$check_src" "$allowlist" >&2; exit 1; }

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

{
  printf '{"Replace":{"%s":"%s"' "$pkgdir/zz_authz_routes_test.go" "$check_src"
  [ -n "$override" ] && printf ',"%s":"%s"' "$pkgdir/router.go" "$override"
  printf '}}\n'
} >"$tmp/overlay.json"

rc=0
AUTHZ_PUBLIC_ROUTES="$allowlist" go test -count=1 -v -run '^TestAuthzRouteTable$' -overlay "$tmp/overlay.json" "$pkg" >"$tmp/out.txt" 2>&1 || rc=$?
grep -E 'route |mounted routes|no protected routes|authz negative test|^(---|FAIL|ok|panic)' "$tmp/out.txt" || cat "$tmp/out.txt"

if [ "$rc" -ne 0 ]; then
  printf 'check-authz-routes: FAIL (a mounted route is not public and not proven by an authz negative test, or the check could not run)\n' >&2
  exit 1
fi
# The test must actually have run and passed: `go test -run` with no match also exits 0.
if ! grep -q -- '--- PASS: TestAuthzRouteTable' "$tmp/out.txt"; then
  printf 'check-authz-routes: FAIL (TestAuthzRouteTable did not run; refusing a silent pass)\n' >&2
  exit 1
fi
if grep -q 'no protected routes exist yet' "$tmp/out.txt"; then
  echo "check-authz-routes: PASS (no protected routes exist yet; every mounted route is public by allowlist)"
else
  echo "check-authz-routes: PASS"
fi
