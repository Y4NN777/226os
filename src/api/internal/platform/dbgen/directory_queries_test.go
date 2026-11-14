package dbgen_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// Contract test for the people-directory query (db/queries/identity.sql, ListDirectoryProfiles): SQL sent,
// bind order, scan order and every error path. Visibility and filters on real PostgreSQL are covered by
// pgtest/directory_integration_test.go.
func TestListDirectoryProfilesQuery(t *testing.T) {
	ctx := t.Context()
	boom := errors.New("boom")
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	text := func(s string) pgtype.Text { return pgtype.Text{String: s, Valid: true} }

	row := []any{id, "ada", "Ada", "BF", true, []string{"fr"}, true, []string{"go"}, "ada-gh"}
	db := &fakeDB{many: [][]any{row}}
	got, err := dbgen.New(db).ListDirectoryProfiles(ctx, dbgen.ListDirectoryProfilesParams{
		Member: true, Country: text("BF"), Language: text("fr"), Skill: text("go"),
		AfterUsername: text("aa"), AfterID: id, RowLimit: 21,
	})
	if err != nil || len(got) != 1 || got[0].Username != "ada" || got[0].CountryCode != "BF" ||
		!got[0].LanguagesVisible || !got[0].SkillsVisible || got[0].GithubUsername != "ada-gh" {
		t.Fatalf("ListDirectoryProfiles = %+v, %v", got, err)
	}
	assertCall(t, db, "ORDER BY p.username, p.user_id", []any{true, text("BF"), text("fr"), text("go"), text("aa"), id, int32(21)})

	call := func(db *fakeDB) error {
		_, err := dbgen.New(db).ListDirectoryProfiles(ctx, dbgen.ListDirectoryProfilesParams{})
		return err
	}
	if err := call(&fakeDB{manyErr: boom}); err == nil {
		t.Fatal("expected query error")
	}
	if err := call(&fakeDB{scanErr: boom, many: [][]any{nil}}); err == nil {
		t.Fatal("expected scan error")
	}
	if err := call(&fakeDB{rowsErr: boom}); err == nil {
		t.Fatal("expected rows error")
	}
}
