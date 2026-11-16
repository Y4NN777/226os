-- Connect: public GitHub developers ranked by committers.top (ADR-0016), listed as external directory entries.
-- The member directory query (identity.sql ListDirectoryProfiles) is unchanged.

-- GitHubDevelopersGeneratedAt is the committers.top `generated` date of a country's stored snapshot; no row
-- means the country was never imported. The import job skips a country whose date did not change.
-- name: GitHubDevelopersGeneratedAt :one
SELECT generated_at FROM github_developers WHERE country_code = $1 LIMIT 1;

-- DeleteGitHubDevelopersByCountry clears one country inside the replace transaction (then InsertGitHubDeveloper).
-- name: DeleteGitHubDevelopersByCountry :exec
DELETE FROM github_developers WHERE country_code = $1;

-- name: InsertGitHubDeveloper :exec
INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at, imported_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- ListExternalDevelopers is the GitHub segment of GET /profiles. It drops opted-out logins and every login equal
-- (case-insensitive) to the github_username of a member who is listable for the caller by the same rule as
-- ListDirectoryProfiles (active, include_demo, profile_visibility, open_to_collab and its visibility) AND whose
-- github field is visible to the caller: a hidden github field never decides which entry is shown. The page's
-- country filter applies to the external entries only, not to the dedup. Order and keyset: (rank, country_code,
-- login), unique since (country_code, login) is the primary key; served by github_developers_rank_idx for one
-- country and github_developers_global_rank_idx otherwise.
-- name: ListExternalDevelopers :many
SELECT g.country_code, g.login, g.rank, g.name, g.avatar_url, g.contributions, g.organizations, g.generated_at
FROM github_developers g
WHERE (sqlc.narg('country')::text IS NULL OR g.country_code = sqlc.narg('country')::text)
  AND (sqlc.narg('after_rank')::integer IS NULL
       OR (g.rank, g.country_code, g.login) > (sqlc.narg('after_rank')::integer, sqlc.narg('after_country')::text, sqlc.narg('after_login')::text))
  AND NOT EXISTS (SELECT 1 FROM github_developer_optouts o WHERE o.login = lower(g.login))
  AND NOT EXISTS (
    SELECT 1 FROM profiles p JOIN users u ON u.id = p.user_id
    WHERE lower(p.github_username) = lower(g.login)
      AND u.status = 'active'
      AND (sqlc.arg('include_demo')::boolean OR u.provenance <> 'demo')
      AND (p.profile_visibility = 'public' OR (sqlc.arg('member')::boolean AND p.profile_visibility = 'members'))
      AND p.open_to_collab
      AND (p.open_to_collab_visibility = 'public' OR (sqlc.arg('member')::boolean AND p.open_to_collab_visibility = 'members'))
      AND (p.github_visibility = 'public' OR (sqlc.arg('member')::boolean AND p.github_visibility = 'members')))
ORDER BY g.rank, g.country_code, g.login
LIMIT sqlc.arg('row_limit');

-- GitHubRankForLogin is the best committers.top rank of a GitHub login (case-insensitive) for the public profile;
-- an opted-out login has none. The caller checks that the github field is visible before asking.
-- name: GitHubRankForLogin :one
SELECT g.country_code, g.rank, g.generated_at
FROM github_developers g
WHERE lower(g.login) = lower(sqlc.arg('login')::text)
  AND NOT EXISTS (SELECT 1 FROM github_developer_optouts o WHERE o.login = lower(g.login))
ORDER BY g.rank, g.country_code
LIMIT 1;
