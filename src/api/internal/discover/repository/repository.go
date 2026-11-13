// Package repository loads the published catalogue through sqlc.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

const queryTimeout = 2 * time.Second

// maxRelated is the contract cap for Resource.related.
const maxRelated = 20

// Repository is the sqlc adapter for the catalogue reader.
type Repository struct{ q *dbgen.Queries }

// New returns a repository over db.
func New(db dbgen.DBTX) *Repository { return &Repository{q: dbgen.New(db)} }

type entry struct {
	id                pgtype.UUID
	slug, name, kind  string
	category, summary string
	docLanguages      []string
	docsURL           pgtype.Text
	status            string
	verifiedAt        pgtype.Timestamptz
	updatedAt         pgtype.Timestamptz
	verifiedSourceID  pgtype.UUID
	providerSlug      string
	providerName      string
	providerURL       pgtype.Text
	sourceCount       int32
}

// List returns published resources for q, with countries, SDKs and related links.
func (r *Repository) List(ctx context.Context, q domain.Query) ([]domain.Resource, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	after, err := optionalID(q.AfterID)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListCatalogue(ctx, dbgen.ListCatalogueParams{
		Category: text(q.Category), ResourceType: text(q.Type), Provider: text(q.Provider),
		Country: text(q.Country), Sdk: text(q.SDK), Doclang: text(q.Doclang), Q: text(q.Q),
		AfterName: text(q.AfterName), AfterID: after, RowLimit: int32(q.Limit), //nolint:gosec // page size is at most 51
	})
	if err != nil {
		return nil, fmt.Errorf("catalogue list: %w", err)
	}
	entries := make([]entry, len(rows))
	for i, row := range rows {
		entries[i] = entryFromList(row)
	}
	return r.finish(ctx, entries)
}

// Get returns one published resource.
func (r *Repository) Get(ctx context.Context, id string) (domain.Resource, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	row, err := r.getEntry(ctx, id)
	if err != nil {
		return domain.Resource{}, err
	}
	items, err := r.finish(ctx, []entry{row})
	if err != nil {
		return domain.Resource{}, err
	}
	return items[0], nil
}

// Sources returns the provenance of a published resource. A missing resource is ErrNotFound.
func (r *Repository) Sources(ctx context.Context, id string) ([]domain.Source, error) {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	row, err := r.getEntry(ctx, id)
	if err != nil {
		return nil, err
	}
	rows, err := r.q.ListCatalogueSources(ctx, row.id)
	if err != nil {
		return nil, fmt.Errorf("catalogue sources: %w", err)
	}
	out := make([]domain.Source, len(rows))
	for i, row := range rows {
		checked := timePtr(row.CheckedAt)
		status, at := domain.SourceStatus(row.Kind, row.VerificationStatus, idString(row.ID), idString(row.VerifiedSourceID), checked)
		out[i] = domain.Source{ID: idString(row.ID), Kind: row.Kind, Title: row.Title, URL: row.Url, Status: status, VerifiedAt: at}
	}
	return out, nil
}

func (r *Repository) getEntry(ctx context.Context, id string) (entry, error) {
	uid, err := requiredID(id)
	if err != nil {
		return entry{}, domain.ErrNotFound
	}
	row, err := r.q.GetCatalogueEntry(ctx, uid)
	if errors.Is(err, pgx.ErrNoRows) {
		return entry{}, domain.ErrNotFound
	}
	if err != nil {
		return entry{}, fmt.Errorf("catalogue get: %w", err)
	}
	return entryFromGet(row), nil
}

func (r *Repository) finish(ctx context.Context, rows []entry) ([]domain.Resource, error) {
	out := make([]domain.Resource, len(rows))
	if len(rows) == 0 {
		return out, nil
	}
	ids := make([]pgtype.UUID, len(rows))
	for i, row := range rows {
		ids[i] = row.id
		out[i] = resourceFrom(row)
	}
	countries, err := r.q.CatalogueCountries(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("catalogue countries: %w", err)
	}
	for _, c := range countries {
		if i, ok := indexOf(rows, c.ResourceID); ok {
			out[i].Countries = append(out[i].Countries, c.CountryCode)
		}
	}
	sdks, err := r.q.CatalogueSDKs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("catalogue sdks: %w", err)
	}
	for _, s := range sdks {
		url := textPtr(s.Url)
		if url == nil {
			continue
		}
		if i, ok := indexOf(rows, s.ResourceID); ok {
			out[i].SDKs = append(out[i].SDKs, domain.SDK{Language: s.Language, Name: s.Name, URL: *url})
		}
	}
	related, err := r.q.CatalogueRelated(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("catalogue related: %w", err)
	}
	for _, link := range related {
		if i, ok := indexOf(rows, link.SrcResourceID); ok {
			out[i].Related = appendLink(out[i].Related, link.TargetType, idString(link.TargetID), link.Title, link.Lang)
		}
	}
	discussions, err := r.q.CatalogueDiscussions(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("catalogue discussions: %w", err)
	}
	for _, d := range discussions {
		if i, ok := indexOf(rows, d.ResourceID); ok {
			out[i].Related = appendLink(out[i].Related, "question", idString(d.QuestionID), d.Title, d.Language)
		}
	}
	for i := range out {
		out[i].Countries = orEmpty(out[i].Countries)
		out[i].SDKs = orEmptySDK(out[i].SDKs)
		out[i].Related = capRelated(out[i].Related)
		out[i].DocLanguages = orEmpty(out[i].DocLanguages)
	}
	return out, nil
}

func appendLink(links []domain.Link, kind, id, title, lang string) []domain.Link {
	switch kind {
	case "question", "resource", "project", "article":
	default:
		return links
	}
	if id == "" || title == "" {
		return links
	}
	return append(links, domain.Link{ID: id, Type: kind, Title: title, Lang: lang})
}

func capRelated(links []domain.Link) []domain.Link {
	if links == nil {
		return []domain.Link{}
	}
	if len(links) > maxRelated {
		return links[:maxRelated]
	}
	return links
}

func resourceFrom(row entry) domain.Resource {
	return domain.Resource{
		ID: idString(row.id), Slug: row.slug, Name: row.name, Type: row.kind, Category: row.category,
		Summary: row.summary, DocLanguages: row.docLanguages, DocsURL: textPtr(row.docsURL),
		Provider:           domain.Provider{Name: row.providerName, Slug: row.providerSlug, URL: textPtr(row.providerURL)},
		SourceCount:        int(row.sourceCount),
		VerificationStatus: row.status,
		LastVerifiedAt:     timePtr(row.verifiedAt),
		UpdatedAt:          row.updatedAt.Time,
	}
}

func entryFromList(row dbgen.ListCatalogueRow) entry {
	return entry{row.ID, row.Slug, row.Name, row.Type, row.CategoryCode, row.Summary, row.DocLanguages, row.DocsUrl, row.VerificationStatus, row.VerifiedAt, row.UpdatedAt, row.VerifiedSourceID, row.ProviderSlug, row.ProviderName, row.ProviderUrl, row.SourceCount}
}

func entryFromGet(row dbgen.GetCatalogueEntryRow) entry {
	return entry{row.ID, row.Slug, row.Name, row.Type, row.CategoryCode, row.Summary, row.DocLanguages, row.DocsUrl, row.VerificationStatus, row.VerifiedAt, row.UpdatedAt, row.VerifiedSourceID, row.ProviderSlug, row.ProviderName, row.ProviderUrl, row.SourceCount}
}

func indexOf(rows []entry, id pgtype.UUID) (int, bool) {
	for i, row := range rows {
		if row.id.Valid && id.Valid && row.id.Bytes == id.Bytes {
			return i, true
		}
	}
	return 0, false
}

func text(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func textPtr(v pgtype.Text) *string {
	if !v.Valid || v.String == "" {
		return nil
	}
	s := v.String
	return &s
}

func timePtr(v pgtype.Timestamptz) *time.Time {
	if !v.Valid {
		return nil
	}
	t := v.Time
	return &t
}

func requiredID(s string) (pgtype.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

func optionalID(s string) (pgtype.UUID, error) {
	if s == "" {
		return pgtype.UUID{}, nil
	}
	return requiredID(s)
}

func idString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	return uuid.UUID(id.Bytes).String()
}

func orEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func orEmptySDK(s []domain.SDK) []domain.SDK {
	if s == nil {
		return []domain.SDK{}
	}
	return s
}
