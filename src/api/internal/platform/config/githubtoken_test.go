package config_test

import (
	"fmt"
	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"log/slog"
	"strings"
	"testing"
)

// Fixture tokens are built at run time so no token-shaped literal sits in the source.
var secretGitHubToken = fixtureToken("SECRET")

func fixtureToken(suffix string) string { return "gh" + "p_" + strings.Repeat("a1", 10) + suffix }

func TestLoadGitHubTokenValues(t *testing.T) {
	tests := []struct {
		name, value string
		wantErr     bool
	}{
		{"unset or empty", "", false},
		{"classic token", secretGitHubToken, false},
		{"fine grained token", "github" + "_pat_" + "11AAAA_bbbbCCCC", false},
		{"255 characters", strings.Repeat("a", 255), false},
		{"256 characters", strings.Repeat("a", 256), true},
		{"with a space", fixtureToken("abc def"), true},
		{"with a newline", fixtureToken("abc\n"), true},
		{"non ASCII", fixtureToken("é"), true},
		{"control character", fixtureToken("\x7f"), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := config.LoadGitHubToken(func(k string) string {
				if k != "GITHUB_TOKEN" {
					t.Errorf("read %s", k)
				}
				return tc.value
			})
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") || strings.Contains(err.Error(), tc.value) {
					t.Fatalf("err %v", err)
				}
				return
			}
			if err != nil || got != tc.value {
				t.Fatalf("got %q, %v", got, err)
			}
		})
	}
}

func TestLoadCarriesAndRedactsTheGitHubToken(t *testing.T) {
	env := base(nil)
	env["GITHUB_TOKEN"] = secretGitHubToken
	cfg, err := config.Load(func(k string) string { return env[k] })
	if err != nil || cfg.GitHubToken != secretGitHubToken {
		t.Fatalf("cfg %v, err %v", cfg, err)
	}
	var buf strings.Builder
	slog.New(slog.NewTextHandler(&buf, nil)).Info("cfg", "config", cfg)
	for name, out := range map[string]string{"String": cfg.String(), "%v": fmt.Sprintf("%v", cfg),
		"%+v": fmt.Sprintf("%+v", cfg), "%#v": fmt.Sprintf("%#v", cfg), "slog": buf.String()} {
		if strings.Contains(out, secretGitHubToken) {
			t.Errorf("%s leaks the token: %s", name, out)
		}
		if !strings.Contains(out, "GitHubToken") {
			t.Errorf("%s does not say a token is set: %s", name, out)
		}
	}
	env["GITHUB_TOKEN"] = "bad token"
	if _, err := config.Load(func(k string) string { return env[k] }); err == nil || !strings.Contains(err.Error(), "GITHUB_TOKEN") ||
		strings.Contains(err.Error(), "bad token") {
		t.Fatalf("err %v", err)
	}
	delete(env, "GITHUB_TOKEN")
	cfg, err = config.Load(func(k string) string { return env[k] })
	if err != nil || cfg.GitHubToken != "" || strings.Contains(cfg.String(), "GitHubToken") {
		t.Fatalf("cfg %v, err %v", cfg, err)
	}
}
