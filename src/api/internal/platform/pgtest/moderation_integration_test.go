//go:build integration

package pgtest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	builddomain "github.com/Y4NN777/226os/src/api/internal/build/domain"
	buildrepo "github.com/Y4NN777/226os/src/api/internal/build/repository"
	discoverdomain "github.com/Y4NN777/226os/src/api/internal/discover/domain"
	discoverrepo "github.com/Y4NN777/226os/src/api/internal/discover/repository"
	learndomain "github.com/Y4NN777/226os/src/api/internal/learn/domain"
	learnrepo "github.com/Y4NN777/226os/src/api/internal/learn/repository"
	moderationdomain "github.com/Y4NN777/226os/src/api/internal/moderation/domain"
	moderationrepo "github.com/Y4NN777/226os/src/api/internal/moderation/repository"
	searchdomain "github.com/Y4NN777/226os/src/api/internal/search/domain"
	searchrepo "github.com/Y4NN777/226os/src/api/internal/search/repository"
	solvedomain "github.com/Y4NN777/226os/src/api/internal/solve/domain"
	solverepo "github.com/Y4NN777/226os/src/api/internal/solve/repository"
)

func TestModerationAuthorizationCSRFAndWriteLimit(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	stack := newS2Stack(ctx, t, nil)
	member := stack.signIn(t, "modmember", true, 31)
	moderator := stack.signIn(t, "modrole", true, 32)
	if _, err := stack.conn.Exec(ctx, `UPDATE users SET role = 'moderator' WHERE id = $1`, moderator.id); err != nil {
		t.Fatal(err)
	}
	path := "/api/v1/moderation/resource/" + catAlpha + "/hide"
	restorePath := "/api/v1/moderation/resource/" + catAlpha + "/restore"
	body := `{"reason":"spam"}`

	preCookie, preCSRF := stack.preSession(t)
	assertProblem(t, stack.do(t, http.MethodPost, path, body, map[string]string{"X-CSRF-Token": preCSRF}, preCookie), http.StatusUnauthorized, "UNAUTHENTICATED")
	assertProblem(t, stack.do(t, http.MethodPost, path, body, map[string]string{"X-CSRF-Token": member.csrf}, member.session), http.StatusForbidden, "FORBIDDEN")
	for _, call := range []struct{ path, body string }{
		{path, body},
		{restorePath, body},
		{"/api/v1/moderation/error-reports/" + catAlpha + "/resolve", `{"status":"resolved"}`},
	} {
		assertProblem(t, stack.do(t, http.MethodPost, call.path, call.body, nil, moderator.session), http.StatusForbidden, "CSRF_FAILED")
	}

	if rec := stack.do(t, http.MethodPost, path, body, map[string]string{"X-CSRF-Token": moderator.csrf}, moderator.session); rec.Code != http.StatusNoContent {
		t.Fatalf("moderator hide: %d %s", rec.Code, rec.Body.String())
	}
	for i := 1; i < 60; i++ {
		next := path
		if i%2 == 1 {
			next = restorePath
		}
		if rec := stack.do(t, http.MethodPost, next, body, map[string]string{"X-CSRF-Token": moderator.csrf}, moderator.session); rec.Code != http.StatusNoContent {
			t.Fatalf("moderator call %d: %d %s", i+1, rec.Code, rec.Body.String())
		}
	}
	limited := stack.do(t, http.MethodPost, path, body, map[string]string{"X-CSRF-Token": moderator.csrf}, moderator.session)
	assertProblem(t, limited, http.StatusTooManyRequests, "RATE_LIMITED")
	if limited.Header().Get("Retry-After") == "" {
		t.Fatal("rate-limited moderation response lacks Retry-After")
	}
}

func moderationDB(t *testing.T) (*moderationrepo.Repository, *pgx.Conn) {
	t.Helper()
	conn := migratedDB(t.Context(), t)
	if _, err := conn.Exec(t.Context(), fixtureGraph); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	return moderationrepo.New(conn), conn
}

func TestModerationRepositoryTransitionsEveryTargetAndAudits(t *testing.T) {
	repo, conn := moderationDB(t)
	tests := []struct {
		kind, id, table, hidden, visible string
	}{
		{"question", qAlice, "questions", "removed", "published"},
		{"answer", aBob, "answers", "removed", "published"},
		{"resource", rKal, "resources", "removed", "published"},
		{"project", prAlice, "projects", "removed", "published"},
		{"article", arAlice, "learn_articles", "removed", "published"},
		{"relation", crOne, "content_relations", "hidden", "active"},
	}
	for _, tc := range tests {
		t.Run(tc.kind, func(t *testing.T) {
			if err := repo.Hide(t.Context(), tc.kind, tc.id, uMod, "spam", "req-hide"); err != nil {
				t.Fatal(err)
			}
			assertModerationState(t, conn, tc.table, tc.id, tc.hidden)
			if err := repo.Hide(t.Context(), tc.kind, tc.id, uMod, "spam", "req-repeat"); !errors.Is(err, moderationdomain.ErrStateConflict) {
				t.Fatalf("repeat hide = %v", err)
			}
			if err := repo.Restore(t.Context(), tc.kind, tc.id, uMod, "appel accepte", "req-restore"); err != nil {
				t.Fatal(err)
			}
			assertModerationState(t, conn, tc.table, tc.id, tc.visible)
			if err := repo.Restore(t.Context(), tc.kind, tc.id, uMod, "appel accepte", "req-repeat"); !errors.Is(err, moderationdomain.ErrStateConflict) {
				t.Fatalf("repeat restore = %v", err)
			}
			var audits int
			if err := conn.QueryRow(t.Context(), `SELECT count(*) FROM audit_log WHERE target_type=$1 AND target_id=$2 AND action IN ($1 || '.hide', $1 || '.restore')`, tc.kind, tc.id).Scan(&audits); err != nil || audits != 2 {
				t.Fatalf("audit count = %d, err %v", audits, err)
			}
		})
	}
}

func TestModerationAuditFailureRollsBackContentChange(t *testing.T) {
	repo, conn := moderationDB(t)
	missingActor := idp + "0000000000ff"
	if err := repo.Hide(t.Context(), "question", qBob, missingActor, "spam", "req-rollback"); err == nil {
		t.Fatal("hide with missing audit actor succeeded")
	}
	assertModerationState(t, conn, "questions", qBob, "published")
	var audits int
	if err := conn.QueryRow(t.Context(), `SELECT count(*) FROM audit_log WHERE target_id=$1`, qBob).Scan(&audits); err != nil || audits != 0 {
		t.Fatalf("orphan audits = %d, err %v", audits, err)
	}
}

func TestModerationHiddenContentLeavesReadsAndSearchThenReturns(t *testing.T) {
	repo, conn := moderationDB(t)
	ctx := t.Context()
	solve := solverepo.New(conn)
	discover := discoverrepo.New(conn)
	build := buildrepo.New(conn)
	learn := learnrepo.New(conn)
	search := searchrepo.New(conn)

	tests := []struct {
		kind, id      string
		assertHidden  func(*testing.T)
		assertVisible func(*testing.T)
	}{
		{"question", qAlice, func(t *testing.T) {
			if _, err := solve.GetQuestion(ctx, qAlice); err == nil {
				t.Fatal("hidden question detail visible")
			}
			items, err := solve.ListQuestions(ctx, solvedomain.QuestionQuery{Limit: 20})
			if err != nil || hasQuestion(items, qAlice) {
				t.Fatalf("question list: %v %+v", err, items)
			}
			assertNoSearchHit(t, search, searchdomain.Questions, "signature", qAlice)
		}, func(t *testing.T) {
			if _, err := solve.GetQuestion(ctx, qAlice); err != nil {
				t.Fatal(err)
			}
		}},
		{"resource", rKal, func(t *testing.T) {
			if _, err := discover.Get(ctx, rKal); err == nil {
				t.Fatal("hidden resource detail visible")
			}
			items, err := discover.List(ctx, discoverdomain.Query{Limit: 20})
			if err != nil || hasResource(items, rKal) {
				t.Fatalf("resource list: %v %+v", err, items)
			}
			assertNoSearchHit(t, search, searchdomain.Resources, "Kalimba", rKal)
		}, func(t *testing.T) {
			if _, err := discover.Get(ctx, rKal); err != nil {
				t.Fatal(err)
			}
		}},
		{"project", prAlice, func(t *testing.T) {
			if _, err := build.GetProject(ctx, prAlice, ""); err == nil {
				t.Fatal("hidden project detail visible")
			}
			items, err := build.ListProjects(ctx, builddomain.ProjectQuery{Limit: 20})
			if err != nil || hasProject(items, prAlice) {
				t.Fatalf("project list: %v %+v", err, items)
			}
			assertNoSearchHit(t, search, searchdomain.Projects, "webhook", prAlice)
		}, func(t *testing.T) {
			if _, err := build.GetProject(ctx, prAlice, ""); err != nil {
				t.Fatal(err)
			}
		}},
		{"article", arAlice, func(t *testing.T) {
			if _, err := learn.GetArticle(ctx, arAlice, ""); err == nil {
				t.Fatal("hidden article detail visible")
			}
			items, err := learn.ListArticles(ctx, learndomain.ArticleQuery{Limit: 20})
			if err != nil || hasArticle(items, arAlice) {
				t.Fatalf("article list: %v %+v", err, items)
			}
			assertNoSearchHit(t, search, searchdomain.Learn, "Idempotence", arAlice)
		}, func(t *testing.T) {
			if _, err := learn.GetArticle(ctx, arAlice, ""); err != nil {
				t.Fatal(err)
			}
		}},
	}
	for _, tc := range tests {
		t.Run(tc.kind, func(t *testing.T) {
			if err := repo.Hide(ctx, tc.kind, tc.id, uMod, "spam", "req-hide"); err != nil {
				t.Fatal(err)
			}
			tc.assertHidden(t)
			if err := repo.Restore(ctx, tc.kind, tc.id, uMod, "appel", "req-restore"); err != nil {
				t.Fatal(err)
			}
			tc.assertVisible(t)
		})
	}

	if err := repo.Hide(ctx, "answer", aBob, uMod, "spam", "req-answer"); err != nil {
		t.Fatal(err)
	}
	answers, err := solve.ListAnswers(ctx, solvedomain.AnswerQuery{QuestionID: qAlice, Limit: 20})
	if err != nil || hasAnswer(answers, aBob) {
		t.Fatalf("answer list: %v %+v", err, answers)
	}
	if err := repo.Restore(ctx, "answer", aBob, uMod, "appel", "req-answer-restore"); err != nil {
		t.Fatal(err)
	}

	before, err := solve.GetQuestion(ctx, qAlice)
	if err != nil || len(before.Related) == 0 {
		t.Fatalf("related before: %v %+v", err, before.Related)
	}
	if err := repo.Hide(ctx, "relation", crOne, uMod, "spam", "req-relation"); err != nil {
		t.Fatal(err)
	}
	after, err := solve.GetQuestion(ctx, qAlice)
	if err != nil || len(after.Related) != 0 {
		t.Fatalf("related hidden: %v %+v", err, after.Related)
	}
}

func TestModerationResolveReportIsAuditedAndTerminal(t *testing.T) {
	repo, conn := moderationDB(t)
	got, err := repo.ResolveReport(t.Context(), erKal, uMod, "resolved", "corrige", "req-report")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != erKal || got.ResourceID != rKal || got.Kind != "broken_link" || got.Status != "resolved" {
		t.Fatalf("report = %+v", got)
	}
	if _, err := repo.ResolveReport(t.Context(), erKal, uMod, "rejected", "non", "req-repeat"); !errors.Is(err, moderationdomain.ErrStateConflict) {
		t.Fatalf("repeat resolve = %v", err)
	}
	var actor, action, reason string
	if err := conn.QueryRow(t.Context(), `SELECT actor_id::text, action, reason FROM audit_log WHERE target_type='report' AND target_id=$1 AND request_id='req-report'`, erKal).Scan(&actor, &action, &reason); err != nil || actor != uMod || action != "error_report.resolve" || reason != "corrige" {
		t.Fatalf("audit = %q %q %q, err %v", actor, action, reason, err)
	}
}

func assertModerationState(t *testing.T, conn *pgx.Conn, table, id, want string) {
	t.Helper()
	var got string
	if err := conn.QueryRow(t.Context(), `SELECT status FROM `+table+` WHERE id=$1`, id).Scan(&got); err != nil || got != want {
		t.Fatalf("%s status = %q, want %q, err %v", table, got, want, err)
	}
}

func assertNoSearchHit(t *testing.T, repo *searchrepo.Repository, group searchdomain.Group, q, id string) {
	t.Helper()
	hits, err := repo.Search(t.Context(), group, searchdomain.Query{Text: q, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	for _, hit := range hits {
		if hit.ID == id {
			t.Fatalf("hidden %s remained in search", id)
		}
	}
}

func hasQuestion(items []solvedomain.Question, id string) bool {
	for _, v := range items {
		if v.ID == id {
			return true
		}
	}
	return false
}
func hasAnswer(items []solvedomain.Answer, id string) bool {
	for _, v := range items {
		if v.ID == id {
			return true
		}
	}
	return false
}
func hasResource(items []discoverdomain.Resource, id string) bool {
	for _, v := range items {
		if v.ID == id {
			return true
		}
	}
	return false
}
func hasProject(items []builddomain.Project, id string) bool {
	for _, v := range items {
		if v.ID == id {
			return true
		}
	}
	return false
}
func hasArticle(items []learndomain.ArticleSummary, id string) bool {
	for _, v := range items {
		if v.ID == id {
			return true
		}
	}
	return false
}
