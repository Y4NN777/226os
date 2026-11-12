---
name: ship-check
description: Read-only 226os release readiness check for a develop -> main promotion. Never deploys or merges.
argument-hint: "[SHA]"
disable-model-invocation: true
---

Target: $ARGUMENTS (default: current `origin/develop` SHA)

Read-only. Never merge, push, tag, deploy, or access secrets.

1. Confirm `origin/main` is an ancestor of the target, so the promotion preserves the tested SHA (merge or fast-forward, not squash).
2. For G0-G7 (`AGENTS.md`, `docs/SHIP.md`, `docs/spec-kit/QUALITY_RELEASE.md`), find evidence tied to this exact SHA. Missing evidence is `NOT RUN`.
3. Check the G7 evidence for this SHA: Datacloud account and product, DNS records for the production and staging hosts, a named production approver, a backup taken and a restore drill, a rollback-by-digest drill, and TLS verified on staging. Any missing item is `NOT RUN`, and release is `BLOCKED` until it is recorded. Also check the licence decisions still recorded as open owner decisions (`docs/SHIP.md`, `THIRD_PARTY_NOTICES.md`); an open one blocks release.
4. Confirm a distinct human production approver is recorded.
5. Output a gate table plus GO / NO-GO, and list the human actions needed. A GO here is advisory; production approval stays human.
