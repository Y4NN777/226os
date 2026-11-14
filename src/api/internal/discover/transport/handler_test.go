package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/discover/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

type fakeReader struct {
	list    func(domain.Query) ([]domain.Resource, error)
	get     func(string) (domain.Resource, error)
	sources func(string) ([]domain.Source, error)
}

func (f fakeReader) List(_ context.Context, q domain.Query) ([]domain.Resource, error) {
	return f.list(q)
}
func (f fakeReader) Get(_ context.Context, id string) (domain.Resource, error) { return f.get(id) }
func (f fakeReader) Sources(_ context.Context, id string) ([]domain.Source, error) {
	return f.sources(id)
}

func handler(t *testing.T, reader fakeReader) *Handler {
	t.Helper()
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return New(service.New(reader, c), nil, nil)
}

func TestListResourcesEncodesAPage(t *testing.T) {
	when := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	docs := "https://docs.example"
	h := handler(t, fakeReader{list: func(q domain.Query) ([]domain.Resource, error) {
		if q.Limit != 2 {
			t.Fatalf("limit %d", q.Limit)
		}
		return []domain.Resource{{
			ID: "00000000-0000-7000-8000-0000000000a1", Slug: "alpha", Name: "Alpha", Type: "api",
			Category: "mobile-money", Summary: "Paiement", DocLanguages: []string{"fr"}, DocsURL: &docs,
			Provider: domain.Provider{Name: "Acme", Slug: "acme"}, Countries: []string{"BF"},
			SDKs:        []domain.SDK{{Language: "go", Name: "Go", URL: "https://pkg.example"}, {Language: "js", URL: "https://js.example"}},
			Related:     []domain.Link{{ID: "00000000-0000-7000-8000-0000000000c1", Type: "question", Title: "How", Lang: "fr"}, {ID: "00000000-0000-7000-8000-0000000000c2", Type: "project", Title: "Proj"}},
			SourceCount: 1, VerificationStatus: "verified", Provenance: "demo", LastVerifiedAt: &when, UpdatedAt: when,
		}, {Name: "Beta", Provenance: "live"}}, nil
	}})
	rec := httptest.NewRecorder()
	one := 1
	kind := apigen.ResourceType("api")
	doc := apigen.Lang("fr")
	blank := ""
	h.ListResources(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/resources", nil), apigen.ListResourcesParams{Limit: &one, Type: &kind, Doclang: &doc, Cursor: &blank})
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status %d type %s", rec.Code, rec.Header().Get("Content-Type"))
	}
	var page apigen.ResourcePage
	if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.NextCursor == nil || page.Items[0].VerificationStatus != "verified" {
		t.Fatalf("%+v", page)
	}
	if page.Items[0].Provenance != apigen.ResourceProvenance("demo") {
		t.Fatalf("provenance %q, want demo", page.Items[0].Provenance)
	}
	if page.Items[0].Countries[0] != "BF" || page.Items[0].SdkLanguages[0].Url == "" || len(page.Items[0].Related) != 2 {
		t.Fatalf("%+v", page.Items[0])
	}
	if page.Items[0].SdkLanguages[1].Name != nil || page.Items[0].Related[1].Lang != nil {
		t.Fatal("empty sdk name and link language should be omitted")
	}
}

func TestCatalogueErrors(t *testing.T) {
	id := uuid.MustParse("00000000-0000-7000-8000-0000000000a1")
	h := handler(t, fakeReader{
		list:    func(domain.Query) ([]domain.Resource, error) { return nil, errors.New("db down secret") },
		get:     func(string) (domain.Resource, error) { return domain.Resource{}, domain.ErrNotFound },
		sources: func(string) ([]domain.Source, error) { return nil, domain.ErrNotFound },
	})
	rec := httptest.NewRecorder()
	h.ListResources(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/resources", nil), apigen.ListResourcesParams{})
	if rec.Code != http.StatusInternalServerError || contains(rec.Body.String(), "secret") {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	bad := 99
	rec = httptest.NewRecorder()
	h.ListResources(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/resources?limit=99", nil), apigen.ListResourcesParams{Limit: &bad})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("limit status %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.GetResource(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), id)
	if rec.Code != http.StatusNotFound || rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("get %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}
	rec = httptest.NewRecorder()
	h.ListResourceSources(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), id)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("sources %d", rec.Code)
	}
}

func TestListResourceSourcesEncodesStatus(t *testing.T) {
	when := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	h := handler(t, fakeReader{sources: func(string) ([]domain.Source, error) {
		return []domain.Source{{ID: "00000000-0000-7000-8000-0000000000b1", Kind: "official_docs", Title: "Docs", URL: "https://docs.example", Status: "verified", VerifiedAt: &when}, {ID: "00000000-0000-7000-8000-0000000000b2", Kind: "claim", URL: "https://claim.example", Status: "claimed"}}, nil
	}})
	rec := httptest.NewRecorder()
	h.ListResourceSources(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), uuid.MustParse("00000000-0000-7000-8000-0000000000a1"))
	var body apigen.ResourceSourceList
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Items[0].Status != "verified" || body.Items[0].VerifiedAt == nil || body.Items[0].Title == nil {
		t.Fatalf("%+v", body.Items[0])
	}
}

func TestGetResourceAndBadIDs(t *testing.T) {
	when := time.Date(2026, 11, 12, 0, 0, 0, 0, time.UTC)
	h := handler(t, fakeReader{
		get: func(string) (domain.Resource, error) {
			return domain.Resource{ID: "00000000-0000-7000-8000-0000000000a1", Name: "Alpha", Type: "api", Category: "mobile-money", Summary: "Paiement", Provider: domain.Provider{Name: "Acme", Slug: "acme"}, Countries: []string{}, SDKs: []domain.SDK{}, Related: []domain.Link{}, DocLanguages: []string{}, UpdatedAt: when, VerificationStatus: "unverified", Provenance: "demo"}, nil
		},
		list: func(domain.Query) ([]domain.Resource, error) {
			return []domain.Resource{{ID: "bad", Name: "Alpha"}}, nil
		},
		sources: func(string) ([]domain.Source, error) {
			return []domain.Source{{ID: "bad", URL: "https://docs.example", Status: "unverified"}}, nil
		},
	})
	rec := httptest.NewRecorder()
	h.GetResource(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), uuid.MustParse("00000000-0000-7000-8000-0000000000a1"))
	if rec.Code != http.StatusOK {
		t.Fatalf("get %d %s", rec.Code, rec.Body.String())
	}
	var one apigen.Resource
	if err := json.Unmarshal(rec.Body.Bytes(), &one); err != nil || one.Provenance != apigen.ResourceProvenance("demo") {
		t.Fatalf("provenance %q err %v", one.Provenance, err)
	}
	rec = httptest.NewRecorder()
	h.ListResources(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), apigen.ListResourcesParams{})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("bad resource id %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ListResourceSources(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), uuid.MustParse("00000000-0000-7000-8000-0000000000a1"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("bad source id %d", rec.Code)
	}
	badLink := handler(t, fakeReader{get: func(string) (domain.Resource, error) {
		return domain.Resource{ID: "00000000-0000-7000-8000-0000000000a1", Related: []domain.Link{{ID: "bad", Type: "project", Title: "P"}}}, nil
	}})
	rec = httptest.NewRecorder()
	badLink.GetResource(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), uuid.MustParse("00000000-0000-7000-8000-0000000000a1"))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("bad link id %d", rec.Code)
	}
}

func contains(s, part string) bool {
	return len(s) >= len(part) && (s == part || len(part) == 0 || (len(s) > 0 && containsAt(s, part)))
}

func containsAt(s, part string) bool {
	for i := 0; i+len(part) <= len(s); i++ {
		if s[i:i+len(part)] == part {
			return true
		}
	}
	return false
}

func TestListResourcesPassesTheCountryList(t *testing.T) {
	var got domain.Query
	h := handler(t, fakeReader{list: func(q domain.Query) ([]domain.Resource, error) {
		got = q
		return nil, nil
	}})
	countries := []apigen.CountryCode{"CI", "BF"}
	rec := httptest.NewRecorder()
	h.ListResources(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), apigen.ListResourcesParams{Country: &countries})
	if rec.Code != http.StatusOK || len(got.Countries) != 2 || got.Countries[0] != "BF" || got.Countries[1] != "CI" {
		t.Fatalf("status %d countries %v", rec.Code, got.Countries)
	}
	rec = httptest.NewRecorder()
	h.ListResources(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil), apigen.ListResourcesParams{})
	if rec.Code != http.StatusOK || got.Countries != nil {
		t.Fatalf("status %d countries %v, want none", rec.Code, got.Countries)
	}
}
