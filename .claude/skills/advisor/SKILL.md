---
name: advisor
description: Get an independent read-only architecture opinion from the Opus advisor subagent on a genuinely disputed 226os decision.
argument-hint: "<decision question>"
disable-model-invocation: true
---

Question: $ARGUMENTS

1. Confirm the question is a genuinely disputed architecture/ADR decision. If it is routine inventory, formatting, or already decided in `docs/decisions/` or ADR-0001, say so and stop.
2. Delegate to the `advisor` subagent with: the question, the relevant doc paths, what is already decided, and what is ruled out. Do not include your preferred answer.
3. Relay its recommendation, evidence paths, and whether an ADR is required. Do not write the ADR unless the user asks.
