## Shorter UI copy, part 1 (landing, home, learn, search, identity)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked for shorter, less AI-sounding UI copy and confirmed a style guide. The AI applied the rewritten strings of the landing, home, learn, search and identity catalogues (FR and EN), deleted the explanatory ledes and notes with their renders (HomePage section sub, landing pillar and catalogue ledes, learn and write ledes, invitation body, not-found bodies, sign-in and sign-up card ledes, welcome intro), shortened the learn empty states added by the filters pass, and updated the expected strings of the tests."
repository_or_source_context: "branch feature/web/copy-rewrite, cut from feature/web/prototype-fidelity; S7 web pages; owner copy brief of 2026-11-15"
files_or_sections_affected: ["src/web/src/features/landing/", "src/web/src/features/home/", "src/web/src/features/learn/", "src/web/src/features/search/", "src/web/src/features/identity/", "src/web/src/features/solve/SolvePage.test.tsx", "src/web/src/features/solve/AskAboutLesson.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-copy-rewrite.md"]
third_party_sources_consulted: ["none beyond the style guide approved by the owner"]
license_review: "NOT APPLICABLE: original text, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["only expected strings changed in tests; the two checks of the deleted sign-in and sign-up ledes were removed, not replaced", "LearnPage 'offers writing to a member': the header action and the invitation now share the label « Write an article », so the single findByRole became findAllByRole with length 2 and an href check on each (stricter)", "solve tests SolvePage and AskAboutLesson assert strings of the search and learn catalogues and were updated", "pnpm typecheck, lint and format:check exit 0; pnpm test --maxWorkers=1 leaves only the known Shell.test.tsx banner failure (see the PR)"]
open_risks: ["copy is not reviewed by a native French editor", "catalogues outside these five (src/i18n/messages.ts and others) are part 2", "the assistant notice no longer says that only the question is sent", "learn empty-state body « Retirez un filtre. » sits above the « Effacer les filtres » button"]
```
