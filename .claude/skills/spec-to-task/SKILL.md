---
name: spec-to-task
description: Turn a 226os SPEC user story or G0 decision into a slice or task entry for TODO.md, without inventing requirements.
argument-hint: "<story ID or decision ID>"
disable-model-invocation: true
---

Source: $ARGUMENTS

1. Locate the source in `docs/SPEC.md`, `docs/PRD.md`, or `docs/archive/spec-kit/G0_DECISIONS.md`. Quote it with `path:line`. If it is not found, stop.
2. Draft the entry in the format of the existing slice blocks of `TODO.md` (owner, reviewer, branches, depends on, tasks with RED tests) and the work-item fields from `AGENTS.md` (goal, story_id, priority, owner, reviewer, allowed_paths, acceptance_criteria, threats_and_risks, planned_tests, required_gates, evidence_paths, status: PLANNED).
3. Copy acceptance criteria and BDD cases from the source verbatim. Mark any gap `OPEN`, never fill it in.
4. Leave owner/reviewer as `OPEN` unless the user names them; assignments follow ADR-0008.
5. Show the draft and propose where to insert it. Edit `TODO.md` only after the user approves.
