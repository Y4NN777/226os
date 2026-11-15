## Third-party licences verified instead of accepted unverified

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (owner's lead session, no subagent)"
purpose: "#87 had closed the licences as accepted without verification. The AI read each licence from its upstream source on 2026-11-13: the GitHub licence API of each repository, and the licence file itself where the API answered NOASSERTION (PostgreSQL, Node.js, goose). This covered the container images, the Go tool entries, go.yaml.in/yaml/v3 (Apache-2.0 with MIT libyaml portions), the CI and Make tools, claude-code-action and the 431 npm development dependencies (none strong copyleft). It wrote THIRD_PARTY_NOTICES.md with a Verification section stating the method and its limits. golangci-lint and ShellCheck (GPL-3.0) and Semgrep (LGPL-2.1) are recorded as build and CI tools, not distributed or linked. The AI-generated brand assets stay an owner decision."
repository_or_source_context: "branch chore/docs/verify-licences, PR #88 (merged 2026-11-13)"
files_or_sections_affected: ["THIRD_PARTY_NOTICES.md", "TODO.md"]
third_party_sources_consulted: ["GitHub licence API of each upstream repository", "upstream licence files of PostgreSQL, Node.js and goose"]
license_review: "VERIFIED: read from the upstream licence API or licence file (method in THIRD_PARTY_NOTICES.md, Verification)"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #88)"]
tests_or_evidence: ["THIRD_PARTY_NOTICES.md Verification section; runtime Go and npm dependencies still come from make licenses"]
open_risks: ["record written on 2026-11-15, after the merge: none was filed at the time", "the brand assets' licence stays an owner decision"]
```
