// Package service is the Build use case (US-BUILD-001): projects registered from a public GitHub
// repository and their contribution needs. Only the project owner changes a project or publishes a
// need (C1); a project_maintainers row grants nothing at P0.
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

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	defaultLimit = 20
	maxLimit     = 50
	maxSummary   = 500 // projects_summary_ck

	// Every authenticated write draws on the one write:user budget (ADR-0003, Shared pieces);
	// registering a project also draws on the GitHub import budget, 10 per user per hour.
	writeScope   = "write:user"
	writeLimit   = 60
	writeWindow  = time.Minute
	importScope  = "import:github"
	importLimit  = 10
	importWindow = time.Hour
)

// Store loads and saves Build content. The service owns this port.
//
// UpdateProject and CreateNeed lock the project and call authorize with its owner and visibility
// inside the writing transaction; an error from authorize aborts the write and is returned as is.
type Store interface {
	URLTaken(ctx context.Context, url string) (bool, error)
	CreateProject(ctx context.Context, in domain.NewProject) (string, error)
	UpdateProject(ctx context.Context, id, actorID string, p domain.ProjectPatch, authorize func(domain.Access) error) error
	GetProject(ctx context.Context, id, viewerID string) (domain.Project, error)
	ListProjects(ctx context.Context, q domain.ProjectQuery) ([]domain.Project, error)
	ProjectAccess(ctx context.Context, id string) (domain.Access, error)
	CreateNeed(ctx context.Context, in domain.NewNeed, authorize func(domain.Access) error) (domain.Need, error)
	ListNeeds(ctx context.Context, q domain.NeedQuery) ([]domain.Need, error)
	CreateApplication(ctx context.Context, in domain.NewApplication, authorize func(domain.NeedAccess) error) (string, error)
	DecideApplication(ctx context.Context, id, actorID, status string, authorize func(domain.ApplicationAccess) error) error
	GetApplication(ctx context.Context, id string) (domain.Application, error)
	ListApplications(ctx context.Context, q domain.ApplicationQuery) ([]domain.Application, error)
}

// GitHub reads one public repository. The service owns this port.
type GitHub interface {
	Repo(ctx context.Context, owner, name string) (domain.Repo, error)
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

// ProjectListRequest is listProjects. Nil means unset; nil Limit means 20.
type ProjectListRequest struct {
	Q, Technology, Skill *string
	OpenNeeds            *bool
	Cursor               string
	Limit                *int
}

// CreateProjectRequest is the createProject body. Nil Visibility means public.
type CreateProjectRequest struct {
	GitHubURL, Lang         string
	Description, Visibility *string
	Technologies            []string
	Links                   []domain.LinkRef
}

// UpdateProjectRequest is the updateProject merge patch; nil means unchanged.
type UpdateProjectRequest struct {
	Description, Visibility *string
	Technologies            *[]string
	Links                   []domain.LinkRef
}

// NeedListRequest is listContributionNeeds. Nil Limit means 20.
type NeedListRequest struct {
	Cursor string
	Limit  *int
}

// CreateNeedRequest is the createContributionNeed body.
type CreateNeedRequest struct {
	Title, Description, Lang string
	Skills                   []string
}

// ProjectPage is one project page. NextCursor is empty on the last page.
type ProjectPage struct {
	Items      []domain.Project
	NextCursor string
}

// NeedPage is one need page. NextCursor is empty on the last page.
type NeedPage struct {
	Items      []domain.Need
	NextCursor string
}

// Service is the Build use case.
type Service struct {
	store   Store
	github  GitHub
	cursors Cursors
	limiter Limiter
}

// New returns the Build service.
func New(store Store, github GitHub, cursors Cursors, limiter Limiter) *Service {
	return &Service{store: store, github: github, cursors: cursors, limiter: limiter}
}

// ListProjects returns one page of published projects: public ones plus the caller's own private ones.
func (s *Service) ListProjects(ctx context.Context, sub subject.Subject, req ProjectListRequest) (ProjectPage, error) {
	q, err := textFilter(req.Q)
	if err != nil {
		return ProjectPage{}, err
	}
	size, err := pageSize(req.Limit)
	if err != nil {
		return ProjectPage{}, err
	}
	query := domain.ProjectQuery{
		ViewerID: sub.UserID, Q: q, Technology: deref(req.Technology), Skill: deref(req.Skill),
		OpenNeeds: req.OpenNeeds != nil && *req.OpenNeeds, Limit: size + 1,
	}
	hash := lengthPrefixed("project", query.ViewerID, query.Q, query.Technology, query.Skill, strconv.FormatBool(query.OpenNeeds))
	if req.Cursor != "" {
		at, id, err := s.decode(req.Cursor, hash)
		if err != nil {
			return ProjectPage{}, err
		}
		query.AfterPublished, query.AfterID = &at, id
	}
	rows, err := s.store.ListProjects(ctx, query)
	if err != nil {
		return ProjectPage{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	page := ProjectPage{Items: rows}
	if len(rows) > size {
		page.Items = rows[:size]
		last := page.Items[size-1]
		page.NextCursor = s.encode(last.PublishedAt, last.ID, hash)
	}
	if page.Items == nil {
		page.Items = []domain.Project{}
	}
	return page, nil
}

// GetProject returns a published project; a private one only to its owner (404 otherwise).
func (s *Service) GetProject(ctx context.Context, sub subject.Subject, id string) (domain.Project, error) {
	p, err := s.store.GetProject(ctx, id, sub.UserID)
	if err != nil {
		return domain.Project{}, storeErr(err)
	}
	return p, nil
}

// CreateProject registers a public GitHub repository for the session subject. The URL is normalised
// and checked against the allowlist, then the repository is read from the GitHub API (name and
// description); a duplicate never costs GitHub quota. Both limiters fail CLOSED (writes, ADR-0003).
func (s *Service) CreateProject(ctx context.Context, sub subject.Subject, req CreateProjectRequest) (domain.Project, error) {
	if err := publisher(sub); err != nil {
		return domain.Project{}, err
	}
	url, owner, name, ok := domain.NormalizeGitHubURL(req.GitHubURL)
	if !ok {
		return domain.Project{}, apperr.New(apperr.CodeInvalidGitHubURL, "github_url must be https://github.com/<owner>/<repository>.").
			WithViolations(apperr.Violation{Field: "/github_url", Code: "invalid_format"})
	}
	description, err := optionalText("/description", req.Description)
	if err != nil {
		return domain.Project{}, err
	}
	links, err := validLinks(req.Links)
	if err != nil {
		return domain.Project{}, err
	}
	visibility := domain.VisibilityPublic
	if req.Visibility != nil {
		visibility = *req.Visibility
	}
	if err := s.hit(ctx, writeScope, sub.UserID, writeLimit, writeWindow); err != nil {
		return domain.Project{}, err
	}
	taken, err := s.store.URLTaken(ctx, url)
	if err != nil {
		return domain.Project{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	if taken {
		return domain.Project{}, duplicateURL()
	}
	if err := s.hit(ctx, importScope, sub.UserID, importLimit, importWindow); err != nil {
		return domain.Project{}, err
	}
	repo, err := s.github.Repo(ctx, owner, name)
	switch {
	case errors.Is(err, domain.ErrRepoNotFound):
		return domain.Project{}, apperr.New(apperr.CodeInvalidGitHubURL, "The repository does not exist or is not public.").
			WithViolations(apperr.Violation{Field: "/github_url", Code: "invalid_value"})
	case err != nil:
		return domain.Project{}, apperr.Wrap(apperr.CodeUpstreamUnavailable, err)
	}
	id, err := s.store.CreateProject(ctx, domain.NewProject{
		OwnerID: sub.UserID, Name: repo.Name, Summary: summary(repo), Description: description, Lang: req.Lang,
		GitHubURL: url, Visibility: visibility, Technologies: dedupe(req.Technologies), Links: links,
	})
	if err != nil {
		return domain.Project{}, storeErr(err)
	}
	return s.reread(ctx, id, sub.UserID)
}

// UpdateProject applies a merge patch for the project owner. A non-owner gets 403 on a public project
// and 404 on a private one; the decision is taken on the locked row. github_url never changes.
func (s *Service) UpdateProject(ctx context.Context, sub subject.Subject, id string, req UpdateProjectRequest) (domain.Project, error) {
	if sub.Anonymous() {
		return domain.Project{}, unauthenticated()
	}
	if _, err := optionalText("/description", req.Description); err != nil {
		return domain.Project{}, err
	}
	links, err := validLinks(req.Links)
	if err != nil {
		return domain.Project{}, err
	}
	patch := domain.ProjectPatch{Description: req.Description, Visibility: req.Visibility, Links: links}
	if req.Technologies != nil {
		techs := dedupe(*req.Technologies)
		patch.Technologies = &techs
	}
	if err := s.hit(ctx, writeScope, sub.UserID, writeLimit, writeWindow); err != nil {
		return domain.Project{}, err
	}
	if err := s.store.UpdateProject(ctx, id, sub.UserID, patch, ownerOnly(sub)); err != nil {
		return domain.Project{}, storeErr(err)
	}
	return s.reread(ctx, id, sub.UserID)
}

// ListNeeds returns one page of a project's needs; the project's visibility applies (private: 404 to
// everyone but the owner).
func (s *Service) ListNeeds(ctx context.Context, sub subject.Subject, projectID string, req NeedListRequest) (NeedPage, error) {
	size, err := pageSize(req.Limit)
	if err != nil {
		return NeedPage{}, err
	}
	access, err := s.store.ProjectAccess(ctx, projectID)
	if err != nil {
		return NeedPage{}, storeErr(err)
	}
	if access.Visibility != domain.VisibilityPublic && access.OwnerID != sub.UserID {
		return NeedPage{}, notFound()
	}
	query := domain.NeedQuery{ProjectID: projectID, ViewerID: sub.UserID, Limit: size + 1}
	hash := lengthPrefixed("need", projectID)
	if req.Cursor != "" {
		at, id, err := s.decode(req.Cursor, hash)
		if err != nil {
			return NeedPage{}, err
		}
		query.AfterCreated, query.AfterID = &at, id
	}
	rows, err := s.store.ListNeeds(ctx, query)
	if err != nil {
		return NeedPage{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	page := NeedPage{Items: rows}
	if len(rows) > size {
		page.Items = rows[:size]
		last := page.Items[size-1]
		page.NextCursor = s.encode(last.CreatedAt, last.ID, hash)
	}
	if page.Items == nil {
		page.Items = []domain.Need{}
	}
	return page, nil
}

// CreateNeed publishes a contribution need on a project the subject owns (C1: owner only).
func (s *Service) CreateNeed(ctx context.Context, sub subject.Subject, projectID string, req CreateNeedRequest) (domain.Need, error) {
	if err := publisher(sub); err != nil {
		return domain.Need{}, err
	}
	if !usableText(req.Title) {
		return domain.Need{}, invalid("/title", "invalid_value", "title must be non-blank valid text.")
	}
	if !usableText(req.Description) {
		return domain.Need{}, invalid("/description", "invalid_value", "description must be non-blank valid text.")
	}
	if err := s.hit(ctx, writeScope, sub.UserID, writeLimit, writeWindow); err != nil {
		return domain.Need{}, err
	}
	need, err := s.store.CreateNeed(ctx, domain.NewNeed{
		ProjectID: projectID, CreatedBy: sub.UserID, Title: req.Title, Description: req.Description,
		Lang: req.Lang, Skills: dedupe(req.Skills),
	}, ownerOnly(sub))
	if err != nil {
		return domain.Need{}, storeErr(err)
	}
	return need, nil
}

func (s *Service) reread(ctx context.Context, id, viewerID string) (domain.Project, error) {
	p, err := s.store.GetProject(ctx, id, viewerID)
	if err != nil {
		return domain.Project{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	return p, nil
}

func (s *Service) hit(ctx context.Context, scope, value string, limit int, window time.Duration) error {
	ok, retry, err := s.limiter.Hit(ctx, scope, value, limit, window)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	if !ok {
		return apperr.New(apperr.CodeRateLimited, "Too many requests.").WithRetryAfter(retry)
	}
	return nil
}

// ownerOnly is the C1 rule on the locked project: the owner passes; anyone else gets 404 when the
// project is private (its existence must not leak) and 403 when it is public (ADR-0003).
func ownerOnly(sub subject.Subject) func(domain.Access) error {
	return func(a domain.Access) error {
		if a.OwnerID == sub.UserID {
			return nil
		}
		if a.Visibility != domain.VisibilityPublic {
			return notFound()
		}
		return apperr.New(apperr.CodeForbidden, "Only the project owner can do this.")
	}
}

func publisher(sub subject.Subject) error {
	if sub.Anonymous() {
		return unauthenticated()
	}
	if !sub.EmailVerified {
		return apperr.New(apperr.CodeEmailNotVerified, "Verify your email address to publish.")
	}
	return nil
}

// summary is the GitHub description cut to the column bound, or the repository name when GitHub
// has none (projects.summary is NOT NULL).
func summary(r domain.Repo) string {
	d := strings.TrimSpace(r.Description)
	if d == "" {
		return r.Name
	}
	if utf8.RuneCountInString(d) > maxSummary {
		d = string([]rune(d)[:maxSummary])
	}
	return d
}

func validLinks(in []domain.LinkRef) ([]domain.LinkRef, error) {
	out := make([]domain.LinkRef, 0, len(in))
	for _, l := range in {
		if !slices.Contains(domain.LinkTypes, l.Type) {
			return nil, invalid("/links", "invalid_value", "a project can link to a question, a resource or an article.")
		}
		if !slices.Contains(out, l) {
			out = append(out, l)
		}
	}
	return out, nil
}

func optionalText(field string, v *string) (string, error) {
	if v == nil {
		return "", nil
	}
	if !validText(*v) {
		return "", invalid(field, "invalid_value", field[1:]+" must be valid text.")
	}
	return *v, nil
}

// usableText is valid text that is not only whitespace.
func usableText(s string) bool { return validText(s) && strings.TrimSpace(s) != "" }

// validText refuses what PostgreSQL text cannot hold (NUL, invalid UTF-8) rather than answer 500.
func validText(s string) bool { return utf8.ValidString(s) && !strings.ContainsRune(s, 0) }

func textFilter(q *string) (string, error) {
	if q == nil {
		return "", nil
	}
	if !usableText(*q) {
		return "", invalid("q", "invalid_value", "q must be non-blank valid text.")
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

// lengthPrefixed binds a cursor to its filters; every field is length-prefixed so no value can
// imitate a boundary.
func lengthPrefixed(fields ...string) string {
	var b strings.Builder
	for _, f := range fields {
		b.WriteString(strconv.Itoa(len(f)))
		b.WriteByte(':')
		b.WriteString(f)
	}
	return b.String()
}

func dedupe(in []string) []string {
	out := make([]string, 0, len(in))
	for _, v := range in {
		if !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	return out
}

// storeErr maps store sentinels; an apperr from authorize passes through unchanged.
func storeErr(err error) error {
	var ae *apperr.Error
	switch {
	case errors.As(err, &ae):
		return ae
	case errors.Is(err, domain.ErrNotFound):
		return notFound()
	case errors.Is(err, domain.ErrDuplicateURL):
		return duplicateURL()
	case errors.Is(err, domain.ErrDuplicateApplication):
		return apperr.New(apperr.CodeDuplicateApplication, "You already applied to this need.")
	case errors.Is(err, domain.ErrStateConflict):
		return stateConflict("The application is no longer pending.")
	case errors.Is(err, domain.ErrLinkTarget):
		return invalid("/links", "invalid_value", "each link must point to a published, public item.")
	default:
		return apperr.Wrap(apperr.CodeInternal, err)
	}
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func invalid(field, code, detail string) error {
	return apperr.New(apperr.CodeValidation, detail).WithViolations(apperr.Violation{Field: field, Code: code})
}

func unauthenticated() error {
	return apperr.New(apperr.CodeUnauthenticated, "Authentication required.")
}

func notFound() error { return apperr.New(apperr.CodeNotFound, "Not found.") }

func duplicateURL() error {
	return apperr.New(apperr.CodeStateConflict, "This GitHub repository is already registered.").
		WithViolations(apperr.Violation{Field: "/github_url", Code: "duplicate"})
}

func stateConflict(detail string) error { return apperr.New(apperr.CodeStateConflict, detail) }

func invalidCursor() error {
	return apperr.New(apperr.CodeInvalidCursor, "The cursor is invalid or expired.")
}
