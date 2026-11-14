package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

func directorySvc(t *testing.T) (*Service, *fakeStore) {
	t.Helper()
	svc, st, _ := testSvc(t)
	codec, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	WithCursors(codec)(svc)
	return svc, st
}

func entries(n int) []domain.DirectoryEntry {
	out := make([]domain.DirectoryEntry, n)
	for i := range out {
		out[i] = domain.DirectoryEntry{
			UserID: fmt.Sprintf("00000000-0000-7000-8000-%012d", i+1), Username: fmt.Sprintf("user%02d", i+1), DisplayName: "U",
		}
	}
	return out
}

func appErr(t *testing.T, err error) *apperr.Error {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("error %v is not an apperr", err)
	}
	return ae
}

func TestListProfilesAudienceComesFromTheSession(t *testing.T) {
	for _, tc := range []struct {
		name, caller string
		member       bool
	}{
		{"anonymous", "", false},
		{"member", "00000000-0000-7000-8000-0000000000aa", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, st := directorySvc(t)
			if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{CallerUserID: tc.caller, ClientIP: "198.51.100.7"}); err != nil {
				t.Fatal(err)
			}
			if st.lastDir.Member != tc.member {
				t.Fatalf("Member = %v, want %v", st.lastDir.Member, tc.member)
			}
		})
	}
}

func TestListProfilesPagesWithOneLookAheadRow(t *testing.T) {
	svc, st := directorySvc(t)
	st.directory = entries(5)
	limit := 2
	var seen []string
	next := ""
	for pages := 0; pages < 5; pages++ {
		page, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7", Limit: &limit, Cursor: next})
		if err != nil {
			t.Fatal(err)
		}
		if st.lastDir.Limit != 3 {
			t.Fatalf("store limit %d, want page size + 1", st.lastDir.Limit)
		}
		for _, e := range page.Items {
			seen = append(seen, e.Username)
		}
		if next = page.NextCursor; next == "" {
			break
		}
	}
	want := []string{"user01", "user02", "user03", "user04", "user05"}
	if fmt.Sprint(seen) != fmt.Sprint(want) {
		t.Fatalf("pages walked %v, want %v", seen, want)
	}
}

func TestListProfilesEmptyPageIsNotNil(t *testing.T) {
	svc, _ := directorySvc(t)
	page, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7"})
	if err != nil || page.Items == nil || len(page.Items) != 0 || page.NextCursor != "" {
		t.Fatalf("page %+v err %v", page, err)
	}
	if st := page; st.NextCursor != "" {
		t.Fatal("empty page has a cursor")
	}
}

func TestListProfilesDefaultLimitIs20(t *testing.T) {
	svc, st := directorySvc(t)
	if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7"}); err != nil {
		t.Fatal(err)
	}
	if st.lastDir.Limit != 21 {
		t.Fatalf("store limit %d, want 21", st.lastDir.Limit)
	}
}

func TestListProfilesValidatesTheRequest(t *testing.T) {
	ptr := func(s string) *string { return &s }
	num := func(n int) *int { return &n }
	for _, tc := range []struct {
		name  string
		req   DirectoryRequest
		field string
	}{
		{"limit 51", DirectoryRequest{Limit: num(51)}, "limit"},
		{"limit 0", DirectoryRequest{Limit: num(0)}, "limit"},
		{"limit negative", DirectoryRequest{Limit: num(-1)}, "limit"},
		{"country lower case", DirectoryRequest{Country: ptr("bf")}, "country"},
		{"country three letters", DirectoryRequest{Country: ptr("BFA")}, "country"},
		{"language region subtag", DirectoryRequest{Language: ptr("fr-BF")}, "language"},
		{"language upper case", DirectoryRequest{Language: ptr("FR")}, "language"},
		{"skill with space", DirectoryRequest{Skill: ptr("go lang")}, "skill"},
		{"skill sql", DirectoryRequest{Skill: ptr("x'; DROP TABLE profiles;--")}, "skill"},
		{"skill too long", DirectoryRequest{Skill: ptr("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")}, "skill"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, st := directorySvc(t)
			tc.req.ClientIP = "198.51.100.7"
			_, err := svc.ListProfiles(context.Background(), tc.req)
			ae := appErr(t, err)
			if ae.Code != apperr.CodeValidation || len(ae.Violations) != 1 || ae.Violations[0].Field != tc.field {
				t.Fatalf("error %+v, want a %s violation", ae, tc.field)
			}
			if len(st.hits) != 0 {
				t.Fatal("an invalid request consumed rate limit budget")
			}
		})
	}
}

func TestListProfilesPassesFiltersToTheStore(t *testing.T) {
	svc, st := directorySvc(t)
	c, l, k := "BF", "fr", "go"
	if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7", Country: &c, Language: &l, Skill: &k}); err != nil {
		t.Fatal(err)
	}
	if st.lastDir.Country != "BF" || st.lastDir.Language != "fr" || st.lastDir.Skill != "go" {
		t.Fatalf("query %+v", st.lastDir)
	}
}

func TestListProfilesRejectsABadCursor(t *testing.T) {
	svc, st := directorySvc(t)
	st.directory = entries(3)
	limit := 1
	first, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7", Limit: &limit})
	if err != nil || first.NextCursor == "" {
		t.Fatalf("first page %+v %v", first, err)
	}
	other := "BF"
	tampered := first.NextCursor[:len(first.NextCursor)-2] + "AA"

	// A cursor whose payload is signed but not a keyset position.
	codec, _ := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	hash := lengthPrefixed("directory", "", "", "")
	badID := codec.Encode(cursor.Payload{Sort: "user01", ID: "not-a-uuid", FilterHash: hash})
	noSort := codec.Encode(cursor.Payload{Sort: "", ID: "00000000-0000-7000-8000-000000000001", FilterHash: hash})

	for _, tc := range []struct {
		name string
		req  DirectoryRequest
	}{
		{"tampered", DirectoryRequest{Cursor: tampered}},
		{"other filters", DirectoryRequest{Cursor: first.NextCursor, Country: &other}},
		{"garbage", DirectoryRequest{Cursor: "AAAA"}},
		{"bad id", DirectoryRequest{Cursor: badID}},
		{"no username", DirectoryRequest{Cursor: noSort}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.ClientIP = "198.51.100.7"
			_, err := svc.ListProfiles(context.Background(), tc.req)
			if ae := appErr(t, err); ae.Code != apperr.CodeInvalidCursor {
				t.Fatalf("code %v, want INVALID_CURSOR", ae.Code)
			}
		})
	}
}

func TestListProfilesStoreErrorIsInternal(t *testing.T) {
	svc, st := directorySvc(t)
	st.errs = map[string]error{"ListProfiles": errors.New("db down")}
	_, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7"})
	if ae := appErr(t, err); ae.Code != apperr.CodeInternal {
		t.Fatalf("code %v, want internal", ae.Code)
	}
}

func TestListProfilesRateLimit(t *testing.T) {
	const member = "00000000-0000-7000-8000-0000000000aa"
	for _, tc := range []struct {
		name string
		req  DirectoryRequest
		same DirectoryRequest // another caller that shares no budget with req
	}{
		{"anonymous per network", DirectoryRequest{ClientIP: "198.51.100.7"}, DirectoryRequest{ClientIP: "203.0.113.9"}},
		{"member per user", DirectoryRequest{CallerUserID: member}, DirectoryRequest{CallerUserID: "00000000-0000-7000-8000-0000000000bb"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, _ := directorySvc(t)
			for i := 0; i < 60; i++ {
				if _, err := svc.ListProfiles(context.Background(), tc.req); err != nil {
					t.Fatalf("call %d: %v", i+1, err)
				}
			}
			_, err := svc.ListProfiles(context.Background(), tc.req)
			ae := appErr(t, err)
			if ae.Code != apperr.CodeRateLimited || ae.RetryAfter <= 0 {
				t.Fatalf("61st call: %+v, want RATE_LIMITED with Retry-After", ae)
			}
			if _, err := svc.ListProfiles(context.Background(), tc.same); err != nil {
				t.Fatalf("another caller was limited: %v", err)
			}
		})
	}
}

func TestListProfilesAnonymousBudgetIsPerNetworkPrefix(t *testing.T) {
	svc, _ := directorySvc(t)
	for i := 0; i < 60; i++ {
		if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7"}); err != nil {
			t.Fatal(err)
		}
	}
	// Same /24, another host: the same bucket.
	if _, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.200"}); appErr(t, err).Code != apperr.CodeRateLimited {
		t.Fatal("a neighbour in the same /24 got its own budget")
	}
}

func TestListProfilesLimiterFailureFailsClosed(t *testing.T) {
	svc, st := directorySvc(t)
	st.errs = map[string]error{"Hit": errors.New("db down")}
	_, err := svc.ListProfiles(context.Background(), DirectoryRequest{ClientIP: "198.51.100.7"})
	if ae := appErr(t, err); ae.Code != apperr.CodeInternal {
		t.Fatalf("code %v, want internal", ae.Code)
	}
}
