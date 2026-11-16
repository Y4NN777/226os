package transport

import (
	"net/http"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// ListProfiles serves GET /profiles, the read-only people directory (E5):
// anonymous callers and members. The caller is the session subject; the
// service filters listability and fields for that audience. The body depends
// on the caller, so no shared cache may store it (issue #116).
func (h *Handler) ListProfiles(w http.ResponseWriter, r *http.Request, params apigen.ListProfilesParams) {
	w.Header().Set("Cache-Control", "no-store")
	req := service.DirectoryRequest{
		CallerUserID: subject.From(r.Context()).UserID, ClientIP: clientIP(r),
		Country: countryStr(params.Country), Language: params.Language, Skill: params.Skill,
		Cursor: derefStr(params.Cursor), Limit: params.Limit,
	}
	if params.Source != nil {
		req.Source = string(*params.Source)
	}
	page, err := h.svc.ListProfiles(r.Context(), req)
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	items := make([]apigen.ProfileSummary, len(page.Items))
	for i, e := range page.Items {
		items[i] = profileSummaryJSON(e)
	}
	var next *string
	if page.NextCursor != "" {
		next = &page.NextCursor
	}
	out := apigen.ProfilePage{Items: items, NextCursor: next}
	if len(page.External) > 0 {
		external := make([]apigen.ExternalDeveloper, len(page.External))
		for i, d := range page.External {
			external[i] = externalDeveloperJSON(d)
		}
		out.External = &external
	}
	writeJSON(w, http.StatusOK, out)
}

// externalDeveloperJSON maps a committers.top entry (ADR-0016); source is always "committers.top".
func externalDeveloperJSON(d domain.GitHubDeveloper) apigen.ExternalDeveloper {
	orgs := d.Organizations
	if orgs == nil {
		orgs = []string{}
	}
	return apigen.ExternalDeveloper{
		Login: d.Login, Name: d.Name, AvatarUrl: d.AvatarURL, Country: apigen.CountryCode(d.Country),
		Rank: d.Rank, Contributions: d.Contributions, Organizations: orgs, GeneratedAt: d.GeneratedAt.UTC(),
		Source: apigen.CommittersTop,
	}
}

// profileSummaryJSON maps a listed entry: a hidden or unset field stays nil
// and is omitted from the JSON, never sent as null or a placeholder.
func profileSummaryJSON(e domain.DirectoryEntry) apigen.ProfileSummary {
	out := apigen.ProfileSummary{
		Username: apigen.Username(e.Username), DisplayName: e.DisplayName,
		Country: countryPtr(e.Country), Github: e.GitHub,
	}
	if e.Languages != nil {
		out.Languages = &e.Languages
	}
	if e.Skills != nil {
		out.Skills = tagSlugsPtr(e.Skills)
	}
	return out
}
