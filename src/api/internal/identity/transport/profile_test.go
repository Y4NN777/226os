package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// seedOwnProfile creates the profile row the sign-up flow would have produced
// for the account created by signupAndSignIn.
func seedOwnProfile(t *testing.T, st *fakeStore, email string) domain.ProfileRecord {
	t.Helper()
	acct, err := st.AccountByEmail(t.Context(), email)
	if err != nil {
		t.Fatal(err)
	}
	rec := domain.ProfileRecord{
		Profile: domain.Profile{
			UserID: acct.ID, Username: acct.Username, DisplayName: acct.DisplayName,
			Visibility: domain.VisibilityPublic, VisCountry: domain.VisibilityPublic,
			VisLanguages: domain.VisibilityPublic, VisSkills: domain.VisibilityPublic,
			VisGitHub: domain.VisibilityPublic, VisBio: domain.VisibilityPublic,
			VisOpenCollab: domain.VisibilityPublic,
			Languages:     []string{}, Skills: []string{},
		},
		AccountStatus: "active", JoinedAt: time.Now(),
	}
	st.profiles[acct.Username] = rec
	return rec
}

// memberCall signs up + signs in, then invokes op as that member with the
// session cookie and CSRF header the middleware would have verified.
func memberCall(t *testing.T, h *Handler, st *fakeStore, op op, method, path, body, email string, seed func()) *httptest.ResponseRecorder {
	t.Helper()
	plain, csrf := signupAndSignIn(t, h, email)
	if seed != nil {
		seed()
	}
	var rdr *strings.Reader
	if body != "" {
		rdr = strings.NewReader(body)
	} else {
		rdr = strings.NewReader("")
	}
	req := httptest.NewRequestWithContext(t.Context(), method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrf)
	req.AddCookie(&http.Cookie{Name: "226os_session", Value: plain})
	acct, err := st.AccountByEmail(t.Context(), email)
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(withSubjectID(req, acct.ID, "member"))
	rec := httptest.NewRecorder()
	op(rec, req)
	return rec
}

func strptrLocal(s string) *string { return &s }

// Caller-dependent reads must not be stored by a shared cache (issue #116).
func TestGetMeAndGetProfileSendNoStore(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	signupAndSignIn(t, h, "cache@example.com")
	seedOwnProfile(t, st, "cache@example.com")
	acct, _ := st.AccountByEmail(t.Context(), "cache@example.com")

	rec := memberCall(t, h, st, h.GetMe, http.MethodGet, "/api/v1/me", "", "me@example.com",
		func() { seedOwnProfile(t, st, "me@example.com") })
	if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("getMe %d cache-control %q", rec.Code, rec.Header().Get("Cache-Control"))
	}
	rec = call(t, func(w http.ResponseWriter, r *http.Request) {
		h.GetProfile(w, r, acct.Username)
	}, doReq(t, http.MethodGet, "/api/v1/profiles/"+acct.Username, nil))
	if rec.Code != http.StatusOK || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("getProfile %d cache-control %q", rec.Code, rec.Header().Get("Cache-Control"))
	}
}

func TestGetMeReturnsAccountAndProfile(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	rec := memberCall(t, h, st, h.GetMe, http.MethodGet, "/api/v1/me", "", "me@example.com",
		func() { seedOwnProfile(t, st, "me@example.com") })
	if rec.Code != http.StatusOK {
		t.Fatalf("getMe %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Account struct {
			Email string `json:"email"`
		} `json:"account"`
		Profile struct {
			Username  string   `json:"username"`
			Checklist []string `json:"checklist_missing"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Account.Email != "me@example.com" || out.Profile.Username == "" {
		t.Fatalf("body %s", rec.Body.String())
	}
	if len(out.Profile.Checklist) == 0 {
		t.Fatal("expected non-empty checklist_missing on a fresh profile")
	}
}

func TestGetMeAnonymous401(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	rec := call(t, h.GetMe, doReq(t, http.MethodGet, "/api/v1/me", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous getMe: %d", rec.Code)
	}
}

func TestAuthzGetMeAnonymous(t *testing.T) {
	// Negative authz for GET /api/v1/me (AUTHORIZATION_MATRIX, allow: member).
	h, _, _, _ := testStack(t, false)
	rec := call(t, h.GetMe, doReq(t, http.MethodGet, "/api/v1/me", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "401") && !strings.Contains(rec.Body.String(), "UNAUTHENTICATED") {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestUpdateMyProfileMergePatch(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	rec := memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile",
		`{"display_name":"Patched","bio":"hello"}`, "patch@example.com",
		func() { seedOwnProfile(t, st, "patch@example.com") })
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	var direct struct {
		DisplayName string  `json:"display_name"`
		Bio         *string `json:"bio"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &direct); err != nil {
		t.Fatal(err)
	}
	if direct.DisplayName != "Patched" || direct.Bio == nil || *direct.Bio != "hello" {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestUpdateMyProfileExplicitNullClears(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	rec := memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile",
		`{"bio":null,"github":null}`, "null@example.com",
		func() {
			rec0 := seedOwnProfile(t, st, "null@example.com")
			rec0.Bio = strptrLocal("old")
			rec0.GitHub = strptrLocal("octocat")
			st.profiles[rec0.Username] = rec0
		})
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	var direct struct {
		Bio    *string `json:"bio"`
		Github *string `json:"github"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &direct); err != nil {
		t.Fatal(err)
	}
	if direct.Bio != nil || direct.Github != nil {
		t.Fatalf("expected cleared fields, body %s", rec.Body.String())
	}
}

func TestUpdateMyProfileSkills(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	st.tagIDs["go"] = "tag-1"
	rec := memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile",
		`{"skills":["go"]}`, "skills@example.com",
		func() { seedOwnProfile(t, st, "skills@example.com") })
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	var direct struct {
		Skills []string `json:"skills"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &direct); err != nil {
		t.Fatal(err)
	}
	if len(direct.Skills) != 1 || direct.Skills[0] != "go" {
		t.Fatalf("skills %v", direct.Skills)
	}
}

func TestUpdateMyProfileUnknownSkillRejected(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	rec := memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile",
		`{"skills":["nope"]}`, "badskill@example.com",
		func() { seedOwnProfile(t, st, "badskill@example.com") })
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestUpdateMyProfileAnonymous401(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/api/v1/me/profile",
		strings.NewReader(`{"display_name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.UpdateMyProfile(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous patch: %d", rec.Code)
	}
}

func TestAuthzUpdateMyProfileAnonymous(t *testing.T) {
	// Negative authz for PATCH /api/v1/me/profile (AUTHORIZATION_MATRIX,
	// allow: member): anonymous callers get 401 before any store access.
	h, _, _, _ := testStack(t, false)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/api/v1/me/profile",
		strings.NewReader(`{"display_name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.UpdateMyProfile(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "401") && !strings.Contains(rec.Body.String(), "UNAUTHENTICATED") {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestUpdateMyProfileMalformedBody(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	rec := memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile", `{not json`, "bad@example.com",
		func() { seedOwnProfile(t, st, "bad@example.com") })
	if rec.Code == http.StatusOK {
		t.Fatalf("malformed body accepted: %s", rec.Body.String())
	}
}

func TestGetProfileServesPublicProfile(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	signupAndSignIn(t, h, "pub@example.com")
	seedOwnProfile(t, st, "pub@example.com")
	acct, _ := st.AccountByEmail(t.Context(), "pub@example.com")
	rec := call(t, func(w http.ResponseWriter, r *http.Request) {
		h.GetProfile(w, r, acct.Username)
	}, doReq(t, http.MethodGet, "/api/v1/profiles/"+acct.Username, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("getProfile %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Username != acct.Username {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestGetProfileUnknown404(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	rec := call(t, func(w http.ResponseWriter, r *http.Request) {
		h.GetProfile(w, r, "ghost")
	}, doReq(t, http.MethodGet, "/api/v1/profiles/ghost", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("unknown profile: %d", rec.Code)
	}
}

// ---------- coverage: remaining branches ----------

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("read boom") }

func TestGetMeMemberWithoutProfileIs500(t *testing.T) {
	// Account exists but no profile row: the service reports an internal
	// error rather than a misleading 404.
	h, _, st, _ := testStack(t, false)
	rec := memberCall(t, h, st, h.GetMe, http.MethodGet, "/api/v1/me", "", "noprofile@example.com", nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestGetMeAccountJSONErrorIs500(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	// An account whose id is not a UUID makes accountJSON fail: internal.
	rec := memberCall(t, h, st, h.GetMe, http.MethodGet, "/api/v1/me", "", "badid@example.com", func() {
		for key, a := range st.accounts {
			if a.Email == "badid@example.com" {
				a.ID = "not-a-uuid"
				st.accounts[key] = a
			}
		}
		st.profiles["bad"] = domain.ProfileRecord{
			Profile:       domain.Profile{UserID: "not-a-uuid", Username: "bad", DisplayName: "B"},
			AccountStatus: "active",
		}
	})
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d %s", rec.Code, rec.Body.String())
	}
}

// A request whose session cannot be resolved is a 401, whatever the subject
// in the context says (the account always comes from the session).
func TestGetMeWithoutALiveSessionIs401(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	req := doReq(t, http.MethodGet, "/api/v1/me", nil)
	req = req.WithContext(withSubjectID(req, "019a0000-0000-7000-8000-0000000000aa", "member"))
	rec := call(t, h.GetMe, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d %s", rec.Code, rec.Body.String())
	}
}

func TestGetMeOptionalFieldsMapped(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	rec := memberCall(t, h, st, h.GetMe, http.MethodGet, "/api/v1/me", "", "opt@example.com",
		func() {
			r := seedOwnProfile(t, st, "opt@example.com")
			r.Intent = strptrLocal("help")
			r.CountryCode = strptrLocal("FR")
			r.Languages = nil // exercises the nonNilSlice nil branch
			st.profiles[r.Username] = r
		})
	if rec.Code != http.StatusOK {
		t.Fatalf("getMe %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Profile struct {
			Intent    *string  `json:"intent"`
			Country   *string  `json:"country"`
			Languages []string `json:"languages"`
		} `json:"profile"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Profile.Intent == nil || *out.Profile.Intent != "help" ||
		out.Profile.Country == nil || *out.Profile.Country != "FR" ||
		out.Profile.Languages == nil {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestUpdateMyProfileFullPatch(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	st.tagIDs["go"] = "tag-1"
	rec := memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile",
		`{"display_name":"Full","intent":"help","country":"FR","languages":["fr","en"],`+
			`"github":"octocat","bio":"bio text","open_to_collab":true,`+
			`"profile_visibility":"members",`+
			`"field_visibility":{"bio":"members","country":"private","languages":"public",`+
			`"skills":"members","github":"private"},"skills":["go"]}`,
		"full@example.com",
		func() { seedOwnProfile(t, st, "full@example.com") })
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Intent            *string  `json:"intent"`
		Country           *string  `json:"country"`
		Languages         []string `json:"languages"`
		Github            *string  `json:"github"`
		Bio               *string  `json:"bio"`
		OpenToCollab      bool     `json:"open_to_collab"`
		ProfileVisibility string   `json:"profile_visibility"`
		FieldVisibility   struct {
			Bio *string `json:"bio"`
		} `json:"field_visibility"`
		Skills []string `json:"skills"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Intent == nil || *out.Intent != "help" || out.Country == nil || *out.Country != "FR" ||
		len(out.Languages) != 2 || out.Github == nil || out.Bio == nil || !out.OpenToCollab ||
		out.ProfileVisibility != "members" || len(out.Skills) != 1 {
		t.Fatalf("body %s", rec.Body.String())
	}
}

func TestUpdateMyProfileNullScalars(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	rec := memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile",
		`{"intent":null,"country":null,"github":null}`,
		"nulls@example.com",
		func() {
			r := seedOwnProfile(t, st, "nulls@example.com")
			r.Intent = strptrLocal("help")
			r.CountryCode = strptrLocal("FR")
			r.GitHub = strptrLocal("octocat")
			st.profiles[r.Username] = r
		})
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Intent  *string `json:"intent"`
		Country *string `json:"country"`
		Github  *string `json:"github"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Intent != nil || out.Country != nil || out.Github != nil {
		t.Fatalf("expected cleared, body %s", rec.Body.String())
	}
}

// Explicit null on contract-non-nullable members is dropped by the handler
// (the embedded OpenAPI validator rejects it before this point in
// production). Pinned so the merge-patch mapping stays honest: a null-only
// body is an empty patch (422), and null members do not block other fields.
func TestUpdateMyProfileNullNonNullableDropped(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	seed := func() { seedOwnProfile(t, st, "nulls@example.com") }

	rec := memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile",
		`{"skills":null,"languages":null,"open_to_collab":null}`, "nulls@example.com", seed)
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("null-only patch: expected 422, got %d %s", rec.Code, rec.Body.String())
	}

	rec = memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile",
		`{"skills":null,"display_name":"Nulls"}`, "nulls@example.com", seed)
	if rec.Code != http.StatusOK {
		t.Fatalf("mixed patch: expected 200, got %d %s", rec.Code, rec.Body.String())
	}
	var body struct {
		DisplayName string   `json:"display_name"`
		Skills      []string `json:"skills"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.DisplayName != "Nulls" || len(body.Skills) != 0 {
		t.Fatalf("mixed patch result %s", rec.Body.String())
	}
}

func TestUpdateMyProfileBodyTooLarge(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	big := `{"display_name":"` + strings.Repeat("a", 1<<20) + `"}`
	rec := memberCall(t, h, st, h.UpdateMyProfile, http.MethodPatch, "/api/v1/me/profile",
		big, "big@example.com",
		func() { seedOwnProfile(t, st, "big@example.com") })
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestUpdateMyProfileReadError(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/api/v1/me/profile", errReader{})
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(withSubject(req, "member"))
	rec := httptest.NewRecorder()
	h.UpdateMyProfile(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("read error accepted: %s", rec.Body.String())
	}
}

func TestUpdateMyProfileUnknownFieldRejected(t *testing.T) {
	h, _, _, _ := testStack(t, false)
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/api/v1/me/profile",
		strings.NewReader(`{"bogus_field":1}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(withSubject(req, "member"))
	rec := httptest.NewRecorder()
	h.UpdateMyProfile(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("unknown field accepted: %s", rec.Body.String())
	}
}

func TestGetProfileAllFieldsVisible(t *testing.T) {
	h, _, st, _ := testStack(t, false)
	signupAndSignIn(t, h, "all@example.com")
	rec0 := seedOwnProfile(t, st, "all@example.com")
	rec0.Bio = strptrLocal("bio")
	rec0.CountryCode = strptrLocal("FR")
	rec0.Languages = []string{"fr"}
	rec0.Skills = []string{"go"}
	rec0.GitHub = strptrLocal("octocat")
	rec0.OpenToCollab = true
	st.profiles[rec0.Username] = rec0
	acct, _ := st.AccountByEmail(t.Context(), "all@example.com")
	rec := call(t, func(w http.ResponseWriter, r *http.Request) {
		h.GetProfile(w, r, acct.Username)
	}, doReq(t, http.MethodGet, "/api/v1/profiles/"+acct.Username, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("getProfile %d %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Bio          *string   `json:"bio"`
		Country      *string   `json:"country"`
		Languages    *[]string `json:"languages"`
		Skills       *[]string `json:"skills"`
		Github       *string   `json:"github"`
		OpenToCollab *bool     `json:"open_to_collab"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if out.Bio == nil || out.Country == nil || out.Languages == nil || out.Skills == nil ||
		out.Github == nil || out.OpenToCollab == nil {
		t.Fatalf("expected all fields, body %s", rec.Body.String())
	}
}

// GetMe is on the page-load path: the session middleware already loaded the
// account, so the request must not read it a second time.
func TestGetMeLoadsTheAccountOnce(t *testing.T) {
	h, svc, st, _ := testStack(t, false)
	plain, _ := signupAndSignIn(t, h, "once-me@example.com")
	// A profile exists, so the request succeeds: the count below is the success path (review of #110: the test used to
	// pass on a 500 caused by a missing profile and proved nothing about a successful GET /me).
	seedOwnProfile(t, st, "once-me@example.com")
	before := st.accountLookups
	chain := subject.Middleware(NewResolver(svc, false), nil)(http.HandlerFunc(h.GetMe))
	req := doReq(t, http.MethodGet, "/api/v1/me", nil, &http.Cookie{Name: sessionCookieName(false), Value: plain})
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /me: %d %s", rec.Code, rec.Body.String())
	}
	if got := st.accountLookups - before; got != 1 {
		t.Fatalf("account lookups for GET /me = %d, want 1 (the session resolution)", got)
	}
}
