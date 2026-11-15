//go:build integration

package pgtest

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// List totals (prototype data gaps, owner decision 2026-11-15): total counts what the pages of the same
// filters and viewer return, whatever the cursor. Each case walks every page one item at a time and
// compares; a count that crossed a visibility boundary differently from the page would differ here.

type totalPage struct {
	Items      []json.RawMessage `json:"items"`
	NextCursor *string           `json:"next_cursor"`
	Total      *int              `json:"total"`
}

// walkTotal pages through path one item at a time as viewer (nil = anonymous) and returns how many
// items the pages held and the total every page reported (they must agree).
func walkTotal(t *testing.T, s *s2Stack, viewer *member, path string) (items, total int) {
	t.Helper()
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	next, total := "", -1
	for page := 0; page < 200; page++ {
		p := path + sep + "limit=1"
		if next != "" {
			p += "&cursor=" + url.QueryEscape(next)
		}
		rec := s.read(t, viewer, p)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", p, rec.Code, rec.Body.String())
		}
		got := decodeAs[totalPage](t, rec)
		if got.Total == nil {
			t.Fatalf("%s: no total in %s", p, rec.Body.String())
		}
		if total >= 0 && *got.Total != total {
			t.Fatalf("%s: total moved from %d to %d between pages", p, total, *got.Total)
		}
		total = *got.Total
		items += len(got.Items)
		if got.NextCursor == nil {
			return items, total
		}
		next = *got.NextCursor
	}
	t.Fatalf("%s: more than 200 pages", path)
	return 0, 0
}

func assertTotal(t *testing.T, s *s2Stack, viewer *member, path string, want int) {
	t.Helper()
	items, total := walkTotal(t, s, viewer, path)
	if items != total || (want >= 0 && total != want) {
		t.Fatalf("%s: pages %d, total %d, want %d", path, items, total, want)
	}
}

func TestListTotalsProjectsFollowVisibility(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "total-alice", true, 73)
	bob := s.signIn(t, "total-bob", true, 74)
	s.createProject(t, &alice, "https://github.com/alice/one", "public")
	s.createProject(t, &alice, "https://github.com/alice/two", "private")
	s.createProject(t, &bob, "https://github.com/bob/three", "public")
	s.createProject(t, &bob, "https://github.com/bob/four", "private")
	// Another user's private project is neither listed nor counted.
	assertTotal(t, s, nil, "/api/v1/projects", 2)
	assertTotal(t, s, &alice, "/api/v1/projects", 3)
	assertTotal(t, s, &bob, "/api/v1/projects", 3)
	assertTotal(t, s, &alice, "/api/v1/projects?technology=go", 3)
	assertTotal(t, s, nil, "/api/v1/projects?stage=active", 0)
	assertTotal(t, s, nil, "/api/v1/projects?q=three", 1)
	assertTotal(t, s, &bob, "/api/v1/projects?open_needs=true", 0)
}

func TestListTotalsCatalogueAndLearn(t *testing.T) {
	s := newS2Stack(t.Context(), t, nil)
	var published int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM resources r JOIN providers p ON p.id = r.provider_id
		WHERE r.status = 'published' AND p.status = 'published' AND p.removed_at IS NULL`).Scan(&published); err != nil {
		t.Fatal(err)
	}
	if published == 0 {
		t.Fatal("the catalogue fixture has no published resource")
	}
	assertTotal(t, s, nil, "/api/v1/resources", published)
	for _, q := range []string{"country=BF", "country=BF,CI", "type=api", "q=paiement", "doclang=fr", "provider=nope"} {
		assertTotal(t, s, nil, "/api/v1/resources?"+q, -1)
	}
	for _, q := range []string{"", "?type=tutorial", "?lang=fr", "?q=paiement"} {
		assertTotal(t, s, nil, "/api/v1/learn/articles"+q, -1)
	}
	var articles int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM learn_articles WHERE status = 'published'`).Scan(&articles); err != nil {
		t.Fatal(err)
	}
	assertTotal(t, s, nil, "/api/v1/learn/articles", articles)
	for _, q := range []string{"", "?lang=fr", "?level=beginner"} {
		assertTotal(t, s, nil, "/api/v1/learn/series"+q, -1)
	}
}

func TestListTotalsQuestions(t *testing.T) {
	s := newSolveStack(t)
	var published int
	member := s.signIn(t, "total-asker", true, 75)
	for _, title := range []string{"Paiement mobile en Go", "Webhook en retard", "Quota dépassé"} {
		s.ask(t, &member, title)
	}
	// The page joins the byline (profiles), so the fixture question of a user without a profile is out of
	// both the pages and the total.
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM questions q JOIN profiles p ON p.user_id = q.author_id
		WHERE q.status = 'published'`).Scan(&published); err != nil {
		t.Fatal(err)
	}
	if published != 3 {
		t.Fatalf("published with a byline: %d", published)
	}
	// A removed question is never counted.
	assertTotal(t, s, nil, "/api/v1/questions", published)
	for _, q := range []string{"state=open", "state=resolved", "lang=fr", "country=BF", "tag=go", "q=paiement", "q=-x"} {
		assertTotal(t, s, nil, "/api/v1/questions?"+q, -1)
	}
}
