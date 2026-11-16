package config_test

import (
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
)

func TestLoadAppBaseURL(t *testing.T) {
	tests := []struct {
		name    string
		appEnv  string
		value   string
		want    string
		wantErr bool
	}{
		{"unset", "local-dev", "", "", false},
		{"local http origin", "local-dev", "http://127.0.0.1:8088", "http://127.0.0.1:8088", false},
		{"local localhost", "local-dev", "http://localhost:8088", "http://localhost:8088", false},
		{"local https", "local-dev", "https://226os.local", "https://226os.local", false},
		{"staging https", "staging-test", "https://staging.226os.example", "https://staging.226os.example", false},
		{"production https", "production-live", "https://226os.example", "https://226os.example", false},
		{"one trailing slash is dropped", "production-live", "https://226os.example/", "https://226os.example", false},
		{"unset in production stays allowed", "production-live", "", "", false},
		{"http in staging", "staging-test", "http://staging.226os.example", "", true},
		{"http in production", "production-live", "http://226os.example", "", true},
		{"not a url", "local-dev", "not a url", "", true},
		{"no scheme", "local-dev", "226os.example", "", true},
		{"scheme-relative", "local-dev", "//226os.example", "", true},
		{"ftp scheme", "local-dev", "ftp://226os.example", "", true},
		{"javascript scheme", "local-dev", "javascript:alert(1)", "", true},
		{"opaque", "local-dev", "https:226os.example", "", true},
		{"empty host", "local-dev", "https://", "", true},
		{"path", "local-dev", "https://226os.example/app", "", true},
		{"path and slash", "local-dev", "https://226os.example//", "", true},
		{"query", "local-dev", "https://226os.example?x=1", "", true},
		{"bare question mark", "local-dev", "https://226os.example?", "", true},
		{"fragment", "local-dev", "https://226os.example#frag", "", true},
		{"bare fragment mark", "local-dev", "https://226os.example#", "", true},
		{"user info", "local-dev", "https://user:pw0rd-s3cret@226os.example", "", true},
		{"user only", "local-dev", "https://user@226os.example", "", true},
		{"surrounding space", "local-dev", " https://226os.example", "", true},
		{"control character", "local-dev", "https://226os.example\n", "", true},
		{"bad port", "local-dev", "https://226os.example:bad", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := config.Load(env(base(map[string]string{"APP_ENV": tc.appEnv, "APP_BASE_URL": tc.value})))
			if tc.wantErr {
				if err == nil || !strings.Contains(err.Error(), "APP_BASE_URL") {
					t.Fatalf("error = %v, want one naming APP_BASE_URL", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.AppBaseURL != tc.want {
				t.Fatalf("AppBaseURL = %q, want %q", got.AppBaseURL, tc.want)
			}
		})
	}
}

// A refused value never reaches the error text, so a pasted secret in the URL is not logged at startup.
func TestAppBaseURLErrorsNeverEchoTheValue(t *testing.T) {
	for _, v := range []string{
		"https://user:pw0rd-s3cret@226os.example",
		"https://226os.example/pw0rd-s3cret",
		"https://226os.example?token=pw0rd-s3cret",
		"https://226os.example#pw0rd-s3cret",
		"pw0rd-s3cret",
		"https://226os.example:pw0rd-s3cret",
	} {
		_, err := config.Load(env(base(map[string]string{"APP_BASE_URL": v})))
		if err == nil {
			t.Fatalf("%q: expected an error", v)
		}
		if strings.Contains(err.Error(), "pw0rd-s3cret") || strings.Contains(err.Error(), "226os.example") {
			t.Fatalf("%q: error %q echoes the value", v, err)
		}
	}
}

func TestAppBaseURLAppearsInTheRedactedForms(t *testing.T) {
	cfg, err := config.Load(env(base(map[string]string{"APP_BASE_URL": "http://127.0.0.1:8088"})))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cfg.String(), "AppBaseURL:http://127.0.0.1:8088") {
		t.Fatalf("String() = %q", cfg.String())
	}
	unset, err := config.Load(env(base(nil)))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(unset.String(), "AppBaseURL") {
		t.Fatalf("an unset base URL should not be printed: %q", unset.String())
	}
}
