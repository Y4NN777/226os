---
name: verify-slice
description: Verify a 226OS slice before declaring success. Reports PASS/FAIL/NOT RUN/BLOCKED per gate from actual command evidence on the target SHA.
argument-hint: "<TASK-ID>"
disable-model-invocation: true
---

Task: $ARGUMENTS

1. Record the target SHA (`git rev-parse HEAD`) and the changed paths (`git diff --name-only origin/develop...HEAD`). Flag any path outside the task's `allowed_paths`.
2. For each gate in the task's `required_gates`, run only the reviewed Make target that `docs/TEST.md` maps to it, after reading its recipe. Capture the exact command, exit code, and report path.
3. A gate without an implemented target is `NOT RUN`. A gate blocked by a missing decision or owner is `BLOCKED`. Never infer PASS from a summary, cache, or skipped command.
4. Delegate an independent review to the `skeptic` subagent with the diff range and the task contract.
5. Write the closeout: Goal, Scope, Changes, Evidence, Gates (table), Residual risks, Next actions. Save it under the task's `evidence_paths` only if the user approves.
