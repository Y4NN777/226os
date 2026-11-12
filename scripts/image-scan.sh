#!/usr/bin/env bash
# Image vulnerability scan (G4, ADR-0006 `security`): Trivy on the built api, web and migrate images.
# Fails on any fixable HIGH or CRITICAL vulnerability (`--ignore-unfixed`: the owner rule is
# "fixable"; unfixable findings are not reported by this gate).
#
# Tool: aquasec/trivy pinned by image digest (below). Images are exported with `docker save`
# and scanned from a tar file: the Docker socket is NEVER mounted into the scanner. The scanner
# container runs as the caller's uid with a read-only root filesystem.
# Network: Trivy downloads its vulnerability database at gate time (ghcr.io); this is the one
# gate-time network use of this step and is documented in mk/30-security.mk. If the database
# cannot be fetched the scan fails (fail closed).
# Exceptions: an optional .trivyignore (CVE ids, each with a justification comment) is honoured.
# Usage: scripts/image-scan.sh [tag]      (default: scripts/image-tag.sh, i.e. what scripts/build-images.sh tagged)
set -euo pipefail

TRIVY_IMAGE="${TRIVY_IMAGE:-aquasec/trivy@sha256:e2b22eac59c02003d8749f5b8d9bd073b62e30fefaef5b7c8371204e0a4b0c08}" # aquasec/trivy:0.67.2
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
tag="${1:-$(scripts/image-tag.sh)}"
[ -n "$tag" ] || { printf 'usage: %s [image-tag]\n' "$0" >&2; exit 64; }
reports="${REPORTS_DIR:-reports}/trivy"
mkdir -p "$reports"
# Trivy database cache on the host (owned by the caller): a fresh named volume is root-owned and the
# scanner runs as the caller's uid. Holds only the public vulnerability DB.
cache="${XDG_CACHE_HOME:-$HOME/.cache}/226os-trivy"
mkdir -p "$cache"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

# Only the ignore file is mounted (never the whole repository, which may hold local env files).
mount_args=()
ignore_args=()
if [ -f .trivyignore ]; then
  mount_args=(-v "$root/.trivyignore:/work/.trivyignore:ro")
  ignore_args=(--ignorefile /work/.trivyignore)
fi

fail=0
for name in 226os-api 226os-migrate 226os-web; do
  img="$name:$tag"
  docker image inspect "$img" >/dev/null 2>&1 || { printf 'image-scan: BLOCKED: %s not built (run make build)\n' "$img" >&2; exit 1; }
  docker save --output "$tmp/$name.tar" "$img"
  printf '\n== trivy %s ==\n' "$img"
  rc=0
  docker run --rm --read-only --tmpfs /tmp \
    --user "$(id -u):$(id -g)" -e HOME=/tmp -e TRIVY_CACHE_DIR=/cache \
    -v "$cache:/cache" \
    -v "$tmp:/scan:ro" \
    ${mount_args[@]+"${mount_args[@]}"} \
    "$TRIVY_IMAGE" image --input "/scan/$name.tar" --scanners vuln \
    --severity HIGH,CRITICAL --ignore-unfixed --exit-code 1 --no-progress \
    ${ignore_args[@]+"${ignore_args[@]}"} --format table 2>&1 | tee "$reports/$name.txt" || rc=$?
  # tee masks docker's status unless pipefail reports it (set above), so $rc is trivy's.
  if [ "$rc" -eq 0 ]; then
    printf 'image-scan: %s: no fixable HIGH/CRITICAL vulnerabilities\n' "$img"
  else
    printf 'image-scan: FAIL: %s (trivy exit %d)\n' "$img" "$rc" >&2
    fail=1
  fi
done
[ "$fail" -eq 0 ] || exit 1
echo "image-scan: PASS"
