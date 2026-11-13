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
	st.accounts["u1"] = acct
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	view, err := svc.GetMe(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if view.Account.ID != "u1" || view.Profile.Username != "ann" {
		t.Fatalf("unexpected view %+v", view)
	}
}

func TestGetMeAccountNotFoundIsUnauthenticated(t *testing.T) {
	svc, _, _ := testSvc(t)
	_, err := svc.GetMe(context.Background(), "ghost")
	if apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("expected %v, got %v", apperr.CodeUnauthenticated, err)
	}
}

func TestGetMeAccountStoreErrorIsInternal(t *testing.T) {
	svc, st, _ := testSvc(t)
	st.errs = map[string]error{"AccountByID": errors.New("db down")}
	_, err := svc.GetMe(context.Background(), "u1")
	if apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("expected %v, got %v", apperr.CodeInternal, err)
	}
}

func TestGetMeProfileErrorIsInternal(t *testing.T) {
	svc, st, _ := testSvc(t)
	st.accounts["u1"] = domain.Account{ID: "u1", Status: "active"}
	st.errs = map[string]error{"ProfileByUserID": errors.New("db down")}
	_, err := svc.GetMe(context.Background(), "u1")
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
	if errors.As(err, &pe) && (len(pe.Violations) != 1 || pe.Violations[0].Field != "languages") {
		t.Fatalf("fr-BF violations %+v", pe.Violations)
	}
}

func TestUpdateMyProfileSkillsResolveAndRejectUnknown(t *testing.T) {
	svc, st, _ := testSvc(t)
	seedProfile(st, "ann", "u1", domain.VisibilityPublic)
	st.tagIDs["go"] = "tag-1"
	st.tagIDs["rust"] = "tag-2"

	slugs := []string{"go", "rust"}
	w := domain.ProfileWrite{SkillTagIDs: domain.PatchField[[]string]{Present: true, Value: &slugs}}
	prof, err := svc.UpdateMyProfile(context.Background(), "u1", w)
	if err != nil {
		t.Fatal(err)
	}
	if len(prof.Skills) != 2 {
		t.Fatalf("skills not applied: %+v", prof.Skills)
	}

	// unknown slug rejects the whole patch
	slugs = []string{"go", "nope"}
	w = domain.ProfileWrite{SkillTagIDs: domain.PatchField[[]string]{Present: true, Value: &slugs}}
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
	w := domain.ProfileWrite{SkillTagIDs: domain.PatchField[[]string]{Present: true, Value: nil}}
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
	w = domain.ProfileWrite{SkillTagIDs: domain.PatchField[[]string]{Present: true, Value: &many}}
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
	w2 := domain.ProfileWrite{SkillTagIDs: domain.PatchField[[]string]{Present: true, Value: &slugs}}
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
