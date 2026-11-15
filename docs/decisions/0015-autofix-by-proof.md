# ADR-0015: auto-fix by proof. A finding becomes a test, a fix is accepted on facts, a human merges it

- **Status:** ACCEPTED (owner `Y4NN777`, 2026-11-15)
- **Date:** 2026-11-15
- **Approver:** `Y4NN777`, repository owner
- **Scope:** stage 3 (auto-fix on a pull request) and stage 5 (batch) of the review pipeline, and how a finding is marked fixed. Amends [ADR-0014](0014-ai-review-merges-develop.md) on one point: a pull request whose only change is an auto-fix commit no longer merges without a human review.
- **Evidence:** the 2026-11-15 research report summarised in "Sources" below, and the run history of #252.
- **Related:** `scripts/review-autofix.py`, `review-comment.py`, `findings.py`, `review-automerge.py`; `.github/workflows/review-autofix.yml`, `review-batch.yml`; `.claude/commands/review-fix.md`; [`REVIEW_CONTEXT.md`](../quality/REVIEW_CONTEXT.md).

## Context

On 2026-11-15 the auto-fix on #252 showed five failures:
1. A run that stopped because the branch moved still used up an attempt. The `autofix-1` label is written before that check.
2. Attempts were spent on findings in `scripts/` and `.github/`, which the fixer may not edit (allowlist), so they could never be fixed.
3. The fixer marked an E2E test `fixme` because a feature was not yet on `develop`. The feature arrived six minutes later. This weakened a test to clear a finding.
4. A fix is judged by the next LLM review, so a model decides whether a model's fix is right.
5. Runs that did nothing end green, so the owner cannot tell what happened.

The research behind this ADR found the same pattern in every deployed system: Meta SapFix and Getafix, Google's suggested edits and sanitizer fixes, GitHub Copilot Autofix, Uber Piranha. A deterministic oracle (the crash, the analyser or the test, run again) decides that the defect is gone, filters check for regressions, and a human lands the change. None lets a model judge its own fix. A model judging patches agreed with humans at κ = 0.57. About 30 % of SWE-bench patches that pass the tests behave differently from the reference fix. In classic program repair, most "plausible" patches simply deleted functionality. Claude models cheat mainly by editing tests, and a prompt rule barely changes that, while locking the tests brings it close to zero.

## Decision

1. **The reviewer finds, it never judges a fix.** A finding is marked fixed only from a proof record (point 5), never from the model's text.
2. **Each finding gets a class when it is created:**
   - `executable`: a behaviour or security claim at `file:line`;
   - `static`: a deterministic tool already in the gates can check it (lint, typecheck, gosec, semgrep, a grep rule);
   - `prose`: design, naming, documentation.

   An entry without a class counts as `prose`.
3. **An executable finding needs a reproducer before any fix.**
   - A separate reproducer job writes one test per finding.
   - A trusted step runs it on the pull request head, three times. It must be red each time, and for the stated reason: not a compile error, not a missing symbol.
   - The step records the test file's `sha256` and the SHA in the ledger.
   - A finding that is not red on the head is `NOT_REPRODUCIBLE` and leaves the auto-fix lane. Testing on the head, not on `develop`, is the rule that would have prevented failure 3.

   A static finding uses its tool as the oracle: the rule fires on the head, and no longer fires after the fix.
4. **The fixer never writes or changes a test.** The reproducer and the existing tests and fixtures are read-only for it. The trusted push step refuses the diff, and the finding goes to the owner, when any of these holds:
   - it adds `skip`, `only`, `fixme`, `xit`, `t.Skip`, `nolint`, `eslint-disable`, `@ts-ignore` or `@ts-expect-error`;
   - a test file loses assertions;
   - it adds a coverage exclusion or raises a timeout;
   - the reproducer's hash changed.
5. **A fix is accepted by proof:**
   - the reproducer, same hash, goes from red on the SHA before the fix to green on the SHA after it;
   - the full `make` gates are green;
   - the guard of point 4 passed;
   - the reproducers run again on the merge result before the merge.

   This proves the fix is **plausible**, not correct.
6. **A plausible fix waits for a human review before it merges** (owner, 2026-11-15). The reviewer is the owner (`Y4NN777`) or Joseph (`mibienpanjoe`); an approval from either one, on the head commit, is enough. Stage 4 does not auto-merge a pull request while an auto-fix commit on it has not been reviewed. The auto-fix prepares the change and its proof; it does not land it.
7. **Prose findings go to the batch** (owner, 2026-11-15). They no longer block the merge of the pull request where they were found. The batch tries to fix them, and point 6 applies to the batch pull request.
8. **Attempts count per finding, and only when they happened.**
   - The key is `(pr, finding, head SHA)`, written when the run ends, never before. A run whose head moved, that was cancelled, or that the infrastructure aborted is `SUPERSEDED` and uses up nothing.
   - At most two attempts per finding, then `needs-owner`.
   - A finding whose `file` is outside the fixer's allowed paths goes to the owner before any attempt, with the reason.
9. **Routing by path.** A route table, read from `develop`, maps a path to the fixer agent: `src/web` to web-implementer, `src/api` and `db/queries` to go-implementer, `scripts` and `.github` to ci-implementer. Sensitive paths (`.github/sensitive-paths.txt`) get a prepared patch and its proof as a run artefact, never a push.
10. **One computed status per pull request.** A single comment is recomputed from facts on every event. It has one row per finding: class, state, last fact (run id, SHA), next action, who acts. A run with nothing to do ends `NO-OP` or `NOT RUN`, never as a green success.

11. **`develop` moves to the GitHub merge queue in phase 2** (owner, 2026-11-15). A ready pull request enters the queue, and GitHub builds the merge commit with what is ahead of it (`merge_group`). The reproducers and gates run on that exact commit, which is what lands. This replaces the "exact evaluated head" check of `review-automerge.py`. It needs the branch setting and a `merge_group` trigger on the required workflows.

### Per-finding states

```
NEW ─ classify ─▶ PROSE ─────────────────────────────▶ BATCH
                  STATIC ─ rule fires on head ─▶ READY   (else STALE)
                  EXECUTABLE ─▶ REPRO_PENDING ─▶ REPRO_RED   (else NOT_REPRODUCIBLE ▶ BATCH or owner)
READY | REPRO_RED ─ route ─▶ PREPARED (sensitive path) ─▶ NEEDS_OWNER
                         └─▶ FIXING (attempt k, SHA)
FIXING ─ head moved, cancelled, infrastructure ─▶ SUPERSEDED (k not used)
FIXING ─ guard refuses ─▶ NEEDS_OWNER
FIXING ─ still red or gates red ─▶ FAILED (k) ─▶ FIXING (k+1), or NEEDS_OWNER after 2
FIXING ─ red to green, same hash, gates green ─▶ FIXED_PLAUSIBLE ─ human review ─▶ ACCEPTED
```

A new push recomputes every state from the facts on the new head.

## Consequences

- **Quality:** acceptance rests on tests and tools, not on a model's opinion. Test weakening is refused by a check, not discouraged by a prompt. The reproducers stay as regression tests.
- **Speed:** a fixed pull request no longer merges by itself (point 6). Fewer findings are auto-fixed: research systems obtain a reproducer for only about 70 % of real bug reports, and LLM review findings include more prose.
- **Cost:** three reruns per reproducer and one extra job. No new dependency in phases 0 to 2.
- **Owner:** clearer status (point 10), and fewer `needs-owner` labels caused by wasted attempts.

## Delivery, in phases (efforts are estimates)

| Phase | Content | Estimate |
|---|---|---|
| 0 | Points 8 and 10, the guard of point 4, and the path prefilter. No new model work. This fixes failures 1, 2, 3 and 5. | 2 to 3 session-days |
| 1 | Points 2 and 3: classes in the ledger, the reproducer job, static findings through the gate tools. | 4 to 6 session-days |
| 2 | Points 1, 5, 6, 9 and 11: acceptance by proof, the merge queue, routing, the review gate on auto-fix commits. | 3 to 4 session-days |
| 3, optional | A diff-scoped mutation check on P0 paths and precision tracking per finding category. These need new dependencies and an ADR. | 3 to 5 session-days |

Each phase is one pull request on pipeline paths, so the owner merges it.

**Phase 0 status (2026-11-15):** implemented on `feature/review/adr-0015-phase-0`, not merged. Point 8: attempts per (finding, head SHA) in the status comment, recorded by `review-autofix.py record` when the run ends; the path prefilter. Point 4: the guard in `check`. The fixer may ADD new test files, under a lexical scan only; existing tests, fixtures, test infrastructure (directories with a known name such as `tests`, `testdata`, `testutil`, `pgtest`, and under `src/api/internal` any directory whose non-test Go file imports `testing` at the base commit, `vitest.setup`, `playwright.config`, `jest.config`, `.mocharc`, `go.work`, `.golangci*`) and test configs are LOCKED: any change, deletion or rename is refused (the literal reading of "read-only"); the infrastructure and configs are also owner paths. A NEW test file, and every added line, gets only a LEXICAL scan (skip-like tokens also when split over lines, nolint, eslint-disable, ts-ignore, coverage exclusions, a timeout option, TestMain/init/os.Exit/process.exit, a mock of the assertion library, any added `//go:build` or `// +build` line in a Go file): best effort, it cannot prove a new test is strong. The reproducer-hash rule waits for phase 1. Aborted runs are recorded too (2 per finding and head send the finding to the owner; completed, pending and aborted runs together are capped at 7 per pull request, the constant `MAX_PR_RUNS`), a missing push token is refused before the model, routing per finding (`fixer` or `owner` with a reason code, kept in the state, an owner finding is never fixed again), two labels (`needs-owner` is informational and blocks nothing; `autofix-stopped` stops the fixer for the whole pull request, set only by the PR-level fail-closed causes: 7 runs in total, a failed record, pending entries of two dead runs, a state too large; only a human removes it, and removing it restarts the cap for 7 more runs and clears the dead pending entries (the marker that makes a removal a reset is saved only on a stop whose label was written, and a label already there is never written again; a state too large is saved compacted, routes, marker and the attempts the caps need, with the label only, so its removal is a reset too; too large even compacted, the owner deletes the status comment then removes the label); auto-merge refuses both). Aborted runs are per head and never a route; a guard refusal or a fixer answer routes to the owner only the findings it names (all of them if it names none). `LOCKED_PATH` names the real test directories, and a directory under `src/api/internal` is test infrastructure by a fact of the base commit read from git objects (a non-test Go file imports `testing`), never by a name ending in `test`, and a pending attempt is written before the fixer runs (one left by a dead run counts). Point 10: the status comment; a run with nothing to do prints `NO-OP` or `NOT RUN` and a workflow annotation (a GitHub job has no neutral conclusion). Also, the batch starts after a merge (`workflow_run` of "Review merged"), the cron staying as the fallback (#264).

## Defaults (owner may revisit)

- reproducers stay as ordinary tests next to the code they cover;
- three reruns;
- an intermittently red reproducer is `NOT_REPRODUCIBLE` and opens a flaky-test note.

## Sources

Meta SapFix (ICSE-SEIP 2019), Getafix (OOPSLA 2019), TestGen-LLM (FSE 2024) and ACH (FSE 2025); Google's ML-suggested edits for review comments (ICSE-SEIP 2024), its sanitizer-fix report (2024) and Passerine (ICSE-SEIP 2025); the GitHub Copilot Autofix responsible-use documentation; Uber Piranha (ICSE-SEIP 2020); Agentless (2024); SWT-bench (NeurIPS 2024); SpecRover (2024); PatchDiff (ICSE 2026); Smith et al. (FSE 2015) and Qi et al., Kali (ISSTA 2015); ImpossibleBench (2025); LLM-judge studies (arXiv 2511.10865, 2604.16790); bors, Zuul and GitHub merge queue documentation; GitHub Actions concurrency documentation. Some figures (ImpossibleBench, the Claude system card wording, Google flakiness rates) come from secondary summaries and were not checked against the primary text.
