// Package service is the search use case: validation, per-group fan-out and has_more.
package service

import (
	"context"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/search/domain"
)

const (
	defaultPerGroup = 5
	maxPerGroup     = 20
	minQuery        = 2
	maxQuery        = 200
)

// Reader runs the bounded ranked search of one group. The service owns this port.
type Reader interface {
	Search(ctx context.Context, group domain.Group, q domain.Query) ([]domain.Hit, error)
}

// Request is the validated shape of GET /search. Nil optional fields mean "not given".
type Request struct {
	Q        string
	Types    []string
	PerGroup *int
	Country  *string
	Lang     *string
}

// GroupResult is one group of hits; HasMore means more than PerGroup rows matched.
type GroupResult struct {
	Items   []domain.Hit
	HasMore bool
}

// Results holds only the requested groups. Query is the original q, not the truncated one.
type Results struct {
	Query  string
	Groups map[domain.Group]GroupResult
}

// Service searches the published, visible content.
type Service struct{ store Reader }

// New returns a search service.
func New(store Reader) *Service { return &Service{store: store} }

// Search runs one query per requested group (all four by default), each fetching per_group+1 rows.
// The text sent to the reader is truncated by C4 PROPOSED option 1 (ADR-0005 point 6); the
// response still echoes the original q.
func (s *Service) Search(ctx context.Context, req Request) (Results, error) {
	if n := utf8.RuneCountInString(req.Q); n < minQuery || n > maxQuery {
		return Results{}, invalid("q", "q must be between 2 and 200 characters.")
	}
	// PostgreSQL text holds neither NUL nor invalid UTF-8: refuse them here rather than answer 500.
	if !utf8.ValidString(req.Q) || strings.ContainsRune(req.Q, 0) {
		return Results{}, invalid("q", "q must be valid text.")
	}
	size := defaultPerGroup
	if req.PerGroup != nil {
		size = *req.PerGroup
	}
	if size < 1 || size > maxPerGroup {
		return Results{}, invalid("per_group", "per_group must be between 1 and 20.")
	}
	groups, err := groupsOf(req.Types)
	if err != nil {
		return Results{}, err
	}
	query := domain.Query{Text: domain.TruncateQuery(req.Q), Country: deref(req.Country), Lang: deref(req.Lang), Limit: size + 1}
	out := Results{Query: req.Q, Groups: make(map[domain.Group]GroupResult, len(groups))}
	for _, g := range groups {
		hits, err := s.store.Search(ctx, g, query)
		if err != nil {
			return Results{}, err
		}
		result := GroupResult{Items: hits, HasMore: len(hits) > size}
		if result.HasMore {
			result.Items = hits[:size]
		}
		if result.Items == nil {
			result.Items = []domain.Hit{}
		}
		out.Groups[g] = result
	}
	return out, nil
}

// groupsOf validates types: each must be a known group, with no repeats. None means all groups.
func groupsOf(types []string) ([]domain.Group, error) {
	if len(types) == 0 {
		return domain.Groups(), nil
	}
	groups := make([]domain.Group, 0, len(types))
	for _, t := range types {
		g := domain.Group(t)
		if !g.Valid() || slices.Contains(groups, g) {
			return nil, invalid("types", "types must list each of questions, resources, projects and learn at most once.")
		}
		groups = append(groups, g)
	}
	return groups, nil
}

func invalid(field, detail string) error {
	return apperr.New(apperr.CodeValidation, detail).WithViolations(apperr.Violation{Field: field, Code: "out_of_range"})
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
