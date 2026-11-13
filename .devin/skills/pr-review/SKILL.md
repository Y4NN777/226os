---
name: pr-review
description: Rigorous pull-request review for 226os. Reads the full diff (never just the description), verifies every claim against code/CI/tests, classifies findings as blocking or follow-up, and posts reviews in English with inline comments. Use when assigned or asked to review any PR in this repository.
---

# PR review — 226os

Reviews are read-only until the human approves posting. Posting a review is a GitHub mutation: only after explicit approval.

## 1. Context before reading code

- `gh pr view <N> --json title,body,headRefOid,commits,files,reviewRequests` — read the whole body. Note stacked PRs (`stacked on #X`): review only the author's own commits (`git diff <base-of-stack>...HEAD`), but keep the stack's merge constraints in mind.
- `gh pr checks <N>` — CI is authoritative. A failing required check is blocking regardless of local results. Read the failed log (`gh run view --job <id> --log-failed`) and reproduce locally before commenting.
- Read the slice block in `TODO.md` and the issue the PR claims to close: allowed_paths, acceptance criteria, required gates, reviewer pair (ADR-0008).

## 2. Read the whole diff, not the summary

- `git fetch origin pull/<N>/head:pr-<N>` then `git worktree add /tmp/pr<N> pr-<N>` — never review in the user's checkout.
- `git diff origin/develop...HEAD` (fetch `origin/develop` first; a stale ref silently inflates the diff with already-merged PRs).
- Read every changed file fully, including tests. For stacked PRs diff against the stack parent, not `develop`.
- Invoke the domain skills for whatever the diff touches: `go-backend`, `postgres-data`, `api-contract`, `security-authz`, `tdd-unit`, `integration-testing`, `ci-cd-delivery`, `web-frontend`, `bdd-acceptance`.

## 3. Verify claims, don't trust them

For each claim in the PR body, find the artifact: exact command, exit code, file:line. Specifically check:

- **Tests**: are RED tests real (would they fail without the change)? Are skips honest (`NOT RUN`, never counted as pass)? Run `gofmt -l`, `go vet`, `go test -race -count=1` on the touched packages in the worktree.
- **Coverage/quality claims**: re-run or read the coverage report; 100% first-party Go statements, >70% web on four metrics.
- **Contract**: `api/openapi.yaml` is design-first — new/changed operations, error codes (closed enum), headers, status policy (401/403/404 anti-enumeration), Idempotency-Key/ETag rules must match `apigen` output and `docs/spec-kit/AUTHORIZATION_MATRIX.md`.
- **Migrations**: forward-only, immutable after merge, complete header (purpose, locks, volume, tx, compat, restore, owner), named constraints, no `Down`, sequential numbering via `make migrate-new`. Data migrations need upgrade-path evidence (`TestUpgradeFromVersion*`).
- **Security**: secrets redacted in `String`/`GoString`/`LogValue`/errors/logs; no credential in URL; authz decisions server-side with negative tests; injection surfaces (SMTP, LLM prompts, headers, SQL); fail-closed config.
- **Generated code**: never hand-edited; `make generate` drift clean.
- **CI vs local discrepancies**: if local passes but CI fails (or vice versa), find the environmental cause — do not assume flake.
- **Honesty**: `NOT RUN` items must stay open; deferred items must be recorded in TODO/issue, not silently dropped.

## 4. Classify findings

- **Blocking**: CI red, contract violation, security hole, missing required gate evidence, out-of-scope changes, weakened gates/tests, secrets exposure, authorship violations.
- **Non-blocking**: residual risks worth a TODO/issue, style, theoretical edge cases with a documented rationale, deferred decisions already tracked.
- Every finding cites file:line or CI evidence. No vague approvals ("looks good") and no approval on a red required check.

## 5. Write the review (English, always)

- Body: verdict first, verification table (claim → evidence), blocking findings, non-blocking observations, merge notes (order, merge-commit vs squash constraints, fingerprints).
- Inline comments via the reviews API for file-specific points: `gh api repos/<owner>/<repo>/pulls/<N>/reviews --input review.json` with `comments[]` (`path`, `line`, `side: RIGHT`, `commit_id` = full head SHA). Events: `APPROVE`, `REQUEST_CHANGES`, `COMMENT` (post-merge reviews use `COMMENT`).
- Match the project's tone: precise, evidence-cited, honest about what was NOT verified.

### House style (owner's model — keep reviews this tight)

Numbered findings, one line each on the bug, the expected fix, and the RED test to add. No praise, no filler, no restating the diff. End with what was verified and an explicit next step ("re-request review once pushed"). Example shape:

```
Requesting changes on three security points (all observed in <file>):

1. **<area>.** <what is wrong, one sentence>. <expected fix>. Add a test for <case>.
2. ...
3. ...
4. **To confirm, <file> <func>.** <question to verify against infra/config>.

Keep 100% coverage. Everything else looks good. Re-request review once pushed.
```

## 6. Repo-specific pitfalls learned

- Stacked PRs: the gitleaks `.gitleaksignore` fingerprint is commit-scoped — merging with squash/rebase breaks it; merge commits required.
- `make lint` runs every `scripts/tests/*_test.sh` — script tests must pass without optional tools (sops, age) installed, or CI fails.
- `check-todo-updated.sh` requires TODO.md to move with any `src/ db/ api/ scripts/ mk/ Makefile` change.
- `notStub`/route gate: structural (autogenerated marker) + behavioral (zero-value call, panic) checks; a 501 stub must not mask a real implementation.
- Middleware order is contract: apigen `Middlewares` slice — last element runs first; `scripts/router-middleware.txt` must match exactly.
- Commits must be authored by the human contributor only; AI assistance goes in `AI_USAGE.md`, never in trailers.

## 7. Advisory bot review (github-actions triage) — self-check before push

The `code-review` workflow runs Claude + open-code-review on every push to a PR targeting `develop`. Its rules live in `.opencodereview/rule.json` and its method in `.claude/commands/delegate-review.md` — read them; the bot re-reviews the **whole** PR diff each push, not just the delta.

Before pushing auth/security-adjacent Go, run the same selection locally and audit the diff against the Go rule: handlers transport-only, authz on every route, no secrets/emails/passwords in logs or error text, wrapped errors, 100% statement coverage. Then think like the bot's security pass — it has repeatedly caught real enumeration oracles and unbounded-work issues:

- Rate-limit buckets keyed by account/email must be hit **before** existence lookups, or 429-vs-4xx becomes an oracle.
- Any path returning early on "already exists" must pay the same crypto cost (dummy Argon2id) and the same validation (policy checks before lookup, or 422-vs-202 leaks existence).
- Anonymous endpoints that run Argon2id need a per-IP rate-limit bucket; `password.Hash` is capped by the `sem` semaphore — keep expensive work inside it.
- One-shot tokens: validate/hash inputs **before** consuming; mutating multi-step writes go in one atomic query.
- Client IP: right-most X-Forwarded-For only (Traefik appends; earlier entries are spoofable).

Local preview: `scripts/review.sh ocr delegate preview --from origin/develop --to HEAD` (same exclusions as CI).

## 8. Lessons from the S1 auth-sessions hardening rounds

Patterns that reviewers (human and bot) actually caught — apply before pushing security-adjacent code:

- **Timing equalization is symmetric or it is an oracle.** A dummy Argon2id on the silent path is only correct if the known path pays equivalent crypto work. On password reset the known path does none (lookup + insert + detached mail), so the dummy made the *unknown* path slower — an inverted oracle. Check both directions: list the work each branch does, then equalize the difference, not just "add a dummy".
- **Validate policy before existence lookups.** Sign-up returned 202 for existing emails before checking password policy → 422-vs-202 leaked existence. Order: rate-limit bucket → input validation → lookup → crypto.
- **Rate-limit scope strings are DB-constrained.** `rate_limit_buckets.scope` allows one `:` (`^[a-z_]+(:[a-z_]+)?$`). `reset:confirm:ip` violated it → 500 on real PostgreSQL while unit fakes passed. Fake stores do not enforce CHECK constraints — only pgtest does. Keep scope names two-segment.
- **Store errors must not silently downgrade.** Only `domain.ErrNotFound` may map to anonymous/no-op (SignOut, ResolveSession→AccountByID). Everything else → `INTERNAL_ERROR`. A swallowed store error turns an outage into a false 204/401.
- **Session expiry writes must respect CHECK constraints.** `sessions_expiry_ck` requires `idle_expires_at <= absolute_expires_at`; `TouchSession` must use `min(now+idleTTL, absoluteExpiresAt)`, not raw `now+idleTTL`.
- **Length limits count runes, not bytes.** `utf8.RuneCountInString` for the 12..128 password contract; byte counting rejects valid multi-byte passwords and accepts short CJK ones.
- **One-shot tokens: validate/hash first, consume atomically last.** Weak passwords must not burn reset tokens; multi-step consume+update+revoke goes in a single CTE. Sibling tokens for the same user must be burned in the same statement (a second open reset token survives otherwise → account takeover window).
- **Detached mail needs thread-safe fakes and polling tests.** `context.WithoutCancel` + timeout keeps SMTP off the request path (timing oracle), but `fakeMail`/transport fakes need mutexes and tests must poll `waitSent`/`waitMail` instead of reading `Sent()` once.
- **`password.Hash` shares the Argon2id semaphore.** `sem` lives in `encode` so Hash, dummy-hash init and Verify are all capped at 4 concurrent; anonymous endpoints that can trigger hashing need a per-IP bucket.
- **Unresolved review threads block merge even with green CI.** After pushing fixes, reply to each inline thread with the fix commit + covering test, then `resolveReviewThread` via GraphQL. Follow-ups deferred to the owner get an explicit "acknowledged, deferred" reply — do not leave threads silently open.
- **Re-request review after resolving**; `reviewDecision: CHANGES_REQUESTED` only clears on human re-review, never on green checks.

## 9. Pre-push / pre-PR checklist (226os)

Run in order before every `git push` on a slice branch, and again before opening or updating a PR:

1. **Diff audit vs bot rules** — `scripts/review.sh ocr delegate preview --from origin/develop --to HEAD`; self-review the same files the bot will read (`.opencodereview/rule.json` + §8 patterns). The bot re-reads the whole PR each push, so a fixed bug can resurface in triage — verify no regression in untouched files.
2. **Format + vet + lint** — `gofmt -l` clean on touched files; `go vet`; `go tool golangci-lint run` on touched packages (and `--build-tags integration` when integration files changed). Note: `golangci-lint` is a `go tool`, not on PATH.
2b. **Contract + script gates** — `make lint-contract` (sqlc vet requires `LIMIT` on every `:many` query — CI-only rule unit tests never see) and `make lint-scripts` (runs every `scripts/tests/*_test.sh`; self-test fixtures that append methods to the real `server.go` break when the probed operation gets implemented — repoint them to a still-stubbed op). These two failed on PR #94 despite green unit/integration tests.
3. **Unit + race** — `go test -race -count=1` on every touched package (identity, app, platform/…). Never `-count=1`-less cached results for evidence.
4. **Coverage** — first-party Go must stay 100% statements; `go tool cover -func` and check for dead branches (unreachable defensive checks fail the gate — remove or test them).
5. **Real PostgreSQL** — `go test -race -count=1 -tags integration ./internal/platform/pgtest/` whenever `db/queries/**`, migrations, repositories or service SQL paths changed. Unit fakes do NOT enforce CHECK constraints, scope formats, or CTE atomicity — pgtest is the only proof.
6. **sqlc drift** — regenerate (`make generate` / sqlc) after any `db/queries/**` change and commit `dbgen` output; drift check must be clean.
7. **Contract/tracker sync** — `scripts/public-routes.txt`, `router-middleware.txt`, unimplemented-route lists, authz probes, and `TODO.md` must move with route/middleware changes (`check-todo-updated.sh` requires TODO.md movement on `src/ db/ api/ scripts/ mk/ Makefile` diffs).
8. **AI_USAGE.md** — record the AI assistance for the slice; keep `license_review`/`closure` fields honest (PENDING/open until owner sign-off).
9. **Commit hygiene** — human identity only (`git config user.*`), imperative message focused on why, **no AI co-author/trailers**, no secrets, no unrelated files (`.devin/` stays untracked unless owner asks).
10. **Push + PR** — non-force push to the slice branch; PR to `develop` with the work-item contract, RED test named, evidence table (PASS/FAIL/NOT RUN per gate — never claim `make quality` on a different SHA without saying so).
11. **After push** — watch the bot triage AND human reviews; verify every finding against code before fixing; reply to each inline thread with fix commit + test, resolve threads, re-request reviewers. `CHANGES_REQUESTED` + unresolved threads both block merge independently of green CI.
