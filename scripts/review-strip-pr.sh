#!/usr/bin/env bash
# Strip the pull request tree the deep review reads (stage 2, .github/workflows/code-review.yml job `deep`) of every
# file Claude Code would load as instructions or settings when it reads inside it: CLAUDE.md, CLAUDE.local.md,
# AGENTS.md, .mcp.json and .claude/ directories, at any depth. The code under review stays. The deep job runs this right
# after checking the pull request head out under pr/ (actions/checkout cannot put it outside the workspace).
# Usage: scripts/review-strip-pr.sh DIR
set -euo pipefail

dir="${1:-}"
[ -n "$dir" ] && [ -d "$dir" ] || { echo "review-strip-pr: BLOCKED: no directory '$dir'" >&2; exit 2; }
find "$dir" -depth \( -name .git -prune \) -o \( -type d -name .claude -exec rm -rf {} + \)
find "$dir" \( -name .git -prune \) -o \( \( -name CLAUDE.md -o -name CLAUDE.local.md -o -name AGENTS.md -o -name .mcp.json \) -exec rm -rf {} + \)
left="$(find "$dir" \( -name .git -prune \) -o \( -name CLAUDE.md -o -name CLAUDE.local.md -o -name AGENTS.md -o -name .mcp.json -o -name .claude \) -print)"
[ -z "$left" ] || { echo "review-strip-pr: FAIL: still present: $left" >&2; exit 1; }
echo "review-strip-pr: OK (no instruction or settings file left under $dir)"
