# 226os Requirements Traceability

**Status:** SPECIFIED; execution evidence NOT RUN

## Domain and differentiator coverage

| Stable scope ID | PRD scope | CADEV/P0 slice | SPEC/BDD | Initial task | Gates |
|---|---|---|---|---|---|
| DOM-SOLVE | Solve | Q&A, snippets, tags, votes, accepted answer, search, relations | US-SOLVE-001 | TASK-API-001, TASK-WEB-001 | G0-G5 |
| DOM-DISCOVER | Discover | catalogue, country/category filters, sourced resource detail, relations | US-DISCOVER-001 | TASK-DATA-001, TASK-API-001, TASK-WEB-001 | G0-G6 |
| DOM-BUILD | Build | public GitHub URL, project, contribution need, unique application, private applications | US-BUILD-001 | TASK-DATA-001, TASK-API-001, TASK-WEB-001 | G0-G5 |
| DOM-LEARN | Learn (PRD 5.9, amendement 3.1; priorities APPROVED by the owner on 2026-11-11) | read/publish articles and tutorials (Markdown, code, tags, level, language), relations, public free training series | US-LEARN-001 | TASK-DATA-001, TASK-API-001, TASK-WEB-001 | G0-G5 |
| DOM-IDENTITY | Identity & Community | account, session, minimum profile, shared identity | US-IDENTITY-001 | TASK-SEC-001, TASK-API-001, TASK-WEB-001 | G0-G5 |
| DOM-CONNECT | Connect | roadmap contract; no false CADEV promise | PRD 5.5, 8 | TASK-SPEC-001 | G0 |
| DOM-LANGUAGE | Language | FR/EN UI, original content preserved; translation/search later | US-IDENTITY-001; PRD 5.6 | TASK-WEB-001, TASK-SPEC-001 | G0, G5 |
| DOM-ACCESS | Access | responsive, low-bandwidth P0; offline drafts later | US-ACCESS-001 | TASK-WEB-001, TASK-OPS-001 | G0, G5, G6 |
| DOM-INTELLIGENCE | Intelligence | deterministic PostgreSQL search and context links; semantic/AI later | US-DISCOVER-001, US-IDENTITY-001; PRD 5.8 | TASK-DATA-001, TASK-API-001 | G0-G6 |

LEARN P1/P2 (structured series with progress, author highlights, richer formats; certificates, interactive exercises) stay roadmap scope. P1/P2 domains remain in the product vision. They are roadmap scope, not hidden deletions, and must not be represented by fake UI.

## Differentiating experiences

| ID | Experience | Observable P0 slice | Evidence |
|---|---|---|---|
| DIFF-01 | African Integration Intelligence | A resource shows supported countries, category, source, verification date, docs link, and related Solve content. | API response, UI trace, seeded fixture, G5/G6 report |
| DIFF-02 | Problem-to-Project | A question can expose a related resource and project; the relation resolves in both directions where authorized. | Relation invariant, E2E trace, G3/G5 |
| DIFF-03 | Cross-Border Collaboration | A project need can be discovered by skills/country context; UI is available in FR and EN without changing original content. | E2E locale/mobile trace, G5 |
| DIFF-04 | Evidence-Based Technical Knowledge | Questions/resources/projects expose source, version/context when available, verification date, and confirmation state without implying official endorsement. | DTO/schema, UI state tests, provenance review, G3-G5 |

## Journey coverage

| Journey | Required cases | Existing source | Planned evidence |
|---|---|---|---|
| A Solve | nominal, invalid input, unauthorized accept, duplicate vote, search retrieval | PRD 7A, SPEC US-SOLVE-001 | API integration, BDD, E2E |
| B Discover | country/category filter, source display, stale/unknown data honesty | PRD 7B, SPEC US-DISCOVER-001 | search integration, E2E, provenance fixture |
| C Build | public GitHub URL, unique application, private application denial | PRD 7C, SPEC US-BUILD-001 | SSRF/security tests, API integration, E2E |
| D Cross-domain | question/resource/project relations resolve and remain bounded | PRD 7D, SPEC US-IDENTITY-001 | relation constraints, E2E |
| E Language | FR/EN UI, original text preserved, no implicit auto-translation | PRD 7E, SPEC US-IDENTITY-001 | locale E2E and text-preservation test |
| G Learn | public read, authenticated publish, unauthenticated publish denied, original FR/EN content preserved | PRD 7G, SPEC US-LEARN-001 | API integration, authz negative test, locale E2E |
| F Access | 360px mobile, slow network, loading/empty/error, no essential heavy media | PRD 7F, SPEC US-ACCESS-001 | throttled E2E, a11y, performance report |

## Traceability rule

No implementation task is eligible for `GREEN` until it points to a stable scope ID, an acceptance example, an API/data/UI contract, a test-first plan, and a gate evidence path. Missing normative contracts remain `OPEN` and block parallel implementation.
