package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// maxReportAttempts bounds the insert/select dance when the open report is resolved between the two
// statements; the second attempt then inserts a fresh report.
const maxReportAttempts = 2

// CreateReport files an error report for a published resource, or returns the reporter's open report
// of the same kind. Each statement is its own snapshot (READ COMMITTED), so when a concurrent request
// has just inserted the open report, ON CONFLICT DO NOTHING waits for its commit and the following
// select sees it: two simultaneous identical requests yield exactly one row and the same id.
func (r *Repository) CreateReport(ctx context.Context, in domain.NewReport) (domain.ErrorReport, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	resource, err := requiredID(in.ResourceID)
	if err != nil {
		return domain.ErrorReport{}, domain.ErrNotFound
	}
	reporter, err := requiredID(in.ReporterID)
	if err != nil {
		return domain.ErrorReport{}, fmt.Errorf("error report reporter: %w", err)
	}
	source, err := optionalID(in.SourceID)
	if err != nil {
		return domain.ErrorReport{}, domain.ErrSourceMismatch
	}
	for range maxReportAttempts {
		created, err := r.q.CreateErrorReport(ctx, dbgen.CreateErrorReportParams{
			ReporterID: reporter, Kind: in.Kind, Message: in.Body, SourceID: source, ResourceID: resource,
		})
		if err == nil {
			return report(created.ID, created.ResourceID, created.Kind, created.Status, created.CreatedAt), nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrorReport{}, reportErr(err)
		}
		open, err := r.q.GetOpenErrorReport(ctx, dbgen.GetOpenErrorReportParams{ResourceID: resource, ReporterID: reporter, Kind: in.Kind})
		if err == nil {
			return report(open.ID, open.ResourceID, open.Kind, open.Status, open.CreatedAt), nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrorReport{}, fmt.Errorf("error report open lookup: %w", err)
		}
	}
	return domain.ErrorReport{}, domain.ErrNotFound
}

// reportErr maps the source foreign key violation to a sentinel, so a source of another resource is
// a 422 that names no other resource.
func reportErr(err error) error {
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23503" && pg.ConstraintName == "error_reports_source_fk" {
		return domain.ErrSourceMismatch
	}
	return fmt.Errorf("error report create: %w", err)
}

func report(id, resourceID pgtype.UUID, kind, status string, created pgtype.Timestamptz) domain.ErrorReport {
	return domain.ErrorReport{ID: idString(id), ResourceID: idString(resourceID), Kind: kind, Status: status, CreatedAt: created.Time}
}
