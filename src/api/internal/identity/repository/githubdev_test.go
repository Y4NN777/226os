package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
)

var ghTime = time.Date(2026, 11, 14, 6, 0, 0, 0, time.UTC)

func TestGitHubGeneratedAt(t *testing.T) {
	got, ok, err := New(&fakeDB{steps: []dbStep{{vals: []any{pgtype.Timestamptz{Time: ghTime, Valid: true}}}}}).GitHubGeneratedAt(context.Background(), "BF")
	if err != nil || !ok || !got.Equal(ghTime) {
		t.Fatalf("%v %v %v", got, ok, err)
	}
	if _, ok, err := New(&fakeDB{steps: []dbStep{{}}}).GitHubGeneratedAt(context.Background(), "BF"); ok || err != nil {
		t.Fatalf("no row: %v %v", ok, err)
	}
	boom := errors.New("db down")
	if _, _, err := New(&fakeDB{steps: []dbStep{{err: boom}}}).GitHubGeneratedAt(context.Background(), "BF"); !errors.Is(err, boom) {
		t.Fatalf("error: %v", err)
	}
}

func TestReplaceGitHubDevelopers(t *testing.T) {
	devs := []domain.GitHubDeveloper{{Login: "a", Rank: 1, Name: "A", GeneratedAt: ghTime}, {Login: "b", Rank: 2, Name: "B", Organizations: []string{"o"}}}
	if err := New(&fakeDB{steps: []dbStep{{}, {}, {}}}).ReplaceGitHubDevelopers(context.Background(), "BF", devs); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("insert failed")
	for name, db := range map[string]*fakeDB{
		"delete": {steps: []dbStep{{err: boom}}},
		"insert": {steps: []dbStep{{}, {err: boom}}},
		"begin":  {beginErr: boom},
		"commit": {steps: []dbStep{{}, {}, {}}, commitErr: boom},
	} {
		if err := New(db).ReplaceGitHubDevelopers(context.Background(), "BF", devs); !errors.Is(err, boom) {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestListExternalMapsRowsAndKeyset(t *testing.T) {
	row := []any{"BF", "ada", int32(1), "Ada", "https://avatars.githubusercontent.com/u/1", int32(5), []string(nil), pgtype.Timestamptz{Time: ghTime, Valid: true}}
	db := &fakeDB{queryRows: [][]any{row}}
	got, err := New(db).ListExternal(context.Background(), domain.ExternalQuery{Member: true, Country: "BF", AfterRank: 3, AfterCountry: "BF", AfterLogin: "z", Limit: 21})
	if err != nil || len(got) != 1 {
		t.Fatalf("%+v %v", got, err)
	}
	d := got[0]
	if d.Country != "BF" || d.Login != "ada" || d.Rank != 1 || d.Contributions != 5 || d.Organizations == nil || !d.GeneratedAt.Equal(ghTime) {
		t.Fatalf("row %+v", d)
	}
	if r, ok := db.queryArgs[1].(pgtype.Int4); !ok || r.Int32 != 3 || !r.Valid {
		t.Fatalf("keyset args %v", db.queryArgs)
	}
	boom := errors.New("db down")
	if _, err := New(&fakeDB{queryErr: boom}).ListExternal(context.Background(), domain.ExternalQuery{Limit: 21}); !errors.Is(err, boom) {
		t.Fatalf("error: %v", err)
	}
}

func TestGitHubRank(t *testing.T) {
	got, err := New(&fakeDB{steps: []dbStep{{vals: []any{"SN", int32(4), pgtype.Timestamptz{Time: ghTime, Valid: true}}}}}).GitHubRank(context.Background(), "ada")
	if err != nil || got.Country != "SN" || got.Rank != 4 || !got.GeneratedAt.Equal(ghTime) {
		t.Fatalf("%+v %v", got, err)
	}
	if _, err := New(&fakeDB{steps: []dbStep{{}}}).GitHubRank(context.Background(), "ada"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no row: %v", err)
	}
	boom := errors.New("db down")
	if _, err := New(&fakeDB{steps: []dbStep{{err: boom}}}).GitHubRank(context.Background(), "ada"); !errors.Is(err, boom) {
		t.Fatalf("error: %v", err)
	}
}
