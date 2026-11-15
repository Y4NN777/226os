package config_test

import (
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
)

// CONNECT_GITHUB_IMPORT switches the committers.top import (ADR-0016): on by default in staging-test and
// production-live, off by default in local-dev; an explicit true or false wins.
func TestLoadGitHubImport(t *testing.T) {
	tests := []struct {
		env, value string
		want, err  bool
	}{
		{"local-dev", "", false, false},
		{"local-dev", "true", true, false},
		{"staging-test", "", true, false},
		{"production-live", "", true, false},
		{"production-live", "false", false, false},
		{"staging-test", "yes", false, true},
	}
	for _, tc := range tests {
		got, err := config.LoadGitHubImport(func(k string) string {
			if k == "CONNECT_GITHUB_IMPORT" {
				return tc.value
			}
			return ""
		}, tc.env)
		if got != tc.want || (err != nil) != tc.err {
			t.Errorf("%s %q: %v %v", tc.env, tc.value, got, err)
		}
	}
}

func TestLoadCarriesGitHubImport(t *testing.T) {
	env := map[string]string{
		"APP_ENV": "staging-test", "DATABASE_URL": "postgres://u@h/db",
		"CURSOR_KEY": "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY", "SMTP_HOST": "", "APP_BASE_URL": "https://staging.example",
	}
	cfg, err := config.Load(func(k string) string { return env[k] })
	if err != nil || !cfg.GitHubImport {
		t.Fatalf("%v %v", cfg.GitHubImport, err)
	}
	env["CONNECT_GITHUB_IMPORT"] = "maybe"
	if _, err := config.Load(func(k string) string { return env[k] }); err == nil {
		t.Fatal("an invalid value fails the load")
	}
}
