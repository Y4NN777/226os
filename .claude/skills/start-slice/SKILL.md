---
name: start-slice
description: Open one 226OS implementation slice from docs/spec-kit/TASKS.md after checking G0 conditions, owner/reviewer, branch, allowed paths, tests, gates, and stop condition.
argument-hint: "<TASK-ID>"
disable-model-invocation: true
---

Task: $ARGUMENTS

Run these checks and print each as PASS / BLOCKED with its source:

1. The task exists in `docs/spec-kit/TASKS.md`. Quote its YAML contract.
2. `owner` and `reviewer` are named, distinct humans (not `OPEN`).
3. The ADR-0001 conditions for this slice are closed (`docs/decisions/0001-g0-conditional-baseline.md`), and every `ADR-OPEN-*` it depends on has a numbered ADR in `docs/decisions/`.
4. Dependencies listed under the task are DONE with evidence paths.
5. Git: working tree clean; branch matches `^(feature|fix|refactor|test|docs|chore)/<domain>/<name>$` per `docs/GIT_WORKFLOW.md`, created from current `origin/develop`. Author identity is the human contributor's approved identity.
6. Every Make target the slice will run exists and its recipe has been read.

If anything is BLOCKED: stop, list the exact human action needed, and do not create application files.

If everything passes, write the slice brief: goal, story IDs, allowed paths, acceptance criteria, RED tests to write first, gates to run, delegate (`go-implementer` / `web-implementer` / none), and stop condition. Wait for user confirmation before delegating.
