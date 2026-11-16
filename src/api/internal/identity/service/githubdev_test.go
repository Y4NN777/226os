package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

type fakeGitHub struct {
	devs    []domain.GitHubDeveloper // sorted by (rank, country, login)
	queries []domain.ExternalQuery
	listErr error
	ranks   map[string]domain.Ranking
	rankErr error
	asked   []string
}

func (f *fakeGitHub) ListExternal(_ context.Context, q domain.ExternalQuery) ([]domain.GitHubDeveloper, error) {
	f.queries = append(f.queries, q)
	if f.listErr != nil {
		return nil, f.listErr
	}
	var out []domain.GitHubDeveloper
	for _, d := range f.devs {
		if q.Country != "" && d.Country != q.Country {
			continue
		}
		if q.AfterRank > 0 && !afterKey(d, q) {
			continue
		}
		if len(out) == q.Limit {
			break
		}
		out = append(out, d)
	}
	return out, nil
}

func afterKey(d domain.GitHubDeveloper, q domain.ExternalQuery) bool {
	if d.Rank != q.AfterRank {
		return d.Rank > q.AfterRank
	}
	if d.Country != q.AfterCountry {
		return d.Country > q.AfterCountry
	}
	return d.Login > q.AfterLogin
}

func (f *fakeGitHub) GitHubRank(_ context.Context, login string) (domain.Ranking, error) {
	f.asked = append(f.asked, login)
	if f.rankErr != nil {
		return domain.Ranking{}, f.rankErr
	}
	r, ok := f.ranks[login]
	if !ok {
		return domain.Ranking{}, domain.ErrNotFound
	}
	return r, nil
}

func ghDevs(n int) []domain.GitHubDeveloper {
	out := make([]domain.GitHubDeveloper, n)
	for i := range out {
		out[i] = domain.GitHubDeveloper{Country: "BF", Login: fmt.Sprintf("gh%02d", i+1), Rank: i + 1, Name: "G"}
	}
	return out
}

func githubSvc(t *testing.T) (*Service, *fakeStore, *fakeGitHub) {
	t.Helper()
	svc, st := directorySvc(t)
	gh := &fakeGitHub{}
	WithGitHubDevelopers(gh)(svc)
	return svc, st, gh
}

func ptr[T any](v T) *T { return &v }

// walk pages through the directory and returns "m:<username>" and "g:<login>" in order.
func walk(t *testing.T, svc *Service, req DirectoryRequest) []string {
	t.Helper()
	var seen []string
	for pages := 0; pages < 20; pages++ {
		page, err := svc.ListProfiles(context.Background(), req)
		if err != nil {
			t.Fatal(err)
		}
		if n := len(page.Items) + len(page.External); req.Limit != nil && n > *req.Limit {
			t.Fatalf("page holds %d entries, limit %d", n, *req.Limit)
		}
		for _, e := range page.Items {
			seen = append(seen, "m:"+e.Username)
		}
		for _, d := range page.External {
			seen = append(seen, "g:"+d.Login)
		}
		if page.NextCursor == "" {
			return seen
		}
		req.Cursor = page.NextCursor
	}
	t.Fatal("too many pages")
	return nil
}

func TestListProfilesMembersThenExternalAcrossPages(t *testing.T) {
	for _, limit := range []int{1, 2, 3, 5} { // 3 members: 3 fills a page exactly, the cursor moves to GitHub
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			svc, st, gh := githubSvc(t)
			st.directory = entries(3)
			gh.devs = ghDevs(4)
			got := walk(t, svc, DirectoryRequest{ClientIP: "198.51.100.7", Limit: ptr(limit)})
			want := []string{"m:user01", "m:user02", "m:user03", "g:gh01", "g:gh02", "g:gh03", "g:gh04"}
			if !slices.Equal(got, want) {
				t.Fatalf("walked %v, want %v", got, want)
			}
		})
	}
}

func TestListProfilesSourceFilters(t *testing.T) {
	svc, st, gh := githubSvc(t)
	st.directory = entries(2)
	gh.devs = ghDevs(2)
	if got := walk(t, svc, DirectoryRequest{ClientIP: "198.51.100.7", Source: "members"}); !slices.Equal(got, []string{"m:user01", "m:user02"}) {
		t.Fatalf("members: %v", got)
	}
	if got := walk(t, svc, DirectoryRequest{ClientIP: "198.51.100.7", Source: "github", Limit: ptr(1)}); !slices.Equal(got, []string{"g:gh01", "g:gh02"}) {
		t.Fatalf("github: %v", got)
	}
	// language or skill filter members only: external entries are omitted.
	if got := walk(t, svc, DirectoryRequest{ClientIP: "198.51.100.7", Language: ptr("fr")}); !slices.Equal(got, []string{"m:user01", "m:user02"}) {
		t.Fatalf("language: %v", got)
	}
	if got := walk(t, svc, DirectoryRequest{ClientIP: "198.51.100.7", Source: "github", Skill: ptr("go")}); len(got) != 0 {
		t.Fatalf("github with a skill: %v", got)
	}
	if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7", Source: "everyone"}); appErr(t, err).Code != apperr.CodeValidation {
		t.Fatalf("bad source: %v", err)
	}
}

func TestListProfilesExternalQueryCarriesAudienceAndCountry(t *testing.T) {
	svc, _, gh := githubSvc(t)
	WithDemoProfiles(true)(svc)
	if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{
		CallerUserID: "00000000-0000-7000-8000-0000000000aa", Country: ptr("BF"), Limit: ptr(4),
	}); err != nil {
		t.Fatal(err)
	}
	q := gh.queries[0]
	if !q.Member || !q.IncludeDemo || q.Country != "BF" || q.Limit != 5 || q.AfterRank != 0 {
		t.Fatalf("query %+v", q)
	}
}

func TestListProfilesCursorIsBoundToSource(t *testing.T) {
	svc, st, gh := githubSvc(t)
	st.directory = entries(3)
	gh.devs = ghDevs(3)
	page, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7", Limit: ptr(1)})
	if err != nil || page.NextCursor == "" {
		t.Fatal(page, err)
	}
	if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7", Limit: ptr(1), Source: "members", Cursor: page.NextCursor}); appErr(t, err).Code != apperr.CodeInvalidCursor {
		t.Fatalf("a cursor of another source: %v", err)
	}
}

func TestListProfilesRejectsForgedGitHubCursors(t *testing.T) {
	svc, _, _ := githubSvc(t)
	codec, _ := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	hash := lengthPrefixed("directory", "", "", "", "all")
	for _, p := range []cursor.Payload{
		{Sort: "ada", ID: "not-a-uuid"}, {Sort: "gh:0:BF", ID: "ada"}, {Sort: "gh:501:BF", ID: "ada"}, {Sort: "gh:x:BF", ID: "ada"},
		{Sort: "gh:1:bf", ID: "ada"}, {Sort: "gh:1:BF", ID: "-ada"}, {Sort: "gh:1", ID: "ada"}, {Sort: "gh:", ID: "ada"},
	} {
		p.FilterHash = hash
		if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7", Cursor: codec.Encode(p)}); appErr(t, err).Code != apperr.CodeInvalidCursor {
			t.Errorf("%+v: %v", p, err)
		}
	}
}

func TestListProfilesExternalStoreError(t *testing.T) {
	svc, _, gh := githubSvc(t)
	gh.listErr = errors.New("db down")
	if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7"}); appErr(t, err).Code != apperr.CodeInternal {
		t.Fatalf("%v", err)
	}
}

func TestListProfilesWithoutTheGitHubPortListsMembersOnly(t *testing.T) {
	svc, st := directorySvc(t)
	st.directory = entries(2)
	if got := walk(t, svc, DirectoryRequest{ClientIP: "198.51.100.7", Source: "github"}); len(got) != 0 {
		t.Fatalf("github without a port: %v", got)
	}
}

func TestGetProfileRankingFollowsGitHubVisibility(t *testing.T) {
	svc, st, _ := testSvc(t)
	gh := &fakeGitHub{ranks: map[string]domain.Ranking{"ada-gh": {Country: "BF", Rank: 3, GeneratedAt: time.Unix(0, 0)}}}
	WithGitHubDevelopers(gh)(svc)
	login := "ada-gh"
	rec := domain.ProfileRecord{AccountStatus: "active", Profile: domain.Profile{
		UserID: "00000000-0000-7000-8000-000000000001", Username: "ada", DisplayName: "Ada", GitHub: &login,
		Visibility: domain.VisibilityPublic, VisGitHub: domain.VisibilityMembers, VisOpenCollab: domain.VisibilityPublic,
	}}
	st.profiles["ada"] = rec
	anon, err := svc.GetProfile(context.Background(), "ada", "")
	if err != nil || anon.Ranking != nil || len(gh.asked) != 0 {
		t.Fatalf("anonymous must not see a ranking of a members-only github: %+v %v %v", anon.Ranking, err, gh.asked)
	}
	member, err := svc.GetProfile(context.Background(), "ada", "00000000-0000-7000-8000-0000000000aa")
	if err != nil || member.Ranking == nil || member.Ranking.Rank != 3 {
		t.Fatalf("member: %+v %v", member.Ranking, err)
	}
	delete(gh.ranks, "ada-gh")
	if p, err := svc.GetProfile(context.Background(), "ada", "00000000-0000-7000-8000-0000000000aa"); err != nil || p.Ranking != nil {
		t.Fatalf("unranked: %+v %v", p.Ranking, err)
	}
	gh.rankErr = errors.New("db down")
	if _, err := svc.GetProfile(context.Background(), "ada", "00000000-0000-7000-8000-0000000000aa"); appErr(t, err).Code != apperr.CodeInternal {
		t.Fatalf("rank error: %v", err)
	}
}
