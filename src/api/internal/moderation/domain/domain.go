// Package domain holds the moderation state and repository errors.
package domain

import (
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("moderation: target not found")
	ErrStateConflict = errors.New("moderation: state conflict")
)

// ErrorReport is the reporter-safe result of resolving an error report.
type ErrorReport struct {
	ID, ResourceID, Kind, Status string
	CreatedAt                    time.Time
}
