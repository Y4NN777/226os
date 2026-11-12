---
name: start-slice
description: Open one 226os implementation slice from its block in TODO.md after checking ADR coverage, owner/reviewer, branch, allowed paths, tests, gates, and stop condition.
argument-hint: "<slice id, e.g. S1>"
disable-model-invocation: true
---

Slice: $ARGUMENTS

Run these checks and print each as PASS / BLOCKED with its source:

1. The slice exists in `TODO.md` (slice block such as `S1`). Quote its owner, reviewer, branches, dependencies, tasks and allowed paths.
2. `owner` and `reviewer` are named GitHub logins, distinct, and the pair matches ADR-0008 (`docs/decisions/0008-team-work-areas-and-ownership.md`): a PR is never reviewed by its author, backend PRs by the lead or the other backend developer, frontend PRs by the owner, ops and platform PRs by a backend developer.
3. ADR-0001..0010 (`docs/decisions/`) cover the slice and no ADR the slice needs is still open.
4. Dependencies listed under the slice are DONE with evidence paths.
5. Git: working tree clean; branch matches `^(feature|fix|refactor|test|docs|chore)/<domain>/<name>$` created from current `origin/develop`, or `hotfix/<domain>/<name>` created from `origin/main` for an approved incident (`docs/GIT_WORKFLOW.md`, `.github/workflows/branch-policy.yml`). Author identity is the human contributor's approved identity.
6. Every Make target the slice will run exists and its recipe has been read.

If anything is BLOCKED: stop, list the exact human action needed, and do not create application files.

If everything passes, write the slice brief: goal, story IDs, allowed paths, acceptance criteria, RED tests to write first, gates to run, delegate (`go-implementer` / `web-implementer` / none), and stop condition. Wait for user confirmation before delegating.
