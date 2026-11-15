## Landing v2: landing-only header and footer, Africa globe hero, photo sections, scroll reveals (Geist, landing only)

```yaml
date: "2026-11-15"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent) led by Claude Opus 5.5"
purpose: "Build version 2 of the public landing after the owner's feedback: landing-only header and footer, full-viewport hero with one call to action and a large Africa-centred dotted globe with arcs between ten African cities, five distinct sections, Unsplash photos and decorative portrait circles, a reusable data-reveal scroll system, count-up of real totals, demo content excluded, FR and EN in one language each."
repository_or_source_context: "branch feature/landing/v2 (follows the merged PR #307); src/web/src/features/landing, src/web/src/app/Shell.tsx; captures of the dev server reading GET /api/v1 of production anonymously"
files_or_sections_affected: ["src/web/src/features/landing/**", "src/web/src/app/Shell.tsx", "src/web/src/app/*.test.tsx", "src/web/src/components/icons.tsx", "src/web/src/styles/shell.css", "src/web/public/landing/**", "TODO.md"]
third_party_sources_consulted: ["https://github.com/vercel/geist-font v1.9.2 (Geist[wght].woff2, OFL-1.1, licence file copied)", "https://github.com/nvkelso/natural-earth-vector ne_110m_land.geojson (public domain), used offline only to compute dot coordinates", "Unsplash photos listed with photographer and page in src/web/public/landing/CREDITS.md"]
license_review: "Geist: SIL OFL-1.1; Natural Earth: public domain; photos: Unsplash License (free use, no attribution required), pages checked on 2026-11-15; no new npm dependency. Owner to confirm the photo licence."
closure: "open: waits for the owner's decision and review"
human_reviewers: []
tests_or_evidence: ["RED: LandingPage.test.tsx rewritten first, 20 of 28 tests failed against the v1 page", "GREEN: pnpm exec vitest run src/features/landing (28 PASS); whole web suite, typecheck, lint, format:check see the final report", "Playwright captures 1440, 768 and 390 px, dark and light, FR and EN, hero and full page, no horizontal scroll", "pixel-sampled contrast audit of every text node, worst real case 5.02:1"]
open_risks: ["owner decision: Geist vs DESIGN.md section 4 (system fonts)", "owner decision: « Membres » on the landing vs « Connect » in the app", "owner decision: Unsplash photos and licence", "the member avatar chip was dropped, so the landing no longer depends on #306", "totals of a list longer than one page include demo items (the server total is used)", "no axe run (@axe-core/playwright not installed), no screen reader run", "the compact menu starts below 1024 px, not 768 px, because the full bar does not fit at 768 px"]
```
