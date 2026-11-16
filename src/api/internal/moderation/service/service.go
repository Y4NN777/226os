// Package service implements the moderator-only use cases.
package service

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/moderation/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	writeScope  = "write:user"
	writeLimit  = 60
	writeWindow = time.Minute
)

type Store interface {
	Hide(context.Context, string, string, string, string, string) error
	Restore(context.Context, string, string, string, string, string) error
	ResolveReport(context.Context, string, string, string, string, string) (domain.ErrorReport, error)
}

type Limiter interface {
	Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error)
}

type Service struct {
	store   Store
	limiter Limiter
}

func New(store Store, limiter Limiter) *Service { return &Service{store: store, limiter: limiter} }

func (s *Service) authorize(ctx context.Context, sub subject.Subject) error {
	if sub.Anonymous() {
		return apperr.New(apperr.CodeUnauthenticated, "Authentication required.")
	}
	if !slices.Contains(sub.Roles, "moderator") {
		return apperr.New(apperr.CodeForbidden, "Moderator role required.")
	}
	allowed, retry, err := s.limiter.Hit(ctx, writeScope, sub.UserID, writeLimit, writeWindow)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	if !allowed {
		return apperr.New(apperr.CodeRateLimited, "Too many requests.").WithRetryAfter(retry)
	}
	return nil
}

func (s *Service) Hide(ctx context.Context, sub subject.Subject, kind, id, reason, requestID string) error {
	if err := s.authorize(ctx, sub); err != nil {
		return err
	}
	return mapStoreError(s.store.Hide(ctx, kind, id, sub.UserID, reason, requestID))
}

func (s *Service) Restore(ctx context.Context, sub subject.Subject, kind, id, reason, requestID string) error {
	if err := s.authorize(ctx, sub); err != nil {
		return err
	}
	return mapStoreError(s.store.Restore(ctx, kind, id, sub.UserID, reason, requestID))
}

func (s *Service) ResolveReport(ctx context.Context, sub subject.Subject, id, status, note, requestID string) (domain.ErrorReport, error) {
	if err := s.authorize(ctx, sub); err != nil {
		return domain.ErrorReport{}, err
	}
	r, err := s.store.ResolveReport(ctx, id, sub.UserID, status, note, requestID)
	return r, mapStoreError(err)
}

func mapStoreError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrNotFound):
		return apperr.New(apperr.CodeNotFound, "Target not found.")
	case errors.Is(err, domain.ErrStateConflict):
		return apperr.New(apperr.CodeStateConflict, "The target is not in the required state.")
	default:
		return apperr.Wrap(apperr.CodeInternal, err)
	}
}
