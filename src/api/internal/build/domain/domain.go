// Package domain holds the Build types (S5, US-BUILD-001): projects referenced by a public GitHub
// repository, their contribution needs, and the GitHub URL rule.
package domain

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Visibility values of a project (projects_visibility_ck).
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

// Errors the store and the GitHub adapter return; the service maps them to problems.
var (
	// ErrNotFound is a project that is missing, removed, or private to someone else.
	ErrNotFound = errors.New("build: not found")
	// ErrDuplicateURL is a GitHub URL already used by a project that is not removed.
	ErrDuplicateURL = errors.New("build: github url already registered")
	// ErrLinkTarget is a link whose target is missing, unpublished or not public.
	ErrLinkTarget = errors.New("build: link target not usable")
	// ErrRepoNotFound is a repository that does not exist, is not public, or was renamed.
	ErrRepoNotFound = errors.New("build: repository not found or not public")
	// ErrUpstream is GitHub being unreachable, slow, rate limited or answering an error.
	ErrUpstream = errors.New("build: github unavailable")
	// ErrDuplicateApplication is a second application to the same need that cannot be reopened.
	ErrDuplicateApplication = errors.New("build: application already exists")
	// ErrStateConflict is a transition from a state that does not allow it.
	ErrStateConflict = errors.New("build: state conflict")
)

// Need and application status values (contribution_needs_status_ck, applications_status_ck).
const (
	NeedOpen             = "open"
	ApplicationPending   = "pending"
	ApplicationAccepted  = "accepted"
	ApplicationDeclined  = "declined"
	ApplicationWithdrawn = "withdrawn"
)

// LinkTypes a project link may name: the pairs content_relations allows with a project
// (question -> project, resource -> project, project -> article).
var LinkTypes = []string{"question", "resource", "article"}

// LinkRef is one typed link of a write body.
type LinkRef struct {
	Type, ID string
}

// Related is one resolved, visible link of a read.
type Related struct {
	Type, ID, Title, Lang string
}

// Owner is the byline of a project: username and display name only.
type Owner struct {
	ID, Username, DisplayName string
}

// Repo is what the GitHub API says about a public repository.
type Repo struct {
	Name, Description string
}

// Project is one project as a read returns it.
type Project struct {
	ID, Name, GitHubURL, Description, Lang, Visibility string
	Technologies                                       []string
	Owner                                              Owner
	OpenNeeds                                          int
	Related                                            []Related
	PublishedAt, CreatedAt, UpdatedAt                  time.Time
}

// ProjectQuery is one listProjects page request. Limit is the page size plus one.
type ProjectQuery struct {
	ViewerID, Q, Technology, Skill string
	OpenNeeds                      bool
	AfterPublished                 *time.Time
	AfterID                        string
	Limit                          int
}

// NewProject is a validated createProject.
type NewProject struct {
	OwnerID, Name, Summary, Description, Lang, GitHubURL, Visibility string
	Technologies                                                     []string
	Links                                                            []LinkRef
}

// ProjectPatch is a validated updateProject; nil means unchanged. Links are added, never removed.
type ProjectPatch struct {
	Description, Visibility *string
	Technologies            *[]string
	Links                   []LinkRef
}

// Access is what an authorization decision needs to know about a project.
type Access struct {
	OwnerID, Visibility string
}

// Need is one contribution need.
type Need struct {
	ID, ProjectID, Title, Description, Lang, Status string
	Skills                                          []string
	CreatedAt                                       time.Time
}

// NeedQuery is one listContributionNeeds page request. Limit is the page size plus one.
type NeedQuery struct {
	ProjectID    string
	AfterCreated *time.Time
	AfterID      string
	Limit        int
}

// NewNeed is a validated createContributionNeed.
type NewNeed struct {
	ProjectID, CreatedBy, Title, Description, Lang string
	Skills                                         []string
}

var (
	githubOwner = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,38}$`)
	githubRepo  = regexp.MustCompile(`^[a-z0-9._-]{1,100}$`)
)

// NormalizeGitHubURL is the single place of the GitHub URL rule (DATABASE_DESIGN 3.5): https, host
// github.com, exactly owner/repo, no user info, port, query or fragment. It returns the canonical form
// (lower case, no trailing slash, no .git) with its owner and repository, or ok false.
func NormalizeGitHubURL(raw string) (canonical, owner, repo string, ok bool) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.ForceQuery ||
		u.Fragment != "" || u.Opaque != "" || !strings.EqualFold(u.Host, "github.com") {
		return "", "", "", false
	}
	parts := strings.Split(strings.TrimSuffix(strings.ToLower(u.Path), "/"), "/")
	if len(parts) != 3 {
		return "", "", "", false
	}
	owner, repo = parts[1], strings.TrimSuffix(parts[2], ".git")
	if !githubOwner.MatchString(owner) || !githubRepo.MatchString(repo) || repo == "." || repo == ".." ||
		strings.HasSuffix(repo, ".git") {
		return "", "", "", false
	}
	return "https://github.com/" + owner + "/" + repo, owner, repo, true
}

// NeedAccess is what an application decision needs to know about the need and its project.
type NeedAccess struct {
	NeedStatus, OwnerID, Visibility string
}

// NewApplication is a validated createApplication; the applicant is the session subject.
type NewApplication struct {
	NeedID, UserID, Message string
}

// ApplicationAccess is what a decision needs to know, read on the locked application row.
type ApplicationAccess struct {
	ApplicantID, OwnerID, Status string
}

// Application is one application as the project owner or the applicant reads it.
type Application struct {
	ID, NeedID, ProjectID, Message, Status string
	Applicant                              Owner
	CreatedAt                              time.Time
	DecidedAt                              *time.Time
}

// ApplicationQuery is one listProjectApplications page request. Limit is the page size plus one.
type ApplicationQuery struct {
	ProjectID, OwnerID, Status string
	AfterCreated               *time.Time
	AfterID                    string
	Limit                      int
}
