//go:build integration

package pgtest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Y4NN777/226os/src/api/internal/learn"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

// US-LEARN-001 on the production router over a migrated database: public reads, member publishing,
// draft and removed visibility, links, idempotency, and the original text kept byte for byte.

// learnBody is French text with accents, Markdown, raw HTML, a trailing space and mixed newlines: it
// must come back exactly as sent (no trimming, no translation, no normalisation).
const learnBody = "# Intégrer Mobile Money\r\n\nÉtape 1 : créer un compte <b>marchand</b>.  \n\n```go\nfmt.Println(\"ça marche\")\n```\n"

func articleJSON(status, topic string, links ...string) string {
	b := map[string]any{
		"type": "tutorial", "title": "Intégrer Mobile Money au Burkina", "summary": "Un tutoriel court.",
		"body_md": learnBody, "lang": "fr", "level": "beginner", "tags": []string{"mobile-money", "paiement"},
		"status": status,
	}
	if topic != "" {
		b["topic"] = topic
	}
	if len(links) > 0 {
		var ls []map[string]string
		for i := 0; i+1 < len(links); i += 2 {
			ls = append(ls, map[string]string{"type": links[i], "id": links[i+1]})
		}
		b["links"] = ls
	}
	out, _ := json.Marshal(b)
	return string(out)
}

func (s *s2Stack) publish(t *testing.T, m *member, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	h := map[string]string{}
	for k, v := range headers {
		h[k] = v
	}
	if m == nil {
		cookie, token := s.preSession(t)
		h["X-CSRF-Token"] = token
		return s.do(t, http.MethodPost, "/api/v1/learn/articles", body, h, cookie)
	}
	if _, ok := h["X-CSRF-Token"]; !ok {
		h["X-CSRF-Token"] = m.csrf
	}
	return s.do(t, http.MethodPost, "/api/v1/learn/articles", body, h, m.session)
}

func (s *s2Stack) articleCount(t *testing.T) int {
	t.Helper()
	var n int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM learn_articles`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func decodeAs[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %d %s: %v", rec.Code, rec.Body.String(), err)
	}
	return v
}

func TestLearnPublishingAuthorization(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	unverified := s.signIn(t, "learn-unverified", false, 41)

	// Anonymous: 401 for a published article and for a draft; nothing created.
	assertProblem(t, s.publish(t, nil, articleJSON("published", ""), nil), http.StatusUnauthorized, "UNAUTHENTICATED")
	assertProblem(t, s.publish(t, nil, articleJSON("draft", ""), nil), http.StatusUnauthorized, "UNAUTHENTICATED")
	// Unverified member: publishing is 403, a draft is allowed.
	assertProblem(t, s.publish(t, &unverified, articleJSON("published", ""), nil), http.StatusForbidden, "EMAIL_NOT_VERIFIED")
	if n := s.articleCount(t); n != 0 {
		t.Fatalf("%d articles after denied writes", n)
	}
	// CSRF: a member without the session token is refused before anything is written.
	rec := s.publish(t, &unverified, articleJSON("draft", ""), map[string]string{"X-CSRF-Token": "wrong"})
	if rec.Code != http.StatusForbidden || s.articleCount(t) != 0 {
		t.Fatalf("csrf %d %s", rec.Code, rec.Body.String())
	}
	rec = s.publish(t, &unverified, articleJSON("draft", ""), nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("draft %d %s", rec.Code, rec.Body.String())
	}
	if a := decodeAs[apigen.Article](t, rec); a.Status != "draft" || a.PublishedAt != nil {
		t.Fatalf("draft %+v", a)
	}
}

func TestLearnPublishReadListAndSearchKeepTheOriginalText(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "learn-alice", true, 42)

	rec := s.publish(t, &alice, articleJSON("published", "payments", "question", catQuestion, "resource", catAlpha, "project", catProject), nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("publish %d %s", rec.Code, rec.Body.String())
	}
	created := decodeAs[apigen.Article](t, rec)
	if rec.Header().Get("Location") != "/api/v1/learn/articles/"+created.Id.String() {
		t.Fatalf("location %q", rec.Header().Get("Location"))
	}
	if created.BodyMd != learnBody || created.Lang != "fr" || created.Author.Id.String() != alice.id || created.PublishedAt == nil {
		t.Fatalf("created %+v", created)
	}
	var stored string
	if err := s.conn.QueryRow(ctx, `SELECT body_md FROM learn_articles WHERE id = $1`, created.Id).Scan(&stored); err != nil || stored != learnBody {
		t.Fatalf("stored body %q, %v", stored, err)
	}

	// Anonymous read: same original body, language and resolved links of the three kinds.
	rec = s.do(t, http.MethodGet, "/api/v1/learn/articles/"+created.Id.String(), "", nil)
	got := decodeAs[apigen.Article](t, rec)
	if rec.Code != http.StatusOK || got.BodyMd != learnBody || got.Lang != "fr" || len(got.Tags) != 2 {
		t.Fatalf("read %d %+v", rec.Code, got)
	}
	kinds := map[apigen.RelatedLinkType]string{}
	for _, l := range got.Related {
		kinds[l.Type] = l.Title
	}
	if kinds["question"] != "Comment encaisser" || kinds["resource"] != "Alpha Pay" || kinds["project"] != "Caisse" {
		t.Fatalf("related %+v", got.Related)
	}

	// Listed in Apprendre, filtered by topic code, and found by text.
	for _, path := range []string{"/api/v1/learn/articles", "/api/v1/learn/articles?topic=payments&lang=fr&type=tutorial", "/api/v1/learn/articles?q=marchand"} {
		page := decodeAs[apigen.ArticlePage](t, s.do(t, http.MethodGet, path, "", nil))
		if len(page.Items) != 1 || page.Items[0].Id != created.Id {
			t.Fatalf("%s: %+v", path, page)
		}
	}
	page := decodeAs[apigen.ArticlePage](t, s.do(t, http.MethodGet, "/api/v1/learn/articles?topic=mobile", "", nil))
	if len(page.Items) != 0 {
		t.Fatalf("other topic listed %+v", page)
	}
	if rec := s.do(t, http.MethodGet, "/api/v1/search?q=marchand", "", nil); !strings.Contains(rec.Body.String(), created.Id.String()) {
		t.Fatalf("search %d %s", rec.Code, rec.Body.String())
	}

	// Removed content never leaks: 404 and not listed.
	if _, err := s.conn.Exec(ctx, `UPDATE learn_articles SET status = 'removed', removed_at = now() WHERE id = $1`, created.Id); err != nil {
		t.Fatal(err)
	}
	if rec := s.do(t, http.MethodGet, "/api/v1/learn/articles/"+created.Id.String(), "", nil, alice.session); rec.Code != http.StatusNotFound {
		t.Fatalf("removed read %d", rec.Code)
	}
	if page := decodeAs[apigen.ArticlePage](t, s.do(t, http.MethodGet, "/api/v1/learn/articles", "", nil)); len(page.Items) != 0 {
		t.Fatalf("removed listed %+v", page)
	}
}

func TestLearnDraftIsVisibleOnlyToItsAuthor(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "learn-draft-a", true, 43)
	bob := s.signIn(t, "learn-draft-b", true, 44)
	rec := s.publish(t, &alice, articleJSON("draft", ""), nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("draft %d %s", rec.Code, rec.Body.String())
	}
	path := "/api/v1/learn/articles/" + decodeAs[apigen.Article](t, rec).Id.String()
	if rec := s.do(t, http.MethodGet, path, "", nil, alice.session); rec.Code != http.StatusOK {
		t.Fatalf("author %d", rec.Code)
	}
	unknown := assertProblem(t, s.do(t, http.MethodGet, "/api/v1/learn/articles/019a0000-0000-7000-8000-00000000ffff", "", nil), http.StatusNotFound, "NOT_FOUND")
	for name, rec := range map[string]*httptest.ResponseRecorder{
		"anonymous": s.do(t, http.MethodGet, path, "", nil),
		"bob":       s.do(t, http.MethodGet, path, "", nil, bob.session),
	} {
		p := assertProblem(t, rec, http.StatusNotFound, "NOT_FOUND")
		if p["detail"] != unknown["detail"] || p["title"] != unknown["title"] {
			t.Fatalf("%s: draft 404 %v differs from unknown %v", name, p, unknown)
		}
	}
	if page := decodeAs[apigen.ArticlePage](t, s.do(t, http.MethodGet, "/api/v1/learn/articles", "", nil, alice.session)); len(page.Items) != 0 {
		t.Fatalf("draft listed %+v", page)
	}
}

func TestLearnRejectsUnusableLinksAndTopicsAtomically(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "learn-links", true, 45)
	if _, err := s.conn.Exec(ctx, `INSERT INTO projects (id, owner_id, name, summary, language, visibility) VALUES
		('019a0000-0000-7000-8000-0000000000e9', '`+catUser+`', 'Privé', 'Caché.', 'fr', 'private')`); err != nil {
		t.Fatal(err)
	}
	for name, body := range map[string]string{
		"removed question":     articleJSON("published", "", "question", catGoneQ),
		"removed resource":     articleJSON("published", "", "resource", catRemoved),
		"unpublished provider": articleJSON("published", "", "resource", catSecret),
		"private project":      articleJSON("published", "", "project", "019a0000-0000-7000-8000-0000000000e9"),
		"unknown target":       articleJSON("published", "", "question", "019a0000-0000-7000-8000-00000000fffe"),
		"good then bad":        articleJSON("published", "", "question", catQuestion, "question", catGoneQ),
		"unknown topic":        articleJSON("published", "not-a-topic"),
	} {
		t.Run(name, func(t *testing.T) {
			p := assertProblem(t, s.publish(t, &alice, body, nil), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
			// The violation names the body member as a JSON pointer with a contract code (issue #116 rule).
			v, _ := json.Marshal(p["violations"])
			pointer := strings.Contains(string(v), `"/links"`) || strings.Contains(string(v), `"/topic"`)
			if !strings.Contains(string(v), `"invalid_value"`) || !pointer {
				t.Fatalf("violations %s", v)
			}
		})
	}
	// A series target is not in the contract's LinkRef any more (no content_relations pair has one,
	// owner decision 2026-11-14), so request validation refuses it and names the exact member.
	t.Run("series is not in the contract", func(t *testing.T) {
		p := assertProblem(t, s.publish(t, &alice, articleJSON("published", "", "series", catQuestion), nil), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		v, _ := json.Marshal(p["violations"])
		if string(v) != `[{"code":"invalid_value","field":"/links/0/type"}]` {
			t.Fatalf("violations %s", v)
		}
	})
	var articles, relations, tags int
	if err := s.conn.QueryRow(ctx, `SELECT (SELECT count(*) FROM learn_articles), (SELECT count(*) FROM content_relations WHERE tgt_article_id IS NOT NULL),
		(SELECT count(*) FROM tags WHERE slug = 'paiement')`).Scan(&articles, &relations, &tags); err != nil {
		t.Fatal(err)
	}
	if articles != 0 || relations != 0 || tags != 0 {
		t.Fatalf("rolled back writes left %d articles, %d relations, %d tags", articles, relations, tags)
	}
}

func TestLearnIdempotentCreateReplays(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "learn-idem", true, 46)
	key := map[string]string{"Idempotency-Key": "learn-key-0123456789"}
	first := s.publish(t, &alice, articleJSON("published", ""), key)
	second := s.publish(t, &alice, articleJSON("published", ""), key)
	if first.Code != http.StatusCreated || second.Code != http.StatusCreated || second.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatalf("first %d second %d replayed %q", first.Code, second.Code, second.Header().Get("Idempotency-Replayed"))
	}
	if n := s.articleCount(t); n != 1 {
		t.Fatalf("%d articles after a replay", n)
	}
	assertProblem(t, s.publish(t, &alice, articleJSON("draft", ""), key), http.StatusUnprocessableEntity, "IDEMPOTENCY_KEY_REUSED")
}

func TestLearnSeriesShowsOnlyPublishedLessons(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "learn-series", true, 47)
	var ids []string
	for i := range 3 {
		rec := s.publish(t, &alice, articleJSON("published", ""), nil)
		if rec.Code != http.StatusCreated {
			t.Fatalf("lesson article %d: %d %s", i, rec.Code, rec.Body.String())
		}
		ids = append(ids, decodeAs[apigen.Article](t, rec).Id.String())
	}
	const series, gone = "019a0000-0000-7000-8000-0000000000f1", "019a0000-0000-7000-8000-0000000000f2"
	for _, stmt := range []string{
		`INSERT INTO training_series (id, author_id, origin, title, summary, language, level, topic_code) VALUES
		  ('` + series + `', '` + alice.id + `', 'platform', 'Formation IA', 'Gratuite.', 'fr', 'beginner', 'ai'),
		  ('` + gone + `', '` + alice.id + `', 'platform', 'Retirée', 'Non.', 'fr', 'beginner', 'ai')`,
		`UPDATE training_series SET status = 'removed', removed_at = now() WHERE id = '` + gone + `'`,
		`INSERT INTO training_lessons (series_id, position, title, language, article_id) VALUES
		  ('` + series + `', 2, 'Deux', 'fr', '` + ids[1] + `'),
		  ('` + series + `', 1, 'Un', 'fr', '` + ids[0] + `'),
		  ('` + series + `', 3, 'Trois', 'fr', '` + ids[2] + `')`,
		`UPDATE learn_articles SET status = 'archived' WHERE id = '` + ids[2] + `'`,
	} {
		if _, err := s.conn.Exec(ctx, stmt); err != nil {
			t.Fatalf("%v\n%s", err, stmt)
		}
	}
	page := decodeAs[apigen.SeriesPage](t, s.do(t, http.MethodGet, "/api/v1/learn/series?topic=ai", "", nil))
	if len(page.Items) != 1 || page.Items[0].Id.String() != series || page.Items[0].LessonCount != 2 {
		t.Fatalf("series page %+v", page)
	}
	rec := s.do(t, http.MethodGet, "/api/v1/learn/series/"+series, "", nil, alice.session)
	got := decodeAs[apigen.TrainingSeries](t, rec)
	if rec.Code != http.StatusOK || len(got.Lessons) != 2 || got.Lessons[0].Title != "Un" || got.Lessons[1].ArticleId.String() != ids[1] {
		t.Fatalf("series %d %+v", rec.Code, got)
	}
	if strings.Contains(rec.Body.String(), "my_progress") {
		t.Fatal("my_progress served before C3")
	}
	assertProblem(t, s.do(t, http.MethodGet, "/api/v1/learn/series/"+gone, "", nil), http.StatusNotFound, "NOT_FOUND")
}

// beginRecorder is a recordingDB that can also open a transaction, as the Learn repository needs.
type beginRecorder struct {
	*recordingDB
	pool *pgxpool.Pool
}

func (b beginRecorder) Begin(ctx context.Context) (pgx.Tx, error) { return b.pool.Begin(ctx) }

// The list plans of the exact SQL a request sends are logged as EXPLAIN (ANALYZE, BUFFERS) evidence
// (ADR-0005 point 6); each is bounded by its Limit.
func TestLearnListPlansAreBounded(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	alice := s.signIn(t, "learn-plans", true, 48)
	if rec := s.publish(t, &alice, articleJSON("published", "payments"), nil); rec.Code != http.StatusCreated {
		t.Fatalf("publish %d %s", rec.Code, rec.Body.String())
	}
	codec, err := cursor.New(s2Secret)
	if err != nil {
		t.Fatal(err)
	}
	rec := &recordingDB{db: s.pool}
	h := learn.NewHandler(beginRecorder{rec, s.pool}, codec, nil, nil)
	q, topic := "marchand", "payments"
	for name, call := range map[string]func(http.ResponseWriter, *http.Request){
		"articles q filter": func(w http.ResponseWriter, r *http.Request) {
			h.ListLearnArticles(w, r, apigen.ListLearnArticlesParams{Q: &q})
		},
		"articles topic filter": func(w http.ResponseWriter, r *http.Request) {
			h.ListLearnArticles(w, r, apigen.ListLearnArticlesParams{Topic: &topic})
		},
		"series topic filter": func(w http.ResponseWriter, r *http.Request) {
			h.ListTrainingSeries(w, r, apigen.ListTrainingSeriesParams{Topic: &topic})
		},
	} {
		rec.reset()
		w := httptest.NewRecorder()
		call(w, httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", name, w.Code, w.Body.String())
		}
		stmt := rec.first(t)
		plan := explainPlan(ctx, t, s.conn, stmt.sql, stmt.args...)
		t.Logf("learn list plan, %s:\n%s", name, plan)
		if !strings.Contains(plan, "Limit") {
			t.Fatalf("%s plan has no Limit:\n%s", name, plan)
		}
	}
}

// The S4 content script (scripts/content/free-series.sql) on a migrated database: it refuses without an
// active owner, creates one public series with 4 published lessons attributed to that owner, and a second
// run changes nothing.
func TestFreeSeriesContentScript(t *testing.T) {
	ctx := t.Context()
	s := newS2Stack(ctx, t, nil)
	sqlText, err := os.ReadFile(filepath.Join(repoRoot(t), "scripts", "content", "free-series.sql"))
	if err != nil {
		t.Fatal(err)
	}
	run := func(owner string) error {
		conn, err := s.pool.Acquire(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Release()
		if _, err := conn.Exec(ctx, `SELECT set_config('content.owner_username', $1, false)`, owner); err != nil {
			t.Fatal(err)
		}
		_, err = conn.Exec(ctx, string(sqlText))
		return err
	}
	count := func() (series, articles int) {
		if err := s.conn.QueryRow(ctx, `SELECT (SELECT count(*) FROM training_series), (SELECT count(*) FROM learn_articles)`).Scan(&series, &articles); err != nil {
			t.Fatal(err)
		}
		return series, articles
	}
	for _, owner := range []string{"", "nobody-here"} {
		if err := run(owner); err == nil {
			t.Fatalf("owner %q accepted", owner)
		}
	}
	owner := s.signIn(t, "learn-owner", true, 49)
	if err := run("learn-owner"); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if series, articles := count(); series != 1 || articles != 4 {
		t.Fatalf("%d series, %d articles", series, articles)
	}
	if err := run("learn-owner"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("second run: %v", err)
	}
	if series, articles := count(); series != 1 || articles != 4 {
		t.Fatalf("second run changed data: %d series, %d articles", series, articles)
	}
	// A series loaded by an earlier version of the script, under another title, still blocks a new run.
	if _, err := s.conn.Exec(ctx, `UPDATE training_series SET title = 'Bien démarrer avec l''IA générative'`); err != nil {
		t.Fatal(err)
	}
	if err := run("learn-owner"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("run after a retitled series: %v", err)
	}
	if series, articles := count(); series != 1 || articles != 4 {
		t.Fatalf("run after a retitled series changed data: %d series, %d articles", series, articles)
	}
	var demo int
	if err := s.conn.QueryRow(ctx, `SELECT (SELECT count(*) FROM training_series WHERE provenance <> 'live') + (SELECT count(*) FROM learn_articles WHERE provenance <> 'live')`).Scan(&demo); err != nil || demo != 0 {
		t.Fatalf("non-live rows %d, %v", demo, err)
	}

	page := decodeAs[apigen.SeriesPage](t, s.do(t, http.MethodGet, "/api/v1/learn/series?topic=ai&lang=fr", "", nil))
	if len(page.Items) != 1 || page.Items[0].LessonCount != 4 || page.Items[0].Author.Id.String() != owner.id {
		t.Fatalf("series page %+v", page)
	}
	got := decodeAs[apigen.TrainingSeries](t, s.do(t, http.MethodGet, "/api/v1/learn/series/"+page.Items[0].Id.String(), "", nil))
	if len(got.Lessons) != 4 {
		t.Fatalf("lessons %+v", got.Lessons)
	}
	for i, l := range got.Lessons {
		rec := s.do(t, http.MethodGet, "/api/v1/learn/articles/"+l.ArticleId.String(), "", nil)
		a := decodeAs[apigen.Article](t, rec)
		if l.Position != i+1 || rec.Code != http.StatusOK || a.Type != "tutorial" || a.Lang != "fr" || len(a.Tags) != 4 || len(a.Title) < 5 {
			t.Fatalf("lesson %d: %+v / %d %+v", i, l, rec.Code, a)
		}
	}
}
