# ADR-0009 — Aggregated content and ingestion

- **Status:** PROPOSED (not in force until the owner sets it to ACCEPTED)
- **Date:** 2026-11-12
- **Approver:** `Y4NN77`, repository owner (pending)
- **Scope:** how real third-party content reaches every environment, including `production-live`, and how it is labelled
- **Supersedes in part:** [ADR-0005](0005-data-model-migrations-search.md) point 17 (seeds). Only the production rule for real sourced content changes. Everything else in point 17 stays.
- **Related:** [ADR-0003](0003-authentication-sessions-csrf-rate-limits.md) (moderator scope), [ADR-0005](0005-data-model-migrations-search.md), [ADR-0006](0006-toolchain-environments-quality-commands.md) (Make targets, environments), `docs/spec-kit/DATABASE_DESIGN.md` section 6 (seed policy), `docs/SPEC.md` section 7 (GitHub quotas, timeouts, anti-SSRF), `docs/SHIP.md` section 6 (CADEV audit), `THIRD_PARTY_NOTICES.md`, `ROADMAP.md` and `TODO.md` (slice S12)

## Context

The owner decided on 2026-11-12 that content comes from aggregation first. The team produces the code and the application. Content is searched, aggregated, organised and loaded for the demo, and members' own content grows beside it later.

Today, real content has no path into production:

- ADR-0005 point 17 says seeds run only in `local-dev` and `staging-test`, never in `production-live`.
- `DATABASE_DESIGN.md` section 6 adds a release check that production has zero `provenance='demo'` rows.
- P0 has no create operation for resources, providers, training series or lessons. The authorization matrix says the owner "loads catalogue content outside the API" (`docs/spec-kit/AUTHORIZATION_MATRIX.md:37`).
- Series and lessons have read operations only (`api/openapi.yaml`, `/learn/series*`).

So a public demo on production would be empty apart from what members type during the week (`ROADMAP.md`, blocker B5).

Seeded demo fixtures are fake: fake users, votes and activity. They must never be presented as real (ADR-0005 point 17). Aggregated content is different. It is real, published elsewhere, and attributable to a source. Its risks are copyright, terms of use, quality and staleness rather than deception.

## Options considered

1. **Human-authored content only.** Members and the team write everything. This is clean on rights, but it cannot fill Discover, Build and Learn before the CADEV deadline (`docs/SHIP.md` section 6), and it puts authoring load on a team that is already capacity-bound (`ROADMAP.md` K1, K23). Rejected as the only source; it remains the long-term growth path.
2. **Run seeds in production.** This contradicts ADR-0005 point 17 and the zero-`demo` release check. It would present fake users and activity as real. Rejected.
3. **Aggregated real content with explicit provenance, loaded by an explicit ingestion command (chosen).** Content is real, attributed and labelled. Fake fixtures stay out of production.

## Decision (proposed)

1. **Two different things, two rules.**
   - **Demo fixtures** (fake users, votes, answers, applications, activity) keep `provenance='demo'`. They run only in `local-dev` and `staging-test` through `make seed`, never in production. ADR-0005 point 17 is unchanged for them.
   - **Aggregated content** is real, sourced and attributed, with `provenance='aggregated'`. It may exist in every environment, including `production-live`. It is loaded only by an explicit ingestion command.
2. **Where aggregated content may go.** Only catalogue and content tables: `tags`, `providers`, `resources` (with `resource_sources`), `projects`, `learn_articles`, and `training_series` with its `training_lessons`.
   - **Solve** (questions, answers, votes), **identity** (users, profiles) and **applications** stay member-only.
   - No user, vote, answer or application is ever created by ingestion.
3. **Schema.** One new forward-only migration, numbered by `make migrate-new` and not 00011, which is reserved for P1. Its single writer is `mibienpanjoe`.
   - It widens the provenance CHECK from `('live','demo')` to `('live','demo','aggregated')` on exactly the six tables in point 2.
   - It adds the missing source metadata: source name, source URL, licence, `fetched_at` and content hash. Either per-table nullable columns or one shared table; the owner chooses.
   - It adds a CHECK that `provenance='aggregated'` implies a source URL and `fetched_at`.
   - The pgtest provenance tests change in the same PR.
4. **Ownership of aggregated rows.**
   - The tables in point 2 require a user foreign key: `created_by`, `owner_id` or `author_id`. Only `tags.created_by` is nullable.
   - Owner ruling needed (`ROADMAP.md` C15). Options:
     - (a) one clearly labelled platform account that is never shown as a person, with `training_series.origin='platform'` already available;
     - (b) nullable author or owner columns through an expand migration.
   - Either way, aggregated projects must not give anyone owner rights over needs or applications unless a real owner claims them. Claiming is P1.
5. **Ingestion tool.** A command-line tool (proposed `src/api/cmd/ingest`) that writes **through the service layer, never raw SQL**, so invariants hold. Write services missing today (providers, resources, series, lessons) are added inside their domain packages (`ROADMAP.md` C17).
   - **Input:** a committed allowlist of sources. Path and format are OPEN; the proposal is a YAML file. Each entry records the name, kind, URL, licence, terms URL, robots.txt status and rate limit.
   - **Connectors, in priority order:**
     1. the GitHub REST API for public open-source projects, with the quotas, timeouts and anti-SSRF rules of `docs/SPEC.md` section 7;
     2. RSS/Atom feeds for articles;
     3. allowlisted pages or data files for API and tool directories (providers, docs links, countries, categories).
   - **Pipeline:** normalise, dedupe by canonical URL plus content hash, classify to a topic or category by rules, write with provenance and source metadata, then publish.
     - LLM-assisted classification is a later option, behind the S9 LLM port, on public content only.
   - **Review.** Status columns have no pending-review state (`published | removed`, plus `draft` on `learn_articles` only), and no new moderation system is added. Proposal: review the allowlist and the pinned snapshot manifest in a pull request before ingestion. Moderators can hide afterwards (ADR-0003 scope) (`ROADMAP.md` C16).
   - **Pinned snapshot.** The demo uses a deterministic snapshot that records the source list and content hashes. The same snapshot is ingested on staging and then production, and its hash is recorded each time.
6. **`make ingest`.** A new Make target, added to the ADR-0006 table.
   - It fails closed. A missing allowlist, snapshot or confirmation exits non-zero.
   - Outside `local-dev` it requires explicit confirmation and backup evidence, as `migrate-up` does.
   - It is never run silently or automatically in production.
7. **Rules for sources.** These are acceptance criteria.
   - Allowlist only.
   - Respect each source's licence, robots.txt and terms. These are **verified per source, never assumed**.
   - Store facts, a short excerpt and the canonical link with attribution. Republish full text only where the licence allows it.
   - Use an identifiable user-agent, and apply a rate limit and timeouts per source.
   - Provide a takedown path: moderator hide plus removal from the allowlist. A re-run never resurrects a taken-down item.
   - Add one `THIRD_PARTY_NOTICES.md` entry per source (CADEV audit, `docs/SHIP.md` section 6).
   - Ingest no personal data.
   - Treat fetched content as untrusted: sanitise it, bound its length, and defend against prompt injection if it later feeds the assistant.
8. **Labels.** The UI shows "via <source>", with a link, for aggregated content and "by member" for member content. No fake users or fake votes appear anywhere.
9. **Taxonomy.** Rows in `learn_topics`, `resource_categories` and `tags` are data. They evolve through ingestion classification or a moderator path, never by editing old migrations. Removals still need explicit human approval.

## Consequences

- **Security.**
  - New outbound network access from wherever `make ingest` runs, and new untrusted input (HTML, feeds, JSON).
  - Mitigations: anti-SSRF, timeouts, sanitising, bounded sizes, and no execution of fetched content.
  - Running ingestion against production is a remote mutation, so it needs explicit human approval each time (`AGENTS.md`, "Human checkpoint").
- **Legal.** Copyright and terms of use are the highest risk. Licence and terms are checked per source. The default is facts, excerpt and link rather than full text. Attribution and notices are required. The originality rule of CADEV applies to the code; how the contest treats third-party content is UNVERIFIED and must be checked against the rules.
- **Data.**
  - One additive migration.
  - The zero-`demo` release check in production stays. `aggregated` rows are allowed.
  - Deduplication and content hashes keep re-runs idempotent.
  - Restore and rollback follow ADR-0005: a restore removes the ingested rows with everything else.
- **Ops.** A new Make target and a recorded snapshot. Ingestion runs on `staging-test` before `production-live`, with a backup first.
- **Cost.** GitHub API quota and network only. No new service. If LLM classification is added later, it falls under the S9 cost and quota rules.
- **Product.** Discover, Build and Learn have real content for the demo, labelled. Aggregated volume could drown member content; provenance labels now and ranking later mitigate this.

## Evidence / verification required

- The migration passes fresh and upgrade runs. Provenance CHECK tests: `aggregated` is accepted on the six tables and rejected on `users`, `questions` and `answers`.
- Connector tests run against recorded fixtures, with no network in `make quality`.
- Dedupe test: ingesting the same snapshot twice leaves row counts unchanged.
- The takedown path is tested.
- Sanitising tests cover XSS and oversized payloads.
- The `make ingest` refusal is tested: outside `local-dev` without confirmation, and without backup evidence.
- Each source's licence, robots.txt and terms are verified, with evidence attached to the allowlist PR.
- `THIRD_PARTY_NOTICES.md` has one entry per source.
- The snapshot hash is recorded for `staging-test` and `production-live`.
- Gates G1, G3, G4 and G7.
