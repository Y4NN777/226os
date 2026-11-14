// Package repository is the sqlc adapter of the Learn store.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/learn/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

const (
	queryTimeout = 2 * time.Second
	writeTimeout = 5 * time.Second

	topicFK = "learn_articles_topic_fk"
)

// Beginner is a DBTX that can also open a transaction (*pgxpool.Pool, *pgx.Conn).
type Beginner interface {
	dbgen.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Repository implements the Learn store over sqlc.
type Repository struct {
	db Beginner
	q  *dbgen.Queries
}

// New returns a repository over db.
func New(db Beginner) *Repository { return &Repository{db: db, q: dbgen.New(db)} }

// ListArticles returns published article summaries with their active tags.
func (r *Repository) ListArticles(ctx context.Context, q domain.ArticleQuery) ([]domain.ArticleSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	after, err := optionalID(q.AfterID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListLearnArticles(ctx, dbgen.ListLearnArticlesParams{
		Kind: text(q.Kind), Topic: text(q.Topic), Level: text(q.Level), Lang: text(q.Lang), Q: text(q.Q),
		AfterPublished: timestamp(q.AfterPublished), AfterID: after,
		RowLimit: int32(q.Limit), //nolint:gosec // page size is at most 51
	})
	if err != nil {
		return nil, fmt.Errorf("learn list articles: %w", err)
	}
	out := make([]domain.ArticleSummary, len(rows))
	ids := make([]pgtype.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
		out[i] = domain.ArticleSummary{
			ID: idString(row.ID), Kind: row.Kind, Title: row.Title, Summary: row.Summary, Lang: row.Language,
			Level: row.Level, PublishedAt: row.PublishedAt.Time,
			Author: author(row.AuthorID, row.AuthorUsername, row.AuthorDisplayName, row.AuthorCountry),
		}
	}
	tags, err := r.tags(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Tags = orEmpty(tags[out[i].ID])
	}
	return out, nil
}

// GetArticle returns a published article or viewerID's own draft or archived article.
func (r *Repository) GetArticle(ctx context.Context, id, viewerID string) (domain.Article, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	aid, err := requiredID(id)
	if err != nil {
		return domain.Article{}, domain.ErrNotFound
	}
	viewer, err := optionalID(viewerID)
	if err != nil {
		return domain.Article{}, err
	}
	row, err := r.q.GetLearnArticle(ctx, dbgen.GetLearnArticleParams{ID: aid, ViewerID: viewer})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Article{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Article{}, fmt.Errorf("learn get article: %w", err)
	}
	tags, err := r.tags(ctx, []pgtype.UUID{row.ID})
	if err != nil {
		return domain.Article{}, err
	}
	related, err := r.q.LearnArticleRelated(ctx, row.ID)
	if err != nil {
		return domain.Article{}, fmt.Errorf("learn article related: %w", err)
	}
	links := make([]domain.Link, len(related))
	for i, l := range related {
		links[i] = domain.Link{Type: l.LinkType, ID: idString(l.LinkID), Title: l.Title, Lang: l.Lang}
	}
	a := domain.Article{
		ID: idString(row.ID), Kind: row.Kind, Title: row.Title, Summary: row.Summary, BodyMD: row.BodyMd,
		Lang: row.Language, Level: row.Level, Status: row.Status, Tags: orEmpty(tags[idString(row.ID)]),
		Author:  author(row.AuthorID, row.AuthorUsername, row.AuthorDisplayName, row.AuthorCountry),
		Related: links, UpdatedAt: row.UpdatedAt.Time,
	}
	if row.PublishedAt.Valid {
		t := row.PublishedAt.Time
		a.PublishedAt = &t
	}
	return a, nil
}

// CreateArticle inserts the article, its tags (creating missing ones) and its links in one
// transaction. Each link target is locked FOR SHARE and must be published and public.
func (r *Repository) CreateArticle(ctx context.Context, in domain.NewArticle) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	authorID, err := requiredID(in.AuthorID)
	if err != nil {
		return "", err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("learn create begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	id, err := q.InsertLearnArticle(ctx, dbgen.InsertLearnArticleParams{
		AuthorID: authorID, Kind: in.Kind, Title: in.Title, Summary: text(in.Summary), BodyMd: in.BodyMD,
		Language: in.Lang, Level: in.Level, Topic: text(in.Topic), Status: in.Status,
	})
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23503" && pe.ConstraintName == topicFK {
			return "", domain.ErrUnknownTopic
		}
		return "", fmt.Errorf("learn insert article: %w", err)
	}
	if err := q.EnsureTags(ctx, dbgen.EnsureTagsParams{Slugs: in.Tags, CreatedBy: authorID}); err != nil {
		return "", fmt.Errorf("learn ensure tags: %w", err)
	}
	if err := q.InsertLearnArticleTags(ctx, dbgen.InsertLearnArticleTagsParams{ArticleID: id, Slugs: in.Tags}); err != nil {
		return "", fmt.Errorf("learn article tags: %w", err)
	}
	for _, l := range in.Links {
		if err := link(ctx, q, id, authorID, l); err != nil {
			return "", err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("learn create commit: %w", err)
	}
	return idString(id), nil
}

func link(ctx context.Context, q *dbgen.Queries, articleID, by pgtype.UUID, l domain.LinkRef) error {
	target, err := requiredID(l.ID)
	if err != nil {
		return domain.ErrLinkTarget
	}
	var n int64
	switch l.Type {
	case "question":
		n, err = q.LinkQuestionToArticle(ctx, dbgen.LinkQuestionToArticleParams{ArticleID: articleID, CreatedBy: by, TargetID: target})
	case "resource":
		n, err = q.LinkResourceToArticle(ctx, dbgen.LinkResourceToArticleParams{ArticleID: articleID, CreatedBy: by, TargetID: target})
	case "project":
		n, err = q.LinkProjectToArticle(ctx, dbgen.LinkProjectToArticleParams{ArticleID: articleID, CreatedBy: by, TargetID: target})
	default:
		return domain.ErrLinkTarget
	}
	if err != nil {
		return fmt.Errorf("learn link %s: %w", l.Type, err)
	}
	if n == 0 {
		return domain.ErrLinkTarget
	}
	return nil
}

// ListSeries returns published series summaries.
func (r *Repository) ListSeries(ctx context.Context, q domain.SeriesQuery) ([]domain.SeriesSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	after, err := optionalID(q.AfterID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListTrainingSeries(ctx, dbgen.ListTrainingSeriesParams{
		Topic: text(q.Topic), Level: text(q.Level), Lang: text(q.Lang),
		AfterPublished: timestamp(q.AfterPublished), AfterID: after,
		RowLimit: int32(q.Limit), //nolint:gosec // page size is at most 51
	})
	if err != nil {
		return nil, fmt.Errorf("learn list series: %w", err)
	}
	out := make([]domain.SeriesSummary, len(rows))
	for i, row := range rows {
		out[i] = domain.SeriesSummary{
			ID: idString(row.ID), Title: row.Title, Summary: row.Summary, Lang: row.Language, Level: row.Level,
			Author:      author(row.AuthorID, row.AuthorUsername, row.AuthorDisplayName, row.AuthorCountry),
			LessonCount: int(row.LessonCount), PublishedAt: row.PublishedAt.Time,
		}
	}
	return out, nil
}

// GetSeries returns a published series with its published lessons in order.
func (r *Repository) GetSeries(ctx context.Context, id string) (domain.Series, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	sid, err := requiredID(id)
	if err != nil {
		return domain.Series{}, domain.ErrNotFound
	}
	row, err := r.q.GetTrainingSeries(ctx, sid)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Series{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Series{}, fmt.Errorf("learn get series: %w", err)
	}
	lessons, err := r.q.TrainingSeriesLessons(ctx, sid)
	if err != nil {
		return domain.Series{}, fmt.Errorf("learn series lessons: %w", err)
	}
	out := domain.Series{
		ID: idString(row.ID), Title: row.Title, Summary: row.Summary, Lang: row.Language, Level: row.Level,
		Author:      author(row.AuthorID, row.AuthorUsername, row.AuthorDisplayName, row.AuthorCountry),
		Lessons:     make([]domain.Lesson, len(lessons)),
		PublishedAt: row.PublishedAt.Time,
	}
	for i, l := range lessons {
		out.Lessons[i] = domain.Lesson{ID: idString(l.ID), Title: l.Title, ArticleID: idString(l.ArticleID), Position: int(l.Position)}
	}
	return out, nil
}

func (r *Repository) tags(ctx context.Context, ids []pgtype.UUID) (map[string][]string, error) {
	out := map[string][]string{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.q.LearnArticleTags(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("learn article tags: %w", err)
	}
	for _, row := range rows {
		id := idString(row.ArticleID)
		out[id] = append(out[id], row.Slug)
	}
	return out, nil
}

func author(id pgtype.UUID, username, displayName, country string) domain.Author {
	return domain.Author{ID: idString(id), Username: username, DisplayName: displayName, Country: country}
}

func text(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func timestamp(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func requiredID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("learn id: %w", err)
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

func optionalID(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{}, nil
	}
	return requiredID(s)
}

// idString formats an id read from a NOT NULL uuid column.
func idString(id pgtype.UUID) string { return uuid.UUID(id.Bytes).String() }

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
