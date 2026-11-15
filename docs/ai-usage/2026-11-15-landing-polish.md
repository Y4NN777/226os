## Landing polish: product screenshots, gentle motion, clearer hero

```yaml
date: "2026-11-15"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent) led by Claude Opus 5.5"
purpose: "Polish the public landing page: four real screenshots of the live product (live content only, no demo item), a small IntersectionObserver reveal with reduced-motion and no-observer fallbacks, hover states, FR/EN captions and alt texts."
repository_or_source_context: "branch feature/landing/polish; TODO.md S7 block; src/web/src/features/landing; captures of https://226os.tech taken anonymously with Playwright (read-only GET)"
files_or_sections_affected: ["src/web/src/features/landing/**", "src/web/public/landing/**", "TODO.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: screenshots of the project's own product, no stock or third-party image, no new dependency"
closure: "open: waits for the owner's decision and review"
human_reviewers: []
tests_or_evidence: ["RED: LandingPage.test.tsx, 4 new tests failed (no screenshots region, no data-reveal state)", "GREEN: pnpm exec vitest run src/features/landing src/app src/i18n --maxWorkers=3 (94 tests PASS)", "pnpm run typecheck and pnpm run lint PASS", "layout check at 360, 768, 1024, 1440 px: no horizontal page scroll", "each image viewed: no Démo badge and no demo- author"]
open_risks: ["the latest-resources section still lists the production catalogue, which currently holds only demo entries (badged)", "no Découvrir screenshot: every catalogue entry in production is a demo fixture", "screenshots will age as the product changes", "no axe or contrast run"]
```
