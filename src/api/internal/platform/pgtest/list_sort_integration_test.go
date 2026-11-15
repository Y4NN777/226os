//go:build integration

package pgtest

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

// List sort (prototype data gaps, owner decision 2026-11-15): listResources, listProjects and
// listQuestions take sort; every order is walked page by page on real PostgreSQL (no duplicate, no gap,
// stable tie-breaker), and a cursor issued under one sort is 400 INVALID_CURSOR under another.

// walkIDs pages through path two items at a time as viewer and returns the ids in page order, plus
// the first next_cursor (empty when one page held everything).
func walkIDs(t *testing.T, s *s2Stack, viewer *member, path string) ([]string, string) {
	t.Helper()
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	var ids []string
	first, next := "", ""
	for page := 0; page < 100; page++ {
		p := path + sep + "limit=2"
		if next != "" {
			p += "&cursor=" + url.QueryEscape(next)
		}
		rec := s.read(t, viewer, p)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", p, rec.Code, rec.Body.String())
		}
		var got struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			NextCursor *string `json:"next_cursor"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		for _, it := range got.Items {
			ids = append(ids, it.ID)
		}
		if got.NextCursor == nil {
			return ids, first
		}
		next = *got.NextCursor
		if first == "" {
			first = next
		}
	}
	t.Fatalf("%s: more than 100 pages", path)
	return nil, ""
}

func assertOrder(t *testing.T, label string, got []string, want ...string) {
	t.Helper()
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("%s: %v, want %v", label, got, want)
	}
}

// assertCursorBoundToSort replays a cursor of sort from under every other sort of path.
func assertCursorBoundToSort(t *testing.T, s *s2Stack, viewer *member, path, from, token string, sorts ...string) {
	t.Helper()
	if token == "" {
		t.Fatalf("%s sort=%s: no cursor to replay", path, from)
	}
	for _, to := range sorts {
		if to == from {
			continue
		}
		assertProblem(t, s.read(t, viewer, path+"?sort="+to+"&cursor="+url.QueryEscape(token)), http.StatusBadRequest, "INVALID_CURSOR")
	}
}

func TestListSortProjects(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "sort-alice", true, 76)
	first := s.createProject(t, &alice, "https://github.com/alice/first", "public")
	second := s.createProject(t, &alice, "https://github.com/alice/second", "public")
	third := s.createProject(t, &alice, "https://github.com/alice/third", "public")
	// first gets two open needs (a need does not touch the project row); second is the most recently
	// updated.
	for _, title := range []string{"Backend Go", "Frontend"} {
		rec := s.write(t, &alice, http.MethodPost, "/api/v1/projects/"+first.Id.String()+"/contribution-needs",
			`{"title":"`+title+`","description":"D","lang":"fr","skills":["go"]}`, nil)
		if rec.Code != http.StatusCreated {
			t.Fatalf("need %d %s", rec.Code, rec.Body.String())
		}
	}
	if _, err := s.conn.Exec(t.Context(), `UPDATE projects SET updated_at = now() + interval '1 minute' WHERE id = $1`, second.Id.String()); err != nil {
		t.Fatal(err)
	}
	f, sc, th := first.Id.String(), second.Id.String(), third.Id.String()

	ids, _ := walkIDs(t, s, nil, "/api/v1/projects")
	assertOrder(t, "default (updated)", ids, sc, th, f)
	ids, upd := walkIDs(t, s, nil, "/api/v1/projects?sort=updated")
	assertOrder(t, "updated", ids, sc, th, f)
	ids, created := walkIDs(t, s, nil, "/api/v1/projects?sort=created")
	assertOrder(t, "created", ids, th, sc, f)
	ids, needs := walkIDs(t, s, nil, "/api/v1/projects?sort=needs")
	assertOrder(t, "needs", ids, f, sc, th)
	if p := decodeAs[apigen.ProjectPage](t, s.read(t, nil, "/api/v1/projects?sort=needs&limit=1")); p.Items[0].OpenNeedsCount != 2 {
		t.Fatalf("needs first %+v", p.Items[0])
	}
	sorts := []string{"updated", "created", "needs"}
	assertCursorBoundToSort(t, s, nil, "/api/v1/projects", "updated", upd, sorts...)
	assertCursorBoundToSort(t, s, nil, "/api/v1/projects", "created", created, sorts...)
	assertCursorBoundToSort(t, s, nil, "/api/v1/projects", "needs", needs, sorts...)
	assertProblem(t, s.read(t, nil, "/api/v1/projects?sort=name"), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
}

func TestListSortResources(t *testing.T) {
	s := newS2Stack(t.Context(), t, nil)
	// Two more never-verified entries next to the fixture, so both orders span several pages and the
	// never-verified tail has a tie broken by id.
	if _, err := s.conn.Exec(t.Context(), `INSERT INTO resources (provider_id, category_code, type, slug, name, summary,
		summary_language, doc_languages, created_by, status)
		SELECT r.provider_id, r.category_code, r.type, r.slug || '-' || v.n, v.n || ' ' || r.name, r.summary,
		       r.summary_language, r.doc_languages, r.created_by, r.status
		FROM resources r JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
		CROSS JOIN (VALUES ('aa'), ('zz')) v(n)
		WHERE r.status = 'published' AND r.verified_at IS NOT NULL
		LIMIT 2`); err != nil {
		t.Fatal(err)
	}
	names, byName := walkIDs(t, s, nil, "/api/v1/resources?sort=name")
	def, _ := walkIDs(t, s, nil, "/api/v1/resources")
	assertOrder(t, "default is name", def, names...)
	verified, byVerified := walkIDs(t, s, nil, "/api/v1/resources?sort=verified")
	if len(verified) != len(names) || len(names) < 3 {
		t.Fatalf("verified %d rows, name %d rows", len(verified), len(names))
	}
	// Each step is not more recently verified than the one before; never verified come last.
	var prev *time.Time
	seenNull := false
	for _, id := range verified {
		r := decodeAs[apigen.Resource](t, s.read(t, nil, "/api/v1/resources/"+id))
		switch {
		case r.LastVerifiedAt == nil:
			seenNull = true
		case seenNull:
			t.Fatalf("%s verified after a never-verified entry", id)
		case prev != nil && r.LastVerifiedAt.After(*prev):
			t.Fatalf("%s out of order", id)
		}
		if r.LastVerifiedAt != nil {
			at := *r.LastVerifiedAt
			prev = &at
		}
	}
	assertCursorBoundToSort(t, s, nil, "/api/v1/resources", "name", byName, "name", "verified")
	assertCursorBoundToSort(t, s, nil, "/api/v1/resources", "verified", byVerified, "name", "verified")
}

func TestListSortQuestions(t *testing.T) {
	s := newSolveStack(t)
	asker := s.signIn(t, "sort-asker", true, 77)
	helper := s.signIn(t, "sort-helper", true, 78)
	q1 := s.ask(t, &asker, "Première question").Id.String()
	q2 := s.ask(t, &asker, "Deuxième question").Id.String()
	q3 := s.ask(t, &asker, "Troisième question").Id.String()
	s.answer(t, &helper, q1, "Une réponse.")
	if rec := s.vote(t, &helper, "question", q2, 1, nil); rec.Code != http.StatusOK {
		t.Fatalf("vote %d %s", rec.Code, rec.Body.String())
	}

	ids, byNew := walkIDs(t, s, nil, "/api/v1/questions")
	assertOrder(t, "new", ids, q3, q2, q1)
	ids, byVotes := walkIDs(t, s, nil, "/api/v1/questions?sort=votes")
	assertOrder(t, "votes", ids, q2, q3, q1)
	ids, byUnanswered := walkIDs(t, s, nil, "/api/v1/questions?sort=unanswered")
	assertOrder(t, "unanswered", ids, q3, q2, q1)
	// With q, the sort still applies inside the matches.
	ids, _ = walkIDs(t, s, nil, "/api/v1/questions?q=question&sort=votes")
	assertOrder(t, "q + votes", ids, q2, q3, q1)
	sorts := []string{"new", "votes", "unanswered"}
	assertCursorBoundToSort(t, s, nil, "/api/v1/questions", "new", byNew, sorts...)
	assertCursorBoundToSort(t, s, nil, "/api/v1/questions", "votes", byVotes, sorts...)
	assertCursorBoundToSort(t, s, nil, "/api/v1/questions", "unanswered", byUnanswered, sorts...)
	assertProblem(t, s.read(t, nil, "/api/v1/questions?sort=oldest"), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
}
