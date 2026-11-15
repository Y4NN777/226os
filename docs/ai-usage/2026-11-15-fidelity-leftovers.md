## Fidelity leftovers: Discover type chips, #ask on /solve, landing landmarks

```yaml
date: "2026-11-15"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Sonnet 5.5 (web-implementer subagent of the lead session)"
purpose: "Three leftovers of the prototype-fidelity pass. The Discover type chips showed truncated labels at tablet width: they now take their own grid row and scroll inside their strip if still too narrow. The shell link « Poser une question » (/solve#ask) did nothing when already on /solve: SolvePage now reacts to the hash changing, opens the ask view and focuses its first field, and clears the hash when the view closes so the link works again. The landing section heads are div elements instead of header elements, so they cannot count as banner landmarks."
repository_or_source_context: "branch feature/web/fidelity-leftovers; S7 web pages; prototypes/html/pages/discover.html"
files_or_sections_affected: ["src/web/src/features/discover/DiscoverPage.tsx", "src/web/src/features/discover/discover.css", "src/web/src/features/solve/SolvePage.tsx", "src/web/src/features/solve/SolvePage.test.tsx", "src/web/src/features/landing/LandingPage.tsx", "src/web/src/features/landing/LandingPage.test.tsx", "TODO.md", "docs/ai-usage/2026-11-15-fidelity-leftovers.md"]
third_party_sources_consulted: ["none"]
license_review: "NOT APPLICABLE: original code, no new dependency"
closure: "open"
human_reviewers: ["pending PR review"]
tests_or_evidence: ["RED: navigating to /solve#ask while on the list did not open the ask view (1 failing test), then GREEN", "landing banner test (one banner landmark) already passed in jsdom before the change, kept as a guard", "captures of /discover at 375, 800, 1024 and 1280 px in the session scratchpad"]
open_risks: ["the prototype has no type chips (a select in a collapsible panel); the app keeps its segmented control", "the type chips fix is CSS only and checked by capture, not by a unit test"]
```
