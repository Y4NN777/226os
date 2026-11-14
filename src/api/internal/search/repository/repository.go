// Package repository runs the ranked, bounded search queries through sqlc.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
	"github.com/Y4NN777/226os/src/api/internal/search/domain"
)

const queryTimeout = 2 * time.Second

// Repository is the sqlc adapter for the search reader.
type Repository struct{ q *dbgen.Queries }

// New returns a repository over db.
func New(db dbgen.DBTX) *Repository { return &Repository{q: dbgen.New(db)} }

// Search returns at most q.Limit ranked hits of one group. Each group applies only the filters its
// table supports: learn articles have no country, so q.Country is ignored for them.
func (r *Repository) Search(ctx context.Context, group domain.Group, q domain.Query) ([]domain.Hit, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	country, lang := text(q.Country), text(q.Lang)
	limit := int32(q.Limit) //nolint:gosec // the limit is at most 21
	var hits []domain.Hit
	switch group {
	case domain.Questions:
		rows, err := r.q.SearchQuestions(ctx, dbgen.SearchQuestionsParams{Q: q.Text, Country: country, Lang: lang, RowLimit: limit})
		if err != nil {
			return nil, fmt.Errorf("search %s: %w", group, err)
		}
		for _, row := range rows {
			hits = append(hits, hit(group, row.ID, row.Title, row.Excerpt, row.Language, row.PublishedAt))
		}
	case domain.Resources:
		rows, err := r.q.SearchResources(ctx, dbgen.SearchResourcesParams{Q: q.Text, Country: country, Lang: lang, RowLimit: limit})
		if err != nil {
			return nil, fmt.Errorf("search %s: %w", group, err)
		}
		for _, row := range rows {
			hits = append(hits, hit(group, row.ID, row.Title, row.Excerpt, row.Language, row.PublishedAt))
		}
	case domain.Projects:
		rows, err := r.q.SearchProjects(ctx, dbgen.SearchProjectsParams{Q: q.Text, Country: country, Lang: lang, RowLimit: limit})
		if err != nil {
			return nil, fmt.Errorf("search %s: %w", group, err)
		}
		for _, row := range rows {
			hits = append(hits, hit(group, row.ID, row.Title, row.Excerpt, row.Language, row.PublishedAt))
		}
	case domain.Learn:
		rows, err := r.q.SearchLearn(ctx, dbgen.SearchLearnParams{Q: q.Text, Lang: lang, RowLimit: limit})
		if err != nil {
			return nil, fmt.Errorf("search %s: %w", group, err)
		}
		for _, row := range rows {
			hits = append(hits, hit(group, row.ID, row.Title, row.Excerpt, row.Language, row.PublishedAt))
		}
	default:
		return nil, fmt.Errorf("search: unknown group %q", group)
	}
	if hits == nil {
		hits = []domain.Hit{}
	}
	return hits, nil
}

// hit maps one row. A NULL published_at (a learn article that was never published cannot match, but the
// column is nullable) is the zero Timestamptz, whose Time is already the zero time.
func hit(group domain.Group, id pgtype.UUID, title, excerpt, lang string, publishedAt pgtype.Timestamptz) domain.Hit {
	return domain.Hit{Type: group, ID: uuid.UUID(id.Bytes).String(), Title: title, Snippet: domain.Snippet(excerpt), Lang: lang, PublishedAt: publishedAt.Time.UTC()}
}

func text(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}
