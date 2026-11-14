package repository

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
)

const (
	appID = "00000000-0000-7000-8000-0000000000f1"
	bobID = "00000000-0000-7000-8000-0000000000b2"
)

var (
	openNeed = rows([]any{"open", "fr", uid(userID), "public"})
	newApp   = domain.NewApplication{NeedID: needID, UserID: bobID, Message: "Je peux aider"}
	appLock  = rows([]any{uid(bobID), "pending", uid(userID)})
)

func allowNeed(domain.NeedAccess) error { return nil }

func allowApp(domain.ApplicationAccess) error { return nil }

func appRow(decided pgtype.Timestamptz) []any {
	return []any{uid(appID), uid(needID), uid(projID), uid(bobID), "bob", "Bob", "Je peux aider", "pending", ts(), decided}
}

func TestCreateApplication(t *testing.T) {
	var seen domain.NeedAccess
	capture := func(a domain.NeedAccess) error { seen = a; return nil }
	f := db(t, openNeed, rows([]any{uid(appID)}))
	id, err := New(f).CreateApplication(t.Context(), newApp, capture)
	if err != nil || id != appID || !f.committed || seen != (domain.NeedAccess{NeedStatus: "open", OwnerID: userID, Visibility: "public"}) {
		t.Fatalf("create %q %v %+v", id, err, seen)
	}
	// Already withdrawn: the same row is reopened.
	f = db(t, openNeed, none, rows([]any{uid(appID), "withdrawn"}), ok)
	if id, err := New(f).CreateApplication(t.Context(), newApp, allowNeed); err != nil || id != appID || !f.committed {
		t.Fatalf("reopen %q %v", id, err)
	}
	badNeed, badUser := newApp, newApp
	badNeed.NeedID, badUser.UserID = "x", "x"
	tests := []struct {
		name          string
		f             *fakeDB
		in            domain.NewApplication
		auth          func(domain.NeedAccess) error
		want          error
		begin, commit error
	}{
		{name: "bad need", f: db(t), in: badNeed, auth: allowNeed},
		{name: "bad user", f: db(t), in: badUser, auth: allowNeed},
		{name: "begin", f: db(t), in: newApp, auth: allowNeed, begin: errDB, want: errDB},
		{name: "missing need", f: db(t, none), in: newApp, auth: allowNeed, want: domain.ErrNotFound},
		{name: "lock need", f: db(t, fail(errDB)), in: newApp, auth: allowNeed, want: errDB},
		{name: "denied", f: db(t, openNeed), in: newApp, auth: func(domain.NeedAccess) error { return errDeny }, want: errDeny},
		{name: "insert", f: db(t, openNeed, fail(errDB)), in: newApp, auth: allowNeed, want: errDB},
		{name: "duplicate pending", f: db(t, openNeed, none, rows([]any{uid(appID), "pending"})), in: newApp, auth: allowNeed, want: domain.ErrDuplicateApplication},
		{name: "after a decline", f: db(t, openNeed, none, rows([]any{uid(appID), "declined"})), in: newApp, auth: allowNeed, want: domain.ErrDuplicateApplication},
		{name: "lock own", f: db(t, openNeed, none, fail(errDB)), in: newApp, auth: allowNeed, want: errDB},
		{name: "reopen", f: db(t, openNeed, none, rows([]any{uid(appID), "withdrawn"}), fail(errDB)), in: newApp, auth: allowNeed, want: errDB},
		{name: "commit", f: db(t, openNeed, rows([]any{uid(appID)})), in: newApp, auth: allowNeed, commit: errDB, want: errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.f.t, tc.f.beginErr, tc.f.commitErr = t, tc.begin, tc.commit
			_, err := New(tc.f).CreateApplication(t.Context(), tc.in, tc.auth)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || tc.f.committed {
				t.Fatalf("err %v, want %v", err, tc.want)
			}
		})
	}
}

func TestDecideApplication(t *testing.T) {
	var seen domain.ApplicationAccess
	capture := func(a domain.ApplicationAccess) error { seen = a; return nil }
	f := db(t, appLock, ok)
	if err := New(f).DecideApplication(t.Context(), appID, userID, "accepted", capture); err != nil || !f.committed ||
		seen != (domain.ApplicationAccess{ApplicantID: bobID, OwnerID: userID, Status: "pending"}) {
		t.Fatalf("decide %v %+v", err, seen)
	}
	tests := []struct {
		name          string
		f             *fakeDB
		id, actor     string
		auth          func(domain.ApplicationAccess) error
		want          error
		begin, commit error
	}{
		{name: "bad id", f: db(t), id: "x", actor: userID, auth: allowApp},
		{name: "bad actor", f: db(t), id: appID, actor: "x", auth: allowApp},
		{name: "begin", f: db(t), id: appID, actor: userID, auth: allowApp, begin: errDB, want: errDB},
		{name: "missing", f: db(t, none), id: appID, actor: userID, auth: allowApp, want: domain.ErrNotFound},
		{name: "lock", f: db(t, fail(errDB)), id: appID, actor: userID, auth: allowApp, want: errDB},
		{name: "denied", f: db(t, appLock), id: appID, actor: userID, auth: func(domain.ApplicationAccess) error { return errDeny }, want: errDeny},
		{name: "no longer pending", f: db(t, appLock, step{zero: true}), id: appID, actor: userID, auth: allowApp, want: domain.ErrStateConflict},
		{name: "update", f: db(t, appLock, fail(errDB)), id: appID, actor: userID, auth: allowApp, want: errDB},
		{name: "commit", f: db(t, appLock, ok), id: appID, actor: userID, auth: allowApp, commit: errDB, want: errDB},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.f.t, tc.f.beginErr, tc.f.commitErr = t, tc.begin, tc.commit
			err := New(tc.f).DecideApplication(t.Context(), tc.id, tc.actor, "accepted", tc.auth)
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || tc.f.committed {
				t.Fatalf("err %v, want %v", err, tc.want)
			}
		})
	}
}

func TestGetAndListApplications(t *testing.T) {
	decided := ts()
	a, err := New(db(t, rows(appRow(decided)))).GetApplication(t.Context(), appID)
	at := decided.Time
	want := domain.Application{ID: appID, NeedID: needID, ProjectID: projID, Message: "Je peux aider", Status: "pending",
		Applicant: domain.Owner{ID: bobID, Username: "bob", DisplayName: "Bob"}, CreatedAt: ts().Time, DecidedAt: &at}
	if err != nil || !reflect.DeepEqual(a, want) {
		t.Fatalf("get %+v %v", a, err)
	}
	if _, err := New(db(t)).GetApplication(t.Context(), "x"); err == nil {
		t.Fatal("bad id accepted")
	}
	if _, err := New(db(t, none)).GetApplication(t.Context(), appID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing %v", err)
	}
	if _, err := New(db(t, fail(errDB))).GetApplication(t.Context(), appID); !errors.Is(err, errDB) {
		t.Fatalf("err %v", err)
	}

	after := time.Now()
	list, err := New(db(t, rows(appRow(pgtype.Timestamptz{})))).ListApplications(t.Context(), domain.ApplicationQuery{
		ProjectID: projID, OwnerID: userID, Status: "pending", AfterCreated: &after, AfterID: appID, Limit: 21,
	})
	if err != nil || len(list) != 1 || list[0].DecidedAt != nil {
		t.Fatalf("list %+v %v", list, err)
	}
	for name, q := range map[string]domain.ApplicationQuery{
		"bad project": {ProjectID: "x", OwnerID: userID},
		"bad owner":   {ProjectID: projID, OwnerID: "x"},
		"bad cursor":  {ProjectID: projID, OwnerID: userID, AfterID: "x"},
	} {
		if _, err := New(db(t)).ListApplications(t.Context(), q); err == nil {
			t.Fatalf("%s accepted", name)
		}
	}
	if _, err := New(db(t, fail(errDB))).ListApplications(t.Context(), domain.ApplicationQuery{ProjectID: projID, OwnerID: userID}); !errors.Is(err, errDB) {
		t.Fatalf("err %v", err)
	}
}
