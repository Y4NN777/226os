package service

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

// Profile-store port additions live on Store (service.go); this file holds the
// S1 profile use cases: getMe, updateMyProfile (JSON Merge Patch) and
// getProfile with the C14b visibility matrix.

const (
	profileWriteLimit  = 60 // authenticated writes per user per minute (ADR-0003)
	profileWriteWindow = time.Minute

	maxDisplayNameRunes = 80
	maxBioRunes         = 500
	maxLanguages        = 10
	maxSkills           = 30
)

var (
	// The contract also accepts region subtags (fr-BF); the server rejects
	// them with 422 on purpose until profiles_languages_ck is widened.
	languagePattern  = regexp.MustCompile(`^[a-z]{2,3}$`)
	githubPattern    = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,37}[A-Za-z0-9])?$`)
	countryPattern   = regexp.MustCompile(`^[A-Z]{2}$`)
	skillSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.+#-]*$`)
)

var validIntents = map[string]bool{
	"ask": true, "help": true, "integrate": true, "contribute": true, "learn": true,
}

var validVisibilities = map[string]bool{
	string(domain.VisibilityPublic): true, string(domain.VisibilityMembers): true,
	string(domain.VisibilityPrivate): true,
}

// MeView is the getMe payload: the account (the only place email is returned)
// plus the caller's own unfiltered profile.
type MeView struct {
	Account domain.Account
	Profile domain.Profile
}

// GetMe returns the caller's account and own profile. The account comes from
// the live session the middleware already resolved (acct), so it is read once
// per request; only the profile is loaded here. The transport rejects
// anonymous callers first.
func (s *Service) GetMe(ctx context.Context, acct domain.Account) (MeView, error) {
	prof, err := s.store.ProfileByUserID(ctx, acct.ID)
	if err != nil {
		return MeView{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	return MeView{Account: acct, Profile: prof}, nil
}

// UpdateMyProfile applies a resolved merge patch to the caller's own profile
// and returns the fresh profile. The patch arrives already tri-state
// (domain.PatchField); this method owns validation policy and skill
// resolution, the store owns the atomic write.
func (s *Service) UpdateMyProfile(ctx context.Context, userID string, w domain.ProfileWrite) (domain.Profile, error) {
	if err := s.allow(ctx, "profile:write", userID, profileWriteLimit, profileWriteWindow); err != nil {
		return domain.Profile{}, err
	}
	if err := validateProfileWrite(&w); err != nil {
		return domain.Profile{}, err
	}
	if w.SkillSlugs.Present {
		if err := s.resolveSkills(ctx, &w); err != nil {
			return domain.Profile{}, err
		}
	}
	deriveBioLanguage(&w)
	if err := s.store.UpdateProfile(ctx, userID, w); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.Profile{}, unauthenticated()
		}
		if errors.Is(err, domain.ErrCountryUnknown) {
			return domain.Profile{}, apperr.New(apperr.CodeValidation, "country is not a known country code.").
				WithViolations(apperr.Violation{Field: "/country", Code: "invalid_value"})
		}
		return domain.Profile{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	prof, err := s.store.ProfileByUserID(ctx, userID)
	if err != nil {
		return domain.Profile{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	return prof, nil
}

// GetProfile serves GET /profiles/{username} under the C14b matrix: the page
// opens only when profile_visibility allows the caller and the account is
// active; every other case — unknown username, deleted or suspended account,
// hidden profile — is the same 404. Field-level visibility then filters the
// payload; the owner sees everything.
func (s *Service) GetProfile(ctx context.Context, username, callerUserID string) (domain.PublicProfile, error) {
	rec, err := s.store.ProfileByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.PublicProfile{}, notFound()
		}
		return domain.PublicProfile{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	anonymous := callerUserID == ""
	owner := !anonymous && callerUserID == rec.UserID
	if rec.AccountStatus != "active" || !rec.Visibility.Allows(anonymous, owner) {
		return domain.PublicProfile{}, notFound()
	}
	out := domain.PublicProfile{
		Username: rec.Username, DisplayName: rec.DisplayName,
		Visibility: rec.Visibility, JoinedAt: rec.JoinedAt,
	}
	if rec.VisBio.Allows(anonymous, owner) {
		out.Bio = rec.Bio
	}
	if rec.VisCountry.Allows(anonymous, owner) {
		out.CountryCode = rec.CountryCode
	}
	if rec.VisLanguages.Allows(anonymous, owner) {
		out.Languages = nonNil(rec.Languages)
	}
	if rec.VisSkills.Allows(anonymous, owner) {
		out.Skills = nonNil(rec.Skills)
	}
	if rec.VisGitHub.Allows(anonymous, owner) {
		out.GitHub = rec.GitHub
	}
	if rec.VisOpenCollab.Allows(anonymous, owner) {
		v := rec.OpenToCollab
		out.OpenToCollab = &v
	}
	return out, nil
}

// ---------- validation ----------

func validateProfileWrite(w *domain.ProfileWrite) error {
	if !w.DisplayName.Present && !w.Intent.Present && !w.CountryCode.Present &&
		!w.Languages.Present && !w.GitHub.Present && !w.Bio.Present &&
		!w.OpenToCollab.Present && !w.VisCountry.Present && !w.VisLanguages.Present &&
		!w.VisSkills.Present && !w.VisGitHub.Present && !w.VisBio.Present &&
		!w.Visibility.Present && !w.SkillSlugs.Present {
		return apperr.New(apperr.CodeValidation, "The patch must change at least one field.")
	}
	violations := []apperr.Violation{}
	// field is the JSON pointer of the request member and code one of the
	// Violation.code enum values (api/openapi.yaml): the S7 form maps them.
	add := func(field, code string) { violations = append(violations, apperr.Violation{Field: field, Code: code}) }

	if w.DisplayName.Present {
		if w.DisplayName.Value == nil {
			add("/display_name", "too_short")
		} else if !validText(*w.DisplayName.Value) {
			add("/display_name", "invalid_value")
		} else if n := utf8.RuneCountInString(*w.DisplayName.Value); n < 1 {
			add("/display_name", "too_short")
		} else if n > maxDisplayNameRunes {
			add("/display_name", "too_long")
		}
	}
	if w.Intent.Present && w.Intent.Value != nil && !validIntents[*w.Intent.Value] {
		add("/intent", "invalid_value")
	}
	if w.CountryCode.Present && w.CountryCode.Value != nil && !countryPattern.MatchString(*w.CountryCode.Value) {
		add("/country", "invalid_format")
	}
	if w.Languages.Present && w.Languages.Value != nil {
		langs := *w.Languages.Value
		if len(langs) > maxLanguages {
			add("/languages", "too_many_items")
		}
		seen := map[string]bool{}
		for _, l := range langs {
			if !languagePattern.MatchString(l) {
				add("/languages", "invalid_format")
				break
			}
			if seen[l] {
				add("/languages", "duplicate")
				break
			}
			seen[l] = true
		}
	}
	if w.GitHub.Present && w.GitHub.Value != nil && !githubPattern.MatchString(*w.GitHub.Value) {
		add("/github", "invalid_format")
	}
	if w.Bio.Present && w.Bio.Value != nil {
		if !validText(*w.Bio.Value) {
			add("/bio", "invalid_value")
		} else if utf8.RuneCountInString(*w.Bio.Value) > maxBioRunes {
			add("/bio", "too_long")
		}
	}
	for _, f := range []struct {
		name string
		v    domain.PatchField[domain.Visibility]
	}{
		{"/field_visibility/country", w.VisCountry}, {"/field_visibility/languages", w.VisLanguages},
		{"/field_visibility/skills", w.VisSkills}, {"/field_visibility/github", w.VisGitHub},
		{"/field_visibility/bio", w.VisBio}, {"/profile_visibility", w.Visibility},
	} {
		if f.v.Present && f.v.Value != nil && !validVisibilities[string(*f.v.Value)] {
			add(f.name, "invalid_value")
		}
	}
	if w.SkillSlugs.Present && w.SkillSlugs.Value != nil && len(*w.SkillSlugs.Value) > maxSkills {
		add("/skills", "too_many_items")
	}
	if len(violations) > 0 {
		return apperr.New(apperr.CodeValidation, "The profile patch is invalid.").WithViolations(violations...)
	}
	return nil
}

// resolveSkills maps the patch's skill slugs to active tag ids. Any slug that
// does not resolve rejects the whole patch: a partial skill set would silently
// drop the owner's choice.
func (s *Service) resolveSkills(ctx context.Context, w *domain.ProfileWrite) error {
	if w.SkillSlugs.Value == nil || *w.SkillSlugs.Value == nil {
		empty := []string{}
		w.SkillTagIDs = domain.PatchField[[]string]{Present: true, Value: &empty} // present + nil: clear every skill
		return nil
	}
	slugs := *w.SkillSlugs.Value
	ids, err := s.store.ResolveSkillSlugs(ctx, slugs)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	if len(ids) != len(uniqueStrings(slugs)) {
		return apperr.New(apperr.CodeValidation, "One or more skills are unknown.").
			WithViolations(apperr.Violation{Field: "/skills", Code: "invalid_value"})
	}
	w.SkillTagIDs = domain.PatchField[[]string]{Present: true, Value: &ids}
	return nil
}

// deriveBioLanguage keeps the bio/bio_language pair consistent with the CHECK
// constraint: setting bio stores 'en' (contract exposes no language field),
// clearing bio clears the language, untouched bio leaves both alone.
func deriveBioLanguage(w *domain.ProfileWrite) {
	if !w.Bio.Present {
		return
	}
	if w.Bio.Value == nil {
		w.BioLanguage = domain.PatchField[string]{Present: true, Value: nil}
		return
	}
	lang := "en"
	w.BioLanguage = domain.PatchField[string]{Present: true, Value: &lang}
}

// validText refuses what PostgreSQL text cannot hold (NUL, invalid UTF-8):
// the write answers 422 instead of a 22021 constraint failure mapped to 500.
func validText(s string) bool { return utf8.ValidString(s) && !strings.ContainsRune(s, 0) }

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func uniqueStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}
