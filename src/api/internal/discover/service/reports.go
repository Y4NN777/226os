package service

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	// Members may file 60 reports per minute (ADR-0003, authenticated writes).
	reportScope  = "error_report:user"
	reportLimit  = 60
	reportWindow = time.Minute
)

// ReportStore saves an error report. The service owns this port.
//
// CreateReport is idempotent on an open report: when the reporter already has an open (received or
// triaged) report of the same kind for the resource, it returns that report and creates no row.
// It returns domain.ErrNotFound when the resource is not published with a published provider and
// domain.ErrSourceMismatch when the source belongs to another resource.
type ReportStore interface {
	CreateReport(ctx context.Context, in domain.NewReport) (domain.ErrorReport, error)
}

// Limiter is the fixed-window rate limiter. The service owns this port.
type Limiter interface {
	Hit(ctx context.Context, scope, value string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)
}

// ReportRequest is the body of createResourceErrorReport. SourceID is empty when not given.
type ReportRequest struct {
	ResourceID, Kind, Body, SourceID string
}

// Reports is the error report use case.
type Reports struct {
	store   ReportStore
	limiter Limiter
}

// NewReports returns the error report service.
func NewReports(store ReportStore, limiter Limiter) *Reports {
	return &Reports{store: store, limiter: limiter}
}

// Create files an error report for the session subject: a member with a verified email, at most 60
// per minute. The reporter always comes from the subject, never from the request. The limiter fails
// CLOSED: this is a write, and ADR-0003 forbids letting a limiter outage lift the limit.
func (s *Reports) Create(ctx context.Context, sub subject.Subject, req ReportRequest) (domain.ErrorReport, error) {
	if sub.Anonymous() {
		return domain.ErrorReport{}, apperr.New(apperr.CodeUnauthenticated, "Sign in to report an error.")
	}
	if !sub.EmailVerified {
		return domain.ErrorReport{}, apperr.New(apperr.CodeEmailNotVerified, "Verify your email address to report an error.")
	}
	if err := validateReport(req); err != nil {
		return domain.ErrorReport{}, err
	}
	ok, retry, err := s.limiter.Hit(ctx, reportScope, sub.UserID, reportLimit, reportWindow)
	if err != nil {
		return domain.ErrorReport{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	if !ok {
		return domain.ErrorReport{}, apperr.New(apperr.CodeRateLimited, "Too many requests.").WithRetryAfter(retry)
	}
	report, err := s.store.CreateReport(ctx, domain.NewReport{
		ResourceID: req.ResourceID, ReporterID: sub.UserID, Kind: req.Kind, Body: req.Body, SourceID: req.SourceID,
	})
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return domain.ErrorReport{}, notFound()
	case errors.Is(err, domain.ErrSourceMismatch):
		return domain.ErrorReport{}, invalid("source_id", "invalid", "source_id must be a source of this resource.")
	case err != nil:
		return domain.ErrorReport{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	return report, nil
}

func validateReport(req ReportRequest) error {
	if !isKind(req.Kind) {
		return invalid("kind", "invalid", "kind must be one of broken_link, outdated, incorrect, unsupported_country, other.")
	}
	n := utf8.RuneCountInString(req.Body)
	if strings.TrimSpace(req.Body) == "" || n > domain.MaxReportBody {
		return invalid("body", "out_of_range", "body must be between 1 and 2000 characters.")
	}
	// PostgreSQL text holds neither NUL nor invalid UTF-8: refuse them here rather than answer 500.
	if !utf8.ValidString(req.Body) || strings.ContainsRune(req.Body, 0) {
		return invalid("body", "invalid", "body must be valid text.")
	}
	if req.SourceID != "" {
		if _, err := uuid.Parse(req.SourceID); err != nil {
			return invalid("source_id", "invalid", "source_id must be a UUID.")
		}
	}
	return nil
}

func isKind(k string) bool { return slices.Contains(domain.ReportKinds, k) }

func invalid(field, code, detail string) error {
	return apperr.New(apperr.CodeValidation, detail).WithViolations(apperr.Violation{Field: field, Code: code})
}

// notFound is the one 404 of the discover module, so an unknown id, a removed resource and an
// unpublished provider are indistinguishable.
func notFound() error { return apperr.New(apperr.CodeNotFound, "Resource not found.") }
