package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// ---------- profiles ----------

func (r *Repository) ProfileByUserID(ctx context.Context, userID string) (domain.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	uid, err := parseUUID(userID)
	if err != nil {
		return domain.Profile{}, domain.ErrNotFound
	}
	row, err := r.q.GetProfileByUserID(ctx, uid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Profile{}, domain.ErrNotFound
		}
		return domain.Profile{}, fmt.Errorf("identity profile by user: %w", err)
	}
	return profileFromRow(row.UserID, row.Username, row.DisplayName, row.Intent, row.CountryCode,
		row.Languages, row.GithubUsername, row.Bio, row.OpenToCollab,
		row.CountryVisibility, row.LanguagesVisibility, row.SkillsVisibility,
		row.GithubVisibility, row.BioVisibility, row.OpenToCollabVisibility,
		row.ProfileVisibility, row.CreatedAt, row.UpdatedAt, row.Skills), nil
}

func (r *Repository) ProfileByUsername(ctx context.Context, username string) (domain.ProfileRecord, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	row, err := r.q.GetProfileByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ProfileRecord{}, domain.ErrNotFound
		}
		return domain.ProfileRecord{}, fmt.Errorf("identity profile by username: %w", err)
	}
	p := profileFromRow(row.UserID, row.Username, row.DisplayName, row.Intent, row.CountryCode,
		row.Languages, row.GithubUsername, row.Bio, row.OpenToCollab,
		row.CountryVisibility, row.LanguagesVisibility, row.SkillsVisibility,
		row.GithubVisibility, row.BioVisibility, row.OpenToCollabVisibility,
		row.ProfileVisibility, row.CreatedAt, row.UpdatedAt, row.Skills)
	return domain.ProfileRecord{
		Profile: p, AccountStatus: row.AccountStatus, JoinedAt: row.JoinedAt.Time,
	}, nil
}

// profileFromRow maps the shared projection. skillsArr is []string thanks to
// the explicit ::text[] cast in the query.
func profileFromRow(userID pgtype.UUID, username, displayName string, intent, country pgtype.Text,
	languages []string, github, bio pgtype.Text, openToCollab bool,
	visCountry, visLanguages, visSkills, visGitHub, visBio, visOpenCollab, visProfile string,
	created, updated pgtype.Timestamptz, skills []string) domain.Profile {
	p := domain.Profile{
		UserID: uuidString(userID), Username: username, DisplayName: displayName,
		Intent: textPtr(intent), CountryCode: textPtr(country),
		Languages: languages, GitHub: textPtr(github), Bio: textPtr(bio),
		OpenToCollab: openToCollab,
		VisCountry:   domain.Visibility(visCountry), VisLanguages: domain.Visibility(visLanguages),
		VisSkills: domain.Visibility(visSkills), VisGitHub: domain.Visibility(visGitHub),
		VisBio: domain.Visibility(visBio), VisOpenCollab: domain.Visibility(visOpenCollab),
		Visibility: domain.Visibility(visProfile),
		Skills:     skills,
		CreatedAt:  created.Time, UpdatedAt: updated.Time,
	}
	return p
}

func (r *Repository) ResolveSkillSlugs(ctx context.Context, slugs []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	rows, err := r.q.ResolveSkillSlugs(ctx, slugs)
	if err != nil {
		return nil, fmt.Errorf("identity resolve skill slugs: %w", err)
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, uuidString(row.ID))
	}
	return ids, nil
}

// UpdateProfile translates the resolved patch into the guarded UPDATE. Each
// PatchField becomes a set-flag plus a value; a present field with a nil value
// clears the column to NULL.
func (r *Repository) UpdateProfile(ctx context.Context, userID string, w domain.ProfileWrite) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	uid, err := parseUUID(userID)
	if err != nil {
		return domain.ErrNotFound
	}
	params := dbgen.UpdateProfileParams{UserID: uid}
	params.Column2, params.DisplayName = patchString(w.DisplayName)
	params.Column4, params.Intent = patchText(w.Intent)
	params.Column6, params.CountryCode = patchText(w.CountryCode)
	params.Column8, params.Languages = patchStrings(w.Languages)
	params.Column10, params.GithubUsername = patchText(w.GitHub)
	params.Column12, params.Bio = patchText(w.Bio)
	params.Column14, params.BioLanguage = patchText(w.BioLanguage)
	params.Column16, params.OpenToCollab = patchBool(w.OpenToCollab)
	params.Column18, params.CountryVisibility = patchVisibility(w.VisCountry)
	params.Column20, params.LanguagesVisibility = patchVisibility(w.VisLanguages)
	params.Column22, params.SkillsVisibility = patchVisibility(w.VisSkills)
	params.Column24, params.GithubVisibility = patchVisibility(w.VisGitHub)
	params.Column26, params.BioVisibility = patchVisibility(w.VisBio)
	params.Column28, params.ProfileVisibility = patchVisibility(w.Visibility)
	mapErr := func(err error) error {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "profiles_country_fk" {
			return domain.ErrCountryUnknown
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("identity update profile: %w", err)
	}
	if !w.SkillTagIDs.Present {
		if _, err := r.q.UpdateProfile(ctx, params); err != nil {
			return mapErr(err)
		}
		return nil
	}
	// A skill replacement runs in a transaction: the guarded UPDATE takes the
	// profile row lock first, then DELETE/INSERT run as separate statements,
	// each on a fresh READ COMMITTED snapshot. Concurrent patches therefore
	// serialize on the row lock and the last writer's set wins exactly,
	// instead of merging two snapshots taken before the lock wait.
	ids, err := patchUUIDs(w.SkillTagIDs)
	if err != nil {
		return err
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("identity update profile begin: %w", err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(ctx)) }()
	q := dbgen.New(tx)
	if _, err := q.UpdateProfile(ctx, params); err != nil {
		return mapErr(err)
	}
	if err := q.DeleteUserSkills(ctx, uid); err != nil {
		return fmt.Errorf("identity replace skills delete: %w", err)
	}
	if len(ids) > 0 {
		if err := q.InsertUserSkills(ctx, dbgen.InsertUserSkillsParams{UserID: uid, Column2: ids}); err != nil {
			return fmt.Errorf("identity replace skills insert: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("identity update profile commit: %w", err)
	}
	return nil
}

// ---------- patch mapping helpers ----------

func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func patchText[T ~string](f domain.PatchField[T]) (bool, pgtype.Text) {
	if !f.Present {
		return false, pgtype.Text{}
	}
	if f.Value == nil {
		return true, pgtype.Text{}
	}
	return true, pgtype.Text{String: string(*f.Value), Valid: true}
}

// patchString is patchText for NOT NULL text columns: the value param is a
// plain string, and a present field always carries a non-nil value (service
// validation rejects a clear on a required column).
func patchString(f domain.PatchField[string]) (bool, string) {
	if !f.Present {
		return false, ""
	}
	if f.Value == nil {
		return true, ""
	}
	return true, *f.Value
}

func patchVisibility(f domain.PatchField[domain.Visibility]) (bool, string) {
	flag, text := patchText(f)
	return flag, text.String
}

func patchStrings(f domain.PatchField[[]string]) (bool, []string) {
	if !f.Present {
		return false, nil
	}
	if f.Value == nil || *f.Value == nil {
		return true, []string{}
	}
	return true, *f.Value
}

func patchBool(f domain.PatchField[bool]) (bool, bool) {
	if !f.Present || f.Value == nil {
		return f.Present, false
	}
	return true, *f.Value
}

func patchUUIDs(f domain.PatchField[[]string]) ([]pgtype.UUID, error) {
	if !f.Present {
		return nil, nil // NULL: leave user_skills untouched
	}
	if f.Value == nil || *f.Value == nil {
		return []pgtype.UUID{}, nil // present + empty: clear every skill
	}
	out := make([]pgtype.UUID, 0, len(*f.Value))
	for _, s := range *f.Value {
		u, err := parseUUID(s)
		if err != nil {
			return nil, fmt.Errorf("identity skill id %q: %w", s, err)
		}
		out = append(out, u)
	}
	return out, nil // non-nil even when empty: an empty set clears all skills
}
