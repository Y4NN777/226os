# ADR-0014 — The AI review pipeline merges into `develop`; the owner steps in on CRITICAL, sensitive and pipeline changes

- **Status:** ACCEPTED (owner `Y4NN777`, 2026-11-15)
- **Date:** 2026-11-15
- **Approver:** `Y4NN777`, repository owner
- **Scope:** who must review a pull request into `develop` before it merges. Amends the review rule of [ADR-0008](0008-team-work-areas-and-ownership.md) (the "Reviews" paragraph of decision 2 and the last sentence of decision 5) for pull requests into `develop` only. The promotion `develop` to `main`, the owner decisions of ADR-0008 decision 3 and the work areas are not changed.
- **Answers:** the contradiction between `docs/team/OWNER_PREFERENCES.md` ("a pull request into `develop` merges when CI is green and the AI review has no open finding") and ADR-0008, `AGENTS.md`, `docs/DEV.md` and `docs/GIT_WORKFLOW.md` (human review of the area reviewer required), found by the documentation audit of 2026-11-15.
- **Related:** `.github/workflows/code-review.yml`, `review-autofix.yml`, `review-automerge.yml`, `review-batch.yml`, `review-merged.yml`; `scripts/review-gate.py`, `review-ping.py`, `review-autofix.py`, `review-automerge.py`, `review-comment.py`, `findings.py`; `.github/sensitive-paths.txt`, `.github/deep-review-paths.txt`, `.github/pipeline-paths.txt`; [ADR-0013](0013-default-branch-develop.md); [`REVIEW_CONTEXT.md`](../quality/REVIEW_CONTEXT.md) ("Pipeline", the full rule set).

## Context

Since 14 November the review pipeline runs end to end on `develop` (ADR-0013). Stage 1 reviews each push, stage 2 runs a deep adversarial review when a P0 path is touched, stage 3 fixes open findings, stage 4 merges and stage 5 works the leftover findings and notes in batches. Every merge decision is in `scripts/review-automerge.py`, run from `develop`'s copy, and not in the model's text. The repository variables read on 2026-11-15 are `CODE_REVIEW_ENABLED=true`, `CODE_REVIEW_GATE=report`, `AUTO_FIX=on` and `AUTO_MERGE=on`. Merges now follow these rules, and no area reviewer approves them. The docs still say a human area review is mandatory.

## Decision

1. **A pull request into `develop` needs no human review by default.** Stage 4 merges it (a merge commit of exactly the evaluated head) when every condition of `review-automerge.py` holds:
   - the required checks (`Validate branch promotion`, `quality`, `Review gate`) pass on the head;
   - the review ledger covers the head and its verdict is neither NOT RUN nor CRITICAL;
   - no finding blocks (an open finding of any severity, or a HIGH or CRITICAL one not marked `fixed`; notes never block);
   - a P0 path has a deep review PASS of the head;
   - no `needs-owner`, `do-not-merge` or `owner-override-pending` label is set, and the pull request is not a draft or a fork.

   A teammate may still review, and a human `CHANGES_REQUESTED` from the owner refuses the merge. The authoritative list is the script's docstring and `REVIEW_CONTEXT.md` "Pipeline".
2. **The owner steps in only where the pipeline hands over:**
   - **CRITICAL.** An open CRITICAL backed by OBSERVED `file:line` evidence is never auto-fixed. The pull request is labelled `needs-owner` and auto-merge refuses it. The owner fixes it or lifts it with the `owner-override` label, which only the owner can add. A CRITICAL without that evidence counts as HIGH. A stale CRITICAL keeps blocking, and a push never lifts it.
   - **HIGH.** An open HIGH not marked `fixed` labels the pull request `needs-owner` and requests the owner's review (`review-ping.py`). Stage 3 may still fix it. The owner settles a HIGH that was answered or accepted.
   - **Sensitive paths** (`.github/sensitive-paths.txt`, read at the tip of `develop`). The owner must approve exactly the head commit, or must have opened the pull request and authored every commit, each verified by GitHub. The only exemption is a signed review-batch pull request whose sensitive paths are all documentation (`docs/decisions/*`, `THIRD_PARTY_NOTICES.md`), under the HMAC conditions of `REVIEW_CONTEXT.md`.
   - **Pipeline paths** (`.github/pipeline-paths.txt`) and any `.github/workflows/` change. These are never auto-merged: they are labelled `needs-owner` and the owner merges them by hand (ADR-0013 decision 3).
   - **Auto-fix that gives up.** After a third attempt, or when the push guard refuses the fixer's diff, the pull request is labelled `needs-owner`.
   - **Batch verdicts that are not proof.** NOT A DEFECT, DECIDED, OWNER, AFTER FREEZE, and a FIXED that the merged diff does not back, stay open with a comment for a human to close (`findings.py settle`).
3. **Unchanged:**
   - the production promotion `develop` to `main` needs G7 evidence and explicit human approval;
   - ADR-0008 decision 3 still applies: ADRs, new dependencies, migrations on live data, secrets, costs, remote access, production approval, and the G4 and G7 security and legal reviews stay with the owner;
   - nobody reviews their own pull request when a human review is requested.
4. **An outage never merges.** A review that did not run is NOT RUN, and auto-merge refuses it. The Review gate fails open on an outage, so it does not turn the pull request red.

## Consequences

- Merges into `develop` no longer wait for a reviewer of the area. Review capacity, the bottleneck named in ADR-0008, moves to the CRITICAL, HIGH, sensitive and pipeline cases.
- **The Review gate check is in `report` mode** (`CODE_REVIEW_GATE=report`), so the check itself fails nothing. Auto-merge applies its own conditions whatever the gate mode, so CRITICAL and sensitive paths are enforced by auto-merge refusing them, not by a red check. A person with write access can still merge by hand. Like the pipeline-path rule of ADR-0013, this is a convention backed by auto-merge, not a technical block. Switching the gate to `enforce` is an owner setting and is not decided here.
- `docs/team/CONTRIBUTORS.md` reviewer pairs still name who reviews when a human review is requested, for example a HIGH ping or an owner request.
- The docs that stated the old rule now point here: `AGENTS.md`, `docs/DEV.md` section 3 ("Revue IA") and `docs/GIT_WORKFLOW.md` promotion flow step 2.

## Alternatives considered

- **Keep the mandatory area review and leave auto-merge off.** Rejected by the owner: with four people and one owner holding two areas, every pull request would wait for a human, and the pipeline was built to remove that wait.
- **Rewrite ADR-0008.** Rejected: an accepted ADR is not edited. This ADR amends only its review rule.
- **Switch the Review gate to `enforce` at the same time.** Not decided: it is a separate owner setting, and auto-merge already refuses the same cases.

## Evidence / verification required

The behaviour is the code: the docstrings of `scripts/review-automerge.py`, `review-gate.py`, `review-ping.py`, `review-autofix.py`, `review-comment.py` and `findings.py`, and their tests under `scripts/tests/review-*` and `scripts/tests/findings*`. The repository variables were read with `gh variable list` on 2026-11-15. NOT RUN: no new test, because this ADR records existing behaviour.
