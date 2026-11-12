#!/usr/bin/env bash
# `make licenses` (G4/G7): runtime dependency licence inventory against the owner-approved list.
#   Go  github.com/google/go-licenses/v2 (pinned through `go tool`) over ./src/api/... : the packages the
#       API binary is built from, not build tools.
#   JS  pnpm licenses list --prod --json for the web workspace package (production dependencies only).
# Writes the inventory under $REPORTS_DIR/licenses (default reports/licenses, git-ignored).
# THIRD_PARTY_NOTICES.md is NOT edited. The licence list itself lives in scripts/check-licenses.py.
# Network: none beyond what the lockfile-pinned dependencies already are (pnpm reads node_modules metadata).
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.27.4}"
out="${REPORTS_DIR:-reports}/licenses"
mkdir -p "$out"

module="$(go list -m)"
go tool go-licenses report ./src/api/... >"$out/go-licenses.csv" 2>"$out/go-licenses.stderr"
pnpm licenses list --prod --json --filter web >"$out/js-licenses.raw.json"

exec python3 -I scripts/check-licenses.py \
  --go-csv "$out/go-licenses.csv" \
  --js-json "$out/js-licenses.raw.json" \
  --out "$out" \
  --first-party "$module"
