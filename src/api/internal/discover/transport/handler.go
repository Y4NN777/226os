// Package transport is the HTTP edge of the catalogue reads.
package transport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/discover/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// Handler serves the three public catalogue reads and the member error report.
type Handler struct {
	svc     *service.Service
	reports *service.Reports
	logger  *slog.Logger
}

// New returns a handler. logger may be nil.
func New(svc *service.Service, reports *service.Reports, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Handler{svc: svc, reports: reports, logger: logger}
}

// ListResources serves GET /resources.
func (h *Handler) ListResources(w http.ResponseWriter, r *http.Request, params apigen.ListResourcesParams) {
	sub := subject.From(r.Context())
	page, err := h.svc.List(r.Context(), service.ListRequest{
		Q: params.Q, Countries: derefList(params.Country), Category: params.Category, Type: defined(params.Type),
		SDK: params.Sdk, Provider: params.Provider, Doclang: lang(params.Doclang),
		Cursor: deref(params.Cursor), Limit: params.Limit, Anonymous: sub.Anonymous(), ClientIP: httpx.ClientIPKey(r),
	})
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	items := make([]apigen.Resource, len(page.Items))
	for i, item := range page.Items {
		body, err := resourceJSON(item)
		if err != nil {
			writeErr(h.logger, w, r, err)
			return
		}
		items[i] = body
	}
	var next *string
	if page.NextCursor != "" {
		next = &page.NextCursor
	}
	writeJSON(w, apigen.ResourcePage{Items: items, NextCursor: next})
}

// GetResource serves GET /resources/{resourceId}.
func (h *Handler) GetResource(w http.ResponseWriter, r *http.Request, resourceID apigen.ResourceId) {
	item, err := h.svc.Get(r.Context(), resourceID.String())
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	body, err := resourceJSON(item)
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	writeJSON(w, body)
}

// ListResourceSources serves GET /resources/{resourceId}/sources.
func (h *Handler) ListResourceSources(w http.ResponseWriter, r *http.Request, resourceID apigen.ResourceId) {
	rows, err := h.svc.Sources(r.Context(), resourceID.String())
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	items := make([]apigen.ResourceSource, len(rows))
	for i, row := range rows {
		var title *string
		if row.Title != "" {
			title = &row.Title
		}
		id, err := parseID(row.ID)
		if err != nil {
			writeErr(h.logger, w, r, err)
			return
		}
		items[i] = apigen.ResourceSource{
			Id: id, Kind: apigen.ResourceSourceKind(row.Kind), Status: apigen.VerificationStatus(row.Status),
			Title: title, Url: row.URL, VerifiedAt: row.VerifiedAt,
		}
	}
	writeJSON(w, apigen.ResourceSourceList{Items: items})
}

func resourceJSON(item domain.Resource) (apigen.Resource, error) {
	langs := make([]apigen.Lang, len(item.DocLanguages))
	for i, lang := range item.DocLanguages {
		langs[i] = apigen.Lang(lang)
	}
	sdks := make([]struct {
		Language apigen.Slug `json:"language"`
		Name     *string     `json:"name,omitempty"`
		Url      string      `json:"url"`
	}, len(item.SDKs))
	for i, sdk := range item.SDKs {
		var name *string
		if sdk.Name != "" {
			name = &sdk.Name
		}
		sdks[i].Language = sdk.Language
		sdks[i].Name = name
		sdks[i].Url = sdk.URL
	}
	related := make([]apigen.RelatedLink, len(item.Related))
	for i, link := range item.Related {
		var lang *apigen.Lang
		if link.Lang != "" {
			l := apigen.Lang(link.Lang)
			lang = &l
		}
		id, err := parseID(link.ID)
		if err != nil {
			return apigen.Resource{}, err
		}
		related[i] = apigen.RelatedLink{Id: id, Lang: lang, Title: link.Title, Type: apigen.RelatedLinkType(link.Type)}
	}
	id, err := parseID(item.ID)
	if err != nil {
		return apigen.Resource{}, err
	}
	out := apigen.Resource{
		Category: item.Category, Countries: item.Countries, DocLanguages: langs, DocsUrl: item.DocsURL,
		Id: id, LastVerifiedAt: item.LastVerifiedAt, Name: item.Name, Related: related, SdkLanguages: sdks,
		Provenance: apigen.ResourceProvenance(item.Provenance), Slug: item.Slug, SourceCount: item.SourceCount, Summary: item.Summary, Type: apigen.ResourceType(item.Type),
		UpdatedAt: item.UpdatedAt, VerificationStatus: apigen.VerificationStatus(item.VerificationStatus),
	}
	out.Provider.Name = item.Provider.Name
	out.Provider.Slug = item.Provider.Slug
	out.Provider.Url = item.Provider.URL
	return out, nil
}

func parseID(s string) (apigen.Uuid, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return apigen.Uuid{}, err
	}
	return id, nil
}

func defined(v *apigen.ResourceType) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func lang(v *apigen.Lang) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func writeJSON(w http.ResponseWriter, body any) { writeStatusJSON(w, http.StatusOK, body) }

func writeStatusJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		httpx.WriteError(logger)(w, r, ae)
		return
	}
	if errors.Is(err, domain.ErrNotFound) {
		httpx.WriteError(logger)(w, r, apperr.New(apperr.CodeNotFound, "Resource not found."))
		return
	}
	httpx.WriteError(logger)(w, r, apperr.Wrap(apperr.CodeInternal, err))
}

func derefList(v *[]string) []string {
	if v == nil {
		return nil
	}
	return *v
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
