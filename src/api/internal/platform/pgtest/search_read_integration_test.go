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
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/search"
)

// All text below is synthetic. Ids are deterministic: sid(n).
func sid(n int) string { return fmt.Sprintf("019b0000-0000-7000-8000-%012x", n) }

const (
	sUser     = 1
	sProvider = 2
	sGoneProv = 3
)

type searchEnv struct {
	conn     *pgx.Conn
	recorder *recordingDB
	h        interface {
		Search(http.ResponseWriter, *http.Request, apigen.SearchParams)
	}
}

func newSearchEnv(ctx context.Context, t *testing.T) *searchEnv {
	t.Helper()
	conn := migratedDB(ctx, t)
	for _, stmt := range []string{
		fmt.Sprintf(`INSERT INTO users (id, email) VALUES ('%s', 'search-read@example.com')`, sid(sUser)),
		fmt.Sprintf(`INSERT INTO providers (id, slug, name, status, removed_at, created_by) VALUES
		  ('%s', 'search-ok', 'Search Provider', 'published', NULL, '%s'),
		  ('%s', 'search-gone', 'Gone Provider', 'removed', now(), '%s')`, sid(sProvider), sid(sUser), sid(sGoneProv), sid(sUser)),
	} {
		if _, err := conn.Exec(ctx, stmt); err != nil {
			t.Fatalf("seed: %v\n%s", err, stmt)
		}
	}
	rec := &recordingDB{db: conn}
	return &searchEnv{conn: conn, recorder: rec, h: search.NewHandler(rec, unlimited{}, nil)}
}

func (e *searchEnv) exec(ctx context.Context, t *testing.T, stmts ...string) {
	t.Helper()
	for _, stmt := range stmts {
		if _, err := e.conn.Exec(ctx, stmt); err != nil {
			t.Fatalf("seed: %v\n%s", err, stmt)
		}
	}
}

func (e *searchEnv) search(ctx context.Context, t *testing.T, params apigen.SearchParams) (apigen.SearchResults, *httptest.ResponseRecorder) {
	t.Helper()
	rec := httptest.NewRecorder()
	e.h.Search(rec, httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/search", nil), params)
	var res apigen.SearchResults
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
			t.Fatalf("decode: %v: %s", err, rec.Body.String())
		}
	}
	return res, rec
}

func only(g apigen.SearchGroup) *[]apigen.SearchGroup { return &[]apigen.SearchGroup{g} }

func ids(r *apigen.SearchGroupResult) []string {
	if r == nil {
		return nil
	}
	out := make([]string, len(r.Items))
	for i, it := range r.Items {
		out[i] = it.Id.String()
	}
	return out
}

func titlesOf(r *apigen.SearchGroupResult) []string {
	if r == nil {
		return nil
	}
	out := make([]string, len(r.Items))
	for i, it := range r.Items {
		out[i] = it.Title
	}
	return out
}

func sameSet(got, want []string) bool {
	g, w := slices.Clone(got), slices.Clone(want)
	slices.Sort(g)
	slices.Sort(w)
	return slices.Equal(g, w)
}

func TestSearchVisibilityMatrix(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	u, p, gp := sid(sUser), sid(sProvider), sid(sGoneProv)
	e.exec(ctx, t,
		// questions: one published, one removed, one whose only match is inside an answer.
		fmt.Sprintf(`INSERT INTO questions (id, author_id, title, body_md, language) VALUES
		  ('%s', '%s', 'Zebra crossing question', 'visible', 'en'),
		  ('%s', '%s', 'Zebra removed question', 'hidden', 'en'),
		  ('%s', '%s', 'Unrelated title', 'nothing here', 'en')`, sid(10), u, sid(11), u, sid(12), u),
		fmt.Sprintf(`UPDATE questions SET status = 'removed', removed_at = now() WHERE id = '%s'`, sid(11)),
		fmt.Sprintf(`INSERT INTO answers (id, question_id, author_id, body_md, language) VALUES
		  ('%s', '%s', '%s', 'zebra only in an answer', 'en')`, sid(13), sid(12), u),
		// resources: published; removed; provider removed; no provider.
		fmt.Sprintf(`INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, created_by) VALUES
		  ('%s', '%s', 'mobile-money', 'api', 'z-ok', 'Zebra API', 'visible', 'en', '%s'),
		  ('%s', '%s', 'mobile-money', 'api', 'z-removed', 'Zebra removed API', 'hidden', 'en', '%s'),
		  ('%s', '%s', 'mobile-money', 'api', 'z-gone-provider', 'Zebra provider-removed API', 'hidden', 'en', '%s'),
		  ('%s', NULL, 'mobile-money', 'api', 'z-orphan', 'Zebra orphan API', 'hidden', 'en', '%s')`,
			sid(20), p, u, sid(21), p, u, sid(22), gp, u, sid(23), u),
		fmt.Sprintf(`UPDATE resources SET status = 'removed', removed_at = now() WHERE id = '%s'`, sid(21)),
		// projects: public; private (even for its owner); removed.
		fmt.Sprintf(`INSERT INTO projects (id, owner_id, name, summary, language, visibility) VALUES
		  ('%s', '%s', 'Zebra Project', 'visible', 'en', 'public'),
		  ('%s', '%s', 'Zebra Private Project', 'hidden', 'en', 'private'),
		  ('%s', '%s', 'Zebra Removed Project', 'hidden', 'en', 'public')`, sid(30), u, sid(31), u, sid(32), u),
		fmt.Sprintf(`UPDATE projects SET status = 'removed', removed_at = now() WHERE id = '%s'`, sid(32)),
		// learn: published; draft; archived; removed.
		fmt.Sprintf(`INSERT INTO learn_articles (id, author_id, kind, title, summary, body_md, language, level, status, published_at) VALUES
		  ('%s', '%s', 'article', 'Zebra Article', 'visible', 'body', 'en', 'beginner', 'published', now()),
		  ('%s', '%s', 'article', 'Zebra Draft', 'hidden', 'body', 'en', 'beginner', 'draft', NULL),
		  ('%s', '%s', 'article', 'Zebra Archived', 'hidden', 'body', 'en', 'beginner', 'archived', now()),
		  ('%s', '%s', 'article', 'Zebra Removed', 'hidden', 'body', 'en', 'beginner', 'published', now())`,
			sid(40), u, sid(41), u, sid(42), u, sid(43), u),
		fmt.Sprintf(`UPDATE learn_articles SET status = 'removed', removed_at = now() WHERE id = '%s'`, sid(43)),
	)

	res, rec := e.search(ctx, t, apigen.SearchParams{Q: "zebra"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d %s", rec.Code, rec.Body.String())
	}
	want := map[string][]string{
		"questions": {sid(10)}, "resources": {sid(20)}, "projects": {sid(30)}, "learn": {sid(40)},
	}
	got := map[string][]string{
		"questions": ids(res.Groups.Questions), "resources": ids(res.Groups.Resources),
		"projects": ids(res.Groups.Projects), "learn": ids(res.Groups.Learn),
	}
	for group, w := range want {
		if !slices.Equal(got[group], w) {
			t.Errorf("%s = %v, want exactly %v (removed, private, draft, archived, provider-less and answer-only rows must not appear)", group, got[group], w)
		}
	}
	// Answers have no group: the body of an answer never surfaces its question.
	for _, hit := range res.Groups.Questions.Items {
		if hit.Id.String() == sid(12) {
			t.Fatal("a question matched only through its answer must not appear")
		}
	}
}

func TestSearchFoldsAccentsBothWays(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	u := sid(sUser)
	e.exec(ctx, t, fmt.Sprintf(`INSERT INTO questions (id, author_id, title, body_md, language) VALUES
	  ('%s', '%s', 'Éléphant et café', 'Le déploiement échoue', 'fr'),
	  ('%s', '%s', 'elephant and cafe', 'deployment fails', 'en')`, sid(50), u, sid(51), u))
	for _, q := range []string{"éléphant", "elephant", "ÉLÉPHANT", "deploiement"} {
		res, rec := e.search(ctx, t, apigen.SearchParams{Q: q, Types: only(apigen.SearchGroupQuestions)})
		if rec.Code != http.StatusOK {
			t.Fatalf("%q: %d", q, rec.Code)
		}
		items := ids(res.Groups.Questions)
		if q == "deploiement" {
			if !slices.Equal(items, []string{sid(50)}) {
				t.Fatalf("%q = %v", q, items)
			}
			continue
		}
		if !sameSet(items, []string{sid(50), sid(51)}) {
			t.Fatalf("%q = %v, want both rows", q, items)
		}
	}
}

func TestSearchOrderIsStableWithAnIDTieBreaker(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	var values []string
	var want []string
	for i := 0; i < 8; i++ {
		id := sid(100 + i)
		want = append(want, id)
		values = append(values, fmt.Sprintf(`('%s', '%s', 'Banane tiebreak', 'identique', 'fr')`, id, sid(sUser)))
	}
	// Insert in reverse so insertion order cannot explain the result.
	slices.Reverse(values)
	e.exec(ctx, t, `INSERT INTO questions (id, author_id, title, body_md, language) VALUES `+strings.Join(values, ","))
	ten := 10
	var first []string
	for run := 0; run < 3; run++ {
		res, _ := e.search(ctx, t, apigen.SearchParams{Q: "banane", Types: only(apigen.SearchGroupQuestions), PerGroup: &ten})
		got := ids(res.Groups.Questions)
		if run == 0 {
			first = got
		}
		if !slices.Equal(got, want) || !slices.Equal(got, first) {
			t.Fatalf("run %d order = %v, want ascending ids %v", run, got, want)
		}
	}
}

func TestSearchPerGroupAndHasMore(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	var values []string
	for i := 0; i < 7; i++ {
		values = append(values, fmt.Sprintf(`('%s', '%s', 'Girafe %d', 'corps', 'fr')`, sid(200+i), sid(sUser), i))
	}
	e.exec(ctx, t, `INSERT INTO questions (id, author_id, title, body_md, language) VALUES `+strings.Join(values, ","))
	for _, tc := range []struct {
		perGroup *int
		items    int
		more     bool
	}{{nil, 5, true}, {intp(3), 3, true}, {intp(6), 6, true}, {intp(7), 7, false}, {intp(20), 7, false}} {
		res, rec := e.search(ctx, t, apigen.SearchParams{Q: "girafe", Types: only(apigen.SearchGroupQuestions), PerGroup: tc.perGroup})
		if rec.Code != http.StatusOK {
			t.Fatalf("%d", rec.Code)
		}
		g := res.Groups.Questions
		if g == nil || len(g.Items) != tc.items || g.HasMore != tc.more {
			t.Fatalf("per_group %v: %+v, want %d items has_more=%v", tc.perGroup, g, tc.items, tc.more)
		}
		if res.Groups.Resources != nil || res.Groups.Projects != nil || res.Groups.Learn != nil {
			t.Fatal("only the requested group may be present")
		}
	}
}

func intp(v int) *int { return &v }

func TestSearchFiltersApplyWhereTheTableHasTheColumn(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	u, p := sid(sUser), sid(sProvider)
	e.exec(ctx, t,
		fmt.Sprintf(`INSERT INTO questions (id, author_id, title, body_md, language, country_code) VALUES
		  ('%s', '%s', 'Lamantin fr BF', 'x', 'fr', 'BF'),
		  ('%s', '%s', 'Lamantin en GH', 'x', 'en', 'GH'),
		  ('%s', '%s', 'Lamantin fr sans pays', 'x', 'fr', NULL)`, sid(300), u, sid(301), u, sid(302), u),
		fmt.Sprintf(`INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, created_by) VALUES
		  ('%s', '%s', 'mobile-money', 'api', 'l-bf', 'Lamantin BF', 'x', 'fr', '%s'),
		  ('%s', '%s', 'mobile-money', 'api', 'l-gh', 'Lamantin GH', 'x', 'en', '%s')`, sid(310), p, u, sid(311), p, u),
		fmt.Sprintf(`INSERT INTO resource_countries (resource_id, country_code) VALUES ('%s', 'BF'), ('%s', 'GH')`, sid(310), sid(311)),
		fmt.Sprintf(`INSERT INTO projects (id, owner_id, name, summary, language, country_code) VALUES
		  ('%s', '%s', 'Lamantin projet BF', 'x', 'fr', 'BF'),
		  ('%s', '%s', 'Lamantin projet GH', 'x', 'en', 'GH')`, sid(320), u, sid(321), u),
		fmt.Sprintf(`INSERT INTO learn_articles (id, author_id, kind, title, body_md, language, level, status, published_at) VALUES
		  ('%s', '%s', 'article', 'Lamantin guide fr', 'x', 'fr', 'beginner', 'published', now()),
		  ('%s', '%s', 'article', 'Lamantin guide en', 'x', 'en', 'beginner', 'published', now())`, sid(330), u, sid(331), u),
	)
	bf, gh := "BF", "GH"
	fr := apigen.LangFilter("fr")
	cases := []struct {
		name    string
		params  apigen.SearchParams
		wantQ   []string
		wantRes []string
		wantPro []string
		wantLrn []string
	}{
		{"no filter", apigen.SearchParams{Q: "lamantin"}, []string{sid(300), sid(301), sid(302)}, []string{sid(310), sid(311)}, []string{sid(320), sid(321)}, []string{sid(330), sid(331)}},
		{"country BF: learn has no country column and is not narrowed", apigen.SearchParams{Q: "lamantin", Country: &bf}, []string{sid(300)}, []string{sid(310)}, []string{sid(320)}, []string{sid(330), sid(331)}},
		{"country GH", apigen.SearchParams{Q: "lamantin", Country: &gh}, []string{sid(301)}, []string{sid(311)}, []string{sid(321)}, []string{sid(330), sid(331)}},
		{"lang fr", apigen.SearchParams{Q: "lamantin", Lang: &fr}, []string{sid(300), sid(302)}, []string{sid(310)}, []string{sid(320)}, []string{sid(330)}},
		{"country BF and lang fr", apigen.SearchParams{Q: "lamantin", Country: &bf, Lang: &fr}, []string{sid(300)}, []string{sid(310)}, []string{sid(320)}, []string{sid(330)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res, rec := e.search(ctx, t, tc.params)
			if rec.Code != http.StatusOK {
				t.Fatalf("%d %s", rec.Code, rec.Body.String())
			}
			for _, g := range []struct {
				name      string
				got, want []string
			}{
				{"questions", ids(res.Groups.Questions), tc.wantQ}, {"resources", ids(res.Groups.Resources), tc.wantRes},
				{"projects", ids(res.Groups.Projects), tc.wantPro}, {"learn", ids(res.Groups.Learn), tc.wantLrn},
			} {
				if !sameSet(g.got, g.want) {
					t.Errorf("%s = %v, want %v", g.name, g.got, g.want)
				}
			}
		})
	}
}

// C4 PROPOSED option 1, ADR-0005 point 6: q is cut to 100 runes at a word boundary before parsing,
// the response echoes the original q. Without the cut the trailing term would be ANDed and match nothing.
func TestSearchTruncatesLongQueriesC4ProposedOption1(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	e.exec(ctx, t, fmt.Sprintf(`INSERT INTO questions (id, author_id, title, body_md, language) VALUES
	  ('%s', '%s', 'Zebra', 'corps', 'en')`, sid(400), sid(sUser)))
	long := strings.Repeat("zebra ", 16) + "qqqqqqqqqq introuvableterme" // 96 + 10 + 1 + 15 runes
	res, rec := e.search(ctx, t, apigen.SearchParams{Q: long, Types: only(apigen.SearchGroupQuestions)})
	if rec.Code != http.StatusOK {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if res.Query != long {
		t.Fatalf("query echo = %q, want the original", res.Query)
	}
	if got := ids(res.Groups.Questions); !slices.Equal(got, []string{sid(400)}) {
		t.Fatalf("hits = %v: the text past 100 runes must not constrain the match", got)
	}
	// Control: the same terms inside 100 runes do constrain the match.
	res, _ = e.search(ctx, t, apigen.SearchParams{Q: "zebra introuvableterme", Types: only(apigen.SearchGroupQuestions)})
	if len(res.Groups.Questions.Items) != 0 {
		t.Fatal("an unknown term inside the first 100 runes must narrow the result")
	}
}

func TestSearchTreatsInputAsData(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	for _, q := range []string{`'; DROP TABLE questions; --`, `" OR 1=1 --`, `& | ! ( ) <-> :*`, `\\`, `ab\`} {
		if _, rec := e.search(ctx, t, apigen.SearchParams{Q: q}); rec.Code != http.StatusOK {
			t.Fatalf("%q: %d %s", q, rec.Code, rec.Body.String())
		}
	}
	// NUL and invalid UTF-8 cannot be stored in PostgreSQL text: a validation error, never a 500.
	for _, q := range []string{"zebra\x00x", "zebra\xffx"} {
		if _, rec := e.search(ctx, t, apigen.SearchParams{Q: q}); rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("%q: %d %s", q, rec.Code, rec.Body.String())
		}
	}
	var n int
	if err := e.conn.QueryRow(ctx, `SELECT count(*) FROM questions`).Scan(&n); err != nil {
		t.Fatalf("questions table must still exist: %v", err)
	}
}

type relevanceCase struct {
	query string
	group apigen.SearchGroup
	top   string
}

// A small FR/EN corpus with synthetic text; each query names the hit that must rank first.
func TestSearchRelevanceFixture(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	u, p := sid(sUser), sid(sProvider)
	e.exec(ctx, t,
		fmt.Sprintf(`INSERT INTO questions (id, author_id, title, body_md, language) VALUES
		  ('%s', '%s', 'Signature webhook invalide', 'La vérification HMAC échoue après un retry', 'fr'),
		  ('%s', '%s', 'Timeout USSD session', 'La session USSD expire avant la confirmation', 'fr'),
		  ('%s', '%s', 'Webhook retry strategy', 'How to retry failed deliveries with backoff', 'en'),
		  ('%s', '%s', 'Mobile money refund', 'Refund a payment through the callback signature', 'en')`,
			sid(500), u, sid(501), u, sid(502), u, sid(503), u),
		fmt.Sprintf(`INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, created_by) VALUES
		  ('%s', '%s', 'mobile-money', 'api', 'r-kalimba', 'Kalimba Pay API', 'Paiement mobile money pour le Burkina', 'fr', '%s'),
		  ('%s', '%s', 'sms-ussd', 'api', 'r-sahel', 'Sahel SMS Gateway', 'Passerelle SMS et USSD pour opérateurs', 'fr', '%s'),
		  ('%s', '%s', 'mobile-money', 'tool', 'r-savanna', 'Savanna Webhook Tester', 'Tool to replay and verify webhook signatures', 'en', '%s')`,
			sid(510), p, u, sid(511), p, u, sid(512), p, u),
		fmt.Sprintf(`INSERT INTO projects (id, owner_id, name, summary, language) VALUES
		  ('%s', '%s', 'Webhook Kit', 'Verify webhook signatures in Go', 'en'),
		  ('%s', '%s', 'Éléphant Pay', 'Bibliothèque de paiement hors ligne', 'fr')`, sid(520), u, sid(521), u),
		fmt.Sprintf(`INSERT INTO learn_articles (id, author_id, kind, title, summary, body_md, language, level, status, published_at) VALUES
		  ('%s', '%s', 'tutorial', 'Idempotence des webhooks', 'Rejouer un webhook sans doublon', 'Corps du tutoriel', 'fr', 'intermediate', 'published', now()),
		  ('%s', '%s', 'article', 'Introduction to mobile money APIs', 'A tutorial on payment callbacks', 'Body', 'en', 'beginner', 'published', now()),
		  ('%s', '%s', 'article', 'Comprendre USSD', NULL, 'Sessions USSD et délais de confirmation', 'fr', 'beginner', 'published', now())`,
			sid(530), u, sid(531), u, sid(532), u),
	)
	cases := []relevanceCase{
		{"signature webhook", apigen.SearchGroupQuestions, "Signature webhook invalide"},
		{"ussd session", apigen.SearchGroupQuestions, "Timeout USSD session"},
		{"retry backoff", apigen.SearchGroupQuestions, "Webhook retry strategy"},
		{`"retry strategy"`, apigen.SearchGroupQuestions, "Webhook retry strategy"},
		{"paiement mobile", apigen.SearchGroupResources, "Kalimba Pay API"},
		{"sms ussd", apigen.SearchGroupResources, "Sahel SMS Gateway"},
		{"replay webhook", apigen.SearchGroupResources, "Savanna Webhook Tester"},
		{"webhook", apigen.SearchGroupProjects, "Webhook Kit"},
		{"elephant", apigen.SearchGroupProjects, "Éléphant Pay"},
		{"webhook doublon", apigen.SearchGroupLearn, "Idempotence des webhooks"},
		{"mobile money", apigen.SearchGroupLearn, "Introduction to mobile money APIs"},
		{"delais ussd", apigen.SearchGroupLearn, "Comprendre USSD"},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			res, rec := e.search(ctx, t, apigen.SearchParams{Q: tc.query, Types: only(tc.group)})
			if rec.Code != http.StatusOK {
				t.Fatalf("%d %s", rec.Code, rec.Body.String())
			}
			var group *apigen.SearchGroupResult
			switch tc.group {
			case apigen.SearchGroupQuestions:
				group = res.Groups.Questions
			case apigen.SearchGroupResources:
				group = res.Groups.Resources
			case apigen.SearchGroupProjects:
				group = res.Groups.Projects
			default:
				group = res.Groups.Learn
			}
			got := titlesOf(group)
			t.Logf("%q (%s) -> %v", tc.query, tc.group, got)
			if len(got) == 0 || got[0] != tc.top {
				t.Fatalf("top hit of %q = %v, want %q first", tc.query, got, tc.top)
			}
		})
	}
}

func TestSearchSnippetsArePlainTextAndBounded(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	body := "# Titre\n\n**gras** <script>alert(1)</script> [lien](https://example.org) `code` " + strings.Repeat("mot ", 300)
	e.exec(ctx, t, fmt.Sprintf(`INSERT INTO questions (id, author_id, title, body_md, language) VALUES
	  ('%s', '%s', 'Marqueur snippet', $body$%s$body$, 'fr')`, sid(600), sid(sUser), body))
	res, _ := e.search(ctx, t, apigen.SearchParams{Q: "marqueur", Types: only(apigen.SearchGroupQuestions)})
	items := res.Groups.Questions.Items
	if len(items) != 1 {
		t.Fatalf("items = %d", len(items))
	}
	snippet := items[0].Snippet
	if n := len([]rune(snippet)); n == 0 || n > 300 || strings.ContainsAny(snippet, "<>*`#") || !strings.HasPrefix(snippet, "Titre gras") || !strings.Contains(snippet, "lien") {
		t.Fatalf("snippet (%d runes) = %q", len([]rune(snippet)), snippet)
	}
}

// bulk inserts enough synthetic rows for the planner to prefer the GIN indexes.
func (e *searchEnv) bulk(ctx context.Context, t *testing.T, n int) {
	t.Helper()
	u, p := sid(sUser), sid(sProvider)
	e.exec(ctx, t,
		fmt.Sprintf(`INSERT INTO questions (author_id, title, body_md, language, country_code)
		  SELECT '%s', 'bruit ' || md5(i::text), 'corps ' || md5((i + 1)::text) || ' ' || md5((i + 2)::text), 'fr', 'BF' FROM generate_series(1, %d) i`, u, n),
		fmt.Sprintf(`INSERT INTO resources (provider_id, category_code, type, slug, name, summary, summary_language, created_by)
		  SELECT '%s', 'mobile-money', 'api', 'bulk-' || i, 'bruit ' || md5(i::text), 'resume ' || md5((i + 1)::text), 'fr', '%s' FROM generate_series(1, %d) i`, p, u, n),
		fmt.Sprintf(`INSERT INTO projects (owner_id, name, summary, language, country_code)
		  SELECT '%s', 'bruit ' || md5(i::text), 'resume ' || md5((i + 1)::text), 'fr', 'BF' FROM generate_series(1, %d) i`, u, n),
		fmt.Sprintf(`INSERT INTO learn_articles (author_id, kind, title, summary, body_md, language, level, status, published_at)
		  SELECT '%s', 'article', 'bruit ' || md5(i::text), 'resume ' || md5((i + 1)::text), 'corps ' || md5((i + 2)::text), 'fr', 'beginner', 'published', now() FROM generate_series(1, %d) i`, u, n),
	)
}

// Evidence that each group's query is served by its partial GIN index on representative volume.
func TestSearchQueriesUseTheGINIndexes(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	u, p := sid(sUser), sid(sProvider)
	e.bulk(ctx, t, 6000)
	e.exec(ctx, t,
		fmt.Sprintf(`INSERT INTO questions (author_id, title, body_md, language) VALUES ('%s', 'Zebra rare', 'rare', 'fr')`, u),
		fmt.Sprintf(`INSERT INTO resources (provider_id, category_code, type, slug, name, summary, summary_language, created_by) VALUES ('%s', 'mobile-money', 'api', 'zebra-rare', 'Zebra rare', 'rare', 'fr', '%s')`, p, u),
		fmt.Sprintf(`INSERT INTO projects (owner_id, name, summary, language) VALUES ('%s', 'Zebra rare', 'rare', 'fr')`, u),
		fmt.Sprintf(`INSERT INTO learn_articles (author_id, kind, title, body_md, language, level, status, published_at) VALUES ('%s', 'article', 'Zebra rare', 'rare', 'fr', 'beginner', 'published', now())`, u),
		`VACUUM ANALYZE`, // also flushes the GIN pending lists, as autovacuum would in production
	)
	wantIndex := map[apigen.SearchGroup]string{
		apigen.SearchGroupQuestions: "questions_search_gin", apigen.SearchGroupResources: "resources_search_gin",
		apigen.SearchGroupProjects: "projects_search_gin", apigen.SearchGroupLearn: "learn_articles_search_gin",
	}
	for _, group := range []apigen.SearchGroup{apigen.SearchGroupQuestions, apigen.SearchGroupResources, apigen.SearchGroupProjects, apigen.SearchGroupLearn} {
		e.recorder.reset()
		res, rec := e.search(ctx, t, apigen.SearchParams{Q: "zebra", Types: only(group)})
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: %d", group, rec.Code)
		}
		if n := countHits(res, group); n != 1 {
			t.Fatalf("%s hits = %d, want the single rare row", group, n)
		}
		stmt := e.recorder.first(t)
		plan := explainPlan(ctx, t, e.conn, stmt.sql, stmt.args...)
		t.Logf("search plan, group %s (6001 rows):\n%s", group, plan)
		if !strings.Contains(plan, wantIndex[group]) {
			t.Errorf("%s plan does not use %s:\n%s", group, wantIndex[group], plan)
		}
		if !strings.Contains(plan, "Limit") {
			t.Errorf("%s plan has no Limit", group)
		}
	}
	// Worst cases (skeptic review): a term every bulk row carries ("bruit") and a negation-only
	// query ("-zebra") make ts_rank score nearly every published row before the top-N sort. Only
	// LIMIT and the per-query timeout bound them; the plans and timings are logged as evidence.
	for _, q := range []string{"bruit", "-zebra"} {
		for _, group := range []apigen.SearchGroup{apigen.SearchGroupQuestions, apigen.SearchGroupResources, apigen.SearchGroupProjects, apigen.SearchGroupLearn} {
			e.recorder.reset()
			if _, rec := e.search(ctx, t, apigen.SearchParams{Q: q, Types: only(group)}); rec.Code != http.StatusOK {
				t.Fatalf("%q %s: %d", q, group, rec.Code)
			}
			stmt := e.recorder.first(t)
			plan := explainPlan(ctx, t, e.conn, stmt.sql, stmt.args...)
			t.Logf("worst-case search plan, q=%q, group %s (6001 rows):\n%s", q, group, plan)
			if !strings.Contains(plan, "Limit") {
				t.Errorf("%q %s plan has no Limit", q, group)
			}
		}
	}
}

func countHits(res apigen.SearchResults, g apigen.SearchGroup) int {
	switch g {
	case apigen.SearchGroupQuestions:
		return len(res.Groups.Questions.Items)
	case apigen.SearchGroupResources:
		return len(res.Groups.Resources.Items)
	case apigen.SearchGroupProjects:
		return len(res.Groups.Projects.Items)
	}
	return len(res.Groups.Learn.Items)
}
