#!/usr/bin/env bash
# Static test of .claude/commands/review-fix.md: the CI fixer is an orchestrator that runs the project agents in parallel
# (at most 3), and only the orchestrator writes TODO.md and the AI-use record and runs the gates.
# Usage: scripts/tests/review-fix-command_test.sh
set -euo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
root="$(cd "$here/../.." && pwd)"
# shellcheck source=scripts/tests/lib.sh
source "$here/lib.sh"
doc="$root/.claude/commands/review-fix.md"

flat="$(tr -s '[:space:]' ' ' <"$doc")"
has() { grep -Eqi -- "$1" <<<"$flat"; }

has 'at most 3 (agents )?(at once|in parallel)|at most three' && ok "review-fix.md states the parallel limit: at most 3 agents at once" || ko "parallel limit"
has 'in parallel' && ok "review-fix.md launches the agents in parallel" || ko "parallel"
has 'only the orchestrator (writes|runs)[^.]*(TODO\.md)[^.]*(docs/ai-usage|ai-usage)[^.]*(gates)|only the orchestrator writes `?TODO\.md`? and the AI-use record and runs the gates' && ok "review-fix.md: only the orchestrator writes TODO.md / ai-usage and runs the gates" || ko "orchestrator rule"
has 'no file (is )?shared between two groups' && ok "review-fix.md: groups by owning agent and by files, no file shared between two groups" || ko "no shared file"
has 'never `?make`? gates|never run `?make`?' && ok "review-fix.md: an agent never runs make gates" || ko "agents no make"
has 'failed group[^.]*OWNER[^.]*other groups still go out' && ok "review-fix.md: a failed group is OWNER and the other groups still go out" || ko "failed group"
has 'single-item run[^.]*through its owning agent' && ok "review-fix.md: a single item may skip the parallel step but goes through its owning agent" || ko "single item"
for a in go-implementer web-implementer postgres-expert ci-implementer; do
  has "$a" && [ -f "$root/.claude/agents/$a.md" ] && ok "review-fix.md names the agent $a, which exists" || ko "agent $a"
done

finish
