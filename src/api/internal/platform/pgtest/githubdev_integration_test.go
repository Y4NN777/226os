//go:build integration

package pgtest

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/repository"
)

var ghGenerated = time.Date(2026, 11, 14, 6, 0, 0, 0, time.UTC)

func ghDev(country, login string, rank int) domain.GitHubDeveloper {
	return domain.GitHubDeveloper{
		Country: country, Login: login, Rank: rank, Name: "Name " + login,
		AvatarURL: "https://avatars.githubusercontent.com/u/1?v=4", Contributions: 1000 - rank,
		Organizations: []string{"org-" + login}, GeneratedAt: ghGenerated,
	}
}

func ghLogins(devs []domain.GitHubDeveloper) []string {
	out := make([]string, len(devs))
	for i, d := range devs {
		out[i] = d.Country + "/" + d.Login
	}
	return out
}

func TestGitHubDevelopersReplaceIsOneTransaction(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	repo := repository.New(conn)
	if _, ok, err := repo.GitHubGeneratedAt(ctx, "BF"); err != nil || ok {
		t.Fatalf("empty country: ok=%v err=%v", ok, err)
	}
	if err := repo.ReplaceGitHubDevelopers(ctx, "BF", []domain.GitHubDeveloper{ghDev("BF", "ada", 1), ghDev("BF", "bob", 2)}); err != nil {
		t.Fatal(err)
	}
	got, ok, err := repo.GitHubGeneratedAt(ctx, "BF")
	if err != nil || !ok || !got.Equal(ghGenerated) {
		t.Fatalf("generated = %v %v %v", got, ok, err)
	}
	// A failing row (avatar on another host) rolls the whole replace back: the previous snapshot stays.
	bad := ghDev("BF", "eve", 2)
	bad.AvatarURL = "https://evil.example/x"
	if err := repo.ReplaceGitHubDevelopers(ctx, "BF", []domain.GitHubDeveloper{ghDev("BF", "carl", 1), bad}); err == nil {
		t.Fatal("a bad row must fail the replace")
	}
	list, err := repo.ListExternal(ctx, domain.ExternalQuery{Country: "BF", Limit: 10})
	if err != nil || !slices.Equal(ghLogins(list), []string{"BF/ada", "BF/bob"}) {
		t.Fatalf("after a failed replace: %v %v", ghLogins(list), err)
	}
	// A good replace swaps the country and leaves the others alone.
	if err := repo.ReplaceGitHubDevelopers(ctx, "SN", []domain.GitHubDeveloper{ghDev("SN", "sow", 1)}); err != nil {
		t.Fatal(err)
	}
	if err := repo.ReplaceGitHubDevelopers(ctx, "BF", []domain.GitHubDeveloper{ghDev("BF", "carl", 1)}); err != nil {
		t.Fatal(err)
	}
	list, err = repo.ListExternal(ctx, domain.ExternalQuery{Limit: 10})
	if err != nil || !slices.Equal(ghLogins(list), []string{"BF/carl", "SN/sow"}) {
		t.Fatalf("after replace: %v %v", ghLogins(list), err)
	}
	if list[0].Name != "Name carl" || list[0].AvatarURL == "" || list[0].Contributions != 999 ||
		!slices.Equal(list[0].Organizations, []string{"org-carl"}) || !list[0].GeneratedAt.Equal(ghGenerated) {
		t.Fatalf("mapped row %+v", list[0])
	}
}

func TestExternalDevelopersKeysetAcrossCountries(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	repo := repository.New(conn)
	// "twice" is ranked 1 in two countries: the keyset must neither skip nor repeat it.
	for _, c := range []struct {
		cc   string
		devs []domain.GitHubDeveloper
	}{
		{"BF", []domain.GitHubDeveloper{ghDev("BF", "twice", 1), ghDev("BF", "b2", 2)}},
		{"SN", []domain.GitHubDeveloper{ghDev("SN", "twice", 1), ghDev("SN", "s2", 2)}},
	} {
		if err := repo.ReplaceGitHubDevelopers(ctx, c.cc, c.devs); err != nil {
			t.Fatal(err)
		}
	}
	var walked []string
	q := domain.ExternalQuery{Limit: 1}
	for range 10 {
		page, err := repo.ListExternal(ctx, q)
		if err != nil {
			t.Fatal(err)
		}
		if len(page) == 0 {
			break
		}
		walked = append(walked, ghLogins(page)...)
		last := page[len(page)-1]
		q.AfterRank, q.AfterCountry, q.AfterLogin = last.Rank, last.Country, last.Login
	}
	if want := []string{"BF/twice", "SN/twice", "BF/b2", "SN/s2"}; !slices.Equal(walked, want) {
		t.Fatalf("walked %v, want %v", walked, want)
	}
	only, err := repo.ListExternal(ctx, domain.ExternalQuery{Country: "SN", Limit: 10})
	if err != nil || !slices.Equal(ghLogins(only), []string{"SN/twice", "SN/s2"}) {
		t.Fatalf("country filter: %v %v", ghLogins(only), err)
	}
}

func TestExternalDevelopersDedupAndOptOut(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	repo := repository.New(conn)
	seedTags(ctx, t, conn)
	// Members whose github_username matches a ranked login, case differing on purpose.
	listed := newPerson(1, "listed")
	listed.github = "Listed-GH"
	listed.country = "SN" // dedup ignores the page's country filter
	membersOnly := newPerson(2, "membersonly")
	membersOnly.github = "membersonly-gh"
	membersOnly.githubVis = "members"
	hidden := newPerson(3, "hidden")
	hidden.github = "hidden-gh"
	hidden.githubVis = "private"
	notListable := newPerson(4, "closed")
	notListable.github = "closed-gh"
	notListable.collab = false
	for _, p := range []person{listed, membersOnly, hidden, notListable} {
		p.insert(ctx, t, conn)
	}
	demo := newPerson(5, "demo-gh-member")
	demo.github = "demo-gh"
	demo.insert(ctx, t, conn)
	if _, err := conn.Exec(ctx, `UPDATE users SET provenance = 'demo' WHERE id = $1`, personID(5)); err != nil {
		t.Fatal(err)
	}
	devs := []domain.GitHubDeveloper{
		ghDev("BF", "listed-gh", 1), ghDev("BF", "membersonly-gh", 2), ghDev("BF", "hidden-gh", 3),
		ghDev("BF", "closed-gh", 4), ghDev("BF", "demo-gh", 5), ghDev("BF", "Gone", 6), ghDev("BF", "plain", 7),
	}
	if err := repo.ReplaceGitHubDevelopers(ctx, "BF", devs); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO github_developer_optouts (login) VALUES ('gone')`); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		q    domain.ExternalQuery
		want []string
	}{
		{"anonymous", domain.ExternalQuery{Country: "BF", Limit: 20},
			[]string{"BF/membersonly-gh", "BF/hidden-gh", "BF/closed-gh", "BF/demo-gh", "BF/plain"}},
		{"member", domain.ExternalQuery{Member: true, Country: "BF", Limit: 20},
			[]string{"BF/hidden-gh", "BF/closed-gh", "BF/demo-gh", "BF/plain"}},
		{"member with demo accounts (local-dev)", domain.ExternalQuery{Member: true, IncludeDemo: true, Country: "BF", Limit: 20},
			[]string{"BF/hidden-gh", "BF/closed-gh", "BF/plain"}},
	}
	for _, tc := range cases {
		got, err := repo.ListExternal(ctx, tc.q)
		if err != nil || !slices.Equal(ghLogins(got), tc.want) {
			t.Fatalf("%s: %v %v, want %v", tc.name, ghLogins(got), err, tc.want)
		}
	}
	// Rank on a profile: case-insensitive, best rank, none once opted out.
	r, err := repo.GitHubRank(ctx, "LISTED-gh")
	if err != nil || r.Country != "BF" || r.Rank != 1 || !r.GeneratedAt.Equal(ghGenerated) {
		t.Fatalf("rank = %+v %v", r, err)
	}
	if _, err := repo.GitHubRank(ctx, "gone"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("opted-out rank: %v", err)
	}
	if _, err := repo.GitHubRank(ctx, "nobody"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("unknown rank: %v", err)
	}
}

// Every list of the GitHub segment is bounded (DATABASE_DESIGN.md section 7); the plans on 5000 rows (ten
// countries of 500) are logged as EXPLAIN (ANALYZE, BUFFERS) evidence.
func TestExternalDevelopersPlanIsBounded(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	for _, cc := range []string{"BF", "CI", "SN", "ML", "NE", "BJ", "TG", "GN", "GH", "NG"} {
		if _, err := conn.Exec(ctx, `INSERT INTO github_developers (country_code, login, rank, name, avatar_url, contributions, organizations, generated_at)
			SELECT $1, 'dev-' || lower($1) || '-' || g, g, 'Dev', 'https://avatars.githubusercontent.com/u/' || g, 1000 - g, '{}', now()
			FROM generate_series(1, 500) g`, cc); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := conn.Exec(ctx, `ANALYZE`); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, conn.Config().ConnString())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	rec := &recordingDB{db: pool}
	repo := repository.New(beginRecorder{rec, pool})
	for name, q := range map[string]domain.ExternalQuery{
		"first page, all countries": {Limit: 21},
		"country filter":            {Country: "SN", Limit: 21},
		"keyset page":               {AfterRank: 250, AfterCountry: "BF", AfterLogin: "dev-bf-250", Limit: 21},
	} {
		rec.reset()
		rows, err := repo.ListExternal(ctx, q)
		if err != nil || len(rows) != 21 {
			t.Fatalf("%s: %d rows, %v", name, len(rows), err)
		}
		stmt := rec.first(t)
		plan := explainPlan(ctx, t, conn, stmt.sql, stmt.args...)
		t.Logf("external plan, %s:\n%s", name, plan)
		if !strings.Contains(plan, "Limit") {
			t.Fatalf("%s plan has no Limit:\n%s", name, plan)
		}
	}
}
