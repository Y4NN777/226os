// Package committers imports the committers.top ranking of public GitHub developers per country (ADR-0016):
// a fixed-host fetch client with an SSRF dial guard, a validating parser of the location files, and a daily
// job that replaces a country's rows only when its snapshot changed. No account is ever created.
package committers

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
)

const (
	// TopN is how many ranked users of a country are kept.
	TopN                 = 500
	maxNameRunes         = 200 // github_developers_name_ck
	maxOrganizations     = 50  // github_developers_organizations_ck
	maxOrganizationRunes = 100
	avatarPrefix         = "https://avatars.githubusercontent.com/"
	maxAvatarLen         = 500 // github_developers_avatar_ck
)

// loginPattern is profiles_github_ck and github_developers_login_ck.
var loginPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]{0,38}$`)

// Snapshot is one validated country file.
type Snapshot struct {
	Generated  time.Time
	Developers []domain.GitHubDeveloper
}

type fileUser struct {
	Rank          int64  `yaml:"rank"`
	Name          string `yaml:"name"`
	Login         string `yaml:"login"`
	AvatarURL     string `yaml:"avatarUrl"`
	Contributions int64  `yaml:"contributions"`
	Organizations string `yaml:"organizations"`
}

type locationFile struct {
	Users     []fileUser `yaml:"users"`
	Generated yaml.Node  `yaml:"generated"`
}

// Parse validates a committers.top location file for country. Only the `users` list (all contributions) is
// read. An invalid user is skipped; a file without a valid `generated` date or without any valid user is an
// error, so the caller keeps the previous snapshot.
func Parse(body []byte, country string) (Snapshot, error) {
	var f locationFile
	if err := yaml.Unmarshal(body, &f); err != nil {
		return Snapshot{}, fmt.Errorf("committers: parse: %w", err)
	}
	generated, err := parseGenerated(f.Generated)
	if err != nil {
		return Snapshot{}, err
	}
	snap := Snapshot{Generated: generated}
	seen := map[string]bool{}
	for _, u := range f.Users {
		d, ok := validUser(u)
		if !ok || seen[strings.ToLower(d.Login)] {
			continue
		}
		seen[strings.ToLower(d.Login)] = true
		d.Country, d.GeneratedAt = country, generated
		snap.Developers = append(snap.Developers, d)
	}
	if len(snap.Developers) == 0 {
		return Snapshot{}, errors.New("committers: no valid user")
	}
	return snap, nil
}

var generatedLayouts = []string{time.RFC3339, "2006-01-02 15:04:05 -0700", "2006-01-02 15:04:05 MST"}

func parseGenerated(n yaml.Node) (time.Time, error) {
	if n.Kind == yaml.ScalarNode {
		for _, layout := range generatedLayouts {
			if t, err := time.Parse(layout, n.Value); err == nil {
				return t.UTC(), nil
			}
		}
	}
	return time.Time{}, errors.New("committers: missing or invalid generated date")
}

func validUser(u fileUser) (domain.GitHubDeveloper, bool) {
	if u.Rank < 1 || u.Rank > TopN || u.Contributions < 0 || u.Contributions > math.MaxInt32 ||
		!loginPattern.MatchString(u.Login) ||
		!strings.HasPrefix(u.AvatarURL, avatarPrefix) || len(u.AvatarURL) > maxAvatarLen {
		return domain.GitHubDeveloper{}, false
	}
	name := truncate(clean(u.Name), maxNameRunes)
	if name == "" {
		name = u.Login
	}
	return domain.GitHubDeveloper{
		Login: u.Login, Rank: int(u.Rank), Name: name, AvatarURL: u.AvatarURL,
		Contributions: int(u.Contributions), Organizations: organizations(u.Organizations),
	}, true
}

// organizations splits the comma list, drops empty entries and bounds count and length.
func organizations(raw string) []string {
	out := []string{}
	for o := range strings.SplitSeq(raw, ",") {
		o = truncate(clean(o), maxOrganizationRunes)
		if o == "" {
			continue
		}
		if len(out) == maxOrganizations {
			break
		}
		out = append(out, o)
	}
	return out
}

// clean drops what PostgreSQL text cannot hold (NUL, invalid UTF-8) and trims spaces.
func clean(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ToValidUTF8(s, "�"), "\x00", ""))
}

func truncate(s string, runes int) string {
	if utf8.RuneCountInString(s) <= runes {
		return s
	}
	return strings.TrimSpace(string([]rune(s)[:runes]))
}
