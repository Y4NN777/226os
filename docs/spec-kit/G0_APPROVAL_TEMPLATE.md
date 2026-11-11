# 226OS G0 Approval Record

**Status:** CONDITIONAL GO recorded from explicit user approval; formal named-owner/ADR completion remains required

A human approver must complete this record after reviewing `G0_DECISIONS.md`, `AUTHORIZATION_MATRIX.md`, `API_CONTRACT.md`, `DATABASE_DESIGN.md`, `ASTRYX_VERIFICATION.md`, and the governance drafts.

```yaml
scope_decision: CONDITIONAL GO
approver: "User (explicit approval in chat; repository identity not provided)"
date: "2026-11-11"
approved_story_ids: []
approved_tasks: ["TASK-SPEC-001", "TASK-G0-001"]
owners:
  product: "OPEN"
  backend: "OPEN"
  frontend: "OPEN"
  security: "OPEN"
  qa: "OPEN"
  release: "OPEN"
accepted_decisions: ["G0-AUTH-001", "G0-AUTH-002", "G0-AUTH-003", "G0-AUTH-004", "G0-AUTHZ-001", "G0-API-001", "G0-API-002", "G0-DATA-001", "G0-DATA-002", "G0-DATA-003", "G0-UI-001", "G0-GOV-001", "G0-GOV-002", "G0-OPS-001", "G0-OPS-002", "G0-COMP-001"]
accepted_defaults: ["All proposed defaults in G0_DECISIONS.md accepted in principle, subject to the conditions below"]
rejected_or_deferred_decisions: []
required_adr_ids: ["ADR-OPEN-001", "ADR-OPEN-002", "ADR-OPEN-003", "ADR-OPEN-004", "ADR-OPEN-005", "ADR-OPEN-006", "ADR-OPEN-007", "ADR-OPEN-008", "ADR-OPEN-010"]
decision_register_choices:
  ACCEPT: []
  ACCEPT_WITH_CONDITION: []
  REJECT: []
  DEFER: []
security_exceptions: []
conditions: ["Assign named product, backend, frontend, security, QA, and release owners", "Create and approve the required ADRs", "Verify Astryx release/API and OpenAPI/data contracts", "Complete governance/legal/licence review", "Keep G1-G7 NOT RUN until actual evidence exists", "Do not implement or release from this conditional approval alone"]
expiry_or_review_date: "Before implementation kickoff"
rationale: "The proposed G0 defaults are accepted as the planning baseline; conditions preserve human control over security, architecture, legal, and execution evidence."
```

## Approval checklist

- [ ] All eight product domains and four differentiators remain preserved.
- [ ] P0 stories, ACs, owners, reviewers, allowed paths, dependencies, and evidence paths are assigned.
- [ ] Authentication/session and authorization matrix are accepted or explicitly conditioned.
- [ ] OpenAPI/error/idempotency policy is accepted or explicitly conditioned.
- [ ] PostgreSQL schema, migration, search, deletion, and seed policies are accepted or explicitly conditioned.
- [ ] Astryx release/API verification owner and acceptance criteria are assigned.
- [ ] Contribution, AI provenance, and third-party notice policies are reviewed.
- [ ] G0 threats and mitigations have owners.
- [ ] Every `G0_DECISIONS.md` row has an explicit `ACCEPT`, `ACCEPT_WITH_CONDITION`, `REJECT`, or `DEFER` choice.
- [ ] Conditions do not waive required evidence or permit secrets, critical vulnerabilities, or likely data loss.

Until signed, G0 is `BLOCKED` and implementation must not begin.
