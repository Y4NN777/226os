package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

func seedProfile(st *fakeStore, username, userID string, vis domain.Visibility) domain.ProfileRecord {
	rec := domain.ProfileRecord{
		Profile: domain.Profile{
			UserID: userID, Username: username, DisplayName: "Display " + username,
			Visibility: vis, VisCountry: domain.VisibilityPublic,
			VisLanguages: domain.VisibilityPublic, VisSkills: domain.VisibilityPublic,
			VisGitHub: domain.VisibilityPublic, VisBio: domain.VisibilityPublic,
			VisOpenCollab: domain.VisibilityPublic,
			Languages:     []string{}, Skills: []string{},
		},
		AccountStatus: "active", JoinedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	st.profiles[username] = rec
	return rec
}

func strp(s string) *string { return &s }

// ---------- GetMe ----------

func TestGetMeReturnsAccountAndProfile(t *testing.T) {
	svc, st, _ := testSvc(t)
	acct := domain.Account{ID: "u1", Email: "a@b.c", Username: "ann", DisplayName: "Ann", Status: "active"}
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	view, err := svc.GetMe(context.Background(), acct)
	if err != nil {
		t.Fatal(err)
	}
	if view.Account.ID != "u1" || view.Account.Email != "a@b.c" || view.Profile.Username != "ann" {
		t.Fatalf("unexpected view %+v", view)
	}
}

// The account is the one the session resolution already loaded: GetMe never
// reads it again (its not-found and store-error cases are ResolveSession's).
func TestGetMeDoesNotReadTheAccountAgain(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	st.errs = map[string]error{"AccountByID": errors.New("must not be called")}
	if _, err := svc.GetMe(context.Background(), domain.Account{ID: "u1", Status: "active"}); err != nil {
		t.Fatalf("GetMe read the account again: %v", err)
	}
}

func TestGetMeProfileErrorIsInternal(t *testing.T) {
	svc, st, _ := testSvc(t)
	st.errs = map[string]error{"ProfileByUserID": errors.New("db down")}
	_, err := svc.GetMe(context.Background(), domain.Account{ID: "u1", Status: "active"})
	if apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("expected %v, got %v", apperr.CodeInternal, err)
	}
}

// ---------- UpdateMyProfile ----------

func TestUpdateMyProfileMergePatch(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	bio := "hello"
	gh := "octocat"
	w := domain.ProfileWrite{
		Bio:         domain.PatchField[string]{Present: true, Value: &bio},
		GitHub:      domain.PatchField[string]{Present: true, Value: &gh},
		DisplayName: domain.PatchField[string]{Present: true, Value: strp("New Name")},
	}
	prof, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if err != nil {
		t.Fatal(err)
	}
	if prof.Bio == nil || *prof.Bio != "hello" || prof.GitHub == nil || *prof.GitHub != "octocat" ||
		prof.DisplayName != "New Name" {
		t.Fatalf("patch not applied: %+v", prof)
	}
	// bio_language derived: bio set => 'en' stored by the repository layer;
	// the service only guarantees the patch carries it.
}

func TestUpdateMyProfileClearNullableFields(t *testing.T) {
	svc, st, _ := testSvc(t)
	rec := seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	rec.Bio = strp("old bio")
	rec.CountryCode = strp("FR")
	st.profiles["ann"] = rec
	w := domain.ProfileWrite{
		Bio:         domain.PatchField[string]{Present: true, Value: nil},
		CountryCode: domain.PatchField[string]{Present: true, Value: nil},
	}
	prof, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if err != nil {
		t.Fatal(err)
	}
	if prof.Bio != nil || prof.CountryCode != nil {
		t.Fatalf("expected cleared fields, got %+v", prof)
	}
}

func TestUpdateMyProfileAbsentFieldsUnchanged(t *testing.T) {
	svc, st, _ := testSvc(t)
	rec := seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	rec.Intent = strp("help")
	st.profiles["ann"] = rec
	w := domain.ProfileWrite{
		DisplayName: domain.PatchField[string]{Present: true, Value: strp("Only Name")},
	}
	prof, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if err != nil {
		t.Fatal(err)
	}
	if prof.Intent == nil || *prof.Intent != "help" {
		t.Fatalf("absent field was modified: %+v", prof.Intent)
	}
}

func TestUpdateMyProfileEmptyPatchRejected(t *testing.T) {
	svc, _, _ := testSvc(t)
	_, err := svc.UpdateMyProfile(context.Background(), "u1", domain.ProfileWrite{})
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("expected %v, got %v", apperr.CodeValidation, err)
	}
}

func TestUpdateMyProfileValidationViolations(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	bad := "not a country!"
	w := domain.ProfileWrite{
		CountryCode: domain.PatchField[string]{Present: true, Value: &bad},
		Intent:      domain.PatchField[string]{Present: true, Value: strp("bogus")},
		GitHub:      domain.PatchField[string]{Present: true, Value: strp("bad name!")},
		Languages:   domain.PatchField[[]string]{Present: true, Value: &[]string{"en", "en"}},
		Bio:         domain.PatchField[string]{Present: true, Value: strp(longString(501))},
		Visibility:  domain.PatchField[domain.Visibility]{Present: true, Value: func() *domain.Visibility { v := domain.Visibility("bogus"); return &v }()},
	}
	_, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("expected %v, got %v", apperr.CodeValidation, err)
	}
	var pe *apperr.Error
	if errors.As(err, &pe) {
		if len(pe.Violations) < 5 {
			t.Fatalf("expected multiple violations, got %+v", pe.Violations)
		}
	}
}

func TestUpdateMyProfileDisplayNameNilAndTooLong(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	w := domain.ProfileWrite{DisplayName: domain.PatchField[string]{Present: true, Value: nil}}
	_, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("expected %v, got %v", apperr.CodeValidation, err)
	}

	long := longString(81)
	w = domain.ProfileWrite{DisplayName: domain.PatchField[string]{Present: true, Value: &long}}
	_, err = svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("expected %v, got %v", apperr.CodeValidation, err)
	}
}

func TestUpdateMyProfileLanguagesTooManyAndBadFormat(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	many := make([]string, 11)
	for i := range many {
		many[i] = "en"
	}
	// duplicates checked first? both branches exist; use distinct invalid entries
	many[0], many[1] = "EN!!", "fr"
	w := domain.ProfileWrite{Languages: domain.PatchField[[]string]{Present: true, Value: &many}}
	_, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("expected %v, got %v", apperr.CodeValidation, err)
	} // too_many

	bad := []string{"english!!"}
	w = domain.ProfileWrite{Languages: domain.PatchField[[]string]{Present: true, Value: &bad}}
	_, err = svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("expected %v, got %v", apperr.CodeValidation, err)
	}

	// The contract also allows region subtags; the server rejects fr-BF with
	// 422 on purpose until profiles_languages_ck is widened (owner decision).
	regional := []string{"fr-BF"}
	w = domain.ProfileWrite{Languages: domain.PatchField[[]string]{Present: true, Value: &regional}}
	_, err = svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("fr-BF: expected %v, got %v", apperr.CodeValidation, err)
	}
	var pe *apperr.Error
	if errors.As(err, &pe) && (len(pe.Violations) != 1 || pe.Violations[0].Field != "/languages") {
		t.Fatalf("fr-BF violations %+v", pe.Violations)
	}
}

// Violation fields are JSON pointers and codes come from the contract enum
// (api/openapi.yaml Violation): the S7 form maps them, issue #116.
func TestUpdateMyProfileViolationsUseContractCodesAndPointers(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	st.tagIDs["go"] = "tag-1"

	cases := []struct {
		name  string
		write domain.ProfileWrite
		want  apperr.Violation
	}{
		{"display_name nil", domain.ProfileWrite{DisplayName: domain.PatchField[string]{Present: true, Value: nil}}, apperr.Violation{Field: "/display_name", Code: "too_short"}},
		{"display_name empty", domain.ProfileWrite{DisplayName: domain.PatchField[string]{Present: true, Value: strp("")}}, apperr.Violation{Field: "/display_name", Code: "too_short"}},
		{"display_name too long", domain.ProfileWrite{DisplayName: domain.PatchField[string]{Present: true, Value: strp(longString(81))}}, apperr.Violation{Field: "/display_name", Code: "too_long"}},
		{"display_name NUL", domain.ProfileWrite{DisplayName: domain.PatchField[string]{Present: true, Value: strp("a\x00b")}}, apperr.Violation{Field: "/display_name", Code: "invalid_value"}},
		{"bio NUL", domain.ProfileWrite{Bio: domain.PatchField[string]{Present: true, Value: strp("a\x00b")}}, apperr.Violation{Field: "/bio", Code: "invalid_value"}},
		{"bio too long", domain.ProfileWrite{Bio: domain.PatchField[string]{Present: true, Value: strp(longString(501))}}, apperr.Violation{Field: "/bio", Code: "too_long"}},
		{"intent", domain.ProfileWrite{Intent: domain.PatchField[string]{Present: true, Value: strp("bogus")}}, apperr.Violation{Field: "/intent", Code: "invalid_value"}},
		{"country format", domain.ProfileWrite{CountryCode: domain.PatchField[string]{Present: true, Value: strp("zz")}}, apperr.Violation{Field: "/country", Code: "invalid_format"}},
		{"languages too many", domain.ProfileWrite{Languages: domain.PatchField[[]string]{Present: true, Value: &[]string{"en", "fr", "en", "fr", "en", "fr", "en", "fr", "en", "fr", "en"}}}, apperr.Violation{Field: "/languages", Code: "too_many_items"}},
		{"languages format", domain.ProfileWrite{Languages: domain.PatchField[[]string]{Present: true, Value: &[]string{"english!!"}}}, apperr.Violation{Field: "/languages", Code: "invalid_format"}},
		{"languages duplicate", domain.ProfileWrite{Languages: domain.PatchField[[]string]{Present: true, Value: &[]string{"en", "en"}}}, apperr.Violation{Field: "/languages", Code: "duplicate"}},
		{"github", domain.ProfileWrite{GitHub: domain.PatchField[string]{Present: true, Value: strp("bad name!")}}, apperr.Violation{Field: "/github", Code: "invalid_format"}},
		{"visibility", domain.ProfileWrite{Visibility: domain.PatchField[domain.Visibility]{Present: true, Value: func() *domain.Visibility { v := domain.Visibility("bogus"); return &v }()}}, apperr.Violation{Field: "/profile_visibility", Code: "invalid_value"}},
		{"field visibility", domain.ProfileWrite{VisBio: domain.PatchField[domain.Visibility]{Present: true, Value: func() *domain.Visibility { v := domain.Visibility("bogus"); return &v }()}}, apperr.Violation{Field: "/field_visibility/bio", Code: "invalid_value"}},
		{"skills too many", domain.ProfileWrite{SkillSlugs: domain.PatchField[[]string]{Present: true, Value: &[]string{"go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go", "go"}}}, apperr.Violation{Field: "/skills", Code: "too_many_items"}},
		{"skills unknown", domain.ProfileWrite{SkillSlugs: domain.PatchField[[]string]{Present: true, Value: &[]string{"go", "nope"}}}, apperr.Violation{Field: "/skills", Code: "invalid_value"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.UpdateMyProfile(context.Background(), "u1", tc.write)
			var pe *apperr.Error
			if !errors.As(err, &pe) || pe.Code != apperr.CodeValidation {
				t.Fatalf("expected validation error, got %v", err)
			}
			found := false
			for _, v := range pe.Violations {
				if v.Field == tc.want.Field && v.Code == tc.want.Code {
					found = true
				}
			}
			if !found {
				t.Fatalf("want violation %+v in %+v", tc.want, pe.Violations)
			}
		})
	}
}

func TestUpdateMyProfileSkillsResolveAndRejectUnknown(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	st.tagIDs["go"] = "tag-1"
	st.tagIDs["rust"] = "tag-2"

	slugs := []string{"go", "rust"}
	w := domain.ProfileWrite{SkillSlugs: domain.PatchField[[]string]{Present: true, Value: &slugs}}
	prof, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if err != nil {
		t.Fatal(err)
	}
	if len(prof.Skills) != 2 {
		t.Fatalf("skills not applied: %+v", prof.Skills)
	}

	// unknown slug rejects the whole patch
	slugs = []string{"go", "nope"}
	w = domain.ProfileWrite{SkillSlugs: domain.PatchField[[]string]{Present: true, Value: &slugs}}
	_, err = svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("expected %v, got %v", apperr.CodeValidation, err)
	}
}

func TestUpdateMyProfileSkillsClearAndTooMany(t *testing.T) {
	svc, st, _ := testSvc(t)
	rec := seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	rec.Skills = []string{"go"}
	st.profiles["ann"] = rec
	st.tagIDs["go"] = "tag-1"

	// present + nil clears
	w := domain.ProfileWrite{SkillSlugs: domain.PatchField[[]string]{Present: true, Value: nil}}
	prof, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if err != nil {
		t.Fatal(err)
	}
	if len(prof.Skills) != 0 {
		t.Fatalf("skills not cleared: %+v", prof.Skills)
	}

	// too many
	many := make([]string, 31)
	for i := range many {
		many[i] = "go"
	}
	w = domain.ProfileWrite{SkillSlugs: domain.PatchField[[]string]{Present: true, Value: &many}}
	_, err = svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("expected %v, got %v", apperr.CodeValidation, err)
	}
}

func TestUpdateMyProfileRateLimited(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	for i := 0; i < profileWriteLimit; i++ {
		if _, err := svc.UpdateMyProfile(context.Background(), "u1", domain.ProfileWrite{
			DisplayName: domain.PatchField[string]{Present: true, Value: strp("N")},
		}); err != nil {
			t.Fatalf("hit %d: %v", i, err)
		}
	}
	_, err := svc.UpdateMyProfile(context.Background(), "u1", domain.ProfileWrite{
		DisplayName: domain.PatchField[string]{Present: true, Value: strp("N")},
	})
	if apperrCode(t, err) != apperr.CodeRateLimited {
		t.Fatalf("expected %v, got %v", apperr.CodeRateLimited, err)
	}
}

func TestUpdateMyProfileStoreErrors(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	w := domain.ProfileWrite{DisplayName: domain.PatchField[string]{Present: true, Value: strp("N")}}

	st.errs = map[string]error{"UpdateProfile": errors.New("db down")}
	_, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("expected %v, got %v", apperr.CodeInternal, err)
	}

	st.errs = map[string]error{"UpdateProfile": domain.ErrNotFound}
	_, err = svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("expected %v, got %v", apperr.CodeUnauthenticated, err)
	}

	st.errs = map[string]error{"UpdateProfile": domain.ErrCountryUnknown}
	_, err = svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("expected %v, got %v", apperr.CodeValidation, err)
	}

	st.errs = map[string]error{"ProfileByUserID": errors.New("db down")}
	_, err = svc.UpdateMyProfile(context.Background(), "u1", w)
	if apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("expected %v, got %v", apperr.CodeInternal, err)
	}

	st.errs = map[string]error{"ResolveSkillSlugs": errors.New("db down")}
	slugs := []string{"go"}
	w2 := domain.ProfileWrite{SkillSlugs: domain.PatchField[[]string]{Present: true, Value: &slugs}}
	_, err = svc.UpdateMyProfile(context.Background(), "u1", w2)
	if apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("expected %v, got %v", apperr.CodeInternal, err)
	}
}

func TestUpdateMyProfileRateLimitStoreError(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	st.errs = map[string]error{"Hit": errors.New("db down")}
	_, err := svc.UpdateMyProfile(context.Background(), "u1", domain.ProfileWrite{
		DisplayName: domain.PatchField[string]{Present: true, Value: strp("N")},
	})
	if apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("expected %v, got %v", apperr.CodeInternal, err)
	}
}

// ---------- GetProfile (C14b matrix) ----------

func TestGetProfileAnonymousPublic(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	pub, err := svc.GetProfile(context.Background(), "ann", "")
	if err != nil {
		t.Fatal(err)
	}
	if pub.Username != "ann" || pub.Bio != nil || pub.CountryCode != nil {
		t.Fatalf("unexpected public profile %+v", pub)
	}
}

func TestGetProfileUnknownIs404(t *testing.T) {
	svc, _, _ := testSvc(t)
	_, err := svc.GetProfile(context.Background(), "ghost", "")
	if apperrCode(t, err) != apperr.CodeNotFound {
		t.Fatalf("expected %v, got %v", apperr.CodeNotFound, err)
	}
}

func TestGetProfileStoreErrorIsInternal(t *testing.T) {
	svc, st, _ := testSvc(t)
	st.errs = map[string]error{"ProfileByUsername": errors.New("db down")}
	_, err := svc.GetProfile(context.Background(), "ann", "")
	if apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("expected %v, got %v", apperr.CodeInternal, err)
	}
}

func TestGetProfileNonActiveIs404(t *testing.T) {
	svc, st, _ := testSvc(t)
	rec := seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	rec.AccountStatus = "suspended"
	st.profiles["ann"] = rec
	_, err := svc.GetProfile(context.Background(), "ann", "")
	if apperrCode(t, err) != apperr.CodeNotFound {
		t.Fatalf("expected %v, got %v", apperr.CodeNotFound, err)
	}
}

func TestGetProfileVisibilityMatrix(t *testing.T) {
	cases := []struct {
		name      string
		pageVis   domain.Visibility
		caller    string // "" anonymous, "u1" owner, "u2" member
		wantFound bool
	}{
		{"anon public page", domain.VisibilityPublic, "", true},
		{"anon members page", domain.VisibilityMembers, "", false},
		{"member members page", domain.VisibilityMembers, "u2", true},
		{"owner members page", domain.VisibilityMembers, "u1", true},
		{"anon private page", domain.VisibilityPrivate, "", false},
		{"member private page", domain.VisibilityPrivate, "u2", false},
		{"owner private page", domain.VisibilityPrivate, "u1", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, st, _ := testSvc(t)
			seedProfile(st, "ann", "u1", tc.pageVis)
			_, err := svc.GetProfile(context.Background(), "ann", tc.caller)
			if tc.wantFound && err != nil {
				t.Fatalf("expected profile, got %v", err)
			}
			if !tc.wantFound {
				if apperrCode(t, err) != apperr.CodeNotFound {
					t.Fatalf("expected %v, got %v", apperr.CodeNotFound, err)
				}
			}
		})
	}
}

func TestGetProfileFieldVisibilityFiltering(t *testing.T) {
	svc, st, _ := testSvc(t)
	rec := seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	rec.Bio = strp("secret bio")
	rec.CountryCode = strp("FR")
	rec.Languages = []string{"fr"}
	rec.Skills = []string{"go"}
	rec.GitHub = strp("octocat")
	rec.OpenToCollab = true
	rec.VisBio = domain.VisibilityPrivate
	rec.VisCountry = domain.VisibilityMembers
	rec.VisLanguages = domain.VisibilityMembers
	rec.VisSkills = domain.VisibilityPrivate
	rec.VisGitHub = domain.VisibilityPublic
	st.profiles["ann"] = rec

	// anonymous: only github visible
	pub, err := svc.GetProfile(context.Background(), "ann", "")
	if err != nil {
		t.Fatal(err)
	}
	if pub.Bio != nil || pub.CountryCode != nil || pub.Languages != nil || pub.Skills != nil ||
		pub.GitHub == nil || pub.OpenToCollab == nil {
		t.Fatalf("anon filtering wrong: %+v", pub)
	}

	// member u2: country+languages+github+open_to_collab visible
	pub, err = svc.GetProfile(context.Background(), "ann", "u2")
	if err != nil {
		t.Fatal(err)
	}
	if pub.Bio != nil || pub.Skills != nil || pub.CountryCode == nil || pub.Languages == nil ||
		pub.GitHub == nil || pub.OpenToCollab == nil {
		t.Fatalf("member filtering wrong: %+v", pub)
	}

	// owner sees everything
	pub, err = svc.GetProfile(context.Background(), "ann", "u1")
	if err != nil {
		t.Fatal(err)
	}
	if pub.Bio == nil || pub.Skills == nil || pub.CountryCode == nil {
		t.Fatalf("owner filtering wrong: %+v", pub)
	}
}

func longString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func TestGetProfileVisibleNilSlicesBecomeEmpty(t *testing.T) {
	svc, st, _ := testSvc(t)
	rec := seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	rec.Languages = nil
	rec.Skills = nil
	st.profiles["ann"] = rec
	pub, err := svc.GetProfile(context.Background(), "ann", "u2")
	if err != nil {
		t.Fatal(err)
	}
	if pub.Languages == nil || pub.Skills == nil {
		t.Fatalf("expected empty non-nil slices, got %+v", pub)
	}
}

// SkillSlugs is the client's input and SkillTagIDs the resolved ids: a field
// never holds both (review of PR #94).
func TestUpdateMyProfileKeepsSlugsAndIDsInSeparateFields(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	st.tagIDs["go"] = "tag-1"
	slugs := []string{"go"}
	w := domain.ProfileWrite{SkillSlugs: domain.PatchField[[]string]{Present: true, Value: &slugs}}
	if _, err := svc.UpdateMyProfile(context.Background(), "u1", w); err != nil {
		t.Fatal(err)
	}
	got := st.lastWrite
	if !got.SkillTagIDs.Present || got.SkillTagIDs.Value == nil || len(*got.SkillTagIDs.Value) != 1 || (*got.SkillTagIDs.Value)[0] != "tag-1" {
		t.Fatalf("the store must receive the resolved ids: %+v", got.SkillTagIDs)
	}
	if !got.SkillSlugs.Present || (*got.SkillSlugs.Value)[0] != "go" {
		t.Fatalf("the slugs must stay as the client sent them: %+v", got.SkillSlugs)
	}
}
