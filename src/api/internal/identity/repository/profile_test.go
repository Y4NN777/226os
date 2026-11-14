package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
)

func profileVals(withNulls bool) []any {
	uid := pguuid()
	ts := tsVal(testTime)
	var intent, country, github, bio, bioLang pgtype.Text
	var langs, skills []string
	if !withNulls {
		intent = pgtype.Text{String: "help", Valid: true}
		country = pgtype.Text{String: "FR", Valid: true}
		github = pgtype.Text{String: "octocat", Valid: true}
		bio = pgtype.Text{String: "hi", Valid: true}
		bioLang = pgtype.Text{String: "en", Valid: true}
		langs = []string{"fr", "en"}
		skills = []string{"go", "rust"}
	}
	return []any{
		uid, "ann", "Ann", intent, country, langs, github, bio, bioLang,
		true, "public", "members", "private", "public", "members", "public", "public",
		ts, ts, skills,
	}
}

func profileUsernameVals(withNulls bool) []any {
	v := profileVals(withNulls)
	// scan order: ..., created_at, updated_at, account_status, joined_at, skills
	skills := v[len(v)-1]
	v = v[:len(v)-1]
	return append(v, "active", tsVal(testTime), skills)
}

func TestProfileByUserID(t *testing.T) {
	ctx := context.Background()

	db := &fakeDB{steps: []dbStep{{vals: profileVals(false)}}}
	p, err := New(db).ProfileByUserID(ctx, testUUID.String())
	if err != nil {
		t.Fatal(err)
	}
	if p.Username != "ann" || p.Intent == nil || *p.Intent != "help" ||
		p.CountryCode == nil || *p.CountryCode != "FR" || p.Bio == nil ||
		len(p.Languages) != 2 || len(p.Skills) != 2 || p.VisSkills != domain.VisibilityPrivate {
		t.Fatalf("mapped profile %+v", p)
	}
	if p.UserID != testUUID.String() {
		t.Fatalf("user id %q", p.UserID)
	}

	// nullables map to nil pointers
	db = &fakeDB{steps: []dbStep{{vals: profileVals(true)}}}
	p, err = New(db).ProfileByUserID(ctx, testUUID.String())
	if err != nil {
		t.Fatal(err)
	}
	if p.Intent != nil || p.CountryCode != nil || p.GitHub != nil || p.Bio != nil || p.Languages != nil {
		t.Fatalf("expected nil optionals, got %+v", p)
	}

	// bad uuid -> not found without touching the db
	db = &fakeDB{}
	if _, err := New(db).ProfileByUserID(ctx, "not-a-uuid"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad uuid: %v", err)
	}

	// no rows -> not found
	db = &fakeDB{steps: []dbStep{{}}}
	if _, err := New(db).ProfileByUserID(ctx, testUUID.String()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no rows: %v", err)
	}

	// store error -> wrapped
	db = &fakeDB{steps: []dbStep{{err: errors.New("conn reset")}}}
	if _, err := New(db).ProfileByUserID(ctx, testUUID.String()); err == nil || errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("store error: %v", err)
	}
}

func TestProfileByUsername(t *testing.T) {
	ctx := context.Background()

	db := &fakeDB{steps: []dbStep{{vals: profileUsernameVals(false)}}}
	rec, err := New(db).ProfileByUsername(ctx, "ann")
	if err != nil {
		t.Fatal(err)
	}
	if rec.AccountStatus != "active" || !rec.JoinedAt.Equal(testTime) || rec.Username != "ann" {
		t.Fatalf("record %+v", rec)
	}

	db = &fakeDB{steps: []dbStep{{vals: profileUsernameVals(true)}}}
	if _, err := New(db).ProfileByUsername(ctx, "ann"); err != nil {
		t.Fatal(err)
	}

	db = &fakeDB{steps: []dbStep{{}}}
	if _, err := New(db).ProfileByUsername(ctx, "ghost"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no rows: %v", err)
	}

	db = &fakeDB{steps: []dbStep{{err: errors.New("conn reset")}}}
	if _, err := New(db).ProfileByUsername(ctx, "ann"); err == nil || errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("store error: %v", err)
	}
}

func TestResolveSkillSlugs(t *testing.T) {
	ctx := context.Background()
	uid2 := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
	db := &fakeDB{queryRows: [][]any{
		{pguuid(), "go"}, {uid2, "rust"},
	}}
	ids, err := New(db).ResolveSkillSlugs(ctx, []string{"go", "rust"})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != testUUID.String() {
		t.Fatalf("ids %v", ids)
	}

	db = &fakeDB{queryRows: [][]any{}}
	ids, err = New(db).ResolveSkillSlugs(ctx, []string{"ghost"})
	if err != nil || len(ids) != 0 {
		t.Fatalf("empty: %v %v", ids, err)
	}

	db = &fakeDB{queryErr: errors.New("conn reset")}
	if _, err := New(db).ResolveSkillSlugs(ctx, []string{"go"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateProfile(t *testing.T) {
	ctx := context.Background()
	uid := testUUID.String()
	name := "New Name"
	bio := "hello"
	langs := []string{"fr"}
	vis := domain.VisibilityMembers
	tagIDs := []string{testUUID.String(), "01932f8e-7b1a-7c2d-9e3f-4a5b6c7d8e9f"}

	w := domain.ProfileWrite{
		DisplayName: domain.PatchField[string]{Present: true, Value: &name},
		Bio:         domain.PatchField[string]{Present: true, Value: &bio},
		Languages:   domain.PatchField[[]string]{Present: true, Value: &langs},
		Visibility:  domain.PatchField[domain.Visibility]{Present: true, Value: &vis},
		SkillTagIDs: domain.PatchField[[]string]{Present: true, Value: &tagIDs},
	}
	// skills present -> transaction: update row, delete skills, insert skills
	db := &fakeDB{steps: []dbStep{{vals: []any{pguuid()}}, {}, {}}}
	if err := New(db).UpdateProfile(ctx, uid, w); err != nil {
		t.Fatal(err)
	}

	// bad uuid -> not found
	db = &fakeDB{}
	if err := New(db).UpdateProfile(ctx, "bad", w); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad uuid: %v", err)
	}

	// no rows -> not found (skills delete/insert never run)
	db = &fakeDB{steps: []dbStep{{}}}
	if err := New(db).UpdateProfile(ctx, uid, w); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("no rows: %v", err)
	}

	// FK violation on country -> ErrCountryUnknown
	pgErr := &pgconn.PgError{Code: "23503", ConstraintName: "profiles_country_fk"}
	db = &fakeDB{steps: []dbStep{{err: pgErr}}}
	if err := New(db).UpdateProfile(ctx, uid, w); !errors.Is(err, domain.ErrCountryUnknown) {
		t.Fatalf("fk: %v", err)
	}

	// other pg error -> wrapped
	db = &fakeDB{steps: []dbStep{{err: &pgconn.PgError{Code: "23505"}}}}
	if err := New(db).UpdateProfile(ctx, uid, w); err == nil || errors.Is(err, domain.ErrCountryUnknown) {
		t.Fatalf("pg err: %v", err)
	}

	// plain error -> wrapped
	db = &fakeDB{steps: []dbStep{{err: errors.New("conn reset")}}}
	if err := New(db).UpdateProfile(ctx, uid, w); err == nil {
		t.Fatal("expected error")
	}

	// skills delete error -> wrapped, insert never runs
	db = &fakeDB{steps: []dbStep{{vals: []any{pguuid()}}, {err: errors.New("delete reset")}}}
	if err := New(db).UpdateProfile(ctx, uid, w); err == nil {
		t.Fatal("expected delete error")
	}

	// skills insert error -> wrapped
	db = &fakeDB{steps: []dbStep{{vals: []any{pguuid()}}, {}, {err: errors.New("insert reset")}}}
	if err := New(db).UpdateProfile(ctx, uid, w); err == nil {
		t.Fatal("expected insert error")
	}

	// clearing skills (nil value) deletes but skips the insert
	wClear := w
	wClear.SkillTagIDs = domain.PatchField[[]string]{Present: true}
	db = &fakeDB{steps: []dbStep{{vals: []any{pguuid()}}, {}}}
	if err := New(db).UpdateProfile(ctx, uid, wClear); err != nil {
		t.Fatal(err)
	}

	// no skills member -> single-statement path, no transaction
	wNoSkills := w
	wNoSkills.SkillTagIDs = domain.PatchField[[]string]{}
	db = &fakeDB{steps: []dbStep{{vals: []any{pguuid()}}}}
	if err := New(db).UpdateProfile(ctx, uid, wNoSkills); err != nil {
		t.Fatal(err)
	}
	db = &fakeDB{steps: []dbStep{{err: errors.New("conn reset")}}}
	if err := New(db).UpdateProfile(ctx, uid, wNoSkills); err == nil {
		t.Fatal("expected error on non-tx path")
	}

	// begin failure -> wrapped
	db = &fakeDB{beginErr: errors.New("pool exhausted")}
	if err := New(db).UpdateProfile(ctx, uid, w); err == nil {
		t.Fatal("expected begin error")
	}

	// commit failure -> wrapped
	db = &fakeDB{steps: []dbStep{{vals: []any{pguuid()}}, {}, {}}, commitErr: errors.New("commit reset")}
	if err := New(db).UpdateProfile(ctx, uid, w); err == nil {
		t.Fatal("expected commit error")
	}

	// unparseable skill id -> error before any statement runs
	wBad := w
	wBad.SkillTagIDs = domain.PatchField[[]string]{Present: true, Value: &[]string{"bad-uuid"}}
	db = &fakeDB{}
	if err := New(db).UpdateProfile(ctx, uid, wBad); err == nil {
		t.Fatal("expected bad skill id error")
	}
}

func TestPatchHelpers(t *testing.T) {
	s := "x"
	langs := []string{"en"}
	ids := []string{testUUID.String()}
	vis := domain.VisibilityPublic

	// patchText
	if f, v := patchText(domain.PatchField[string]{}); f || v.Valid {
		t.Fatal("absent patchText")
	}
	if f, v := patchText(domain.PatchField[string]{Present: true}); !f || v.Valid {
		t.Fatal("clear patchText")
	}
	if f, v := patchText(domain.PatchField[string]{Present: true, Value: &s}); !f || !v.Valid || v.String != "x" {
		t.Fatal("set patchText")
	}

	// patchString
	if f, v := patchString(domain.PatchField[string]{}); f || v != "" {
		t.Fatal("absent patchString")
	}
	if f, v := patchString(domain.PatchField[string]{Present: true}); !f || v != "" {
		t.Fatal("nil patchString")
	}
	if f, v := patchString(domain.PatchField[string]{Present: true, Value: &s}); !f || v != "x" {
		t.Fatal("set patchString")
	}

	// patchVisibility
	if f, v := patchVisibility(domain.PatchField[domain.Visibility]{}); f || v != "" {
		t.Fatal("absent patchVisibility")
	}
	if f, v := patchVisibility(domain.PatchField[domain.Visibility]{Present: true, Value: &vis}); !f || v != "public" {
		t.Fatal("set patchVisibility")
	}

	// patchStrings
	if f, v := patchStrings(domain.PatchField[[]string]{}); f || v != nil {
		t.Fatal("absent patchStrings")
	}
	if f, v := patchStrings(domain.PatchField[[]string]{Present: true}); !f || len(v) != 0 {
		t.Fatal("nil patchStrings")
	}
	if f, v := patchStrings(domain.PatchField[[]string]{Present: true, Value: &langs}); !f || len(v) != 1 {
		t.Fatal("set patchStrings")
	}

	// patchBool
	if f, v := patchBool(domain.PatchField[bool]{}); f || v {
		t.Fatal("absent patchBool")
	}
	if f, v := patchBool(domain.PatchField[bool]{Present: true}); !f || v {
		t.Fatal("nil patchBool")
	}
	b := true
	if f, v := patchBool(domain.PatchField[bool]{Present: true, Value: &b}); !f || !v {
		t.Fatal("set patchBool")
	}

	// patchUUIDs: absent -> nil, present nil -> empty non-nil, values -> parsed
	if v, err := patchUUIDs(domain.PatchField[[]string]{}); err != nil || v != nil {
		t.Fatal("absent patchUUIDs")
	}
	if v, err := patchUUIDs(domain.PatchField[[]string]{Present: true}); err != nil || v == nil || len(v) != 0 {
		t.Fatal("nil patchUUIDs")
	}
	v, err := patchUUIDs(domain.PatchField[[]string]{Present: true, Value: &ids})
	if err != nil || v == nil || len(v) != 1 || !v[0].Valid {
		t.Fatal("set patchUUIDs")
	}
	bad := []string{"not-a-uuid", testUUID.String()}
	if _, err := patchUUIDs(domain.PatchField[[]string]{Present: true, Value: &bad}); err == nil {
		t.Fatal("bad uuid must error, not skip")
	}
}

func TestTextPtr(t *testing.T) {
	if textPtr(pgtype.Text{}) != nil {
		t.Fatal("invalid text")
	}
	if p := textPtr(pgtype.Text{String: "x", Valid: true}); p == nil || *p != "x" {
		t.Fatal("valid text")
	}
}
