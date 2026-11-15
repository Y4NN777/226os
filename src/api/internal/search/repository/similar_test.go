package repository

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/search/domain"
)

// Similar runs the trigram query of each group and maps its rows like Search.
func TestSimilarRunsTheTrigramQueryOfEveryGroup(t *testing.T) {
	for _, g := range domain.Groups() {
		t.Run(string(g), func(t *testing.T) {
			db := &fakeDB{rows: [][]any{{uid(hitID), "Paiement **mobile**", "*extrait*", "fr", pgtype.Timestamptz{}}}}
			got, err := New(db).Similar(t.Context(), g, domain.Query{Text: "paiemnt", Country: "BF", Lang: "fr", Limit: 6})
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 1 || got[0].Type != g || got[0].ID != hitID || got[0].Snippet != "extrait" {
				t.Fatalf("hits = %+v", got)
			}
			if !strings.Contains(db.sql, "%> i.q") || !strings.Contains(db.sql, "word_similarity") || strings.Contains(db.sql, "ts_rank") {
				t.Fatalf("%s did not run the trigram query:\n%s", g, db.sql)
			}
			if !db.deadline {
				t.Fatal("the query must run under a context deadline")
			}
		})
	}
}

func TestSimilarBindsTheLimitAndOptionalFilters(t *testing.T) {
	db := &fakeDB{}
	got, err := New(db).Similar(t.Context(), domain.Resources, domain.Query{Text: "paiemnt", Limit: 21})
	if err != nil || got == nil || len(got) != 0 {
		t.Fatalf("got %#v err %v, want an empty list", got, err)
	}
	if want := []any{int32(21), "paiemnt", pgtype.Text{}, pgtype.Text{}}; !reflect.DeepEqual(db.args, want) {
		t.Fatalf("args = %#v, want %#v", db.args, want)
	}
	if _, err := New(db).Similar(t.Context(), domain.Learn, domain.Query{Text: "paiemnt", Country: "BF", Lang: "en", Limit: 3}); err != nil {
		t.Fatal(err)
	}
	if want := []any{int32(3), "paiemnt", pgtype.Text{String: "en", Valid: true}}; !reflect.DeepEqual(db.args, want) {
		t.Fatalf("learn args = %#v, want %#v", db.args, want)
	}
}

func TestSimilarErrors(t *testing.T) {
	boom := errors.New("db down")
	for _, g := range domain.Groups() {
		_, err := New(&fakeDB{err: boom}).Similar(t.Context(), g, domain.Query{Text: "paiemnt", Limit: 6})
		if !errors.Is(err, boom) || !strings.Contains(err.Error(), "similar "+string(g)) {
			t.Fatalf("%s: err = %v", g, err)
		}
	}
	if _, err := New(&fakeDB{err: context.DeadlineExceeded}).Similar(t.Context(), domain.Questions, domain.Query{Text: "paiemnt", Limit: 6}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("timeout err = %v", err)
	}
	if _, err := New(&fakeDB{}).Similar(t.Context(), domain.Group("answers"), domain.Query{Limit: 6}); err == nil {
		t.Fatal("an unknown group must be an error")
	}
}
