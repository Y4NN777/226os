package dbgen_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// Contract test for the Learn queries (db/queries/learn.sql): SQL sent, bind order, scan order, and
// every error path. Behaviour on real PostgreSQL is covered by pgtest/learn_integration_test.go.
func TestLearnQueries(t *testing.T) {
	ctx := t.Context()
	boom := errors.New("boom")
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	other := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	ts := pgtype.Timestamptz{Time: time.Unix(10, 0).UTC(), Valid: true}
	text := pgtype.Text{String: "payments", Valid: true}

	listRow := []any{id, "tutorial", "Titre", "", "fr", "beginner", ts, other, "alice", "Alice", "BF"}
	db := &fakeDB{many: [][]any{listRow}}
	arts, err := dbgen.New(db).ListLearnArticles(ctx, dbgen.ListLearnArticlesParams{Topic: text, AfterPublished: ts, AfterID: id, RowLimit: 21})
	if err != nil || len(arts) != 1 || arts[0].AuthorCountry != "BF" {
		t.Fatalf("ListLearnArticles = %+v, %v", arts, err)
	}
	assertCall(t, db, "ORDER BY a.published_at DESC, a.id DESC", []any{pgtype.Text{}, text, pgtype.Text{}, pgtype.Text{}, pgtype.Text{}, ts, id, int32(21)})

	db = &fakeDB{rowVals: []any{id, "article", "Titre", "R", "# B", "fr", "beginner", "draft", pgtype.Timestamptz{}, ts, other, "alice", "Alice", ""}}
	art, err := dbgen.New(db).GetLearnArticle(ctx, dbgen.GetLearnArticleParams{ID: id, ViewerID: other})
	if err != nil || art.Status != "draft" || art.BodyMd != "# B" {
		t.Fatalf("GetLearnArticle = %+v, %v", art, err)
	}
	assertCall(t, db, "a.author_id = $2::uuid", []any{id, other})

	db = &fakeDB{rowVals: []any{id}}
	newID, err := dbgen.New(db).InsertLearnArticle(ctx, dbgen.InsertLearnArticleParams{
		AuthorID: other, Kind: "article", Title: "T", BodyMd: "B", Language: "fr", Level: "beginner", Topic: text, Status: "published",
	})
	if err != nil || newID != id {
		t.Fatalf("InsertLearnArticle = %v, %v", newID, err)
	}
	assertCall(t, db, "INSERT INTO learn_articles", []any{other, "article", "T", pgtype.Text{}, "B", "fr", "beginner", text, "published"})

	db = &fakeDB{}
	if err := dbgen.New(db).EnsureTags(ctx, dbgen.EnsureTagsParams{CreatedBy: other, Slugs: []string{"api"}}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "ON CONFLICT (slug) DO NOTHING", []any{other, []string{"api"}})
	db = &fakeDB{}
	if err := dbgen.New(db).InsertLearnArticleTags(ctx, dbgen.InsertLearnArticleTagsParams{ArticleID: id, Slugs: []string{"api"}}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "INSERT INTO learn_article_tags", []any{id, []string{"api"}})

	links := map[string]func(*fakeDB) (int64, error){
		"question": func(db *fakeDB) (int64, error) {
			return dbgen.New(db).LinkQuestionToArticle(ctx, dbgen.LinkQuestionToArticleParams{ArticleID: id, CreatedBy: other, TargetID: id})
		},
		"resource": func(db *fakeDB) (int64, error) {
			return dbgen.New(db).LinkResourceToArticle(ctx, dbgen.LinkResourceToArticleParams{ArticleID: id, CreatedBy: other, TargetID: id})
		},
		"project": func(db *fakeDB) (int64, error) {
			return dbgen.New(db).LinkProjectToArticle(ctx, dbgen.LinkProjectToArticleParams{ArticleID: id, CreatedBy: other, TargetID: id})
		},
	}
	for kind, link := range links {
		db = &fakeDB{}
		if n, err := link(db); err != nil || n != 3 {
			t.Fatalf("%s link = %d, %v", kind, n, err)
		}
		assertCall(t, db, "FOR SHARE", []any{id, other, id})
		if _, err := link(&fakeDB{execErr: boom}); !errors.Is(err, boom) {
			t.Fatalf("%s link err = %v", kind, err)
		}
	}

	tags, err := dbgen.New(&fakeDB{many: [][]any{{id, "api"}}}).LearnArticleTags(ctx, []pgtype.UUID{id})
	if err != nil || tags[0].Slug != "api" {
		t.Fatalf("LearnArticleTags = %+v, %v", tags, err)
	}
	related, err := dbgen.New(&fakeDB{many: [][]any{{"question", id, "Q", "fr"}}}).LearnArticleRelated(ctx, id)
	if err != nil || related[0].LinkType != "question" {
		t.Fatalf("LearnArticleRelated = %+v, %v", related, err)
	}
	series, err := dbgen.New(&fakeDB{many: [][]any{{id, "S", "R", "fr", "beginner", ts, other, "alice", "Alice", "", int32(4)}}}).
		ListTrainingSeries(ctx, dbgen.ListTrainingSeriesParams{RowLimit: 21})
	if err != nil || series[0].LessonCount != 4 {
		t.Fatalf("ListTrainingSeries = %+v, %v", series, err)
	}
	one, err := dbgen.New(&fakeDB{rowVals: []any{id, "S", "R", "fr", "beginner", ts, other, "alice", "Alice", ""}}).GetTrainingSeries(ctx, id)
	if err != nil || one.Title != "S" {
		t.Fatalf("GetTrainingSeries = %+v, %v", one, err)
	}
	lessons, err := dbgen.New(&fakeDB{many: [][]any{{id, int32(1), "L", other}}}).TrainingSeriesLessons(ctx, id)
	if err != nil || lessons[0].Position != 1 {
		t.Fatalf("TrainingSeriesLessons = %+v, %v", lessons, err)
	}

	many := []func(*fakeDB) error{
		func(db *fakeDB) error {
			_, err := dbgen.New(db).ListLearnArticles(ctx, dbgen.ListLearnArticlesParams{})
			return err
		},
		func(db *fakeDB) error { _, err := dbgen.New(db).LearnArticleTags(ctx, nil); return err },
		func(db *fakeDB) error { _, err := dbgen.New(db).LearnArticleRelated(ctx, id); return err },
		func(db *fakeDB) error {
			_, err := dbgen.New(db).ListTrainingSeries(ctx, dbgen.ListTrainingSeriesParams{})
			return err
		},
		func(db *fakeDB) error { _, err := dbgen.New(db).TrainingSeriesLessons(ctx, id); return err },
	}
	for _, call := range many {
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
	if _, err := dbgen.New(&fakeDB{rowErr: boom}).GetLearnArticle(ctx, dbgen.GetLearnArticleParams{}); err == nil {
		t.Fatal("expected GetLearnArticle error")
	}
	if _, err := dbgen.New(&fakeDB{rowErr: boom}).GetTrainingSeries(ctx, id); err == nil {
		t.Fatal("expected GetTrainingSeries error")
	}
	if _, err := dbgen.New(&fakeDB{rowErr: boom}).InsertLearnArticle(ctx, dbgen.InsertLearnArticleParams{}); err == nil {
		t.Fatal("expected InsertLearnArticle error")
	}
	if err := dbgen.New(&fakeDB{execErr: boom}).EnsureTags(ctx, dbgen.EnsureTagsParams{}); !errors.Is(err, boom) {
		t.Fatalf("EnsureTags err = %v", err)
	}
	if err := dbgen.New(&fakeDB{execErr: boom}).InsertLearnArticleTags(ctx, dbgen.InsertLearnArticleTagsParams{}); !errors.Is(err, boom) {
		t.Fatalf("InsertLearnArticleTags err = %v", err)
	}
}
