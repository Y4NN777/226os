// Package service is the Learn use case (US-LEARN-001): public reads of published articles and series,
// and member publishing with the original language kept.
package service

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/learn/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	defaultLimit = 20
	maxLimit     = 50
	maxQuery     = 200
	maxTags      = 5

	// Article creation draws on the one authenticated-writes budget, 60 per user per minute
	// (ADR-0003); error reports and S3 to S5 share the same write:user scope.
	writeScope  = "write:user"
	writeLimit  = 60
	writeWindow = time.Minute
)

// Store loads and saves Learn content. The service owns this port.
//
// GetArticle returns a published article, or a draft or archived article whose author is viewerID
// (empty for anonymous); anything else is domain.ErrNotFound. CreateArticle inserts the article, its
// tags and its links in one transaction and returns domain.ErrUnknownTopic or domain.ErrLinkTarget
// when the topic or a link target is not usable.
type Store interface {
	ListArticles(ctx context.Context, q domain.ArticleQuery) ([]domain.ArticleSummary, error)
	GetArticle(ctx context.Context, id, viewerID string) (domain.Article, error)
	CreateArticle(ctx context.Context, in domain.NewArticle) (string, error)
	ListSeries(ctx context.Context, q domain.SeriesQuery) ([]domain.SeriesSummary, error)
	GetSeries(ctx context.Context, id string) (domain.Series, error)
}

// Cursors signs page cursors and rejects a cursor used with different filters.
type Cursors interface {
	Encode(cursor.Payload) string
	Decode(token, filterHash string) (cursor.Payload, error)
}

// Limiter is the fixed-window rate limiter. The service owns this port.
type Limiter interface {
	Hit(ctx context.Context, scope, value string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)
}

// ArticleListRequest is listLearnArticles. Nil means unset; nil Limit means 20.
type ArticleListRequest struct {
	Q, Type, Topic, Level, Lang *string
	Cursor                      string
	Limit                       *int
}

// SeriesListRequest is listTrainingSeries. Nil means unset; nil Limit means 20.
type SeriesListRequest struct {
	Topic, Level, Lang *string
	Cursor             string
	Limit              *int
}

// CreateRequest is the createLearnArticle body. Nil Status means published.
type CreateRequest struct {
	Kind, Title, BodyMD, Lang, Level string
	Summary, Topic, Status           *string
	Tags                             []string
	Links                            []domain.LinkRef
}

// ArticlePage is one article page. NextCursor is empty on the last page.
type ArticlePage struct {
	Items      []domain.ArticleSummary
	NextCursor string
}

// SeriesPage is one series page. NextCursor is empty on the last page.
type SeriesPage struct {
	Items      []domain.SeriesSummary
	NextCursor string
}

// Service is the Learn use case.
type Service struct {
	store   Store
	cursors Cursors
	limiter Limiter
}

// New returns the Learn service.
func New(store Store, cursors Cursors, limiter Limiter) *Service {
	return &Service{store: store, cursors: cursors, limiter: limiter}
}

// ListArticles returns one page of published articles, newest first.
func (s *Service) ListArticles(ctx context.Context, req ArticleListRequest) (ArticlePage, error) {
	q, err := textFilter(req.Q)
	if err != nil {
		return ArticlePage{}, err
	}
	size, err := pageSize(req.Limit)
	if err != nil {
		return ArticlePage{}, err
	}
	query := domain.ArticleQuery{
		Q: q, Kind: deref(req.Type), Topic: deref(req.Topic), Level: deref(req.Level), Lang: deref(req.Lang),
		Limit: size + 1,
	}
	hash := articleHash(query)
	if req.Cursor != "" {
		at, id, err := s.decode(req.Cursor, hash)
		if err != nil {
			return ArticlePage{}, err
		}
		query.AfterPublished, query.AfterID = &at, id
	}
	rows, err := s.store.ListArticles(ctx, query)
	if err != nil {
		return ArticlePage{}, err
	}
	page := ArticlePage{Items: rows}
	if len(rows) > size {
		page.Items = rows[:size]
		last := page.Items[size-1]
		page.NextCursor = s.encode(last.PublishedAt, last.ID, hash)
	}
	if page.Items == nil {
		page.Items = []domain.ArticleSummary{}
	}
	return page, nil
}

// GetArticle returns a published article, or the caller's own draft or archived article.
func (s *Service) GetArticle(ctx context.Context, sub subject.Subject, id string) (domain.Article, error) {
	a, err := s.store.GetArticle(ctx, id, sub.UserID)
	if err != nil {
		return domain.Article{}, readErr(err)
	}
	return a, nil
}

// CreateArticle saves an article for the session subject. A draft needs a session; publishing also
// needs a verified email. The author is always the subject, never the request. The text is stored as
// given: no trimming, no translation. The limiter fails CLOSED (ADR-0003: a write).
func (s *Service) CreateArticle(ctx context.Context, sub subject.Subject, req CreateRequest) (domain.Article, error) {
	if sub.Anonymous() {
		return domain.Article{}, apperr.New(apperr.CodeUnauthenticated, "Sign in to write an article.")
	}
	status := domain.StatusPublished
	if req.Status != nil {
		status = *req.Status
	}
	if status == domain.StatusPublished && !sub.EmailVerified {
		return domain.Article{}, apperr.New(apperr.CodeEmailNotVerified, "Verify your email address to publish.")
	}
	in, err := validate(req, status)
	if err != nil {
		return domain.Article{}, err
	}
	in.AuthorID = sub.UserID
	ok, retry, err := s.limiter.Hit(ctx, writeScope, sub.UserID, writeLimit, writeWindow)
	if err != nil {
		return domain.Article{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	if !ok {
		return domain.Article{}, apperr.New(apperr.CodeRateLimited, "Too many requests.").WithRetryAfter(retry)
	}
	id, err := s.store.CreateArticle(ctx, in)
	switch {
	case errors.Is(err, domain.ErrUnknownTopic):
		return domain.Article{}, invalid("/topic", "invalid_value", "topic must be a known topic code.")
	case errors.Is(err, domain.ErrLinkTarget):
		return domain.Article{}, invalid("/links", "invalid_value", "each link must point to a published, public item.")
	case err != nil:
		return domain.Article{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	a, err := s.store.GetArticle(ctx, id, sub.UserID)
	if err != nil {
		return domain.Article{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	return a, nil
}

// ListSeries returns one page of published training series, newest first.
func (s *Service) ListSeries(ctx context.Context, req SeriesListRequest) (SeriesPage, error) {
	size, err := pageSize(req.Limit)
	if err != nil {
		return SeriesPage{}, err
	}
	query := domain.SeriesQuery{Topic: deref(req.Topic), Level: deref(req.Level), Lang: deref(req.Lang), Limit: size + 1}
	hash := seriesHash(query)
	if req.Cursor != "" {
		at, id, err := s.decode(req.Cursor, hash)
		if err != nil {
			return SeriesPage{}, err
		}
		query.AfterPublished, query.AfterID = &at, id
	}
	rows, err := s.store.ListSeries(ctx, query)
	if err != nil {
		return SeriesPage{}, err
	}
	page := SeriesPage{Items: rows}
	if len(rows) > size {
		page.Items = rows[:size]
		last := page.Items[size-1]
		page.NextCursor = s.encode(last.PublishedAt, last.ID, hash)
	}
	if page.Items == nil {
		page.Items = []domain.SeriesSummary{}
	}
	return page, nil
}

// GetSeries returns a published series with its published lessons. my_progress is not served: the
// progress table is P1 (C3 open), and the contract makes the field optional.
func (s *Service) GetSeries(ctx context.Context, id string) (domain.Series, error) {
	series, err := s.store.GetSeries(ctx, id)
	if err != nil {
		return domain.Series{}, readErr(err)
	}
	return series, nil
}

func validate(req CreateRequest, status string) (domain.NewArticle, error) {
	if status != domain.StatusDraft && status != domain.StatusPublished {
		return domain.NewArticle{}, invalid("/status", "invalid_value", "status must be draft or published.")
	}
	if !usableText(req.Title) {
		return domain.NewArticle{}, invalid("/title", "invalid_value", "title must be non-blank valid text.")
	}
	if !usableText(req.BodyMD) {
		return domain.NewArticle{}, invalid("/body_md", "invalid_value", "body_md must be non-blank valid text.")
	}
	summary := deref(req.Summary)
	if !validText(summary) {
		return domain.NewArticle{}, invalid("/summary", "invalid_value", "summary must be valid text.")
	}
	tags := dedupe(req.Tags)
	if len(tags) < 1 || len(tags) > maxTags {
		return domain.NewArticle{}, invalid("/tags", "out_of_range", "tags must hold 1 to 5 distinct slugs.")
	}
	links := make([]domain.LinkRef, 0, len(req.Links))
	for _, l := range req.Links {
		if !slices.Contains(domain.LinkTypes, l.Type) {
			return domain.NewArticle{}, invalid("/links", "invalid_value", "an article can link to a question, a resource or a project.")
		}
		if _, err := uuid.Parse(l.ID); err != nil {
			return domain.NewArticle{}, invalid("/links", "invalid_value", "link id must be a UUID.")
		}
		if !slices.Contains(links, l) {
			links = append(links, l)
		}
	}
	return domain.NewArticle{
		Kind: req.Kind, Title: req.Title, Summary: summary, BodyMD: req.BodyMD, Lang: req.Lang,
		Level: req.Level, Topic: deref(req.Topic), Status: status, Tags: tags, Links: links,
	}, nil
}

// usableText is valid text that is not only whitespace.
func usableText(s string) bool { return validText(s) && strings.TrimSpace(s) != "" }

// validText refuses what PostgreSQL text cannot hold (NUL, invalid UTF-8) rather than answer 500.
func validText(s string) bool { return utf8.ValidString(s) && !strings.ContainsRune(s, 0) }

func dedupe(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	return out
}

func (s *Service) encode(at time.Time, id, hash string) string {
	return s.cursors.Encode(cursor.Payload{Sort: at.UTC().Format(time.RFC3339Nano), ID: id, FilterHash: hash})
}

func (s *Service) decode(token, hash string) (time.Time, string, error) {
	p, err := s.cursors.Decode(token, hash)
	if err != nil {
		return time.Time{}, "", err
	}
	at, err := time.Parse(time.RFC3339Nano, p.Sort)
	if err != nil {
		return time.Time{}, "", invalidCursor()
	}
	if _, err := uuid.Parse(p.ID); err != nil {
		return time.Time{}, "", invalidCursor()
	}
	return at, p.ID, nil
}

func textFilter(q *string) (string, error) {
	if q == nil {
		return "", nil
	}
	n := utf8.RuneCountInString(*q)
	if n < 1 || n > maxQuery {
		return "", invalid("q", "out_of_range", "q must be between 1 and 200 characters.")
	}
	if !validText(*q) {
		return "", invalid("q", "invalid_value", "q must be valid UTF-8 text.")
	}
	return *q, nil
}

func pageSize(v *int) (int, error) {
	if v == nil {
		return defaultLimit, nil
	}
	if *v < 1 || *v > maxLimit {
		return 0, invalid("limit", "out_of_range", "limit must be between 1 and 50.")
	}
	return *v, nil
}

// articleHash and seriesHash bind a cursor to its filters; every field is length-prefixed so no
// value can imitate a boundary.
func articleHash(q domain.ArticleQuery) string {
	return lengthPrefixed("article", q.Q, q.Kind, q.Topic, q.Level, q.Lang)
}

func seriesHash(q domain.SeriesQuery) string {
	return lengthPrefixed("series", q.Topic, q.Level, q.Lang)
}

func lengthPrefixed(fields ...string) string {
	var b strings.Builder
	for _, f := range fields {
		b.WriteString(strconv.Itoa(len(f)))
		b.WriteByte(':')
		b.WriteString(f)
	}
	return b.String()
}

// readErr makes a missing, removed and invisible item one indistinguishable 404.
func readErr(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return apperr.New(apperr.CodeNotFound, "Not found.")
	}
	return apperr.Wrap(apperr.CodeInternal, err)
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// invalid is a 422 with one violation: field is the JSON pointer of a body member or the name of a query
// parameter, and code is one of the Violation.code enum values of api/openapi.yaml.
func invalid(field, code, detail string) error {
	return apperr.New(apperr.CodeValidation, detail).WithViolations(apperr.Violation{Field: field, Code: code})
}

func invalidCursor() error {
	return apperr.New(apperr.CodeInvalidCursor, "The cursor is invalid or expired.")
}
