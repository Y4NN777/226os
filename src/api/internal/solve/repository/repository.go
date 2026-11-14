// Package repository is the sqlc adapter of the Solve store.
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

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
)

const (
	queryTimeout = 2 * time.Second
	writeTimeout = 5 * time.Second

	countryFK = "questions_country_fk"
)

// Beginner is a DBTX that can also open a transaction (*pgxpool.Pool, *pgx.Conn).
type Beginner interface {
	dbgen.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Repository implements the Solve store over sqlc.
type Repository struct {
	db Beginner
	q  *dbgen.Queries
}

// New returns a repository over db.
func New(db Beginner) *Repository { return &Repository{db: db, q: dbgen.New(db)} }

// ListQuestions returns published questions with their active tags and visible links.
func (r *Repository) ListQuestions(ctx context.Context, q domain.QuestionQuery) ([]domain.Question, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	after, err := optionalID(q.AfterID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListSolveQuestions(ctx, dbgen.ListSolveQuestionsParams{
		Tag: text(q.Tag), State: text(q.State), Country: text(q.Country), Lang: text(q.Lang), Q: text(q.Q),
		AfterCreated: timestamp(q.AfterCreated), AfterRank: q.AfterRank, AfterID: after,
		RowLimit: int32(q.Limit), //nolint:gosec // page size is at most 51
	})
	if err != nil {
		return nil, fmt.Errorf("solve list questions: %w", err)
	}
	out := make([]domain.Question, len(rows))
	ids := make([]pgtype.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
		out[i] = question(dbgen.GetSolveQuestionRow{
			ID: row.ID, Title: row.Title, BodyMd: row.BodyMd, Language: row.Language, Snippet: row.Snippet,
			SnippetLanguage: row.SnippetLanguage, CountryCode: row.CountryCode, AnswerCount: row.AnswerCount,
			VoteScore: row.VoteScore, AcceptedAnswerID: row.AcceptedAnswerID, CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt, AuthorID: row.AuthorID, AuthorUsername: row.AuthorUsername,
			AuthorDisplayName: row.AuthorDisplayName, AuthorCountry: row.AuthorCountry,
		})
		out[i].Rank = row.Rank
	}
	if err := r.attach(ctx, r.q, out, ids); err != nil {
		return nil, err
	}
	return out, nil
}

// GetQuestion returns a published question.
func (r *Repository) GetQuestion(ctx context.Context, id string) (domain.Question, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	qid, err := requiredID(id)
	if err != nil {
		return domain.Question{}, domain.ErrNotFound
	}
	return r.load(ctx, r.q, qid)
}

// CreateQuestion inserts the question, its tags (creating missing ones) and its links in one
// transaction. Each link target is locked FOR SHARE and must be published and public.
func (r *Repository) CreateQuestion(ctx context.Context, in domain.NewQuestion) (domain.Question, error) {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	authorID, err := requiredID(in.AuthorID)
	if err != nil {
		return domain.Question{}, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Question{}, fmt.Errorf("solve create begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	snippet, language := snippetColumns(in.Snippets)
	id, err := q.InsertSolveQuestion(ctx, dbgen.InsertSolveQuestionParams{
		AuthorID: authorID, Title: in.Title, BodyMd: in.BodyMD, Language: in.Lang,
		Snippet: snippet, SnippetLanguage: language, CountryCode: text(in.Country),
	})
	if err != nil {
		return domain.Question{}, insertErr("insert question", err)
	}
	if err := writeTags(ctx, q, id, authorID, in.Tags); err != nil {
		return domain.Question{}, err
	}
	if err := writeLinks(ctx, q, id, authorID, in.Links); err != nil {
		return domain.Question{}, err
	}
	out, err := r.load(ctx, q, id)
	if err != nil {
		return domain.Question{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Question{}, fmt.Errorf("solve create commit: %w", err)
	}
	return out, nil
}

// UpdateQuestion applies a merge patch under the question's row lock. guard decides on the locked
// row; its error is returned unchanged and nothing is written.
func (r *Repository) UpdateQuestion(ctx context.Context, id string, patch domain.QuestionPatch,
	guard func(domain.Question) error,
) (domain.Question, error) {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	qid, err := requiredID(id)
	if err != nil {
		return domain.Question{}, domain.ErrNotFound
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Question{}, fmt.Errorf("solve update begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	locked, err := q.LockSolveQuestion(ctx, qid)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Question{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Question{}, fmt.Errorf("solve lock question: %w", err)
	}
	if err := guard(domain.Question{
		ID: idString(locked.ID), Author: domain.Author{ID: idString(locked.AuthorID)},
		AnswerCount: int(locked.AnswerCount), VoteScore: int(locked.VoteScore),
		AcceptedAnswerID: optionalString(locked.AcceptedAnswerID), UpdatedAt: locked.UpdatedAt.Time,
	}); err != nil {
		return domain.Question{}, err
	}
	if err := apply(ctx, q, locked, patch); err != nil {
		return domain.Question{}, err
	}
	out, err := r.load(ctx, q, qid)
	if err != nil {
		return domain.Question{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Question{}, fmt.Errorf("solve update commit: %w", err)
	}
	return out, nil
}

func apply(ctx context.Context, q *dbgen.Queries, locked dbgen.LockSolveQuestionRow, p domain.QuestionPatch) error {
	args := dbgen.UpdateSolveQuestionParams{ID: locked.ID, Title: ptr(p.Title), BodyMd: ptr(p.BodyMD)}
	if p.Snippets != nil {
		args.SetSnippet = true
		args.Snippet, args.SnippetLanguage = snippetColumns(*p.Snippets)
	}
	if p.Country != nil {
		args.SetCountry = true
		args.CountryCode = text(*p.Country)
	}
	if err := q.UpdateSolveQuestion(ctx, args); err != nil {
		return insertErr("update question", err)
	}
	if p.Tags != nil {
		if err := q.DeleteSolveQuestionTags(ctx, locked.ID); err != nil {
			return fmt.Errorf("solve delete tags: %w", err)
		}
		if err := writeTags(ctx, q, locked.ID, locked.AuthorID, *p.Tags); err != nil {
			return err
		}
	}
	if p.Links != nil {
		if err := q.DeleteSolveQuestionLinks(ctx, dbgen.DeleteSolveQuestionLinksParams{QuestionID: locked.ID, CreatedBy: locked.AuthorID}); err != nil {
			return fmt.Errorf("solve delete links: %w", err)
		}
		return writeLinks(ctx, q, locked.ID, locked.AuthorID, *p.Links)
	}
	return nil
}

// ListAnswers returns a published question's answers in the contract order, with their sources.
func (r *Repository) ListAnswers(ctx context.Context, q domain.AnswerQuery) ([]domain.Answer, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	qid, err := requiredID(q.QuestionID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	exists, err := r.q.SolveQuestionExists(ctx, qid)
	if err != nil {
		return nil, fmt.Errorf("solve question exists: %w", err)
	}
	if !exists {
		return nil, domain.ErrNotFound
	}
	args := dbgen.ListSolveAnswersParams{QuestionID: qid, RowLimit: int32(q.Limit)} //nolint:gosec // page size is at most 51
	if k := q.After; k != nil {
		args.AfterID = pgtype.UUID{Bytes: uuid.MustParse(k.ID), Valid: true}
		args.AfterRank, args.AfterNegScore = int32(k.Rank), int32(k.NegScore) //nolint:gosec // rank is 0 or 1, a score is a column int4
		args.AfterCreated = timestamp(&k.CreatedAt)
	}
	rows, err := r.q.ListSolveAnswers(ctx, args)
	if err != nil {
		return nil, fmt.Errorf("solve list answers: %w", err)
	}
	out := make([]domain.Answer, len(rows))
	ids := make([]pgtype.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
		out[i] = answer(dbgen.GetSolveAnswerRow(row))
	}
	if err := r.sources(ctx, r.q, out, ids); err != nil {
		return nil, err
	}
	return out, nil
}

// CreateAnswer inserts the answer and raises the question's answer_count in one transaction: the
// count update row-locks the question and is the visibility check.
func (r *Repository) CreateAnswer(ctx context.Context, in domain.NewAnswer) (domain.Answer, error) {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	qid, err := requiredID(in.QuestionID)
	if err != nil {
		return domain.Answer{}, domain.ErrNotFound
	}
	authorID, err := requiredID(in.AuthorID)
	if err != nil {
		return domain.Answer{}, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Answer{}, fmt.Errorf("solve answer begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	if _, err := q.BumpSolveAnswerCount(ctx, qid); errors.Is(err, pgx.ErrNoRows) {
		return domain.Answer{}, domain.ErrNotFound
	} else if err != nil {
		return domain.Answer{}, fmt.Errorf("solve bump answer count: %w", err)
	}
	id, err := q.InsertSolveAnswer(ctx, dbgen.InsertSolveAnswerParams{QuestionID: qid, AuthorID: authorID, BodyMd: in.BodyMD, Language: in.Lang})
	if err != nil {
		return domain.Answer{}, fmt.Errorf("solve insert answer: %w", err)
	}
	if len(in.Sources) > 0 {
		if err := q.InsertSolveAnswerSources(ctx, dbgen.InsertSolveAnswerSourcesParams{AnswerID: id, Urls: in.Sources}); err != nil {
			return domain.Answer{}, fmt.Errorf("solve answer sources: %w", err)
		}
	}
	row, err := q.GetSolveAnswer(ctx, id)
	if err != nil {
		return domain.Answer{}, fmt.Errorf("solve get answer: %w", err)
	}
	out := []domain.Answer{answer(row)}
	if err := r.sources(ctx, q, out, []pgtype.UUID{id}); err != nil {
		return domain.Answer{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Answer{}, fmt.Errorf("solve answer commit: %w", err)
	}
	return out[0], nil
}

// load reads one published question with its tags and links through q (the pool or a transaction).
func (r *Repository) load(ctx context.Context, q *dbgen.Queries, id pgtype.UUID) (domain.Question, error) {
	row, err := q.GetSolveQuestion(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Question{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Question{}, fmt.Errorf("solve get question: %w", err)
	}
	out := []domain.Question{question(row)}
	if err := r.attach(ctx, q, out, []pgtype.UUID{id}); err != nil {
		return domain.Question{}, err
	}
	return out[0], nil
}

// attach fills Tags and Related of questions (ids in the same order).
func (r *Repository) attach(ctx context.Context, q *dbgen.Queries, qs []domain.Question, ids []pgtype.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	tagRows, err := q.SolveQuestionTags(ctx, ids)
	if err != nil {
		return fmt.Errorf("solve question tags: %w", err)
	}
	tags := map[string][]string{}
	for _, t := range tagRows {
		id := idString(t.QuestionID)
		tags[id] = append(tags[id], t.Slug)
	}
	linkRows, err := q.SolveQuestionRelated(ctx, ids)
	if err != nil {
		return fmt.Errorf("solve question related: %w", err)
	}
	related := map[string][]domain.Link{}
	for _, l := range linkRows {
		id := idString(l.QuestionID)
		related[id] = append(related[id], domain.Link{Type: l.Kind, ID: idString(l.TargetID), Title: l.Title, Lang: l.Lang})
	}
	for i := range qs {
		qs[i].Tags = orEmpty(tags[qs[i].ID])
		qs[i].Related = orEmpty(related[qs[i].ID])
	}
	return nil
}

// sources fills Sources of answers (ids in the same order).
func (r *Repository) sources(ctx context.Context, q *dbgen.Queries, as []domain.Answer, ids []pgtype.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	rows, err := q.SolveAnswerSources(ctx, ids)
	if err != nil {
		return fmt.Errorf("solve answer sources: %w", err)
	}
	urls := map[string][]string{}
	for _, s := range rows {
		id := idString(s.AnswerID)
		urls[id] = append(urls[id], s.Url)
	}
	for i := range as {
		as[i].Sources = orEmpty(urls[as[i].ID])
	}
	return nil
}

func writeTags(ctx context.Context, q *dbgen.Queries, questionID, by pgtype.UUID, slugs []string) error {
	if err := q.EnsureTags(ctx, dbgen.EnsureTagsParams{Slugs: slugs, CreatedBy: by}); err != nil {
		return fmt.Errorf("solve ensure tags: %w", err)
	}
	if err := q.InsertSolveQuestionTags(ctx, dbgen.InsertSolveQuestionTagsParams{QuestionID: questionID, Slugs: slugs}); err != nil {
		return fmt.Errorf("solve question tags: %w", err)
	}
	return nil
}

func writeLinks(ctx context.Context, q *dbgen.Queries, questionID, by pgtype.UUID, links []domain.LinkRef) error {
	for _, l := range links {
		target, err := requiredID(l.ID)
		if err != nil {
			return domain.ErrLinkTarget
		}
		var usable bool
		switch l.Type {
		case "resource":
			usable, err = q.SolveLinkResource(ctx, dbgen.SolveLinkResourceParams{QuestionID: questionID, CreatedBy: by, TargetID: target})
		case "project":
			usable, err = q.SolveLinkProject(ctx, dbgen.SolveLinkProjectParams{QuestionID: questionID, CreatedBy: by, TargetID: target})
		case "article":
			usable, err = q.SolveLinkArticle(ctx, dbgen.SolveLinkArticleParams{QuestionID: questionID, CreatedBy: by, TargetID: target})
		default:
			return domain.ErrLinkTarget
		}
		if err != nil {
			return fmt.Errorf("solve link %s: %w", l.Type, err)
		}
		if !usable {
			return domain.ErrLinkTarget
		}
	}
	return nil
}

// insertErr maps the country foreign key to its domain error and wraps anything else.
func insertErr(op string, err error) error {
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23503" && pe.ConstraintName == countryFK {
		return domain.ErrUnknownCountry
	}
	return fmt.Errorf("solve %s: %w", op, err)
}

func question(row dbgen.GetSolveQuestionRow) domain.Question {
	out := domain.Question{
		ID: idString(row.ID), Title: row.Title, BodyMD: row.BodyMd, Lang: row.Language, Country: row.CountryCode,
		AcceptedAnswerID: optionalString(row.AcceptedAnswerID), Snippets: []domain.Snippet{},
		Author:      author(row.AuthorID, row.AuthorUsername, row.AuthorDisplayName, row.AuthorCountry),
		AnswerCount: int(row.AnswerCount), VoteScore: int(row.VoteScore),
		CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
	if row.Snippet != "" {
		out.Snippets = []domain.Snippet{{Language: row.SnippetLanguage, Code: row.Snippet}}
	}
	return out
}

func answer(row dbgen.GetSolveAnswerRow) domain.Answer {
	return domain.Answer{
		ID: idString(row.ID), QuestionID: idString(row.QuestionID), BodyMD: row.BodyMd, Lang: row.Language,
		Snippets: []domain.Snippet{}, Author: author(row.AuthorID, row.AuthorUsername, row.AuthorDisplayName, row.AuthorCountry),
		VoteScore: int(row.VoteScore), Accepted: row.IsAccepted, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func author(id pgtype.UUID, username, displayName, country string) domain.Author {
	return domain.Author{ID: idString(id), Username: username, DisplayName: displayName, Country: country}
}

// snippetColumns maps at most one snippet to the question columns; none is NULL, NULL.
func snippetColumns(in []domain.Snippet) (code, language pgtype.Text) {
	if len(in) == 0 {
		return pgtype.Text{}, pgtype.Text{}
	}
	return text(in[0].Code), text(in[0].Language)
}

func text(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func ptr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
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
		return pgtype.UUID{}, fmt.Errorf("solve id: %w", err)
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

func optionalString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return idString(id)
}

func orEmpty[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
