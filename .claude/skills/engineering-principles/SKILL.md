---
name: engineering-principles
description: Cross-cutting 226os design and review standards: SOLID applied concretely to Go and React/TS, clean/hexagonal boundaries inside the Go modular monolith, KISS/YAGNI, DRY vs coupling, naming, error handling, the TDD refactor step, and code-review standards. Use when touching src/**, db/** or api/**.
---
# 226os Engineering Principles

## Scope and triggers
- Any design, implementation, refactor or review under `src/api/**`, `src/web/**`, `db/**` or `api/**`.
- Pair this skill with the layer skill, for example `web-frontend`.
- Precedence when sources conflict: security/integrity > PRD > SPEC > accepted ADRs > technical rules > local preference (`docs/RULES.md:3`). The principles below never override a project document. Flag conflicts instead of resolving them silently (`DESIGN.md:284`, `docs/DEV.md:64`).

## Project rules (binding)
**Stack**
- Go modular monolith with Chi, `pgx` and `sqlc`, REST and OpenAPI 3.1. No general-purpose ORM, no external search engine, no microservices, Redis or Kubernetes without an ADR (`docs/DEV.md:7`, `docs/DEV.md:10`, `docs/architecture/ARCHITECTURE.md:27`).

**Boundaries**
- Handlers are transport only. Business invariants live in the domain/service layer (`docs/DEV.md:10`, `docs/RULES.md:12`, `docs/spec-kit/TASKS.md:92`).
- Each domain package has `domain`, `service`, `repository` and `transport` layers, with dependencies pointing toward business logic (`docs/DEV.md:44`).
- A module never imports a neighbour's HTTP transport. Share through explicit interfaces and limited models (`docs/architecture/ARCHITECTURE.md:20`).
- Domains and their allowed dependencies are listed in `docs/SPEC.md:83-94`. Example: solve may depend on identity only through interfaces.
- The web layer talks only to explicit APIs and has no DB access. The API owns invariants and permissions. Access checks sit as close as possible to the business operation (`docs/architecture/ARCHITECTURE.md:19`, `docs/architecture/ARCHITECTURE.md:23`).
- Cross-domain links use typed content relations, never one giant shared table (`docs/architecture/ARCHITECTURE.md:22`, `docs/spec-kit/DATABASE_DESIGN.md:7`).

**YAGNI**
- Don't create empty P1/P2 modules. Folders appear only when features start (`docs/SPEC.md:96`).
- No implicit new features (`docs/RULES.md:6`).

**Go and data**
- `gofmt`, `go vet`, `staticcheck` and `golangci-lint`, each pinned (`docs/DEV.md:46`).
- Every DB, HTTP and external call carries a context deadline (`docs/DEV.md:45`).
- Errors are typed and wrapped (`docs/DEV.md:45`).
- SQL is always parameterized (`docs/RULES.md:12`).
- Multi-table mutations run in a transaction (`docs/SPEC.md:119`).
- Keep generated `sqlc` code in sync with `db/queries` (`docs/DEV.md:47`).
- Invariants are enforced in SQL constraints, not only in code (`docs/SPEC.md:106-116`).
- Migrations are immutable after merge (`docs/spec-kit/DATABASE_DESIGN.md:39`).
- The authorization subject comes from the server session, never from a client-supplied owner ID (`docs/spec-kit/AUTHORIZATION_MATRIX.md:46`).

**Errors and logging**
- Status: problem+json (`docs/SPEC.md:99`) is firm. The field details, the immutable-migration rule and the authorization-matrix items are the ADR-0001 planning baseline (`docs/decisions/0001-g0-conditional-baseline.md:16-19`). They become normative only once the contracts are approved (`docs/spec-kit/API_CONTRACT.md:3`, `docs/spec-kit/DATABASE_DESIGN.md:3`). Re-check before coding.
- Errors leave the API as RFC 9457 problem+json with a stable `code`. Never expose stack traces, SQL, tokens, internal hosts or whether a private object exists (`docs/spec-kit/API_CONTRACT.md:51-57`).
- Never log secrets or unnecessary personal data (`docs/RULES.md:14`, `docs/SPEC.md:104`).

**Contract and change control**
- OpenAPI is the shared contract. Clients are generated or validated, and no drift is allowed (`docs/RULES.md:15`).
- Stack, contract, major-dependency or boundary changes require an ADR. Silent drift is rejected (`docs/DEV.md:64`).

**TDD and coverage**
- Work BDD then RED → GREEN → REFACTOR. Refactoring keeps tests green and invariants validated (`docs/TEST.md:28-31`, `docs/RULES.md:18`).
- Never delete or weaken a test to hide a failure. Changing an expectation needs a business reason and a human reviewer (`docs/TEST.md:32`, `AGENTS.md:74`).
- Go needs 100% first-party statement coverage. Frontend needs strictly more than 70% on all four metrics. Exclusions go only in `docs/quality/COVERAGE_SCOPE.md`, with no opportunistic exclusions (`docs/TEST.md:6-8`, `docs/RULES.md:19`).
- Statement coverage is not enough for authorization, validation and transactions. Those need a positive/negative branch matrix plus real-PostgreSQL integration tests (`docs/TEST.md:9`, `docs/TEST.md:16`).

**Evidence, PRs and identity**
- Report `OBSERVED`, `INFERRED` or `UNVERIFIED`, and `NOT RUN` is never `PASS` (`AGENTS.md:76`, `docs/RULES.md:21`).
- Keep PRs small, with human review for product, security, migration and architecture changes (`docs/RULES.md:33`).
- Commit as the responsible human. Never add AI agents as authors or `Co-authored-by` trailers. Record AI assistance in `AI_USAGE.md` (`AGENTS.md:82-84`).

## Expert practices
**Go interfaces and dependency injection**
- Define interfaces where they are consumed, keep them small, and don't define one until a real caller needs it. Never create interfaces on the implementor side "for mocking" ([Go CodeReviewComments: Interfaces](https://go.dev/wiki/CodeReviewComments#interfaces)).
- "Accept interfaces, return structs." Require no more and promise no less ([Cheney, SOLID Go Design](https://dave.cheney.net/2016/08/20/solid-go-design)).
- Push concrete wiring toward `main`/`cmd` and keep the import graph acyclic and flat ([Cheney](https://dave.cheney.net/2016/08/20/solid-go-design)).
- Inject dependencies through constructors, for example `NewService(repo QuestionStore, clock Clock) *Service`. No package-level mutable singletons and no DI container (226os convention; DIP per [Cheney](https://dave.cheney.net/2016/08/20/solid-go-design)).

**Naming**
- Name packages for what they provide. Avoid names like `util`, `common`, `misc`, `types` and `interfaces`. Don't stutter: `question.Store`, not `question.QuestionStore` ([CodeReviewComments](https://go.dev/wiki/CodeReviewComments#package-names)).
- Receiver names are short and consistent ([CodeReviewComments](https://go.dev/wiki/CodeReviewComments#receiver-names)).
- Domain names follow the SPEC vocabulary (question, answer, contribution need, application), not internal jargon (226os convention; `docs/SPEC.md:83-94`).

**Go errors**
- Never discard an error with `_` ([CodeReviewComments](https://go.dev/wiki/CodeReviewComments#handle-errors)).
- Error strings are lowercase with no trailing punctuation.
- Wrap with `%w` only when callers should inspect the cause. Use `%v` to avoid leaking an implementation detail such as `pgx.ErrNoRows` into a module's API ([Go blog: errors](https://go.dev/blog/go1.13-errors)).
- Repository errors become domain sentinels such as `ErrNotFound`, `ErrConflict` or `ErrForbidden`. The transport layer maps them, through one table, to problem+json `code` and `status` (226os convention; `docs/spec-kit/API_CONTRACT.md:57`).

**Hexagonal boundaries**
- The domain/service core must be drivable by HTTP, tests and batch jobs, isolated from the DB and devices ([Cockburn, Hexagonal](https://alistair.cockburn.us/hexagonal-architecture/)).
- Map hexagonal concepts onto the mandated folder names:
  - ports are interfaces declared in `service`
  - adapters are `repository` (sqlc/pgx) and `transport` (Chi)
- Don't add `ports/` or `adapters/` folders.

**KISS, YAGNI and DRY**
- YAGNI applies to speculative features, not to refactoring that makes code easier to change ([Fowler, Yagni](https://martinfowler.com/bliki/Yagni.html)).
- "Duplication is far cheaper than the wrong abstraction." Wait for the third real occurrence with the same reason to change before extracting. Inline a wrong abstraction rather than adding parameters to it ([Metz](https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction)).
- DRY applies to knowledge, not text. Two domains sharing a struct shape is not a reason to share a type across boundaries, because that couples their change cycles. Map at the boundary instead (226os convention; `docs/architecture/ARCHITECTURE.md:20`).

**Refactor step**
- A refactor changes internal structure without changing observable behaviour ([Fowler, Definition of Refactoring](https://martinfowler.com/bliki/DefinitionOfRefactoring.html)).
- Make it a separate step: tests stay green, the API and OpenAPI stay unchanged, ideally in a separate commit.

**React and TypeScript**
- Apply SRP through hooks and components:
  - custom hooks own data and effects ([react.dev custom hooks](https://react.dev/learn/reusing-logic-with-custom-hooks))
  - components own rendering
  - routes own wiring
- Prefer composition through children and slots over inheritance or configuration-heavy "god components".
- Avoid effects for derived state ([react.dev](https://react.dev/learn/you-might-not-need-an-effect)).
- Keep the `strict` TypeScript family on and ban `any` ([TSConfig](https://www.typescriptlang.org/tsconfig/#strict)).
- Use discriminated unions for UI state (`idle | loading | error | success`) instead of boolean flags ([TS handbook: narrowing](https://www.typescriptlang.org/docs/handbook/2/narrowing.html#discriminated-unions)).

**Review**
- Approve once the change clearly improves code health, even if it is imperfect. Correctness, security and contract compliance are never trade-offs ([Google eng-practices](https://google.github.io/eng-practices/review/reviewer/standard.html)).

## SOLID / design notes
| Principle | Go (src/api) | React/TS (src/web) |
|---|---|---|
| SRP | One reason to change per package. The handler parses, authenticates the session, calls the service and maps errors, and does nothing else | Split data hook, view and route |
| OCP | Add behaviour through a new type that implements a small consumer interface. Avoid flag parameters | Extend through composition and slots, not `variant` prop explosions |
| LSP | Fakes and real repos honour the same contract. Run shared contract tests against both, with the real-PG run being authoritative | Wrapped Astryx primitives keep native semantics (button stays button, keyboard, `ref`) |
| ISP | `QuestionReader` and `AnswerAccepter`, not a 20-method `Repository` | Narrow props. Don't pass whole DTOs when a component needs two fields |
| DIP | `service` declares the interfaces, `repository` implements them, and `cmd` wires them | Components depend on hooks, and hooks depend on the generated client |

- Tension: `docs/DEV.md:44` says "interfaces à la frontière", which is ambiguous. Read it as consumer-side: declared in `service`, implemented in `repository`/`transport`. Escalate if an ADR says otherwise.
- With 100% Go coverage, don't add defensive branches that can't be reached. Remove impossible states through types and constructors so that every branch stays testable.

## TDD & tests for this layer
- Start RED from a SPEC Given/When/Then that covers the nominal path, invalid data, an unauthorized actor, duplicates/idempotency, and DB or network failure (`docs/TEST.md:26`).
- Go tests:
  - use table-driven tests
  - test services with hand-written fakes of consumer interfaces
  - test repositories and invariants against real ephemeral PostgreSQL, not mocks (`docs/TEST.md:16`)
  - run with `-race` (`docs/DEV.md:46`)
- Contract: OpenAPI lint plus a generated-client drift check (`docs/spec-kit/API_CONTRACT.md:66`).
- Web: Vitest + Testing Library at the role/label level. Use Playwright for P0 journeys and axe for accessibility (`docs/DEV.md:55`, `docs/TEST.md:18`, `docs/TEST.md:21`). See `web-frontend`.
- Refactor: re-run the same suite. Coverage and the authorization matrix must not regress.

## Anti-patterns (reject in review)
- SQL, authorization decisions or invariants inside Chi handlers. Domain packages importing `net/http` or a neighbour's transport.
- Owner ID taken from the request body. Missing negative authorization tests.
- Producer-side interfaces, `Repository` god-interfaces, or interfaces with a single implementation and no consumer need.
- `util`, `common` or `helpers` packages. Global mutable state. `init()` wiring.
- Errors swallowed with `_`. `pgx` errors leaked across module APIs. Stack traces or SQL in responses. Secrets in logs.
- Missing context deadlines on DB or HTTP calls. String-built SQL.
- Speculative abstractions, empty P1/P2 scaffolding, or premature generic frameworks. DRY used to couple domains.
- "Refactor" commits that change behaviour or contracts. Tests edited or skipped to go green. Opportunistic coverage exclusions.
- An ORM, Redis, an external search engine or a new service without an ADR.
- AI agents as authors or `Co-authored-by`. Claims of PASS without command evidence.

## Done checklist
- [ ] Traced to a `US-*` story or `TASK-*`, inside its `allowed_paths`. Slice opened with `/start-slice`.
- [ ] RED evidence recorded, then GREEN, then a separate behaviour-preserving REFACTOR.
- [ ] Dependency direction checked: transport → service → domain, with repository implementing interfaces declared in service. No cross-module transport imports.
- [ ] Errors wrapped and mapped to problem+json. Nothing leaks, no secrets are logged, and contexts have deadlines.
- [ ] Coverage meets the threshold with no new exclusions outside `COVERAGE_SCOPE.md`. Authorization matrix and integration tests are present.
- [ ] No drift in OpenAPI, sqlc or generated clients. Any deviation from the architecture has an ADR.
- [ ] Review standards met. Human author identity used. Gates reported PASS, FAIL, NOT RUN or BLOCKED via `/verify-slice`.

## References
- Project: `docs/RULES.md`, `docs/DEV.md`, `docs/TEST.md`, `docs/SPEC.md`, `docs/architecture/ARCHITECTURE.md`, `docs/spec-kit/{API_CONTRACT,AUTHORIZATION_MATRIX,DATABASE_DESIGN,TASKS}.md`, `docs/quality/COVERAGE_SCOPE.md`, `AGENTS.md`, `CONTRIBUTING.md`.
- Web: [Go CodeReviewComments](https://go.dev/wiki/CodeReviewComments), [SOLID Go Design](https://dave.cheney.net/2016/08/20/solid-go-design), [Go 1.13 errors](https://go.dev/blog/go1.13-errors), [Hexagonal](https://alistair.cockburn.us/hexagonal-architecture/), [Yagni](https://martinfowler.com/bliki/Yagni.html), [Refactoring definition](https://martinfowler.com/bliki/DefinitionOfRefactoring.html), [Wrong abstraction](https://sandimetz.com/blog/2016/1/20/the-wrong-abstraction), [Google review standard](https://google.github.io/eng-practices/review/reviewer/standard.html), [react.dev custom hooks](https://react.dev/learn/reusing-logic-with-custom-hooks).
