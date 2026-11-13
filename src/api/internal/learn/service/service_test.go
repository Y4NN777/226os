package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/learn/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	aliceID   = "00000000-0000-7000-8000-0000000000a1"
	articleID = "00000000-0000-7000-8000-0000000000b1"
	seriesID  = "00000000-0000-7000-8000-0000000000c1"
	targetID  = "00000000-0000-7000-8000-0000000000d1"
)

var (
	alice      = subject.Subject{UserID: aliceID, EmailVerified: true}
	unverified = subject.Subject{UserID: aliceID}
	anonymous  = subject.Subject{}
	when       = time.Date(2026, 11, 13, 9, 30, 0, 123456789, time.UTC)
	errDB      = errors.New("db down")
)

type fakeStore struct {
	articles    []domain.ArticleSummary
	series      []domain.SeriesSummary
	article     domain.Article
	oneSeries   domain.Series
	listErr     error
	getErr      error
	createErr   error
	seriesErr   error
	created     *domain.NewArticle
	lastQuery   domain.ArticleQuery
	lastSeriesQ domain.SeriesQuery
	viewer      string
}

func (f *fakeStore) ListArticles(_ context.Context, q domain.ArticleQuery) ([]domain.ArticleSummary, error) {
	f.lastQuery = q
	return f.articles, f.listErr
}

func (f *fakeStore) GetArticle(_ context.Context, _, viewer string) (domain.Article, error) {
	f.viewer = viewer
	return f.article, f.getErr
}

func (f *fakeStore) CreateArticle(_ context.Context, in domain.NewArticle) (string, error) {
	f.created = &in
	return articleID, f.createErr
}

func (f *fakeStore) ListSeries(_ context.Context, q domain.SeriesQuery) ([]domain.SeriesSummary, error) {
	f.lastSeriesQ = q
	return f.series, f.listErr
}

func (f *fakeStore) GetSeries(context.Context, string) (domain.Series, error) {
	return f.oneSeries, f.seriesErr
}

type fakeLimiter struct {
	allow bool
	err   error
	hits  int
	scope string
}

func (l *fakeLimiter) Hit(_ context.Context, scope, _ string, _ int, _ time.Duration) (bool, time.Duration, error) {
	l.hits++
	l.scope = scope
	return l.allow, 7 * time.Second, l.err
}

func codec(t *testing.T) *cursor.Codec {
	t.Helper()
	c, err := cursor.New([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func newSvc(t *testing.T, st *fakeStore, lim *fakeLimiter) *Service {
	t.Helper()
	if lim == nil {
		lim = &fakeLimiter{allow: true}
	}
	return New(st, codec(t), lim)
}

func ptr[T any](v T) *T { return &v }

func code(err error) apperr.Code {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return ae.Code
	}
	return ""
}

func summaries(n int) []domain.ArticleSummary {
	out := make([]domain.ArticleSummary, n)
	for i := range out {
		out[i] = domain.ArticleSummary{ID: "00000000-0000-7000-8000-00000000010" + string(rune('0'+i)), PublishedAt: when.Add(-time.Duration(i) * time.Hour)}
	}
	return out
}

func TestListArticlesPagesWithASignedCursor(t *testing.T) {
	st := &fakeStore{articles: summaries(3)}
	svc := newSvc(t, st, nil)
	page, err := svc.ListArticles(t.Context(), ArticleListRequest{Topic: ptr("payments"), Limit: ptr(2)})
	if err != nil || len(page.Items) != 2 || page.NextCursor == "" {
		t.Fatalf("page %+v, %v", page, err)
	}
	if st.lastQuery.Limit != 3 || st.lastQuery.Topic != "payments" || st.lastQuery.AfterPublished != nil {
		t.Fatalf("query %+v", st.lastQuery)
	}
	st.articles = summaries(1)
	next, err := svc.ListArticles(t.Context(), ArticleListRequest{Topic: ptr("payments"), Limit: ptr(2), Cursor: page.NextCursor})
	if err != nil || next.NextCursor != "" || len(next.Items) != 1 {
		t.Fatalf("next %+v, %v", next, err)
	}
	if st.lastQuery.AfterID != page.Items[1].ID || !st.lastQuery.AfterPublished.Equal(page.Items[1].PublishedAt) {
		t.Fatalf("keyset %+v", st.lastQuery)
	}
}

func TestListArticlesRejectsACursorOfOtherFilters(t *testing.T) {
	st := &fakeStore{articles: summaries(3)}
	svc := newSvc(t, st, nil)
	page, err := svc.ListArticles(t.Context(), ArticleListRequest{Limit: ptr(1)})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.ListArticles(t.Context(), ArticleListRequest{Level: ptr("advanced"), Cursor: page.NextCursor})
	if code(err) != apperr.CodeInvalidCursor {
		t.Fatalf("err %v", err)
	}
}

func TestListArticlesRejectsBadCursorPayloads(t *testing.T) {
	c := codec(t)
	hash := articleHash(domain.ArticleQuery{})
	for name, p := range map[string]cursor.Payload{
		"bad id":   {Sort: when.Format(time.RFC3339Nano), ID: "nope", FilterHash: hash},
		"bad time": {Sort: "yesterday", ID: articleID, FilterHash: hash},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := newSvc(t, &fakeStore{}, nil).ListArticles(t.Context(), ArticleListRequest{Cursor: c.Encode(p)})
			if code(err) != apperr.CodeInvalidCursor {
				t.Fatalf("err %v", err)
			}
		})
	}
}

func TestListArticlesValidatesInput(t *testing.T) {
	for name, req := range map[string]ArticleListRequest{
		"empty q":      {Q: ptr("")},
		"long q":       {Q: ptr(strings.Repeat("a", 201))},
		"nul q":        {Q: ptr("a\x00b")},
		"invalid utf8": {Q: ptr("\xff")},
		"limit 0":      {Limit: ptr(0)},
		"limit 51":     {Limit: ptr(51)},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := newSvc(t, &fakeStore{}, nil).ListArticles(t.Context(), req)
			if code(err) != apperr.CodeValidation {
				t.Fatalf("err %v", err)
			}
		})
	}
}

func TestListArticlesPassesFiltersAndStoreErrors(t *testing.T) {
	st := &fakeStore{listErr: errDB}
	_, err := newSvc(t, st, nil).ListArticles(t.Context(), ArticleListRequest{
		Q: ptr("mobile money"), Type: ptr("tutorial"), Level: ptr("beginner"), Lang: ptr("fr"),
	})
	if !errors.Is(err, errDB) {
		t.Fatalf("err %v", err)
	}
	q := st.lastQuery
	if q.Q != "mobile money" || q.Kind != "tutorial" || q.Level != "beginner" || q.Lang != "fr" || q.Limit != 21 {
		t.Fatalf("query %+v", q)
	}
}

func TestListArticlesEmptyPageHasNoNilItems(t *testing.T) {
	page, err := newSvc(t, &fakeStore{}, nil).ListArticles(t.Context(), ArticleListRequest{})
	if err != nil || page.Items == nil || len(page.Items) != 0 {
		t.Fatalf("page %+v, %v", page, err)
	}
}

func TestGetArticlePassesTheViewer(t *testing.T) {
	st := &fakeStore{article: domain.Article{ID: articleID}}
	svc := newSvc(t, st, nil)
	if _, err := svc.GetArticle(t.Context(), anonymous, articleID); err != nil || st.viewer != "" {
		t.Fatalf("anonymous viewer %q, %v", st.viewer, err)
	}
	if _, err := svc.GetArticle(t.Context(), alice, articleID); err != nil || st.viewer != aliceID {
		t.Fatalf("member viewer %q, %v", st.viewer, err)
	}
}

func TestGetArticleNotFoundIsOneProblem(t *testing.T) {
	_, err := newSvc(t, &fakeStore{getErr: domain.ErrNotFound}, nil).GetArticle(t.Context(), alice, articleID)
	if code(err) != apperr.CodeNotFound {
		t.Fatalf("err %v", err)
	}
	_, err = newSvc(t, &fakeStore{getErr: errDB}, nil).GetArticle(t.Context(), alice, articleID)
	if code(err) != apperr.CodeInternal {
		t.Fatalf("err %v", err)
	}
}

func goodCreate() CreateRequest {
	return CreateRequest{
		Kind: "tutorial", Title: "  Intégrer Mobile Money  ", BodyMD: "# Étape 1\n\n  texte d'origine  \n",
		Lang: "fr", Level: "beginner", Tags: []string{"mobile-money", "api", "mobile-money"},
		Links: []domain.LinkRef{{Type: "question", ID: targetID}, {Type: "question", ID: targetID}},
	}
}

func TestCreatePublishesAndKeepsTheOriginalText(t *testing.T) {
	st := &fakeStore{article: domain.Article{ID: articleID, Status: "published"}}
	lim := &fakeLimiter{allow: true}
	got, err := newSvc(t, st, lim).CreateArticle(t.Context(), alice, goodCreate())
	if err != nil || got.ID != articleID {
		t.Fatalf("got %+v, %v", got, err)
	}
	c := st.created
	if c.AuthorID != aliceID || c.Status != domain.StatusPublished {
		t.Fatalf("created %+v", c)
	}
	if c.Title != goodCreate().Title || c.BodyMD != goodCreate().BodyMD {
		t.Fatalf("text altered: %q %q", c.Title, c.BodyMD)
	}
	if strings.Join(c.Tags, ",") != "mobile-money,api" || len(c.Links) != 1 {
		t.Fatalf("dedupe tags %v links %v", c.Tags, c.Links)
	}
	if st.viewer != aliceID || lim.hits != 1 || lim.scope != "write:user" {
		t.Fatalf("viewer %q hits %d scope %q", st.viewer, lim.hits, lim.scope)
	}
}

func TestCreateCarriesOptionalFields(t *testing.T) {
	st := &fakeStore{}
	req := goodCreate()
	req.Summary, req.Topic, req.Status = ptr("Résumé"), ptr("payments"), ptr("draft")
	if _, err := newSvc(t, st, nil).CreateArticle(t.Context(), unverified, req); err != nil {
		t.Fatal(err)
	}
	if st.created.Summary != "Résumé" || st.created.Topic != "payments" || st.created.Status != domain.StatusDraft {
		t.Fatalf("created %+v", st.created)
	}
}

func TestCreateAuthorization(t *testing.T) {
	tests := []struct {
		name string
		sub  subject.Subject
		stat *string
		want apperr.Code
	}{
		{"anonymous publish", anonymous, nil, apperr.CodeUnauthenticated},
		{"anonymous draft", anonymous, ptr("draft"), apperr.CodeUnauthenticated},
		{"unverified publish", unverified, nil, apperr.CodeEmailNotVerified},
		{"unverified explicit publish", unverified, ptr("published"), apperr.CodeEmailNotVerified},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st, lim := &fakeStore{}, &fakeLimiter{allow: true}
			req := goodCreate()
			req.Status = tc.stat
			_, err := newSvc(t, st, lim).CreateArticle(t.Context(), tc.sub, req)
			if code(err) != tc.want || st.created != nil || lim.hits != 0 {
				t.Fatalf("err %v created %v hits %d", err, st.created, lim.hits)
			}
		})
	}
}

func TestCreateValidation(t *testing.T) {
	tests := map[string]func(*CreateRequest){
		"blank title":       func(r *CreateRequest) { r.Title = "      " },
		"nul title":         func(r *CreateRequest) { r.Title = "Hello\x00World" },
		"invalid body":      func(r *CreateRequest) { r.BodyMD = "\xff\xfe" },
		"blank body":        func(r *CreateRequest) { r.BodyMD = " \n " },
		"nul summary":       func(r *CreateRequest) { r.Summary = ptr("a\x00") },
		"unknown status":    func(r *CreateRequest) { r.Status = ptr("archived") },
		"series link":       func(r *CreateRequest) { r.Links = []domain.LinkRef{{Type: "series", ID: targetID}} },
		"article link":      func(r *CreateRequest) { r.Links = []domain.LinkRef{{Type: "article", ID: targetID}} },
		"bad link id":       func(r *CreateRequest) { r.Links = []domain.LinkRef{{Type: "project", ID: "x"}} },
		"no tags":           func(r *CreateRequest) { r.Tags = nil },
		"six distinct tags": func(r *CreateRequest) { r.Tags = []string{"a", "b", "c", "d", "e", "f"} },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			st, lim := &fakeStore{}, &fakeLimiter{allow: true}
			req := goodCreate()
			mutate(&req)
			_, err := newSvc(t, st, lim).CreateArticle(t.Context(), alice, req)
			if code(err) != apperr.CodeValidation || st.created != nil || lim.hits != 0 {
				t.Fatalf("err %v created %v hits %d", err, st.created, lim.hits)
			}
		})
	}
}

func TestCreateRateLimitFailsClosed(t *testing.T) {
	st := &fakeStore{}
	_, err := newSvc(t, st, &fakeLimiter{allow: false}).CreateArticle(t.Context(), alice, goodCreate())
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != apperr.CodeRateLimited || ae.RetryAfter != 7*time.Second || st.created != nil {
		t.Fatalf("err %v", err)
	}
	_, err = newSvc(t, st, &fakeLimiter{err: errDB}).CreateArticle(t.Context(), alice, goodCreate())
	if code(err) != apperr.CodeInternal || st.created != nil {
		t.Fatalf("err %v", err)
	}
}

func TestCreateMapsStoreErrors(t *testing.T) {
	tests := []struct {
		err   error
		want  apperr.Code
		field string
	}{
		{domain.ErrUnknownTopic, apperr.CodeValidation, "topic"},
		{domain.ErrLinkTarget, apperr.CodeValidation, "links"},
		{errDB, apperr.CodeInternal, ""},
	}
	for _, tc := range tests {
		t.Run(tc.err.Error(), func(t *testing.T) {
			_, err := newSvc(t, &fakeStore{createErr: tc.err}, nil).CreateArticle(t.Context(), alice, goodCreate())
			var ae *apperr.Error
			if !errors.As(err, &ae) || ae.Code != tc.want {
				t.Fatalf("err %v", err)
			}
			if tc.field != "" && (len(ae.Violations) != 1 || ae.Violations[0].Field != tc.field) {
				t.Fatalf("violations %+v", ae.Violations)
			}
		})
	}
}

func TestCreateReadBackFailureIsInternal(t *testing.T) {
	_, err := newSvc(t, &fakeStore{getErr: domain.ErrNotFound}, nil).CreateArticle(t.Context(), alice, goodCreate())
	if code(err) != apperr.CodeInternal {
		t.Fatalf("err %v", err)
	}
}

func TestListSeriesPages(t *testing.T) {
	st := &fakeStore{series: []domain.SeriesSummary{
		{ID: seriesID, PublishedAt: when}, {ID: articleID, PublishedAt: when.Add(-time.Hour)},
	}}
	svc := newSvc(t, st, nil)
	page, err := svc.ListSeries(t.Context(), SeriesListRequest{Lang: ptr("fr"), Topic: ptr("ai"), Level: ptr("beginner"), Limit: ptr(1)})
	if err != nil || len(page.Items) != 1 || page.NextCursor == "" {
		t.Fatalf("page %+v, %v", page, err)
	}
	q := st.lastSeriesQ
	if q.Lang != "fr" || q.Topic != "ai" || q.Level != "beginner" || q.Limit != 2 {
		t.Fatalf("query %+v", q)
	}
	st.series = nil
	next, err := svc.ListSeries(t.Context(), SeriesListRequest{Lang: ptr("fr"), Topic: ptr("ai"), Level: ptr("beginner"), Limit: ptr(1), Cursor: page.NextCursor})
	if err != nil || next.Items == nil || st.lastSeriesQ.AfterID != seriesID {
		t.Fatalf("next %+v, %v, %+v", next, err, st.lastSeriesQ)
	}
}

func TestListSeriesErrors(t *testing.T) {
	if _, err := newSvc(t, &fakeStore{}, nil).ListSeries(t.Context(), SeriesListRequest{Limit: ptr(99)}); code(err) != apperr.CodeValidation {
		t.Fatalf("limit err %v", err)
	}
	if _, err := newSvc(t, &fakeStore{}, nil).ListSeries(t.Context(), SeriesListRequest{Cursor: "garbage"}); code(err) != apperr.CodeInvalidCursor {
		t.Fatalf("cursor err %v", err)
	}
	if _, err := newSvc(t, &fakeStore{listErr: errDB}, nil).ListSeries(t.Context(), SeriesListRequest{}); !errors.Is(err, errDB) {
		t.Fatalf("store err %v", err)
	}
}

func TestGetSeries(t *testing.T) {
	got, err := newSvc(t, &fakeStore{oneSeries: domain.Series{ID: seriesID}}, nil).GetSeries(t.Context(), seriesID)
	if err != nil || got.ID != seriesID {
		t.Fatalf("got %+v, %v", got, err)
	}
	if _, err := newSvc(t, &fakeStore{seriesErr: domain.ErrNotFound}, nil).GetSeries(t.Context(), seriesID); code(err) != apperr.CodeNotFound {
		t.Fatalf("err %v", err)
	}
	if _, err := newSvc(t, &fakeStore{seriesErr: errDB}, nil).GetSeries(t.Context(), seriesID); code(err) != apperr.CodeInternal {
		t.Fatalf("err %v", err)
	}
}

func TestFilterHashesAreUnambiguous(t *testing.T) {
	a := articleHash(domain.ArticleQuery{Q: "a", Kind: "b"})
	b := articleHash(domain.ArticleQuery{Q: "ab"})
	if a == b {
		t.Fatal("article hash collides")
	}
	if seriesHash(domain.SeriesQuery{Topic: "a", Level: "b"}) == seriesHash(domain.SeriesQuery{Topic: "ab"}) {
		t.Fatal("series hash collides")
	}
}
