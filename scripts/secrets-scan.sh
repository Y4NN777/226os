#!/usr/bin/env bash
# Secret scan (G4, ADR-0006 `security`): gitleaks over (1) the working tree and (2) the git
# history reachable from HEAD (the branch). Fails on any finding; findings are redacted.
#
# Tool: zricethezav/gitleaks pinned by image digest (below), run with --network none, a read-only
# root filesystem and the caller's uid. Rules: gitleaks defaults plus .gitleaks.toml when present
# (justified allowlists only).
# (1) Working tree = tracked files plus untracked files that are NOT git-ignored, exported to a
#     temp dir. Git-ignored files (throwaway local .env.* files, node_modules, build output) are
#     therefore not scanned, and nothing needs an allowlist for them.
# (2) History = `gitleaks git` over HEAD. CI must fetch full history (fetch-depth: 0); a shallow
#     clone is refused because it would silently scan less than the branch.
# Network: none (the scanner container has --network none).
set -euo pipefail

GITLEAKS_IMAGE="${GITLEAKS_IMAGE:-zricethezav/gitleaks@sha256:691af3c7c5a48b16f187ce3446d5f194838f91238f27270ed36eef6359a574d9}" # zricethezav/gitleaks:v8.30.0

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

if [ "$(git rev-parse --is-shallow-repository)" = true ]; then
  echo "secrets-scan: BLOCKED: shallow clone; fetch full history (fetch-depth: 0) so the whole branch is scanned" >&2
  exit 1
fi

cfg=()
if [ -f .gitleaks.toml ]; then cfg=(--config /repo-config/.gitleaks.toml); fi
cfg_mount=()
if [ -f .gitleaks.toml ]; then cfg_mount=(-v "$root/.gitleaks.toml:/repo-config/.gitleaks.toml:ro"); fi
# Reviewed, fingerprint-level exceptions only (see the file header); never a path or rule allowlist.
cfg+=(--gitleaks-ignore-path /repo-config/gitleaks-ignore)
cfg_mount+=(-v "$root/scripts/gitleaks-ignore.txt:/repo-config/gitleaks-ignore/.gitleaksignore:ro")

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir "$tmp/tree"

run_gitleaks() { # <mount-spec> <subcommand-and-args...>
  local mount="$1"
  shift
  docker run --rm --network none --read-only --tmpfs /tmp \
    --user "$(id -u):$(id -g)" -e HOME=/tmp \
    -v "$mount" ${cfg_mount[@]+"${cfg_mount[@]}"} \
    "$GITLEAKS_IMAGE" "$@" --redact --no-banner --exit-code 1 ${cfg[@]+"${cfg[@]}"}
}

# 1. Working tree (tracked + untracked, minus ignored and minus files deleted in the tree).
echo "== gitleaks: working tree =="
git ls-files -z --cached --others --exclude-standard |
  while IFS= read -r -d '' f; do [ -e "$f" ] && printf '%s\0' "$f"; done |
  tar --null -T - -cf - | tar -xf - -C "$tmp/tree"
run_gitleaks "$tmp/tree:/scan:ro" dir /scan

# 2. History reachable from HEAD.
echo "== gitleaks: git history of HEAD =="
run_gitleaks "$root:/repo:ro" git /repo --log-opts="HEAD"

echo "secrets-scan: PASS (working tree and history of HEAD: no leaks found)"
