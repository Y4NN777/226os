package repository

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
)

const reporterID = "00000000-0000-7000-8000-0000000000d1"

func newReport() domain.NewReport {
	return domain.NewReport{ResourceID: resID, ReporterID: reporterID, Kind: "outdated", Body: "moved"}
}

func reportRow(status string) []any {
	return row(uid("00000000-0000-7000-8000-0000000000e1"), uid(resID), "outdated", status, when())
}

func TestCreateReportInsertsAndReturnsTheNewReport(t *testing.T) {
	in := newReport()
	in.SourceID = srcID
	got, err := New(&script{t: t, steps: [][][]any{{reportRow("received")}}, errs: []error{nil}}).CreateReport(t.Context(), in)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "00000000-0000-7000-8000-0000000000e1" || got.ResourceID != resID || got.Kind != "outdated" || got.Status != "received" || !got.CreatedAt.Equal(when().Time) {
		t.Fatalf("got %+v", got)
	}
}

func TestCreateReportReturnsTheExistingOpenReportOnDuplicate(t *testing.T) {
	s := &script{t: t, steps: [][][]any{nil, {reportRow("triaged")}}, errs: []error{pgx.ErrNoRows, nil}}
	got, err := New(s).CreateReport(t.Context(), newReport())
	if err != nil || got.Status != "triaged" || s.n != 2 {
		t.Fatalf("got %+v, %v after %d queries; want the open report from the second query", got, err, s.n)
	}
}

// The open report may be resolved between the failed insert and the select: one more attempt.
func TestCreateReportRetriesWhenTheOpenReportVanishes(t *testing.T) {
	s := &script{t: t,
		steps: [][][]any{nil, nil, {reportRow("received")}},
		errs:  []error{pgx.ErrNoRows, pgx.ErrNoRows, nil}}
	got, err := New(s).CreateReport(t.Context(), newReport())
	if err != nil || got.Status != "received" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestCreateReportOfAnInvisibleResourceIsNotFound(t *testing.T) {
	s := &script{t: t, steps: [][][]any{nil, nil, nil, nil}, errs: []error{pgx.ErrNoRows, pgx.ErrNoRows, pgx.ErrNoRows, pgx.ErrNoRows}}
	if _, err := New(s).CreateReport(t.Context(), newReport()); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestCreateReportTranslatesErrors(t *testing.T) {
	boom := errors.New("db down")
	fk := &pgconn.PgError{Code: "23503", ConstraintName: "error_reports_source_fk"}
	otherFK := &pgconn.PgError{Code: "23503", ConstraintName: "error_reports_reporter_fk"}
	cases := []struct {
		name    string
		steps   [][][]any
		errs    []error
		in      domain.NewReport
		target  error
		isPlain bool
	}{
		{"source of another resource", [][][]any{nil}, []error{fk}, newReport(), domain.ErrSourceMismatch, false},
		{"another foreign key is an internal error", [][][]any{nil}, []error{otherFK}, newReport(), otherFK, true},
		{"insert failure", [][][]any{nil}, []error{boom}, newReport(), boom, true},
		{"select failure", [][][]any{nil, nil}, []error{pgx.ErrNoRows, boom}, newReport(), boom, true},
		{"malformed resource id", nil, nil, domain.NewReport{ResourceID: "bad", ReporterID: reporterID}, domain.ErrNotFound, false},
		{"malformed reporter id", nil, nil, domain.NewReport{ResourceID: resID, ReporterID: "bad"}, nil, true},
		{"malformed source id", nil, nil, domain.NewReport{ResourceID: resID, ReporterID: reporterID, SourceID: "bad"}, domain.ErrSourceMismatch, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(&script{t: t, steps: tc.steps, errs: tc.errs}).CreateReport(t.Context(), tc.in)
			if err == nil {
				t.Fatal("want an error")
			}
			if tc.target != nil && !errors.Is(err, tc.target) {
				t.Fatalf("err = %v, want %v", err, tc.target)
			}
			var pg *pgconn.PgError
			if !tc.isPlain && errors.As(err, &pg) {
				t.Fatalf("driver error %v leaked past the repository", err)
			}
		})
	}
}
