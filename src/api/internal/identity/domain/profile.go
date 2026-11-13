package domain

import "time"

// Visibility is the three-level audience shared by profile_visibility and the
// per-field *_visibility columns (C14b, ADR-0005 point 15).
type Visibility string

const (
	// VisibilityPublic shows the value to every caller, anonymous included.
	VisibilityPublic Visibility = "public"
	// VisibilityMembers shows the value to authenticated members only.
	VisibilityMembers Visibility = "members"
	// VisibilityPrivate shows the value to the owner only.
	VisibilityPrivate Visibility = "private"
)

// Allows reports whether a caller with the given authentication state and
// ownership may see a value guarded by v. Anonymous callers only pass
// public; members pass public and members; owners pass everything.
func (v Visibility) Allows(anonymous, owner bool) bool {
	switch v {
	case VisibilityPublic:
		return true
	case VisibilityMembers:
		return !anonymous || owner
	case VisibilityPrivate:
		return owner
	}
	return false
}

// Profile is the profiles row plus the owner's skill slugs. It carries every
// stored column; filtering for a caller happens in the service, never here.
type Profile struct {
	UserID        string
	Username      string
	DisplayName   string
	Intent        *string
	CountryCode   *string
	Languages     []string
	GitHub        *string
	Bio           *string
	OpenToCollab  bool
	VisCountry    Visibility
	VisLanguages  Visibility
	VisSkills     Visibility
	VisGitHub     Visibility
	VisBio        Visibility
	VisOpenCollab Visibility
	Visibility    Visibility // profile_visibility: who may open the page
	Skills        []string   // tag slugs, sorted
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ProfileRecord is a Profile joined with the account facts the public
// visibility decision needs.
type ProfileRecord struct {
	Profile
	AccountStatus string    // users.status; only "active" profiles are served
	JoinedAt      time.Time // users.created_at, exposed as joined_at
}

// PublicProfile is the getProfile result after visibility filtering: fields
// the caller may not see are nil, never redacted placeholders, so transport
// omits them entirely. Username, DisplayName, Visibility and JoinedAt are
// always present (contract required members).
type PublicProfile struct {
	Username     string
	DisplayName  string
	Bio          *string
	CountryCode  *string
	Languages    []string // nil means hidden
	Skills       []string // nil means hidden
	GitHub       *string
	OpenToCollab *bool
	Visibility   Visibility
	JoinedAt     time.Time
}

// ProfileWrite is the resolved merge patch handed to the store: every member
// is a PatchField so "absent" and "explicit null" stay distinct end to end.
type ProfileWrite struct {
	DisplayName PatchField[string]
	Intent      PatchField[string]
	CountryCode PatchField[string]
	Languages   PatchField[[]string]
	GitHub      PatchField[string]
	Bio         PatchField[string]
	// BioLanguage is derived by the service, never taken from the contract:
	// 'en' when Bio is set, cleared together with Bio, untouched otherwise
	// (profiles_bio_language_ck requires the pair to move together).
	BioLanguage  PatchField[string]
	OpenToCollab PatchField[bool]
	VisCountry   PatchField[Visibility]
	VisLanguages PatchField[Visibility]
	VisSkills    PatchField[Visibility]
	VisGitHub    PatchField[Visibility]
	VisBio       PatchField[Visibility]
	Visibility   PatchField[Visibility]
	// SkillSlugs replaces the whole skill set when Present; Value nil +
	// Present clears every skill. Tag ids are resolved by the service.
	SkillTagIDs PatchField[[]string]
}

// PatchField is one JSON Merge Patch member: Present=false leaves the stored
// value alone; Present=true with Value=nil clears a nullable column.
type PatchField[T any] struct {
	Present bool
	Value   *T
}
