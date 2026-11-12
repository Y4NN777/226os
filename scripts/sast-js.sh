#!/usr/bin/env bash
# JS/TS SAST (G4, ADR-0006 `security`): Semgrep over src/web/src with the repository's own rules
# (scripts/semgrep/web-rules.yml). Fails on any finding (ERROR or WARNING, --error) and on any scanner
# failure; also fails when nothing was scanned.
#
# Tool: semgrep/semgrep pinned by image digest (below), run with --network none, a read-only root
# filesystem, the caller's uid, metrics off and the version check disabled. Rules are local: no registry
# download, no drift. The generated src/web/src/api/schema.d.ts is excluded.
# Usage: scripts/sast-js.sh [target-dir]      (default: src/web/src; the dir argument is for tests)
set -euo pipefail

SEMGREP_IMAGE="${SEMGREP_IMAGE:-semgrep/semgrep@sha256:67518c174185a2043fdbed16c70631b79209499b3671178e7194379b48295a0d}" # semgrep/semgrep:1.185.0

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"
target="${1:-src/web/src}"
[ -d "$target" ] || { printf 'sast-js: BLOCKED: target directory not found: %s\n' "$target" >&2; exit 1; }
target="$(cd "$target" && pwd)"
rules="$root/scripts/semgrep/web-rules.yml"
[ -f "$rules" ] || { printf 'sast-js: BLOCKED: %s missing\n' "$rules" >&2; exit 1; }

reports="${REPORTS_DIR:-reports}/semgrep"
mkdir -p "$reports"
reports="$(cd "$reports" && pwd)"
rm -f "$reports/semgrep.json"

rc=0
docker run --rm --network none --read-only --tmpfs /tmp \
  --user "$(id -u):$(id -g)" -e HOME=/tmp -e SEMGREP_SEND_METRICS=off \
  -v "$target:/src:ro" -v "$rules:/rules/web-rules.yml:ro" -v "$reports:/out" -w /src \
  "$SEMGREP_IMAGE" semgrep scan --config /rules/web-rules.yml --metrics=off --disable-version-check \
  --error --exclude schema.d.ts --json-output /out/semgrep.json /src || rc=$?

if [ "$rc" -ne 0 ]; then
  printf 'sast-js: FAIL (semgrep exit %d: findings or scanner failure)\n' "$rc" >&2
  exit 1
fi
python3 -I - "$reports/semgrep.json" <<'PY'
import json, sys
d = json.load(open(sys.argv[1]))
scanned = d.get("paths", {}).get("scanned", [])
if d.get("errors"):
    print("sast-js: FAIL (scanner reported errors)", file=sys.stderr); sys.exit(1)
if not scanned:
    print("sast-js: FAIL (no file was scanned; refusing a silent pass)", file=sys.stderr); sys.exit(1)
print(f"sast-js: PASS ({len(scanned)} files scanned, 0 findings)")
PY
