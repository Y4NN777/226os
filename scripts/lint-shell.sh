#!/usr/bin/env bash
# ShellCheck over every shell script in scripts/ (G1, ADR-0006 `lint`).
# Tool: koalaman/shellcheck pinned by image digest (below), run with --network none and the
# repository mounted read-only. Docker is already a doctor prerequisite, so the pinned image
# is used everywhere instead of whatever shellcheck a machine happens to have.
# Severity floor: warning (warnings and errors fail). Info-level notes are not gate failures; the
# only ones present are SC2015 on the `cond && ok "..." || ko "..."` test idiom, where ok never fails.
set -euo pipefail

SHELLCHECK_IMAGE="${SHELLCHECK_IMAGE:-koalaman/shellcheck@sha256:61862eba1fcf09a484ebcc6feea46f1782532571a34ed51fedf90dd25f925a8d}" # koalaman/shellcheck:v0.11.0

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

shopt -s nullglob
files=(scripts/*.sh scripts/tests/*.sh)
[ "${#files[@]}" -gt 0 ] || { echo "lint-shell: BLOCKED: no scripts found" >&2; exit 1; }

docker run --rm --network none --read-only --user "$(id -u):$(id -g)" \
  -v "$root:/mnt:ro" -w /mnt "$SHELLCHECK_IMAGE" -x --severity=warning "${files[@]}"
echo "lint-shell: OK (${#files[@]} scripts)"
