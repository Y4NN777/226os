# ADR-0005 — Data model, migrations, and search

- **Status:** ACCEPTED
- **Date:** 2026-11-11
- **Approver:** `Y4NN777`
- **Amended:** 2026-11-14, same approver: point 10 follows ROADMAP C1 (project actions belong to the project owner only at P0).
- **Closes:** ADR-OPEN-005
- **Refines:** G0-DATA-001, G0-DATA-002, G0-DATA-003
- **Related:** `docs/spec-kit/DATABASE_DESIGN.md` (entity detail and DDL), [ADR-0003](0003-authentication-sessions-csrf-rate-limits.md), [ADR-0004](0004-api-contract-conventions.md), [ADR-0006](0006-toolchain-environments-quality-commands.md)
- **Metadata note 2026-11-12:** approver spelling normalised to the GitHub login (`Y4NN777`); the decision text is unchanged.

## Context

TASK-DATA-001 needs the migration tool, PostgreSQL version, ID strategy, search strategy, status model and Learn entities fixed. PRD amendment 3.1 makes Learn (articles, tutorials, a free training series) a P0 domain. `DATABASE_DESIGN.md` holds the table-level detail; this ADR records the decisions it implements.

## Options considered

- **Migration tool:** goose (chosen), golang-migrate, Atlas, tern. goose runs plain SQL, embeds in the Go binary and supports `NO TRANSACTION` for concurrent indexes.
- **PostgreSQL version:** 16, 17, **18 (chosen)**. 18 provides a native `uuidv7()`, so IDs need no Go-side generator and every client (including seeds and tests) gets time-ordered keys.
- **Primary keys:** UUIDv4, **UUIDv7 (chosen)**, integers. v7 is time-ordered, which keeps btree inserts local and keyset ordering cheap. Creation time is observable from the ID; accepted, because IDs are never authorization.
- **Search vectors:** generated stored columns (chosen), triggers, external engine (excluded).

## Decision

1. **Engine.** **PostgreSQL 18**, one image tag pinned by digest in every Compose profile; extensions `pg_trgm` and `unaccent` (the latter only through an immutable wrapper, see 6). No ORM; `pgx` v5 plus `sqlc`.
2. **Layout.** Root `db/`: `db/migrations/`, `db/queries/`, `db/seeds/`, `db/sqlc.yaml`. sqlc generates into a Go package under `src/api/internal/` (generated, drift-checked, never hand-edited). `docs/DEV.md` follows this layout (no `src/api/db`).
3. **Migrations.** goose, plain SQL, **forward-only**: no `Down` sections, CI rejects them. Sequential names `00001_<slug>.sql`. Each file is transactional by default; `CREATE INDEX CONCURRENTLY` uses `-- +goose NO TRANSACTION` in its own file. **Files are immutable after merge** (CI compares against the target branch). Every migration has a header: purpose, tables/indexes, lock risk, data volume, transaction behaviour, compatibility window, restore plan, owner. Live incompatible changes use expand/contract. Risky migrations need a verified `pg_dump -Fc` backup first; rollback is restore or redeploy of the previous image by digest, never an untested down. CI runs a fresh install and an upgrade from the previous schema.
4. **IDs and time.** Every UUID primary key is `uuid PRIMARY KEY DEFAULT uuidv7()`; seeds supply fixed UUIDs. `created_at` and `updated_at` are UTC `timestamptz`.
5. **Domains and integrity.** Separate tables per domain (identity, solve, discover, build, learn, relations, taxonomy). Invariants (unique normalized email, one accepted answer per question, one vote per user and target, unique application per user and need, normalized GitHub URL) are enforced by constraints and partial unique indexes, not application code alone. No cascading deletes on user-owned content.
6. **Search.** Stored generated `search_document tsvector` columns with weighted fields and GIN indexes on questions, answers, resources, projects and **Learn articles**; bounded `pg_trgm` GIN indexes on titles and names only. Configuration is **`simple` plus `unaccent` through an immutable wrapper function (`f_unaccent`), in P0** (French-first: `éléphant` matches `elephant`), applied identically to documents and queries. Queries are length-bounded (100 characters), result-bounded (50), stably ordered with a UUID tie-breaker. Content is never auto-translated. A bilingual relevance dataset and `EXPLAIN (ANALYZE, BUFFERS)` evidence are G6 work that may tune ranking, not reopen this choice.
7. **Status model.** Three separate things, never one column:
   - **Publication lifecycle** (column `status`): Learn articles `draft | published | archived | removed`; every other publishable entity (questions, answers, resources, providers, projects, training series, lessons) `published | removed`.
   - **Question solve state** is **derived**, not stored: the API exposes `state` = `open` (no answers), `answered` (has answers, none accepted) or `resolved` (`accepted_answer_id` set), computed from answers and `accepted_answer_id`.
   - **Project visibility**: `public | private` (a column, independent of `status`).
   - A moderator **hide** sets `status = 'removed'` (with `removed_at` and an `audit_log` row); **restore** sets it back and clears `removed_at`. Status vocabulary: `removed` is the publication status of content (not shown publicly, restorable by a moderator); `hidden` is the status of a moderated content relation or tag (a separate column with its own check, see decision 13).
8. **Votes.** Up or down on questions and answers (`value` +1 / -1). **Self-vote is forbidden**: 403 `SELF_VOTE_FORBIDDEN`, a service rule inside the vote transaction.
9. **Applications.** Unique per `(user, need)`. **Re-applying is allowed only if the previous application was withdrawn by the applicant**: the row is reused and reopened (`withdrawn -> pending`). **No re-apply after a decline.** `accepted` and `declined` are terminal at P0.
10. **"Collaborator".** A collaborator is **an applicant whose application was accepted on a project**: a resource-scoped relationship derived from `applications`, **not a global role**. It is not in any role list. `project_maintainers` stays in the schema for owner-added maintainers, but **grants nothing at P0**: project actions (edit, needs, list applications, accept or decline) belong to the project owner only, and the applicant may withdraw (ROADMAP C1, owner decision; amended 2026-11-14, the matrix in `docs/spec-kit/AUTHORIZATION_MATRIX.md` already follows it). Maintainer delegation is a P1 question.
11. **Tags.** Free-form, normalized to a lowercase slug of at most 32 characters, **at most 5 per item**. Moderators can merge and rename (the merge/rename operations are not part of the P0 moderator scope in ADR-0003 and ship later; until then the owner performs them by an audited operational step).
12. **Profile visibility defaults.** Name, username, skills, country, bio, GitHub username, `open_to_collab` and languages are **public by default**; each optional field keeps an owner-controlled visibility (`public | members | private`). **Email is never public** and is not a profile field.
13. **Content relations.** Types: question-resource, question-project, resource-project, article-question, article-resource, article-project. **Only the author of the source item creates a relation, only to published public targets; moderators can remove them** (`status = 'hidden'`, audited).
14. **Resources.** `resources.type` is `api | sdk | tool` (what the item is). The table formerly named `resource_sdks` is **`resource_sdk_languages`**: the languages in which the resource has an official SDK.
15. **Account deletion.** Credentials, sessions and profile personal data are deleted **within 30 days** of the request. Authored public content is kept and **anonymized as "membre supprimé"** (the account row stays with `status = 'deleted'`, the email is replaced by a placeholder, no username or display name is shown).
16. **Learn.** `learn_articles` (`kind` article | tutorial, language, level, tags), `training_series` and `training_lessons` (P0: one public free series readable without login) are P0. **Learn progress is P1**: `learn_progress` is created in a later migration, not in the bootstrap.
17. **Seeds.** Deterministic SQL in `db/seeds/`, fixed UUIDs, every root row `provenance = 'demo'` and visibly marked in the UI; no fake endorsements, verifications or real users. Seeds run **only in `local-dev` and `staging-test`**, never in `production-live`.

## Consequences

- **Security:** constraints block cross-tenant and duplicate states even if the service layer has a bug; SQL is parameterized only; trigram and query-length bounds limit expensive queries.
- **Data:** forward-only means mistakes are fixed by a new migration or restore, so backup discipline is mandatory. `uuidv7()` needs PostgreSQL 18 everywhere, including CI and local Compose.
- **Ops:** goose runs from a CLI or embedded step in each profile; a failed migration stops the deployment. Restore drills gate G7.
- **Cost:** none; one database, no search service.

## Evidence / verification required

Fresh and upgrade migration runs on real PostgreSQL 18; constraint tests for each invariant; no-`Down` and immutability CI checks; sqlc drift check; `f_unaccent` immutability and French accent-matching tests; seed provenance test; vote, application re-apply and relation tests; account-deletion test; relevance dataset with `EXPLAIN (ANALYZE, BUFFERS)` plans; restore drill. Gates G1, G3, G6, G7.
