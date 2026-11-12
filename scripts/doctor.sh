#!/usr/bin/env bash
# `make doctor` (ADR-0006): verify the pinned toolchain before any other target runs.
# Read-only. Exit 0 only when every check passes; otherwise every problem is listed and the exit is 1.
#   Go     the toolchain selected by GOTOOLCHAIN must be exactly go<go.mod "go" line>.
#   Node   at least the `engines.node` floor of package.json. A different major than .nvmrc
#          (which CI uses) is only a NOTE: ADR-0006 allows a newer or older local LTS above the floor.
#   pnpm   exactly the `packageManager` version of package.json.
#   Docker the docker CLI and the compose plugin must be present (a daemon is checked by the targets that use it).
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

fails=0
ok() { printf 'ok    %s\n' "$*"; }
bad() { printf 'FAIL  %s\n' "$*" >&2; fails=$((fails + 1)); }
note() { printf 'NOTE  %s\n' "$*"; }
have() { command -v "$1" >/dev/null 2>&1; }

# --- Go ---------------------------------------------------------------------------------------
want_go="$(awk '$1 == "go" {print $2; exit}' go.mod)"
if [ -z "$want_go" ]; then
  bad "go.mod has no go directive"
elif ! have go; then
  bad "go is not installed (need go${want_go})"
else
  export GOTOOLCHAIN="${GOTOOLCHAIN:-go${want_go}}"
  if [ "$GOTOOLCHAIN" != "go${want_go}" ]; then
    bad "GOTOOLCHAIN=${GOTOOLCHAIN} but go.mod pins go${want_go}"
  elif got_go="$(go env GOVERSION 2>&1)" && [ "$got_go" = "go${want_go}" ]; then
    ok "go ${got_go} (GOTOOLCHAIN=${GOTOOLCHAIN})"
  else
    bad "go toolchain is '${got_go:-unknown}', need go${want_go} (GOTOOLCHAIN=${GOTOOLCHAIN})"
  fi
fi

# --- Node -------------------------------------------------------------------------------------
if ! have node; then
  bad "node is not installed"
elif ! have python3; then
  bad "python3 is required (licence check, container verification)"
else
  floor="$(python3 -I - <<'PY'
import json, re
v = json.load(open("package.json")).get("engines", {}).get("node", "")
m = re.fullmatch(r">=\s*(\d+)(\.\d+){0,2}", v.strip())
print(m.group(1) if m else "")
PY
)"
  have_major="$(node -p 'process.versions.node.split(".")[0]')"
  nvm_major="$(tr -d '[:space:]v' <.nvmrc 2>/dev/null | cut -d. -f1)"
  if [ -z "$floor" ]; then
    bad "package.json engines.node is not of the form '>=N'"
  elif [ "$have_major" -lt "$floor" ]; then
    bad "node $(node -v) is below the engines floor >=${floor}"
  else
    ok "node $(node -v) (engines >=${floor})"
    if [ -n "$nvm_major" ] && [ "$have_major" != "$nvm_major" ]; then
      note "local node major ${have_major} differs from .nvmrc (${nvm_major}); CI uses .nvmrc"
    fi
  fi
fi

# --- pnpm -------------------------------------------------------------------------------------
want_pnpm="$(python3 -I -c 'import json; print(json.load(open("package.json")).get("packageManager", ""))' 2>/dev/null)" || want_pnpm=""
want_pnpm="${want_pnpm#pnpm@}"
want_pnpm="${want_pnpm%%+*}"
if [ -z "$want_pnpm" ]; then
  bad "package.json packageManager is not pnpm@<version>"
elif ! have pnpm; then
  bad "pnpm is not installed (need pnpm ${want_pnpm})"
elif got_pnpm="$(pnpm --version 2>&1)" && [ "$got_pnpm" = "$want_pnpm" ]; then
  ok "pnpm ${got_pnpm} (packageManager)"
else
  bad "pnpm is '${got_pnpm:-unknown}', packageManager pins ${want_pnpm}"
fi

# --- Docker -----------------------------------------------------------------------------------
if ! have docker; then
  bad "docker is not installed"
elif compose_v="$(docker compose version --short 2>&1)"; then
  ok "docker $(docker --version | awk '{print $3}' | tr -d ,) with compose ${compose_v}"
else
  bad "docker compose plugin is not available: ${compose_v}"
fi

have git && ok "git $(git --version | awk '{print $3}')" || bad "git is not installed"

if [ "$fails" -ne 0 ]; then
  printf 'doctor: FAILED (%d problem(s))\n' "$fails" >&2
  exit 1
fi
echo "doctor: OK"
