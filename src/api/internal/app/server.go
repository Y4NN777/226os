package app

import (
	"net/http"

	buildtransport "github.com/Y4NN777/226os/src/api/internal/build/transport"
	"github.com/Y4NN777/226os/src/api/internal/discover/transport"
	identitytransport "github.com/Y4NN777/226os/src/api/internal/identity/transport"
	learntransport "github.com/Y4NN777/226os/src/api/internal/learn/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
	searchtransport "github.com/Y4NN777/226os/src/api/internal/search/transport"
	solvetransport "github.com/Y4NN777/226os/src/api/internal/solve/transport"
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
	build     *buildtransport.Handler
	solve     *solvetransport.Handler
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

// WithBuild sets the S5 project and contribution-need operations. Without it those routes stay 501.
func WithBuild(h *buildtransport.Handler) Option { return func(s *Server) { s.build = h } }

// WithSolve sets the S3 question and answer operations. Without it those routes stay 501.
func WithSolve(h *solvetransport.Handler) Option { return func(s *Server) { s.solve = h } }

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

// ListProjects serves published projects. A server built without WithBuild stays unimplemented.
func (s *Server) ListProjects(w http.ResponseWriter, r *http.Request, params apigen.ListProjectsParams) {
	if s.build == nil {
		s.Unimplemented.ListProjects(w, r, params)
		return
	}
	s.build.ListProjects(w, r, params)
}

// CreateProject registers a member's GitHub project. A server built without WithBuild stays unimplemented.
func (s *Server) CreateProject(w http.ResponseWriter, r *http.Request, params apigen.CreateProjectParams) {
	if s.build == nil {
		s.Unimplemented.CreateProject(w, r, params)
		return
	}
	s.build.CreateProject(w, r, params)
}

// GetProject serves one project. A server built without WithBuild stays unimplemented.
func (s *Server) GetProject(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId) {
	if s.build == nil {
		s.Unimplemented.GetProject(w, r, projectID)
		return
	}
	s.build.GetProject(w, r, projectID)
}

// UpdateProject edits the owner's project. A server built without WithBuild stays unimplemented.
func (s *Server) UpdateProject(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId) {
	if s.build == nil {
		s.Unimplemented.UpdateProject(w, r, projectID)
		return
	}
	s.build.UpdateProject(w, r, projectID)
}

// ListContributionNeeds serves a project's needs. A server built without WithBuild stays unimplemented.
func (s *Server) ListContributionNeeds(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId, params apigen.ListContributionNeedsParams) {
	if s.build == nil {
		s.Unimplemented.ListContributionNeeds(w, r, projectID, params)
		return
	}
	s.build.ListContributionNeeds(w, r, projectID, params)
}

// CreateContributionNeed publishes the owner's need. A server built without WithBuild stays unimplemented.
func (s *Server) CreateContributionNeed(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId, params apigen.CreateContributionNeedParams) {
	if s.build == nil {
		s.Unimplemented.CreateContributionNeed(w, r, projectID, params)
		return
	}
	s.build.CreateContributionNeed(w, r, projectID, params)
}

// CreateApplication records a member's application. A server built without WithBuild stays unimplemented.
func (s *Server) CreateApplication(w http.ResponseWriter, r *http.Request, needID apigen.NeedId, params apigen.CreateApplicationParams) {
	if s.build == nil {
		s.Unimplemented.CreateApplication(w, r, needID, params)
		return
	}
	s.build.CreateApplication(w, r, needID, params)
}

// ListProjectApplications serves a project's applications to its owner. A server built without WithBuild stays unimplemented.
func (s *Server) ListProjectApplications(w http.ResponseWriter, r *http.Request, projectID apigen.ProjectId, params apigen.ListProjectApplicationsParams) {
	if s.build == nil {
		s.Unimplemented.ListProjectApplications(w, r, projectID, params)
		return
	}
	s.build.ListProjectApplications(w, r, projectID, params)
}

// DecideApplication accepts, declines or withdraws an application. A server built without WithBuild stays unimplemented.
func (s *Server) DecideApplication(w http.ResponseWriter, r *http.Request, applicationID apigen.ApplicationId) {
	if s.build == nil {
		s.Unimplemented.DecideApplication(w, r, applicationID)
		return
	}
	s.build.DecideApplication(w, r, applicationID)
}

// ListQuestions serves GET /questions. A server built without WithSolve stays unimplemented.
func (s *Server) ListQuestions(w http.ResponseWriter, r *http.Request, params apigen.ListQuestionsParams) {
	if s.solve == nil {
		s.Unimplemented.ListQuestions(w, r, params)
		return
	}
	s.solve.ListQuestions(w, r, params)
}

// CreateQuestion serves POST /questions. A server built without WithSolve stays unimplemented.
func (s *Server) CreateQuestion(w http.ResponseWriter, r *http.Request, params apigen.CreateQuestionParams) {
	if s.solve == nil {
		s.Unimplemented.CreateQuestion(w, r, params)
		return
	}
	s.solve.CreateQuestion(w, r, params)
}

// GetQuestion serves GET /questions/{questionId}. A server built without WithSolve stays unimplemented.
func (s *Server) GetQuestion(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId) {
	if s.solve == nil {
		s.Unimplemented.GetQuestion(w, r, questionID)
		return
	}
	s.solve.GetQuestion(w, r, questionID)
}

// UpdateQuestion serves PATCH /questions/{questionId}. A server built without WithSolve stays unimplemented.
func (s *Server) UpdateQuestion(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId, params apigen.UpdateQuestionParams) {
	if s.solve == nil {
		s.Unimplemented.UpdateQuestion(w, r, questionID, params)
		return
	}
	s.solve.UpdateQuestion(w, r, questionID, params)
}

// ListAnswers serves GET /questions/{questionId}/answers. A server built without WithSolve stays unimplemented.
func (s *Server) ListAnswers(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId, params apigen.ListAnswersParams) {
	if s.solve == nil {
		s.Unimplemented.ListAnswers(w, r, questionID, params)
		return
	}
	s.solve.ListAnswers(w, r, questionID, params)
}

// CreateAnswer serves POST /questions/{questionId}/answers. A server built without WithSolve stays unimplemented.
func (s *Server) CreateAnswer(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId, params apigen.CreateAnswerParams) {
	if s.solve == nil {
		s.Unimplemented.CreateAnswer(w, r, questionID, params)
		return
	}
	s.solve.CreateAnswer(w, r, questionID, params)
}
