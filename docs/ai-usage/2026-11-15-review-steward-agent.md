## Review pipeline steward agent and its skill

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session)"
purpose: "The AI wrote a project agent dedicated to the maintenance and evolution of the AI review pipeline (.claude/agents/review-steward.md) and its preloaded skill (.claude/skills/review-pipeline/SKILL.md: map of the stages, ledger fields, invariants, failure history, test patterns, ADR-0015 roadmap), routed the pipeline files to that skill in the skill map, and pointed ci-implementer and CLAUDE.md to the new agent."
repository_or_source_context: "branch feature/harness/review-steward, from develop"
files_or_sections_affected: [".claude/agents/review-steward.md", ".claude/skills/review-pipeline/SKILL.md", ".claude/hooks/skill-map.json", ".claude/hooks/tests/test_domain_skills.py", ".claude/agents/ci-implementer.md", "CLAUDE.md"]
third_party_sources_consulted: []
license_review: "no new dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["python3 -I -m unittest discover -s .claude/hooks/tests: OK (28 tests), RED first on the new skill-map cases", "check-docs PASS"]
open_risks: ["the skill describes ADR-0015 phase 0 rules from PR #309; it says to check that #309 is merged before relying on them"]
```
