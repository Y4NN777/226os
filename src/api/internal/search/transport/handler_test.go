package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/search/domain"
	"github.com/Y4NN777/226os/src/api/internal/search/service"
)

const hitID = "00000000-0000-7000-8000-0000000000a1"

type fakeReader struct {
	search func(domain.Group, domain.Query) ([]domain.Hit, error)
}

func (f fakeReader) Search(_ context.Context, g domain.Group, q domain.Query) ([]domain.Hit, error) {
	return f.search(g, q)
}

func handler(r fakeReader) *Handler { return New(service.New(r), nil) }

func call(t *testing.T, h *Handler, params apigen.SearchParams) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	h.Search(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/search", nil), params)
	return rec
}

func TestSearchEncodesGroupedHits(t *testing.T) {
	var limits []int
	h := handler(fakeReader{search: func(g domain.Group, q domain.Query) ([]domain.Hit, error) {
		limits = append(limits, q.Limit)
		switch g {
		case domain.Questions:
			return []domain.Hit{
				{Type: g, ID: hitID, Title: "Webhook", Snippet: "extrait", Lang: "fr"},
				{Type: g, ID: hitID, Title: "Autre", Snippet: "x", Lang: "en"},
			}, nil
		case domain.Learn:
			return []domain.Hit{{Type: g, ID: hitID, Title: "Sans langue", Snippet: "y"}}, nil
		}
		return nil, nil
	}})
	two := 1
	rec := call(t, h, apigen.SearchParams{Q: "webhook", PerGroup: &two})
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status %d type %s body %s", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
	}
	var res apigen.SearchResults
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Query != "webhook" {
		t.Fatalf("query = %q", res.Query)
	}
	q := res.Groups.Questions
	if q == nil || len(q.Items) != 1 || !q.HasMore || q.Items[0].Type != apigen.SearchGroupQuestions ||
		q.Items[0].Id.String() != hitID || q.Items[0].Title != "Webhook" || q.Items[0].Snippet != "extrait" ||
		q.Items[0].Lang == nil || *q.Items[0].Lang != "fr" {
		t.Fatalf("questions = %+v", q)
	}
	if res.Groups.Learn == nil || res.Groups.Learn.Items[0].Lang != nil {
		t.Fatalf("a hit without language must omit lang: %+v", res.Groups.Learn)
	}
	if res.Groups.Resources == nil || res.Groups.Projects == nil || len(res.Groups.Resources.Items) != 0 {
		t.Fatalf("every group is present by default: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Fatalf("empty groups encode as [] not null: %s", rec.Body.String())
	}
	for _, l := range limits {
		if l != 2 {
			t.Fatalf("limits = %v, want per_group+1 = 2", limits)
		}
	}
}

func TestSearchDefaultsAndFiltersReachTheReader(t *testing.T) {
	var got domain.Query
	h := handler(fakeReader{search: func(_ domain.Group, q domain.Query) ([]domain.Hit, error) {
		got = q
		return nil, nil
	}})
	country, lang := "BF", apigen.LangFilter("fr")
	types := []apigen.SearchGroup{apigen.SearchGroupResources}
	rec := call(t, h, apigen.SearchParams{Q: "ab", Types: &types, Country: &country, Lang: &lang})
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	if got.Limit != 6 || got.Country != "BF" || got.Lang != "fr" {
		t.Fatalf("query = %+v, want default per_group 5 (limit 6) and both filters", got)
	}
	var res apigen.SearchResults
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	if res.Groups.Resources == nil || res.Groups.Questions != nil || res.Groups.Projects != nil || res.Groups.Learn != nil {
		t.Fatalf("only the requested group is present: %s", rec.Body.String())
	}
}

func TestSearchMapsErrorsToProblemJSON(t *testing.T) {
	zero := 0
	boom := errors.New("select * from secrets: connection refused")
	tests := []struct {
		name   string
		params apigen.SearchParams
		reader func(domain.Group, domain.Query) ([]domain.Hit, error)
		status int
		code   string
	}{
		{"per_group 0 is a 422", apigen.SearchParams{Q: "ab", PerGroup: &zero}, nil, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"short q is a 422", apigen.SearchParams{Q: "a"}, nil, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"a reader failure is a 500 without internals", apigen.SearchParams{Q: "ab"}, func(domain.Group, domain.Query) ([]domain.Hit, error) { return nil, boom }, http.StatusInternalServerError, "INTERNAL_ERROR"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := handler(fakeReader{search: tc.reader})
			rec := call(t, h, tc.params)
			if rec.Code != tc.status || rec.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("status %d type %s body %s", rec.Code, rec.Header().Get("Content-Type"), rec.Body.String())
			}
			var p struct {
				Code string `json:"code"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil || p.Code != tc.code {
				t.Fatalf("code = %q err %v body %s", p.Code, err, rec.Body.String())
			}
			if strings.Contains(rec.Body.String(), "secrets") || strings.Contains(rec.Body.String(), "connection refused") {
				t.Fatalf("internals leaked: %s", rec.Body.String())
			}
		})
	}
}

func TestSearchRejectsAnUnreadableHitID(t *testing.T) {
	h := handler(fakeReader{search: func(g domain.Group, _ domain.Query) ([]domain.Hit, error) {
		return []domain.Hit{{Type: g, ID: "not-a-uuid", Title: "x"}}, nil
	}})
	if rec := call(t, h, apigen.SearchParams{Q: "ab"}); rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
}
