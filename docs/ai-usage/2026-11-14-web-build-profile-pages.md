## S7 build, project and profile pages

```yaml
date: "2026-11-14"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) and Claude Sonnet 5.5 (web-implementer and skeptic subagents)"
purpose: "The build, project and profile pages on the merged S5 and S1 APIs: project list and registration, project detail with contribution needs and applications (owner decisions, applicant withdrawal), and the public profile with the completion checklist on the member's own profile only."
repository_or_source_context: "branch feature/build/web-build-profile-pages; issue #109 section 5; TODO S1, S5 choices and deviations, S7 page map; prototypes/html/pages/{build,project,profile}.html"
files_or_sections_affected: ["src/web/src/features/build/**", "src/web/src/features/profile/**", "src/web/src/app/router.tsx", "src/web/src/app/navigation.ts", "src/web/src/app/Shell.tsx", "src/web/src/app/Shell.test.tsx", "src/web/src/i18n/messages.ts", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (S7 reviewer, pending)"]
tests_or_evidence: ["RED: BuildPage 13/13, ProjectPage 21/21, ProfilePage 15/15 failing before the routes existed", "GREEN: lint, format:check, typecheck, coverage exit 0 (387 tests; 95.67/87.95/95.48/96.8)"]
open_risks: ["updateProject has no If-Match in the contract, so concurrent owner edits are last-write-wins", "an applicant can withdraw only right after applying: no operation lists a member's own applications", "axe, visual and responsive checks NOT RUN"]
```
