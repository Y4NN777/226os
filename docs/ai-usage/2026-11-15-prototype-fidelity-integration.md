## Prototype fidelity integration: merges, copy rewrite tooling, fidelity report and README screenshots

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (lead session) with web-implementer subagents and two proof:screen-judge reviews"
purpose: "The lead session integrated the page-fidelity branches (shell, rail, harmonize, learn filters, home-auth, leftovers, controls, copy rewrite parts 1 and 2) into feature/web/prototype-fidelity and resolved their conflicts. It also applied the owner-approved copy rewrite by script (only where a string still had its audited value), restored the assistant privacy fact, captured the app next to the prototypes, had an independent reviewer judge them twice, and wrote the fidelity report, the copy style guide and the README screenshots section."
repository_or_source_context: "branch feature/web/prototype-fidelity; prototypes/html/pages; docs/quality/ASTRYX_REPORT.md section 8"
files_or_sections_affected: ["docs/quality/PROTOTYPE_FIDELITY.md", "docs/quality/evidence/fidelity/2026-11-15/", "docs/quality/ASTRYX_REPORT.md", "docs/team/COPY_STYLE.md", "docs/team/OWNER_PREFERENCES.md", "README.md", "TODO.md", "src/web/src/components/{List,PageHeader,StatusBadge,icons}.tsx", "src/web/src/features/{connect,settings,solve,discover,search}/", "src/web/src/i18n/messages.ts"]
third_party_sources_consulted: ["copy research sources listed in docs/team/COPY_STYLE.md (Microsoft fr-FR localization guide, GitHub Primer, Google developer style guide, Atlassian, Mailchimp, OQLF, Mastodon and Discourse FR translations)"]
license_review: "NOT APPLICABLE: original text and code, no dependency added; screenshots are of this project"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["web gates on the merged branch: typecheck, lint, format:check PASS; Vitest 48 files, 646 tests PASS", "two independent screen-judge passes, verdicts in docs/quality/PROTOTYPE_FIDELITY.md", "58 captures in docs/quality/evidence/fidelity/2026-11-15/"]
open_risks: ["PageHeader is a div, not a header element, so it is not an extra banner landmark; tests locate .ui-page-header", "no signed-in captures and no dark captures of detail pages", "the copy has not been reviewed by a native French editor"]
```
