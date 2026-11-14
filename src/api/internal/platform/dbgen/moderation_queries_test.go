package dbgen_test

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

func TestModerationGeneratedQueries(t *testing.T) {
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	actor := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	created := pgtype.Timestamptz{Time: time.Date(2026, 11, 14, 17, 0, 0, 0, time.UTC), Valid: true}

	for _, tc := range []struct {
		name string
		call func(*dbgen.Queries) error
	}{
		{"hide answer", func(q *dbgen.Queries) error { return q.HideModerationAnswer(t.Context(), id) }},
		{"hide article", func(q *dbgen.Queries) error { return q.HideModerationArticle(t.Context(), id) }},
		{"hide project", func(q *dbgen.Queries) error { return q.HideModerationProject(t.Context(), id) }},
		{"hide question", func(q *dbgen.Queries) error { return q.HideModerationQuestion(t.Context(), id) }},
		{"hide relation", func(q *dbgen.Queries) error { return q.HideModerationRelation(t.Context(), id) }},
		{"hide resource", func(q *dbgen.Queries) error { return q.HideModerationResource(t.Context(), id) }},
		{"restore answer", func(q *dbgen.Queries) error { return q.RestoreModerationAnswer(t.Context(), id) }},
		{"restore article", func(q *dbgen.Queries) error { return q.RestoreModerationArticle(t.Context(), id) }},
		{"restore project", func(q *dbgen.Queries) error { return q.RestoreModerationProject(t.Context(), id) }},
		{"restore question", func(q *dbgen.Queries) error { return q.RestoreModerationQuestion(t.Context(), id) }},
		{"restore relation", func(q *dbgen.Queries) error { return q.RestoreModerationRelation(t.Context(), id) }},
		{"restore resource", func(q *dbgen.Queries) error { return q.RestoreModerationResource(t.Context(), id) }},
		{"audit", func(q *dbgen.Queries) error {
			return q.InsertModerationAudit(t.Context(), dbgen.InsertModerationAuditParams{ActorID: actor, Action: "question.hide", TargetType: pgtype.Text{String: "question", Valid: true}, TargetID: id, Reason: "spam", RequestID: "req"})
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &fakeDB{}
			if err := tc.call(dbgen.New(db)); err != nil || len(db.calls) != 1 {
				t.Fatalf("err=%v calls=%d", err, len(db.calls))
			}
		})
	}

	for _, tc := range []struct {
		name string
		call func(*dbgen.Queries) (string, error)
	}{
		{"lock answer", func(q *dbgen.Queries) (string, error) { return q.LockModerationAnswer(t.Context(), id) }},
		{"lock article", func(q *dbgen.Queries) (string, error) { return q.LockModerationArticle(t.Context(), id) }},
		{"lock project", func(q *dbgen.Queries) (string, error) { return q.LockModerationProject(t.Context(), id) }},
		{"lock question", func(q *dbgen.Queries) (string, error) { return q.LockModerationQuestion(t.Context(), id) }},
		{"lock relation", func(q *dbgen.Queries) (string, error) { return q.LockModerationRelation(t.Context(), id) }},
		{"lock resource", func(q *dbgen.Queries) (string, error) { return q.LockModerationResource(t.Context(), id) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &fakeDB{rowVals: []any{"published"}}
			got, err := tc.call(dbgen.New(db))
			if err != nil || got != "published" || len(db.calls) != 1 {
				t.Fatalf("got=%q err=%v calls=%d", got, err, len(db.calls))
			}
		})
	}

	reportVals := []any{id, actor, "outdated", "received", created}
	db := &fakeDB{rowVals: reportVals}
	locked, err := dbgen.New(db).LockModerationErrorReport(t.Context(), id)
	if err != nil || locked.ID != id || locked.ResourceID != actor || locked.Kind != "outdated" {
		t.Fatalf("locked=%+v err=%v", locked, err)
	}
	db = &fakeDB{rowVals: []any{id, actor, "outdated", "resolved", created}}
	resolved, err := dbgen.New(db).ResolveModerationErrorReport(t.Context(), dbgen.ResolveModerationErrorReportParams{ID: id, ActorID: actor, Status: "resolved"})
	if err != nil || resolved.ID != id || resolved.Status != "resolved" {
		t.Fatalf("resolved=%+v err=%v", resolved, err)
	}
}
