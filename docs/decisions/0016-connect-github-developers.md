# ADR-0016: the Connect directory also lists public GitHub developers ranked by committers.top, as labelled external entries

- **Status:** ACCEPTED (owner `Y4NN777`, 2026-11-15)
- **Date:** 2026-11-15
- **Approver:** `Y4NN777`, repository owner
- **Scope:** the E5 directory (`GET /profiles`, the `connect` page) and the public profile (`GET /profiles/{username}`). It does not change who is listed as a member: the C14 opt-in rule (`ListDirectoryProfiles`) is unchanged.
- **Evidence:** the committers.top data files (`_data/locations/<slug>.yml` on the `gh-pages` branch of `ashkulz/committers.top`, MIT licence), read on 2026-11-15; the local-dev seed script of PR #326 for their shape.
- **Related:** migration `00014_github_developers.sql`; `db/queries/identity.sql`; `src/api/internal/identity/committers`; `api/openapi.yaml` (`listProfiles`, `ExternalDeveloper`, `PublicProfile.ranking`); [`AUTHORIZATION_MATRIX.md`](../spec-kit/AUTHORIZATION_MATRIX.md).

## Context

In production the directory lists only members who opted in (`open_to_collab`). At launch that list is short, so the page shows little of the West African developer community it is meant to show. committers.top already ranks the most active public GitHub users per country from public data, and publishes the result under the MIT licence.

## Decision

1. **External entries, never accounts.** The directory may list the top GitHub users of a country from committers.top as **external** entries, clearly labelled "Profil GitHub public" with their rank and the source. No account, profile or user row is created for them, and they never sign in, post or get mail.
2. **Member opt-in is unchanged.** A member is listed only under the C14 rule. External entries come after the members of the page and never replace a member's own choice.
3. **Deduplication.** An external entry is dropped when a member who is listable for the caller shows the same GitHub login (case-insensitive) and that GitHub field is visible to the caller. A hidden GitHub field never decides which external entry is shown.
4. **Source and refresh.** The API process fetches `https://raw.githubusercontent.com/ashkulz/committers.top/gh-pages/_data/locations/<slug>.yml` once shortly after start and then daily. A country whose `generated` date did not change is skipped. Every field is validated, the top 500 users are kept, and a country's rows are replaced in one transaction. On any error the previous rows stay. A setting (`CONNECT_GITHUB_IMPORT`) switches the job off; it is on in `staging-test` and `production-live` and off in `local-dev` unless set.
5. **MVP countries.** West Africa first: BF, CI, SN, ML, NE, BJ, TG, GN, GH, NG. The other African countries and the other committers.top rankings come later.
6. **Opt-out.** A person who asks not to be listed is added to `github_developer_optouts` (login stored in lower case). An opted-out login is never listed and gets no rank on a profile. The directory footer says how to ask.
7. **Privacy.** Only public GitHub data is stored: login, name, avatar URL, contribution count, organisations, rank and the source's `generated` date. Company, email and location text are not stored. Avatars load from `avatars.githubusercontent.com` like members' profile photos; the database refuses any other avatar host. A member's rank appears on their profile only when their GitHub field is visible to the caller.
8. **Schema shape (deliberate).** One ranking per country (`PRIMARY KEY (country_code, login)`), no `ranking` column yet. The plan's three-ranking key is a later migration if the other rankings are added.

## Consequences

- The directory is useful on day one in production, while consent for member data stays where C14 put it.
- The page now shows people who did not sign up. The label, the source line and the opt-out keep that honest.
- The API makes one outbound HTTPS call per country per day to a fixed host, behind the same SSRF guard as the Build GitHub client.
- The directory response gains an optional `external` array; clients that read only `items` keep working.

### Accepted risks and limits (independent review, 2026-11-15)

- **`github_username` is not verified.** A member can enter the login of a ranked developer: their profile then shows that developer's rank, and the external entry disappears for callers who can list that member. Accepted for the MVP; the fix is to show the rank and deduplicate only once GitHub ownership is verified (OAuth).
- **Cursors.** The directory filter hash now includes `source`: a `/profiles` cursor issued before the deploy is refused once (`INVALID_CURSOR`), and the client restarts pagination.
- **Two countries.** A login ranked in two countries appears once per country in the unfiltered list (no duplicate row in a page walk, the keyset is unique). Accepted; a later change may keep the best rank only.
- **Opt-out.** Removal is a manual insert into `github_developer_optouts` until an admin path exists; the public request channel is the owner's to choose.
