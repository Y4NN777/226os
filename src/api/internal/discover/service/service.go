// Package service is the catalogue read use case: filters, page size, and cursors.
package service

import (
	"context"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

const (
	defaultLimit = 20
	maxLimit     = 50
	maxQuery     = 200
)

// Reader loads published catalogue rows. The service owns this port.
type Reader interface {
	List(ctx context.Context, q domain.Query) ([]domain.Resource, error)
	Get(ctx context.Context, id string) (domain.Resource, error)
	Sources(ctx context.Context, id string) ([]domain.Source, error)
}

// Cursors signs page cursors and rejects a cursor used with different filters.
type Cursors interface {
	Encode(cursor.Payload) string
	Decode(token, filterHash string) (cursor.Payload, error)
}

// ListRequest is the public list. Nil Q means no text filter. Nil Limit means 20.
type ListRequest struct {
	Q, Country, Category, Type, SDK, Provider, Doclang *string
	Cursor                                             string
	Limit                                              *int
}

// Page is one catalogue page. NextCursor is empty on the last page.
type Page struct {
	Items      []domain.Resource
	NextCursor string
}

// Service reads the published catalogue.
type Service struct {
	store   Reader
	cursors Cursors
}

// New returns a catalogue service.
func New(store Reader, cursors Cursors) *Service {
	return &Service{store: store, cursors: cursors}
}

// List returns one page of published resources. A cursor from another filter is rejected.
func (s *Service) List(ctx context.Context, req ListRequest) (Page, error) {
	q, err := textFilter(req.Q)
	if err != nil {
		return Page{}, err
	}
	size, err := pageSize(req.Limit)
	if err != nil {
		return Page{}, err
	}
	filters := domain.Query{
		Q: q, Country: deref(req.Country), Category: deref(req.Category), Type: deref(req.Type),
		SDK: deref(req.SDK), Provider: deref(req.Provider), Doclang: deref(req.Doclang),
		Limit: size + 1,
	}
	hash := filterHash(filters)
	if req.Cursor != "" {
		p, err := s.cursors.Decode(req.Cursor, hash)
		if err != nil {
			return Page{}, err
		}
		if _, err := uuid.Parse(p.ID); err != nil || p.Sort == "" {
			return Page{}, invalidCursor()
		}
		filters.AfterName, filters.AfterID = p.Sort, p.ID
	}
	rows, err := s.store.List(ctx, filters)
	if err != nil {
		return Page{}, err
	}
	page := Page{Items: rows}
	if len(rows) > size {
		page.Items = rows[:size]
		last := page.Items[len(page.Items)-1]
		page.NextCursor = s.cursors.Encode(cursor.Payload{Sort: last.Name, ID: last.ID, FilterHash: hash})
	}
	if page.Items == nil {
		page.Items = []domain.Resource{}
	}
	return page, nil
}

// Get returns one published resource.
func (s *Service) Get(ctx context.Context, id string) (domain.Resource, error) {
	return s.store.Get(ctx, id)
}

// Sources returns the provenance of one published resource.
func (s *Service) Sources(ctx context.Context, id string) ([]domain.Source, error) {
	return s.store.Sources(ctx, id)
}

func textFilter(q *string) (string, error) {
	if q == nil {
		return "", nil
	}
	n := utf8.RuneCountInString(*q)
	if n < 1 || n > maxQuery {
		return "", apperr.New(apperr.CodeValidation, "q must be between 1 and 200 characters.").
			WithViolations(apperr.Violation{Field: "q", Code: "out_of_range"})
	}
	return *q, nil
}

func pageSize(v *int) (int, error) {
	if v == nil {
		return defaultLimit, nil
	}
	if *v < 1 || *v > maxLimit {
		return 0, apperr.New(apperr.CodeValidation, "limit must be between 1 and 50.").
			WithViolations(apperr.Violation{Field: "limit", Code: "out_of_range"})
	}
	return *v, nil
}

func filterHash(q domain.Query) string {
	return q.Q + "\x1f" + q.Country + "\x1f" + q.Category + "\x1f" + q.Type + "\x1f" + q.SDK + "\x1f" + q.Provider + "\x1f" + q.Doclang
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func invalidCursor() error {
	return apperr.New(apperr.CodeInvalidCursor, "The cursor is invalid or expired.")
}
