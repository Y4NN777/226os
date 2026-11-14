# ADR-0012 — AI assistant integration (grounded answers, limits, fallback, golden set)

- **Status:** ACCEPTED (owner `Y4NN777`, 2026-11-14, at the 12:00 checkpoint)
- **Date:** 2026-11-14
- **Approver:** `Y4NN777`, repository owner
- **Scope:** how `POST /assistant/answers` (`askAssistant`, S9) retrieves, sends, limits, validates and falls back. The provider and model are decided in [ADR-0011](0011-ai-assistant-provider.md).
- **Answers:** ROADMAP O5, TODO S9 (the PROPOSED ADR task). The SPEC story and the `AUTHORIZATION_MATRIX.md` rows stay with the owner (H13).
- **Related:** `src/api/internal/assistant`, `src/api/internal/platform/llm`, `src/api/internal/platform/ratelimit`, [ADR-0003](0003-authentication-sessions-csrf-rate-limits.md), PRD amendment 3.2

## Context

The PRD amendment 3.2 makes a sourced assistant part of P0: it answers only from published content, cites it, refuses without a source and never invents guarantees. The LLM port, the fake adapter, the output cleaning and the tracking exist. This ADR fixes the integration rules that the code and tests follow.

## Decision

1. **Access.** Members only, with a verified email (401 anonymous, 403 `EMAIL_NOT_VERIFIED`), CSRF on the POST, subject from the session only. Anonymous access is not offered at P0 (O5).
2. **Data sent (AI-10).** Only the question text, after redaction, and the retrieved public passages (also redacted) go to the provider, in one request. No user id, e-mail, session, profile or private content is sent. Redaction removes e-mail addresses, private keys, bearer and JWT tokens, well-known API key shapes (Google, GitHub, AWS, Slack, Stripe), `password=`, `token=` and `api_key=` assignments, URL credentials and any 32+ character alphanumeric run. It errs on the side of removing.
3. **Retrieval.** The existing PostgreSQL search (S2), called as a member, over published, visible, public content only (questions, resources, projects, Learn articles). The question becomes an OR query of its first 8 meaningful words (FR/EN stop words dropped), because the search ANDs plain terms. The best 5 hits are kept, one per content group in turn so one type does not crowd out the others. Each passage is the search snippet (at most 300 characters today), cut to 1,500 characters. Private, removed, draft and archived items are excluded by the search queries themselves. No vector database.
4. **Grounding (AI-1, AI-2).** With zero passages the provider is not called: outcome `no_source` with a link to the search. The prompt asks the model to cite passages as `[S1]`, `[S2]`. The server validates the reply: every marker must name a retrieved passage and at least one must be cited; otherwise the answer is dropped and the reply is `no_source`. Markers are renumbered in order of first use and `sources` lists exactly the cited passages.
5. **Untrusted content.** Passages go in a delimited `<sources>` data section; the system prompt declares it data, never instructions; the delimiter tags are stripped from passage text and titles so a passage cannot close the section. An obeyed injection cannot add a source: it fails the citation check.
6. **Answer format (AI-3, AI-4, AI-5, AI-9).** The reply is labelled `generated_by: ai`; it is never stored, never an `answers` row and has no vote target. The prompt forbids guarantees. The reasoning text is removed by `llm.CleanAnswer`. The "Comment vérifier" or "How to verify" section is split into `how_to_verify`; when the model wrote none, a fixed default is used. Amended 2026-11-14 (owner, the 12:00 cuts back in scope): the search hits now carry `published_at`, each passage is dated in the prompt (or marked undated), each source returns `published_at` (null when undated), and `how_to_verify` ends with a sentence computed from the cited sources: the dates are shown next to each source, or some sources carry no date; the prompt also asks the model to say so when passages disagree (AI-5). AI-3 is enforced on the server: sentences that promise a guarantee or a verified solution (FR and EN patterns, negated forms kept) are removed before the citation check, and a reply left with no valid citation becomes `no_source`; only the class `guarantee_removed` is logged.
7. **Language.** The answer is in the question's language: the optional `lang` field wins, otherwise a simple FR/EN function-word and accent detection (a tie is French).
8. **Limits and where the counters live.** Both use `platform/ratelimit` (table `rate_limit_buckets`, no migration) and both fail closed (a limiter error is 503).
   - Per member: scope `assistant:user`, 10 per hour, 429 `RATE_LIMITED` with `Retry-After`.
   - Global provider guard: scope `assistant:global`, counts only the calls that reach the provider (after the `no_source` check, before the call). **Deviation from the decision of 300 calls a day:** the shared limiter accepts windows of one hour at most and purges buckets two hours after their window start, so a one-day bucket cannot be held without changing `platform/ratelimit`. The guard is therefore 12 calls per hour (288 a day at most), answering 503 `SERVICE_UNAVAILABLE` with `Retry-After` once spent. A true daily cap needs a ratelimit change (longer window and retention), outside this slice; the owner decides.
9. **Timeout and fallback.** The provider call is made once and never retried. Its deadline is `AI_TIMEOUT` capped at 20 s, inside the 30 s route budget (B3). The config default of `AI_TIMEOUT` is still 30 s; the cap makes the effective default 20 s, and aligning `platform/config` is a follow-up. Assistant off (`AI_ENABLED=false`), limiter failure, provider error, provider 429, timeout or empty output: 503 `SERVICE_UNAVAILABLE` with `Retry-After: 30` (the spent global guard keeps the limiter's own wait), a fixed message without provider details. The page and the search never depend on it.
10. **Logging.** Only model, class, status and latency (the `llm.Tracker`) and the outcome with a source count are logged. The question, the passages, the answer and the key are never logged.
11. **Golden set and pass criteria.** `TestAssistant*` in `platform/pgtest` runs 17 cases on the FAKE adapter over test-only fixtures (fixed UUIDs, provenance `demo`, inserted by the test, never seeded): G1 answered with a valid citation; G2 French question; G3 no source, provider not called; G4 injected instruction ignored (and an obeyed one refused); G5 citation to a source not retrieved rejected; G6 private, removed, draft and archived items never retrieved nor cited; G7 no e-mail or secret in the prompt; G8 no `answers`, `votes` or `questions` row; G9 call 11 is 429; G10 disabled, G11 provider down (one call), G12 provider 429 and G13 global guard spent are 503; G14 and G15 (AI-3: a guarantee sentence removed, a reply left without citation refused), G16 (AI-5: dated sources and wording) and G17 (`Retry-After` on the disabled and limiter-failure 503), added 2026-11-14; plus the authorization and CSRF matrix. CI passes only if all pass. The real-provider run (V4: FR/EN latency and quality on the same questions) is recorded separately by the owner and stays NOT RUN until done.

## Consequences

- The free tier (V2 quotas, V3) is protected twice: 10 per member per hour and 12 per hour overall. A demo with several simultaneous users shares the 12 calls an hour.
- Retrieval quality is bounded by the search snippet (300 characters) and the keyword OR query. A richer passage (the full body, 1,500 characters) needs a new query in `db/queries` and is a follow-up.
- Source URLs follow the web routes: `/questions/{id}`, `/discover/{id}`, `/build/{id}` and `/apprendre/articles/{id}` (the last two since 2026-11-14, #173 and #177); a refusal links to `/recherche?q=`.
- **Owner decisions of 2026-11-14, ACCEPTED:** the global guard stays at 12 calls per hour (point 8, in place of 300 a day); passages stay the 300-character search snippets (point 3); the 20 s cap stays in the app wiring (point 9).
- Learn series are not retrieved (the search has no series group).

## Alternatives considered

- **A table for a daily counter:** needs a migration (R6); not needed once the hourly guard is accepted.
- **Dropping a bad citation and keeping the rest:** rejected; an answer that cites a source that was never retrieved is not trusted at all.
- **Anonymous access:** rejected at P0 (abuse and cost on a free tier).

## Evidence / verification required

`make test-unit` and `make coverage` (assistant packages at 100%), `make test-integration-go` (the golden set above), `scripts/check-authz-routes.sh`. NOT RUN: the real-provider run (V4), V2 (real quotas) and the visible notice in the UI.
