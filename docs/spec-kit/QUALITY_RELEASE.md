# 226OS Quality and Release Evidence

**Status:** SPECIFIED; all executable gates are NOT RUN in this documentation phase

## Gate matrix

| Gate | Required evidence | Current status | Blocking condition |
|---|---|---|---|
| G0 | approved stories, AC, threats, owners, plan, explicit decision register, and governance review | CONDITIONAL GO | user accepted proposed defaults; named owners, ADRs, evidence plans, and governance/legal review remain required |
| G1 | format/lint/typecheck/generation logs from target SHA | NOT RUN | no implementation or CI |
| G2 | Go first-party statements 100.00%; frontend statements/branches/functions/lines strictly >70%; unit evidence | NOT RUN | no code/tests/scope file |
| G3 | real PostgreSQL integration, OpenAPI validation/generation diff, permission/invariant tests | NOT RUN | contracts and database absent |
| G4 | negative auth tests, SAST, SCA, secret and license scans | NOT RUN | security implementation and reports absent |
| G5 | Playwright P0 traces, FR/EN, mobile, keyboard, axe/contrast/manual review | NOT RUN | application absent |
| G6 | reproducible dataset/config, SQL plans, API latency, page weight, LCP/INP/CLS, throttled network | NOT RUN | baseline and implementation absent |
| G7 | SHA/image digest, staging smoke, backup/restore, migration and rollback evidence, human approval | NOT RUN | deployment artifacts and approval absent |

No `NOT RUN` result is a pass. No exclusion, threshold, or exception may be created to hide a failure.

## Coverage scope

The canonical exclusion policy must be maintained in `docs/quality/COVERAGE_SCOPE.md`. Until that file is reviewed and populated, coverage gates are blocked. Generated code, vendors, migrations, pure types/configuration, and mocks may be excluded only by a versioned, justified list; business branches and difficult endpoints may not be excluded.

## Evidence record

Each task/PR evidence record must identify:

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

Evidence is retained with the target SHA. Screenshots, traces, and reports must not contain credentials, tokens, private user data, or undisclosed secrets.

## Release prerequisites

The release runbook in `docs/SHIP.md` remains authoritative. Before any human release decision, the team must verify Datacloud eligibility, DNS/TLS, secrets outside Git, protected production approval, immutable image references, database backup, restore evidence, migration plan, post-deploy smoke, and an honest P1/P2 status.
