## Slice

Slice id and section in `TODO.md` (for example `S2`):

## Summary

What changes and why (one slice per PR):

## Allowed paths

- [ ] Every changed file is inside the slice's allowed paths (or the exception is explained below).

## Acceptance criteria

- [ ] <criterion from the slice or SPEC story>

## Risks

Security, migration, licence, performance or rollout risks (or "none"):

## Evidence

Exact commands run, with `PASS` / `FAIL` / `NOT RUN` and the target commit SHA:

| Command | Result |
|---|---|
| `make quality` | |

## Checklist

- [ ] RED test written and named above, then GREEN.
- [ ] `TODO.md` updated in this PR.
- [ ] Reviewer requested per ADR-0008 (name):
- [ ] Co-workers or helpers named (or "none"):
- [ ] No `Co-authored-by` or other AI author trailer; AI assistance recorded in `AI_USAGE.md` if relevant.
- [ ] No secret, key or real email added.
- [ ] CADEV files (`THIRD_PARTY_NOTICES.md`, `AI_USAGE.md`, original-work statement in `docs/research/REFERENCES.md`, submission audit in `docs/SHIP.md`) are untouched or only added to.
