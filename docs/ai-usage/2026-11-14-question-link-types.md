## Contract: link types narrowed to the stored pairs

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session, no subagent)"
purpose: "Owner decision on S3: the contract offered question and series link targets that content_relations cannot store (422 at runtime). QuestionCreate and QuestionUpdate now take QuestionLinkRef (resource, project, article) and LinkRef drops series; the generated Go and TypeScript types follow."
repository_or_source_context: "branch fix/solve/question-link-types; db/migrations/00009_content_relations.sql (content_relations_pair_ck); TODO S3 and S4 deviations"
files_or_sections_affected: ["api/openapi.yaml", "src/api/internal/platform/apigen/apigen.gen.go (generated)", "src/api/internal/platform/apispec/openapi.gen.yaml (generated)", "src/api/internal/platform/apispec/apispec_test.go", "src/web/src/api/schema.d.ts (generated)", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (contract reviewer, R7, pending)"]
tests_or_evidence: ["RED: TestLinkTypesMatchTheStoredPairs failed on develop's embedded spec (LinkRef items, series present)", "GREEN: the same test and TestEmbeddedSpecMatchesContract pass after scripts/check-generated.sh generate"]
open_risks: ["a client that sent series or a question-to-question link now fails schema validation (422) instead of the service's 422; no web code sends them"]
```
