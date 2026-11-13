package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/search/domain"
)

type call struct {
	group domain.Group
	query domain.Query
}

// fakeReader returns n synthetic hits per group (n defaults to 0) and records every call.
type fakeReader struct {
	mu    sync.Mutex
	calls []call
	rows  map[domain.Group]int
	err   map[domain.Group]error
}

func (f *fakeReader) Search(_ context.Context, g domain.Group, q domain.Query) ([]domain.Hit, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, call{g, q})
	if err := f.err[g]; err != nil {
		return nil, err
	}
	n := min(f.rows[g], q.Limit)
	if n == 0 {
		return nil, nil
	}
	hits := make([]domain.Hit, n)
	for i := range hits {
		hits[i] = domain.Hit{Type: g, ID: fmt.Sprintf("id-%s-%d", g, i), Title: "t"}
	}
	return hits, nil
}

func ints(v int) *int      { return &v }
func str(s string) *string { return &s }
func violation(t *testing.T, err error, field string) {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeValidation {
		t.Fatalf("err = %v, want a validation error", err)
	}
	if len(ae.Violations) != 1 || ae.Violations[0].Field != field {
		t.Fatalf("violations = %+v, want one on %q", ae.Violations, field)
	}
}

func TestSearchAllGroupsByDefaultWithPerGroupFive(t *testing.T) {
	r := &fakeReader{rows: map[domain.Group]int{domain.Questions: 9, domain.Resources: 2}}
	got, err := New(r).Search(t.Context(), Request{Q: "paiement"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Groups) != 4 || len(r.calls) != 4 {
		t.Fatalf("groups %d, calls %d, want 4 and 4", len(got.Groups), len(r.calls))
	}
	for _, c := range r.calls {
		if c.query.Limit != 6 {
			t.Fatalf("%s limit = %d, want per_group+1 = 6", c.group, c.query.Limit)
		}
	}
	q := got.Groups[domain.Questions]
	if len(q.Items) != 5 || !q.HasMore {
		t.Fatalf("questions: %d items has_more=%v, want 5 and true", len(q.Items), q.HasMore)
	}
	res := got.Groups[domain.Resources]
	if len(res.Items) != 2 || res.HasMore {
		t.Fatalf("resources: %d items has_more=%v, want 2 and false", len(res.Items), res.HasMore)
	}
	for _, g := range []domain.Group{domain.Projects, domain.Learn} {
		if items := got.Groups[g].Items; items == nil || len(items) != 0 || got.Groups[g].HasMore {
			t.Fatalf("%s: items %#v, want an empty non-nil list", g, items)
		}
	}
}

func TestSearchHasMoreIsExactAtTheBoundary(t *testing.T) {
	for _, tc := range []struct {
		rows     int
		wantMore bool
		want     int
	}{{2, false, 2}, {3, false, 3}, {4, true, 3}, {50, true, 3}} {
		r := &fakeReader{rows: map[domain.Group]int{domain.Learn: tc.rows}}
		got, err := New(r).Search(t.Context(), Request{Q: "ab", Types: []string{"learn"}, PerGroup: ints(3)})
		if err != nil {
			t.Fatal(err)
		}
		g := got.Groups[domain.Learn]
		if len(g.Items) != tc.want || g.HasMore != tc.wantMore {
			t.Fatalf("rows %d: %d items has_more=%v, want %d and %v", tc.rows, len(g.Items), g.HasMore, tc.want, tc.wantMore)
		}
	}
}

func TestSearchOnlyRequestedTypesAreQueriedAndReturned(t *testing.T) {
	r := &fakeReader{}
	got, err := New(r).Search(t.Context(), Request{Q: "ab", Types: []string{"projects", "questions"}, PerGroup: ints(20)})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := got.Groups[domain.Projects]; !ok || len(got.Groups) != 2 || len(r.calls) != 2 {
		t.Fatalf("groups %v, calls %d", got.Groups, len(r.calls))
	}
	if _, ok := got.Groups[domain.Resources]; ok {
		t.Fatal("resources was not requested")
	}
	for _, c := range r.calls {
		if c.query.Limit != 21 {
			t.Fatalf("limit = %d, want 21", c.query.Limit)
		}
	}
	empty, err := New(&fakeReader{}).Search(t.Context(), Request{Q: "ab", Types: []string{}})
	if err != nil || len(empty.Groups) != 4 {
		t.Fatalf("an empty types list means all groups: %d groups, %v", len(empty.Groups), err)
	}
}

// C4 PROPOSED option 1, ADR-0005 point 6.
func TestSearchTruncatesTheTextButEchoesTheOriginalC4ProposedOption1(t *testing.T) {
	original := strings.Repeat("mot ", 40) // 160 runes
	r := &fakeReader{}
	got, err := New(r).Search(t.Context(), Request{Q: original, Types: []string{"learn"}})
	if err != nil {
		t.Fatal(err)
	}
	if got.Query != original {
		t.Fatalf("Query = %q, want the original q echoed", got.Query)
	}
	sent := r.calls[0].query.Text
	if len([]rune(sent)) > 100 || sent == original || !strings.HasPrefix(original, sent) || strings.HasSuffix(sent, " ") {
		t.Fatalf("text sent to the reader = %q", sent)
	}
}

func TestSearchPassesFiltersThrough(t *testing.T) {
	r := &fakeReader{}
	if _, err := New(r).Search(t.Context(), Request{Q: "ab", Types: []string{"questions"}, Country: str("BF"), Lang: str("fr")}); err != nil {
		t.Fatal(err)
	}
	if q := r.calls[0].query; q.Country != "BF" || q.Lang != "fr" {
		t.Fatalf("query = %+v", q)
	}
	if _, err := New(r).Search(t.Context(), Request{Q: "ab", Types: []string{"questions"}}); err != nil {
		t.Fatal(err)
	}
	if q := r.calls[1].query; q.Country != "" || q.Lang != "" {
		t.Fatalf("absent filters must stay empty: %+v", q)
	}
}

func TestSearchValidation(t *testing.T) {
	tests := []struct {
		name  string
		req   Request
		field string
	}{
		{"q of 1 rune", Request{Q: "é"}, "q"},
		{"empty q", Request{Q: ""}, "q"},
		{"q with a NUL byte, which PostgreSQL text cannot hold", Request{Q: "ab\x00cd"}, "q"},
		{"q that is not valid UTF-8", Request{Q: "ab\xffcd"}, "q"},
		{"q over 200 runes", Request{Q: strings.Repeat("é", 201)}, "q"},
		{"per_group 0", Request{Q: "ab", PerGroup: ints(0)}, "per_group"},
		{"per_group 21", Request{Q: "ab", PerGroup: ints(21)}, "per_group"},
		{"unknown type", Request{Q: "ab", Types: []string{"answers"}}, "types"},
		{"duplicate type", Request{Q: "ab", Types: []string{"learn", "learn"}}, "types"},
		{"five types", Request{Q: "ab", Types: []string{"learn", "projects", "questions", "resources", "learn"}}, "types"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &fakeReader{}
			_, err := New(r).Search(t.Context(), tc.req)
			violation(t, err, tc.field)
			if len(r.calls) != 0 {
				t.Fatal("the reader must not run on invalid input")
			}
		})
	}
	if _, err := New(&fakeReader{}).Search(t.Context(), Request{Q: strings.Repeat("é", 200), PerGroup: ints(1)}); err != nil {
		t.Fatalf("200 runes is the maximum and must pass: %v", err)
	}
}

func TestSearchReturnsReaderErrors(t *testing.T) {
	boom := errors.New("db down")
	r := &fakeReader{err: map[domain.Group]error{domain.Projects: boom}}
	if _, err := New(r).Search(t.Context(), Request{Q: "ab"}); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}
