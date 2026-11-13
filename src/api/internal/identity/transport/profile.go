package transport

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// maxPatchBodyBytes matches httpx.DecodeJSON's 1 MiB cap for the pre-read the
// merge patch needs (typed decode plus key-presence detection).
const maxPatchBodyBytes = 1 << 20

// GetMe serves GET /me: member only, the single place the account email is
// returned (contract x-authz note).
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	caller := subject.From(r.Context())
	if caller.Anonymous() {
		writeErr(h.logger, w, r, apperr.New(apperr.CodeUnauthenticated, "Authentication required."))
		return
	}
	view, err := h.svc.GetMe(r.Context(), caller.UserID)
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	acct, err := accountJSON(view.Account)
	if err != nil {
		writeErr(h.logger, w, r, apperr.Wrap(apperr.CodeInternal, err))
		return
	}
	writeJSON(w, http.StatusOK, apigen.Me{Account: acct, Profile: myProfileJSON(view.Profile)})
}

// UpdateMyProfile serves PATCH /me/profile: member only, JSON Merge Patch.
// The OpenAPI validator has already enforced the schema; this handler adds
// key-presence detection because the generated *T fields cannot distinguish
// an absent member from an explicit JSON null (which clears the column).
func (h *Handler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	caller := subject.From(r.Context())
	if caller.Anonymous() {
		writeErr(h.logger, w, r, apperr.New(apperr.CodeUnauthenticated, "Authentication required."))
		return
	}
	body, present, err := decodeProfilePatch(w, r)
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	prof, err := h.svc.UpdateMyProfile(r.Context(), caller.UserID, profileWriteFrom(body, present))
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, myProfileJSON(prof))
}

// GetProfile serves GET /profiles/{username}: anonymous and members allowed;
// the service applies the C14b visibility matrix and answers an identical 404
// for hidden, unknown and non-active profiles.
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request, username apigen.Username) {
	caller := subject.From(r.Context())
	pub, err := h.svc.GetProfile(r.Context(), string(username), caller.UserID)
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, publicProfileJSON(pub))
}

// ---------- merge patch decoding ----------

// decodeProfilePatch strictly decodes the body into apigen.ProfileUpdate and
// returns the set of top-level keys that were present, so nullable members can
// tell "absent" (unchanged) from "null" (clear).
func decodeProfilePatch(w http.ResponseWriter, r *http.Request) (apigen.ProfileUpdate, map[string]json.RawMessage, error) {
	var body apigen.ProfileUpdate
	r.Body = http.MaxBytesReader(w, r.Body, maxPatchBodyBytes)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return body, nil, apperr.New(apperr.CodePayloadTooLarge, "The request body is too large.")
		}
		return body, nil, apperr.New(apperr.CodeMalformedRequest, "The request body is not valid JSON.")
	}
	// Presence first: this rejects non-object JSON the typed decode would
	// also reject, and gives the merge patch its absent-vs-null signal.
	present := map[string]json.RawMessage{}
	if err := json.Unmarshal(data, &present); err != nil {
		return body, nil, apperr.New(apperr.CodeMalformedRequest, "The request body is not valid JSON.")
	}
	r2 := r.Clone(r.Context())
	r2.Body = io.NopCloser(bytes.NewReader(data))
	if err := httpx.DecodeJSON(w, r2, &body); err != nil {
		return body, nil, err
	}
	return body, present, nil
}

// profileWriteFrom converts the decoded body plus key presence into the
// tri-state domain patch the service validates and the store applies.
func profileWriteFrom(b apigen.ProfileUpdate, present map[string]json.RawMessage) domain.ProfileWrite {
	var w domain.ProfileWrite
	setStr := func(dst *domain.PatchField[string], key string, v *string) {
		if _, ok := present[key]; !ok {
			return
		}
		f := domain.PatchField[string]{Present: true}
		if isJSONNull(present[key]) {
			f.Value = nil
		} else if v != nil {
			f.Value = v
		}
		*dst = f
	}
	setStr(&w.DisplayName, "display_name", b.DisplayName)
	setStr(&w.Intent, "intent", intentStr(b.Intent))
	setStr(&w.CountryCode, "country", countryStr(b.Country))
	setStr(&w.GitHub, "github", b.Github)
	setStr(&w.Bio, "bio", b.Bio)

	if _, ok := present["languages"]; ok && b.Languages != nil {
		w.Languages = domain.PatchField[[]string]{Present: true, Value: b.Languages}
	}
	if _, ok := present["skills"]; ok && b.Skills != nil {
		slugs := make([]string, len(*b.Skills))
		for i, s := range *b.Skills {
			slugs[i] = string(s)
		}
		w.SkillTagIDs = domain.PatchField[[]string]{Present: true, Value: &slugs}
	}
	if b.OpenToCollab != nil {
		w.OpenToCollab = domain.PatchField[bool]{Present: true, Value: b.OpenToCollab}
	}
	if b.ProfileVisibility != nil {
		v := domain.Visibility(*b.ProfileVisibility)
		w.Visibility = domain.PatchField[domain.Visibility]{Present: true, Value: &v}
	}
	if b.FieldVisibility != nil {
		fv := b.FieldVisibility
		setVis := func(dst *domain.PatchField[domain.Visibility], v *apigen.FieldVisibility) {
			if v != nil {
				d := domain.Visibility(*v)
				*dst = domain.PatchField[domain.Visibility]{Present: true, Value: &d}
			}
		}
		setVis(&w.VisBio, fv.Bio)
		setVis(&w.VisCountry, fv.Country)
		setVis(&w.VisLanguages, fv.Languages)
		setVis(&w.VisSkills, fv.Skills)
		setVis(&w.VisGitHub, fv.Github)
	}
	return w
}

func isJSONNull(raw json.RawMessage) bool { return bytes.Equal(bytes.TrimSpace(raw), []byte("null")) }

func intentStr(v *apigen.Intent) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func countryStr(v *apigen.CountryCode) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

// ---------- response mapping ----------

func myProfileJSON(p domain.Profile) apigen.MyProfile {
	out := apigen.MyProfile{
		Username: apigen.Username(p.Username), DisplayName: p.DisplayName,
		Intent:            intentPtr(p.Intent),
		Skills:            tagSlugs(p.Skills),
		Country:           countryPtr(p.CountryCode),
		Languages:         nonNilSlice(p.Languages),
		Github:            p.GitHub,
		Bio:               p.Bio,
		OpenToCollab:      p.OpenToCollab,
		ProfileVisibility: apigen.ProfileVisibility(p.Visibility),
		ChecklistMissing:  checklistMissing(p),
	}
	out.UpdatedAt = &p.UpdatedAt
	out.FieldVisibility.Bio = visPtr(p.VisBio)
	out.FieldVisibility.Country = visPtr(p.VisCountry)
	out.FieldVisibility.Languages = visPtr(p.VisLanguages)
	out.FieldVisibility.Skills = visPtr(p.VisSkills)
	out.FieldVisibility.Github = visPtr(p.VisGitHub)
	return out
}

func publicProfileJSON(p domain.PublicProfile) apigen.PublicProfile {
	out := apigen.PublicProfile{
		Username: apigen.Username(p.Username), DisplayName: p.DisplayName,
		ProfileVisibility: apigen.ProfileVisibility(p.Visibility),
		JoinedAt:          p.JoinedAt,
	}
	if p.Bio != nil {
		out.Bio = p.Bio
	}
	if p.CountryCode != nil {
		out.Country = countryPtr(p.CountryCode)
	}
	if p.Languages != nil {
		out.Languages = &p.Languages
	}
	if p.Skills != nil {
		out.Skills = tagSlugsPtr(p.Skills)
	}
	if p.GitHub != nil {
		out.Github = p.GitHub
	}
	if p.OpenToCollab != nil {
		out.OpenToCollab = p.OpenToCollab
	}
	return out
}

func intentPtr(s *string) *apigen.Intent {
	if s == nil {
		return nil
	}
	v := apigen.Intent(*s)
	return &v
}

func countryPtr(s *string) *apigen.CountryCode {
	if s == nil {
		return nil
	}
	v := apigen.CountryCode(*s)
	return &v
}

func visPtr(v domain.Visibility) *apigen.FieldVisibility {
	f := apigen.FieldVisibility(v)
	return &f
}

func tagSlugs(in []string) []apigen.TagSlug {
	out := make([]apigen.TagSlug, len(in))
	for i, s := range in {
		out[i] = apigen.TagSlug(s)
	}
	return out
}

func tagSlugsPtr(in []string) *[]apigen.TagSlug {
	out := tagSlugs(in)
	return &out
}

func nonNilSlice(in []string) []string {
	if in == nil {
		return []string{}
	}
	return in
}

// checklistMissing derives the read-only onboarding checklist: a field counts
// as missing while it holds its zero value.
func checklistMissing(p domain.Profile) *[]apigen.ProfileField {
	missing := []apigen.ProfileField{}
	if p.Intent == nil {
		missing = append(missing, apigen.ProfileFieldIntent)
	}
	if len(p.Skills) == 0 {
		missing = append(missing, apigen.ProfileFieldSkills)
	}
	if p.CountryCode == nil {
		missing = append(missing, apigen.ProfileFieldCountry)
	}
	if len(p.Languages) == 0 {
		missing = append(missing, apigen.ProfileFieldLanguages)
	}
	if p.GitHub == nil {
		missing = append(missing, apigen.ProfileFieldGithub)
	}
	if p.Bio == nil {
		missing = append(missing, apigen.ProfileFieldBio)
	}
	if !p.OpenToCollab {
		missing = append(missing, apigen.ProfileFieldOpenToCollab)
	}
	return &missing
}
