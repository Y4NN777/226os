# 226os P0 PostgreSQL Schema and Migration Design

**Status:** ACCEPTED design (ADR-0005, 2026-11-11, approver `Y4NN777`). The DDL is owned by `db/migrations/00001` to `00012` (immutable once merged). A statement-by-statement diff of this document's former DDL against those migrations was done on 2026-11-12: identical for the 38 P0 tables (names, constraints, indexes); the only differences were the 3 P1 tables below (no migration) and the reference `INSERT`s in `00002`. This document keeps the rationale, the rules SQL cannot express, and the DDL of what has no migration. Items with no migration yet are marked **PLANNED**.
**Target engine:** **PostgreSQL 18** with native `uuidv7()`. **Tooling:** goose forward-only SQL migrations under root `db/migrations/`, sqlc queries under `db/queries/`, `pgx` v5.
**Decision owner:** `Y4NN777` (interim owner, ADR-0002). Migrations `00001` to `00012` run on a **fresh** PostgreSQL 18 database in the integration suite `src/api/internal/platform/pgtest` (`TestFreshMigrationToHead`), and `TestUpgradeFromVersion10KeepsExistingRows` applies `00011` and `00012` on top of version 10. Both run under `make test-integration`, which `make quality` and CI include. sqlc queries exist for identity, platform and reference data only (`db/queries/`).

References to other documents cite section names or IDs, not line numbers.

## 1. Design principles and conventions

PostgreSQL remains the source of truth and initial search engine (BASE-005). Tables stay separated by domain; cross-domain links use typed relations rather than a universal content table. Every ownership and visibility decision must be enforceable in SQL or a transactionally adjacent authorization check.

| Topic | Decision | Justification |
|---|---|---|
| Primary keys | `uuid PRIMARY KEY DEFAULT uuidv7()` (native in PostgreSQL 18) | v7 is time-ordered, so btree inserts stay local and `(created_at, id)` keyset order is cheap, and every client (application, tests, seeds) gets the same default. Cost: creation time is observable from the ID; accepted because `created_at` is public on most content and IDs are never authorization. Seeds use fixed UUIDs. |
| Exceptions | `audit_log.id` is `bigint GENERATED ALWAYS AS IDENTITY` (append-only, never exposed through the API). Reference tables use stable text codes (`countries.code`, etc.). Pure join tables use composite primary keys. | Opaque IDs are only useful for objects addressable through the API. |
| Time | `timestamptz` everywhere (stored UTC, serialized RFC 3339 UTC, `SPEC section 4`). `created_at` immutable; `updated_at` set by the sqlc query (`updated_at = now()`), **not by trigger**, to avoid hidden mutation. | Matches the existing trigger-avoidance stance on search vectors. |
| Status | `text` + named `CHECK (... IN (...))`, not PG `ENUM`, so a value is added by a forward-only `ALTER TABLE ... DROP/ADD CONSTRAINT` without enum-type surgery. **Publication lifecycle** is the `status` column: `learn_articles` `draft/published/archived/removed`, every other publishable entity `published/removed`. **Question solve state** is not stored: the API derives `state` = `open` (no answers), `answered` (answers, none accepted), `resolved` (`accepted_answer_id` set) from `answer_count` and `accepted_answer_id`. **Project visibility** is its own column (`public/private`). | Business-driven, evolving sets; three concerns, three mechanisms. |
| Language | `language text CHECK (language IN ('fr','en'))`, always `NOT NULL` on user-authored content (original language is preserved and displayed, never translated: `SPEC US-LEARN-001`, `SPEC US-LEARN-001`). A `DOMAIN` was rejected because sqlc support for `CREATE DOMAIN` is unverified. | |
| Email | **No `citext`.** Store only the normalized form: `email text CHECK (email = lower(btrim(email)))` with `UNIQUE (email)`. The Go layer lower-cases, trims and applies Unicode NFC before insert; the CHECK is the backstop. | One representation, one unique index, no extension; satisfies `SPEC section 5`. Display-as-typed casing is not needed in P0. |
| Delete policy | Soft delete (`status = 'removed'` + `removed_at`) for all user content and entities that other rows reference; hard delete only for credentials/tokens/sessions/idempotency/rate-limit rows and for pure join rows. See the table in 1.1. | Relations, votes and applications must not dangle (`SPEC section 5`); privacy erasure is done by anonymization, not row deletion. |
| `ON DELETE` | `RESTRICT` for every FK to content or users. `CASCADE` only on (a) auth artefacts (`user_credentials`, `sessions`, token tables, `idempotency_keys`) and (b) pure join/child rows owned by one parent (`question_tags`, `resource_countries`, ...). No `SET NULL` except where a nullable column is documented below. | No cascading delete of user-owned content (`DATABASE_DESIGN` principle). |
| Constraint names | Every `CHECK`, `UNIQUE`, `FK` is named `<table>_<what>_<ck|uq|fk>`; repositories map SQLSTATE + constraint name to domain errors. | `.claude/skills/postgres-data/SKILL.md`. |
| Bounds | Every free-text column has `char_length` CHECK; Markdown bodies are capped; `jsonb` only for bounded audit/idempotency payloads. | Query cost and abuse limits (`SPEC section 7`, `SPEC section 7`). |
| Triggers | None in P0 (updated_at, search vectors, counters are explicit). | Auditable, testable behaviour. |
| Extensions | `pg_trgm` and `unaccent`. `unaccent` is used only through the immutable wrapper `f_unaccent(text)` (3.0) so it can appear in generated columns and indexes. | ADR-0005 points 1 and 6. |

### 1.1 Soft versus hard delete per entity

| Entity | Delete mode | Why |
|---|---|---|
| `users` | Soft, **within 30 days of the request**: `status='deleted'`, `deleted_at`, email replaced by `deleted-<uuid>@invalid`, credentials/sessions/tokens hard-deleted, profile personal data scrubbed (username replaced by `deleted-<8 hex>`). Authored public content is **kept and shown as "membre supprimé"** (the API returns no username or display name for a deleted author). | Keeps answers/votes/relations referentially intact while honouring erasure requests (ADR-0005). |
| `questions`, `answers`, `resources`, `providers`, `projects`, `learn_articles`, `training_series`, `training_lessons` | Soft: `status='removed'`, `removed_at`; a moderator hide writes `audit_log` in the same transaction and a restore sets `status='published'` and clears `removed_at`. Hard delete only by a documented, approved privacy/legal procedure. | Referenced by relations, votes, applications, reports. |
| `contribution_needs` | Status `closed` (no removal column) | Needs are closed, not deleted; moderators hide the parent project. |
| `applications` | Soft: status `withdrawn`; never hard-deleted by users. A withdrawn row is reused when the applicant re-applies. | Preserves the unique `(user_id, contribution_need_id)` rule and the owner's audit trail. |
| `votes` | Hard delete (user retracts vote), score adjusted in the same transaction. | A vote has no independent meaning; `DELETE /votes` exists. |
| `content_relations` | Soft: `status='hidden'` (moderator removal, audited; restore sets `active`). | Moderation reversibility. |
| `sessions`, `email_verifications`, `password_resets`, `idempotency_keys`, `rate_limit_buckets` | Hard delete by expiry job. | Secrets/operational data; minimal retention. |
| join tables (`question_tags`, `user_skills`, `need_skills`, ...) | Hard delete (row is the relation). | |
| `audit_log` | Never deleted or updated by the application role (grants, see 3.8). | Moderation accountability (`AUTHORIZATION_MATRIX`). |

## 2. ERD

```mermaid
erDiagram
    users ||--|| user_credentials : "has secret"
    users ||--|| profiles : "has"
    users ||--o{ sessions : "opens"
    users ||--o{ email_verifications : "requests"
    users ||--o{ password_resets : "requests"
    users ||--o{ user_skills : "declares"
    tags ||--o{ user_skills : "skill"
    countries ||--o{ profiles : "country"

    users ||--o{ questions : "asks"
    questions ||--o{ answers : "has"
    questions |o--o| answers : "accepted answer (composite FK)"
    users ||--o{ answers : "writes"
    users ||--o{ votes : "casts"
    questions ||--o{ votes : "target"
    answers ||--o{ votes : "target"
    questions ||--o{ question_tags : "tagged"
    tags ||--o{ question_tags : "tag"
    questions ||--o{ question_sources : "cites"
    answers ||--o{ answer_sources : "cites"
    countries ||--o{ questions : "context"

    providers ||--o{ resources : "offers"
    resource_categories ||--o{ resources : "classifies"
    users ||--o{ resources : "creates"
    resources ||--o{ resource_countries : "supports"
    countries ||--o{ resource_countries : "country"
    resources ||--o{ resource_sdk_languages : "has official SDK in"
    resources ||--o{ resource_sources : "sourced by"
    resource_sources |o--o| resources : "verified_source (composite FK)"
    resources ||--o{ resource_tags : "tagged"
    tags ||--o{ resource_tags : "tag"
    resources ||--o{ error_reports : "reported"
    users ||--o{ error_reports : "reports"
    resources ||--o{ resource_reviews : "reviewed P1"
    users ||--o{ resource_reviews : "writes P1"

    users ||--o{ projects : "owns"
    projects ||--o{ project_maintainers : "delegates"
    users ||--o{ project_maintainers : "maintains"
    projects ||--o{ project_tags : "stack"
    tags ||--o{ project_tags : "tag"
    projects ||--o{ contribution_needs : "needs"
    contribution_needs ||--o{ need_skills : "requires"
    tags ||--o{ need_skills : "skill"
    contribution_needs ||--o{ applications : "receives"
    users ||--o{ applications : "applies"

    learn_topics ||--o{ learn_articles : "topic"
    users ||--o{ learn_articles : "authors"
    learn_articles ||--o{ learn_article_tags : "tagged"
    tags ||--o{ learn_article_tags : "tag"
    users ||--o{ training_series : "authors"
    learn_topics ||--o{ training_series : "topic"
    training_series ||--o{ training_lessons : "ordered lessons"
    learn_articles |o--o{ training_lessons : "optional article"
    users ||--o{ learn_progress : "tracks P1"
    training_lessons ||--o{ learn_progress : "completed P1"

    questions ||--o{ content_relations : "src/tgt"
    resources ||--o{ content_relations : "src/tgt"
    projects ||--o{ content_relations : "src/tgt"
    learn_articles ||--o{ content_relations : "src/tgt"
    users ||--o{ content_relations : "creates"
    users ||--o{ content_reports : "reports P1"
    users ||--o{ audit_log : "acts"
    users ||--o{ idempotency_keys : "retries"

    users { uuid id PK "email normalized unique" }
    profiles { uuid user_id PK "username unique" }
    sessions { uuid id PK "token_hash unique" }
    questions { uuid id PK "accepted_answer_id composite FK" }
    answers { uuid id PK "unique id+question_id" }
    votes { uuid user_id PK "user+target_type+target_id" }
    resources { uuid id PK "verified_source_id composite FK" }
    projects { uuid id PK "github_url unique partial" }
    applications { uuid id PK "unique user+need" }
    learn_articles { uuid id PK "kind article or tutorial" }
    content_relations { uuid id PK "typed src/tgt FKs" }
```

### 2.1 Table inventory (authoritative, 41 tables: 38 in the P0 bootstrap, 3 deferred to P1)

| Domain | Tables | Count | P1 (deferred) |
|---|---|---|---|
| Reference / taxonomy | `countries`, `tags`, `learn_topics` | 3 | - |
| Identity | `users`, `user_credentials`, `profiles`, `user_skills`, `sessions`, `email_verifications`, `password_resets`, `rate_limit_buckets` | 8 | - |
| Solve | `questions`, `answers`, `votes`, `question_tags`, `question_sources`, `answer_sources` | 6 | - |
| Discover | `resource_categories`, `providers`, `resources`, `resource_countries`, `resource_sdk_languages`, `resource_sources`, `resource_tags`, `error_reports`, `resource_reviews` | 9 | `resource_reviews` |
| Build | `projects`, `project_maintainers`, `project_tags`, `contribution_needs`, `need_skills`, `applications` | 6 | - |
| Learn | `learn_articles`, `learn_article_tags`, `training_series`, `training_lessons`, `learn_progress` | 5 | `learn_progress` |
| Cross-cutting | `content_relations`, `content_reports`, `audit_log`, `idempotency_keys` | 4 | `content_reports` (no P0 endpoint; moderators resolve error reports only) |
| **Total** | | **41** | **3** |

### 2.2 Why `content_relations` is not polymorphic

A classic `(source_type text, source_id uuid, target_type text, target_id uuid)` table cannot carry foreign keys, so dangling links and cross-type confusion are possible (`SPEC section 5` requires allowed source/target types and no pending links; `.claude/skills/postgres-data/SKILL.md`). Per-pair tables (6 unordered pairs) would multiply tables and queries. The design is **one table with one nullable real FK column per entity per side** (4 x 2 columns), a `CHECK` that exactly one column per side is set and agrees with the declared type, a `CHECK` that `(source_type, target_type)` is one of six canonical cross-type pairs (so A-B and B-A cannot both exist), and generated `source_id`/`target_id` columns for a single uniqueness constraint. Referential integrity is therefore enforced by PostgreSQL; "target is published and public" cannot be a column check and is enforced in the creating transaction (`SELECT ... FOR SHARE` on the target) and in every read query (join with `status = 'published'` and visibility). Adding a fifth type (e.g. `training_series`) is an expand-only `ADD COLUMN` plus a new CHECK.

## 3. Table notes (DDL lives in the migrations)

Each migration orders tables so that every FK target already exists, except the two documented composite FKs added with `ALTER TABLE` (`questions` to `answers`, `resources` to `resource_sources`). Primary keys default to `uuidv7()`; seeds supply fixed UUIDs.

### 3.0 Extensions and the `unaccent` wrapper (`00001`)

`unaccent()` is `STABLE`, so it cannot be used in generated columns or expression indexes. The accepted fix is an **immutable wrapper** with a schema-qualified dictionary, used identically for documents and queries (French-first accent folding, ADR-0005 point 6).

DDL: see `db/migrations/00001_extensions.sql`.

(Test: `f_unaccent('éléphant') = 'elephant'`, and creating a generated column with `f_unaccent` succeeds. Both are G3 checks.)

### 3.1 Reference and taxonomy

**`countries`** - stable ISO 3166-1 alpha-2 codes with separate FR/EN labels (labels localized separately from codes). Enforces: validated country (`SPEC section 5`). Reference data (not demo data) is inserted by the migration.

DDL: see `db/migrations/00002_reference_taxonomy.sql`.

**`tags`** - one free-form slug namespace shared by questions, resources, projects, articles and user skills/technologies. Any member may create a tag when publishing; the slug is normalized to lowercase and at most 32 characters (the `TagSlug` pattern of the OpenAPI), **at most 5 tags per item** (service rule), and moderators can merge or rename (P1 operation; `status = 'hidden'` is the P0 moderator tool). Enforces: unique normalized identifiers.

DDL: see `db/migrations/00002_reference_taxonomy.sql`.

**`learn_topics`** - small reference list for Learn so new topics are data, not constraint changes. The current rows are the 9 owner-approved topics inserted by `db/migrations/00002_reference_taxonomy.sql:112-121` (`ai`, `payments`, `ussd-messaging`, `mobile`, `api-integration`, `security-privacy`, `cloud-devops`, `geo-data`, `open-source`); the prototype list that used to be cited here is superseded.

**PLANNED (E2, owner decision 2026-11-12; new migration, not in 00002, which is immutable):** one new forward-only migration, numbered by `make migrate-new` and written by the single migration writer (`ROADMAP.md` R6), adds 5 topics (`learn-to-learn`, `ai-assisted-dev`, `ai-agents`, `open-models`, `african-languages`) and relabels `ai`, `mobile` and `open-source` (same codes, new `name_fr`/`name_en`), for **14 topics**. No row is deleted; removals are deferred and need explicit human approval. The exact labels are in `TODO.md`, section E2. The same PR updates the exact lists and counts in `src/api/internal/platform/pgtest/migrate_integration_test.go`.

DDL: see `db/migrations/00002_reference_taxonomy.sql`.

### 3.2 Identity (`00003`)

**`users`** - account identity and lifecycle. Enforces: unique normalized email (`SPEC section 5`); bounded status; moderator role is data (granted by the repository owner through an audited operational step), never a frontend flag. There is no `collaborator` role: collaboration is derived from accepted `applications` and `project_maintainers`.

DDL: see `db/migrations/00003_identity.sql`.

**`user_credentials`** - Argon2id password hash kept out of `users` so profile/session queries never select it (`G0 register (archived)`, `AUTHORIZATION_MATRIX`). CASCADE: it is a secret artefact, not content.

DDL: see `db/migrations/00003_identity.sql`.

**`profiles`** - one public profile per user with per-field visibility. **Defaults: display name, username, skills, country, bio, GitHub username, `open_to_collab` and languages are public**; the owner can narrow each optional field to `members` or `private`. **Email is never public** and is not a profile field. Enforces: profile belongs to user; unique username; validated country/languages; explicit visibility.

DDL: see `db/migrations/00003_identity.sql`.

**`profiles.profile_visibility` (C14b, decided by the owner and the main session on 2026-11-12; `db/migrations/00012_profile_visibility.sql`, not in 00003, which is immutable).** The accepted DDL of 00003 is unchanged. `00012` adds:

- `profile_visibility text NOT NULL DEFAULT 'public'` on `profiles`, with a named `CHECK (profile_visibility IN ('public','members','private'))`;
- in the same PR, the pgtest tests that assert table structure and constraints.

Semantics (`ROADMAP.md` section 10, C14):

- `GET /profiles/{username}`: `public` is visible to everyone, `members` to authenticated members only, `private` to the owner only. Everyone else gets a 404 whose body is identical to an unknown or deleted username. Moderators get no special profile access.
- Author bylines on published content keep showing the username and display name whatever `profile_visibility` says; deleted accounts still show "membre supprimé" (1.1).
- People directory (E5): a profile is listed for a caller only if the user is `active`, `profile_visibility` allows that caller, `open_to_collab = true` and `open_to_collab_visibility` allows that caller. Listed fields and filters use only fields visible to the caller. Email is never selected.
- Not decided: how suspended users are treated by `GET /profiles/{username}`.

Once this migration merges, decision 8 in section 10 ("only `projects` and profile fields carry visibility") also covers this profile-level column.

**`user_skills`** - user skills/technologies as tags (feeds the skills-based home feed and project needs matching). Join row: CASCADE from either side.

DDL: see `db/migrations/00003_identity.sql`.

**`sessions`** - opaque server-side sessions; only the SHA-256 of the 256-bit token and of the CSRF synchronizer token are stored. Enforces: session hash unique; idle and absolute expiry; revoked/expired sessions unusable (ADR-0003; rotation = new row + old row revoked with reason). IP is a truncated network prefix and the user agent a short parsed summary, no raw values.

DDL: see `db/migrations/00003_identity.sql`.

A session is usable iff `revoked_at IS NULL AND now() < LEAST(idle_expires_at, absolute_expires_at)`; this predicate lives in one sqlc query and is tested. There is no `session_events` table: the revocation reason is a column and security events go to `audit_log` (decided, section 10).

**`email_verifications`** - hashed, single-use, 24 h verification tokens (ADR-0003). A user with `email_verified_at IS NULL` may sign in, read and draft; publishing needs a verified email (service rule, 403 `EMAIL_NOT_VERIFIED`).

DDL: see `db/migrations/00003_identity.sql`.

Single use is enforced by `UPDATE ... SET consumed_at = now() WHERE token_hash = $1 AND consumed_at IS NULL AND expires_at > now() RETURNING user_id`; zero rows means invalid/used/expired (one generic error, no enumeration).

**`password_resets`** - same shape, 30 minute lifetime (ADR-0003), no `email` column.

DDL: see `db/migrations/00003_identity.sql`.

**`rate_limit_buckets`** - PostgreSQL-backed fixed-window counters, no new dependency (ADR-0003). Keys are HMACs of the IP/account so no plaintext personal data is stored. `UNLOGGED` (rebuildable, avoids WAL cost on a hot table); **accepted** (ADR-0003): a crash resets counters, a documented residual risk.

DDL: see `db/migrations/00004_platform_support.sql`.

Increment: `INSERT ... ON CONFLICT (scope, key_hash, window_start) DO UPDATE SET hits = rate_limit_buckets.hits + 1 RETURNING hits`.

### 3.3 Solve (`00005`)

**`questions`** - a technical question. Enforces: owner required, bounded state (`SPEC section 5`); at most one accepted answer per question **by construction** (a single `accepted_answer_id` column) and that the accepted answer belongs to the same question (composite FK, added after `answers` exists) (`SPEC US-SOLVE-001`, `SPEC section 5`); searchable (`SPEC`).

DDL: see `db/migrations/00005_solve.sql`.

The API `state` is derived: `resolved` when `accepted_answer_id IS NOT NULL`, else `answered` when `answer_count > 0`, else `open`. (Decided: the unused `questions_id_author_uq` constraint is dropped from the bootstrap migration.)

**`answers`** - an answer. Enforces: owner required, bounded state (`SPEC section 5`); `UNIQUE (id, question_id)` is the target of the accepted-answer composite FK.

DDL: see `db/migrations/00005_solve.sql`.

`MATCH SIMPLE` means the FK is not checked while `accepted_answer_id IS NULL`. A second acceptance is an `UPDATE` of the same column (at most one value), so "two accepted answers" is unrepresentable; acceptance by a non-author, acceptance of a `removed` answer, and clearing on answer removal are service rules in one transaction with `SELECT ... FOR UPDATE` on the question (the DB cannot see the caller). Concurrent accepts serialize on the row lock. Alternative (partial unique index `ON answers(question_id) WHERE accepted`) rejected: two sources of truth.

**`votes`** - one vote per user per target with real FKs for both target kinds. Enforces: unique `(user_id, target_type, target_id)` (`SPEC section 5`) literally, via a generated `target_id`; no dangling target; counters never taken from the client (`SPEC section 5`).

DDL: see `db/migrations/00005_solve.sql`.

`vote_score` on questions/answers is maintained by the repository in the same transaction as the vote insert/flip/delete (`INSERT ... ON CONFLICT (user_id, target_type, target_id) DO UPDATE`), with a reconciliation test comparing it to `SUM(value)`. Up and down votes are both allowed (`value` +1 / -1). **Self-vote is forbidden** (403 `SELF_VOTE_FORBIDDEN`): the author of the target cannot vote on it, a service rule checked in the vote transaction (a CHECK cannot cross tables).

**`question_tags`**, **`question_sources`**, **`answer_sources`** - tags and cited links (answers are "explained and sourced", PRD pitch). URLs are `https?` only and length-bounded; they are rendered as links, never fetched in P0 (anti-SSRF, `SPEC section 7`).

DDL: see `db/migrations/00005_solve.sql`.

"At most 5 tags per question" is a service rule (decided, ADR-0005).

### 3.4 Discover (`00006`)

**`resource_categories`** - controlled category codes with FR/EN labels. The current rows are the 15 owner-approved categories inserted by `db/migrations/00002_reference_taxonomy.sql:124-139`, `sort_order` in steps of 10 (`mobile-money` through `dev-tools`).

**PLANNED (E2, owner decision 2026-11-12; new migration, not in 00002, which is immutable):** the same E2 migration as for `learn_topics` adds `ai-models`, `ai-tooling` and `ai-datasets` (each with an explicit `sort_order`) and relabels `cloud-hosting`, for **18 categories**. `mobile-money` and `sms-ussd` stay (differentiator 1). No row is deleted.

DDL: see `db/migrations/00002_reference_taxonomy.sql`.

**`providers`** - organisation behind one or more resources. Enforces: no unsupported provider claim: a provider carries no verification flag, only `resources`/`resource_sources` do (`SPEC section 5`, `AUTHORIZATION_MATRIX`).

DDL: see `db/migrations/00006_discover.sql`.

**`resources`** - API, SDK or tool catalogue entry. Enforces: category and country links (`SPEC section 5`, `SPEC US-DISCOVER-001`); `verified` is representable **only with a source row of that resource and a verification date**, via composite FK `(verified_source_id, id)` to `resource_sources(id, resource_id)` (`SPEC section 5`, `SPEC US-DISCOVER-001`); staleness is computed at read time from `resource_sources.checked_at` (not stored), so a stale source is never shown as verified merely because it ranks (search design).

DDL: see `db/migrations/00006_discover.sql`.

**`resource_sources`** - provenance links with publication and last-check dates. Enforces: source + URL + verification date when declared (`SPEC section 5`, `SPEC US-DISCOVER-001`).

DDL: see `db/migrations/00006_discover.sql`.

`kind = 'claim'` (provider claim, not verified) is the only kind allowed without `checked_at`; the service additionally forbids pointing `verified_source_id` at a `claim` source (decided, section 10 item 19: a partial composite FK cannot express this).

**`resource_countries`**, **`resource_tags`**, **`resource_sdk_languages`** - supported countries, tags, and the languages in which the resource has an official SDK (with the package/repository link). `resources.type` says what the item is (`api`, `sdk` or `tool`); `resource_sdk_languages` says in which languages an item offers an SDK, so the two do not overlap. Enforces: country relation unique, no dangling country (`SPEC section 5`, `SPEC`).

DDL: see `db/migrations/00006_discover.sql`.

**`error_reports`** - "outdated / broken API" signalling (PRD Discover). One open report per reporter/resource/kind prevents spam; resolution fields stay consistent. **Moderators resolve (or reject) error reports** at P0, audited in the same transaction.

DDL: see `db/migrations/00006_discover.sql`.

**`resource_reviews`** (**P1**, minimal; not in the initial bootstrap) - one review per user per resource.

```sql
CREATE TABLE resource_reviews (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  resource_id uuid NOT NULL,
  user_id     uuid NOT NULL,
  rating      smallint NOT NULL,
  body_md     text,
  language    text,
  status      text NOT NULL DEFAULT 'published',
  created_at  timestamptz NOT NULL DEFAULT now(),
  updated_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT resource_reviews_resource_fk FOREIGN KEY (resource_id) REFERENCES resources (id) ON DELETE RESTRICT,
  CONSTRAINT resource_reviews_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT resource_reviews_user_resource_uq UNIQUE (user_id, resource_id),
  CONSTRAINT resource_reviews_rating_ck CHECK (rating BETWEEN 1 AND 5),
  CONSTRAINT resource_reviews_body_ck CHECK (body_md IS NULL OR char_length(body_md) <= 2000),
  CONSTRAINT resource_reviews_language_ck CHECK ((body_md IS NULL) = (language IS NULL) AND (language IS NULL OR language IN ('fr','en'))),
  CONSTRAINT resource_reviews_status_ck CHECK (status IN ('published','removed'))
);
CREATE INDEX resource_reviews_resource_idx ON resource_reviews (resource_id, created_at DESC, id DESC) WHERE status = 'published';
```

### 3.5 Build (`00007`)

**`projects`** - an open-source project or idea. Enforces: normalized GitHub URL unique among non-removed projects, internal owner required, explicit visibility (`SPEC section 5`); bounded stage/status. The URL is stored **already normalized** (`https://github.com/<owner>/<repo>`, lower-case, no trailing slash, no `.git`, no query/fragment); the Go normalizer is the single place of the rule and the CHECK is the backstop. Resolving the repo against the GitHub API (SSRF-guarded, quota-bound: `SPEC section 7`) is application behaviour, never a DB concern.

DDL: see `db/migrations/00007_build.sql`.

**`project_maintainers`** - maintainers added by the project owner, with **explicit rights on that project: edit the project and manage its needs and applications**. The owner stays `projects.owner_id`. A **collaborator** is not a role and not a row here: it is an applicant whose `applications` row is `accepted` on that project (a resource-scoped relationship derived from `applications`).

DDL: see `db/migrations/00007_build.sql`.

"Maintainer is not the owner" is a service rule (a CHECK cannot cross tables).

**`project_tags`** (technologies/topics) - join row.

DDL: see `db/migrations/00007_build.sql`.

**`contribution_needs`** - a concrete contribution need (the SPEC/OpenAPI name; "project_needs" in the brief). Enforces: existing project (FK), bounded state, skills linked (`SPEC section 5`).

DDL: see `db/migrations/00007_build.sql`.

**`applications`** - a member's request to contribute. Enforces: unique `(user_id, contribution_need_id)` so a second click or retry never creates a second row (`SPEC US-BUILD-001`, `SPEC section 5`, `AUTHORIZATION_MATRIX`); access restriction is authorization, not schema (`SPEC`, `AUTHORIZATION_MATRIX`); decision fields consistent with status.

DDL: see `db/migrations/00007_build.sql`.

Transitions are guarded in the service with `UPDATE applications SET status = $new ... WHERE id = $1 AND status = $expected` (zero rows = conflict `409`): `pending -> accepted | declined | withdrawn`; `withdrawn -> pending` (the applicant re-applies: the row is reused and reopened, `withdrawn_at` cleared and the message replaced; allowed **only** when the previous application was withdrawn by the applicant); `accepted` and `declined` are terminal, so there is **no re-apply after a decline**. An applicant whose application is `accepted` is the project's "collaborator" (a relationship, not a role). A transition trigger is deliberately not used (decided, section 10 item 13). "Applicant is not the project owner" and "need is open" are service rules inside the creating transaction.

### 3.6 Learn (`00008`, progress in a P1 migration)

Table names follow ADR-0005 point 7 (`learn_articles`, `training_series`, `training_lessons`, `learn_progress`). Enforces: public read and authenticated publish; publishing needs a verified email (`SPEC US-LEARN-001`), mandatory original language (`SPEC US-LEARN-001`), validated level/language (`SPEC US-LEARN-001`), searchable articles (`SPEC US-LEARN-001`).

**`learn_articles`** - article or tutorial.

DDL: see `db/migrations/00008_learn.sql`.

(`learn_articles_topic_fk_idx` duplicates the leading column of `learn_articles_topic_idx`; drop it in review if the topic index stays. Kept here only to make the "every FK is indexed" check visible.)

**`training_series`** - a free training series (P0: one public series readable without login, PRD 5.9). `origin = 'platform'` marks the founder/team series with the "Formation gratuite" label; `author_id` is always a real account (accountability), so the label is data, not a claim by absence of author.

DDL: see `db/migrations/00008_learn.sql`.

**`training_lessons`** - ordered lessons. Order is unique per series; the constraint is `DEFERRABLE INITIALLY DEFERRED` so a reorder can swap positions in one transaction. A lesson has its own Markdown body and may additionally reference a published article/tutorial (ADR-0005 point 7).

DDL: see `db/migrations/00008_learn.sql`.

(`UNIQUE (series_id, position)` already provides the `series_id` FK index.) Note: a unique constraint that is `DEFERRABLE` cannot be used as an `ON CONFLICT` arbiter; no upsert on position is planned.

**`learn_progress`** (**P1**, created in a later migration, not in the P0 bootstrap per ADR-0005) - lesson completion per user.

```sql
CREATE TABLE learn_progress (
  user_id      uuid NOT NULL,
  lesson_id    uuid NOT NULL,
  completed_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT learn_progress_pk PRIMARY KEY (user_id, lesson_id),     -- unique user+lesson
  CONSTRAINT learn_progress_user_fk FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  CONSTRAINT learn_progress_lesson_fk FOREIGN KEY (lesson_id) REFERENCES training_lessons (id) ON DELETE RESTRICT
);
CREATE INDEX learn_progress_lesson_idx ON learn_progress (lesson_id, user_id);
```

### 3.7 Cross-cutting relations and moderation (`00009`)

**`content_relations`** - typed links between question, resource, project and article, with real FKs (rationale in 2.2). The accepted matrix is question-resource, question-project, resource-project and article-question/resource/project; the table stores each pair in one canonical orientation (the pair CHECK below), and `created_by` must be the **author (or owner) of the item the creator was editing**, who may link **only to published public targets**; the item's author is the only member who creates a relation, and **moderators can remove it** (`status = 'hidden'`, audited). Enforces: allowed source/target types and no dangling links (`SPEC section 5`); "no pending/private target exposure" is the read rule plus the creating transaction.

DDL: see `db/migrations/00009_content_relations.sql`.

Canonical orientation means a question's relations to a resource are stored once as `question -> resource` and read from either side with the `src_*`/`tgt_*` indexes. Typical read for a question: `WHERE src_question_id = $1 AND status = 'active'` joined to published, public targets. (The `content_relations_target_idx` and the eight per-FK indexes overlap; review after EXPLAIN evidence and drop the redundant ones.)

**`content_reports`** (**P1**, not in the P0 bootstrap: no P0 endpoint creates it, and the P0 moderator scope covers error reports only) - member reports of abusive/incorrect content for moderators. Same typed-FK pattern, with `answers` added.

```sql
CREATE TABLE content_reports (
  id          uuid PRIMARY KEY DEFAULT uuidv7(),
  reporter_id uuid NOT NULL,
  target_type text NOT NULL,
  question_id uuid,
  answer_id   uuid,
  resource_id uuid,
  project_id  uuid,
  article_id  uuid,
  target_id   uuid GENERATED ALWAYS AS (COALESCE(question_id, answer_id, resource_id, project_id, article_id)) STORED NOT NULL,
  reason      text NOT NULL,
  details     text,
  status      text NOT NULL DEFAULT 'open',
  handled_by  uuid,
  handled_at  timestamptz,
  created_at  timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT content_reports_reporter_fk FOREIGN KEY (reporter_id) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT content_reports_handled_by_fk FOREIGN KEY (handled_by) REFERENCES users (id) ON DELETE RESTRICT,
  CONSTRAINT content_reports_question_fk FOREIGN KEY (question_id) REFERENCES questions (id) ON DELETE RESTRICT,
  CONSTRAINT content_reports_answer_fk   FOREIGN KEY (answer_id)   REFERENCES answers (id) ON DELETE RESTRICT,
  CONSTRAINT content_reports_resource_fk FOREIGN KEY (resource_id) REFERENCES resources (id) ON DELETE RESTRICT,
  CONSTRAINT content_reports_project_fk  FOREIGN KEY (project_id)  REFERENCES projects (id) ON DELETE RESTRICT,
  CONSTRAINT content_reports_article_fk  FOREIGN KEY (article_id)  REFERENCES learn_articles (id) ON DELETE RESTRICT,
  CONSTRAINT content_reports_target_ck CHECK (
    num_nonnulls(question_id, answer_id, resource_id, project_id, article_id) = 1
    AND (target_type = 'question') = (question_id IS NOT NULL)
    AND (target_type = 'answer')   = (answer_id IS NOT NULL)
    AND (target_type = 'resource') = (resource_id IS NOT NULL)
    AND (target_type = 'project')  = (project_id IS NOT NULL)
    AND (target_type = 'article')  = (article_id IS NOT NULL)),
  CONSTRAINT content_reports_reason_ck CHECK (reason IN ('spam','abuse','incorrect','copyright','other')),
  CONSTRAINT content_reports_details_ck CHECK (details IS NULL OR char_length(details) <= 2000),
  CONSTRAINT content_reports_status_ck CHECK (status IN ('open','actioned','dismissed')),
  CONSTRAINT content_reports_handled_ck CHECK ((status <> 'open') = (handled_at IS NOT NULL AND handled_by IS NOT NULL))
);
CREATE UNIQUE INDEX content_reports_open_uq ON content_reports (reporter_id, target_type, target_id) WHERE status = 'open';
CREATE INDEX content_reports_queue_idx ON content_reports (status, created_at, id);
CREATE INDEX content_reports_handled_by_idx ON content_reports (handled_by) WHERE handled_by IS NOT NULL;
CREATE INDEX content_reports_question_idx ON content_reports (question_id) WHERE question_id IS NOT NULL;
CREATE INDEX content_reports_answer_idx   ON content_reports (answer_id)   WHERE answer_id IS NOT NULL;
CREATE INDEX content_reports_resource_idx ON content_reports (resource_id) WHERE resource_id IS NOT NULL;
CREATE INDEX content_reports_project_idx  ON content_reports (project_id)  WHERE project_id IS NOT NULL;
CREATE INDEX content_reports_article_idx  ON content_reports (article_id)  WHERE article_id IS NOT NULL;
```
Reporter identity is visible only to moderators (reporter privacy).

### 3.8 Platform support (`00004`)

**`audit_log`** - append-only moderation/security trail. Enforces: every moderator action (hide, restore, error-report resolution, relation removal) is audited **in the same transaction** as the action and never inferred from a UI flag (ADR-0003). `target_id` has **no FK by design** (an audit row must survive any future hard deletion and may describe objects of several tables); `target_type` is CHECK-bounded. Denied-access attempts go to structured application logs without personal data (`SPEC`), not to this table, to bound write volume (decided: denied attempts go to logs).

DDL: see `db/migrations/00004_platform_support.sql`.

Append-only is enforced by role grants (the application role gets `INSERT, SELECT` only), to be defined with ops in ADR-0006 scope; not provable by schema alone. `details` must never contain secrets, tokens or message bodies (tested).

**`idempotency_keys`** - safe retries for the six create operations that declare the `Idempotency-Key` parameter in `api/openapi.yaml` (createQuestion, createAnswer, createProject, createContributionNeed, createApplication, createLearnArticle; ADR-0004). Votes and relations are not idempotency-keyed: they are idempotent through their unique constraints. The key is bound to subject; route and request hash are stored and compared so reuse with a different request is rejected (`API_CONTRACT`).

DDL: see `db/migrations/00004_platform_support.sql`.

Claim pattern: `INSERT ... ON CONFLICT (user_id, idempotency_key) DO NOTHING RETURNING ...`; on conflict read the row: same hash and `completed` replays the stored response, same hash `in_progress` returns `409`, different hash returns `422`. Idempotency complements, and never replaces, the unique constraints (`applications_user_need_uq`; `votes_pk` covers votes on its own).

### 3.9 Search indexes (`00010`)

All search indexes are **partial on `status = 'published'`** (plus `visibility = 'public'` for projects, the only content table with a visibility column) so unpublished or private content is never in the index and queries must carry the predicate (visibility rules in `AUTHORIZATION_MATRIX`). The configuration is always the two-argument `'simple'` form; accent handling for French is decided (section 10 item 14: `simple` plus `f_unaccent`). Indexes: a `search_document` GIN per searchable table (questions, answers, resources, projects, Learn articles) and a `f_unaccent(title|name) gin_trgm_ops` index on questions, resources, providers, projects and Learn articles, all with the partial predicate; the answers GIN is optional (see 9). `pg_trgm` GIN is used only on bounded fields (titles, names, slugs <= 200 characters); the application enforces a minimum trigram query length of 3 and a maximum query length (100).

DDL: see `db/migrations/00010_search_indexes.sql`.

GIN trigram indexes serve `%`, `ILIKE` and `~`; they cannot serve `<->` `ORDER BY ... LIMIT` (that needs GiST), so typo tolerance is `WHERE title % $q ORDER BY similarity(title, $q) DESC LIMIT n` over a small candidate set, or an `ILIKE` fallback when full-text returns too few rows.

### 3.10 FK self-check

- The 2026-11-12 diff against `db/migrations/` found the 38 P0 tables, constraint names and indexes identical to the former DDL; the three P1 tables below are the only ones without a migration.
- Circular references (`questions` <-> `answers`, `resources` <-> `resource_sources`, `tags` -> `users`) are broken with nullable columns plus `ALTER TABLE ... ADD CONSTRAINT` in the same migration file; neither needs `DEFERRABLE`.
- Every FK column has an index (leading column of a PK/UNIQUE or an explicit index). Review the FK-index list again once EXPLAIN evidence is captured.

## 4. Invariants enforced in the database

| Invariant | Mechanism | Constraint | Source |
|---|---|---|---|
| Unique normalized email | stored lower-cased + CHECK + UNIQUE | `users_email_uq`, `users_email_normalized_ck` | `SPEC section 5` |
| Profile belongs to one user; unique username | PK = FK | `profiles_user_fk`, `profiles_username_uq` | `SPEC section 5` |
| Validated country and language | FK to `countries`; CHECK on `language` | `*_country_fk`, `*_language_ck` | `SPEC section 5` |
| Session hash unique; revoked/expired unusable | UNIQUE hash + CHECKs + usability predicate | `sessions_token_hash_uq`, `sessions_expiry_ck` | `AUTHORIZATION_MATRIX` |
| Question/answer owner required; bounded state | `NOT NULL` FK + CHECK | `questions_author_fk`, `questions_status_ck`, `answers_author_fk` | `SPEC section 5` |
| At most one accepted answer per question; belongs to that question | single column + composite FK | `questions_accepted_answer_fk` | `SPEC US-SOLVE-001`, `SPEC section 5` |
| Unique vote per `(user, target_type, target_id)`; no dangling target | PK on generated `target_id` + real FKs | `votes_pk`, `votes_target_ck` | `SPEC section 5` |
| Resource: source + URL + check date when declared; verified needs a source of the same resource | NOT NULL/CHECK + composite FK | `resource_sources_checked_ck`, `resources_verified_ck`, `resources_verified_source_fk` | `SPEC section 5` |
| Resource countries linked, unique | composite PK + FK | `resource_countries_pk` | `SPEC section 5` |
| Project: normalized unique GitHub URL, internal owner, explicit visibility | CHECK + partial UNIQUE + NOT NULL FK + CHECK | `projects_github_url_ck`, `projects_github_url_uq`, `projects_owner_fk`, `projects_visibility_ck` | `SPEC section 5` |
| Need: existing project, bounded state, skills linked | FK + CHECK + join table | `contribution_needs_project_fk`, `contribution_needs_status_ck`, `need_skills_pk` | `SPEC section 5` |
| Application unique per user + need; no duplicate on second click | UNIQUE | `applications_user_need_uq` | `SPEC US-BUILD-001`, `SPEC section 5` |
| Content relation: allowed types, no dangling links | typed FKs + CHECKs | `content_relations_pair_ck`, `content_relations_source_ck`, `content_relations_target_ck`, `content_relations_uq` | `SPEC section 5` |
| Article/series original language recorded; level valid | NOT NULL + CHECK | `learn_articles_language_ck`, `learn_articles_level_ck` | `SPEC US-LEARN-001` |
| One progress row per user + lesson | PK | `learn_progress_pk` | PRD 5.9 (P1), ADR-0005 |
| Lesson order unique per series | deferrable UNIQUE | `training_lessons_position_uq` | ADR-0005 point 7 |
| Single-use hashed tokens | UNIQUE hash + `consumed_at` guard | `*_token_uq` | `G0 register (archived)`, `AUTHORIZATION_MATRIX` |
| Safe write retries | PK `(user_id, key)` + stored request hash | `idempotency_keys_pk` | `AUTHORIZATION_MATRIX`, `API_CONTRACT` |

Enforced by service plus transaction (not expressible as a constraint): acceptance by the question author only (`SPEC US-SOLVE-001`, `AUTHORIZATION_MATRIX`), self-vote and self-application policy, content relation targets are published/public, allowed application transitions, tag count limits, GitHub URL resolution, and moderator-only actions.

## 5. Initial migration plan

**ADR-0005 is accepted and P0 implementation is authorized (ADR-0001). Files `00001` to `00010` below exist in `db/migrations/` since B0 bootstrap (PR #24) and are immutable. `00011` and `00012` are the E2 and C14b files. The row marked PLANNED is not written yet.** Tool: goose, plain SQL, forward-only: no `-- +goose Down` section; CI rejects one. Sequential names `NNNNN_<slug>.sql` per ADR-0005 point 3. Every file begins with the header required by the migration policy (6.1): purpose, tables/indexes, lock risk, volume, transaction behaviour, compatibility window, restore plan, owner (`Y4NN777` under ADR-0002; `mibienpanjoe` for `00011` and `00012`).

| File | Contents | Lock / impact notes |
|---|---|---|
| `00001_extensions` | `pg_trgm`, `unaccent`, `f_unaccent()` wrapper (3.0) | Needs a role allowed to create the (trusted) extensions; run once. |
| `00002_reference_taxonomy` | `countries`, `tags`, `learn_topics`, `resource_categories` and their reference rows (54 African states, 9 topics, 15 categories) as deterministic `INSERT`s | Reference rows are **not** demo data; labelled in the file header. |
| `00003_identity` | `users`, `tags_created_by_fk`, `user_credentials`, `profiles`, `user_skills`, `sessions`, `email_verifications`, `password_resets` | Later changes to `users` must be expand/contract; it is the most referenced table. |
| `00004_platform_support` | `rate_limit_buckets` (UNLOGGED), `idempotency_keys`, `audit_log` | Unlogged: truncated after a crash (accepted, ADR-0003). Append-only grants on `audit_log` are applied by ops. |
| `00005_solve` | `questions`, `answers`, `questions_accepted_answer_fk`, `votes`, tag and source join tables | Generated `search_document` columns are free on empty tables; adding one later rewrites the table. |
| `00006_discover` | `providers`, `resources`, `resource_sources`, `resources_verified_source_fk`, join tables, `error_reports` | |
| `00007_build` | `projects`, `project_maintainers`, `project_tags`, `contribution_needs`, `need_skills`, `applications` | Partial unique index on `github_url` is inline. |
| `00008_learn` | `learn_articles`, `learn_article_tags`, `training_series`, `training_lessons` | |
| `00009_content_relations` | `content_relations` | Must follow 00005 to 00008 (FK targets). |
| `00010_search_indexes` | GIN/trigram indexes (3.9) | Normal `CREATE INDEX` on empty tables. **After data exists**, each new index is `CREATE INDEX CONCURRENTLY` in its own file starting with `-- +goose NO TRANSACTION`, with `lock_timeout` set. |
| a later migration (number from `make migrate-new`) | **Deferred (P1).** `learn_progress`, `resource_reviews`, `content_reports` (DDL in 3.4, 3.6, 3.7) | Listed so the ERD and ADR-0005 stay consistent. `00011` is not reserved for these tables. |
| `00011_taxonomy_ai_topics` | 5 `learn_topics`, 3 `resource_categories`, 4 relabels (3.1, 3.4); no deletions | `INSERT`/`UPDATE` on small reference tables. Not in `00002` (immutable). |
| `00012_profile_visibility` | `profiles.profile_visibility` with its named CHECK (3.2) | Constant-default `ADD COLUMN` (no rewrite, short `ACCESS EXCLUSIVE` lock); the CHECK scans `profiles`, small at P0. Separate file from E2; not in `00003`. |

Numbering: `scripts/migrate-new.sh` assigns the highest existing number plus one. `00011` and `00012` are the taxonomy and profile-visibility files. The deferred P1 tables take whatever number the script gives when they are written (`ROADMAP.md` C11). The list counts in the `00002` row stay the counts of that immutable file; current totals live in the pgtest list tests.

Cross-file rules: foreign keys to populated tables are added `NOT VALID` then `VALIDATE CONSTRAINT`; destructive changes (drop column/table, type narrowing) follow expand/contract with a backup/restore plan and a human-approved ADR (`AGENTS.md`, `AGENTS.md`); each migration is exercised on a fresh database and on upgrade-from-previous in CI; releases run a dry run and verify a backup first (`docs/SHIP.md`); an application rollback never undoes a data migration (`docs/SHIP.md`).

Migration headers document purpose, tables/indexes, lock risk, data volume, transaction behaviour, compatibility window, restore plan and owner. `sqlc` output is generated from the reviewed schema and queries, checked for drift, never hand-edited.

## 6. Seed policy

**PLANNED:** `db/seeds/` does not exist yet (`mk/50-ops.mk` has no seed implementation); the rules below are the accepted design for when it is written.

- Seeds live in `db/seeds/`, **separate from migrations**, as deterministic SQL with **fixed UUIDs** (no `now()` or random values; timestamps are literal) so two runs on an empty database produce identical rows. Reference data (countries, topics, categories) is a migration, not a seed.
- Everything seeded is **fictional and clearly marked**: every root row carries `provenance = 'demo'` (`users`, `tags`, `providers`, `resources`, `questions`, `answers`, `projects`, `learn_articles`, `training_series`); demo accounts use the reserved `@demo.226os.invalid` email domain and `demo-` usernames; URLs point only to `example.org`; the UI shows a demo badge for `provenance = 'demo'`. No real person, company, provider endorsement, contribution, price/quota/benchmark or verification is implied. A demo resource may be `verified` only against a visibly fictional `example.org` source; demo data must never be presented as real activity or reputation.
- Seeds run only in `local-dev` and `staging-test` (ADR-0005 point 8). They never run automatically, and never in `production-live`: the seed command refuses unless the environment name allows it, and a CI/release check asserts `SELECT count(*) FROM <root tables> WHERE provenance = 'demo'` is `0` against a production-live database (and that no `@demo.226os.invalid` account exists). Seeds are `provenance = 'demo'` and run only in `local-dev` and `staging-test`. A demo account in production is not part of `db/seeds/` and would need its own explicit owner decision.
- The prototype data in `prototypes/html/assets/data/*.js` (fictional people, 9 countries, categories, questions, resources, projects, the AI series) is a UX/shape reference for fixtures only; it is not imported as-is and never presented as real.
- Purging demo data is a dev-only script deleting in reverse dependency order (all content FKs are `RESTRICT`); it is not a migration.
- Seed ownership: the owner (`Y4NN777`, ADR-0002) curates and reviews demo content.

## 7. Query and index notes for P0 screens

General: no `OFFSET` on public lists; `limit` bounded (default 20, max 50, ADR-0004); order is always a stable sort key plus `id` tie-breaker; cursors are opaque and integrity-protected; every public read carries `status = 'published'` (and `visibility = 'public'` where applicable) so the partial indexes apply; `statement_timeout` is set as a backstop to Go context deadlines; evidence is `EXPLAIN (ANALYZE, BUFFERS)` per query on representative data (`SPEC section 7`, `docs/archive/spec-kit/TASKS.md`).

| Screen / query | Filters | Order and cursor key | Index |
|---|---|---|---|
| Question list (Solve) | tag (`EXISTS` on `question_tags`), country, language, unanswered (`answer_count = 0`), accepted (`accepted_answer_id IS NOT NULL`) | newest: `(published_at DESC, id DESC)`; keyset `WHERE (published_at, id) < ($1, $2)` | add `questions_published_idx ON questions (published_at DESC, id DESC) WHERE status = 'published'`; `questions_country_idx` extended to `(country_code, published_at DESC, id DESC)`; `question_tags_tag_idx (tag_id, question_id)`; unanswered partial index only if EXPLAIN justifies it |
| Question list "top" | same | `(vote_score DESC, id DESC)`; `vote_score` changes between pages, so this sort is **not** stable across pages: either accept (documented) or limit "top" to a snapshot/time window | `(vote_score DESC, id DESC) WHERE status = 'published'` only if kept |
| Question detail + answers | by id | answers `(created_at, id)` | `answers_question_idx` |
| Catalogue (Discover) | category, type, country (`resource_countries`), docs language, SDK language (`resource_sdk_languages`), provider, verification status, tag | `(name, id)`; keyset `(name, id) > ($1, $2)`; staleness derived from `max(resource_sources.checked_at)` | `resources_catalogue_idx (category_code, type, name, id)`, `resource_countries_country_idx`, `resource_sdk_languages_language_idx`, `resources_provider_idx`, `resource_tags_tag_idx` |
| Project browse (Build) | country, stage, tag/technology, open needs (`EXISTS contribution_needs ... status = 'open'`), licence | `(published_at DESC, id DESC)` | `projects_browse_idx`, `projects_country_idx`, `project_tags_tag_idx`, `contribution_needs_open_idx` |
| Need discovery by skill | `need_skills.tag_id = ANY(user skills)` | `(contribution_needs.created_at DESC, id DESC)` | `need_skills_tag_idx`, `contribution_needs_open_idx` |
| Applications (owner) | need, status | `(created_at DESC, id DESC)` | `applications_need_idx`; authorization predicate (owner or delegated maintainer of the need's project) is part of the query, never a post-filter |
| My applications | user | `(created_at DESC, id DESC)` | `applications_user_idx` |
| Learn list | kind, level, language, topic, tag | `(published_at DESC, id DESC)` | `learn_articles_browse_idx`, `learn_articles_topic_idx`, `learn_article_tags_tag_idx`; series list via `training_series_browse_idx`; lessons ordered by `(series_id, position)` via `training_lessons_position_uq` |
| Search (all) | `q` length 1-100 (`websearch_to_tsquery('simple', f_unaccent($q))`), type, country, tag, language | rank is unstable for cursors: P0 returns a bounded top-N per source (N <= 50, no deep paging), merged by `(rank DESC, type, id)`; opaque cursor only if the relevance policy fixes a quantized rank | the GIN/trigram indexes in 3.9; candidate set is bounded before `ts_rank` because ranking reads each matching tsvector |
| Explore / home feed by skills | question tags intersect the user's `user_skills` (capped at 20 skills); also open needs and Learn articles matching those tags | `(published_at DESC, id DESC)` | `user_skills` PK, `question_tags_tag_idx`, `learn_article_tags_tag_idx`; if the join is slow on measured data, denormalize `published_at` into the tag join tables or maintain a per-user feed table (not used at P0) |
| Content relations on a detail page | source or target = this entity, `status = 'active'` | `(created_at, id)` | per-FK partial indexes in 3.7, then join to published/public targets |
| Moderation queue (P0: error reports) | `error_reports.status IN ('received','triaged')` | `(created_at, id)` | `error_reports_status_idx` (`content_reports_queue_idx` arrives with `content_reports` in P1) |

Dynamic `ORDER BY` is never built from user input; sort options map to a fixed allowlist of sqlc queries. Optional filters use sqlc `sqlc.narg` with explicit `IS NULL OR` clauses; if the planner handles them poorly (generic plans), split into specialized queries.

## 8. Alternatives and tradeoffs

- **Shared polymorphic content table / per-pair relation tables:** rejected (weak FKs and authorization clarity, or 6+ tables); `content_relations` keeps real FKs (2.2).
- **Integer IDs, UUIDv4, `citext`, `ENUM` types:** rejected (enumeration, locality, extra extension, evolving statuses); see section 1.
- **Trigger-maintained search vectors and counters:** hidden mutation; generated stored vectors and explicit in-transaction counters are used.
- **External search engine:** operational and consistency surface; excluded from P0.
- **Up/down migrations:** unsafe for data-changing production migrations; restore/redeploy is the rollback.
- **Search document including tags:** a generated column cannot read another table; tags are filters and facets, not text-ranked in P0.

## 9. Search design

- Stored generated `search_document tsvector` on questions, answers (optional), resources, projects and Learn articles, with GIN indexes. Always the explicit two-argument form `to_tsvector('simple', ...)`; never `default_text_search_config`.
- Use the PostgreSQL `simple` configuration with `f_unaccent` (immutable `unaccent` wrapper) from P0 so French accents fold; a later relevance dataset may tune ranking, not reverse this.
- Add `pg_trgm` indexes only to bounded, user-facing fields where typo tolerance is required. Do not index arbitrary Markdown or unbounded blobs.
- Queries enforce maximum query length, bounded result count, explicit filters, stable ordering, and a deterministic tie-breaker; use `websearch_to_tsquery`.
- Search ranking and source freshness are separate signals; a stale source is not presented as verified merely because it ranks highly.

## 10. Decisions on the former open questions (all decided, ADR-0003/0004/0005)

| # | Topic | Decision |
|---|---|---|
| 1 | PostgreSQL version and UUID | PostgreSQL 18, `DEFAULT uuidv7()` |
| 2 | "Collaborator" | Not a role: an applicant whose application is accepted on a project. `project_maintainers` are owner-added maintainers with explicit rights (edit project, manage needs and applications) |
| 3 | 403 versus 404 | 403 if the object is visible to the caller, 404 if not |
| 4 | Migration wording | Forward-only plus restore; SPEC amended |
| 5 | Search scope | Questions, resources, projects and Learn articles; SPEC and CONTRACTS amended |
| 6 | `session_events` | Dropped: `sessions.revoked_reason` plus `audit_log` |
| 7 | Tags | Free-form, lowercase slug of at most 32 characters, max 5 per item, moderators merge/rename (P1 operation) |
| 8 | Visibility columns | Only `projects` (public/private) and profile fields carry visibility; every other entity is public iff `status = 'published'` |
| 9 | Profile defaults | Public for name, username, skills, country, bio, GitHub, `open_to_collab` and languages; email never public |
| 10 | Projects without GitHub URL and private URL uniqueness | Only `stage = 'idea'` may omit the URL; the URL uniqueness is across non-removed projects (the existence oracle for private projects is accepted at P0 and revisited in the post-P0 review) |
| 11 | Votes | Up and down; self-vote forbidden (403 `SELF_VOTE_FORBIDDEN`) |
| 12 | Counters | `answer_count` and `vote_score` are denormalized, updated in the same transaction, with a reconcile test |
| 13 | Application transitions | Service-guarded; re-apply only after the applicant's own withdrawal; none after a decline; owner or maintainer decides |
| 14 | French search | `simple` plus `f_unaccent` in P0; answers are indexed; denied attempts go to logs, not to `audit_log` |
| 15 | `rate_limit_buckets` | `UNLOGGED` accepted (ADR-0003) |
| 16 | `resource_sdks` overlap | `resources.type` = api/sdk/tool; table renamed `resource_sdk_languages` |
| 17 | Relation matrix | question-resource, question-project, resource-project, article-question/resource/project; source item's author creates, published public targets only, moderators remove |
| 18 | Platform series labelling | One original-language title per series; a bilingual series is two series rows |
| 19 | Verified source semantics | Service rule: a `claim` source cannot be the verified source; moderators and the owner mark verification |
| 20 | `questions_id_author_uq` | Dropped |
| 21 | db location | Root `db/` (DEV.md fixed) |
| 22 | Erasure | Credentials, sessions and profile personal data deleted within 30 days; authored public content kept as "membre supprimé" |
| 23 | Moderator scope | Global `moderator` role limited to hide/restore and error-report resolution; no per-domain scopes |

## 11. Approval record

ADR-0005 is ACCEPTED (2026-11-11, `Y4NN777`). The items above, the table notes in section 3 and the migration plan in section 5 are the approved design. Migrations `00001` to `00012` run on a fresh PostgreSQL 18 database in the pgtest integration suite, and the upgrade from version 10 is `TestUpgradeFromVersion10KeepsExistingRows` (see the header). The relevance dataset and EXPLAIN evidence, and the restore drill, remain G6 and G7 evidence that is NOT RUN as of 2026-11-12.
