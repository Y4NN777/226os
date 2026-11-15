//go:build integration

package pgtest

import (
	"net/http"
	"slices"
	"testing"
)

// Profile activity filters (prototype data gaps, owner decision 2026-11-15): listProjects owner,
// listQuestions author and answers_by. A username resolves only under the getProfile rule (ADR-0005,
// C14b): active account, and profile_visibility public, members for a member, or the caller's own
// profile. A hidden, private, suspended or unknown username is an empty 200 page with total 0, never a
// 404, and another user's private project is never returned.

func (s *s2Stack) setProfileVisibility(t *testing.T, username, visibility string) {
	t.Helper()
	if _, err := s.conn.Exec(t.Context(), `UPDATE profiles SET profile_visibility = $2 WHERE username = $1`, username, visibility); err != nil {
		t.Fatal(err)
	}
}

func (s *s2Stack) suspend(t *testing.T, username string) {
	t.Helper()
	if _, err := s.conn.Exec(t.Context(), `UPDATE users SET status = 'suspended' WHERE id = (SELECT user_id FROM profiles WHERE username = $1)`, username); err != nil {
		t.Fatal(err)
	}
}

// pageIDs reads one page (limit 50) and returns its ids and total.
func pageIDs(t *testing.T, s *s2Stack, viewer *member, path string) ([]string, int) {
	t.Helper()
	rec := s.read(t, viewer, path)
	if rec.Code != http.StatusOK {
		t.Fatalf("%s: %d %s", path, rec.Code, rec.Body.String())
	}
	got := decodeAs[struct {
		Items []struct {
			ID string `json:"id"`
		} `json:"items"`
		Total int `json:"total"`
	}](t, rec)
	ids := make([]string, len(got.Items))
	for i, it := range got.Items {
		ids[i] = it.ID
	}
	slices.Sort(ids)
	return ids, got.Total
}

func sorted(ids ...string) []string {
	out := slices.Clone(ids)
	slices.Sort(out)
	return out
}

func assertPage(t *testing.T, s *s2Stack, viewer *member, path string, want ...string) {
	t.Helper()
	ids, total := pageIDs(t, s, viewer, path+"&limit=50")
	if !slices.Equal(ids, sorted(want...)) || total != len(want) {
		t.Fatalf("%s: %v (total %d), want %v", path, ids, total, sorted(want...))
	}
}

func TestProfileActivityProjectsByOwner(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "act-alice", true, 81)
	bob := s.signIn(t, "act-bob", true, 82)
	mia := s.signIn(t, "act-mia", true, 83)
	pat := s.signIn(t, "act-pat", true, 84)
	sue := s.signIn(t, "act-sue", true, 85)
	alicePub := s.createProject(t, &alice, "https://github.com/act-alice/pub", "public").Id.String()
	alicePriv := s.createProject(t, &alice, "https://github.com/act-alice/priv", "private").Id.String()
	miaPub := s.createProject(t, &mia, "https://github.com/act-mia/pub", "public").Id.String()
	patPub := s.createProject(t, &pat, "https://github.com/act-pat/pub", "public").Id.String()
	patPriv := s.createProject(t, &pat, "https://github.com/act-pat/priv", "private").Id.String()
	s.createProject(t, &sue, "https://github.com/act-sue/pub", "public")
	s.setProfileVisibility(t, "act-mia", "members")
	s.setProfileVisibility(t, "act-pat", "private")
	s.suspend(t, "act-sue")

	// Public profile: everyone sees the public projects; only the owner also sees the private one.
	assertPage(t, s, nil, "/api/v1/projects?owner=act-alice", alicePub)
	assertPage(t, s, &bob, "/api/v1/projects?owner=act-alice", alicePub)
	assertPage(t, s, &alice, "/api/v1/projects?owner=act-alice", alicePub, alicePriv)
	// Members-only profile: hidden from anonymous, open to members.
	assertPage(t, s, nil, "/api/v1/projects?owner=act-mia")
	assertPage(t, s, &bob, "/api/v1/projects?owner=act-mia", miaPub)
	// Private profile: hidden from everyone but its owner (who also sees the private project).
	assertPage(t, s, nil, "/api/v1/projects?owner=act-pat")
	assertPage(t, s, &bob, "/api/v1/projects?owner=act-pat")
	assertPage(t, s, &pat, "/api/v1/projects?owner=act-pat", patPub, patPriv)
	// The unfiltered list still shows pat's public project: hiding the profile hides the filter only.
	if ids, _ := pageIDs(t, s, nil, "/api/v1/projects?limit=50"); !slices.Contains(ids, patPub) {
		t.Fatalf("public list lost %s: %v", patPub, ids)
	}
	// Suspended and unknown accounts look the same: empty, never a 404.
	assertPage(t, s, nil, "/api/v1/projects?owner=act-sue")
	assertPage(t, s, &bob, "/api/v1/projects?owner=nobody-here")
	// Combined with the other filters, and a bad username is 422 from the contract.
	assertPage(t, s, &alice, "/api/v1/projects?owner=act-alice&sort=created&technology=go", alicePub, alicePriv)
	assertProblem(t, s.read(t, nil, "/api/v1/projects?owner=A!"), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
}

func TestProfileActivityQuestionsByAuthorAndAnswerer(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "act-qalice", true, 86)
	bob := s.signIn(t, "act-qbob", true, 87)
	pat := s.signIn(t, "act-qpat", true, 88)
	aliceQ := s.ask(t, &alice, "Question d'Alice").Id.String()
	bobQ := s.ask(t, &bob, "Question de Bob").Id.String()
	patQ := s.ask(t, &pat, "Question de Pat").Id.String()
	s.answer(t, &alice, bobQ, "Alice répond à Bob.")
	s.answer(t, &pat, aliceQ, "Pat répond à Alice.")
	s.answer(t, &pat, bobQ, "Pat répond à Bob.")
	// A removed answer does not make its author an answerer.
	removed := s.answer(t, &alice, patQ, "Réponse retirée.").Id.String()
	if _, err := s.conn.Exec(t.Context(), `UPDATE answers SET status = 'removed', removed_at = now() WHERE id = $1`, removed); err != nil {
		t.Fatal(err)
	}

	assertPage(t, s, nil, "/api/v1/questions?author=act-qalice", aliceQ)
	assertPage(t, s, nil, "/api/v1/questions?answers_by=act-qalice", bobQ)
	assertPage(t, s, nil, "/api/v1/questions?answers_by=act-qpat", aliceQ, bobQ)
	assertPage(t, s, nil, "/api/v1/questions?author=act-qbob&answers_by=act-qpat", bobQ)
	assertPage(t, s, nil, "/api/v1/questions?answers_by=act-qpat&sort=votes", aliceQ, bobQ)

	// Members-only: hidden from anonymous only. Private: hidden from everyone but the owner.
	s.setProfileVisibility(t, "act-qalice", "members")
	assertPage(t, s, nil, "/api/v1/questions?author=act-qalice")
	assertPage(t, s, nil, "/api/v1/questions?answers_by=act-qalice")
	assertPage(t, s, &bob, "/api/v1/questions?author=act-qalice", aliceQ)
	s.setProfileVisibility(t, "act-qpat", "private")
	assertPage(t, s, nil, "/api/v1/questions?answers_by=act-qpat")
	assertPage(t, s, &bob, "/api/v1/questions?answers_by=act-qpat")
	assertPage(t, s, &bob, "/api/v1/questions?author=act-qpat")
	assertPage(t, s, &pat, "/api/v1/questions?answers_by=act-qpat", aliceQ, bobQ)
	assertPage(t, s, &pat, "/api/v1/questions?author=act-qpat", patQ)
	// Suspended and unknown: empty, never a 404; a bad username is 422.
	s.suspend(t, "act-qbob")
	assertPage(t, s, &alice, "/api/v1/questions?author=act-qbob")
	assertPage(t, s, nil, "/api/v1/questions?answers_by=nobody-here")
	assertProblem(t, s.read(t, nil, "/api/v1/questions?author=x"), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
}
