---
name: ship-check
description: Read-only 226OS release readiness check for a develop -> main promotion. Never deploys or merges.
argument-hint: "[SHA]"
disable-model-invocation: true
---

Target: $ARGUMENTS (default: current `origin/develop` SHA)

Read-only. Never merge, push, tag, deploy, or access secrets.

1. Confirm `origin/main` is an ancestor of the target, so the promotion preserves the tested SHA (merge or fast-forward, not squash).
2. For G0-G7 (`AGENTS.md`, `docs/SHIP.md`, `docs/spec-kit/QUALITY_RELEASE.md`), find evidence tied to this exact SHA. Missing evidence is `NOT RUN`.
3. Check the deferred decisions G0-OPS-002 (Datacloud) and G0-COMP-001 (licensing). If either is still deferred, release is `BLOCKED`.
4. Confirm a distinct human production approver is recorded.
5. Output a gate table plus GO / NO-GO, and list the human actions needed. A GO here is advisory; production approval stays human.
