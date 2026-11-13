// Package domain holds the catalogue read model. It has no database or HTTP types.
package domain

import (
	"errors"
	"time"
)

// ErrNotFound means the resource is missing, removed, or has no published provider.
var ErrNotFound = errors.New("catalogue resource not found")

// Query is one published-catalogue page. Empty filters are unset. Limit is the
// number of rows to fetch, including the extra row that detects a next page.
type Query struct {
	Q, Country, Category, Type, SDK, Provider, Doclang string
	AfterName, AfterID                                 string
	Limit                                              int
}

// Resource is one published catalogue entry.
type Resource struct {
	ID, Slug, Name, Type, Category, Summary string
	DocLanguages                            []string
	DocsURL                                 *string
	Provider                                Provider
	Countries                               []string
	SDKs                                    []SDK
	Related                                 []Link
	SourceCount                             int
	VerificationStatus                      string
	LastVerifiedAt                          *time.Time
	UpdatedAt                               time.Time
}

// Provider is the published provider of a resource.
type Provider struct {
	Name, Slug string
	URL        *string
}

// SDK is one official SDK that has a URL.
type SDK struct {
	Language, Name, URL string
}

// Link is a related published project, article, or question.
type Link struct {
	ID, Type, Title, Lang string
}

// Source is one provenance row of a published resource.
type Source struct {
	ID, Kind, Title, URL, Status string
	VerifiedAt                   *time.Time
}

// SourceStatus maps stored facts to the public status.
//
// verified is returned only when this source is the resource's current verified
// source and the resource itself is verified. A claim that is not that source
// is claimed. Everything else is unverified.
//
// stale is never returned: no freshness window is defined, and inventing one
// would mark a source stale without a decision.
func SourceStatus(kind, resourceStatus, sourceID, verifiedSourceID string, checkedAt *time.Time) (string, *time.Time) {
	if sourceID != "" && sourceID == verifiedSourceID && resourceStatus == "verified" {
		return "verified", checkedAt
	}
	if kind == "claim" {
		return "claimed", nil
	}
	return "unverified", nil
}
