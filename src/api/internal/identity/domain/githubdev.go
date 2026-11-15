package domain

import "time"

// GitHubDeveloper is one public GitHub user ranked by committers.top in a country (ADR-0016). It is an
// external directory entry: no account, profile or user row exists for it.
type GitHubDeveloper struct {
	Country       string
	Login         string
	Rank          int
	Name          string
	AvatarURL     string
	Contributions int
	Organizations []string
	GeneratedAt   time.Time
}

// ExternalQuery is one page request of the GitHub segment of the directory. Member and IncludeDemo decide
// which members deduplicate an entry, exactly as for DirectoryQuery. After* is the keyset position
// (AfterRank 0 means the first page).
type ExternalQuery struct {
	Member, IncludeDemo bool
	Country             string
	AfterRank           int
	AfterCountry        string
	AfterLogin          string
	Limit               int
}

// Ranking is the committers.top rank of a member's visible GitHub login.
type Ranking struct {
	Country     string
	Rank        int
	GeneratedAt time.Time
}
