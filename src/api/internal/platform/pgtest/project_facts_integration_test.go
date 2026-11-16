//go:build integration

package pgtest

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

// Project facts (owner decision 2026-11-15, prototype data gaps): stage, license_spdx and country_code are
// written on create and update, cleared with a merge-patch null, served on reads, and filter
// listProjects; the database checks stay the backstop and come back as 422, never 500.
func TestBuildProjectFacts(t *testing.T) {
	s := newBuildStack(t, &fakeGitHub{})
	alice := s.signIn(t, "facts-alice", true, 71)
	bob := s.signIn(t, "facts-bob", true, 72)

	// Defaults: idea, no license, no country.
	plain := s.createProject(t, &alice, "https://github.com/alice/plain", "public")
	if plain.Stage != apigen.ProjectStageIdea || plain.LicenseSpdx != nil || plain.CountryCode != nil {
		t.Fatalf("defaults %+v", plain)
	}
	rec := s.write(t, &alice, http.MethodPost, "/api/v1/projects",
		projectBody("https://github.com/alice/facts", "public", `"stage":"active"`, `"license_spdx":"Apache-2.0"`, `"country_code":"BF"`), nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	facts := decodeAs[apigen.Project](t, rec)
	if facts.Stage != apigen.ProjectStageActive || *facts.LicenseSpdx != "Apache-2.0" || *facts.CountryCode != "BF" {
		t.Fatalf("facts %+v", facts)
	}
	// Another user's private project with the same facts is never listed for them.
	rec = s.write(t, &bob, http.MethodPost, "/api/v1/projects",
		projectBody("https://github.com/bob/hidden", "private", `"stage":"active"`, `"license_spdx":"Apache-2.0"`, `"country_code":"BF"`), nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("bob create %d %s", rec.Code, rec.Body.String())
	}
	hidden := decodeAs[apigen.Project](t, rec)

	for _, tc := range []struct {
		viewer *member
		query  string
		want   []string
	}{
		{nil, "stage=active", []string{facts.Id.String()}},
		{nil, "country=BF", []string{facts.Id.String()}},
		{nil, "license=Apache-2.0", []string{facts.Id.String()}},
		{nil, "stage=active&country=BF&license=Apache-2.0", []string{facts.Id.String()}},
		{nil, "stage=idea", []string{plain.Id.String()}},
		{nil, "country=SN", nil},
		{&alice, "license=Apache-2.0", []string{facts.Id.String()}},
		{&bob, "license=Apache-2.0", []string{hidden.Id.String(), facts.Id.String()}},
	} {
		page := decodeAs[apigen.ProjectPage](t, s.read(t, tc.viewer, "/api/v1/projects?"+tc.query))
		if strings.Join(projectIDs(page), ",") != strings.Join(tc.want, ",") {
			t.Fatalf("%s: %v, want %v", tc.query, projectIDs(page), tc.want)
		}
	}
	// Bad filter values are 422 from the contract.
	for _, q := range []string{"stage=done", "country=bf", "license=MIT%20OR%20GPL"} {
		assertProblem(t, s.read(t, nil, "/api/v1/projects?"+q), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	}

	// Merge patch: set, clear with null, leave absent members unchanged.
	path := "/api/v1/projects/" + facts.Id.String()
	rec = s.write(t, &alice, http.MethodPatch, path, `{"stage":"maintained","license_spdx":null}`, nil)
	got := decodeAs[apigen.Project](t, rec)
	if rec.Code != http.StatusOK || got.Stage != apigen.ProjectStageMaintained || got.LicenseSpdx != nil || got.CountryCode == nil || *got.CountryCode != "BF" {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	rec = s.write(t, &alice, http.MethodPatch, path, `{"country_code":null,"license_spdx":"MIT"}`, nil)
	got = decodeAs[apigen.Project](t, rec)
	if rec.Code != http.StatusOK || got.CountryCode != nil || *got.LicenseSpdx != "MIT" || got.Stage != apigen.ProjectStageMaintained {
		t.Fatalf("patch 2 %d %s", rec.Code, rec.Body.String())
	}
	// Database backstops: an unknown country (pattern-valid) and a non-idea stage without a repository.
	assertProblemViolations(t, s.write(t, &alice, http.MethodPatch, path, `{"country_code":"ZZ"}`, nil), "/country_code", "invalid_value")
	assertProblemViolations(t, s.write(t, &alice, http.MethodPost, "/api/v1/projects",
		projectBody("https://github.com/alice/nowhere", "public", `"country_code":"ZZ"`), nil), "/country_code", "invalid_value")
	if _, err := s.conn.Exec(t.Context(), `UPDATE projects SET github_url = NULL, stage = 'idea' WHERE id = $1`, plain.Id.String()); err != nil {
		t.Fatal(err)
	}
	assertProblemViolations(t, s.write(t, &alice, http.MethodPatch, "/api/v1/projects/"+plain.Id.String(), `{"stage":"prototype"}`, nil), "/stage", "invalid_value")
	// A non-owner cannot set facts: 403 on a public project, nothing changes.
	assertProblem(t, s.write(t, &bob, http.MethodPatch, path, `{"stage":"idea"}`, nil), http.StatusForbidden, "FORBIDDEN")
	if got := decodeAs[apigen.Project](t, s.read(t, nil, path)); got.Stage != apigen.ProjectStageMaintained {
		t.Fatalf("non-owner changed the stage: %s", got.Stage)
	}
}
