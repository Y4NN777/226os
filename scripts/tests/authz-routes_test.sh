#!/usr/bin/env bash
# RED/GREEN evidence for scripts/check-authz-routes.sh (negative-authz route table gate).
#   GREEN: the real router (only GET /healthz) passes and says no protected routes exist.
#   RED:   a scratch copy of router.go with an extra route makes the gate fail with the
#          "authz negative test is required" message; the real router.go is never edited.
#   RED:   a missing allowlist entry (empty allowlist) also fails.
# Usage: scripts/tests/authz-routes_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
gate="$root/scripts/check-authz-routes.sh"
router="$root/apps/api/internal/platform/httpx/router.go"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

rc=0
out="$("$gate" 2>&1)" || rc=$?
if [ "$rc" -eq 0 ] && grep -q 'no protected routes exist yet' <<<"$out"; then ok "GREEN: only public routes are mounted"; else ko "GREEN (rc=$rc, out=$out)"; fi

# Scratch router with one extra, non-allowlisted route.
if ! grep -q 'r.Get("/healthz", healthz)' "$router"; then ko "cannot inject: router.go no longer contains the healthz registration"; finish; exit 1; fi
sed 's#r.Get("/healthz", healthz)#r.Get("/healthz", healthz)\n\tr.Get("/admin/users", healthz)#' "$router" >"$tmp/router.go"
grep -q '/admin/users' "$tmp/router.go" || { ko "injection did not apply"; finish; exit 1; }
rc=0
out="$("$gate" --router-override "$tmp/router.go" 2>&1)" || rc=$?
if [ "$rc" -ne 0 ] && grep -q 'GET /admin/users' <<<"$out" && grep -q 'authz negative test is required' <<<"$out"; then
  ok "RED: an extra mounted route fails the gate with the authz-negative-test message"
else ko "RED injected route (rc=$rc, out=$out)"; fi

sed 's#r.Get("/healthz", healthz)#r.Get("/healthz", healthz)\n\tr.Post("/healthz", healthz)#' "$router" >"$tmp/router2.go"
rc=0
out="$("$gate" --router-override "$tmp/router2.go" 2>&1)" || rc=$?
[ "$rc" -ne 0 ] && grep -q 'POST /healthz' <<<"$out" && ok "RED: another method on a public path is not public" || ko "RED method (rc=$rc, out=$out)"

# The real router is untouched.
git -C "$root" diff --quiet -- apps/api/internal/platform/httpx/router.go && ok "router.go is unmodified" || ko "router.go was modified"

finish
