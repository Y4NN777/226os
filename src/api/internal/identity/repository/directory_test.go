package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
)

// directoryRow is one ListDirectoryProfiles row in scan order: user_id, username, display_name,
// country_code, languages_visible, languages, skills_visible, skills, github_username.
func directoryRow(country string, langsVisible bool, langs []string, skillsVisible bool, skills []string, github string) []any {
	return []any{pguuid(), "ada", "Ada", country, langsVisible, langs, skillsVisible, skills, github}
}

func TestListProfilesMapsVisibleAndHiddenFields(t *testing.T) {
	db := &fakeDB{queryRows: [][]any{
		directoryRow("BF", true, []string{"fr", "en"}, true, []string{"go"}, "ada"),
		directoryRow("", false, nil, false, nil, ""),
		directoryRow("", true, nil, true, nil, ""), // visible but empty lists stay non-nil
	}}
	got, err := New(db).ListProfiles(context.Background(), domain.DirectoryQuery{
		Member: true, Country: "BF", Language: "fr", Skill: "go",
		AfterUsername: "aa", AfterID: testUUID.String(), Limit: 21,
	})
	if err != nil || len(got) != 3 {
		t.Fatalf("got %+v, %v", got, err)
	}
	a := got[0]
	if a.UserID != testUUID.String() || a.Username != "ada" || a.DisplayName != "Ada" ||
		a.Country == nil || *a.Country != "BF" || a.GitHub == nil || *a.GitHub != "ada" ||
		len(a.Languages) != 2 || len(a.Skills) != 1 {
		t.Fatalf("visible entry %+v", a)
	}
	h := got[1]
	if h.Country != nil || h.GitHub != nil || h.Languages != nil || h.Skills != nil {
		t.Fatalf("hidden fields must map to nil: %+v", h)
	}
	e := got[2]
	if e.Languages == nil || len(e.Languages) != 0 || e.Skills == nil || len(e.Skills) != 0 {
		t.Fatalf("visible empty lists must stay non-nil: %+v", e)
	}
}

func TestListProfilesFirstPageHasNoKeyset(t *testing.T) {
	got, err := New(&fakeDB{}).ListProfiles(context.Background(), domain.DirectoryQuery{Limit: 21})
	if err != nil || len(got) != 0 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestListProfilesErrors(t *testing.T) {
	if _, err := New(&fakeDB{}).ListProfiles(context.Background(), domain.DirectoryQuery{AfterUsername: "aa", AfterID: "not-a-uuid", Limit: 21}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad keyset id: %v", err)
	}
	boom := errors.New("db down")
	if _, err := New(&fakeDB{queryErr: boom}).ListProfiles(context.Background(), domain.DirectoryQuery{Limit: 21}); !errors.Is(err, boom) {
		t.Fatalf("query error: %v", err)
	}
}
