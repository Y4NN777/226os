package config_test

import (
	"strings"
	"testing"

	"github.com/Y4NN777/226os/apps/api/internal/platform/config"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    config.Config
		wantErr string
	}{
		{"defaults address", map[string]string{"APP_ENV": "local-dev"}, config.Config{AppEnv: "local-dev", Addr: ":8080"}, ""},
		{"explicit address", map[string]string{"APP_ENV": "staging-test", "API_ADDR": "127.0.0.1:9000"}, config.Config{AppEnv: "staging-test", Addr: "127.0.0.1:9000"}, ""},
		{"production", map[string]string{"APP_ENV": "production-live"}, config.Config{AppEnv: "production-live", Addr: ":8080"}, ""},
		{"missing APP_ENV", map[string]string{}, config.Config{}, "APP_ENV"},
		{"unknown APP_ENV", map[string]string{"APP_ENV": "prod"}, config.Config{}, "APP_ENV"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := config.Load(env(tc.env))
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error = %v, want containing %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
