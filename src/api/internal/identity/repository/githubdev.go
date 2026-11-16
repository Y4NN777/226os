package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// replaceTimeout bounds one country's replace (500 inserts at most) inside one transaction.
const replaceTimeout = 30 * time.Second

// GitHubGeneratedAt is the committers.top generated date of a country's stored snapshot; ok is false when
// the country holds no row.
func (r *Repository) GitHubGeneratedAt(ctx context.Context, country string) (time.Time, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	ts, err := r.q.GitHubDevelopersGeneratedAt(ctx, country)
	if errors.Is(err, pgx.ErrNoRows) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, fmt.Errorf("identity github generated at: %w", err)
	}
	return ts.Time, true, nil
}

// ReplaceGitHubDevelopers swaps a country's rows in one transaction: on any error nothing changes, so the
// previous snapshot stays (ADR-0016).
func (r *Repository) ReplaceGitHubDevelopers(ctx context.Context, country string, devs []domain.GitHubDeveloper) error {
	ctx, cancel := context.WithTimeout(ctx, replaceTimeout)
	defer cancel()
	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		q := r.q.WithTx(tx)
		if err := q.DeleteGitHubDevelopersByCountry(ctx, country); err != nil {
			return err
		}
		for _, d := range devs {
			if err := q.InsertGitHubDeveloper(ctx, dbgen.InsertGitHubDeveloperParams{
				CountryCode: country, Login: d.Login,
				Rank: int32(d.Rank), //nolint:gosec // the importer bounds rank to 1..500
				Name: d.Name, AvatarUrl: d.AvatarURL,
				Contributions: int32(d.Contributions), //nolint:gosec // the importer bounds it to int32
				Organizations: nonEmpty(d.Organizations),
				GeneratedAt:   pgtype.Timestamptz{Time: d.GeneratedAt, Valid: true},
				ImportedAt:    now,
			}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("identity replace github developers: %w", err)
	}
	return nil
}

// ListExternal runs the GitHub segment of the directory (ListExternalDevelopers). q.Limit is the page size
// plus the look-ahead row, at most 51.
func (r *Repository) ListExternal(ctx context.Context, q domain.ExternalQuery) ([]domain.GitHubDeveloper, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	params := dbgen.ListExternalDevelopersParams{
		Country: optionalText(q.Country), Member: q.Member, IncludeDemo: q.IncludeDemo,
		RowLimit: int32(q.Limit), //nolint:gosec // the service bounds the page size to 51
	}
	if q.AfterRank > 0 {
		params.AfterRank = pgtype.Int4{Int32: int32(q.AfterRank), Valid: true} //nolint:gosec // decoded from a signed cursor, bounded by the service
		params.AfterCountry = pgtype.Text{String: q.AfterCountry, Valid: true}
		params.AfterLogin = pgtype.Text{String: q.AfterLogin, Valid: true}
	}
	rows, err := r.q.ListExternalDevelopers(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("identity list external developers: %w", err)
	}
	out := make([]domain.GitHubDeveloper, len(rows))
	for i, row := range rows {
		out[i] = domain.GitHubDeveloper{
			Country: row.CountryCode, Login: row.Login, Rank: int(row.Rank), Name: row.Name,
			AvatarURL: row.AvatarUrl, Contributions: int(row.Contributions),
			Organizations: nonEmpty(row.Organizations), GeneratedAt: row.GeneratedAt.Time,
		}
	}
	return out, nil
}

// GitHubRank is the best committers.top rank of a login; domain.ErrNotFound when it has none or opted out.
func (r *Repository) GitHubRank(ctx context.Context, login string) (domain.Ranking, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	row, err := r.q.GitHubRankForLogin(ctx, login)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Ranking{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Ranking{}, fmt.Errorf("identity github rank: %w", err)
	}
	return domain.Ranking{Country: row.CountryCode, Rank: int(row.Rank), GeneratedAt: row.GeneratedAt.Time}, nil
}
