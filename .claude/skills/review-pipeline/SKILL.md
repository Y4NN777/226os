---
name: review-pipeline
description: 226os AI review pipeline (ADR-0013, ADR-0014, ADR-0015): stage 1 review and deep review, the JSON ledger, the Review gate, stage 3 auto-fix, stage 4 auto-merge, stage 5 batch, settle on merge, the findings issues and dashboard. Its invariants, failure history, test patterns and roadmap. Use when touching scripts/review-*.py, scripts/findings.py, scripts/close-merged-issues.py, .github/workflows/code-review.yml or review-*.yml, .github/*-paths.txt, .claude/commands/review-fix.md, deep-review.md, delegate-review.md, their tests, or docs/quality/REVIEW_CONTEXT.md.
---
# Review pipeline: maintenance and evolution

The code is the truth. This skill is a map: before you describe, change or judge any rule, read the file and the line
that implements it. Never describe a gate, a check or a review rule from memory or from older docs (CLAUDE.md pitfall).

## Sources to read first
- `docs/quality/REVIEW_CONTEXT.md`: what the review must not raise again (decided or accepted), and the stage rules.
- ADRs: `docs/decisions/0013-*` (default branch `develop`, `workflow_run`/`schedule` run the default-branch copy),
  `0014-*` (the review flow), `0015-autofix-by-proof.md` (fix by proof; phases 0 to 3; its "Phase N status" lines).
- `docs/team/OWNER_PREFERENCES.md` (review automation: merge on CI green and no open finding; notes go to the batch,
  never to the owner; "all" means every review item, notes included; prefer a code fix over "already decided").
- The tracker line of the auto-fix and deep-review entries in `TODO.md` (dated history of every pipeline decision).
- Issue #264 (ADR-0015 phase 0 and open pipeline follow-ups), #212, #201, #218, #227 (older follow-ups), #119 (dashboard).

## Map (stage, file, what it owns)
| Stage | Workflow | Script(s) | Owns |
|---|---|---|---|
| 1 review | `code-review.yml` job `review` | `review-context.py` (context pack), `review-comment.py` (pin, carry forward, write) | one comment per run, line 1 `Triage: <LEVEL>`, line 2 `<!-- 226os-ai-review -->`, a fenced JSON ledger in `<details>` |
| 2 deep | `code-review.yml` job `deep` | `.claude/commands/deep-review.md`, merged by `review-comment.py` | P0 paths (`.github/sensitive-paths.txt`, `.github/deep-review-paths.txt`), `"source":"deep"`, `Deep: PASS|PENDING|NOT RUN` line, `deep.status/sha` in the ledger |
| gate | `code-review.yml` job `gate` | `review-gate.py` | the required check `Review gate`: ledger sha = PR head, no blocking finding, CRITICAL rules |
| harvest | `code-review.yml` job `harvest` | `findings.py` | finding issues (`finding`, `sev:*`, `area:*`, `due:*`), `PR<N> notes` issues, `PR<N> follow-ups`, dashboard #119 |
| 3 auto-fix | `review-autofix.yml` (`decide`, `preflight`, `fix`, `publish`, end job) | `review-autofix.py` (`pr`, `context`, `prepare`, `apply`, `check`, `body`, `record`, `nothing`, `repro-lock`, `repro-run`, `static-after`, `run`; read `main()`), `.claude/commands/review-fix.md` | attempts, routing, the trusted push guard, the status comment |
| 4 auto-merge | `review-automerge.yml` | `review-automerge.py` | merge commit of exactly the evaluated head into `develop`; pipeline paths and sensitive paths go to the owner |
| 5 batch | `review-batch.yml` (`plan`, `batch`, `fix`, `publish`) | `findings.py batch`, `review-autofix.py batch/nothing` | one batch issue and one `fix/review/batch-YYYY-MM-DD` PR |
| settle | `review-merged.yml` | `findings.py settle`, `close-merged-issues.py` | closes or ticks sources by verdict on merge |

Path lists read from `develop` through the API: `.github/pipeline-paths.txt` (never auto-merged), `.github/sensitive-paths.txt`
(owner approval of the head), `.github/deep-review-paths.txt` (deep review).

## Ledger entry (read `review-comment.py` and the deep JSON schema in `code-review.yml` for the exact set)
`id`, `kind` (`finding` | `note`), `evidence` (`OBSERVED` | `INFERRED`), `sev` (`CRITICAL` | `HIGH` | `MEDIUM` | `LOW`),
`file`, `line`, `scenario`, `detail`, `fix`, `status` (`open` | `answered` | `accepted` | `fixed`), `source`
(`review` | `deep`), `first_sha`, `location` (`verified` | `mismatch` with a `probable` file:line | `unverified`). The ledger `sha` must be the PR head
(`PR_HEAD_SHA`), never the checkout HEAD. A declared CRITICAL without OBSERVED evidence and a real `file:line` counts as
HIGH. A stale ledger is NOT RUN, and a stale CRITICAL keeps blocking. Notes never block.

## Invariants (break one and you reopen a fixed incident)
1. **No PR text in a prompt.** Prompts carry the repository, numbers and finding ids only. Everything read from a PR,
   issue or comment is untrusted data, neutralised (`defuse`, `neutralise_free`, `redact_lines`) before it is written back.
2. **Trusted vs untrusted jobs.** A job holding a write token or a secret runs no PR-controlled code; trusted scripts
   come from the base commit as git objects into `$RUNNER_TEMP`. The model-calling job never sees `REVIEW_AGENT_TOKEN`
   (`review-agent` environment, deployment branch `develop` only; `review-agent-triggers_test.sh` guards it).
3. **`workflow_run` and `schedule` run the default-branch copy** (ADR-0013): a workflow change takes effect only after
   it is merged into `develop`. Test statically; say NOT RUN for the live path.
4. **Values reach `run:` through `env:`**, never `${{ }}` of untrusted text. Actions pinned by full SHA (copy existing pins).
5. **Fail closed** for gates, merges and counters; fail open only for advisory steps, and then print `NOT RUN` with the
   reason. A GitHub job has no neutral result: a run with nothing to do says `NO-OP` or `NOT RUN`, never a silent green.
6. **The model never closes, labels or ticks by itself.** `settle` closes FIXED / FIXED ELSEWHERE; NOT A DEFECT and
   DECIDED tick a NOTE only with a reference the script verifies (an ACCEPTED ADR naming the note's file and its id as
   code); on a finding they only comment.
7. **Auto-fix (ADR-0015 phase 0, PR #309, merged 2026-11-15):** attempts per (finding, head SHA),
   recorded when the run ends, pending written before the fixer; 2 heads per finding, 2 aborts per head, 7 runs in total
   per PR (`MAX_PR_RUNS`); routing per finding (`fixer` | `owner` + reason); `needs-owner` informational for the fixer (the other findings keep being fixed) but blocking for auto-merge (`BLOCKING_LABELS` in `review-automerge.py`),
   `autofix-stopped` blocking and removed only by a human; existing tests, fixtures, test infrastructure and test configs
   LOCKED at the trusted push, new test files under a lexical scan; findings outside the fixer allowlist go to the owner.
8. **Auto-merge** merges with `--match-head-commit`, merge commit only, never squash; refuses pipeline paths, an
   unreviewed sensitive path, any open finding, `needs-owner` and `autofix-stopped`.

## Failure history (each one became a rule; do not reintroduce)
- #252: two auto-fix attempts spent on findings in `scripts/` and `.github/` the fixer may not edit (path prefilter);
  the fixer marked a test `fixme` (weakening guard); `autofix-1` written before the "branch moved" check (attempts at run end).
- The batch cron fired once in two days (best-effort GitHub schedules): the batch also runs after `Review merged`.
- #187 independent review: a model verdict never closes a finding; pipeline tests are pipeline paths.
- #266: existence of a reference is not evidence (verified references only).
- Skeptic rounds on #309: counting assertions was bypassable (comment out, weaken matcher, rename, early return, build
  tags, `TestMain` + `os.Exit`, test configs): lock existing files instead of counting.

## Testing patterns
- Plain bash tests in `scripts/tests/`, `scripts/tests/lib.sh` (`ok`, `ko`, `finish`), a fake `gh` on `PATH` that
  records argv and answers from files; PyYAML for static workflow checks. Python runs with `python3 -I`.
- RED first, shown; then GREEN. A probe script in a scratch directory for every guard (try to get around it).
- Gates: every `scripts/tests/review-*_test.sh` and `findings*_test.sh`, `scripts/lint-shell.sh`,
  `python3 -I scripts/check-docs.py`, `bash scripts/check-todo-updated.sh`, a YAML parse of each touched workflow,
  `actionlint` when installed (else NOT RUN). `make lint-scripts` runs them; read `mk/*.mk` first.
- Live behaviour (`workflow_run`, environments, comment edits) cannot run before merge: say NOT RUN and tell the owner
  what to watch on the first live run.

## Roadmap
ADR-0015 phase 1 (classes, reproducer job, static oracles, reproducer hash rule), phase 2 (proof record, human review of
auto-fix commits by Y4NN777 or mibienpanjoe, merge queue, routing by path), phase 3 optional (diff-scoped mutation and
precision tracking, needs ADR-0016). Follow the execution plan the owner hands over, if any; otherwise the ADR's
"Delivery, in phases" table.
