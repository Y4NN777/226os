## Contract violation codes and pointers in Learn and discover reports

```yaml
date: "2026-11-13"
contributor: "mibienpanjoe"
tool_or_model: "Claude Code, Claude Opus 5.5 (main session, no subagents)"
purpose: "Align the Learn and discover error-report 422 violations with the contract, as #127 did for identity (issue #116 rule): JSON-pointer fields for body members and Violation.code enum values instead of 'invalid'; query parameters keep their names. Also marks S4 DONE in TODO.md (PR #111 merged)."
repository_or_source_context: "branch fix/platform/violation-codes; review of PR #127"
files_or_sections_affected: ["src/api/internal/learn/service/service.go", "src/api/internal/learn/service/service_test.go", "src/api/internal/discover/service/reports.go", "src/api/internal/discover/service/reports_test.go", "src/api/internal/discover/service/service.go", "src/api/internal/platform/pgtest/learn_integration_test.go", "src/api/internal/platform/pgtest/errorreport_integration_test.go", "TODO.md", "docs/ai-usage/2026-11-13-violation-codes.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["NICE-DEV226 (pending)"]
tests_or_evidence: ["RED: TestCreateViolationsUseContractCodesAndPointers and TestCreateErrorReportValidation failed on 'invalid' codes and bare field names", "GREEN: see the PR evidence table"]
open_risks: ["identity's password and new_password violations still use 'invalid' (identity area, flagged for NICE-DEV226)"]
```
