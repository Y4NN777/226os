package transport

import (
	"net/http"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/build/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// CreateApplication serves POST /contribution-needs/{needId}/applications. The applicant is the session
// subject; Idempotency-Key is handled by the router middleware.
func (h *Handler) CreateApplication(w http.ResponseWriter, r *http.Request, needID apigen.NeedId, _ apigen.CreateApplicationParams) {
	var body apigen.ApplicationCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	a, err := h.svc.CreateApplication(r.Context(), subject.From(r.Context()), needID.String(), service.CreateApplicationRequest{Message: body.Message})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/applications/"+a.ID)
	writeJSON(w, http.StatusCreated, applicationJSON(a))
}

// ListProjectApplications serves GET /projects/{projectId}/applications (project owner only).
func (h *Handler) ListProjectApplications(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId,
	params apigen.ListProjectApplicationsParams,
) {
	page, err := h.svc.ListApplications(r.Context(), subject.From(r.Context()), projectID.String(), service.ApplicationListRequest{
		Status: (*string)(params.Status), Cursor: deref(params.Cursor), Limit: params.Limit,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	items := make([]apigen.Application, len(page.Items))
	for i, a := range page.Items {
		items[i] = applicationJSON(a)
	}
	writeJSON(w, http.StatusOK, apigen.ApplicationPage{Items: items, NextCursor: next(page.NextCursor)})
}

// DecideApplication serves PATCH /applications/{applicationId}. The note is accepted but not stored
// (applications has no column for it at P0).
func (h *Handler) DecideApplication(w http.ResponseWriter, r *http.Request, applicationID apigen.ApplicationId) {
	var body apigen.ApplicationDecision
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	a, err := h.svc.DecideApplication(r.Context(), subject.From(r.Context()), applicationID.String(), string(body.Status))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, applicationJSON(a))
}

func applicationJSON(a domain.Application) apigen.Application {
	out := apigen.Application{
		Id: mustID(a.ID), NeedId: mustID(a.NeedID), ProjectId: mustID(a.ProjectID), Message: a.Message,
		Status: apigen.ApplicationStatus(a.Status), CreatedAt: a.CreatedAt.UTC(),
		Applicant: apigen.AuthorRef{Id: mustID(a.Applicant.ID), Username: a.Applicant.Username, DisplayName: a.Applicant.DisplayName},
	}
	if a.DecidedAt != nil {
		at := a.DecidedAt.UTC()
		out.DecidedAt = &at
	}
	return out
}
