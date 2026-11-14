// Package repository persists moderation changes and their audit row atomically.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/moderation/domain"
	platformdb "github.com/Y4NN777/226os/src/api/internal/platform/db"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

const writeTimeout = 5 * time.Second

type Beginner interface {
	dbgen.DBTX
	Begin(context.Context) (pgx.Tx, error)
}

type Repository struct{ runner *platformdb.Runner }

func New(db Beginner) *Repository { return &Repository{runner: platformdb.NewRunner(db)} }

func (r *Repository) Hide(ctx context.Context, kind, target, actor, reason, requestID string) error {
	return r.change(ctx, kind, target, actor, reason, requestID, true)
}

func (r *Repository) Restore(ctx context.Context, kind, target, actor, reason, requestID string) error {
	return r.change(ctx, kind, target, actor, reason, requestID, false)
}

func (r *Repository) change(ctx context.Context, kind, target, actor, reason, requestID string, hide bool) error {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	targetID, err := requiredID(target)
	if err != nil {
		return err
	}
	actorID, err := requiredID(actor)
	if err != nil {
		return err
	}
	return r.runner.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := dbgen.New(tx)
		state, mutate, err := targetOps(ctx, q, kind, targetID, hide)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("moderation lock %s: %w", kind, err)
		}
		expected := "published"
		if kind == "relation" {
			expected = "active"
		}
		if !hide {
			expected = "removed"
			if kind == "relation" {
				expected = "hidden"
			}
		}
		if state != expected {
			return domain.ErrStateConflict
		}
		if err := mutate(ctx); err != nil {
			return fmt.Errorf("moderation update %s: %w", kind, err)
		}
		action := kind + ".hide"
		if !hide {
			action = kind + ".restore"
		}
		if err := audit(ctx, q, actorID, action, kind, targetID, reason, requestID); err != nil {
			return fmt.Errorf("moderation audit: %w", err)
		}
		return nil
	})
}

func (r *Repository) ResolveReport(ctx context.Context, report, actor, status, note, requestID string) (out domain.ErrorReport, err error) {
	ctx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	reportID, err := requiredID(report)
	if err != nil {
		return out, err
	}
	actorID, err := requiredID(actor)
	if err != nil {
		return out, err
	}
	err = r.runner.InTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		q := dbgen.New(tx)
		locked, err := q.LockModerationErrorReport(ctx, reportID)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("moderation lock report: %w", err)
		}
		if locked.Status != "received" && locked.Status != "triaged" {
			return domain.ErrStateConflict
		}
		row, err := q.ResolveModerationErrorReport(ctx, dbgen.ResolveModerationErrorReportParams{ID: reportID, ActorID: actorID, Status: status})
		if err != nil {
			return fmt.Errorf("moderation resolve report: %w", err)
		}
		action := "error_report.resolve"
		if status == "rejected" {
			action = "error_report.reject"
		}
		if err := audit(ctx, q, actorID, action, "report", reportID, note, requestID); err != nil {
			return fmt.Errorf("moderation audit report: %w", err)
		}
		out = domain.ErrorReport{ID: idString(row.ID), ResourceID: idString(row.ResourceID), Kind: row.Kind, Status: row.Status, CreatedAt: row.CreatedAt.Time}
		return nil
	})
	return out, err
}

type mutation func(context.Context) error

func targetOps(ctx context.Context, q *dbgen.Queries, kind string, id pgtype.UUID, hide bool) (string, mutation, error) {
	switch kind {
	case "question":
		state, err := q.LockModerationQuestion(ctx, id)
		if hide {
			return state, func(c context.Context) error { return q.HideModerationQuestion(c, id) }, err
		}
		return state, func(c context.Context) error { return q.RestoreModerationQuestion(c, id) }, err
	case "answer":
		state, err := q.LockModerationAnswer(ctx, id)
		if hide {
			return state, func(c context.Context) error { return q.HideModerationAnswer(c, id) }, err
		}
		return state, func(c context.Context) error { return q.RestoreModerationAnswer(c, id) }, err
	case "resource":
		state, err := q.LockModerationResource(ctx, id)
		if hide {
			return state, func(c context.Context) error { return q.HideModerationResource(c, id) }, err
		}
		return state, func(c context.Context) error { return q.RestoreModerationResource(c, id) }, err
	case "project":
		state, err := q.LockModerationProject(ctx, id)
		if hide {
			return state, func(c context.Context) error { return q.HideModerationProject(c, id) }, err
		}
		return state, func(c context.Context) error { return q.RestoreModerationProject(c, id) }, err
	case "article":
		state, err := q.LockModerationArticle(ctx, id)
		if hide {
			return state, func(c context.Context) error { return q.HideModerationArticle(c, id) }, err
		}
		return state, func(c context.Context) error { return q.RestoreModerationArticle(c, id) }, err
	case "relation":
		state, err := q.LockModerationRelation(ctx, id)
		if hide {
			return state, func(c context.Context) error { return q.HideModerationRelation(c, id) }, err
		}
		return state, func(c context.Context) error { return q.RestoreModerationRelation(c, id) }, err
	default:
		return "", nil, domain.ErrNotFound
	}
}

func audit(ctx context.Context, q *dbgen.Queries, actor pgtype.UUID, action, kind string, target pgtype.UUID, reason, requestID string) error {
	return q.InsertModerationAudit(ctx, dbgen.InsertModerationAuditParams{ActorID: actor, Action: action, TargetType: pgtype.Text{String: kind, Valid: true}, TargetID: target, Reason: reason, RequestID: requestID})
}

func requiredID(raw string) (pgtype.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("moderation uuid: %w", err)
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

func idString(id pgtype.UUID) string { return uuid.UUID(id.Bytes).String() }
