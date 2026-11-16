## Review fix of batch 352: mobile search close label, S9 tracker line

```yaml
date: "2026-11-16"
contributor: "Y4NN777"
tool_or_model: "Claude Code in CI (review-batch), Claude Opus 5.5 orchestrator, web-implementer agent"
purpose: "Work the 15 items of review batch 352. Two fixes: the close button of the phone search overlay had the search label as its accessible name (issue 348, mobile-search-close-label), and the S9 production-enablement line of TODO.md sat in the S7 block (notes todo-line-in-s7-block and todo-placement). The other items were settled without a change: pipeline and ADR paths held by the owner, owner-decided deploy and quota notes, and search notes that are not defects."
repository_or_source_context: "branch fix/review/batch-2026-11-16, batch issue 352"
files_or_sections_affected: ["src/web/src/app/Shell.tsx", "src/web/src/i18n/messages.ts", "src/web/src/app/Shell.mobileSearchClose.test.tsx", "TODO.md", "docs/ai-usage/2026-11-16-review-fix-batch.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: no dependency, no third-party code"
closure: "open"
human_reviewers: []
tests_or_evidence: ["RED: Shell.mobileSearchClose.test.tsx failed, no button named Fermer la recherche in the open overlay", "GREEN: pnpm --filter web run test src/app/Shell, 2 files and 46 tests passed; typecheck and lint exit 0", "RED: the chore/ops/enable-ai-production line of TODO.md was at line 388, outside the S9 block (424 to 462)", "GREEN: the same line is at 459, inside the S9 block (424 to 463)", "make lint: FAIL at ShellCheck only, no Docker daemon in the fixer; OpenAPI, migrations, ESLint and Prettier passed before it; the tracker check is NOT RUN (CI lint-scripts runs it)", "make test-unit: FAIL in src/api/cmd/api only, the runner's GITHUB_TOKEN breaks that config test; no Go file changed, every other Go package passed", "pnpm --filter web run test (the web part of test-unit): 66 files, 820 tests passed", "make coverage, make test-integration-go, make security: NOT RUN (web-only change, no Go, database or identity code touched)"]
open_risks: []
```
