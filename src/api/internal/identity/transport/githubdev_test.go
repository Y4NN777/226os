package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

type fakeGitHub struct {
	devs []domain.GitHubDeveloper
	last domain.ExternalQuery
}

func (f *fakeGitHub) ListExternal(_ context.Context, q domain.ExternalQuery) ([]domain.GitHubDeveloper, error) {
	f.last = q
	return f.devs, nil
}

func (f *fakeGitHub) GitHubRank(context.Context, string) (domain.Ranking, error) {
	return domain.Ranking{Country: "BF", Rank: 7, GeneratedAt: time.Date(2026, 11, 14, 6, 0, 0, 0, time.UTC)}, nil
}

func githubStack(t *testing.T, gh *fakeGitHub) (*Handler, *fakeStore) {
	t.Helper()
	st := newFakeStore()
	svc := service.New(st, mailer.NewFake(), []byte("0123456789abcdef0123456789abcdef"),
		service.WithCursors(testCodec()), service.WithGitHubDevelopers(gh))
	return New(svc, false, nil), st
}

func TestListProfilesServesExternalEntries(t *testing.T) {
	gen := time.Date(2026, 11, 14, 6, 0, 0, 0, time.UTC)
	gh := &fakeGitHub{devs: []domain.GitHubDeveloper{{
		Country: "BF", Login: "Ada-GH", Rank: 1, Name: "Ada", AvatarURL: "https://avatars.githubusercontent.com/u/1",
		Contributions: 42, GeneratedAt: gen,
	}}}
	h, _ := githubStack(t, gh)
	src := apigen.ListProfilesParamsSourceGithub
	code, _, body := listProfiles(t, h, apigen.ListProfilesParams{Source: &src}, false)
	if code != http.StatusOK {
		t.Fatalf("%d %s", code, body)
	}
	var out struct {
		Items    []map[string]any `json:"items"`
		External []map[string]any `json:"external"`
	}
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatal(err)
	}
	if out.Items == nil || len(out.Items) != 0 || len(out.External) != 1 {
		t.Fatalf("body %s", body)
	}
	e := out.External[0]
	if e["login"] != "Ada-GH" || e["source"] != "committers.top" || e["rank"] != float64(1) || e["country"] != "BF" ||
		e["generated_at"] != "2026-11-14T06:00:00Z" || e["organizations"] == nil || len(e) != 9 {
		t.Fatalf("external entry %v", e)
	}
}

func TestListProfilesOmitsExternalWhenNone(t *testing.T) {
	h, _ := githubStack(t, &fakeGitHub{})
	members := apigen.ListProfilesParamsSourceMembers
	for _, p := range []apigen.ListProfilesParams{{}, {Source: &members}} {
		_, _, body := listProfiles(t, h, p, false)
		if body != "{\"items\":[],\"next_cursor\":null}\n" {
			t.Fatalf("body %q", body)
		}
	}
}

func TestPublicProfileJSONCarriesTheRanking(t *testing.T) {
	gen := time.Date(2026, 11, 14, 6, 0, 0, 0, time.UTC)
	out := publicProfileJSON(domain.PublicProfile{Username: "ada", Ranking: &domain.Ranking{Country: "SN", Rank: 2, GeneratedAt: gen}})
	if out.Ranking == nil || out.Ranking.Country != "SN" || out.Ranking.Rank != 2 || !out.Ranking.GeneratedAt.Equal(gen) {
		t.Fatalf("ranking %+v", out.Ranking)
	}
	if publicProfileJSON(domain.PublicProfile{Username: "ada"}).Ranking != nil {
		t.Fatal("no ranking: omitted")
	}
}
