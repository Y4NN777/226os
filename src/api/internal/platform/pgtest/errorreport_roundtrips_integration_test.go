//go:build integration

package pgtest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	discoverrepo "github.com/Y4NN777/226os/src/api/internal/discover/repository"
)

// Issue #117 (note-404-round-trips): reporting on an invisible resource used to cost four round trips
// (two attempts of insert then select). The insert now answers visibility itself, so the cost is the
// number of statements a case needs and no more: one for a fresh report and for a 404, two for a
// duplicate of an open report (the insert, then the select of the existing one).
func TestErrorReportRoundTrips(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	conn := migratedDB(ctx, t)
	seedCatalogue(ctx, t, conn)
	if _, err := conn.Exec(ctx, `INSERT INTO resource_sources (id, resource_id, kind, title, url, checked_at) VALUES
		('`+catSrcBeta+`', '`+catBeta+`', 'official_docs', 'Beta docs', 'https://docs.example/beta', now())`); err != nil {
		t.Fatal(err)
	}
	rec := &recordingDB{db: conn}
	repo := discoverrepo.New(rec)
	report := func(resource, kind string) (domain.ErrorReport, int, error) {
		rec.reset()
		got, err := repo.CreateReport(ctx, domain.NewReport{ResourceID: resource, ReporterID: catUser, Kind: kind, Body: "moved"})
		return got, len(rec.calls), err
	}

	for _, tc := range []struct{ name, id string }{
		{"unknown id", "019a0000-0000-7000-8000-00000000ffff"},
		{"removed resource", catRemoved},
		{"resource without a provider", catOrphan},
		{"resource of a removed provider", catSecret},
	} {
		t.Run(tc.name+" is a 404 in one round trip", func(t *testing.T) {
			if _, n, err := report(tc.id, "outdated"); !errors.Is(err, domain.ErrNotFound) || n != 1 {
				t.Fatalf("err %v after %d statements, want ErrNotFound after 1", err, n)
			}
		})
	}

	var first domain.ErrorReport
	t.Run("a fresh report is one round trip", func(t *testing.T) {
		got, n, err := report(catAlpha, "outdated")
		if err != nil || n != 1 || got.Status != "received" {
			t.Fatalf("%+v, %v after %d statements, want a received report after 1", got, err, n)
		}
		first = got
	})
	t.Run("a duplicate of the open report is the insert and one select", func(t *testing.T) {
		got, n, err := report(catAlpha, "outdated")
		if err != nil || n != 2 || got.ID != first.ID {
			t.Fatalf("%+v, %v after %d statements, want id %s after 2", got, err, n, first.ID)
		}
	})
	t.Run("a resolved report no longer counts as open", func(t *testing.T) {
		if _, err := conn.Exec(ctx, `UPDATE error_reports SET status = 'resolved', handled_by = $1, handled_at = now() WHERE id = $2`, catUser, first.ID); err != nil {
			t.Fatal(err)
		}
		got, n, err := report(catAlpha, "outdated")
		if err != nil || n != 1 || got.ID == first.ID || got.Status != "received" {
			t.Fatalf("%+v, %v after %d statements, want a new report after 1", got, err, n)
		}
	})
	t.Run("a source of another resource stays a mismatch, not a 404", func(t *testing.T) {
		rec.reset()
		_, err := repo.CreateReport(ctx, domain.NewReport{ResourceID: catAlpha, ReporterID: catUser, Kind: "other", Body: "x", SourceID: catSrcBeta})
		if !errors.Is(err, domain.ErrSourceMismatch) || len(rec.calls) != 1 {
			t.Fatalf("err %v after %d statements, want ErrSourceMismatch after 1", err, len(rec.calls))
		}
	})
}

// The plan of the single statement, on the exact SQL and arguments of a real request. ANALYZE executes
// it, so the writes run in a transaction that is rolled back. Read it with -v: the resource is one index probe
// by id, the providers table is a small reference table, and the insert is the only write.
func TestErrorReportStatementPlan(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	conn := migratedDB(ctx, t)
	seedCatalogue(ctx, t, conn)
	rec := &recordingDB{db: conn}
	if _, err := discoverrepo.New(rec).CreateReport(ctx, domain.NewReport{ResourceID: catAlpha, ReporterID: catUser, Kind: "outdated", Body: "moved"}); err != nil {
		t.Fatal(err)
	}
	stmt := rec.first(t)
	if _, err := conn.Exec(ctx, `DELETE FROM error_reports`); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, id string }{{"visible resource", catAlpha}, {"invisible resource", catRemoved}} {
		tx, err := conn.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		args := append([]any{}, stmt.args...)
		args[0] = pgUUID(tc.id)
		rows, err := tx.Query(ctx, "EXPLAIN (ANALYZE, BUFFERS) "+stmt.sql, args...)
		if err != nil {
			t.Fatalf("explain: %v", err)
		}
		var plan strings.Builder
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				t.Fatal(err)
			}
			plan.WriteString(line + "\n")
		}
		rows.Close()
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		t.Logf("EXPLAIN (ANALYZE, BUFFERS), CreateErrorReport, %s:\n%s", tc.name, plan.String())
	}
}
