// Package transport is the HTTP edge of Learn: decode, call one service method, encode.
package transport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/learn/domain"
	"github.com/Y4NN777/226os/src/api/internal/learn/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// Handler serves the five P0 Learn operations.
type Handler struct {
	svc    *service.Service
	logger *slog.Logger
}

// New returns a handler. logger may be nil.
func New(svc *service.Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Handler{svc: svc, logger: logger}
}

// ListLearnArticles serves GET /learn/articles.
func (h *Handler) ListLearnArticles(w http.ResponseWriter, r *http.Request, params apigen.ListLearnArticlesParams) {
	page, err := h.svc.ListArticles(r.Context(), service.ArticleListRequest{
		Q: params.Q, Type: str(params.Type), Topic: params.Topic, Level: str(params.Level), Lang: str(params.Lang),
		Cursor: deref(params.Cursor), Limit: params.Limit,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	items := make([]apigen.ArticleSummary, len(page.Items))
	for i, a := range page.Items {
		items[i] = apigen.ArticleSummary{
			Id: mustID(a.ID), Type: apigen.ArticleType(a.Kind), Title: a.Title, Summary: a.Summary,
			Lang: apigen.Lang(a.Lang), Level: apigen.Level(a.Level), Tags: a.Tags,
			Author: authorJSON(a.Author), PublishedAt: a.PublishedAt.UTC(),
		}
	}
	writeJSON(w, http.StatusOK, apigen.ArticlePage{Items: items, NextCursor: next(page.NextCursor)})
}

// GetLearnArticle serves GET /learn/articles/{articleId}.
func (h *Handler) GetLearnArticle(w http.ResponseWriter, r *http.Request, articleID apigen.ArticleId) {
	a, err := h.svc.GetArticle(r.Context(), subject.From(r.Context()), articleID.String())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, articleJSON(a))
}

// CreateLearnArticle serves POST /learn/articles. The author is the session subject; Idempotency-Key
// is handled by the router middleware.
func (h *Handler) CreateLearnArticle(w http.ResponseWriter, r *http.Request, _ apigen.CreateLearnArticleParams) {
	var body apigen.ArticleCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	req := service.CreateRequest{
		Kind: string(body.Type), Title: body.Title, BodyMD: body.BodyMd, Lang: string(body.Lang),
		Level: string(body.Level), Summary: body.Summary, Topic: body.Topic, Tags: body.Tags,
	}
	if body.Status != nil {
		s := string(*body.Status)
		req.Status = &s
	}
	if body.Links != nil {
		for _, l := range *body.Links {
			req.Links = append(req.Links, domain.LinkRef{Type: string(l.Type), ID: l.Id.String()})
		}
	}
	a, err := h.svc.CreateArticle(r.Context(), subject.From(r.Context()), req)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/learn/articles/"+a.ID)
	writeJSON(w, http.StatusCreated, articleJSON(a))
}

// ListTrainingSeries serves GET /learn/series.
func (h *Handler) ListTrainingSeries(w http.ResponseWriter, r *http.Request, params apigen.ListTrainingSeriesParams) {
	page, err := h.svc.ListSeries(r.Context(), service.SeriesListRequest{
		Topic: params.Topic, Level: str(params.Level), Lang: str(params.Lang),
		Cursor: deref(params.Cursor), Limit: params.Limit,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	items := make([]apigen.SeriesSummary, len(page.Items))
	for i, s := range page.Items {
		items[i] = apigen.SeriesSummary{
			Id: mustID(s.ID), Title: s.Title, Summary: s.Summary, Lang: apigen.Lang(s.Lang), Level: apigen.Level(s.Level),
			Tags: []string{}, Author: authorJSON(s.Author), LessonCount: s.LessonCount,
		}
	}
	writeJSON(w, http.StatusOK, apigen.SeriesPage{Items: items, NextCursor: next(page.NextCursor)})
}

// GetTrainingSeries serves GET /learn/series/{seriesId}. Series carry no tags and no relations at P0
// (no such tables), and my_progress is not served (C3, P1).
func (h *Handler) GetTrainingSeries(w http.ResponseWriter, r *http.Request, seriesID apigen.SeriesId) {
	s, err := h.svc.GetSeries(r.Context(), seriesID.String())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	lessons := make([]apigen.Lesson, len(s.Lessons))
	for i, l := range s.Lessons {
		lessons[i] = apigen.Lesson{Id: mustID(l.ID), Position: l.Position, Title: l.Title, ArticleId: mustID(l.ArticleID)}
	}
	writeJSON(w, http.StatusOK, apigen.TrainingSeries{
		Id: mustID(s.ID), Title: s.Title, Summary: s.Summary, Lang: apigen.Lang(s.Lang), Level: apigen.Level(s.Level),
		Tags: []string{}, Author: authorJSON(s.Author), Lessons: lessons, Related: []apigen.RelatedLink{},
		PublishedAt: s.PublishedAt.UTC(),
	})
}

func articleJSON(a domain.Article) apigen.Article {
	related := make([]apigen.RelatedLink, len(a.Related))
	for i, l := range a.Related {
		link := apigen.RelatedLink{Id: mustID(l.ID), Type: apigen.RelatedLinkType(l.Type), Title: l.Title}
		if l.Lang != "" {
			lang := apigen.Lang(l.Lang)
			link.Lang = &lang
		}
		related[i] = link
	}
	out := apigen.Article{
		Id: mustID(a.ID), Type: apigen.ArticleType(a.Kind), Title: a.Title, Summary: a.Summary, BodyMd: a.BodyMD,
		Lang: apigen.Lang(a.Lang), Level: apigen.Level(a.Level), Tags: a.Tags, Author: authorJSON(a.Author),
		Related: related, Status: apigen.ArticleStatus(a.Status), UpdatedAt: a.UpdatedAt.UTC(),
	}
	if a.PublishedAt != nil {
		t := a.PublishedAt.UTC()
		out.PublishedAt = &t
	}
	return out
}

func authorJSON(a domain.Author) apigen.AuthorRef {
	out := apigen.AuthorRef{Id: mustID(a.ID), Username: a.Username, DisplayName: a.DisplayName}
	if a.Country != "" {
		c := a.Country
		out.Country = &c
	}
	return out
}

// mustID parses an id the repository formatted from a uuid column; it cannot fail.
func mustID(s string) apigen.Uuid { return uuid.MustParse(s) }

func next(c string) *string {
	if c == "" {
		return nil
	}
	return &c
}

func str[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
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
