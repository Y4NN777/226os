package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

const (
	directoryDefaultLimit = 20
	directoryMaxLimit     = 50

	// The directory is limited to 60 requests per minute: per client network
	// (/24 or /48) for anonymous callers, per user for members. ADR-0003 has no
	// window for this route; 60 per minute is the anonymous search figure and
	// is the owner's to confirm (TODO.md, E5).
	directoryIPScope   = "profiles:ip"
	directoryUserScope = "profiles:user"
	directoryLimit     = 60
	directoryWindow    = time.Minute
	maxSkillSlugRunes  = 32
)

// Cursors signs page cursors and rejects a cursor used with other filters.
// The service owns this port; *cursor.Codec implements it.
type Cursors interface {
	Encode(cursor.Payload) string
	Decode(token, filterHash string) (cursor.Payload, error)
}

// DirectoryRequest is the validated shape of GET /profiles. Nil filters mean
// "not given"; a nil Limit means 20. CallerUserID is empty for anonymous
// callers (the session subject, never a client value); ClientIP is the trusted
// client address, used only to key the anonymous rate limit.
type DirectoryRequest struct {
	CallerUserID string
	ClientIP     string
	Country      *string
	Language     *string
	Skill        *string
	Source       string // all (or empty), members or github (ADR-0016)
	Cursor       string
	Limit        *int
}

// DirectoryPage is one page of listed profiles. NextCursor is empty on the
// last page.
type DirectoryPage struct {
	Items      []domain.DirectoryEntry
	External   []domain.GitHubDeveloper // after the members; nil when the page has none
	NextCursor string
}

// ListProfiles serves GET /profiles: the read-only people directory (E5).
// Listability and field visibility are decided in the store query for the
// caller's audience (anonymous or member; the owner has no extra reach), so a
// hidden value never reaches this layer. The service validates the filters
// and the page size, rate-limits the caller and signs the cursor.
func (s *Service) ListProfiles(ctx context.Context, req DirectoryRequest) (DirectoryPage, error) {
	q, err := directoryQuery(req)
	if err != nil {
		return DirectoryPage{}, err
	}
	source, err := directorySource(req.Source)
	if err != nil {
		return DirectoryPage{}, err
	}
	if err := s.allowDirectory(ctx, req); err != nil {
		return DirectoryPage{}, err
	}
	q.IncludeDemo = s.demoProfiles
	hash := lengthPrefixed("directory", q.Country, q.Language, q.Skill, source)
	var at externalPosition
	if req.Cursor != "" {
		p, err := s.cursors.Decode(req.Cursor, hash)
		if err != nil {
			return DirectoryPage{}, err
		}
		if at, err = decodePosition(p); err != nil {
			return DirectoryPage{}, err
		}
		if !at.github {
			q.AfterUsername, q.AfterID = p.Sort, p.ID
		}
	}
	size := q.Limit
	page := DirectoryPage{Items: []domain.DirectoryEntry{}}
	if source != sourceGitHub && !at.github {
		q.Limit = size + 1 // one extra row tells whether another page exists
		rows, err := s.store.ListProfiles(ctx, q)
		if err != nil {
			return DirectoryPage{}, apperr.Wrap(apperr.CodeInternal, err)
		}
		if len(rows) > size {
			page.Items = rows[:size]
			last := page.Items[size-1]
			page.NextCursor = s.cursors.Encode(cursor.Payload{Sort: last.Username, ID: last.UserID, FilterHash: hash})
			return page, nil
		}
		if rows != nil {
			page.Items = rows
		}
	}
	// External entries follow the members (ADR-0016). language and skill describe members only.
	if s.github == nil || source == sourceMembers || q.Language != "" || q.Skill != "" {
		return page, nil
	}
	return s.externalPage(ctx, page, q, at, size, hash)
}

const (
	sourceAll     = "all"
	sourceMembers = "members"
	sourceGitHub  = "github"
	// githubSortPrefix marks a cursor of the GitHub segment: "gh:" at its start, "gh:<rank>:<country>" with
	// the login as ID after an entry. A username cannot hold ':', so no member cursor looks like one.
	githubSortPrefix = "gh:"
	maxExternalRank  = 500
)

// externalPosition is the decoded keyset of the GitHub segment; github is false in the member segment.
type externalPosition struct {
	github  bool
	rank    int
	country string
	login   string
}

func directorySource(raw string) (string, error) {
	switch raw {
	case "", sourceAll:
		return sourceAll, nil
	case sourceMembers, sourceGitHub:
		return raw, nil
	default:
		return "", directoryInvalid("source", "invalid_value", "source must be all, members or github.")
	}
}

// decodePosition reads a signed cursor payload. Only the server writes payloads, so a malformed one is a
// tampered or stale cursor: 400.
func decodePosition(p cursor.Payload) (externalPosition, error) {
	invalid := apperr.New(apperr.CodeInvalidCursor, "The cursor is invalid or expired.")
	rest, ok := strings.CutPrefix(p.Sort, githubSortPrefix)
	if !ok {
		if _, err := uuid.Parse(p.ID); err != nil || p.Sort == "" {
			return externalPosition{}, invalid
		}
		return externalPosition{}, nil
	}
	if rest == "" && p.ID == "" {
		return externalPosition{github: true}, nil
	}
	rankText, country, ok := strings.Cut(rest, ":")
	rank, err := strconv.Atoi(rankText)
	if !ok || err != nil || rank < 1 || rank > maxExternalRank || !countryPattern.MatchString(country) ||
		!githubLoginPattern.MatchString(p.ID) {
		return externalPosition{}, invalid
	}
	return externalPosition{github: true, rank: rank, country: country, login: p.ID}, nil
}

// externalPage fills the rest of page with external entries. When the members filled it exactly, one probe
// row tells whether a GitHub segment follows.
func (s *Service) externalPage(ctx context.Context, page DirectoryPage, q domain.DirectoryQuery, at externalPosition, size int, hash string) (DirectoryPage, error) {
	remaining := size - len(page.Items)
	eq := domain.ExternalQuery{
		Member: q.Member, IncludeDemo: q.IncludeDemo, Country: q.Country,
		AfterRank: at.rank, AfterCountry: at.country, AfterLogin: at.login, Limit: remaining + 1,
	}
	rows, err := s.github.ListExternal(ctx, eq)
	if err != nil {
		return DirectoryPage{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	if remaining == 0 {
		if len(rows) > 0 {
			page.NextCursor = s.cursors.Encode(cursor.Payload{Sort: githubSortPrefix, FilterHash: hash})
		}
		return page, nil
	}
	if len(rows) > remaining {
		rows = rows[:remaining]
		last := rows[remaining-1]
		page.NextCursor = s.cursors.Encode(cursor.Payload{
			Sort: githubSortPrefix + strconv.Itoa(last.Rank) + ":" + last.Country, ID: last.Login, FilterHash: hash,
		})
	}
	if len(rows) > 0 {
		page.External = rows
	}
	return page, nil
}

// directoryQuery validates the request and builds the store query.
func directoryQuery(req DirectoryRequest) (domain.DirectoryQuery, error) {
	q := domain.DirectoryQuery{Member: req.CallerUserID != "", Limit: directoryDefaultLimit}
	if req.Limit != nil {
		if *req.Limit < 1 || *req.Limit > directoryMaxLimit {
			return q, directoryInvalid("limit", "out_of_range", "limit must be between 1 and 50.")
		}
		q.Limit = *req.Limit
	}
	if req.Country != nil {
		if !countryPattern.MatchString(*req.Country) {
			return q, directoryInvalid("country", "invalid_format", "country must be an ISO 3166-1 alpha-2 code.")
		}
		q.Country = *req.Country
	}
	if req.Language != nil {
		if !languagePattern.MatchString(*req.Language) {
			return q, directoryInvalid("language", "invalid_format", "language must be a 2 or 3 letter language tag.")
		}
		q.Language = *req.Language
	}
	if req.Skill != nil {
		if !skillSlugPattern.MatchString(*req.Skill) || len(*req.Skill) > maxSkillSlugRunes {
			return q, directoryInvalid("skill", "invalid_format", "skill must be a tag slug.")
		}
		q.Skill = *req.Skill
	}
	return q, nil
}

// allowDirectory applies the directory rate limit: per client network for
// anonymous callers, per user for members. Like every identity limit it fails
// closed; the list query uses the same database, so a failing limiter store
// would fail the read anyway.
func (s *Service) allowDirectory(ctx context.Context, req DirectoryRequest) error {
	if req.CallerUserID == "" {
		return s.allow(ctx, directoryIPScope, ipKey(req.ClientIP), directoryLimit, directoryWindow)
	}
	return s.allow(ctx, directoryUserScope, req.CallerUserID, directoryLimit, directoryWindow)
}

// lengthPrefixed binds a cursor to its filters; every field is length-prefixed
// so no value can imitate a boundary.
func lengthPrefixed(fields ...string) string {
	var b strings.Builder
	for _, f := range fields {
		b.WriteString(strconv.Itoa(len(f)))
		b.WriteByte(':')
		b.WriteString(f)
	}
	return b.String()
}

// directoryInvalid is a 422 with one violation: field is the query parameter
// name and code one of the Violation.code enum values of api/openapi.yaml.
func directoryInvalid(field, code, detail string) error {
	return apperr.New(apperr.CodeValidation, detail).WithViolations(apperr.Violation{Field: field, Code: code})
}
