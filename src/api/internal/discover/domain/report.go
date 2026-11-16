package domain

import (
	"errors"
	"time"
)

// ErrSourceMismatch means the named source does not belong to the reported resource (or does not exist).
var ErrSourceMismatch = errors.New("source does not belong to the resource")

// ReportKinds are the kinds a member may report, equal to the error_reports_kind_ck constraint.
var ReportKinds = []string{"broken_link", "outdated", "incorrect", "unsupported_country", "other"}

// Report limits (contract ErrorReportCreate.body).
const MaxReportBody = 2000

// NewReport is one error report to store. SourceID is empty when the report names no source.
type NewReport struct {
	ResourceID, ReporterID, Kind, Body, SourceID string
}

// ErrorReport is a stored report as the reporter sees it. It carries no reporter identity.
type ErrorReport struct {
	ID, ResourceID, Kind, Status string
	CreatedAt                    time.Time
}
