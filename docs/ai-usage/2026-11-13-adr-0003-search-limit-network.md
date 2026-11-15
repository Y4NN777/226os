## ADR-0003 amended: anonymous search limited per client network

```yaml
date: "2026-11-13"
contributor: "Y4NN777"
tool_or_model: "Claude Code, Claude Opus 5.5 (owner's lead session, no subagent)"
purpose: "Owner decision of 2026-11-13 (item 1 of the PR96 audit, #117). ADR-0003 said anonymous search is limited to 60 per IP per minute, while the code limits per client network (/24 for IPv4, /48 for IPv6). The AI drafted the one-line amendment of ADR-0003: the per-network rule, the reason (it bounds the cost of an anonymous scan per network and cannot be dodged by changing the last byte), the trade-off (people behind one shared network share the budget) and when to revisit it. The owner decided the rule."
repository_or_source_context: "branch docs/decisions/search-limit-network, PR #123 (merged 2026-11-13)"
files_or_sections_affected: ["docs/decisions/0003-authentication-sessions-csrf-rate-limits.md"]
third_party_sources_consulted: []
license_review: "NOT APPLICABLE: documentation only, no dependency"
closure: "closed by the owner (2026-11-15)"
human_reviewers: ["Y4NN777 (owner closure, 2026-11-15; merged in #123)"]
tests_or_evidence: ["scripts/check-docs.py PASS (PR #123)"]
open_risks: ["record written on 2026-11-15, after the merge: none was filed at the time"]
```
