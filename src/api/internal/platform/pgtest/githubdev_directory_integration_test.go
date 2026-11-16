//go:build integration

package pgtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/repository"
)

// ADR-0016 on the production router over a migrated database: external committers.top entries in GET
// /profiles and the rank on GET /profiles/{username}, with the member rules of C14 unchanged.

type mixedPage struct {
	Items      []map[string]any `json:"items"`
	External   []map[string]any `json:"external"`
	NextCursor *string          `json:"next_cursor"`
}

func (s *s2Stack) mixed(t *testing.T, m *member, query string) mixedPage {
	t.Helper()
	rec := s.read(t, m, "/api/v1/profiles?"+query)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s: %d %s", query, rec.Code, rec.Body.String())
	}
	var p mixedPage
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func (s *s2Stack) walkMixed(t *testing.T, m *member, query string) []string {
	t.Helper()
	var seen []string
	cursor := ""
	for range 20 {
		q := query
		if cursor != "" {
			q += "&cursor=" + cursor
		}
		p := s.mixed(t, m, q)
		for _, it := range p.Items {
			seen = append(seen, fmt.Sprint("m:", it["username"]))
		}
		for _, e := range p.External {
			seen = append(seen, fmt.Sprint("g:", e["login"]))
		}
		if p.NextCursor == nil {
			return seen
		}
		cursor = *p.NextCursor
	}
	t.Fatal("too many pages")
	return nil
}

func seedRanked(t *testing.T, s *s2Stack, country string, logins ...string) {
	t.Helper()
	devs := make([]domain.GitHubDeveloper, len(logins))
	for i, l := range logins {
		devs[i] = ghDev(country, l, i+1)
	}
	if err := repository.New(s.pool).ReplaceGitHubDevelopers(t.Context(), country, devs); err != nil {
		t.Fatal(err)
	}
}

func TestDirectoryListsMembersThenExternalAcrossPages(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedTags(ctx, t, s.conn)
	newPerson(1, "anna").insert(ctx, t, s.conn)
	newPerson(2, "binta").insert(ctx, t, s.conn)
	seedRanked(t, s, "BF", "g1", "g2", "g3")
	want := []string{"m:anna", "m:binta", "g:g1", "g:g2", "g:g3"}
	for _, limit := range []int{1, 2, 3, 10} {
		if got := s.walkMixed(t, nil, fmt.Sprintf("limit=%d", limit)); !slices.Equal(got, want) {
			t.Fatalf("limit %d: %v, want %v", limit, got, want)
		}
	}
	if got := s.walkMixed(t, nil, "source=github&limit=2"); !slices.Equal(got, want[2:]) {
		t.Fatalf("github: %v", got)
	}
	if got := s.walkMixed(t, nil, "source=members&limit=1"); !slices.Equal(got, want[:2]) {
		t.Fatalf("members: %v", got)
	}
	if got := s.walkMixed(t, nil, "language=fr"); !slices.Equal(got, want[:2]) {
		t.Fatalf("a language filter omits external entries: %v", got)
	}
	if got := s.walkMixed(t, nil, "country=SN"); len(got) != 0 {
		t.Fatalf("country filter on both segments: %v", got)
	}
	// A cursor of one source is refused on another.
	first := s.mixed(t, nil, "limit=1")
	if rec := s.read(t, nil, "/api/v1/profiles?source=members&limit=1&cursor="+*first.NextCursor); rec.Code != http.StatusBadRequest {
		t.Fatalf("cross-source cursor: %d", rec.Code)
	}
	if rec := s.read(t, nil, "/api/v1/profiles?source=everyone"); rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("bad source: %d", rec.Code)
	}
	// The external entry carries public data and its source, never an account field.
	e := s.mixed(t, nil, "source=github&limit=1").External[0]
	if e["source"] != "committers.top" || e["country"] != "BF" || e["rank"] != float64(1) ||
		!strings.HasPrefix(fmt.Sprint(e["avatar_url"]), "https://avatars.githubusercontent.com/") || len(e) != 9 {
		t.Fatalf("external entry %v", e)
	}
}

func TestDirectoryMembersOutputIsUnchanged(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedMatrix(ctx, t, s.conn)
	viewer := s.signIn(t, "unchanged-viewer", true, 70)
	before := map[string]string{}
	for _, who := range []*member{nil, &viewer} {
		rec := s.read(t, who, "/api/v1/profiles?limit=50")
		before[fmt.Sprint(who != nil)] = rec.Body.String()
	}
	seedRanked(t, s, "BF", "g1", "g2")
	for _, who := range []*member{nil, &viewer} {
		got := s.read(t, who, "/api/v1/profiles?limit=50&source=members").Body.String()
		if got != before[fmt.Sprint(who != nil)] {
			t.Fatalf("source=members changed the member page:\nbefore %s\nafter  %s", before[fmt.Sprint(who != nil)], got)
		}
		all := s.mixed(t, who, "limit=50")
		raw, _ := json.Marshal(all.Items)
		var old mixedPage
		_ = json.Unmarshal([]byte(before[fmt.Sprint(who != nil)]), &old)
		rawOld, _ := json.Marshal(old.Items)
		if string(raw) != string(rawOld) || len(all.External) != 2 {
			t.Fatalf("source=all items differ or external missing: %s vs %s, %d external", raw, rawOld, len(all.External))
		}
	}
}

func TestDirectoryDedupFollowsTheCallersView(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedTags(ctx, t, s.conn)
	pub := newPerson(1, "pubgh")
	pub.github = "PubGH" // case differs from the ranked login
	mem := newPerson(2, "memgh")
	mem.github = "memgh"
	mem.githubVis = "members"
	priv := newPerson(3, "privgh")
	priv.github = "privgh"
	priv.githubVis = "private"
	for _, p := range []person{pub, mem, priv} {
		p.insert(ctx, t, s.conn)
	}
	seedRanked(t, s, "BF", "pubgh", "memgh", "privgh", "optedout")
	if _, err := s.conn.Exec(ctx, `INSERT INTO github_developer_optouts (login) VALUES ('optedout')`); err != nil {
		t.Fatal(err)
	}
	viewer := s.signIn(t, "dedup-viewer", true, 71)
	if got := s.walkMixed(t, nil, "source=github"); !slices.Equal(got, []string{"g:memgh", "g:privgh"}) {
		t.Fatalf("anonymous: %v", got)
	}
	if got := s.walkMixed(t, &viewer, "source=github"); !slices.Equal(got, []string{"g:privgh"}) {
		t.Fatalf("member: %v", got)
	}
}

func TestProfileRankingIsShownOnlyWithAVisibleGitHub(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedTags(ctx, t, s.conn)
	p := newPerson(1, "ranked")
	p.github = "Ranked-GH"
	p.githubVis = "members"
	p.insert(ctx, t, s.conn)
	seedRanked(t, s, "SN", "other", "ranked-gh")
	viewer := s.signIn(t, "rank-viewer", true, 72)
	ranking := func(who *member) map[string]any {
		rec := s.read(t, who, "/api/v1/profiles/ranked")
		var body map[string]any
		if rec.Code != http.StatusOK || json.Unmarshal(rec.Body.Bytes(), &body) != nil {
			t.Fatalf("%d %s", rec.Code, rec.Body.String())
		}
		r, _ := body["ranking"].(map[string]any)
		return r
	}
	if r := ranking(nil); r != nil {
		t.Fatalf("anonymous sees a members-only github's rank: %v", r)
	}
	if r := ranking(&viewer); r == nil || r["country"] != "SN" || r["rank"] != float64(2) || r["generated_at"] == nil {
		t.Fatalf("member: %v", r)
	}
	if _, err := s.conn.Exec(ctx, `INSERT INTO github_developer_optouts (login) VALUES ('ranked-gh')`); err != nil {
		t.Fatal(err)
	}
	if r := ranking(&viewer); r != nil {
		t.Fatalf("opted out: %v", r)
	}
}
