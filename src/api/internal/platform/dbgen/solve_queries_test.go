package dbgen_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// Contract test for the Solve queries (db/queries/solve.sql): SQL sent, bind order, scan order, and
// every error path. Behaviour on real PostgreSQL is covered by pgtest/solve_integration_test.go.
func TestSolveQueries(t *testing.T) {
	ctx := t.Context()
	boom := errors.New("boom")
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	other := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	ts := pgtype.Timestamptz{Time: time.Unix(10, 0).UTC(), Valid: true}
	text := pgtype.Text{String: "go", Valid: true}

	questionRow := []any{id, "Titre", "Corps", "fr", "code", "go", "BF", int32(2), int32(3), other, ts, ts, id, "alice", "Alice", "BF"}
	listRow := append(append([]any(nil), questionRow...), float32(0.5))
	answerRow := []any{id, other, "Corps", "fr", int32(4), ts, ts, true, id, "alice", "Alice", ""}

	db := &fakeDB{many: [][]any{listRow}}
	list, err := dbgen.New(db).ListSolveQuestions(ctx, dbgen.ListSolveQuestionsParams{Tag: text, AfterCreated: ts, AfterID: id, RowLimit: 21})
	if err != nil || len(list) != 1 || list[0].AnswerCount != 2 || list[0].AuthorCountry != "BF" || list[0].Rank != 0.5 {
		t.Fatalf("ListSolveQuestions = %+v, %v", list, err)
	}
	assertCall(t, db, "ORDER BY c.rank DESC, c.created_at DESC, c.id DESC", []any{ts, float32(0), id, int32(21), pgtype.Text{}, text, pgtype.Text{}, pgtype.Text{}, pgtype.Text{}})

	if sql := db.calls[0].sql; !strings.Contains(sql, "THEN NULL ELSE 500 END") {
		t.Errorf("ListSolveQuestions does not cap the ranked candidates at 500: %s", sql)
	}

	db = &fakeDB{rowVals: questionRow}
	q, err := dbgen.New(db).GetSolveQuestion(ctx, id)
	if err != nil || q.Title != "Titre" || q.VoteScore != 3 || q.AcceptedAnswerID != other {
		t.Fatalf("GetSolveQuestion = %+v, %v", q, err)
	}
	assertCall(t, db, "q.status = 'published'", []any{id})

	db = &fakeDB{rowVals: []any{id, other, int32(1), int32(2), other, ts, ts}}
	locked, err := dbgen.New(db).LockSolveQuestion(ctx, id)
	if err != nil || locked.AnswerCount != 1 || locked.AuthorID != other {
		t.Fatalf("LockSolveQuestion = %+v, %v", locked, err)
	}
	assertCall(t, db, "FOR UPDATE", []any{id})

	db = &fakeDB{rowVals: []any{id}}
	newID, err := dbgen.New(db).InsertSolveQuestion(ctx, dbgen.InsertSolveQuestionParams{
		AuthorID: other, Title: "T", BodyMd: "B", Language: "fr", Snippet: text, SnippetLanguage: text, CountryCode: pgtype.Text{},
	})
	if err != nil || newID != id {
		t.Fatalf("InsertSolveQuestion = %v, %v", newID, err)
	}
	assertCall(t, db, "INSERT INTO questions", []any{other, "T", "B", "fr", text, text, pgtype.Text{}})

	db = &fakeDB{}
	err = dbgen.New(db).UpdateSolveQuestion(ctx, dbgen.UpdateSolveQuestionParams{ID: id, Title: text, SetSnippet: true, SetCountry: true})
	if err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "updated_at = now()", []any{text, pgtype.Text{}, true, pgtype.Text{}, pgtype.Text{}, true, pgtype.Text{}, id})

	db = &fakeDB{}
	if err := dbgen.New(db).DeleteSolveQuestionTags(ctx, id); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "DELETE FROM question_tags", []any{id})
	db = &fakeDB{rowVals: []any{int32(1)}}
	if n, err := dbgen.New(db).InsertSolveQuestionTags(ctx, dbgen.InsertSolveQuestionTagsParams{QuestionID: id, Slugs: []string{"go"}}); err != nil || n != 1 {
		t.Fatalf("hidden count %d, %v", n, err)
	}
	assertCall(t, db, "INSERT INTO question_tags", []any{[]string{"go"}, id})
	db = &fakeDB{}
	if err := dbgen.New(db).DeleteSolveQuestionLinks(ctx, dbgen.DeleteSolveQuestionLinksParams{QuestionID: id, CreatedBy: other}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "DELETE FROM content_relations", []any{id, other})

	links := map[string]func(*fakeDB) (bool, error){
		"resource": func(db *fakeDB) (bool, error) {
			return dbgen.New(db).SolveLinkResource(ctx, dbgen.SolveLinkResourceParams{TargetID: id, QuestionID: other, CreatedBy: other})
		},
		"project": func(db *fakeDB) (bool, error) {
			return dbgen.New(db).SolveLinkProject(ctx, dbgen.SolveLinkProjectParams{TargetID: id, QuestionID: other, CreatedBy: other})
		},
		"article": func(db *fakeDB) (bool, error) {
			return dbgen.New(db).SolveLinkArticle(ctx, dbgen.SolveLinkArticleParams{TargetID: id, QuestionID: other, CreatedBy: other})
		},
	}
	for kind, link := range links {
		db = &fakeDB{rowVals: []any{true}}
		if ok, err := link(db); err != nil || !ok {
			t.Fatalf("%s link = %v, %v", kind, ok, err)
		}
		assertCall(t, db, "FOR SHARE", []any{id, other, other})
		if _, err := link(&fakeDB{rowErr: boom}); !errors.Is(err, boom) {
			t.Fatalf("%s link err = %v", kind, err)
		}
	}

	db = &fakeDB{rowVals: []any{true}}
	if ok, err := dbgen.New(db).SolveQuestionExists(ctx, id); err != nil || !ok {
		t.Fatalf("SolveQuestionExists = %v, %v", ok, err)
	}
	assertCall(t, db, "SELECT EXISTS", []any{id})
	db = &fakeDB{rowVals: []any{id}}
	if got, err := dbgen.New(db).BumpSolveAnswerCount(ctx, id); err != nil || got != id {
		t.Fatalf("BumpSolveAnswerCount = %v, %v", got, err)
	}
	assertCall(t, db, "answer_count = answer_count + 1", []any{id})

	db = &fakeDB{rowVals: []any{id}}
	if got, err := dbgen.New(db).InsertSolveAnswer(ctx, dbgen.InsertSolveAnswerParams{QuestionID: id, AuthorID: other, BodyMd: "B", Language: "fr"}); err != nil || got != id {
		t.Fatalf("InsertSolveAnswer = %v, %v", got, err)
	}
	assertCall(t, db, "INSERT INTO answers", []any{id, other, "B", "fr"})
	db = &fakeDB{}
	if err := dbgen.New(db).InsertSolveAnswerSources(ctx, dbgen.InsertSolveAnswerSourcesParams{AnswerID: id, Urls: []string{"https://x"}}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "INSERT INTO answer_sources", []any{id, []string{"https://x"}})

	db = &fakeDB{rowVals: answerRow}
	ans, err := dbgen.New(db).GetSolveAnswer(ctx, id)
	if err != nil || !ans.IsAccepted || ans.VoteScore != 4 {
		t.Fatalf("GetSolveAnswer = %+v, %v", ans, err)
	}
	assertCall(t, db, "a.status = 'published' AND q.status = 'published'", []any{id})

	db = &fakeDB{many: [][]any{answerRow}}
	answers, err := dbgen.New(db).ListSolveAnswers(ctx, dbgen.ListSolveAnswersParams{QuestionID: id, AfterID: other, AfterRank: 1, AfterNegScore: -2, AfterCreated: ts, RowLimit: 21})
	if err != nil || len(answers) != 1 || answers[0].Language != "fr" {
		t.Fatalf("ListSolveAnswers = %+v, %v", answers, err)
	}
	assertCall(t, db, "a.created_at, a.id", []any{id, other, int32(1), int32(-2), ts, int32(21)})
	// A question removed after the existence check leaks no answer: the list itself filters on it.
	if sql := db.calls[0].sql; !strings.Contains(sql, "a.status = 'published' AND q.status = 'published'") {
		t.Errorf("ListSolveAnswers does not filter on the question status: %s", sql)
	}
	// An edit replaces only the author's active links; a hidden relation is a moderator decision.
	db = &fakeDB{}
	_ = dbgen.New(db).DeleteSolveQuestionLinks(ctx, dbgen.DeleteSolveQuestionLinksParams{})
	if sql := db.calls[0].sql; !strings.Contains(sql, "status = 'active'") {
		t.Errorf("DeleteSolveQuestionLinks does not keep hidden relations: %s", sql)
	}

	tags, err := dbgen.New(&fakeDB{many: [][]any{{id, "go"}}}).SolveQuestionTags(ctx, []pgtype.UUID{id})
	if err != nil || tags[0].Slug != "go" {
		t.Fatalf("SolveQuestionTags = %+v, %v", tags, err)
	}
	related, err := dbgen.New(&fakeDB{many: [][]any{{id, "resource", other, "Alpha", "fr"}}}).SolveQuestionRelated(ctx, []pgtype.UUID{id})
	if err != nil || related[0].Kind != "resource" {
		t.Fatalf("SolveQuestionRelated = %+v, %v", related, err)
	}
	sources, err := dbgen.New(&fakeDB{many: [][]any{{id, "https://x"}}}).SolveAnswerSources(ctx, []pgtype.UUID{id})
	if err != nil || sources[0].Url != "https://x" {
		t.Fatalf("SolveAnswerSources = %+v, %v", sources, err)
	}

	many := []func(*fakeDB) error{
		func(db *fakeDB) error {
			_, err := dbgen.New(db).ListSolveQuestions(ctx, dbgen.ListSolveQuestionsParams{})
			return err
		},
		func(db *fakeDB) error {
			_, err := dbgen.New(db).ListSolveAnswers(ctx, dbgen.ListSolveAnswersParams{})
			return err
		},
		func(db *fakeDB) error { _, err := dbgen.New(db).SolveQuestionTags(ctx, nil); return err },
		func(db *fakeDB) error { _, err := dbgen.New(db).SolveQuestionRelated(ctx, nil); return err },
		func(db *fakeDB) error { _, err := dbgen.New(db).SolveAnswerSources(ctx, nil); return err },
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
	one := []func(*fakeDB) error{
		func(db *fakeDB) error { _, err := dbgen.New(db).GetSolveQuestion(ctx, id); return err },
		func(db *fakeDB) error { _, err := dbgen.New(db).GetSolveAnswer(ctx, id); return err },
		func(db *fakeDB) error { _, err := dbgen.New(db).LockSolveQuestion(ctx, id); return err },
		func(db *fakeDB) error {
			_, err := dbgen.New(db).InsertSolveQuestion(ctx, dbgen.InsertSolveQuestionParams{})
			return err
		},
		func(db *fakeDB) error {
			_, err := dbgen.New(db).InsertSolveAnswer(ctx, dbgen.InsertSolveAnswerParams{})
			return err
		},
		func(db *fakeDB) error {
			_, err := dbgen.New(db).InsertSolveQuestionTags(ctx, dbgen.InsertSolveQuestionTagsParams{})
			return err
		},
		func(db *fakeDB) error { _, err := dbgen.New(db).SolveQuestionExists(ctx, id); return err },
		func(db *fakeDB) error { _, err := dbgen.New(db).BumpSolveAnswerCount(ctx, id); return err },
	}
	for _, call := range one {
		if err := call(&fakeDB{rowErr: boom}); !errors.Is(err, boom) {
			t.Fatalf("expected row error, got %v", err)
		}
	}
	exec := []func(*fakeDB) error{
		func(db *fakeDB) error {
			return dbgen.New(db).UpdateSolveQuestion(ctx, dbgen.UpdateSolveQuestionParams{})
		},
		func(db *fakeDB) error { return dbgen.New(db).DeleteSolveQuestionTags(ctx, id) },
		func(db *fakeDB) error {
			return dbgen.New(db).DeleteSolveQuestionLinks(ctx, dbgen.DeleteSolveQuestionLinksParams{})
		},
		func(db *fakeDB) error {
			return dbgen.New(db).InsertSolveAnswerSources(ctx, dbgen.InsertSolveAnswerSourcesParams{})
		},
	}
	for _, call := range exec {
		if err := call(&fakeDB{execErr: boom}); !errors.Is(err, boom) {
			t.Fatalf("expected exec error, got %v", err)
		}
	}
}
