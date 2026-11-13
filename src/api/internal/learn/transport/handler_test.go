package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/learn/domain"
	"github.com/Y4NN777/226os/src/api/internal/learn/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	userID    = "00000000-0000-7000-8000-0000000000a1"
	articleID = "00000000-0000-7000-8000-0000000000b1"
	seriesID  = "00000000-0000-7000-8000-0000000000c1"
	lessonID  = "00000000-0000-7000-8000-0000000000c2"
	qID       = "00000000-0000-7000-8000-0000000000d1"
)

var when = time.Date(2026, 11, 13, 10, 0, 0, 0, time.FixedZone("x", 3600))

type fakeStore struct {
	articles []domain.ArticleSummary
	article  domain.Article
	series   []domain.SeriesSummary
	one      domain.Series
	err      error
	created  []domain.NewArticle
}

func (f *fakeStore) ListArticles(context.Context, domain.ArticleQuery) ([]domain.ArticleSummary, error) {
	return f.articles, f.err
}

func (f *fakeStore) GetArticle(context.Context, string, string) (domain.Article, error) {
	return f.article, f.err
}

func (f *fakeStore) CreateArticle(_ context.Context, in domain.NewArticle) (string, error) {
	f.created = append(f.created, in)
	return articleID, f.err
}

func (f *fakeStore) ListSeries(context.Context, domain.SeriesQuery) ([]domain.SeriesSummary, error) {
	return f.series, f.err
}

func (f *fakeStore) GetSeries(context.Context, string) (domain.Series, error) { return f.one, f.err }

type allowAll struct{}

func (allowAll) Hit(context.Context, string, string, int, time.Duration) (bool, time.Duration, error) {
	return true, 0, nil
}

func handler(t *testing.T, st *fakeStore) *Handler {
	t.Helper()
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return New(service.New(st, c, allowAll{}), nil)
}

var alice = domain.Author{ID: userID, Username: "alice", DisplayName: "Alice", Country: "BF"}

func article() domain.Article {
	at := when
	return domain.Article{
		ID: articleID, Kind: "tutorial", Title: "Intégrer Mobile Money", Summary: "", BodyMD: "# Étape\n<b>brut</b>\n",
		Lang: "fr", Level: "beginner", Status: "published", Tags: []string{"api"}, Author: alice,
		Related:     []domain.Link{{Type: "question", ID: qID, Title: "Q ?", Lang: "fr"}, {Type: "project", ID: qID, Title: "P"}},
		PublishedAt: &at, UpdatedAt: when,
	}
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("%v: %s", err, rec.Body.String())
	}
	return v
}

func get(t *testing.T, sub subject.Subject) *http.Request {
	t.Helper()
	return httptest.NewRequestWithContext(subject.With(t.Context(), sub), http.MethodGet, "/", nil)
}

func TestListLearnArticlesEncodesThePage(t *testing.T) {
	st := &fakeStore{articles: []domain.ArticleSummary{
		{ID: articleID, Kind: "article", Title: "Titre", Lang: "en", Level: "advanced", Tags: []string{}, Author: domain.Author{ID: userID, Username: "bob", DisplayName: "Bob"}, PublishedAt: when},
		{ID: qID, Kind: "article", Title: "Autre", Lang: "en", Level: "advanced", Tags: []string{}, Author: alice, PublishedAt: when},
	}}
	rec := httptest.NewRecorder()
	typ, level, lang := apigen.ArticleType("article"), apigen.Level("advanced"), apigen.Lang("en")
	limit := 1
	handler(t, st).ListLearnArticles(rec, get(t, subject.Subject{}), apigen.ListLearnArticlesParams{Type: &typ, Level: &level, Lang: &lang, Limit: &limit})
	page := decode[apigen.ArticlePage](t, rec)
	if rec.Code != http.StatusOK || len(page.Items) != 1 || page.NextCursor == nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if it := page.Items[0]; it.Author.Country != nil || it.PublishedAt.Location() != time.UTC || it.Title != "Titre" {
		t.Fatalf("item %+v", it)
	}
}

func TestListLearnArticlesLastPageHasNullCursor(t *testing.T) {
	rec := httptest.NewRecorder()
	handler(t, &fakeStore{}).ListLearnArticles(rec, get(t, subject.Subject{}), apigen.ListLearnArticlesParams{})
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"next_cursor":null`) || !strings.Contains(rec.Body.String(), `"items":[]`) {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

func TestListLearnArticlesErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	bad := "garbage"
	handler(t, &fakeStore{}).ListLearnArticles(rec, get(t, subject.Subject{}), apigen.ListLearnArticlesParams{Cursor: &bad})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("cursor %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{err: errors.New("pg: secret detail")}).ListLearnArticles(rec, get(t, subject.Subject{}), apigen.ListLearnArticlesParams{})
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "secret") {
		t.Fatalf("store %d %s", rec.Code, rec.Body.String())
	}
}

func TestGetLearnArticleReturnsTheOriginalBody(t *testing.T) {
	rec := httptest.NewRecorder()
	handler(t, &fakeStore{article: article()}).GetLearnArticle(rec, get(t, subject.Subject{}), uuid.MustParse(articleID))
	a := decode[apigen.Article](t, rec)
	if rec.Code != http.StatusOK || a.BodyMd != "# Étape\n<b>brut</b>\n" || a.Lang != "fr" || a.PublishedAt == nil {
		t.Fatalf("%d %+v", rec.Code, a)
	}
	if len(a.Related) != 2 || a.Related[0].Lang == nil || a.Related[1].Lang != nil || a.Author.Country == nil || *a.Author.Country != "BF" {
		t.Fatalf("related %+v author %+v", a.Related, a.Author)
	}
}

func TestGetLearnArticleDraftAndNotFound(t *testing.T) {
	draft := article()
	draft.PublishedAt, draft.Status = nil, "draft"
	rec := httptest.NewRecorder()
	handler(t, &fakeStore{article: draft}).GetLearnArticle(rec, get(t, subject.Subject{UserID: userID}), uuid.MustParse(articleID))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"published_at":null`) {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{err: domain.ErrNotFound}).GetLearnArticle(rec, get(t, subject.Subject{}), uuid.MustParse(articleID))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("%d", rec.Code)
	}
}

const createBody = `{"type":"tutorial","title":"Intégrer Mobile Money","body_md":"# Étape\n","lang":"fr","level":"beginner",` +
	`"tags":["api"],"summary":"Court","topic":"payments","status":"published","links":[{"type":"question","id":"` + qID + `"}]}`

func post(t *testing.T, h *Handler, sub subject.Subject, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(subject.With(t.Context(), sub), http.MethodPost, "/api/v1/learn/articles", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.CreateLearnArticle(rec, req, apigen.CreateLearnArticleParams{})
	return rec
}

func TestCreateLearnArticleAnswers201WithLocation(t *testing.T) {
	st := &fakeStore{article: article()}
	rec := post(t, handler(t, st), subject.Subject{UserID: userID, EmailVerified: true}, createBody)
	if rec.Code != http.StatusCreated || rec.Header().Get("Location") != "/api/v1/learn/articles/"+articleID {
		t.Fatalf("%d %q %s", rec.Code, rec.Header().Get("Location"), rec.Body.String())
	}
	in := st.created[0]
	if in.AuthorID != userID || in.Topic != "payments" || in.Summary != "Court" || len(in.Links) != 1 || in.Links[0].ID != qID {
		t.Fatalf("created %+v", in)
	}
}

func TestCreateLearnArticleRejectsUnknownFields(t *testing.T) {
	st := &fakeStore{}
	rec := post(t, handler(t, st), subject.Subject{UserID: userID, EmailVerified: true}, `{"author_id":"`+userID+`"}`)
	if rec.Code < 400 || rec.Code >= 500 || len(st.created) != 0 {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
}

// Negative authz for POST /api/v1/learn/articles (AUTHORIZATION_MATRIX, allow: member; publishing
// needs a verified email): anonymous callers get 401 (even for a draft), unverified members 403 when
// publishing, and nothing is created.
func TestAuthzCreateLearnArticleAnonymous(t *testing.T) {
	draft := strings.Replace(createBody, `"status":"published"`, `"status":"draft"`, 1)
	cases := []struct {
		name   string
		sub    subject.Subject
		body   string
		status int
	}{
		{"anonymous publish", subject.Subject{}, createBody, http.StatusUnauthorized},
		{"anonymous draft", subject.Subject{}, draft, http.StatusUnauthorized},
		{"member without a verified email publishing", subject.Subject{UserID: userID}, createBody, http.StatusForbidden},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{}
			req := httptest.NewRequestWithContext(subject.With(t.Context(), tc.sub), http.MethodPost, "/api/v1/learn/articles", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler(t, st).CreateLearnArticle(rec, req, apigen.CreateLearnArticleParams{})
			if rec.Code != tc.status || len(st.created) != 0 {
				t.Fatalf("status %d, created %d; want %d and nothing created", rec.Code, len(st.created), tc.status)
			}
		})
	}
}

func TestListTrainingSeriesEncodesThePage(t *testing.T) {
	st := &fakeStore{series: []domain.SeriesSummary{{ID: seriesID, Title: "Formation", Summary: "S", Lang: "fr", Level: "beginner", Author: alice, LessonCount: 3, PublishedAt: when}}}
	topic, level, lang := "ai", apigen.Level("beginner"), apigen.Lang("fr")
	rec := httptest.NewRecorder()
	handler(t, st).ListTrainingSeries(rec, get(t, subject.Subject{}), apigen.ListTrainingSeriesParams{Topic: &topic, Level: &level, Lang: &lang})
	page := decode[apigen.SeriesPage](t, rec)
	if rec.Code != http.StatusOK || len(page.Items) != 1 || page.Items[0].LessonCount != 3 || page.Items[0].Tags == nil || page.NextCursor != nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	limit := 0
	handler(t, st).ListTrainingSeries(rec, get(t, subject.Subject{}), apigen.ListTrainingSeriesParams{Limit: &limit})
	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("limit %d", rec.Code)
	}
}

func TestGetTrainingSeriesHasLessonsAndNoProgress(t *testing.T) {
	st := &fakeStore{one: domain.Series{ID: seriesID, Title: "Formation", Summary: "S", Lang: "fr", Level: "beginner", Author: alice,
		Lessons: []domain.Lesson{{ID: lessonID, Title: "Leçon 1", ArticleID: articleID, Position: 1}}, PublishedAt: when}}
	rec := httptest.NewRecorder()
	handler(t, st).GetTrainingSeries(rec, get(t, subject.Subject{UserID: userID}), uuid.MustParse(seriesID))
	s := decode[apigen.TrainingSeries](t, rec)
	if rec.Code != http.StatusOK || len(s.Lessons) != 1 || s.Lessons[0].ArticleId.String() != articleID || s.Related == nil || s.Tags == nil {
		t.Fatalf("%d %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "my_progress") {
		t.Fatal("my_progress served before C3")
	}
	rec = httptest.NewRecorder()
	handler(t, &fakeStore{err: domain.ErrNotFound}).GetTrainingSeries(rec, get(t, subject.Subject{}), uuid.MustParse(seriesID))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("%d", rec.Code)
	}
}

func TestNewHandlerWithNilLogger(t *testing.T) {
	if New(nil, nil).logger == nil {
		t.Fatal("nil logger")
	}
}
