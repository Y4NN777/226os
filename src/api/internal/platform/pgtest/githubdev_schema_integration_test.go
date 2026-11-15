//go:build integration

package pgtest

import (
	"testing"

	"github.com/jackc/pgx/v5"
)

// ADR-0016: the committers.top snapshot tables (00014) refuse anything the importer must never store.
func TestGitHubDevelopersConstraints(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	const ok = `INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at)
		VALUES ('BF', 'ada', 1, 'Ada', 'https://avatars.githubusercontent.com/u/1?v=4', 10, '{}', now())`
	if _, err := conn.Exec(ctx, ok); err != nil {
		t.Fatalf("valid row: %v", err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO github_developer_optouts (login) VALUES ('ada')`); err != nil {
		t.Fatalf("valid opt-out: %v", err)
	}
	cases := []struct{ name, stmt, code, constraint string }{
		{"unknown country", `INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at) VALUES ('ZZ', 'bob', 1, 'B', 'https://avatars.githubusercontent.com/u/2', 1, '{}', now())`, fkViol, "github_developers_country_fk"},
		{"duplicate login in a country", ok, uniqueViol, "github_developers_pkey"},
		{"rank zero", `INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at) VALUES ('BF', 'bob', 0, 'B', 'https://avatars.githubusercontent.com/u/2', 1, '{}', now())`, checkViol, "github_developers_rank_ck"},
		{"negative contributions", `INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at) VALUES ('BF', 'bob', 2, 'B', 'https://avatars.githubusercontent.com/u/2', -1, '{}', now())`, checkViol, "github_developers_contributions_ck"},
		{"bad login", `INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at) VALUES ('BF', '-bob', 2, 'B', 'https://avatars.githubusercontent.com/u/2', 1, '{}', now())`, checkViol, "github_developers_login_ck"},
		{"avatar on another host", `INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at) VALUES ('BF', 'bob', 2, 'B', 'https://evil.example/u/2', 1, '{}', now())`, checkViol, "github_developers_avatar_ck"},
		{"avatar over http", `INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at) VALUES ('BF', 'bob', 2, 'B', 'http://avatars.githubusercontent.com/u/2', 1, '{}', now())`, checkViol, "github_developers_avatar_ck"},
		{"overlong name", `INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at) VALUES ('BF', 'bob', 2, repeat('x', 201), 'https://avatars.githubusercontent.com/u/2', 1, '{}', now())`, checkViol, "github_developers_name_ck"},
		{"opt-out not lower case", `INSERT INTO github_developer_optouts (login) VALUES ('Ada')`, checkViol, "github_developer_optouts_login_ck"},
		{"duplicate opt-out", `INSERT INTO github_developer_optouts (login) VALUES ('ada')`, uniqueViol, "github_developer_optouts_pkey"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, tc.stmt)
				return err
			})
			pe := pgErr(err)
			if pe == nil || pe.Code != tc.code || pe.ConstraintName != tc.constraint {
				t.Fatalf("got %v, want %s on %s", err, tc.code, tc.constraint)
			}
		})
	}
}
