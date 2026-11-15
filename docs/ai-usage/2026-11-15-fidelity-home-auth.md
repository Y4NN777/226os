## Page fidelity pass: landing, member home, search and account pages

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked that every page match its HTML prototype visually. For `/` (landing and member home), `/recherche` and the account pages the AI rebuilt the layout after prototypes/html (index, accueil, recherche, connexion, inscription, mot-de-passe-oublie, reinitialiser, verifier-email, bienvenue): dense list rows with lead, title link, meta line and state pills, section heads with icon and see-all link, hero, icon pillar cards and catalogue cards, search groups, the account card head with title and lead, the forgot-password link on the password label line and welcome choice tiles. Only contract fields are shown; behaviour, routes and tests are kept. It wrote RED tests for the new visible elements first, then the code, and added one TODO.md line."
repository_or_source_context: "branch feature/web/fidelity-home-auth; S7 web pages; prototypes/html (read only)"
files_or_sections_affected: ["src/web/src/features/home/**", "src/web/src/features/landing/**", "src/web/src/features/search/SearchPage.tsx", "src/web/src/features/search/SearchPage.test.tsx", "src/web/src/features/search/messages.ts", "src/web/src/features/search/search.css", "src/web/src/features/identity/**", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-home-auth.md"]
third_party_sources_consulted: ["@astryxdesign/core installed type declarations and astryx.css token names, read from node_modules", "prototypes/html icon paths (the project's own prototype, copied into FeedIcon.tsx)"]
license_review: "NOT APPLICABLE: original code and the project's own prototype assets, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["Order: RED tests first (each failed on the previous page), then the pages", "new tests: home rows (pills, counts, see-all links, never Vérifié when unverified), landing catalogue cards, search title and complete-group count, sign-in lead and forgot link placement, sign-up lead, welcome tile lines", "changed tests: sign-in h1 is now « Content de vous revoir » / « Welcome back » (prototype copy; the document title stays « Connexion »)", "captures of prototype and app at 1280 and 375 px, light and dark, in the session scratchpad"]
open_risks: ["the member home was captured with a stubbed session (dev-only Vite middleware, not committed), not a real sign-in", "the logo above the account card and its legal footer are in app/AuthLayout.tsx, outside this branch: the lead applies them", "the local rows, pills and icons duplicate what the shell branch adds in ui/ and components/: the lead swaps them", "no axe or screen-reader run on the new markup"]
```
