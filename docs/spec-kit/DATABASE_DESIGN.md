# 226os P0 PostgreSQL Schema and Migration Design

**Status:** PROPOSED design; no SQL or application code is implemented

## Design principles

PostgreSQL remains the source of truth and initial search engine. Tables stay separated by domain; cross-domain links use typed relations rather than a universal content table. Every ownership and visibility decision must be enforceable in SQL or a transactionally adjacent authorization check.

## Proposed entity groups

| Group | Tables / key fields | Required invariants |
|---|---|---|
| Identity | `users`, `profiles`, `sessions`, `session_events` | normalized email unique; profile belongs to user; session hash unique; revoked/expired sessions unusable |
| Solve | `questions`, `answers`, `votes`, `tags`, `question_tags` | question/answer owner required; one accepted answer per question; one vote per subject/target; controlled publication state |
| Discover | `resources`, `resource_countries`, `resource_tags`, `resource_sources` | source URL and verification metadata; country relation unique; source/provenance explicit; no unsupported provider claim |
| Build | `projects`, `project_technologies`, `contribution_needs`, `applications` | normalized public GitHub URL unique per visibility policy; project owner required; application unique per user/need; bounded statuses |
| Relations | `content_relations` | typed source/target pairs; foreign keys or equivalent existence checks; no pending/private target exposure |
| Shared taxonomy | `countries`, `languages`, `technologies`, `tags` | normalized identifiers unique; display labels localized separately from stable codes |

## ID, time, and lifecycle defaults proposed

- UUID primary keys generated server-side; IDs are opaque but never treated as an authorization control.
- `created_at` and `updated_at` are UTC `timestamptz`; immutable event times are separate from mutable update times.
- Public content uses explicit status checks such as `draft`, `published`, `archived`, `removed`; visibility is explicit rather than inferred from status.
- Deletes are not assumed safe for content with relations. Prefer controlled archival with audit metadata; hard deletion requires a documented privacy/legal reason and relation strategy.
- Foreign keys use restrictive or explicit set-null behavior by relation type. No cascading delete is allowed by default for user-owned content.
- Email and external URL normalization rules are centralized and versioned; unique indexes apply to normalized forms.

## Search proposal

- Add a stored or generated `search_document tsvector` to questions, resources, and projects, with a GIN index.
- Use PostgreSQL `simple` configuration as the safe initial fallback while a representative French/English relevance dataset decides whether language-specific configurations improve results.
- Add `pg_trgm` indexes only to bounded, user-facing fields where typo tolerance is required. Do not index arbitrary Markdown or unbounded blobs.
- Search queries enforce maximum query length, bounded result count, explicit filters, stable ordering, and a deterministic tie-breaker.
- Search ranking and source freshness are separate signals; a stale source is not presented as verified merely because it ranks highly.

## Migration policy proposal

- Use versioned SQL migrations in `db/migrations/`; migration files are immutable after merge.
- Prefer forward-only migrations. Destructive changes require expand/contract sequencing, an explicit backup/restore plan, and a human-approved ADR.
- Each migration documents: purpose, affected tables/indexes, lock duration risk, estimated data volume, transaction behavior, compatibility window, rollback/restore plan, and owner.
- CI must exercise a fresh database and an upgrade from the previous schema. Release must perform a dry run and verify backup availability before risky migrations.
- `sqlc` queries are generated from the reviewed schema/query set; generated output is checked for drift but not hand-edited.
- Seed data is deterministic and synthetic, with provenance labels; no seed may imply a real provider endorsement, user, contribution, or verification.

## Alternatives and tradeoffs

- **Shared polymorphic content table:** less duplication but weaker foreign-key and authorization clarity; rejected as the default proposal for auditability.
- **Integer IDs:** smaller indexes but easier enumeration; UUIDs are proposed for opaque public identifiers, with authorization still mandatory.
- **Trigger-maintained search vectors:** portable across writes but hidden mutation behavior; generated/stored vectors are proposed pending PostgreSQL-version verification.
- **External search engine:** richer ranking but adds operational/data-consistency surface; excluded from P0 by the accepted architecture.
- **Up/down rollback for every migration:** attractive locally but unsafe for data-changing production migrations; restore/redeploy is proposed as the release rollback strategy.

## G0 approval blockers

Human approval is required for migration tool/version, generated versus trigger search vectors, French/English search configuration, deletion/archive policy, seed ownership, and the exact content-relation type matrix. No schema is accepted by this design note alone.
