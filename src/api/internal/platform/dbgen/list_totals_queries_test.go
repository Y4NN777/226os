package dbgen_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// Contract test for the list total queries (prototype data gaps, owner decision 2026-11-15): SQL sent,
// bind order, the single scanned column, and the error path. Behaviour on real PostgreSQL (the total
// equals what the pages return) is covered by pgtest/list_totals_integration_test.go.
func TestListTotalQueries(t *testing.T) {
	ctx := t.Context()
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	a := pgtype.Text{String: "a", Valid: true}
	b := pgtype.Text{String: "b", Valid: true}
	none := pgtype.Text{}
	cases := []struct {
		name     string
		run      func(*fakeDB) (int32, error)
		fragment string
		args     []any
	}{
		{"CountProjects", func(db *fakeDB) (int32, error) {
			return dbgen.New(db).CountProjects(ctx, dbgen.CountProjectsParams{ViewerID: id, Stage: a, Country: b, Owner: b, Q: a, OpenNeeds: true})
		}, "FROM projects p", []any{id, a, b, none, b, a, none, none, true}},
		{"CountCatalogue", func(db *fakeDB) (int32, error) {
			return dbgen.New(db).CountCatalogue(ctx, dbgen.CountCatalogueParams{Category: a, Countries: []string{"BF"}, Q: b})
		}, "FROM resources r", []any{a, none, none, []string{"BF"}, none, none, b}},
		{"CountLearnArticles", func(db *fakeDB) (int32, error) {
			return dbgen.New(db).CountLearnArticles(ctx, dbgen.CountLearnArticlesParams{Kind: a, Q: b})
		}, "FROM learn_articles a", []any{a, none, none, none, b}},
		{"CountTrainingSeries", func(db *fakeDB) (int32, error) {
			return dbgen.New(db).CountTrainingSeries(ctx, dbgen.CountTrainingSeriesParams{Topic: a, Lang: b})
		}, "FROM training_series s", []any{a, none, b}},
		{"CountSolveQuestions", func(db *fakeDB) (int32, error) {
			return dbgen.New(db).CountSolveQuestions(ctx, dbgen.CountSolveQuestionsParams{Q: a, Tag: b, Author: a, ViewerID: id, AnswersBy: b})
		}, "LIMIT (SELECT CASE WHEN tsq IS NULL THEN NULL ELSE 500 END FROM i)", []any{a, b, none, none, none, a, id, b}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := &fakeDB{rowVals: []any{int32(4)}}
			n, err := tc.run(db)
			if err != nil || n != 4 {
				t.Fatalf("%s = %d, %v", tc.name, n, err)
			}
			assertCall(t, db, tc.fragment, tc.args)
			boom := errors.New("boom")
			if _, err := tc.run(&fakeDB{rowErr: boom}); !errors.Is(err, boom) {
				t.Fatalf("%s error = %v", tc.name, err)
			}
		})
	}
}
