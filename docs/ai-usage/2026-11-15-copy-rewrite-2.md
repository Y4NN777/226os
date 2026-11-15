## Shorter UI copy, part 2 (shared, solve, build, discover, connect, profile)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner confirmed the style guide and its decisions. The AI applied the rewritten strings of the shared i18n catalogue and of the solve, build, discover, connect and profile catalogues (FR and EN), deleted the explanatory ledes, not-found bodies, Connect spaces teasers and dead keys with their renders, moved « réglages » to « paramètres », and updated the expected strings of the tests."
repository_or_source_context: "branch feature/web/copy-rewrite-2, cut from feature/web/prototype-fidelity; S7 web pages; owner copy brief of 2026-11-15"
files_or_sections_affected: ["src/web/src/i18n/messages.ts", "src/web/src/features/solve/", "src/web/src/features/build/messages.ts", "src/web/src/features/discover/", "src/web/src/features/connect/", "src/web/src/features/profile/messages.ts", "src/web/src/features/settings/", "src/web/src/features/NotFound.tsx", "TODO.md", "docs/ai-usage/2026-11-15-copy-rewrite-2.md"]
third_party_sources_consulted: ["none beyond the style guide approved by the owner"]
license_review: "NOT APPLICABLE: original text, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["only expected strings changed in tests; the one check of the deleted sign-in note on the question page was removed, not replaced", "scoped queries where two controls now share a name: the Enregistrer buttons of the settings page (by section) and the Effacer les filtres buttons of the build empty state", "pnpm typecheck, lint, format:check and test --maxWorkers=1 results are in the PR"]
open_risks: ["copy is not reviewed by a native French editor", "the deleted not-found bodies leave the title and the back link alone on those pages", "shell.search is « Rechercher sur le site » / « Search the site », not the table value, so that the top bar, the home feed search and the learn search are search landmarks with different names", "an accepted answer now carries the pill « Résolue » (solve.pill.resolved), the same string as a resolved question"]
```
