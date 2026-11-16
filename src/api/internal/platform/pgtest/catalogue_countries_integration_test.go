//go:build integration

package pgtest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

const catGamma = "019a0000-0000-7000-8000-0000000000b6"

// seedMultiCountry adds Gamma Pay, which supports both BF and GH, and marks Beta Pay as a demo fixture.
func seedMultiCountry(ctx context.Context, t *testing.T, s *s2Stack) {
	t.Helper()
	for _, stmt := range []string{
		`INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, doc_languages, created_by)
		 VALUES ('` + catGamma + `', '` + catProvider + `', 'mobile-money', 'api', 'gamma-pay', 'Gamma Pay', 'Deux pays', 'fr', '{fr}', '` + catUser + `')`,
		`INSERT INTO resource_countries (resource_id, country_code) VALUES ('` + catGamma + `', 'BF'), ('` + catGamma + `', 'GH')`,
		`UPDATE resources SET provenance = 'demo' WHERE id = '` + catBeta + `'`,
	} {
		if _, err := s.conn.Exec(ctx, stmt); err != nil {
			t.Fatalf("seed: %v\n%s", err, stmt)
		}
	}
}

func (s *s2Stack) list(t *testing.T, query string) (*apigen.ResourcePage, int, string) {
	t.Helper()
	rec := s.do(t, http.MethodGet, "/api/v1/resources?"+query, "", nil)
	var page apigen.ResourcePage
	if rec.Code == http.StatusOK {
		if err := json.Unmarshal(rec.Body.Bytes(), &page); err != nil {
			t.Fatal(err)
		}
	}
	return &page, rec.Code, rec.Body.String()
}

func TestResourceProvenanceIsServed(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	seedMultiCountry(ctx, t, s)

	page, code, body := s.list(t, "")
	if code != http.StatusOK {
		t.Fatalf("list %d %s", code, body)
	}
	got := map[string]string{}
	for _, item := range page.Items {
		got[item.Name] = string(item.Provenance)
	}
	want := map[string]string{"Alpha Pay": "live", "Beta Pay": "demo", "Gamma Pay": "live"}
	if len(got) != len(want) || got["Alpha Pay"] != "live" || got["Beta Pay"] != "demo" || got["Gamma Pay"] != "live" {
		t.Fatalf("provenance by name = %v, want %v", got, want)
	}

	rec := s.do(t, http.MethodGet, "/api/v1/resources/"+catBeta, "", nil)
	var one apigen.Resource
	if err := json.Unmarshal(rec.Body.Bytes(), &one); err != nil || rec.Code != http.StatusOK || one.Provenance != apigen.ResourceProvenance("demo") {
		t.Fatalf("get %d provenance %q err %v: %s", rec.Code, one.Provenance, err, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"provenance":"demo"`) {
		t.Fatalf("wire body has no provenance: %s", rec.Body.String())
	}
}

func TestResourceCountryListIsAUnion(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	seedMultiCountry(ctx, t, s)

	namesOf := func(query string) []string {
		page, code, body := s.list(t, query)
		if code != http.StatusOK {
			t.Fatalf("%q: %d %s", query, code, body)
		}
		return names(*page)
	}
	tests := []struct {
		query string
		want  []string
	}{
		{"country=BF", []string{"Alpha Pay", "Gamma Pay"}},
		{"country=GH", []string{"Beta Pay", "Gamma Pay"}},
		// Union: Gamma supports both requested countries and still appears once.
		{"country=BF,GH", []string{"Alpha Pay", "Beta Pay", "Gamma Pay"}},
		{"country=GH,BF,BF", []string{"Alpha Pay", "Beta Pay", "Gamma Pay"}},
		{"country=BF,CI", []string{"Alpha Pay", "Gamma Pay"}},
		{"country=CI,SN", []string{}},
		{"", []string{"Alpha Pay", "Beta Pay", "Gamma Pay"}},
	}
	for _, tc := range tests {
		if got := namesOf(tc.query); !slices.Equal(got, tc.want) {
			t.Errorf("%q = %v, want %v", tc.query, got, tc.want)
		}
	}
}

func TestResourceCountryListPaginatesStably(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)
	seedMultiCountry(ctx, t, s)

	var seen []string
	query := "country=GH,BF&limit=1"
	for range 5 {
		page, code, body := s.list(t, query)
		if code != http.StatusOK {
			t.Fatalf("%d %s", code, body)
		}
		seen = append(seen, names(*page)...)
		if page.NextCursor == nil {
			break
		}
		// The same filter written in another order accepts the cursor: BF,GH is GH,BF.
		query = "country=BF,GH&limit=1&cursor=" + url.QueryEscape(*page.NextCursor)
	}
	if want := []string{"Alpha Pay", "Beta Pay", "Gamma Pay"}; !slices.Equal(seen, want) {
		t.Fatalf("walked %v, want %v", seen, want)
	}

	first, _, _ := s.list(t, "country=BF,GH&limit=1")
	rec := s.do(t, http.MethodGet, "/api/v1/resources?country=BF&limit=1&cursor="+url.QueryEscape(*first.NextCursor), "", nil)
	assertProblem(t, rec, http.StatusBadRequest, "INVALID_CURSOR")
}

func TestResourceCountryListRefusesBadInput(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 90*time.Second)
	defer cancel()
	s := newS2Stack(ctx, t, nil)

	twenty := make([]string, 20)
	twentyOne := make([]string, 21)
	for i := range twentyOne {
		c := string(rune('A'+i/26)) + string(rune('A'+i%26))
		twentyOne[i] = c
		if i < 20 {
			twenty[i] = c
		}
	}
	if _, code, body := s.list(t, "country="+strings.Join(twenty, ",")); code != http.StatusOK {
		t.Fatalf("20 codes: %d %s", code, body)
	}
	for name, query := range map[string]string{
		"21 codes":  "country=" + strings.Join(twentyOne, ","),
		"lowercase": "country=bf",
		"one bad":   "country=BF,XYZ",
	} {
		rec := s.do(t, http.MethodGet, "/api/v1/resources?"+query, "", nil)
		p := assertProblem(t, rec, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		v, _ := json.Marshal(p["violations"])
		if !strings.Contains(string(v), `"field":"country"`) {
			t.Errorf("%s: violations %s, want one on country", name, v)
		}
	}
	// A value the parameter binder cannot read at all (an empty list, the parameter given twice) is the
	// existing 400 for a malformed parameter, with the violation on country.
	for name, query := range map[string]string{"empty": "country=", "repeated": "country=BF&country=GH"} {
		rec := s.do(t, http.MethodGet, "/api/v1/resources?"+query, "", nil)
		p := assertProblem(t, rec, http.StatusBadRequest, "MALFORMED_REQUEST")
		if v, _ := json.Marshal(p["violations"]); !strings.Contains(string(v), `"field":"country"`) {
			t.Errorf("%s: violations %s, want one on country", name, v)
		}
	}
}
