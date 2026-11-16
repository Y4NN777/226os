package dbgen_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// Contract test for the application queries of db/queries/build.sql: SQL sent, bind order, scan
// order, and every error path. Behaviour on real PostgreSQL is in pgtest/build_applications_integration_test.go.
func TestBuildApplicationQueries(t *testing.T) {
	ctx := t.Context()
	boom := errors.New("boom")
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	other := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	ts := pgtype.Timestamptz{Time: time.Unix(10, 0).UTC(), Valid: true}
	text := pgtype.Text{String: "pending", Valid: true}
	appRow := []any{id, other, other, id, "bob", "Bob", "Je peux aider", "pending", ts, pgtype.Timestamptz{}}

	db := &fakeDB{rowVals: []any{"open", "fr", other, "public"}}
	need, err := dbgen.New(db).LockNeedForApply(ctx, id)
	if err != nil || need.Status != "open" || need.OwnerID != other || need.Visibility != "public" {
		t.Fatalf("LockNeedForApply = %+v, %v", need, err)
	}
	assertCall(t, db, "FOR SHARE OF n", []any{id})

	db = &fakeDB{rowVals: []any{id}}
	newID, err := dbgen.New(db).InsertApplication(ctx, dbgen.InsertApplicationParams{NeedID: other, UserID: id, Message: "M", Language: "fr"})
	if err != nil || newID != id {
		t.Fatalf("InsertApplication = %v, %v", newID, err)
	}
	assertCall(t, db, "ON CONFLICT (user_id, contribution_need_id) DO NOTHING", []any{other, id, "M", "fr"})

	db = &fakeDB{rowVals: []any{id, "withdrawn"}}
	own, err := dbgen.New(db).LockOwnApplication(ctx, dbgen.LockOwnApplicationParams{UserID: id, NeedID: other})
	if err != nil || own.Status != "withdrawn" {
		t.Fatalf("LockOwnApplication = %+v, %v", own, err)
	}
	assertCall(t, db, "FOR UPDATE", []any{id, other})

	db = &fakeDB{}
	n, err := dbgen.New(db).ReopenApplication(ctx, dbgen.ReopenApplicationParams{Message: "M", Language: "fr", ID: id})
	if err != nil || n != 3 {
		t.Fatalf("ReopenApplication = %d, %v", n, err)
	}
	assertCall(t, db, "status = 'withdrawn'", []any{"M", "fr", id})

	db = &fakeDB{rowVals: []any{id, "pending", other}}
	lock, err := dbgen.New(db).LockApplicationForDecision(ctx, id)
	if err != nil || lock.UserID != id || lock.OwnerID != other {
		t.Fatalf("LockApplicationForDecision = %+v, %v", lock, err)
	}
	assertCall(t, db, "FOR UPDATE OF a", []any{id})

	db = &fakeDB{}
	n, err = dbgen.New(db).DecideApplication(ctx, dbgen.DecideApplicationParams{Status: "accepted", ActorID: other, ID: id})
	if err != nil || n != 3 {
		t.Fatalf("DecideApplication = %d, %v", n, err)
	}
	assertCall(t, db, "status = 'pending'", []any{"accepted", other, id})

	db = &fakeDB{rowVals: appRow}
	app, err := dbgen.New(db).GetApplication(ctx, id)
	if err != nil || app.Username != "bob" || app.Message != "Je peux aider" {
		t.Fatalf("GetApplication = %+v, %v", app, err)
	}
	assertCall(t, db, "WHERE a.id = $1::uuid", []any{id})

	db = &fakeDB{many: [][]any{appRow}}
	list, err := dbgen.New(db).ListProjectApplications(ctx, dbgen.ListProjectApplicationsParams{
		ProjectID: other, OwnerID: id, Status: text, AfterCreated: ts, AfterID: id, RowLimit: 21,
	})
	if err != nil || len(list) != 1 || list[0].DisplayName != "Bob" {
		t.Fatalf("ListProjectApplications = %+v, %v", list, err)
	}
	assertCall(t, db, "p.owner_id = $2::uuid", []any{other, id, text, ts, id, int32(21)})

	list2 := func(db *fakeDB) error {
		_, err := dbgen.New(db).ListProjectApplications(ctx, dbgen.ListProjectApplicationsParams{})
		return err
	}
	if err := list2(&fakeDB{manyErr: boom}); err == nil {
		t.Fatal("expected query error")
	}
	if err := list2(&fakeDB{scanErr: boom, many: [][]any{nil}}); err == nil {
		t.Fatal("expected scan error")
	}
	if err := list2(&fakeDB{rowsErr: boom}); err == nil {
		t.Fatal("expected rows error")
	}
	for name, call := range map[string]func(*fakeDB) error{
		"LockNeedForApply": func(db *fakeDB) error { _, err := dbgen.New(db).LockNeedForApply(ctx, id); return err },
		"InsertApplication": func(db *fakeDB) error {
			_, err := dbgen.New(db).InsertApplication(ctx, dbgen.InsertApplicationParams{})
			return err
		},
		"LockOwnApplication": func(db *fakeDB) error {
			_, err := dbgen.New(db).LockOwnApplication(ctx, dbgen.LockOwnApplicationParams{})
			return err
		},
		"LockApplicationForDecision": func(db *fakeDB) error {
			_, err := dbgen.New(db).LockApplicationForDecision(ctx, id)
			return err
		},
		"GetApplication": func(db *fakeDB) error { _, err := dbgen.New(db).GetApplication(ctx, id); return err },
	} {
		if err := call(&fakeDB{rowErr: boom}); !errors.Is(err, boom) {
			t.Fatalf("%s err = %v", name, err)
		}
	}
	if _, err := dbgen.New(&fakeDB{execErr: boom}).ReopenApplication(ctx, dbgen.ReopenApplicationParams{}); !errors.Is(err, boom) {
		t.Fatalf("ReopenApplication err = %v", err)
	}
	if _, err := dbgen.New(&fakeDB{execErr: boom}).DecideApplication(ctx, dbgen.DecideApplicationParams{}); !errors.Is(err, boom) {
		t.Fatalf("DecideApplication err = %v", err)
	}
}
