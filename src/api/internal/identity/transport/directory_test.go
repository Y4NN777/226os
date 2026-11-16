package transport

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

func listProfiles(t *testing.T, h *Handler, params apigen.ListProfilesParams, member bool) (code int, header http.Header, body string) {
	t.Helper()
	req := doReq(t, http.MethodGet, "/api/v1/profiles", map[string]string{"X-Forwarded-For": "198.51.100.7"})
	if member {
		req = req.WithContext(withSubjectID(req, "019a0000-0000-7000-8000-0000000000aa", "member"))
	}
	rec := call(t, func(w http.ResponseWriter, r *http.Request) { h.ListProfiles(w, r, params) }, req)
	return rec.Code, rec.Header(), rec.Body.String()
}

func ptr[T any](v T) *T { return &v }

func TestListProfilesServesTheSummaryWithoutHiddenFields(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	st.directory = []domain.DirectoryEntry{
		{UserID: "019a0000-0000-7000-8000-000000000001", Username: "ada", DisplayName: "Ada", Country: ptr("BF"),
			Languages: []string{"fr"}, Skills: []string{"go"}, GitHub: ptr("ada")},
		{UserID: "019a0000-0000-7000-8000-000000000002", Username: "bob", DisplayName: "Bob"}, // everything hidden
	}
	code, header, body := listProfiles(t, h, apigen.ListProfilesParams{}, false)
	if code != http.StatusOK || header.Get("Cache-Control") != "no-store" {
		t.Fatalf("status %d cache-control %q body %s", code, header.Get("Cache-Control"), body)
	}
	var out struct {
		Items      []map[string]any `json:"items"`
		NextCursor *string          `json:"next_cursor"`
	}
	if err := json.Unmarshal([]byte(body), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Items) != 2 || out.NextCursor != nil {
		t.Fatalf("body %s", body)
	}
	if len(out.Items[0]) != 6 || out.Items[0]["country"] != "BF" || out.Items[0]["github"] != "ada" {
		t.Fatalf("visible entry %v", out.Items[0])
	}
	if len(out.Items[1]) != 2 {
		t.Fatalf("hidden fields must be omitted, not null: %v", out.Items[1])
	}
	if strings.Contains(strings.ToLower(body), "email") || strings.Contains(body, "user_id") {
		t.Fatalf("body leaks an identity detail: %s", body)
	}
	if st.lastDir.Member {
		t.Fatal("an anonymous call was treated as a member")
	}
}

func TestListProfilesMemberAudienceAndNextCursor(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	for i := 1; i <= 3; i++ {
		st.directory = append(st.directory, domain.DirectoryEntry{
			UserID: "019a0000-0000-7000-8000-00000000000" + string(rune('0'+i)), Username: "user" + string(rune('0'+i)), DisplayName: "U",
		})
	}
	code, _, body := listProfiles(t, h, apigen.ListProfilesParams{Limit: ptr(2)}, true)
	if code != http.StatusOK || !st.lastDir.Member {
		t.Fatalf("status %d member %v", code, st.lastDir.Member)
	}
	var out struct {
		Items      []map[string]any `json:"items"`
		NextCursor *string          `json:"next_cursor"`
	}
	if err := json.Unmarshal([]byte(body), &out); err != nil || len(out.Items) != 2 || out.NextCursor == nil {
		t.Fatalf("body %s err %v", body, err)
	}
	// The cursor comes back as a param and reaches the store as the keyset position.
	code, _, _ = listProfiles(t, h, apigen.ListProfilesParams{Limit: ptr(2), Cursor: out.NextCursor}, true)
	if code != http.StatusOK || st.lastDir.AfterUsername != "user2" {
		t.Fatalf("status %d after %q", code, st.lastDir.AfterUsername)
	}
}

func TestListProfilesRejectsInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name   string
		params apigen.ListProfilesParams
		status int
		code   string
	}{
		{"limit above 50", apigen.ListProfilesParams{Limit: ptr(51)}, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"bad skill", apigen.ListProfilesParams{Skill: ptr("not a slug")}, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"bad language", apigen.ListProfilesParams{Language: ptr("FR")}, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"bad country", apigen.ListProfilesParams{Country: ptr(apigen.CountryCode("xx"))}, http.StatusUnprocessableEntity, "VALIDATION_ERROR"},
		{"tampered cursor", apigen.ListProfilesParams{Cursor: ptr("AAAA")}, http.StatusBadRequest, "INVALID_CURSOR"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, _, _, _ := testStack(t, false)
			code, header, body := listProfiles(t, h, tc.params, false)
			if code != tc.status || !strings.Contains(body, tc.code) || !strings.HasPrefix(header.Get("Content-Type"), "application/problem+json") {
				t.Fatalf("status %d %q body %s", code, header.Get("Content-Type"), body)
			}
		})
	}
}

func TestListProfilesRateLimitIs429WithRetryAfter(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	for i := 0; i < 60; i++ {
		if code, _, body := listProfiles(t, h, apigen.ListProfilesParams{}, false); code != http.StatusOK {
			t.Fatalf("call %d: %d %s", i+1, code, body)
		}
	}
	code, header, body := listProfiles(t, h, apigen.ListProfilesParams{}, false)
	if code != http.StatusTooManyRequests || header.Get("Retry-After") == "" || !strings.Contains(body, "RATE_LIMITED") {
		t.Fatalf("61st call: %d retry-after %q body %s", code, header.Get("Retry-After"), body)
	}
}

func TestListProfilesStoreFailureIsARedacted500(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	st.failList = true
	code, _, body := listProfiles(t, h, apigen.ListProfilesParams{}, false)
	if code != http.StatusInternalServerError || strings.Contains(body, "db down") {
		t.Fatalf("status %d body %s", code, body)
	}
}
