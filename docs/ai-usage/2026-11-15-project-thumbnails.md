## Project thumbnails and the card view of the Build, Discover and Learn lists

```yaml
date: "2026-11-15"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Card view by default on Build, Discover and Apprendre with a Cartes / Liste switch (view search param, remembered per browser), and a thumbnail for each project in Build (list rows and project page) from the repository's GitHub social preview, with the monogram tile as fallback; no contract change, no migration, no server storage."
repository_or_source_context: "branch feature/build/web-project-thumbnails; TODO.md S5 block; DESIGN.md (iconography, low bandwidth); src/web/src/components/Avatar.tsx"
files_or_sections_affected: ["src/web/src/features/build/ProjectThumb.tsx", "src/web/src/features/build/ProjectThumb.test.tsx", "src/web/src/features/build/BuildPage.tsx", "src/web/src/features/build/ProjectPage.tsx", "src/web/src/features/build/build.css", "src/web/nginx/security-headers.conf", "src/web/src/components/ViewToggle.tsx", "src/web/src/components/CardGrid.tsx", "src/web/src/components/useListView.ts", "src/web/src/components/ViewToggle.test.tsx", "src/web/src/components/CardGrid.test.tsx", "src/web/src/components/useListView.test.tsx", "src/web/src/styles/ui.css", "src/web/src/i18n/messages.ts", "src/web/src/features/build/routeSearch.ts", "src/web/src/features/build/BuildPage.view.test.tsx", "src/web/src/features/discover/DiscoverPage.tsx", "src/web/src/features/discover/routeSearch.ts", "src/web/src/features/discover/DiscoverPage.view.test.tsx", "src/web/src/features/learn/LearnPage.tsx", "src/web/src/features/learn/present.tsx", "src/web/src/features/learn/routeSearch.ts", "src/web/src/features/learn/LearnPage.view.test.tsx", "TODO.md"]
third_party_sources_consulted: ["GitHub social preview images (opengraph.githubassets.com)"]
license_review: "NOT APPLICABLE: no new dependency; images are loaded from GitHub at view time, not copied or redistributed"
closure: "open: waits for the owner's decision and review"
human_reviewers: []
tests_or_evidence: ["RED: vitest run src/features/build/ProjectThumb.test.tsx fails (module ./ProjectThumb not found)", "GREEN: vitest run src/features/build: 7 files, 107 tests pass", "RED (card view): vitest run on the new view tests fails (modules ./ViewToggle, ./CardGrid, ./useListView not found; no Cartes / Liste buttons on the three pages)", "GREEN (card view): the same files pass, then the build, discover, learn and components suites", "pnpm run typecheck, pnpm run lint, prettier: PASS"]
open_risks: ["privacy: GitHub receives the viewer's IP for each preview (no referrer sent)", "bandwidth: one preview per project card or row, lazy-loaded", "CSP img-src now allows opengraph.githubassets.com", "Learn cards keep the existing icon tile (KindThumb), not the monogram: Learn rows have no monogram", "reverses the earlier no-remote-picture choice for projects: owner decision"]
```
