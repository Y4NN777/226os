#!/usr/bin/env bash
# Route-table gate (ADR-0006 `security`, G4; decision B1).
#
# Passes ONLY while:
#   0. the middleware chain installed by NewRouter equals scripts/router-middleware.txt;
#   1. every route mounted by httpx.NewRouter (the chi route table, walked with chi.Walk) is in
#      EXACTLY ONE of: scripts/public-routes.txt (public by design), scripts/unimplemented-routes.txt
#      (still answered 501; starts with all 47 operations and shrinks), scripts/authz-tests.txt
#      ("METHOD /path TestName": the named Go test must exist, be named TestAuthz<OperationID>..., not skip,
#      name the route and mention a denial status);
#   2. no router, mux or route is built or registered outside httpx.NewRouter (go/ast scan of the
#      non-test, non-generated Go files under src/api).
# Needs no Docker: the test files come from `go list` (no TestMain), the test names from their Go AST.
# A missing list file, a malformed line, a route in two lists, an unimplemented entry that is not an
# operation of the contract, or a public/authz entry for a route that is not mounted fails the gate.
#
# Mechanism: scripts/authz/routes_check.go.txt is injected into the httpx package as an external
# test file through `go test -overlay` (internal packages cannot be imported from scripts/), so no
# application file is edited. Exit 0 = pass, 1 = fail.
#
# Usage: scripts/check-authz-routes.sh [--router-override FILE] [--lists-dir DIR] [--scan-extra DIR]
#   Test hooks (used only by scripts/tests/authz-routes_test.sh, never passed by the Makefile):
#   --router-override FILE  compile FILE instead of router.go
#   --lists-dir DIR         read the three list files from DIR instead of scripts/
#   --scan-extra DIR        also scan DIR for routes built outside NewRouter and for Test functions
#   --server-override FILE  compile FILE instead of src/api/internal/app/server.go (an implemented operation)
#   --scan-extra-module DIR a scratch Go module whose test files count like the project's
#   --test-run-pattern RE   run another go test pattern (proves a run that matches nothing is not a pass)
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.27.4}"
# The gate must not be steered by the caller's Go environment (GOFLAGS=-exec could forge a PASS).
export GOFLAGS="" GOENV=off GOWORK=off

pkg=./src/api/internal/platform/httpx
pkgdir="$root/src/api/internal/platform/httpx"
check_src="$root/scripts/authz/routes_check.go.txt"
listdir="$root/scripts"
override=""
scan_extra=""
run_pattern='^TestAuthzRouteTable$'
scan_module=""
server_override=""

abs() { (cd "$1" && pwd); }
while [ $# -gt 0 ]; do
  case "$1" in
    --router-override)
      [ -f "${2-}" ] || { printf 'check-authz-routes: override file not found: %s\n' "${2-}" >&2; exit 2; }
      override="$(abs "$(dirname "$2")")/$(basename "$2")"
      shift 2
      ;;
    --lists-dir)
      [ -d "${2-}" ] || { printf 'check-authz-routes: lists dir not found: %s\n' "${2-}" >&2; exit 2; }
      listdir="$(abs "$2")"
      shift 2
      ;;
    --scan-extra)
      [ -d "${2-}" ] || { printf 'check-authz-routes: scan dir not found: %s\n' "${2-}" >&2; exit 2; }
      scan_extra="$(abs "$2")"
      shift 2
      ;;
    --server-override)
      [ -f "${2-}" ] || { printf 'check-authz-routes: server override file not found: %s\n' "${2-}" >&2; exit 2; }
      server_override="$(abs "$(dirname "$2")")/$(basename "$2")"
      shift 2
      ;;
    --scan-extra-module)
      [ -f "${2-}/go.mod" ] || { printf 'check-authz-routes: scan module not found: %s\n' "${2-}" >&2; exit 2; }
      scan_module="$(abs "$2")"
      shift 2
      ;;
    --test-run-pattern)
      [ -n "${2-}" ] || { printf 'check-authz-routes: --test-run-pattern needs a value\n' >&2; exit 2; }
      run_pattern="$2"
      shift 2
      ;;
    *) printf 'usage: %s [--router-override FILE] [--lists-dir DIR] [--scan-extra DIR] [--test-run-pattern RE] [--scan-extra-module DIR] [--server-override FILE]\n' "$0" >&2; exit 64 ;;
  esac
done

for f in "$check_src" "$listdir/public-routes.txt" "$listdir/unimplemented-routes.txt" "$listdir/authz-tests.txt" "$listdir/router-middleware.txt"; do
  [ -f "$f" ] || { printf 'check-authz-routes: BLOCKED: %s is missing\n' "$f" >&2; exit 1; }
done

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

{
  printf '{"Replace":{"%s":"%s"' "$pkgdir/zz_authz_routes_test.go" "$check_src"
  [ -n "$override" ] && printf ',"%s":"%s"' "$pkgdir/router.go" "$override"
  [ -n "$server_override" ] && printf ',"%s":"%s"' "$root/src/api/internal/app/server.go" "$server_override"
  printf '}}\n'
} >"$tmp/overlay.json"

# The test files the Go toolchain builds (default and integration tags), as `go list` reports them:
# build constraints, _ and . prefixes, _GOOS/_GOARCH suffixes and ignored directories all apply.
# `go list` loads packages; it runs no TestMain, so it needs no Docker.
list_tests() { # DIR TAGS
  (cd "$1" && go list -tags "$2" -f '{{range .TestGoFiles}}{{$.Dir}}/{{.}}{{"\n"}}{{end}}{{range .XTestGoFiles}}{{$.Dir}}/{{.}}{{"\n"}}{{end}}' "${3:-./...}")
}
: >"$tmp/testfiles.txt"
for tags in "" "integration"; do
  if ! list_tests "$root" "$tags" ./src/api/... >>"$tmp/testfiles.txt" 2>"$tmp/list.err"; then
    cat "$tmp/list.err" >&2
    printf 'check-authz-routes: BLOCKED: go list failed (tags=%s)\n' "$tags" >&2
    exit 1
  fi
  if [ -n "$scan_module" ] && ! list_tests "$scan_module" "$tags" >>"$tmp/testfiles.txt" 2>"$tmp/list.err"; then
    cat "$tmp/list.err" >&2
    printf 'check-authz-routes: BLOCKED: go list failed in the scan module\n' >&2
    exit 1
  fi
done

rc=0
AUTHZ_TEST_FILES="$tmp/testfiles.txt" AUTHZ_OPENAPI="$root/api/openapi.yaml" AUTHZ_ROUTER_OVERRIDE="$override" AUTHZ_LISTS_DIR="$listdir" AUTHZ_SCAN_ROOT="$root/src/api" AUTHZ_SCAN_EXTRA="$scan_extra" \
  go test -count=1 -v -run "$run_pattern" -overlay "$tmp/overlay.json" "$pkg" >"$tmp/out.txt" 2>&1 || rc=$?
# A build failure has no "--- " test line: show the compiler output in full, never a filtered tail.
if grep -q -- '^--- ' "$tmp/out.txt"; then
  grep -E 'route |mounted routes|authz-gate|^(---|FAIL|ok|panic)' "$tmp/out.txt" || true
else
  cat "$tmp/out.txt"
fi

if [ "$rc" -ne 0 ]; then
  printf 'check-authz-routes: FAIL (a mounted route is not in exactly one of the three lists, a named authz test is missing, a router is built outside httpx.NewRouter, or the check could not run)\n' >&2
  exit 1
fi
# The test must actually have run and passed: `go test -run` with no match also exits 0.
if ! grep -q -- '--- PASS: TestAuthzRouteTable' "$tmp/out.txt"; then
  printf 'check-authz-routes: FAIL (TestAuthzRouteTable did not run; refusing a silent pass)\n' >&2
  exit 1
fi
echo "check-authz-routes: PASS"
