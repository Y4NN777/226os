## Project thumbnails from the GitHub social preview

```yaml
date: "2026-11-15"
contributor: "PANK4SS"
tool_or_model: "Claude Code, Claude Opus 5.5"
purpose: "Show a thumbnail for each project in Build (list rows and project page) from the repository's GitHub social preview, with the monogram tile as fallback; no contract change, no migration, no storage."
repository_or_source_context: "branch feature/build/web-project-thumbnails; TODO.md S5 block; DESIGN.md (iconography, low bandwidth); src/web/src/components/Avatar.tsx"
files_or_sections_affected: ["src/web/src/features/build/ProjectThumb.tsx", "src/web/src/features/build/ProjectThumb.test.tsx", "src/web/src/features/build/BuildPage.tsx", "src/web/src/features/build/ProjectPage.tsx", "src/web/src/features/build/build.css", "src/web/nginx/security-headers.conf", "TODO.md"]
third_party_sources_consulted: ["GitHub social preview images (opengraph.githubassets.com)"]
license_review: "NOT APPLICABLE: no new dependency; images are loaded from GitHub at view time, not copied or redistributed"
closure: "open: waits for the owner's decision and review"
human_reviewers: []
tests_or_evidence: ["RED: vitest run src/features/build/ProjectThumb.test.tsx fails (module ./ProjectThumb not found)", "GREEN: vitest run src/features/build: 7 files, 107 tests pass", "pnpm run typecheck, pnpm run lint, prettier: PASS"]
open_risks: ["privacy: GitHub receives the viewer's IP for each preview (no referrer sent)", "bandwidth: one preview per project row, lazy-loaded", "CSP img-src now allows opengraph.githubassets.com", "reverses the earlier no-remote-picture choice for projects: owner decision"]
```
