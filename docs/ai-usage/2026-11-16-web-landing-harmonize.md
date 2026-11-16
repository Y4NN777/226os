## Inner app harmonized with the landing

```yaml
date: "2026-11-16"
contributor: "Y4NN777"
tool_or_model: "OpenAI Codex CLI (gpt-5.6-sol, nine runs: design system, seven areas in parallel, one polish pass), driven by Claude Code (Claude Opus 5.5, lead session: briefs, visual review of the captures, integration)"
purpose: "The AI updated DESIGN.md to 1.1 with the landing as the product reference, moved the landing's scales to app-wide tokens (styles/tokens.css), applied them to the shell, the shared components and every area of the signed-in app, unified the form controls, made the mobile header one row, fixed the three review notes of #332, and refreshed the README and quality captures."
repository_or_source_context: "branch feature/web/landing-harmonize, from develop after #332"
files_or_sections_affected: ["DESIGN.md", "src/web/src/styles/**", "src/web/src/theme/theme226os.ts", "src/web/src/app/**", "src/web/src/components/**", "src/web/src/features/**", "README.md", "docs/quality/ASTRYX_REPORT.md", "docs/quality/evidence/harmonize/2026-11-16/", "TODO.md"]
third_party_sources_consulted: []
license_review: "no new dependency; Geist (OFL-1.1) was already self-hosted by #332"
closure: "open"
human_reviewers: []
tests_or_evidence: ["pnpm --filter web run typecheck and lint: exit 0", "vitest: 63 files, 809 tests passed", "check-docs PASS", "48 captures (12 pages, light and dark, 1280 and 375 px) against the public production data, reviewed page by page by the lead session"]
open_risks: ["no pixel diff and no signed-in captures (member home, own profile, settings forms)", "the README captures show the public production data of 16 Nov 2026"]
```
