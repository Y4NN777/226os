package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
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
func (f fakeReader) Get(_ context.Context, id string) (domain.Resource, error) {
	return f.get(id)
}
func (f fakeReader) Sources(_ context.Context, id string) ([]domain.Source, error) {
	return f.sources(id)
}

func codec(t *testing.T) *cursor.Codec {
	t.Helper()
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func str(s string) *string { return &s }

func TestListValidatesPageAndQuery(t *testing.T) {
	var got domain.Query
	svc := New(fakeReader{list: func(q domain.Query) ([]domain.Resource, error) {
		got = q
		return nil, nil
	}}, codec(t))
	long := strings.Repeat("é", 201)
	nul, badUTF8 := "pay\x00ment", "pay\xffment"
	zero, fiftyOne := 0, 51
	tests := []struct {
		name string
		req  ListRequest
		code apperr.Code
	}{
		{"limit zero", ListRequest{Limit: &zero}, apperr.CodeValidation},
		{"limit 51", ListRequest{Limit: &fiftyOne}, apperr.CodeValidation},
		{"empty q", ListRequest{Q: str("")}, apperr.CodeValidation},
		{"q over 200", ListRequest{Q: &long}, apperr.CodeValidation},
		{"q with a NUL byte", ListRequest{Q: &nul}, apperr.CodeValidation},
		{"q with invalid UTF-8", ListRequest{Q: &badUTF8}, apperr.CodeValidation},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.List(t.Context(), tc.req)
			var ae *apperr.Error
			if !errors.As(err, &ae) || ae.Code != tc.code {
				t.Fatalf("err = %v, want %s", err, tc.code)
			}
		})
	}
	one := 1
	if _, err := svc.List(t.Context(), ListRequest{Limit: &one, Q: str("paiement")}); err != nil {
		t.Fatal(err)
	}
	if got.Limit != 2 || got.Q != "paiement" {
		t.Fatalf("query = %+v, want limit 2 and q", got)
	}
}

func TestListCursorIsBoundToTheFilters(t *testing.T) {
	rows := []domain.Resource{{ID: "00000000-0000-7000-8000-000000000001", Name: "Alpha"}, {ID: "00000000-0000-7000-8000-000000000002", Name: "Beta"}}
	svc := New(fakeReader{list: func(q domain.Query) ([]domain.Resource, error) {
		if q.AfterName == "" {
			return rows, nil
		}
		if q.AfterName != "Alpha" || q.Country != "BF" {
			t.Fatalf("continued query = %+v", q)
		}
		return rows[1:], nil
	}}, codec(t))
	one := 1
	first, err := svc.List(t.Context(), ListRequest{Limit: &one, Country: str("BF")})
	if err != nil || first.NextCursor == "" || len(first.Items) != 1 || first.Items[0].Name != "Alpha" {
		t.Fatalf("first = %+v, err %v", first, err)
	}
	second, err := svc.List(t.Context(), ListRequest{Limit: &one, Country: str("BF"), Cursor: first.NextCursor})
	if err != nil || len(second.Items) != 1 || second.Items[0].Name != "Beta" || second.NextCursor != "" {
		t.Fatalf("second = %+v, err %v", second, err)
	}
	if _, err := svc.List(t.Context(), ListRequest{Limit: &one, Country: str("CI"), Cursor: first.NextCursor}); !errors.As(err, new(*apperr.Error)) {
		t.Fatalf("replay on another filter: %v", err)
	} else {
		var ae *apperr.Error
		_ = errors.As(err, &ae)
		if ae.Code != apperr.CodeInvalidCursor {
			t.Fatalf("code = %s", ae.Code)
		}
	}
	if _, err := svc.List(t.Context(), ListRequest{Cursor: "not-a-cursor"}); !errors.As(err, new(*apperr.Error)) {
		t.Fatal(err)
	}
}

func TestListRejectsACursorWithNoPosition(t *testing.T) {
	c := codec(t)
	token := c.Encode(cursor.Payload{FilterHash: filterHash(domain.Query{})})
	svc := New(fakeReader{list: func(domain.Query) ([]domain.Resource, error) {
		t.Fatal("store should not be called")
		return nil, nil
	}}, c)
	_, err := svc.List(t.Context(), ListRequest{Cursor: token})
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInvalidCursor {
		t.Fatalf("err = %v", err)
	}
}

func TestListReturnsStoreErrors(t *testing.T) {
	boom := errors.New("db down")
	svc := New(fakeReader{list: func(domain.Query) ([]domain.Resource, error) { return nil, boom }}, codec(t))
	_, err := svc.List(t.Context(), ListRequest{})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
}

func TestGetAndSourcesPassThrough(t *testing.T) {
	want := domain.Resource{ID: "r", Name: "Alpha", VerificationStatus: "verified", UpdatedAt: time.Unix(10, 0).UTC()}
	sources := []domain.Source{{ID: "s", Status: "unverified"}}
	svc := New(fakeReader{
		get:     func(id string) (domain.Resource, error) { return want, nil },
		sources: func(id string) ([]domain.Source, error) { return sources, nil },
	}, codec(t))
	got, err := svc.Get(t.Context(), "r")
	if err != nil || got.Name != "Alpha" || got.VerificationStatus != "verified" {
		t.Fatalf("got %+v err %v", got, err)
	}
	gotSources, err := svc.Sources(t.Context(), "r")
	if err != nil || len(gotSources) != 1 || gotSources[0].Status != "unverified" {
		t.Fatalf("sources %+v err %v", gotSources, err)
	}
}

// Two filter sets that the old "\x1f" join mapped to the same string must hash differently,
// otherwise a cursor issued for one filter set would be accepted for the other.
func TestFilterHashSeparatesFieldsThatContainTheOldSeparator(t *testing.T) {
	a := domain.Query{Q: "x\x1f", Country: "y"}
	b := domain.Query{Q: "x", Country: "\x1fy"}
	if oldJoin(a) != oldJoin(b) {
		t.Fatal("the fixture no longer collides under the old join")
	}
	if filterHash(a) == filterHash(b) {
		t.Fatalf("filterHash collides: %q", filterHash(a))
	}
	again := domain.Query{Q: "x\x1f", Country: "y"}
	if filterHash(a) != filterHash(again) || filterHash(domain.Query{}) == filterHash(domain.Query{Q: "a"}) {
		t.Fatal("filterHash must be deterministic and sensitive to each field")
	}
}

func TestListCursorIsRefusedOnACollidingFilterSet(t *testing.T) {
	rows := []domain.Resource{{ID: "00000000-0000-7000-8000-000000000001", Name: "Alpha"}, {ID: "00000000-0000-7000-8000-000000000002", Name: "Beta"}}
	svc := New(fakeReader{list: func(domain.Query) ([]domain.Resource, error) { return rows, nil }}, codec(t))
	one := 1
	first, err := svc.List(t.Context(), ListRequest{Limit: &one, Q: str("x\x1f"), Country: str("y")})
	if err != nil || first.NextCursor == "" {
		t.Fatalf("first = %+v, err %v", first, err)
	}
	_, err = svc.List(t.Context(), ListRequest{Limit: &one, Q: str("x"), Country: str("\x1fy"), Cursor: first.NextCursor})
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeInvalidCursor {
		t.Fatalf("err = %v", err)
	}
}

// oldJoin is the pre-fix encoding, kept only to prove the collision fixture is real.
func oldJoin(q domain.Query) string {
	return q.Q + "\x1f" + q.Country + "\x1f" + q.Category + "\x1f" + q.Type + "\x1f" + q.SDK + "\x1f" + q.Provider + "\x1f" + q.Doclang
}
