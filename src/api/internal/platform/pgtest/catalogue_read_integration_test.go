//go:build integration

package pgtest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Y4NN777/226os/src/api/internal/discover"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

const (
	catUser     = "019a0000-0000-7000-8000-0000000000a1"
	catProvider = "019a0000-0000-7000-8000-0000000000a2"
	catHidden   = "019a0000-0000-7000-8000-0000000000a3"
	catAlpha    = "019a0000-0000-7000-8000-0000000000b1"
	catBeta     = "019a0000-0000-7000-8000-0000000000b2"
	catRemoved  = "019a0000-0000-7000-8000-0000000000b3"
	catOrphan   = "019a0000-0000-7000-8000-0000000000b4"
	catSecret   = "019a0000-0000-7000-8000-0000000000b5"
	catSrcDocs  = "019a0000-0000-7000-8000-0000000000c1"
	catSrcOther = "019a0000-0000-7000-8000-0000000000c2"
	catSrcClaim = "019a0000-0000-7000-8000-0000000000c3"
	catQuestion = "019a0000-0000-7000-8000-0000000000d1"
	catGoneQ    = "019a0000-0000-7000-8000-0000000000d2"
	catProject  = "019a0000-0000-7000-8000-0000000000d3"
)

func TestCatalogueReadsPublishedRowsOnly(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	conn := migratedDB(ctx, t)
	seedCatalogue(ctx, t, conn)

	codec, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	h := discover.NewHandler(conn, codec, nil)
	alpha := uuid.MustParse(catAlpha)

	rec := callList(ctx, h, apigen.ListResourcesParams{})
	if rec.Code != http.StatusOK {
		t.Fatalf("list %d %s", rec.Code, rec.Body.String())
	}
	var page apigen.ResourcePage
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Items[0].Name != "Alpha Pay" || page.Items[1].Name != "Beta Pay" {
		t.Fatalf("items = %+v", names(page))
	}
	if page.Items[0].VerificationStatus != "verified" || page.Items[0].Provider.Slug != "acme" {
		t.Fatalf("alpha = %+v", page.Items[0])
	}
	if len(page.Items[0].Countries) != 1 || page.Items[0].Countries[0] != "BF" || len(page.Items[0].SdkLanguages) != 1 {
		t.Fatalf("alpha facts %+v sdk %+v", page.Items[0].Countries, page.Items[0].SdkLanguages)
	}
	if len(page.Items[0].Related) != 2 {
		t.Fatalf("related %+v", page.Items[0].Related)
	}

	one := 1
	firstPage := callList(ctx, h, apigen.ListResourcesParams{Limit: &one})
	var first apigen.ResourcePage
	if err := json.Unmarshal(firstPage.Body.Bytes(), &first); err != nil || len(first.Items) != 1 || first.Items[0].Name != "Alpha Pay" || first.NextCursor == nil {
		t.Fatalf("first page %d %s", firstPage.Code, firstPage.Body.String())
	}
	second := callList(ctx, h, apigen.ListResourcesParams{Limit: &one, Cursor: first.NextCursor})
	var rest apigen.ResourcePage
	if err := json.Unmarshal(second.Body.Bytes(), &rest); err != nil || len(rest.Items) != 1 || rest.Items[0].Name != "Beta Pay" || rest.NextCursor != nil {
		t.Fatalf("second page %s", second.Body.String())
	}
	country := "BF"
	filtered := callList(ctx, h, apigen.ListResourcesParams{Country: &country})
	var only apigen.ResourcePage
	if err := json.Unmarshal(filtered.Body.Bytes(), &only); err != nil || len(only.Items) != 1 || only.Items[0].Name != "Alpha Pay" {
		t.Fatalf("country filter %s", filtered.Body.String())
	}

	q := "'; DROP TABLE resources; --"
	injected := callList(ctx, h, apigen.ListResourcesParams{Q: &q})
	if injected.Code != http.StatusOK {
		t.Fatalf("search text must stay a parameter: %d %s", injected.Code, injected.Body.String())
	}
	var still int
	if err := conn.QueryRow(ctx, "SELECT count(*) FROM resources").Scan(&still); err != nil || still < 4 {
		t.Fatalf("resources after the search text: %d %v", still, err)
	}

	got := callGet(ctx, h, alpha)
	if got.Code != http.StatusOK {
		t.Fatalf("get %d %s", got.Code, got.Body.String())
	}
	missing := callGet(ctx, h, uuid.MustParse(catRemoved))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("removed status %d", missing.Code)
	}
	hidden := callGet(ctx, h, uuid.MustParse(catSecret))
	if hidden.Code != http.StatusNotFound {
		t.Fatalf("unpublished provider status %d", hidden.Code)
	}

	src := httptest.NewRecorder()
	h.ListResourceSources(src, httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil), alpha)
	var sources apigen.ResourceSourceList
	if err := json.Unmarshal(src.Body.Bytes(), &sources); err != nil || len(sources.Items) != 3 {
		t.Fatalf("sources %d %s", src.Code, src.Body.String())
	}
	byID := map[string]apigen.ResourceSource{}
	for _, item := range sources.Items {
		byID[item.Id.String()] = item
	}
	if byID[catSrcDocs].Status != "verified" || byID[catSrcDocs].VerifiedAt == nil {
		t.Fatalf("verified source %+v", byID[catSrcDocs])
	}
	if byID[catSrcOther].Status != "unverified" || byID[catSrcOther].VerifiedAt != nil {
		t.Fatalf("other source %+v", byID[catSrcOther])
	}
	if byID[catSrcClaim].Status != "claimed" {
		t.Fatalf("claim %+v", byID[catSrcClaim])
	}

	plan := explainPlan(ctx, t, conn, `
		SELECT r.id FROM resources r
		JOIN providers p ON p.id = r.provider_id AND p.status = 'published' AND p.removed_at IS NULL
		WHERE r.status = 'published'
		ORDER BY r.name, r.id
		LIMIT 21`)
	t.Logf("catalogue list plan:\n%s", plan)
	if !strings.Contains(plan, "Limit") {
		t.Fatalf("plan has no Limit:\n%s", plan)
	}

	// The q and sdk filter paths: EXPLAIN the exact statement and arguments the repository sends.
	// The dataset here is tiny, so the planner is free to scan; the plans are logged as evidence of the
	// shape (bounded by Limit, filters as index or semi-join conditions), not as a scale claim.
	recorder := &recordingDB{db: conn}
	probe := discover.NewHandler(recorder, codec, nil)
	sdk := "go"
	text := "paiement"
	for _, path := range []struct {
		name   string
		params apigen.ListResourcesParams
	}{
		{"q filter", apigen.ListResourcesParams{Q: &text}},
		{"sdk filter", apigen.ListResourcesParams{Sdk: &sdk}},
	} {
		recorder.reset()
		if rec := callList(ctx, probe, path.params); rec.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", path.name, rec.Code, rec.Body.String())
		}
		stmt := recorder.first(t)
		filterPlan := explainPlan(ctx, t, conn, stmt.sql, stmt.args...)
		t.Logf("catalogue list plan, %s:\n%s", path.name, filterPlan)
		if !strings.Contains(filterPlan, "Limit") {
			t.Fatalf("%s plan has no Limit:\n%s", path.name, filterPlan)
		}
	}
}

func callList(ctx context.Context, h interface {
	ListResources(http.ResponseWriter, *http.Request, apigen.ListResourcesParams)
}, params apigen.ListResourcesParams) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ListResources(rec, httptest.NewRequestWithContext(ctx, http.MethodGet, "/api/v1/resources", nil), params)
	return rec
}

func callGet(ctx context.Context, h interface {
	GetResource(http.ResponseWriter, *http.Request, apigen.ResourceId)
}, id apigen.ResourceId) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.GetResource(rec, httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil), id)
	return rec
}

func names(page apigen.ResourcePage) []string {
	out := make([]string, len(page.Items))
	for i, item := range page.Items {
		out[i] = item.Name
	}
	return out
}

func seedCatalogue(ctx context.Context, t *testing.T, conn *pgx.Conn) {
	t.Helper()
	stmts := []string{
		`INSERT INTO users (id, email) VALUES ('` + catUser + `', 'catalogue-read@example.com')`,
		`INSERT INTO providers (id, slug, name, website_url, status, removed_at, created_by) VALUES
		  ('` + catProvider + `', 'acme', 'Acme', 'https://acme.example', 'published', NULL, '` + catUser + `'),
		  ('` + catHidden + `', 'hidden', 'Hidden', NULL, 'removed', now(), '` + catUser + `')`,
		`INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, doc_languages, docs_url, created_by) VALUES
		  ('` + catAlpha + `', '` + catProvider + `', 'mobile-money', 'api', 'alpha-pay', 'Alpha Pay', 'Paiement mobile au Burkina', 'fr', '{fr}', 'https://docs.example', '` + catUser + `'),
		  ('` + catBeta + `', '` + catProvider + `', 'sms-ussd', 'tool', 'beta-pay', 'Beta Pay', 'SMS au Ghana', 'fr', '{fr}', NULL, '` + catUser + `'),
		  ('` + catRemoved + `', '` + catProvider + `', 'mobile-money', 'api', 'removed-pay', 'Removed Pay', 'Retiré', 'fr', '{fr}', NULL, '` + catUser + `'),
		  ('` + catOrphan + `', NULL, 'mobile-money', 'api', 'orphan-pay', 'Orphan Pay', 'Sans fournisseur', 'fr', '{fr}', NULL, '` + catUser + `'),
		  ('` + catSecret + `', '` + catHidden + `', 'mobile-money', 'api', 'secret-pay', 'Secret Pay', 'Fournisseur retiré', 'fr', '{fr}', NULL, '` + catUser + `')`,
		`UPDATE resources SET status = 'removed', removed_at = now() WHERE id = '` + catRemoved + `'`,
		`INSERT INTO resource_sources (id, resource_id, kind, title, url, checked_at) VALUES
		  ('` + catSrcDocs + `', '` + catAlpha + `', 'official_docs', 'Docs', 'https://docs.example/ref', now()),
		  ('` + catSrcOther + `', '` + catAlpha + `', 'community', 'Forum', 'https://forum.example/alpha', now())`,
		`INSERT INTO resource_sources (id, resource_id, kind, title, url) VALUES
		  ('` + catSrcClaim + `', '` + catAlpha + `', 'claim', 'Claim', 'https://claim.example/alpha')`,
		`UPDATE resources SET verification_status = 'verified', verified_at = now(), verified_source_id = '` + catSrcDocs + `' WHERE id = '` + catAlpha + `'`,
		`INSERT INTO resource_countries (resource_id, country_code) VALUES ('` + catAlpha + `', 'BF'), ('` + catBeta + `', 'GH')`,
		`INSERT INTO resource_sdk_languages (resource_id, language, name, package_url) VALUES
		  ('` + catAlpha + `', 'go', 'Go SDK', 'https://pkg.example/go'),
		  ('` + catAlpha + `', 'js', 'JS SDK', NULL)`,
		`INSERT INTO questions (id, author_id, title, body_md, language) VALUES
		  ('` + catQuestion + `', '` + catUser + `', 'Comment encaisser', 'Avec le SDK.', 'fr'),
		  ('` + catGoneQ + `', '` + catUser + `', 'Retirée', 'Non.', 'fr')`,
		`UPDATE questions SET status = 'removed', removed_at = now() WHERE id = '` + catGoneQ + `'`,
		`INSERT INTO projects (id, owner_id, name, summary, language) VALUES
		  ('` + catProject + `', '` + catUser + `', 'Caisse', 'Un projet.', 'fr')`,
		`INSERT INTO content_relations (source_type, src_resource_id, target_type, tgt_project_id, created_by) VALUES
		  ('resource', '` + catAlpha + `', 'project', '` + catProject + `', '` + catUser + `')`,
		`INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_resource_id, created_by) VALUES
		  ('question', '` + catQuestion + `', 'resource', '` + catAlpha + `', '` + catUser + `'),
		  ('question', '` + catGoneQ + `', 'resource', '` + catAlpha + `', '` + catUser + `')`,
	}
	for _, stmt := range stmts {
		if _, err := conn.Exec(ctx, stmt); err != nil {
			t.Fatalf("seed: %v\n%s", err, stmt)
		}
	}
}
