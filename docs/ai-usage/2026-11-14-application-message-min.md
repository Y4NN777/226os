## Contract: an application message is not empty (#135)

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session, no subagent)"
purpose: "Owner decision on #135: the contract allowed an empty ApplicationCreate.message while applications_message_ck and the service required 1 to 2000 characters. The contract now says minLength 1; a test pins it."
repository_or_source_context: "branch fix/build/application-message-min; issue #135 (PR132 notes); review batch #157 item 10"
files_or_sections_affected: ["api/openapi.yaml", "src/api/internal/platform/apispec/openapi.gen.yaml (generated)", "src/api/internal/platform/apispec/apispec_test.go", "src/api/internal/build/service/applications.go (comment)", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original change, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (contract co-reviewer, R7, pending)"]
tests_or_evidence: ["RED: TestApplicationMessageIsNotEmpty failed on minLength 0", "GREEN: apispec, build packages and TestBuildApplication* (real PostgreSQL) pass; make lint-contract and docs-check exit 0"]
open_risks: ["none: the server already refused an empty message with 422; only the documented contract changes"]
```
