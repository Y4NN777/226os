## Two slow Build tests made fast

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent)"
purpose: "Two Build tests timed out under full-suite load in CI. The AI measured them, found that the cost was the accessible-name role query over the whole page after opening a Selector (about 600 ms per pick, more with 54 countries), and added a test helper that opens the Selector by keyboard and finds the option by text inside the listbox. The create-form test was also split in two (stage and licence, then country), with the same assertions, and its 30 s timeout was removed. No assertion was removed or loosened and no timeout was raised."
repository_or_source_context: "branch fix/web/slow-build-tests, from develop 1045d41"
files_or_sections_affected: ["src/web/src/test/select.ts", "src/web/src/features/build/BuildPage.data.test.tsx", "src/web/src/features/build/ProjectForm.data.test.tsx"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: test code only, no dependency"
closure: "open"
human_reviewers: []
tests_or_evidence: ["vitest --maxWorkers=1 on both files and on src/features/build: PASS; typecheck, lint, format:check PASS"]
open_risks: ["the helper looks the option up by text, not by accessible name"]
```
