## Harness tests in CI

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (ci-implementer subagent)"
purpose: "Fix the agent inventory test for the new ci-implementer agent, and run the harness tests (.claude/hooks/tests) in CI through a new make target test-harness, a step of make quality and of the CI group harness."
repository_or_source_context: "branch chore/claude/ci-implementer-inventory; owner-approved pipeline task; ADR-0010 (CI groups); mk/40-quality.mk"
files_or_sections_affected: [".claude/hooks/tests/test_domain_skills.py", "mk/20-gates.mk", "mk/40-quality.mk", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["Y4NN777 (pipeline owner, pending)"]
tests_or_evidence: ["RED: python3 -m unittest discover -s .claude/hooks/tests failed, ci-implementer missing from the expected set", "GREEN: same command exit 0 (28 tests)"]
open_risks: ["the harness CI group is NOT RUN in CI until this pull request runs", "scripts/tests/make_test.sh was not touched: it pins the repo group steps, so the step lives in its own group"]
```
