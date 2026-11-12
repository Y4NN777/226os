# ADR-0002 — Interim single-maintainer ownership

- **Status:** ACCEPTED
- **Date:** 2026-11-11
- **Approver:** `Y4NN777`, repository owner (explicit decision recorded 2026-11-11; reconfirmed 2026-11-11 with the final G0 reconciliation)
- **Scope:** who holds which role until P0 completes; relation to ADR-0001 condition 1
- **Related:** [ADR-0001](0001-g0-conditional-baseline.md), `docs/GIT_WORKFLOW.md`, `AGENTS.md`, `docs/DEV.md` section 6
- **Amended by:** [ADR-0008](0008-team-work-areas-and-ownership.md) (work areas for the four-person team)
- **Metadata note 2026-11-12:** approver spelling normalised to the GitHub login (`Y4NN777`); the decision text is unchanged.

## Context

ADR-0001 condition 1 requires named product, backend, frontend, security, QA, and release owners before implementation begins. `docs/DEV.md` records three team members with responsibilities not yet assigned. On 2026-11-11 the owner decided explicitly that roles are not assigned yet: "finish P0 first before deciding who does what". Delaying P0 for a staffing discussion is not acceptable with the CADEV submission deadline of 2026-11-15 17:00 GMT.

## Options considered

1. **Assign roles now.** Rejected by the owner as premature and a distraction from P0.
2. **Leave roles undefined.** Tasks have no `owner`/`reviewer`, so accountability and approvals become ambiguous; rejected.
3. **One named interim owner for every role until P0 completes.** Keeps accountability explicit, keeps the work item contract (`owner`, `reviewer`) fillable, and is reversible. Chosen.

## Decision

- `Y4NN777` is the interim owner for product, backend, frontend, security, QA, and release until P0 completes. Work items set `owner: Y4NN777`; `reviewer` is `Y4NN777` unless another human is named on the item.
- Pull requests from work branches into `develop` may be owner-reviewed, as `docs/GIT_WORKFLOW.md` already allows for routine low-risk staging integration. Required CI checks still apply; owner review does not replace them.
- Promotion of `develop` to `main` (production) requires explicit owner approval recorded on the production PR with G7 evidence, plus the protected `production` environment approval. **During the interim, owner approval is sufficient for `main` promotion**; `docs/GIT_WORKFLOW.md` refers to this ADR for the "distinct human approval on `main`" rule, so that rule is satisfied by the owner approval until this ADR is superseded.
- The security and legal reviews required by G4 and G7 are still required. This ADR does not waive them. Owner self-review is not an independent security or legal review; the `skeptic` agent output is advisory evidence, not a substitute for the human review that G4/G7 name.
- **Revisit at P0 completion.** At that point the owner assigns named roles (and a second reviewer if available) in a superseding or amending ADR, before any post-P0 slice starts.
- Commit identity rules are unchanged: the owner's approved Git identity only; no AI agent as author or `Co-authored-by`; AI assistance recorded in `AI_USAGE.md`.

This ADR meets ADR-0001 condition 1 (named owners) with one interim owner for every role, and waives nothing else. The other conditions of ADR-0001, all gates G0 to G7, and all other human checkpoints in `AGENTS.md` section 4 stay in force.

## Consequences

- **Security:** one person can author, review, and approve a change in `develop`; separation of duties is lost for routine work. Compensating controls: mandatory green CI, deterministic gates, and `skeptic` review recommended for auth, authorization, and migration slices. Production keeps an explicit approval step.
- **Data:** no change. Destructive or live-data migrations still need explicit human approval and a restore plan.
- **Ops:** the release operator and approver are the same person; the release runbook in `docs/SHIP.md` is not weakened.
- **Cost:** none. Single-person bus factor is an accepted schedule risk.

## Evidence / verification required

- Work items and PR descriptions show `owner`/`reviewer` populated per this ADR.
- Production PR shows the owner approval and G7 evidence on the target SHA.
- G4 and G7 closeouts cite the actual human security and legal reviews, or report `BLOCKED`.

## Conditions

1. Expires at P0 completion; the owner revisits it then and assigns named roles (and a second reviewer if available) in a superseding or amending ADR before any post-P0 slice starts.
2. Does not authorize skipping G4/G7 reviews, branch protection, or production approval.
3. Decided 2026-11-11: with one maintainer, the "distinct human approval on `main`" rule of `docs/GIT_WORKFLOW.md` is satisfied for the interim by the owner approval recorded on the production PR. The production PR states that the approver is also the author; this limitation is accepted by the owner and ends with this ADR.
