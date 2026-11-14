package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// ListProfiles runs the people-directory query (E5). Visibility is decided in
// SQL for the audience, so a hidden value is never selected: Languages and
// Skills come back nil when hidden, Country and GitHub nil when hidden or
// unset. q.Limit is the page size plus the look-ahead row, at most 51.
func (r *Repository) ListProfiles(ctx context.Context, q domain.DirectoryQuery) ([]domain.DirectoryEntry, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	afterID, err := parseOptionalUUID(q.AfterID)
	if err != nil {
		return nil, domain.ErrNotFound
	}
	rows, err := r.q.ListDirectoryProfiles(ctx, dbgen.ListDirectoryProfilesParams{
		Member:        q.Member,
		Country:       optionalText(q.Country),
		Language:      optionalText(q.Language),
		Skill:         optionalText(q.Skill),
		AfterUsername: optionalText(q.AfterUsername),
		AfterID:       afterID,
		RowLimit:      int32(q.Limit), //nolint:gosec // the service bounds the page size to 51
	})
	if err != nil {
		return nil, fmt.Errorf("identity list profiles: %w", err)
	}
	out := make([]domain.DirectoryEntry, len(rows))
	for i, row := range rows {
		e := domain.DirectoryEntry{
			UserID: uuidString(row.UserID), Username: row.Username, DisplayName: row.DisplayName,
			Country: emptyToNil(row.CountryCode), GitHub: emptyToNil(row.GithubUsername),
		}
		if row.LanguagesVisible {
			e.Languages = nonEmpty(row.Languages)
		}
		if row.SkillsVisible {
			e.Skills = nonEmpty(row.Skills)
		}
		out[i] = e
	}
	return out, nil
}

func optionalText(s string) pgtype.Text { return pgtype.Text{String: s, Valid: s != ""} }

func parseOptionalUUID(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{}, nil
	}
	return parseUUID(s)
}

func emptyToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// nonEmpty keeps a visible-but-empty list distinct from a hidden one (nil).
func nonEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
