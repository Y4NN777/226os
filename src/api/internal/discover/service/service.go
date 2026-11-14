// Package service is the catalogue read use case: filters, page size, and cursors.
package service

import (
	"context"
	"slices"
	"strconv"
	"strings"
	"time"
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
	maxCountries = 20
	searchScope  = "search:ip"
	searchLimit  = 60
	searchWindow = time.Minute
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
	Q, Category, Type, SDK, Provider, Doclang *string
	Anonymous                                 bool
	ClientIP                                  string
	// Countries is the raw list from the request. Nil means no country filter.
	Countries []string
	Cursor    string
	Limit     *int
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
	limiter Limiter
}

// New returns a catalogue service.
func New(store Reader, cursors Cursors, limiter ...Limiter) *Service {
	var l Limiter
	if len(limiter) > 0 {
		l = limiter[0]
	}
	return &Service{store: store, cursors: cursors, limiter: l}
}

// List returns one page of published resources. A cursor from another filter is rejected.
func (s *Service) List(ctx context.Context, req ListRequest) (Page, error) {
	q, err := textFilter(req.Q)
	if err != nil {
		return Page{}, err
	}
	if q != "" && req.Anonymous && s.limiter != nil {
		ok, retry, limitErr := s.limiter.Hit(ctx, searchScope, req.ClientIP, searchLimit, searchWindow)
		if limitErr == nil && !ok {
			return Page{}, apperr.New(apperr.CodeRateLimited, "Too many requests.").WithRetryAfter(retry)
		}
	}
	size, err := pageSize(req.Limit)
	if err != nil {
		return Page{}, err
	}
	countries, err := countryFilter(req.Countries)
	if err != nil {
		return Page{}, err
	}
	filters := domain.Query{
		Q: q, Countries: countries, Category: deref(req.Category), Type: deref(req.Type),
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
	// PostgreSQL text holds neither NUL nor invalid UTF-8: refuse them here rather than answer 500.
	if !utf8.ValidString(*q) || strings.ContainsRune(*q, 0) {
		return "", apperr.New(apperr.CodeValidation, "q must be valid UTF-8 text.").
			WithViolations(apperr.Violation{Field: "q", Code: "invalid_value"})
	}
	return *q, nil
}

// countryFilter validates the requested country codes (1..20, each two capital letters) and returns them
// sorted without duplicates, so BF,CI and CI,BF are one filter with one cursor hash. Nil means no filter.
func countryFilter(in []string) ([]string, error) {
	if in == nil {
		return nil, nil
	}
	if len(in) < 1 {
		return nil, countryViolation("too_short", "country needs at least one code.")
	}
	if len(in) > maxCountries {
		return nil, countryViolation("too_many_items", "country accepts at most 20 codes.")
	}
	for _, c := range in {
		if !validCountry(c) {
			return nil, countryViolation("invalid_format", "country codes are ISO 3166-1 alpha-2, two capital letters.")
		}
	}
	out := slices.Clone(in)
	slices.Sort(out)
	return slices.Compact(out), nil
}

func validCountry(c string) bool {
	return len(c) == 2 && c[0] >= 'A' && c[0] <= 'Z' && c[1] >= 'A' && c[1] <= 'Z'
}

func countryViolation(code, detail string) error {
	return apperr.New(apperr.CodeValidation, detail).WithViolations(apperr.Violation{Field: "country", Code: code})
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

// filterHash binds a cursor to its filters. Every field is length-prefixed, so no field value can
// imitate a boundary (a plain separator join let two different filter sets share one hash).
func filterHash(q domain.Query) string {
	var b strings.Builder
	for _, f := range []string{q.Q, strings.Join(q.Countries, ","), q.Category, q.Type, q.SDK, q.Provider, q.Doclang} {
		b.WriteString(strconv.Itoa(len(f)))
		b.WriteByte(':')
		b.WriteString(f)
	}
	return b.String()
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
