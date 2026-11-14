// Package transport is the HTTP edge of Build: decode, call one service method, encode.
package transport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/build/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// Handler serves the P0 project and contribution-need operations.
type Handler struct {
	svc    *service.Service
	logger *slog.Logger
}

// New returns a handler. logger may be nil.
func New(svc *service.Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Handler{svc: svc, logger: logger}
}

// ListProjects serves GET /projects.
func (h *Handler) ListProjects(w http.ResponseWriter, r *http.Request, params apigen.ListProjectsParams) {
	page, err := h.svc.ListProjects(r.Context(), subject.From(r.Context()), service.ProjectListRequest{
		Q: params.Q, Technology: params.Technology, Skill: params.Skill, OpenNeeds: params.OpenNeeds,
		Cursor: deref(params.Cursor), Limit: params.Limit,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	items := make([]apigen.Project, len(page.Items))
	for i, p := range page.Items {
		items[i] = projectJSON(p)
	}
	writeJSON(w, http.StatusOK, apigen.ProjectPage{Items: items, NextCursor: next(page.NextCursor)})
}

// CreateProject serves POST /projects. The owner is the session subject; Idempotency-Key is handled
// by the router middleware.
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request, _ apigen.CreateProjectParams) {
	var body apigen.ProjectCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	p, err := h.svc.CreateProject(r.Context(), subject.From(r.Context()), service.CreateProjectRequest{
		GitHubURL: body.GithubUrl, Lang: string(body.Lang), Description: body.Description,
		Visibility: (*string)(body.Visibility), Technologies: slice(body.Technologies), Links: links(body.Links),
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/projects/"+p.ID)
	writeJSON(w, http.StatusCreated, projectJSON(p))
}

// GetProject serves GET /projects/{projectId}.
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId) {
	p, err := h.svc.GetProject(r.Context(), subject.From(r.Context()), projectID.String())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, projectJSON(p))
}

// UpdateProject serves PATCH /projects/{projectId} (JSON Merge Patch).
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId) {
	var body apigen.ProjectUpdate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	p, err := h.svc.UpdateProject(r.Context(), subject.From(r.Context()), projectID.String(), service.UpdateProjectRequest{
		Description: body.Description, Visibility: (*string)(body.Visibility), Technologies: body.Technologies,
		Links: links(body.Links),
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, projectJSON(p))
}

// ListContributionNeeds serves GET /projects/{projectId}/contribution-needs.
func (h *Handler) ListContributionNeeds(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId,
	params apigen.ListContributionNeedsParams,
) {
	page, err := h.svc.ListNeeds(r.Context(), subject.From(r.Context()), projectID.String(), service.NeedListRequest{
		Cursor: deref(params.Cursor), Limit: params.Limit,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	items := make([]apigen.ContributionNeed, len(page.Items))
	for i, n := range page.Items {
		items[i] = needJSON(n)
	}
	httpx.ViewerScoped(w)
	writeJSON(w, http.StatusOK, apigen.NeedPage{Items: items, NextCursor: next(page.NextCursor)})
}

// CreateContributionNeed serves POST /projects/{projectId}/contribution-needs; Idempotency-Key is
// handled by the router middleware.
func (h *Handler) CreateContributionNeed(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId,
	_ apigen.CreateContributionNeedParams,
) {
	var body apigen.NeedCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	n, err := h.svc.CreateNeed(r.Context(), subject.From(r.Context()), projectID.String(), service.CreateNeedRequest{
		Title: body.Title, Description: body.Description, Lang: string(body.Lang), Skills: body.Skills,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/projects/"+projectID.String()+"/contribution-needs")
	writeJSON(w, http.StatusCreated, needJSON(n))
}

func projectJSON(p domain.Project) apigen.Project {
	related := make([]apigen.RelatedLink, len(p.Related))
	for i, l := range p.Related {
		lang := apigen.Lang(l.Lang)
		related[i] = apigen.RelatedLink{Id: mustID(l.ID), Type: apigen.RelatedLinkType(l.Type), Title: l.Title, Lang: &lang}
	}
	return apigen.Project{
		Id: mustID(p.ID), Name: p.Name, GithubUrl: p.GitHubURL, Description: p.Description, Lang: apigen.Lang(p.Lang),
		Technologies: p.Technologies, Visibility: apigen.ProjectVisibility(p.Visibility),
		Owner:          apigen.AuthorRef{Id: mustID(p.Owner.ID), Username: p.Owner.Username, DisplayName: p.Owner.DisplayName},
		OpenNeedsCount: p.OpenNeeds, Related: related, CreatedAt: p.CreatedAt.UTC(), UpdatedAt: p.UpdatedAt.UTC(),
	}
}

func needJSON(n domain.Need) apigen.ContributionNeed {
	out := apigen.ContributionNeed{
		Id: mustID(n.ID), ProjectId: mustID(n.ProjectID), Title: n.Title, Description: n.Description,
		Lang: apigen.Lang(n.Lang), Skills: n.Skills, Status: apigen.NeedStatus(n.Status), CreatedAt: n.CreatedAt.UTC(),
	}
	if n.ViewerApplication != nil {
		out.ViewerApplication = &apigen.ViewerApplication{
			Id: mustID(n.ViewerApplication.ID), Status: apigen.ApplicationStatus(n.ViewerApplication.Status),
		}
	}
	return out
}

func links(in *[]apigen.LinkRef) []domain.LinkRef {
	if in == nil {
		return nil
	}
	out := make([]domain.LinkRef, len(*in))
	for i, l := range *in {
		out[i] = domain.LinkRef{Type: string(l.Type), ID: l.Id.String()}
	}
	return out
}

func slice(in *[]string) []string {
	if in == nil {
		return nil
	}
	return *in
}

// mustID parses an id the repository formatted from a uuid column; it cannot fail.
func mustID(s string) apigen.Uuid { return uuid.MustParse(s) }

func next(c string) *string {
	if c == "" {
		return nil
	}
	return &c
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		ae = apperr.Wrap(apperr.CodeInternal, err)
	}
	httpx.WriteError(h.logger)(w, r, ae)
}
