# Owner preferences (binding for every human and AI contributor)

**Owner:** `Y4NN777`. These are standing decisions of the owner, gathered from the working sessions of 11 to 14 Nov 2026. They apply to every session, local or in CI (the review, the deep review and the automatic fixer read this file). `AGENTS.md`, `CLAUDE.md` and the ADRs still rule; this file only adds the owner's working preferences.

## Work and process

- **Automate the review and merge flow.** A pull request into `develop` merges when CI is green and the AI review has no open finding. Notes never block a merge: they are filed on merge as a "PR<N> notes" issue and ranked by the review batch (`scripts/findings.py batch`), fixed by one pull request per batch. Do not ask the owner to triage notes or approve merges that meet the rule.
- **Ask the owner only for real decisions:** scope, contract changes (R7), migrations (R6), ADRs, security exceptions, secrets and settings. Everything else is decided from the docs and done.
- **Work to the end.** Stop only for a real blocker; never re-ask something already decided.
- **Fewer pull requests and issues.** Bundle related changes in one pull request (each costs a CI wait); a small edit never gets its own pull request. Follow-ups of one piece of work go in one issue with a checklist.
- **Analysis means analysis.** When asked to analyse or propose, give findings and two or three options with a recommendation, then stop; build only after the owner chooses.
- **"Fix before merge" is binding.** If a later short message is ambiguous ("close", "no time"), ask one short question before merging.
- **No AI attribution.** No `Co-authored-by` for an AI and no "Generated with" line in commits or pull request bodies; AI use is recorded as one file in `docs/ai-usage/`.

## Product copy (French)

- Natural, idiomatic developer French; never translated-sounding or "AI French" (no calques such as « Voir en tant que », no marketing filler, no exclamation marks).
- Keep the jargon developers use (webhook, API, SDK, pull request, open source); use French where natural French exists (Résoudre, Découvrir, Construire, Communauté, Mentorat).
- No inclusive-writing median dots (« Candidat·e »): rephrase neutrally.
- French typography: a non-breaking space before `:` `;` `?` `!` and inside « ».
- One form of address across the whole product.

## Interface (minimal, premium)

- Design what a real user needs and nothing more; no prototype scaffolding in view (demo switches, "simulé" badges, explanation callouts) except behind one discreet toggle.
- Icon buttons with an `aria-label` for small actions (show password, copy, vote) instead of text buttons.
- Language detected from the browser; no prominent FR/EN switch in the header (a discreet override in the account menu is enough).
- Few borders: keep them for inputs and tables; use spacing or background elsewhere.
- Never vertical accent bars or left-border stripes for active, accepted or callout states.
- Sign-in and sign-up are standalone pages without app chrome. The product has a public landing page.
- Profile fields stay off the account forms; ask for a profile item just in time, where its value is concrete. The member's own profile page carries a completion checklist (progress, one-line value per item, completion in place, gone at 100 %); never on the home page.
- Base design changes on researched expert references, not guesses.
