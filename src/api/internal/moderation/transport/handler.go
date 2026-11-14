// Package transport is the HTTP edge of moderation.
package transport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/moderation/domain"
	"github.com/Y4NN777/226os/src/api/internal/moderation/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

type Handler struct {
	svc    *service.Service
	logger *slog.Logger
}

func New(svc *service.Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) HideContent(w http.ResponseWriter, r *http.Request, targetType apigen.ModerationTargetType, targetID apigen.Uuid) {
	var body apigen.ModerationReason
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	if err := h.svc.Hide(r.Context(), subject.From(r.Context()), string(targetType), targetID.String(), body.Reason, middleware.GetReqID(r.Context())); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RestoreContent(w http.ResponseWriter, r *http.Request, targetType apigen.ModerationTargetType, targetID apigen.Uuid) {
	var body apigen.ModerationReason
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	if err := h.svc.Restore(r.Context(), subject.From(r.Context()), string(targetType), targetID.String(), body.Reason, middleware.GetReqID(r.Context())); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ResolveErrorReport(w http.ResponseWriter, r *http.Request, reportID apigen.Uuid) {
	var body apigen.ErrorReportResolution
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	note := ""
	if body.Note != nil {
		note = *body.Note
	}
	report, err := h.svc.ResolveReport(r.Context(), subject.From(r.Context()), reportID.String(), string(body.Status), note, middleware.GetReqID(r.Context()))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, reportJSON(report))
}

func reportJSON(r domain.ErrorReport) apigen.ErrorReport {
	return apigen.ErrorReport{Id: uuid.MustParse(r.ID), ResourceId: uuid.MustParse(r.ResourceID), Kind: apigen.ErrorReportKind(r.Kind), Status: apigen.ErrorReportStatus(r.Status), CreatedAt: r.CreatedAt.UTC()}
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
