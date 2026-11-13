package app

import (
	"net/http"

	"github.com/Y4NN777/226os/src/api/internal/discover/transport"
	identitytransport "github.com/Y4NN777/226os/src/api/internal/identity/transport"
	learntransport "github.com/Y4NN777/226os/src/api/internal/learn/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
	searchtransport "github.com/Y4NN777/226os/src/api/internal/search/transport"
)

// Server is the ONLY production apigen.ServerInterface. Every operation answers
// 501 until its slice implements it by overriding the method on this type;
// no other constructor of a ServerInterface may exist in production code.
// Besides the operations it exposes one accessor, Mailer, for the session mail flows (S1).
type Server struct {
	apigen.Unimplemented
	mailer    mailer.Mailer
	catalogue *transport.Handler
	reports   *transport.Handler
	search    *searchtransport.Handler
	auth      *identitytransport.Handler
	learn     *learntransport.Handler
}

// Option configures NewServer.
type Option func(*Server)

// WithMailer sets the mail port the operations use.
func WithMailer(m mailer.Mailer) Option { return func(s *Server) { s.mailer = m } }

// WithCatalogue sets the published-catalogue reads. Without it those routes stay 501.
func WithCatalogue(h *transport.Handler) Option { return func(s *Server) { s.catalogue = h } }

// WithErrorReports sets the member error report. Without it POST /resources/{resourceId}/error-reports stays 501.
func WithErrorReports(h *transport.Handler) Option { return func(s *Server) { s.reports = h } }

// WithSearch sets the search read. Without it GET /search stays unimplemented.
func WithSearch(h *searchtransport.Handler) Option { return func(s *Server) { s.search = h } }

// WithIdentity sets the S1 auth operations. Without it those routes stay 501.
func WithIdentity(h *identitytransport.Handler) Option { return func(s *Server) { s.auth = h } }

// WithLearn sets the S4 Learn operations. Without it those routes stay 501.
func WithLearn(h *learntransport.Handler) Option { return func(s *Server) { s.learn = h } }

// NewServer returns the production server with all operations unimplemented.
// Without WithMailer its mailer is mailer.Disabled.
func NewServer(opts ...Option) *Server {
	s := &Server{mailer: mailer.Disabled{}}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Mailer is the mail port: operations send mail through it and never import net/smtp.
func (s *Server) Mailer() mailer.Mailer { return s.mailer }

// ListResources serves the published catalogue. A server built without WithCatalogue stays unimplemented.
func (s *Server) ListResources(w http.ResponseWriter, r *http.Request, params apigen.ListResourcesParams) {
	if s.catalogue == nil {
		s.Unimplemented.ListResources(w, r, params)
		return
	}
	s.catalogue.ListResources(w, r, params)
}

// GetResource serves one published resource. A server built without WithCatalogue stays unimplemented.
func (s *Server) GetResource(w http.ResponseWriter, r *http.Request, resourceID apigen.ResourceId) {
	if s.catalogue == nil {
		s.Unimplemented.GetResource(w, r, resourceID)
		return
	}
	s.catalogue.GetResource(w, r, resourceID)
}

// ListResourceSources serves provenance for one published resource.
func (s *Server) ListResourceSources(w http.ResponseWriter, r *http.Request, resourceID apigen.ResourceId) {
	if s.catalogue == nil {
		s.Unimplemented.ListResourceSources(w, r, resourceID)
		return
	}
	s.catalogue.ListResourceSources(w, r, resourceID)
}

// CreateResourceErrorReport files a member's error report. A server built without WithErrorReports stays unimplemented.
func (s *Server) CreateResourceErrorReport(w http.ResponseWriter, r *http.Request, resourceID apigen.ResourceId) {
	if s.reports == nil {
		s.Unimplemented.CreateResourceErrorReport(w, r, resourceID)
		return
	}
	s.reports.CreateResourceErrorReport(w, r, resourceID)
}

// Search serves the grouped search. A server built without WithSearch stays unimplemented.
func (s *Server) Search(w http.ResponseWriter, r *http.Request, params apigen.SearchParams) {
	if s.search == nil {
		s.Unimplemented.Search(w, r, params)
		return
	}
	s.search.Search(w, r, params)
}

// GetCsrfToken serves the CSRF token endpoint. A server built without WithIdentity stays unimplemented.
func (s *Server) GetCsrfToken(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.GetCsrfToken(w, r)
		return
	}
	s.auth.GetCsrfToken(w, r)
}

// SignUp serves account creation. A server built without WithIdentity stays unimplemented.
func (s *Server) SignUp(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.SignUp(w, r)
		return
	}
	s.auth.SignUp(w, r)
}

// SignIn serves session creation. A server built without WithIdentity stays unimplemented.
func (s *Server) SignIn(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.SignIn(w, r)
		return
	}
	s.auth.SignIn(w, r)
}

// SignOut revokes the caller's session. A server built without WithIdentity stays unimplemented.
func (s *Server) SignOut(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.SignOut(w, r)
		return
	}
	s.auth.SignOut(w, r)
}

// GetSession serves the current session state. A server built without WithIdentity stays unimplemented.
func (s *Server) GetSession(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.GetSession(w, r)
		return
	}
	s.auth.GetSession(w, r)
}

// VerifyEmail confirms an email address. A server built without WithIdentity stays unimplemented.
func (s *Server) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.VerifyEmail(w, r)
		return
	}
	s.auth.VerifyEmail(w, r)
}

// RequestPasswordReset serves the reset request. A server built without WithIdentity stays unimplemented.
func (s *Server) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.RequestPasswordReset(w, r)
		return
	}
	s.auth.RequestPasswordReset(w, r)
}

// ConfirmPasswordReset sets the new password. A server built without WithIdentity stays unimplemented.
func (s *Server) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.ConfirmPasswordReset(w, r)
		return
	}
	s.auth.ConfirmPasswordReset(w, r)
}

// GetMe serves the caller's account and own profile. A server built without WithIdentity stays unimplemented.
func (s *Server) GetMe(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.GetMe(w, r)
		return
	}
	s.auth.GetMe(w, r)
}

// UpdateMyProfile applies a JSON Merge Patch to the caller's profile. A server built without WithIdentity stays unimplemented.
func (s *Server) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil {
		s.Unimplemented.UpdateMyProfile(w, r)
		return
	}
	s.auth.UpdateMyProfile(w, r)
}

// GetProfile serves a visibility-filtered public profile. A server built without WithIdentity stays unimplemented.
func (s *Server) GetProfile(w http.ResponseWriter, r *http.Request, username apigen.Username) {
	if s.auth == nil {
		s.Unimplemented.GetProfile(w, r, username)
		return
	}
	s.auth.GetProfile(w, r, username)
}

// ListLearnArticles serves published articles. A server built without WithLearn stays unimplemented.
func (s *Server) ListLearnArticles(w http.ResponseWriter, r *http.Request, params apigen.ListLearnArticlesParams) {
	if s.learn == nil {
		s.Unimplemented.ListLearnArticles(w, r, params)
		return
	}
	s.learn.ListLearnArticles(w, r, params)
}

// CreateLearnArticle saves a member's article. A server built without WithLearn stays unimplemented.
func (s *Server) CreateLearnArticle(w http.ResponseWriter, r *http.Request, params apigen.CreateLearnArticleParams) {
	if s.learn == nil {
		s.Unimplemented.CreateLearnArticle(w, r, params)
		return
	}
	s.learn.CreateLearnArticle(w, r, params)
}

// GetLearnArticle serves one article. A server built without WithLearn stays unimplemented.
func (s *Server) GetLearnArticle(w http.ResponseWriter, r *http.Request, articleID apigen.ArticleId) {
	if s.learn == nil {
		s.Unimplemented.GetLearnArticle(w, r, articleID)
		return
	}
	s.learn.GetLearnArticle(w, r, articleID)
}

// ListTrainingSeries serves published series. A server built without WithLearn stays unimplemented.
func (s *Server) ListTrainingSeries(w http.ResponseWriter, r *http.Request, params apigen.ListTrainingSeriesParams) {
	if s.learn == nil {
		s.Unimplemented.ListTrainingSeries(w, r, params)
		return
	}
	s.learn.ListTrainingSeries(w, r, params)
}

// GetTrainingSeries serves one series with its lessons. A server built without WithLearn stays unimplemented.
func (s *Server) GetTrainingSeries(w http.ResponseWriter, r *http.Request, seriesID apigen.SeriesId) {
	if s.learn == nil {
		s.Unimplemented.GetTrainingSeries(w, r, seriesID)
		return
	}
	s.learn.GetTrainingSeries(w, r, seriesID)
}
