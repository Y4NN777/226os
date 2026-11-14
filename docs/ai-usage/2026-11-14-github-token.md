## S5 optional GitHub token for the repository check

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (go-implementer subagent)"
purpose: "An optional GITHUB_TOKEN so project registration stops sharing the unauthenticated 60 requests/hour/IP GitHub quota, with redaction, a fixed host and a one-time unauthenticated retry when the token is rejected."
repository_or_source_context: "branch feature/build/github-token; TODO S5 choices (no token at first); K10 anti-SSRF rules"
files_or_sections_affected: ["src/api/internal/build/github/**", "src/api/internal/platform/config/**", "src/api/internal/app/**", "src/api/internal/build/build_test.go", "compose.yaml", ".env.example", "scripts/compose.sh", "scripts/tests/compose-wrapper_test.sh", "secrets/staging-test.enc.env", "secrets/production-live.enc.env", "README.md", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["NICE-DEV226 (S5 reviewer, pending)", "Y4NN777 (K10 co-review)"]
tests_or_evidence: ["RED: compile failures before New(token, logger) and LoadGitHubToken existed", "GREEN: TestRepoSendsTheTokenOnlyWhenSet, TestRepoRetriesOnceWithoutTheTokenOn401, TestClientFormattingHidesTheToken, TestLoadGitHubTokenValues; lint-go, race unit, pgtest Build/Project exit 0; Go coverage 100.0 % (6271/6271)"]
open_risks: ["not tested against the real GitHub API with a token", "a revoked token silently falls back to the shared 60/h quota (warn log github_token_rejected)", "the token is tied to the owner's account until a GitHub App replaces it"]
```
