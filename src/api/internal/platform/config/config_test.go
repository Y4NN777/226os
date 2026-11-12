package config_test

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

var (
	rawKey    = bytes.Repeat([]byte{0xAB}, 32)
	keyURL    = base64.RawURLEncoding.EncodeToString(rawKey)
	keyStd    = base64.StdEncoding.EncodeToString(rawKey)
	keyStdRaw = base64.RawStdEncoding.EncodeToString(rawKey)
	keyURLPad = base64.URLEncoding.EncodeToString(rawKey)
)

const (
	dbURL    = "postgres://app:s3cr3t-pw@db:5432/app?sslmode=disable"
	dbURLAlt = "postgresql://app:s3cr3t-pw@db/app"
)

func base(extra map[string]string) map[string]string {
	m := map[string]string{"APP_ENV": "local-dev", "DATABASE_URL": dbURL, "CURSOR_KEY": keyURL}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantEnv string
		wantAdr string
		wantDB  string
		wantErr string
	}{
		{"defaults address", base(nil), "local-dev", ":8080", dbURL, ""},
		{"explicit address", base(map[string]string{"APP_ENV": "staging-test", "API_ADDR": "127.0.0.1:9000"}), "staging-test", "127.0.0.1:9000", dbURL, ""},
		{"production", base(map[string]string{"APP_ENV": "production-live"}), "production-live", ":8080", dbURL, ""},
		{"postgresql scheme", base(map[string]string{"DATABASE_URL": dbURLAlt}), "local-dev", ":8080", dbURLAlt, ""},
		{"key base64 std", base(map[string]string{"CURSOR_KEY": keyStd}), "local-dev", ":8080", dbURL, ""},
		{"key base64 std unpadded", base(map[string]string{"CURSOR_KEY": keyStdRaw}), "local-dev", ":8080", dbURL, ""},
		{"key base64url padded", base(map[string]string{"CURSOR_KEY": keyURLPad}), "local-dev", ":8080", dbURL, ""},
		{"missing APP_ENV", base(map[string]string{"APP_ENV": ""}), "", "", "", "APP_ENV"},
		{"unknown APP_ENV", base(map[string]string{"APP_ENV": "prod"}), "", "", "", "APP_ENV"},
		{"missing DATABASE_URL", base(map[string]string{"DATABASE_URL": ""}), "", "", "", "DATABASE_URL"},
		{"mysql scheme", base(map[string]string{"DATABASE_URL": "mysql://u:s3cr3t-pw@h/db"}), "", "", "", "DATABASE_URL"},
		{"keyword dsn", base(map[string]string{"DATABASE_URL": "host=db password=s3cr3t-pw"}), "", "", "", "DATABASE_URL"},
		{"unparseable url", base(map[string]string{"DATABASE_URL": "postgres://u:s3cr3t-pw@h:bad/db"}), "", "", "", "DATABASE_URL"},
		{"missing CURSOR_KEY", base(map[string]string{"CURSOR_KEY": ""}), "", "", "", "CURSOR_KEY"},
		{"CURSOR_KEY not base64", base(map[string]string{"CURSOR_KEY": "***not base64***"}), "", "", "", "CURSOR_KEY"},
		{"CURSOR_KEY too short", base(map[string]string{"CURSOR_KEY": base64.RawURLEncoding.EncodeToString(rawKey[:31])}), "", "", "", "CURSOR_KEY"},
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
			if got.AppEnv != tc.wantEnv || got.Addr != tc.wantAdr || got.DatabaseURL != tc.wantDB || !bytes.Equal(got.CursorKey, rawKey) {
				t.Fatalf("got %+v", got)
			}
		})
	}
}

func TestLoadErrorsNeverEchoSecrets(t *testing.T) {
	secrets := []string{"s3cr3t-pw", dbURL, keyURL, keyStd, "***not base64***"}
	bad := []map[string]string{
		base(map[string]string{"DATABASE_URL": "mysql://u:s3cr3t-pw@h/db"}),
		base(map[string]string{"DATABASE_URL": "host=db password=s3cr3t-pw"}),
		base(map[string]string{"DATABASE_URL": "postgres://u:s3cr3t-pw@h:bad/db"}),
		base(map[string]string{"CURSOR_KEY": "***not base64***"}),
		base(map[string]string{"CURSOR_KEY": base64.RawURLEncoding.EncodeToString(rawKey[:31])}),
	}
	for i, m := range bad {
		_, err := config.Load(env(m))
		if err == nil {
			t.Fatalf("case %d: expected error", i)
		}
		for _, s := range secrets {
			if strings.Contains(err.Error(), s) {
				t.Fatalf("case %d: error %q echoes a secret", i, err)
			}
		}
	}
}

func TestConfigStringRedactsSecrets(t *testing.T) {
	cfg, err := config.Load(env(base(nil)))
	if err != nil {
		t.Fatal(err)
	}
	for _, out := range []string{cfg.String(), fmt.Sprintf("%v", cfg), fmt.Sprintf("%+v", cfg)} {
		for _, s := range []string{"s3cr3t-pw", keyURL, "db:5432"} {
			if strings.Contains(out, s) {
				t.Fatalf("%q leaks %q", out, s)
			}
		}
		if !strings.Contains(out, "local-dev") {
			t.Fatalf("%q lost AppEnv", out)
		}
	}
}

// slog.Any("cfg", cfg) and %#v must not print the database URL or the cursor key.
func TestConfigIsRedactedForSlogAndGoSyntax(t *testing.T) {
	cfg := config.Config{AppEnv: "local-dev", Addr: ":8080", DatabaseURL: "postgres://u:hunter2@db/x", CursorKey: []byte("0123456789abcdef0123456789abcdef")}
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("cfg", "config", cfg)
	for name, out := range map[string]string{"slog": buf.String(), "%#v": fmt.Sprintf("%#v", cfg), "%+v": fmt.Sprintf("%+v", cfg)} {
		if strings.Contains(out, "hunter2") || strings.Contains(out, "0123456789abcdef") {
			t.Fatalf("%s leaks a secret: %s", name, out)
		}
	}
}
