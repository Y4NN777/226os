package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// CreateApplication locks the need, lets authorize decide on its status and its project, then inserts
// the application. When the applicant already has one for this need, a withdrawn application is
// reopened with the new message and anything else is domain.ErrDuplicateApplication. The unique
// (user, need) pair makes a concurrent second click wait for the first and then fail the same way.
func (r *Repository) CreateApplication(ctx context.Context, in domain.NewApplication, authorize func(domain.NeedAccess) error) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	need, err := requiredID(in.NeedID)
	if err != nil {
		return "", err
	}
	user, err := requiredID(in.UserID)
	if err != nil {
		return "", err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("build apply begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	row, err := q.LockNeedForApply(ctx, need)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("build lock need: %w", err)
	}
	if err := authorize(domain.NeedAccess{NeedStatus: row.Status, OwnerID: idString(row.OwnerID), Visibility: row.Visibility}); err != nil {
		return "", err
	}
	id, err := q.InsertApplication(ctx, dbgen.InsertApplicationParams{NeedID: need, UserID: user, Message: in.Message, Language: row.Language})
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		id, err = reopen(ctx, q, need, user, in.Message, row.Language)
		if err != nil {
			return "", err
		}
	case err != nil:
		return "", fmt.Errorf("build insert application: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("build apply commit: %w", err)
	}
	return idString(id), nil
}

func reopen(ctx context.Context, q *dbgen.Queries, need, user pgtype.UUID, message, lang string) (pgtype.UUID, error) {
	own, err := q.LockOwnApplication(ctx, dbgen.LockOwnApplicationParams{UserID: user, NeedID: need})
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("build lock own application: %w", err)
	}
	if own.Status != domain.ApplicationWithdrawn {
		return pgtype.UUID{}, domain.ErrDuplicateApplication
	}
	if _, err := q.ReopenApplication(ctx, dbgen.ReopenApplicationParams{Message: message, Language: lang, ID: own.ID}); err != nil {
		return pgtype.UUID{}, fmt.Errorf("build reopen application: %w", err)
	}
	return own.ID, nil
}

// DecideApplication locks the application, lets authorize decide on its applicant, project owner and
// status, then moves it from pending to status in the same transaction.
func (r *Repository) DecideApplication(ctx context.Context, id, actorID, status string, authorize func(domain.ApplicationAccess) error) error {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	aid, err := requiredID(id)
	if err != nil {
		return err
	}
	actor, err := requiredID(actorID)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("build decide begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	row, err := q.LockApplicationForDecision(ctx, aid)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("build lock application: %w", err)
	}
	if err := authorize(domain.ApplicationAccess{ApplicantID: idString(row.UserID), OwnerID: idString(row.OwnerID), Status: row.Status}); err != nil {
		return err
	}
	n, err := q.DecideApplication(ctx, dbgen.DecideApplicationParams{Status: status, ActorID: actor, ID: aid})
	if err != nil {
		return fmt.Errorf("build decide application: %w", err)
	}
	if n == 0 {
		return domain.ErrStateConflict
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("build decide commit: %w", err)
	}
	return nil
}

// GetApplication returns one application. Authorization is the caller's.
func (r *Repository) GetApplication(ctx context.Context, id string) (domain.Application, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	aid, err := requiredID(id)
	if err != nil {
		return domain.Application{}, err
	}
	row, err := r.q.GetApplication(ctx, aid)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Application{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Application{}, fmt.Errorf("build get application: %w", err)
	}
	return application(dbgen.ListProjectApplicationsRow(row)), nil
}

// ListApplications returns the applications of a project owned by q.OwnerID, newest first. The owner
// predicate is in the query, so a wrong owner lists nothing.
func (r *Repository) ListApplications(ctx context.Context, q domain.ApplicationQuery) ([]domain.Application, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	pid, err := requiredID(q.ProjectID)
	if err != nil {
		return nil, err
	}
	owner, err := requiredID(q.OwnerID)
	if err != nil {
		return nil, err
	}
	after, err := optionalID(q.AfterID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListProjectApplications(ctx, dbgen.ListProjectApplicationsParams{
		ProjectID: pid, OwnerID: owner, Status: text(q.Status), AfterCreated: timestamp(q.AfterCreated), AfterID: after,
		RowLimit: int32(q.Limit), //nolint:gosec // page size is at most 51
	})
	if err != nil {
		return nil, fmt.Errorf("build list applications: %w", err)
	}
	out := make([]domain.Application, len(rows))
	for i, row := range rows {
		out[i] = application(row)
	}
	return out, nil
}

func application(row dbgen.ListProjectApplicationsRow) domain.Application {
	a := domain.Application{
		ID: idString(row.ID), NeedID: idString(row.ContributionNeedID), ProjectID: idString(row.ProjectID),
		Message: row.Message, Status: row.Status, CreatedAt: row.CreatedAt.Time,
		Applicant: domain.Owner{ID: idString(row.UserID), Username: row.Username, DisplayName: row.DisplayName},
	}
	if row.DecidedAt.Valid {
		at := row.DecidedAt.Time
		a.DecidedAt = &at
	}
	return a
}
