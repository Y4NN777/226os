---
name: spec-to-task
description: Turn a 226OS SPEC user story or G0 decision into a work-item contract for docs/spec-kit/TASKS.md, without inventing requirements.
argument-hint: "<story ID or decision ID>"
disable-model-invocation: true
---

Source: $ARGUMENTS

1. Locate the source in `docs/SPEC.md`, `docs/PRD.md`, or `docs/spec-kit/G0_DECISIONS.md`. Quote it with `path:line`. If it is not found, stop.
2. Draft the work-item YAML from `AGENTS.md` (goal, story_id, priority, owner, reviewer, allowed_paths, acceptance_criteria, threats_and_risks, planned_tests, required_gates, evidence_paths, status: PLANNED).
3. Copy acceptance criteria and BDD cases from the source verbatim. Mark any gap `OPEN`, never fill it in.
4. Leave owner/reviewer as `OPEN` unless the user names them.
5. Show the draft and propose where to insert it. Edit `TASKS.md` only after the user approves.
