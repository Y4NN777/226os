package committers

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// sample is the committers.top location file shape (gh-pages, _data/locations/<slug>.yml).
const sample = `users:
- rank: 1
  name: "Adé Ouédraogo"
  login: "ade-dev"
  avatarUrl: "https://avatars.githubusercontent.com/u/1?v=4"
  contributions: 4321
  company: "@acme"
  organizations: "acme, open-bf,  "
- rank: 2
  name: ""
  login: "NoName"
  avatarUrl: "https://avatars.githubusercontent.com/u/2?v=4"
  contributions: 10
  company: ""
  organizations: ""
users_public_contributions:
- rank: 1
  name: "Other list"
  login: "other"
  avatarUrl: "https://avatars.githubusercontent.com/u/3?v=4"
  contributions: 1
  company: ""
  organizations: ""
private_users: []
organizations: []
generated: 2026-11-14 06:12:44 +0000
total_user_count: 812
title: "Burkina Faso"
`

func TestParseReadsUsersWithUnicodeEscapes(t *testing.T) {
	snap, err := Parse([]byte(sample), "BF")
	if err != nil {
		t.Fatal(err)
	}
	if !snap.Generated.Equal(time.Date(2026, 11, 14, 6, 12, 44, 0, time.UTC)) {
		t.Fatalf("generated %v", snap.Generated)
	}
	if len(snap.Developers) != 2 {
		t.Fatalf("developers %+v", snap.Developers)
	}
	a := snap.Developers[0]
	if a.Name != "Adé Ouédraogo" || a.Login != "ade-dev" || a.Rank != 1 || a.Contributions != 4321 ||
		a.Country != "BF" || a.AvatarURL != "https://avatars.githubusercontent.com/u/1?v=4" ||
		strings.Join(a.Organizations, "|") != "acme|open-bf" || !a.GeneratedAt.Equal(snap.Generated) {
		t.Fatalf("first %+v", a)
	}
	if b := snap.Developers[1]; b.Name != "NoName" || len(b.Organizations) != 0 {
		t.Fatalf("an empty name falls back to the login, empty organisations are none: %+v", b)
	}
}

func TestParseAcceptsRFC3339Generated(t *testing.T) {
	src := strings.Replace(sample, "generated: 2026-11-14 06:12:44 +0000", "generated: \"2026-11-14T06:12:44Z\"", 1)
	snap, err := Parse([]byte(src), "BF")
	if err != nil || !snap.Generated.Equal(time.Date(2026, 11, 14, 6, 12, 44, 0, time.UTC)) {
		t.Fatalf("%v %v", snap.Generated, err)
	}
}

func TestParseSkipsInvalidUsers(t *testing.T) {
	entry := func(rank int, login, avatar string, contributions int) string {
		return fmt.Sprintf("- rank: %d\n  name: \"x\"\n  login: %q\n  avatarUrl: %q\n  contributions: %d\n  organizations: \"\"\n",
			rank, login, avatar, contributions)
	}
	good := "https://avatars.githubusercontent.com/u/9"
	src := "users:\n" +
		entry(1, "-bad", good, 1) + // login pattern
		entry(2, "evil", "https://evil.example/a.png", 1) + // avatar host
		entry(3, "plain", "http://avatars.githubusercontent.com/u/9", 1) + // not https
		entry(0, "zero", good, 1) + // rank
		entry(4, "neg", good, -1) + // contributions
		entry(5, "huge", good, 1<<40) + // contributions above int32
		entry(6, "ok", good, 1) +
		entry(7, "OK", good, 1) + // duplicate login, case-insensitive
		entry(501, "late", good, 1) + // past the top 500
		"generated: 2026-11-14T06:12:44Z\n"
	snap, err := Parse([]byte(src), "BF")
	if err != nil {
		t.Fatal(err)
	}
	if len(snap.Developers) != 1 || snap.Developers[0].Login != "ok" {
		t.Fatalf("kept %+v", snap.Developers)
	}
}

func TestParseBoundsNameAndOrganizations(t *testing.T) {
	orgs := strings.Repeat("org,", 80) + strings.Repeat("o", 200)
	src := fmt.Sprintf("users:\n- rank: 1\n  name: %q\n  login: \"a\"\n  avatarUrl: \"https://avatars.githubusercontent.com/u/1\"\n  contributions: 1\n  organizations: %q\ngenerated: 2026-11-14T06:12:44Z\n",
		strings.Repeat("é", 300)+"\x00", orgs)
	snap, err := Parse([]byte(src), "BF")
	if err != nil {
		t.Fatal(err)
	}
	d := snap.Developers[0]
	if n := len([]rune(d.Name)); n != maxNameRunes || strings.ContainsRune(d.Name, 0) {
		t.Fatalf("name runes %d", n)
	}
	if len(d.Organizations) != maxOrganizations {
		t.Fatalf("organizations %d", len(d.Organizations))
	}
	for _, o := range d.Organizations {
		if len([]rune(o)) > maxOrganizationRunes {
			t.Fatalf("organization %q", o)
		}
	}
}

func TestParseRejectsBadFiles(t *testing.T) {
	for name, src := range map[string]string{
		"not yaml":          "users: [",
		"no generated":      "users:\n- rank: 1\n  login: \"a\"\n  avatarUrl: \"https://avatars.githubusercontent.com/u/1\"\n",
		"bad generated":     "generated: yesterday\nusers:\n- rank: 1\n  login: \"a\"\n  avatarUrl: \"https://avatars.githubusercontent.com/u/1\"\n",
		"no valid user":     "generated: 2026-11-14T06:12:44Z\nusers:\n- rank: 1\n  login: \"-\"\n",
		"no users":          "generated: 2026-11-14T06:12:44Z\n",
		"generated as list": "generated: [1]\nusers: []\n",
	} {
		if _, err := Parse([]byte(src), "BF"); err == nil {
			t.Errorf("%s: want an error", name)
		}
	}
}
