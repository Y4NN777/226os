## S7 frontend PR 3: landing, member home and parametres

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (integrator, parametres) and Claude Sonnet 5.5 (landing and home lanes in separate worktrees)"
purpose: "S7 PR 3: public landing page and member home on a session-aware index route, profile visibility setting on parametres (C14b), settings moved to /parametres."
repository_or_source_context: "branch feature/platform/web-home-settings; lane branches feature/platform/web-landing and feature/platform/web-home-feed; prototypes/html for copy"
files_or_sections_affected: ["src/web/src/features/{landing,home,settings}/**", "src/web/src/app/**", "src/web/src/i18n/**", "TODO.md", "AI_USAGE.md"]
third_party_sources_consulted: ["installed @astryxdesign/cli 0.7.14 component docs"]
license_review: "NOT APPLICABLE: original code and text, no new dependency"
closure: "open"
human_reviewers: ["PANK4SS (pending)"]
tests_or_evidence: ["RED then GREEN (ProfileVisibility.test.tsx, LandingPage.test.tsx, HomePage.test.tsx, translate.test.ts key-collision check)", "Real-browser journeys on make up-local-dev with Mailpit: sign-up, verify, sign-in, sign-out, wrong password, reset, sign-in with the new password; axe on 13 pages; no horizontal scroll at 360/768/1024/1440 and 200% zoom (56/56)"]
open_risks: ["Demo examples on the member home stay until listLearnArticles, listQuestions and listProjects are live.", "The local display-name field on parametres is a prototype leftover kept for now (owner decision)."]
```

