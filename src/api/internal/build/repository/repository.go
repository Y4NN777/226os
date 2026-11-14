// Package repository implements the Build store over sqlc: projects, their technologies and links,
// and contribution needs. It translates driver errors to domain sentinels.
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

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

const (
	queryTimeout = 2 * time.Second
	writeTimeout = 5 * time.Second

	githubURLUnique = "projects_github_url_uq"
)

// Beginner is a DBTX that can also open a transaction (*pgxpool.Pool, *pgx.Conn).
type Beginner interface {
	dbgen.DBTX
	Begin(ctx context.Context) (pgx.Tx, error)
}

// Repository implements the Build store.
type Repository struct {
	db Beginner
	q  *dbgen.Queries
}

// New returns a repository over db.
func New(db Beginner) *Repository { return &Repository{db: db, q: dbgen.New(db)} }

// URLTaken reports whether a project that is not removed already uses url.
func (r *Repository) URLTaken(ctx context.Context, url string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	taken, err := r.q.ProjectURLTaken(ctx, url)
	if err != nil {
		return false, fmt.Errorf("build url taken: %w", err)
	}
	return taken, nil
}

// CreateProject inserts the project, its technologies (creating missing tags) and its links in one
// transaction. A URL registered meanwhile is domain.ErrDuplicateURL; an unusable link target is
// domain.ErrLinkTarget.
func (r *Repository) CreateProject(ctx context.Context, in domain.NewProject) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	owner, err := requiredID(in.OwnerID)
	if err != nil {
		return "", err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("build create begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	id, err := q.InsertProject(ctx, dbgen.InsertProjectParams{
		OwnerID: owner, Name: in.Name, Summary: in.Summary, DescriptionMd: text(in.Description),
		Language: in.Lang, GithubUrl: in.GitHubURL, Visibility: in.Visibility,
	})
	if err != nil {
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" && pe.ConstraintName == githubURLUnique {
			return "", domain.ErrDuplicateURL
		}
		return "", fmt.Errorf("build insert project: %w", err)
	}
	if err := setTechnologies(ctx, q, id, owner, in.Technologies, false); err != nil {
		return "", err
	}
	if err := addLinks(ctx, q, id, owner, in.Links); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("build create commit: %w", err)
	}
	return idString(id), nil
}

// UpdateProject locks the project, lets authorize decide on its owner and visibility, then applies the
// patch in the same transaction. A missing or removed project is domain.ErrNotFound.
func (r *Repository) UpdateProject(ctx context.Context, id, actorID string, p domain.ProjectPatch,
	authorize func(domain.Access) error,
) error {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	pid, err := requiredID(id)
	if err != nil {
		return err
	}
	actor, err := requiredID(actorID)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("build update begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	if err := lockAndAuthorize(ctx, q, pid, authorize); err != nil {
		return err
	}
	params := dbgen.UpdateProjectParams{ID: pid, SetDescription: p.Description != nil}
	if p.Description != nil {
		params.DescriptionMd = text(*p.Description)
	}
	if p.Visibility != nil {
		params.Visibility = text(*p.Visibility)
	}
	if err := q.UpdateProject(ctx, params); err != nil {
		return fmt.Errorf("build update project: %w", err)
	}
	if p.Technologies != nil {
		if err := setTechnologies(ctx, q, pid, actor, *p.Technologies, true); err != nil {
			return err
		}
	}
	if err := addLinks(ctx, q, pid, actor, p.Links); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("build update commit: %w", err)
	}
	return nil
}

// GetProject returns a published project that is public or owned by viewerID (empty for anonymous).
func (r *Repository) GetProject(ctx context.Context, id, viewerID string) (domain.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	pid, err := requiredID(id)
	if err != nil {
		return domain.Project{}, err
	}
	viewer, err := optionalID(viewerID)
	if err != nil {
		return domain.Project{}, err
	}
	row, err := r.q.GetProject(ctx, dbgen.GetProjectParams{ID: pid, ViewerID: viewer})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("build get project: %w", err)
	}
	related, err := r.related(ctx, []pgtype.UUID{pid})
	if err != nil {
		return domain.Project{}, err
	}
	p := project(dbgen.ListProjectsRow(row))
	p.Related = orNone(related[p.ID])
	return p, nil
}

// ListProjects returns published projects that are public or owned by the viewer, newest first.
func (r *Repository) ListProjects(ctx context.Context, q domain.ProjectQuery) ([]domain.Project, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	viewer, err := optionalID(q.ViewerID)
	if err != nil {
		return nil, err
	}
	after, err := optionalID(q.AfterID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListProjects(ctx, dbgen.ListProjectsParams{
		ViewerID: viewer, Q: text(q.Q), Technology: text(q.Technology), Skill: text(q.Skill),
		OpenNeeds: q.OpenNeeds, AfterPublished: timestamp(q.AfterPublished), AfterID: after,
		RowLimit: int32(q.Limit), //nolint:gosec // page size is at most 51
	})
	if err != nil {
		return nil, fmt.Errorf("build list projects: %w", err)
	}
	ids := make([]pgtype.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}
	related, err := r.related(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Project, len(rows))
	for i, row := range rows {
		out[i] = project(row)
		out[i].Related = orNone(related[out[i].ID])
	}
	return out, nil
}

// ProjectAccess returns the owner and visibility of a published project.
func (r *Repository) ProjectAccess(ctx context.Context, id string) (domain.Access, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	pid, err := requiredID(id)
	if err != nil {
		return domain.Access{}, err
	}
	row, err := r.q.ProjectAccess(ctx, pid)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Access{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Access{}, fmt.Errorf("build project access: %w", err)
	}
	return domain.Access{OwnerID: idString(row.OwnerID), Visibility: row.Visibility}, nil
}

// CreateNeed locks the project, lets authorize decide, then inserts the need and its skills (creating
// missing tags) in the same transaction.
func (r *Repository) CreateNeed(ctx context.Context, in domain.NewNeed, authorize func(domain.Access) error) (domain.Need, error) {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	pid, err := requiredID(in.ProjectID)
	if err != nil {
		return domain.Need{}, err
	}
	by, err := requiredID(in.CreatedBy)
	if err != nil {
		return domain.Need{}, err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return domain.Need{}, fmt.Errorf("build need begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	if err := lockAndAuthorize(ctx, q, pid, authorize); err != nil {
		return domain.Need{}, err
	}
	row, err := q.InsertNeed(ctx, dbgen.InsertNeedParams{
		ProjectID: pid, Title: in.Title, DescriptionMd: in.Description, Language: in.Lang, CreatedBy: by,
	})
	if err != nil {
		return domain.Need{}, fmt.Errorf("build insert need: %w", err)
	}
	if err := q.EnsureTags(ctx, dbgen.EnsureTagsParams{Slugs: in.Skills, CreatedBy: by}); err != nil {
		return domain.Need{}, fmt.Errorf("build ensure skill tags: %w", err)
	}
	if err := q.InsertNeedSkills(ctx, dbgen.InsertNeedSkillsParams{NeedID: row.ID, Slugs: in.Skills}); err != nil {
		return domain.Need{}, fmt.Errorf("build need skills: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Need{}, fmt.Errorf("build need commit: %w", err)
	}
	return domain.Need{
		ID: idString(row.ID), ProjectID: in.ProjectID, Title: in.Title, Description: in.Description,
		Lang: in.Lang, Status: row.Status, Skills: in.Skills, CreatedAt: row.CreatedAt.Time,
	}, nil
}

// ListNeeds returns the needs of one project, newest first. Visibility is the caller's check.
func (r *Repository) ListNeeds(ctx context.Context, q domain.NeedQuery) ([]domain.Need, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	pid, err := requiredID(q.ProjectID)
	if err != nil {
		return nil, err
	}
	after, err := optionalID(q.AfterID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListNeeds(ctx, dbgen.ListNeedsParams{
		ProjectID: pid, AfterCreated: timestamp(q.AfterCreated), AfterID: after,
		RowLimit: int32(q.Limit), //nolint:gosec // page size is at most 51
	})
	if err != nil {
		return nil, fmt.Errorf("build list needs: %w", err)
	}
	out := make([]domain.Need, len(rows))
	for i, row := range rows {
		out[i] = domain.Need{
			ID: idString(row.ID), ProjectID: idString(row.ProjectID), Title: row.Title, Description: row.DescriptionMd,
			Lang: row.Language, Status: row.Status, Skills: orEmpty(row.Skills), CreatedAt: row.CreatedAt.Time,
		}
	}
	return out, nil
}

func lockAndAuthorize(ctx context.Context, q *dbgen.Queries, id pgtype.UUID, authorize func(domain.Access) error) error {
	row, err := q.LockProjectAccess(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("build lock project: %w", err)
	}
	return authorize(domain.Access{OwnerID: idString(row.OwnerID), Visibility: row.Visibility})
}

// setTechnologies makes the project's active technologies exactly slugs; replace first drops the
// current active ones (hidden ones stay, they are moderator state).
func setTechnologies(ctx context.Context, q *dbgen.Queries, id, by pgtype.UUID, slugs []string, replace bool) error {
	if replace {
		if err := q.DeleteProjectTags(ctx, id); err != nil {
			return fmt.Errorf("build delete technologies: %w", err)
		}
	}
	if len(slugs) == 0 {
		return nil
	}
	if err := q.EnsureTags(ctx, dbgen.EnsureTagsParams{Slugs: slugs, CreatedBy: by}); err != nil {
		return fmt.Errorf("build ensure technology tags: %w", err)
	}
	if err := q.InsertProjectTags(ctx, dbgen.InsertProjectTagsParams{ProjectID: id, Slugs: slugs}); err != nil {
		return fmt.Errorf("build technologies: %w", err)
	}
	return nil
}

func addLinks(ctx context.Context, q *dbgen.Queries, project, by pgtype.UUID, links []domain.LinkRef) error {
	for _, l := range links {
		target, err := requiredID(l.ID)
		if err != nil {
			return domain.ErrLinkTarget
		}
		var ok bool
		switch l.Type {
		case "question":
			ok, err = q.LinkQuestionToProject(ctx, dbgen.LinkQuestionToProjectParams{TargetID: target, ProjectID: project, CreatedBy: by})
		case "resource":
			ok, err = q.LinkResourceToProject(ctx, dbgen.LinkResourceToProjectParams{TargetID: target, ProjectID: project, CreatedBy: by})
		case "article":
			ok, err = q.LinkProjectToArticleTarget(ctx, dbgen.LinkProjectToArticleTargetParams{TargetID: target, ProjectID: project, CreatedBy: by})
		default:
			return domain.ErrLinkTarget
		}
		if err != nil {
			return fmt.Errorf("build link %s: %w", l.Type, err)
		}
		if !ok {
			return domain.ErrLinkTarget
		}
	}
	return nil
}

func (r *Repository) related(ctx context.Context, ids []pgtype.UUID) (map[string][]domain.Related, error) {
	out := map[string][]domain.Related{}
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.q.ProjectRelated(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("build related: %w", err)
	}
	for _, row := range rows {
		id := idString(row.ProjectID)
		out[id] = append(out[id], domain.Related{Type: row.Kind, ID: idString(row.TargetID), Title: row.Title, Lang: row.Lang})
	}
	return out, nil
}

func project(row dbgen.ListProjectsRow) domain.Project {
	return domain.Project{
		ID: idString(row.ID), Name: row.Name, GitHubURL: row.GithubUrl, Description: row.Description,
		Lang: row.Language, Visibility: row.Visibility, Technologies: orEmpty(row.Technologies),
		Owner:       domain.Owner{ID: idString(row.OwnerID), Username: row.Username, DisplayName: row.DisplayName},
		OpenNeeds:   int(row.OpenNeeds),
		PublishedAt: row.PublishedAt.Time, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
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
		return pgtype.UUID{}, fmt.Errorf("build id: %w", err)
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

func orNone(r []domain.Related) []domain.Related {
	if r == nil {
		return []domain.Related{}
	}
	return r
}
