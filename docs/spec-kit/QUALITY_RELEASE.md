# 226os Quality and Release Evidence

**Status:** G1 to G4 run in CI through `make quality`; G5 and G6 are NOT IMPLEMENTED; G7 is NOT RUN (nothing is deployed). The gate matrix, thresholds and commands are owned by [docs/TEST.md](../TEST.md); the exclusion policy by [docs/quality/COVERAGE_SCOPE.md](../quality/COVERAGE_SCOPE.md).

No `NOT RUN` or `NOT IMPLEMENTED` result is a pass. No exclusion, threshold or exception may be created to hide a failure.

## Evidence record

Each task/PR evidence record identifies:

```yaml
story_ids: []
commit_sha: ""
environment: ""
commands: []
result: PASS | FAIL | NOT RUN | BLOCKED
coverage: {}
security_checks: []
performance_measurements: []
reviewer: ""
known_risks: []
```

Evidence is retained with the target SHA. Screenshots, traces and reports must not contain credentials, tokens, private user data or secrets.

## Release prerequisites

The runbook in `docs/SHIP.md` is authoritative. Before any release decision: Datacloud account/product eligibility, DNS/TLS, secrets outside Git, protected production approval, immutable image references, database backup, restore evidence, migration plan, post-deploy smoke, and an honest P1/P2 status.
