//go:build integration

package pgtest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/repository"
)

// E5 people directory (GET /profiles, C14b) on the production router over a migrated database. Every
// person below is one visibility case; the tests assert who is listed, which fields each audience sees,
// and that no filter matches a field the caller cannot see.

type person struct {
	n                     int
	username, status      string
	profileVis, collabVis string
	collab                bool
	country, countryVis   string
	languages             []string
	languagesVis          string
	skills                []string
	skillsVis             string
	github, githubVis     string
}

// newPerson is a fully public, listable member from Burkina Faso who speaks French and knows Go.
func newPerson(n int, username string) person {
	return person{
		n: n, username: username, status: "active", profileVis: "public", collabVis: "public", collab: true,
		country: "BF", countryVis: "public", languages: []string{"fr"}, languagesVis: "public",
		skills: []string{"go"}, skillsVis: "public", github: username + "-gh", githubVis: "public",
	}
}

func personID(n int) string { return fmt.Sprintf("00000000-0000-7000-8000-%012d", n) }

func (p person) insert(ctx context.Context, t *testing.T, conn *pgx.Conn) {
	t.Helper()
	id := personID(p.n)
	deleted := "NULL"
	if p.status == "deleted" {
		deleted = "now()"
	}
	if _, err := conn.Exec(ctx, fmt.Sprintf(
		`INSERT INTO users (id, email, status, deleted_at) VALUES ($1, $2, '%s', %s)`, p.status, deleted),
		id, fmt.Sprintf("secret-mail-%d@example.com", p.n)); err != nil {
		t.Fatalf("user %s: %v", p.username, err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO profiles (user_id, username, display_name, country_code, languages, github_username,
		open_to_collab, country_visibility, languages_visibility, skills_visibility, github_visibility,
		open_to_collab_visibility, profile_visibility)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), $7, $8, $9, $10, $11, $12, $13)`,
		id, p.username, "Display "+p.username, p.country, p.languages, p.github, p.collab,
		p.countryVis, p.languagesVis, p.skillsVis, p.githubVis, p.collabVis, p.profileVis); err != nil {
		t.Fatalf("profile %s: %v", p.username, err)
	}
	for _, slug := range p.skills {
		if _, err := conn.Exec(ctx, `INSERT INTO user_skills (user_id, tag_id) SELECT $1, id FROM tags WHERE slug = $2`, id, slug); err != nil {
			t.Fatalf("skill %s: %v", slug, err)
		}
	}
}

func seedTags(ctx context.Context, t *testing.T, conn *pgx.Conn) {
	t.Helper()
	if _, err := conn.Exec(ctx, `INSERT INTO tags (slug, label) VALUES ('go', 'Go'), ('rust', 'Rust')`); err != nil {
		t.Fatal(err)
	}
}

// matrix is every listing case with who may see the person: A for anonymous, M for members.
func matrix() []struct {
	p            person
	anon, member bool
} {
	mod := func(n int, name string, f func(*person)) person {
		p := newPerson(n, name)
		f(&p)
		return p
	}
	return []struct {
		p            person
		anon, member bool
	}{
		{newPerson(1, "open"), true, true},
		{mod(2, "profile-members", func(p *person) { p.profileVis = "members" }), false, true},
		{mod(3, "profile-private", func(p *person) { p.profileVis = "private" }), false, false},
		{mod(4, "collab-off", func(p *person) { p.collab = false }), false, false},
		{mod(5, "collab-members", func(p *person) { p.collabVis = "members" }), false, true},
		{mod(6, "collab-private", func(p *person) { p.collabVis = "private" }), false, false},
		{mod(7, "suspended", func(p *person) { p.status = "suspended" }), false, false},
		{mod(8, "deleted", func(p *person) { p.status = "deleted" }), false, false},
		{mod(9, "hidden-country", func(p *person) { p.country, p.countryVis = "ML", "members" }), true, true},
		{mod(10, "hidden-languages", func(p *person) { p.languages, p.languagesVis = []string{"bm"}, "private" }), true, true},
		{mod(11, "hidden-skills", func(p *person) { p.skills, p.skillsVis = []string{"rust"}, "members" }), true, true},
		{mod(12, "hidden-github", func(p *person) { p.githubVis = "private" }), true, true},
	}
}

func seedMatrix(ctx context.Context, t *testing.T, conn *pgx.Conn) {
	t.Helper()
	seedTags(ctx, t, conn)
	for _, c := range matrix() {
		c.p.insert(ctx, t, conn)
	}
}

type directoryPage struct {
	Items      []map[string]any `json:"items"`
	NextCursor *string          `json:"next_cursor"`
}

func (d directoryPage) usernames() []string {
	out := make([]string, len(d.Items))
	for i, it := range d.Items {
		out[i], _ = it["username"].(string)
	}
	return out
}

func (d directoryPage) item(t *testing.T, username string) map[string]any {
	t.Helper()
	for _, it := range d.Items {
		if it["username"] == username {
			return it
		}
	}
	t.Fatalf("%s is not listed in %v", username, d.usernames())
	return nil
}

func (s *s2Stack) directory(t *testing.T, m *member, query string) (directoryPage, *httptest.ResponseRecorder) {
	t.Helper()
	path := "/api/v1/profiles"
	if query != "" {
		path += "?" + query
	}
	rec := s.read(t, m, path)
	var page directoryPage
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
			t.Fatalf("decode %s: %v", rec.Body.String(), err)
		}
	}
	return page, rec
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

func TestDirectoryVisibilityMatrix(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedMatrix(ctx, t, s.conn)
	viewer := s.signIn(t, "dir-viewer", true, 60)

	var wantAnon, wantMember []string
	for _, c := range matrix() {
		if c.anon {
			wantAnon = append(wantAnon, c.p.username)
		}
		if c.member {
			wantMember = append(wantMember, c.p.username)
		}
	}
	slices.Sort(wantAnon)
	slices.Sort(wantMember)

	for _, tc := range []struct {
		name string
		m    *member
		want []string
	}{{"anonymous", nil, wantAnon}, {"member", &viewer, wantMember}} {
		t.Run(tc.name, func(t *testing.T) {
			page, rec := s.directory(t, tc.m, "")
			if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("status %d cache-control %q: %s", rec.Code, rec.Header().Get("Cache-Control"), rec.Body.String())
			}
			if got := page.usernames(); !slices.Equal(got, tc.want) {
				t.Fatalf("listed %v, want %v", got, tc.want)
			}
			if page.NextCursor != nil {
				t.Fatalf("a single page has a next_cursor: %v", *page.NextCursor)
			}
		})
	}
}

func TestDirectoryFieldsFollowFieldVisibility(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedMatrix(ctx, t, s.conn)
	viewer := s.signIn(t, "dir-viewer", true, 61)

	anon, _ := s.directory(t, nil, "")
	member, _ := s.directory(t, &viewer, "")
	all := []string{"country", "display_name", "github", "languages", "skills", "username"}

	if got := keys(anon.item(t, "open")); !slices.Equal(got, all) {
		t.Fatalf("open (anonymous) keys %v", got)
	}
	// Each hidden-* person hides exactly one field from anonymous callers and shows it to members when it is members-only.
	for _, tc := range []struct {
		username, hidden string
		visibleToMember  bool
	}{
		{"hidden-country", "country", true},
		{"hidden-languages", "languages", false}, // private: nobody
		{"hidden-skills", "skills", true},
		{"hidden-github", "github", false}, // private: nobody
	} {
		withoutHidden := slices.DeleteFunc(slices.Clone(all), func(k string) bool { return k == tc.hidden })
		if got := keys(anon.item(t, tc.username)); !slices.Equal(got, withoutHidden) {
			t.Errorf("%s anonymous keys %v, want %v", tc.username, got, withoutHidden)
		}
		wantMember := all
		if !tc.visibleToMember {
			wantMember = withoutHidden
		}
		if got := keys(member.item(t, tc.username)); !slices.Equal(got, wantMember) {
			t.Errorf("%s member keys %v, want %v", tc.username, got, wantMember)
		}
	}
	if v := member.item(t, "hidden-country")["country"]; v != "ML" {
		t.Errorf("member sees country %v", v)
	}
}

func TestDirectoryFiltersNeverMatchAHiddenField(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedMatrix(ctx, t, s.conn)
	viewer := s.signIn(t, "dir-viewer", true, 62)

	for _, tc := range []struct {
		name, query  string
		anon, member []string
	}{
		{"country public", "country=BF", []string{"hidden-github", "hidden-languages", "hidden-skills", "open"}, []string{"collab-members", "hidden-github", "hidden-languages", "hidden-skills", "open", "profile-members"}},
		// hidden-country lives in ML with a members-only country: invisible to anonymous, a match for members.
		{"country members-only", "country=ML", nil, []string{"hidden-country"}},
		{"language public", "language=fr", []string{"hidden-country", "hidden-github", "hidden-skills", "open"}, []string{"collab-members", "hidden-country", "hidden-github", "hidden-skills", "open", "profile-members"}},
		// hidden-languages speaks bm with a private list: it matches nobody.
		{"language private", "language=bm", nil, nil},
		{"skill public", "skill=go", []string{"hidden-country", "hidden-github", "hidden-languages", "open"}, []string{"collab-members", "hidden-country", "hidden-github", "hidden-languages", "open", "profile-members"}},
		// hidden-skills has rust with a members-only list: invisible to anonymous, a match for members.
		{"skill members-only", "skill=rust", nil, []string{"hidden-skills"}},
		{"no match", "country=BF&language=bm", nil, nil},
		{"combined", "country=BF&language=fr&skill=go", []string{"hidden-github", "open"}, []string{"collab-members", "hidden-github", "open", "profile-members"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, who := range []struct {
				label string
				m     *member
				want  []string
			}{{"anonymous", nil, tc.anon}, {"member", &viewer, tc.member}} {
				page, rec := s.directory(t, who.m, tc.query)
				if rec.Code != http.StatusOK {
					t.Fatalf("%s: %d %s", who.label, rec.Code, rec.Body.String())
				}
				want := slices.Clone(who.want)
				if want == nil {
					want = []string{}
				}
				if got := page.usernames(); !slices.Equal(got, want) {
					t.Errorf("%s %s: %v, want %v", who.label, tc.query, got, want)
				}
			}
		})
	}
}

func TestDirectoryNeverReturnsAnEmail(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedMatrix(ctx, t, s.conn)
	viewer := s.signIn(t, "dir-viewer", true, 63)
	allowed := map[string]bool{"username": true, "display_name": true, "country": true, "languages": true, "skills": true, "github": true}
	for _, m := range []*member{nil, &viewer} {
		page, rec := s.directory(t, m, "")
		body := strings.ToLower(rec.Body.String())
		for _, leak := range []string{"email", "@example.com", "secret-mail", "dir-viewer@", "user_id", "provenance", "visibility", "open_to_collab"} {
			if strings.Contains(body, leak) {
				t.Fatalf("body leaks %q: %s", leak, rec.Body.String())
			}
		}
		for _, it := range page.Items {
			for k := range it {
				if !allowed[k] {
					t.Fatalf("unexpected key %q in %v", k, it)
				}
			}
		}
	}
}

func TestDirectoryPaginationIsStable(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedTags(ctx, t, s.conn)
	for i := 1; i <= 7; i++ {
		newPerson(i, fmt.Sprintf("page-%02d", 8-i)).insert(ctx, t, s.conn) // inserted in reverse username order
	}
	var walked []string
	cursor := ""
	for pages := 0; pages < 10; pages++ {
		q := "limit=3"
		if cursor != "" {
			q += "&cursor=" + cursor
		}
		page, rec := s.directory(t, nil, q)
		if rec.Code != http.StatusOK {
			t.Fatalf("page %d: %d %s", pages, rec.Code, rec.Body.String())
		}
		walked = append(walked, page.usernames()...)
		if page.NextCursor == nil {
			break
		}
		cursor = *page.NextCursor
	}
	want := []string{"page-01", "page-02", "page-03", "page-04", "page-05", "page-06", "page-07"}
	if !slices.Equal(walked, want) {
		t.Fatalf("walked %v, want %v (ordered, no duplicate, no gap)", walked, want)
	}
	// A row inserted behind the cursor does not move the next page.
	first, _ := s.directory(t, nil, "limit=3")
	newPerson(20, "page-00").insert(ctx, t, s.conn)
	second, _ := s.directory(t, nil, "limit=3&cursor="+*first.NextCursor)
	if got := second.usernames(); !slices.Equal(got, []string{"page-04", "page-05", "page-06"}) {
		t.Fatalf("second page after an insert behind the cursor: %v", got)
	}
}

func TestDirectoryRejectsBadInput(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedTags(ctx, t, s.conn)
	for i := 1; i <= 3; i++ {
		newPerson(i, fmt.Sprintf("bad-%d", i)).insert(ctx, t, s.conn)
	}
	first, _ := s.directory(t, nil, "limit=1")
	if first.NextCursor == nil {
		t.Fatal("no cursor to tamper with")
	}
	tampered := (*first.NextCursor)[:len(*first.NextCursor)-3] + "AAA"
	for _, tc := range []struct {
		name, query string
		status      int
		code        string
	}{
		{"limit 51", "limit=51", http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"limit 0", "limit=0", http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"limit not a number", "limit=many", http.StatusBadRequest, "MALFORMED_REQUEST"},
		{"country lower case", "country=bf", http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"language upper case", "language=FR", http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"skill injection", "skill=x%27%3B+DROP+TABLE+profiles%3B--", http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"tampered cursor", "limit=1&cursor=" + tampered, http.StatusBadRequest, "INVALID_CURSOR"},
		{"cursor with other filters", "limit=1&country=BF&cursor=" + *first.NextCursor, http.StatusBadRequest, "INVALID_CURSOR"},
		{"limit 50 is fine", "limit=50", http.StatusOK, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := s.read(t, nil, "/api/v1/profiles?"+tc.query)
			if rec.Code != tc.status || (tc.code != "" && !strings.Contains(rec.Body.String(), tc.code)) {
				t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
			}
		})
	}
}

// The owner has no branch of its own: a private profile is not in its owner's list, and once the
// profile is members-only the owner sees its own row like any other member would.
func TestDirectoryOwnerSeesItsRowLikeAnyListing(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedTags(ctx, t, s.conn)
	owner := s.signIn(t, "dir-owner", true, 64)
	set := func(profileVis, countryVis string) {
		t.Helper()
		if _, err := s.conn.Exec(ctx, `UPDATE profiles SET open_to_collab = true, country_code = 'BF', profile_visibility = $2, country_visibility = $3 WHERE user_id = $1`,
			owner.id, profileVis, countryVis); err != nil {
			t.Fatal(err)
		}
	}

	set("private", "private")
	if page, _ := s.directory(t, &owner, ""); slices.Contains(page.usernames(), "dir-owner") {
		t.Fatal("a private profile is listed to its owner")
	}
	set("members", "private")
	page, _ := s.directory(t, &owner, "")
	it := page.item(t, "dir-owner")
	if _, ok := it["country"]; ok {
		t.Fatalf("the owner gets extra fields in the listing: %v", it)
	}
	if anon, _ := s.directory(t, nil, ""); slices.Contains(anon.usernames(), "dir-owner") {
		t.Fatal("a members-only profile is listed to anonymous callers")
	}
	if page, _ := s.directory(t, &owner, "country=BF"); slices.Contains(page.usernames(), "dir-owner") {
		t.Fatal("the owner's own private country matched a filter")
	}
}

// A role change or a status change takes effect on the next request: nothing is cached.
func TestDirectoryDropsAnAccountOnceItIsSuspendedOrDeleted(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedTags(ctx, t, s.conn)
	newPerson(1, "alive").insert(ctx, t, s.conn)
	if page, _ := s.directory(t, nil, ""); !slices.Equal(page.usernames(), []string{"alive"}) {
		t.Fatalf("listed %v", page.usernames())
	}
	for _, stmt := range []string{
		`UPDATE users SET status = 'suspended' WHERE id = $1`,
		`UPDATE users SET status = 'deleted', deleted_at = now() WHERE id = $1`,
	} {
		if _, err := s.conn.Exec(ctx, `UPDATE users SET status = 'active', deleted_at = NULL WHERE id = $1`, personID(1)); err != nil {
			t.Fatal(err)
		}
		if _, err := s.conn.Exec(ctx, stmt, personID(1)); err != nil {
			t.Fatal(err)
		}
		if page, _ := s.directory(t, nil, ""); len(page.Items) != 0 {
			t.Fatalf("%s: still listed %v", stmt, page.usernames())
		}
	}
}

func TestDirectoryRateLimit(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedTags(ctx, t, s.conn)
	newPerson(1, "limited").insert(ctx, t, s.conn)
	viewer := s.signIn(t, "dir-viewer", true, 65)
	// The limiter uses fixed one-minute windows: each 61-call burst must start inside one window,
	// or a minute rollover resets the bucket and the 61st call is allowed.
	awaitFreshMinute(t)
	for i := 1; i <= 60; i++ {
		if rec := s.read(t, nil, "/api/v1/profiles"); rec.Code != http.StatusOK {
			t.Fatalf("anonymous call %d: %d %s", i, rec.Code, rec.Body.String())
		}
	}
	rec := s.read(t, nil, "/api/v1/profiles")
	if rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" || !strings.Contains(rec.Body.String(), "RATE_LIMITED") {
		t.Fatalf("61st anonymous call: %d retry-after %q %s", rec.Code, rec.Header().Get("Retry-After"), rec.Body.String())
	}
	// The member has its own per-user budget, untouched by the anonymous network's.
	awaitFreshMinute(t)
	if rec := s.read(t, &viewer, "/api/v1/profiles"); rec.Code != http.StatusOK {
		t.Fatalf("member call after the anonymous limit: %d %s", rec.Code, rec.Body.String())
	}
	for i := 2; i <= 60; i++ {
		if rec := s.read(t, &viewer, "/api/v1/profiles"); rec.Code != http.StatusOK {
			t.Fatalf("member call %d: %d", i, rec.Code)
		}
	}
	if rec := s.read(t, &viewer, "/api/v1/profiles"); rec.Code != http.StatusTooManyRequests || rec.Header().Get("Retry-After") == "" {
		t.Fatalf("61st member call: %d %s", rec.Code, rec.Body.String())
	}
}

// The plan of the exact statement a request sends is logged as EXPLAIN (ANALYZE, BUFFERS) evidence
// (ADR-0005 point 6) on 3000 listable profiles: the page is bounded by its LIMIT and ordered by the
// username index.
func TestDirectoryListPlanIsBounded(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	seedTags(ctx, t, s.conn)
	for _, stmt := range []string{
		`INSERT INTO users (id, email)
		 SELECT ('00000000-0000-7000-8001-' || lpad(g::text, 12, '0'))::uuid, 'bulk-' || g || '@example.com' FROM generate_series(1, 3000) g`,
		`INSERT INTO profiles (user_id, username, display_name, country_code, languages, open_to_collab)
		 SELECT ('00000000-0000-7000-8001-' || lpad(g::text, 12, '0'))::uuid, 'bulk-' || lpad(g::text, 5, '0'), 'Bulk',
		        (ARRAY['BF','SN','CI','ML'])[1 + g % 4], ARRAY['fr'], true FROM generate_series(1, 3000) g`,
	} {
		if _, err := s.conn.Exec(ctx, stmt); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.conn.Exec(ctx, `INSERT INTO user_skills (user_id, tag_id) SELECT p.user_id, t.id FROM profiles p, tags t WHERE t.slug = 'go' AND p.username LIKE 'bulk-%' AND substr(p.username, 6)::int % 3 = 0`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.conn.Exec(ctx, `ANALYZE`); err != nil {
		t.Fatal(err)
	}
	rec := &recordingDB{db: s.pool}
	repo := repository.New(beginRecorder{rec, s.pool})
	for name, q := range map[string]domain.DirectoryQuery{
		"first page, anonymous": {Limit: 21},
		"first page, member":    {Member: true, Limit: 21},
		"country filter":        {Country: "SN", Limit: 21},
		"skill filter":          {Skill: "go", Limit: 21},
		"keyset page":           {AfterUsername: "bulk-01500", AfterID: personID(1), Limit: 21},
	} {
		rec.reset()
		rows, err := repo.ListProfiles(ctx, q)
		if err != nil || len(rows) != 21 {
			t.Fatalf("%s: %d rows, %v", name, len(rows), err)
		}
		stmt := rec.first(t)
		plan := explainPlan(ctx, t, s.conn, stmt.sql, stmt.args...)
		t.Logf("directory plan, %s:\n%s", name, plan)
		if !strings.Contains(plan, "Limit") {
			t.Fatalf("%s plan has no Limit:\n%s", name, plan)
		}
	}
}
