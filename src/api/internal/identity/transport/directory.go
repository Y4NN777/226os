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
	writeJSON(w, http.StatusOK, apigen.ProfilePage{Items: items, NextCursor: next})
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
