# ADR-0011 — AI assistant provider

- **Status:** ACCEPTED
- **Date:** 2026-11-12
- **Approver:** `Y4NN777`, repository owner (decision taken by the owner, recorded in `ROADMAP.md` section 14 and section 10, O5)
- **Scope:** which provider and model back the sourced assistant, how the key is handled, and what stays open
- **Related:** [PRD amendment 3.2](../PRD.md), [ADR-0009](0009-secrets-with-sops-and-age.md), `ROADMAP.md` (O5, K13, K21), `src/api/internal/platform/llm`, `src/api/cmd/aicheck`

## Context

The proposed PRD amendment 3.2 makes a sourced contextual assistant part of P0: it answers only from published content, cites its sources, and stays disabled by configuration until it is integrated. It needs an external language model. ROADMAP O5 lists the questions the decision had to settle, and K13 and K21 record the risks: unknown model id and limits, and the provider terms for users in the EEA, Switzerland and the UK.

## Decision

1. **Provider and model.** Google AI Studio, Gemma 4 family. The staging model is `gemma-4-26b-a4b-it`. The free tier was chosen by the owner on 2026-11-12 (V3 decided).
2. **Provider-agnostic port.** The code depends on a port in `src/api/internal/platform/llm`; the AI Studio adapter, a fake and a disabled client sit behind it. Switching provider is a configuration change.
3. **Key handling.** `AI_API_KEY` is a secret (see ADR-0009) and only reaches the `api` container; `scripts/verify-containers.sh` fails if another service receives it.
4. **Switch.** The assistant is off unless `AI_ENABLED=true` (default `false`, set per environment), and `AI_MODEL` has no default.
5. **Measured so far** (`aicheck` on the real key, 2026-11-12): V1 done, the model list includes `gemma-4-26b-a4b-it` and `gemma-4-31b-it`; V2 probes and V4: about 6 s per answer, 3 of 3 probes ok for `gemma-4-26b-a4b-it`, while `gemma-4-31b-it` was slower and 2 of 3 probes were unavailable.
6. **Caveats, not hidden.** The model returns reasoning text in its answer: it must be dropped and one answer format applied. The real project quotas (V2) are still open: the probes do not measure them. 429, 500, 503 and timeouts must be tracked.
7. **EEA, Switzerland and UK (K21), OPEN.** Under the AI Studio terms only paid services may be offered to users in those regions, and the free tier lets the provider use and review submitted content. The owner must decide between restricting the assistant for those users or linking a billing account. Until then the assistant stays switchable off per environment.

## Consequences

- Development and the golden evaluation set run against the fake adapter; the real provider is exercised separately and recorded as such.
- Only the question text and retrieved public content are sent, after redaction of obvious secrets and emails, with a visible notice (PRD amendment 3.2).
- Free-tier limits and terms may change; the port keeps the exit cheap.
- Open work: K21, V2, the cost cap or quota guard, timeouts and anonymous access are tracked in `ROADMAP.md` and `TODO.md`.

## Alternatives considered

- **OpenAI:** recorded first in the ROADMAP decisions log, then superseded by the entry of 2026-11-12 that chose Google AI Studio.
- **A paid or billing-linked project:** possible at the owner's decision, mainly to settle K21; not chosen on 2026-11-12.
- **Another provider through the port:** the stated fallback if the free tier is rate-limited or its terms are unsuitable.

## Evidence / verification required

`aicheck` (`models`, `ping`, `limits`) output, the ROADMAP O5, K13 and K21 entries and the decisions log of 2026-11-12, `scripts/verify-containers.sh`, and the `llm` package tests.
