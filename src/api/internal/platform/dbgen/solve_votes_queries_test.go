package dbgen_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// Contract test for the Solve vote and accept queries (db/queries/solve.sql): SQL sent, bind order, scan
// order and error paths. Their behaviour on real PostgreSQL is covered by pgtest/solve_votes_integration_test.go.
func TestSolveVoteAndAcceptQueries(t *testing.T) {
	ctx := t.Context()
	boom := errors.New("boom")
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	other := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	// The target lock is the first statement of a vote and names the lock that serialises the votes of
	// one target; the author comes back for the self-vote rule.
	for name, tc := range map[string]struct {
		call func(*fakeDB) (pgtype.UUID, error)
		sql  string
	}{
		"question": {func(db *fakeDB) (pgtype.UUID, error) { return dbgen.New(db).LockSolveQuestionForVote(ctx, id) }, "FROM questions"},
		"answer":   {func(db *fakeDB) (pgtype.UUID, error) { return dbgen.New(db).LockSolveAnswerForVote(ctx, id) }, "FROM answers a"},
	} {
		db := &fakeDB{rowVals: []any{other}}
		got, err := tc.call(db)
		if err != nil || got != other {
			t.Fatalf("%s lock = %v, %v", name, got, err)
		}
		assertCall(t, db, tc.sql, []any{id})
		if sql := db.calls[0].sql; !strings.Contains(sql, "FOR NO KEY UPDATE") || !strings.Contains(sql, "'published'") {
			t.Errorf("%s lock is not a published row lock: %s", name, sql)
		}
		if _, err := tc.call(&fakeDB{rowErr: boom}); !errors.Is(err, boom) {
			t.Fatalf("%s lock err = %v", name, err)
		}
	}

	// A vote and its score delta are one statement: the upsert or delete, and the atomic UPDATE.
	db := &fakeDB{rowVals: []any{int32(7)}}
	n, err := dbgen.New(db).SetSolveQuestionVote(ctx, dbgen.SetSolveQuestionVoteParams{UserID: other, QuestionID: id, Value: -1})
	if err != nil || n != 7 {
		t.Fatalf("SetSolveQuestionVote = %v, %v", n, err)
	}
	assertCall(t, db, "vote_score = vote_score +", []any{int16(-1), id, other})
	db = &fakeDB{rowVals: []any{int32(-3)}}
	n, err = dbgen.New(db).SetSolveAnswerVote(ctx, dbgen.SetSolveAnswerVoteParams{UserID: other, AnswerID: id, Value: 0})
	if err != nil || n != -3 {
		t.Fatalf("SetSolveAnswerVote = %v, %v", n, err)
	}
	assertCall(t, db, "ON CONFLICT ON CONSTRAINT votes_pk", []any{int16(0), id, other})
	if sql := db.calls[0].sql; strings.Contains(sql, "vote_score = $") {
		t.Errorf("the score must move by a delta, not be set from a value read earlier: %s", sql)
	}
	if _, err := dbgen.New(&fakeDB{rowErr: boom}).SetSolveQuestionVote(ctx, dbgen.SetSolveQuestionVoteParams{}); !errors.Is(err, boom) {
		t.Fatalf("SetSolveQuestionVote err = %v", err)
	}
	if _, err := dbgen.New(&fakeDB{rowErr: boom}).SetSolveAnswerVote(ctx, dbgen.SetSolveAnswerVoteParams{}); !errors.Is(err, boom) {
		t.Fatalf("SetSolveAnswerVote err = %v", err)
	}

	db = &fakeDB{rowVals: []any{id}}
	got, err := dbgen.New(db).LockSolveAnswerForAccept(ctx, dbgen.LockSolveAnswerForAcceptParams{ID: id, QuestionID: other})
	if err != nil || got != id {
		t.Fatalf("LockSolveAnswerForAccept = %v, %v", got, err)
	}
	assertCall(t, db, "FOR SHARE", []any{id, other})
	if _, err := dbgen.New(&fakeDB{rowErr: boom}).LockSolveAnswerForAccept(ctx, dbgen.LockSolveAnswerForAcceptParams{}); !errors.Is(err, boom) {
		t.Fatalf("LockSolveAnswerForAccept err = %v", err)
	}

	db = &fakeDB{}
	if err := dbgen.New(db).SetSolveAcceptedAnswer(ctx, dbgen.SetSolveAcceptedAnswerParams{ID: id, AnswerID: other}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "accepted_answer_id = ", []any{other, id})
	if sql := db.calls[0].sql; strings.Contains(sql, "updated_at") {
		t.Errorf("accepting must not change updated_at (the ETag): %s", sql)
	}
	if err := dbgen.New(&fakeDB{execErr: boom}).SetSolveAcceptedAnswer(ctx, dbgen.SetSolveAcceptedAnswerParams{}); !errors.Is(err, boom) {
		t.Fatalf("SetSolveAcceptedAnswer err = %v", err)
	}
}
