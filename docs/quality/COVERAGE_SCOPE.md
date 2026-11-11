# 226OS Coverage Scope

**Status:** OPEN / awaiting implementation inventory and human review

This file is the only authorized location for coverage exclusions. No exclusion is active yet.

## Backend

- Required: 100.00% of instrumentable first-party Go statements.
- Proposed exclusions: none until generated/vendor/configuration paths are identified and reviewed.
- Required companion evidence: authorization branch matrix, transaction/invariant integration tests, race tests.

## Frontend

- Required: strictly greater than 70% for statements, branches, functions, and lines across application source.
- Proposed exclusions: none until generated/vendor/pure-type/configuration paths are identified and reviewed.
- Required companion evidence: P0 E2E, keyboard/accessibility, locale, and throttled-network tests.

## Change control

Every exclusion must state path pattern, reason, owner, reviewer, date, and why the exclusion cannot hide behavior. A coverage report without this file's reviewed version is `BLOCKED`, not `PASS`.
