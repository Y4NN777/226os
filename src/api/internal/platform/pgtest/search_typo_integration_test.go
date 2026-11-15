//go:build integration

package pgtest

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

// SPEC 6 typo tolerance (issue #248 item 2): when full text finds nothing in a group, a title or name that
// holds a word close to the query is found, with the same visibility rules and filters as full text.
func TestSearchToleratesTypos(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	u, p, gp := sid(sUser), sid(sProvider), sid(sGoneProv)
	e.exec(ctx, t,
		fmt.Sprintf(`INSERT INTO questions (id, author_id, title, body_md, language, country_code) VALUES
		  ('%s', '%s', 'Paiement mobile en Côte d''Ivoire', 'x', 'fr', 'CI'),
		  ('%s', '%s', 'Paiement mobile supprimé', 'x', 'fr', 'CI')`, sid(500), u, sid(501), u),
		fmt.Sprintf(`UPDATE questions SET status = 'removed', removed_at = now() WHERE id = '%s'`, sid(501)),
		fmt.Sprintf(`INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, created_by) VALUES
		  ('%s', '%s', 'mobile-money', 'api', 'pm-ok', 'API Paiement mobile', 'x', 'fr', '%s'),
		  ('%s', '%s', 'mobile-money', 'api', 'pm-gone', 'API Paiement mobile retirée', 'x', 'fr', '%s')`, sid(510), p, u, sid(511), gp, u),
		fmt.Sprintf(`INSERT INTO projects (id, owner_id, name, summary, language, visibility, country_code) VALUES
		  ('%s', '%s', 'Paiement mobile ouvert', 'x', 'fr', 'public', 'CI'),
		  ('%s', '%s', 'Paiement mobile privé', 'x', 'fr', 'private', 'CI')`, sid(520), u, sid(521), u),
		fmt.Sprintf(`INSERT INTO learn_articles (id, author_id, kind, title, body_md, language, level, status, published_at) VALUES
		  ('%s', '%s', 'article', 'Comprendre le paiement mobile', 'x', 'fr', 'beginner', 'published', now()),
		  ('%s', '%s', 'article', 'Brouillon paiement mobile', 'x', 'fr', 'beginner', 'draft', NULL)`, sid(530), u, sid(531), u),
	)
	want := map[string][]string{"questions": {sid(500)}, "resources": {sid(510)}, "projects": {sid(520)}, "learn": {sid(530)}}
	// A missing letter ("paiemnt", word similarity 0.63) and an extra one ("paiemment"): full text finds
	// neither word. Swapped letters ("paeiment", 0.39) stay below the 0.6 threshold: a known limit.
	for _, q := range []string{"paiemnt", "paiemment", "paiemnt mobile"} {
		res, rec := e.search(ctx, t, apigen.SearchParams{Q: q})
		if rec.Code != http.StatusOK {
			t.Fatalf("%q: %d %s", q, rec.Code, rec.Body.String())
		}
		got := map[string][]string{
			"questions": ids(res.Groups.Questions), "resources": ids(res.Groups.Resources),
			"projects": ids(res.Groups.Projects), "learn": ids(res.Groups.Learn),
		}
		for group, w := range want {
			if !slices.Equal(got[group], w) {
				t.Errorf("%q %s = %v, want exactly %v (removed, private, draft and provider-removed rows never appear)", q, group, got[group], w)
			}
		}
	}

	// Filters apply to the fallback as to full text: CI matches, GH does not; learn has no country.
	ci, gh := "CI", "GH"
	res, _ := e.search(ctx, t, apigen.SearchParams{Q: "paiemnt", Country: &ci})
	if len(res.Groups.Questions.Items) != 1 || len(res.Groups.Projects.Items) != 1 {
		t.Errorf("country CI: questions %v projects %v", ids(res.Groups.Questions), ids(res.Groups.Projects))
	}
	res, _ = e.search(ctx, t, apigen.SearchParams{Q: "paiemnt", Country: &gh})
	if len(res.Groups.Questions.Items) != 0 || len(res.Groups.Projects.Items) != 0 || len(res.Groups.Learn.Items) != 1 {
		t.Errorf("country GH: questions %v projects %v learn %v", ids(res.Groups.Questions), ids(res.Groups.Projects), ids(res.Groups.Learn))
	}

	// Far from every word, a negation only, a negated word (no fallback: pg_trgm ignores "-") and 2 characters: nothing.
	for _, q := range []string{"zzzqqq", "-paiemnt", "paiemnt -mobile", "pq"} {
		res, rec := e.search(ctx, t, apigen.SearchParams{Q: q})
		if rec.Code != http.StatusOK {
			t.Fatalf("%q: %d", q, rec.Code)
		}
		for _, g := range []*apigen.SearchGroupResult{res.Groups.Questions, res.Groups.Resources, res.Groups.Projects, res.Groups.Learn} {
			if len(g.Items) != 0 {
				t.Errorf("%q matched %v", q, titlesOf(g))
			}
		}
	}
}

// Evidence that each group's fallback is served by its partial trigram GIN index on representative
// volume, and ranks a bounded candidate set.
func TestSearchTypoFallbackUsesTheTrigramIndexes(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 120*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	u, p := sid(sUser), sid(sProvider)
	e.bulk(ctx, t, 6000)
	e.exec(ctx, t,
		fmt.Sprintf(`INSERT INTO questions (author_id, title, body_md, language) VALUES ('%s', 'Lamantin rare', 'rare', 'fr')`, u),
		fmt.Sprintf(`INSERT INTO resources (provider_id, category_code, type, slug, name, summary, summary_language, created_by) VALUES ('%s', 'mobile-money', 'api', 'lamantin-rare', 'Lamantin rare', 'rare', 'fr', '%s')`, p, u),
		fmt.Sprintf(`INSERT INTO projects (owner_id, name, summary, language) VALUES ('%s', 'Lamantin rare', 'rare', 'fr')`, u),
		fmt.Sprintf(`INSERT INTO learn_articles (author_id, kind, title, body_md, language, level, status, published_at) VALUES ('%s', 'article', 'Lamantin rare', 'rare', 'fr', 'beginner', 'published', now())`, u),
		`VACUUM ANALYZE`,
	)
	wantIndex := map[apigen.SearchGroup]string{
		apigen.SearchGroupQuestions: "questions_title_trgm", apigen.SearchGroupResources: "resources_name_trgm",
		apigen.SearchGroupProjects: "projects_name_trgm", apigen.SearchGroupLearn: "learn_articles_title_trgm",
	}
	for _, tc := range []struct {
		q       string
		hits    int
		maxRows int
	}{{"lamantn", 1, 1}, {"bruti", 0, searchCandidates}} {
		for _, group := range []apigen.SearchGroup{apigen.SearchGroupQuestions, apigen.SearchGroupResources, apigen.SearchGroupProjects, apigen.SearchGroupLearn} {
			e.recorder.reset()
			res, rec := e.search(ctx, t, apigen.SearchParams{Q: tc.q, Types: only(group)})
			if rec.Code != http.StatusOK {
				t.Fatalf("%q %s: %d", tc.q, group, rec.Code)
			}
			if tc.hits > 0 && countHits(res, group) != tc.hits {
				t.Fatalf("%q %s hits = %d, want %d", tc.q, group, countHits(res, group), tc.hits)
			}
			stmt := e.recorder.last(t)
			if !strings.Contains(stmt.sql, "word_similarity") {
				t.Fatalf("%q %s: the last statement is not the fallback:\n%s", tc.q, group, stmt.sql)
			}
			plan := explainPlan(ctx, t, e.conn, stmt.sql, stmt.args...)
			t.Logf("typo fallback plan, q=%q, group %s (6001 rows):\n%s", tc.q, group, plan)
			if !strings.Contains(plan, wantIndex[group]) {
				t.Errorf("%q %s plan does not use %s:\n%s", tc.q, group, wantIndex[group], plan)
			}
			if n := maxScanRows(plan); n > tc.maxRows {
				t.Errorf("%q %s: a scan returned %d rows, want at most %d:\n%s", tc.q, group, n, tc.maxRows, plan)
			}
		}
	}
}
