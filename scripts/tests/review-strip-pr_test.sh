#!/usr/bin/env bash
# Tests for scripts/review-strip-pr.sh: the pull request tree the deep review reads carries no instruction or settings
# file Claude Code would load (CLAUDE.md, CLAUDE.local.md, AGENTS.md, .claude/, .mcp.json), at any depth; code stays.
# Usage: scripts/tests/review-strip-pr_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
tool="$root/scripts/review-strip-pr.sh"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
pr="$tmp/pr"
mkdir -p "$pr/.claude/agents" "$pr/src/api/.claude" "$pr/src/web" "$pr/db"
for f in CLAUDE.md CLAUDE.local.md AGENTS.md .mcp.json .claude/settings.json .claude/agents/skeptic.md src/api/CLAUDE.md src/web/AGENTS.md src/api/.claude/settings.json db/query.sql src/api/main.go; do
  echo x >"$pr/$f"
done

rc=0; out="$("$tool" "$pr" 2>&1)" || rc=$?
left="$(cd "$pr" && find . -name CLAUDE.md -o -name CLAUDE.local.md -o -name AGENTS.md -o -name .mcp.json -o -name .claude | sort)"
[ "$rc" -eq 0 ] && [ -z "$left" ] && ok "every CLAUDE.md, CLAUDE.local.md, AGENTS.md, .mcp.json and .claude/ is removed, at any depth" || ko "strip (rc=$rc out=$out left=$left)"
[ -f "$pr/db/query.sql" ] && [ -f "$pr/src/api/main.go" ] && ok "the code under review stays" || ko "code removed"
rc=0; "$tool" "$tmp/missing" >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] && ok "a missing directory fails (the deep review then does not run)" || ko "missing dir (rc=$rc)"
rc=0; "$tool" >/dev/null 2>&1 || rc=$?
[ "$rc" -ne 0 ] && ok "no argument fails" || ko "no arg (rc=$rc)"

finish
