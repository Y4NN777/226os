package dbgen_test

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// Contract test for the Build queries (db/queries/build.sql): SQL sent, bind order, scan order, and
// every error path. Behaviour on real PostgreSQL is covered by pgtest/build_integration_test.go.
func TestBuildQueries(t *testing.T) {
	ctx := t.Context()
	boom := errors.New("boom")
	id := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	other := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	ts := pgtype.Timestamptz{Time: time.Unix(10, 0).UTC(), Valid: true}
	text := pgtype.Text{String: "go", Valid: true}
	projectRow := []any{id, "Repo", "https://github.com/o/r", "D", "fr", "public", other, "alice", "Alice", []string{"go"}, int32(2), ts, ts, ts}

	db := &fakeDB{rowVals: []any{true}}
	taken, err := dbgen.New(db).ProjectURLTaken(ctx, "https://github.com/o/r")
	if err != nil || !taken {
		t.Fatalf("ProjectURLTaken = %v, %v", taken, err)
	}
	assertCall(t, db, "status <> 'removed'", []any{"https://github.com/o/r"})

	db = &fakeDB{rowVals: []any{id}}
	newID, err := dbgen.New(db).InsertProject(ctx, dbgen.InsertProjectParams{
		OwnerID: other, Name: "Repo", Summary: "S", DescriptionMd: text, Language: "fr", GithubUrl: "u", Visibility: "public",
	})
	if err != nil || newID != id {
		t.Fatalf("InsertProject = %v, %v", newID, err)
	}
	assertCall(t, db, "INSERT INTO projects", []any{other, "Repo", "S", text, "fr", "u", "public"})

	db = &fakeDB{}
	if err := dbgen.New(db).InsertProjectTags(ctx, dbgen.InsertProjectTagsParams{ProjectID: id, Slugs: []string{"go"}}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "INSERT INTO project_tags", []any{id, []string{"go"}})
	db = &fakeDB{}
	if err := dbgen.New(db).DeleteProjectTags(ctx, id); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "t.status = 'active'", []any{id})

	db = &fakeDB{rowVals: []any{other, "private"}}
	lock, err := dbgen.New(db).LockProjectAccess(ctx, id)
	if err != nil || lock.Visibility != "private" || lock.OwnerID != other {
		t.Fatalf("LockProjectAccess = %+v, %v", lock, err)
	}
	assertCall(t, db, "FOR UPDATE", []any{id})
	db = &fakeDB{rowVals: []any{other, "public"}}
	acc, err := dbgen.New(db).ProjectAccess(ctx, id)
	if err != nil || acc.Visibility != "public" {
		t.Fatalf("ProjectAccess = %+v, %v", acc, err)
	}
	assertCall(t, db, "status = 'published'", []any{id})

	db = &fakeDB{}
	if err := dbgen.New(db).UpdateProject(ctx, dbgen.UpdateProjectParams{SetDescription: true, DescriptionMd: text, ID: id}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "UPDATE projects SET", []any{true, text, pgtype.Text{}, id})

	db = &fakeDB{rowVals: projectRow}
	p, err := dbgen.New(db).GetProject(ctx, dbgen.GetProjectParams{ID: id, ViewerID: other})
	if err != nil || p.Name != "Repo" || p.OpenNeeds != 2 || p.Technologies[0] != "go" {
		t.Fatalf("GetProject = %+v, %v", p, err)
	}
	assertCall(t, db, "p.owner_id = $2::uuid", []any{id, other})

	db = &fakeDB{many: [][]any{projectRow}}
	list, err := dbgen.New(db).ListProjects(ctx, dbgen.ListProjectsParams{ViewerID: other, Technology: text, OpenNeeds: true, AfterPublished: ts, AfterID: id, RowLimit: 21})
	if err != nil || len(list) != 1 || list[0].Username != "alice" {
		t.Fatalf("ListProjects = %+v, %v", list, err)
	}
	assertCall(t, db, "ORDER BY p.published_at DESC, p.id DESC", []any{other, pgtype.Text{}, text, pgtype.Text{}, true, ts, id, int32(21)})

	db = &fakeDB{many: [][]any{{id, "question", other, "Q", "fr"}}}
	related, err := dbgen.New(db).ProjectRelated(ctx, []pgtype.UUID{id})
	if err != nil || related[0].Kind != "question" || related[0].TargetID != other {
		t.Fatalf("ProjectRelated = %+v, %v", related, err)
	}
	assertCall(t, db, "x.n <= 20", []any{[]pgtype.UUID{id}})

	links := map[string]func(*fakeDB) (bool, error){
		"question": func(db *fakeDB) (bool, error) {
			return dbgen.New(db).LinkQuestionToProject(ctx, dbgen.LinkQuestionToProjectParams{TargetID: id, ProjectID: other, CreatedBy: other})
		},
		"resource": func(db *fakeDB) (bool, error) {
			return dbgen.New(db).LinkResourceToProject(ctx, dbgen.LinkResourceToProjectParams{TargetID: id, ProjectID: other, CreatedBy: other})
		},
		"article": func(db *fakeDB) (bool, error) {
			return dbgen.New(db).LinkProjectToArticleTarget(ctx, dbgen.LinkProjectToArticleTargetParams{TargetID: id, ProjectID: other, CreatedBy: other})
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

	db = &fakeDB{rowVals: []any{id, "open", ts}}
	need, err := dbgen.New(db).InsertNeed(ctx, dbgen.InsertNeedParams{ProjectID: other, Title: "T", DescriptionMd: "D", Language: "en", CreatedBy: other})
	if err != nil || need.Status != "open" || need.ID != id {
		t.Fatalf("InsertNeed = %+v, %v", need, err)
	}
	assertCall(t, db, "INSERT INTO contribution_needs", []any{other, "T", "D", "en", other})
	db = &fakeDB{}
	if err := dbgen.New(db).InsertNeedSkills(ctx, dbgen.InsertNeedSkillsParams{NeedID: id, Slugs: []string{"go"}}); err != nil {
		t.Fatal(err)
	}
	assertCall(t, db, "INSERT INTO need_skills", []any{id, []string{"go"}})

	db = &fakeDB{many: [][]any{{id, other, "T", "D", "en", "open", ts, []string{"go"}, other, pgtype.Text{String: "pending", Valid: true}}}}
	needs, err := dbgen.New(db).ListNeeds(ctx, dbgen.ListNeedsParams{ProjectID: other, AfterCreated: ts, AfterID: id, ViewerID: id, RowLimit: 21})
	if err != nil || needs[0].Skills[0] != "go" || needs[0].ViewerApplicationID != other || needs[0].ViewerApplicationStatus.String != "pending" {
		t.Fatalf("ListNeeds = %+v, %v", needs, err)
	}
	assertCall(t, db, "ORDER BY n.created_at DESC, n.id DESC", []any{id, other, ts, id, int32(21)})

	many := []func(*fakeDB) error{
		func(db *fakeDB) error {
			_, err := dbgen.New(db).ListProjects(ctx, dbgen.ListProjectsParams{})
			return err
		},
		func(db *fakeDB) error { _, err := dbgen.New(db).ProjectRelated(ctx, nil); return err },
		func(db *fakeDB) error { _, err := dbgen.New(db).ListNeeds(ctx, dbgen.ListNeedsParams{}); return err },
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
	rowCalls := map[string]func(*fakeDB) error{
		"ProjectURLTaken": func(db *fakeDB) error { _, err := dbgen.New(db).ProjectURLTaken(ctx, ""); return err },
		"InsertProject": func(db *fakeDB) error {
			_, err := dbgen.New(db).InsertProject(ctx, dbgen.InsertProjectParams{})
			return err
		},
		"LockProjectAccess": func(db *fakeDB) error { _, err := dbgen.New(db).LockProjectAccess(ctx, id); return err },
		"ProjectAccess":     func(db *fakeDB) error { _, err := dbgen.New(db).ProjectAccess(ctx, id); return err },
		"GetProject":        func(db *fakeDB) error { _, err := dbgen.New(db).GetProject(ctx, dbgen.GetProjectParams{}); return err },
		"InsertNeed":        func(db *fakeDB) error { _, err := dbgen.New(db).InsertNeed(ctx, dbgen.InsertNeedParams{}); return err },
	}
	for name, call := range rowCalls {
		if err := call(&fakeDB{rowErr: boom}); !errors.Is(err, boom) {
			t.Fatalf("%s err = %v", name, err)
		}
	}
	execCalls := map[string]func(*fakeDB) error{
		"InsertProjectTags": func(db *fakeDB) error {
			return dbgen.New(db).InsertProjectTags(ctx, dbgen.InsertProjectTagsParams{})
		},
		"DeleteProjectTags": func(db *fakeDB) error { return dbgen.New(db).DeleteProjectTags(ctx, id) },
		"UpdateProject":     func(db *fakeDB) error { return dbgen.New(db).UpdateProject(ctx, dbgen.UpdateProjectParams{}) },
		"InsertNeedSkills": func(db *fakeDB) error {
			return dbgen.New(db).InsertNeedSkills(ctx, dbgen.InsertNeedSkillsParams{})
		},
	}
	for name, call := range execCalls {
		if err := call(&fakeDB{execErr: boom}); !errors.Is(err, boom) {
			t.Fatalf("%s err = %v", name, err)
		}
	}
}
