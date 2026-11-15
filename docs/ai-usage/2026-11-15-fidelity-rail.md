## Tablet icon rail (S7 fidelity pass)

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "The owner asked every page to match its prototype. From 768 to 1023 px the prototype shows a 92 px icon rail (icon over a short label); the app kept the 232 px sidebar, which crowded Découvrir. The AI wrote RED tests first, then a RailNav in the shell (flat list of areas, same order, aria-current on the current one, same « Navigation principale » label) chosen by a matchMedia hook on the tablet range, its CSS, the rail-wide footer and brand offsets, and viewport-fit=cover in index.html for the bottom bar's safe-area padding."
repository_or_source_context: "branch feature/web/fidelity-rail from feature/web/prototype-fidelity; S7 web shell; prototypes/html/assets/css/layout.css (.nav-item rail rules)"
files_or_sections_affected: ["src/web/src/app/Shell.tsx", "src/web/src/app/Shell.test.tsx", "src/web/src/app/navigation.ts", "src/web/src/styles/shell.css", "src/web/index.html", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-rail.md"]
third_party_sources_consulted: ["none beyond the project prototype CSS and the installed Astryx AppShell"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED first: the rail test failed (the sidebar group titles were still shown)", "new tests: rail links, order and decorative icons; only the current area has aria-current; one primary navigation and no bottom bar; no rail above 1023 px", "captures of /discover and / at 800x1000 and 1024x800, light and dark, next to the prototype (scratch, not committed)", "gates in src/web: typecheck 0, lint 0, format:check 0, vitest run --maxWorkers=1 (see PR)"]
open_risks: ["the Découvrir type filter still truncates at 800 px (« S... », « T... ») because it is a feature-owned control, not the shell", "rail labels are 11 px, as in the prototype, below the 16 px body rule", "no real screen reader run"]
```
