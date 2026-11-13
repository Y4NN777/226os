// Package transport is the HTTP edge of the search read.
package transport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
	"github.com/Y4NN777/226os/src/api/internal/search/domain"
	"github.com/Y4NN777/226os/src/api/internal/search/service"
)

// Handler serves GET /search.
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

// Search serves GET /search.
func (h *Handler) Search(w http.ResponseWriter, r *http.Request, params apigen.SearchParams) {
	req := service.Request{
		Q: params.Q, PerGroup: params.PerGroup, Country: params.Country,
		Anonymous: subject.From(r.Context()).Anonymous(), Client: httpx.ClientIPKey(r),
	}
	if params.Types != nil {
		for _, t := range *params.Types {
			req.Types = append(req.Types, string(t))
		}
	}
	if params.Lang != nil {
		lang := string(*params.Lang)
		req.Lang = &lang
	}
	results, err := h.svc.Search(r.Context(), req)
	if err != nil {
		writeErr(h.logger, w, r, err)
		return
	}
	body := apigen.SearchResults{Query: results.Query}
	for group, result := range results.Groups {
		out, err := groupJSON(result)
		if err != nil {
			writeErr(h.logger, w, r, err)
			return
		}
		switch group {
		case domain.Questions:
			body.Groups.Questions = &out
		case domain.Resources:
			body.Groups.Resources = &out
		case domain.Projects:
			body.Groups.Projects = &out
		case domain.Learn:
			body.Groups.Learn = &out
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(body)
}

func groupJSON(result service.GroupResult) (apigen.SearchGroupResult, error) {
	items := make([]apigen.SearchHit, len(result.Items))
	for i, hit := range result.Items {
		id, err := uuid.Parse(hit.ID)
		if err != nil {
			return apigen.SearchGroupResult{}, err
		}
		items[i] = apigen.SearchHit{Id: id, Type: apigen.SearchGroup(hit.Type), Title: hit.Title, Snippet: hit.Snippet}
		if hit.Lang != "" {
			lang := apigen.Lang(hit.Lang)
			items[i].Lang = &lang
		}
	}
	return apigen.SearchGroupResult{Items: items, HasMore: result.HasMore}, nil
}

func writeErr(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		httpx.WriteError(logger)(w, r, ae)
		return
	}
	httpx.WriteError(logger)(w, r, apperr.Wrap(apperr.CodeInternal, err))
}
