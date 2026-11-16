package dbgen_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// Contract test for the committers.top queries (db/queries/connect.sql, ADR-0016): SQL sent, bind order,
// scan order and every error path. Behaviour on real PostgreSQL is in pgtest/githubdev_integration_test.go.
func TestConnectQueries(t *testing.T) {
	ctx := t.Context()
	boom := errors.New("boom")
	ts := pgtype.Timestamptz{Time: time.Date(2026, 11, 14, 6, 0, 0, 0, time.UTC), Valid: true}

	db := &fakeDB{rowVals: []any{ts}}
	if got, err := dbgen.New(db).GitHubDevelopersGeneratedAt(ctx, "BF"); err != nil || got != ts {
		t.Fatalf("GitHubDevelopersGeneratedAt = %v, %v", got, err)
	}
	assertCall(t, db, "FROM github_developers WHERE country_code = $1 LIMIT 1", []any{"BF"})

	db = &fakeDB{}
	if err := dbgen.New(db).DeleteGitHubDevelopersByCountry(ctx, "BF"); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "DELETE FROM github_developers WHERE country_code = $1", []any{"BF"})

	db = &fakeDB{}
	in := dbgen.InsertGitHubDeveloperParams{CountryCode: "BF", Login: "ada", Rank: 1, Name: "Ada", AvatarUrl: "a",
		Contributions: 2, Organizations: []string{"o"}, GeneratedAt: ts, ImportedAt: ts}
	if err := dbgen.New(db).InsertGitHubDeveloper(ctx, in); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "INSERT INTO github_developers", []any{"BF", "ada", int32(1), "Ada", "a", int32(2), []string{"o"}, ts, ts})

	row := []any{"BF", "ada", int32(1), "Ada", "a", int32(2), []string{"o"}, ts}
	db = &fakeDB{many: [][]any{row}}
	text := func(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }
	got, err := dbgen.New(db).ListExternalDevelopers(ctx, dbgen.ListExternalDevelopersParams{
		Country: text("BF"), AfterRank: pgtype.Int4{Int32: 3, Valid: true}, AfterCountry: text("BF"), AfterLogin: text("z"),
		IncludeDemo: true, Member: true, RowLimit: 21,
	})
	if err != nil || len(got) != 1 || got[0].Login != "ada" || got[0].Rank != 1 || got[0].GeneratedAt != ts {
		t.Fatalf("ListExternalDevelopers = %+v, %v", got, err)
	}
	assertCall(t, db, "ORDER BY g.rank, g.country_code, g.login",
		[]any{text("BF"), pgtype.Int4{Int32: 3, Valid: true}, text("BF"), text("z"), true, true, int32(21)})
	list := func(db *fakeDB) error {
		_, err := dbgen.New(db).ListExternalDevelopers(ctx, dbgen.ListExternalDevelopersParams{})
		return err
	}
	if err := list(&fakeDB{manyErr: boom}); err == nil {
		t.Fatal("expected query error")
	}
	if err := list(&fakeDB{scanErr: boom, many: [][]any{nil}}); err == nil {
		t.Fatal("expected scan error")
	}
	if err := list(&fakeDB{rowsErr: boom}); err == nil {
		t.Fatal("expected rows error")
	}

	db = &fakeDB{rowVals: []any{"SN", int32(4), ts}}
	r, err := dbgen.New(db).GitHubRankForLogin(ctx, "Ada")
	if err != nil || r.CountryCode != "SN" || r.Rank != 4 || r.GeneratedAt != ts {
		t.Fatalf("GitHubRankForLogin = %+v, %v", r, err)
	}
	assertCall(t, db, "WHERE lower(g.login) = lower($1::text)", []any{"Ada"})
}
