package config_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
)

// Fixture credentials are built at run time so no secret-shaped literal sits in the source.
var (
	smtpUser = fmt.Sprintf("mailer-%d@example.org", 42)
	smtpPass = fmt.Sprintf("pw-%d-%s", 7319, "zebra")
)

func smtpEnv(extra map[string]string) map[string]string {
	m := map[string]string{
		"SMTP_HOST": "smtp.example.org", "SMTP_PORT": "587", "SMTP_FROM": "no-reply@example.org",
		"SMTP_USERNAME": smtpUser, "SMTP_PASSWORD": smtpPass,
	}
	for k, v := range extra {
		m[k] = v
	}
	return m
}

func TestLoadSMTPNilWhenNothingSet(t *testing.T) {
	got, err := config.LoadSMTP(env(map[string]string{}), "local-dev")
	if err != nil || got != nil {
		t.Fatalf("got %v, %v; want nil, nil", got, err)
	}
}

func TestLoadSMTPValid(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		appEnv  string
		wantTLS string
	}{
		{"default starttls with credentials", smtpEnv(nil), "production-live", "starttls"},
		{"explicit tls", smtpEnv(map[string]string{"SMTP_TLS": "tls", "SMTP_PORT": "465"}), "staging-test", "tls"},
		{"none in local-dev without credentials", smtpEnv(map[string]string{"SMTP_TLS": "none", "SMTP_USERNAME": "", "SMTP_PASSWORD": "", "SMTP_HOST": "mailpit", "SMTP_PORT": "1025"}), "local-dev", "none"},
		{"ip host", smtpEnv(map[string]string{"SMTP_HOST": "127.0.0.1"}), "local-dev", "starttls"},
		{"ipv6 host", smtpEnv(map[string]string{"SMTP_HOST": "::1"}), "local-dev", "starttls"},
		{"none on localhost", smtpEnv(map[string]string{"SMTP_TLS": "none", "SMTP_USERNAME": "", "SMTP_PASSWORD": "", "SMTP_HOST": "localhost"}), "local-dev", "none"},
		{"none on 127.0.0.1", smtpEnv(map[string]string{"SMTP_TLS": "none", "SMTP_USERNAME": "", "SMTP_PASSWORD": "", "SMTP_HOST": "127.0.0.1"}), "local-dev", "none"},
		{"none on ::1", smtpEnv(map[string]string{"SMTP_TLS": "none", "SMTP_USERNAME": "", "SMTP_PASSWORD": "", "SMTP_HOST": "::1"}), "local-dev", "none"},
		{"no credentials starttls", smtpEnv(map[string]string{"SMTP_USERNAME": "", "SMTP_PASSWORD": ""}), "local-dev", "starttls"},
		{"port 65535", smtpEnv(map[string]string{"SMTP_PORT": "65535"}), "local-dev", "starttls"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := config.LoadSMTP(env(tc.env), tc.appEnv)
			if err != nil || got == nil {
				t.Fatalf("got %v, %v", got, err)
			}
			if got.TLS != tc.wantTLS || got.Host != tc.env["SMTP_HOST"] || got.From != tc.env["SMTP_FROM"] ||
				got.Username != tc.env["SMTP_USERNAME"] || got.Password != tc.env["SMTP_PASSWORD"] {
				t.Fatalf("got %#v", got)
			}
		})
	}
}

func TestLoadSMTPRejects(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		appEnv  string
		wantErr string
	}{
		{"only host", map[string]string{"SMTP_HOST": "h"}, "local-dev", "SMTP_PORT"},
		{"missing host", smtpEnv(map[string]string{"SMTP_HOST": ""}), "local-dev", "SMTP_HOST"},
		{"missing port", smtpEnv(map[string]string{"SMTP_PORT": ""}), "local-dev", "SMTP_PORT"},
		{"port not a number", smtpEnv(map[string]string{"SMTP_PORT": "abc"}), "local-dev", "SMTP_PORT"},
		{"port zero", smtpEnv(map[string]string{"SMTP_PORT": "0"}), "local-dev", "SMTP_PORT"},
		{"port too high", smtpEnv(map[string]string{"SMTP_PORT": "65536"}), "local-dev", "SMTP_PORT"},
		{"missing from", smtpEnv(map[string]string{"SMTP_FROM": ""}), "local-dev", "SMTP_FROM"},
		{"from invalid", smtpEnv(map[string]string{"SMTP_FROM": "not an address"}), "local-dev", "SMTP_FROM"},
		{"from with newline", smtpEnv(map[string]string{"SMTP_FROM": "a@example.org\nBcc: x@example.org"}), "local-dev", "SMTP_FROM"},
		{"from with NUL", smtpEnv(map[string]string{"SMTP_FROM": "a@example.org\x00"}), "local-dev", "SMTP_FROM"},
		{"username without password", smtpEnv(map[string]string{"SMTP_PASSWORD": ""}), "local-dev", "SMTP_USERNAME"},
		{"password without username", smtpEnv(map[string]string{"SMTP_USERNAME": ""}), "local-dev", "SMTP_USERNAME"},
		{"bad tls", smtpEnv(map[string]string{"SMTP_TLS": "ssl"}), "local-dev", "SMTP_TLS"},
		{"from display name", smtpEnv(map[string]string{"SMTP_FROM": "226os <no-reply@example.org>"}), "local-dev", "SMTP_FROM"},
		{"from angle brackets", smtpEnv(map[string]string{"SMTP_FROM": "<no-reply@example.org>"}), "local-dev", "SMTP_FROM"},
		{"from quoted injection", smtpEnv(map[string]string{"SMTP_FROM": `"a> BODY=8BITMIME"@example.org`}), "local-dev", "SMTP_FROM"},
		{"from non-ASCII", smtpEnv(map[string]string{"SMTP_FROM": "é@example.org"}), "local-dev", "SMTP_FROM"},
		{"from NEL", smtpEnv(map[string]string{"SMTP_FROM": "a@example.org\u0085"}), "local-dev", "SMTP_FROM"},
		{"none on a remote host in local-dev", smtpEnv(map[string]string{"SMTP_TLS": "none", "SMTP_USERNAME": "", "SMTP_PASSWORD": "", "SMTP_HOST": "smtp.gmail.com"}), "local-dev", "SMTP_TLS"},
		{"none on an arbitrary IP in local-dev", smtpEnv(map[string]string{"SMTP_TLS": "none", "SMTP_USERNAME": "", "SMTP_PASSWORD": "", "SMTP_HOST": "192.0.2.10"}), "local-dev", "SMTP_TLS"},
		{"none outside local-dev", smtpEnv(map[string]string{"SMTP_TLS": "none", "SMTP_USERNAME": "", "SMTP_PASSWORD": ""}), "staging-test", "SMTP_TLS"},
		{"none with credentials in local-dev", smtpEnv(map[string]string{"SMTP_TLS": "none", "SMTP_HOST": "mailpit"}), "local-dev", "SMTP_TLS"},
		{"host with scheme", smtpEnv(map[string]string{"SMTP_HOST": "smtp://mail.example.org"}), "local-dev", "SMTP_HOST"},
		{"host with slash", smtpEnv(map[string]string{"SMTP_HOST": "mail.example.org/x"}), "local-dev", "SMTP_HOST"},
		{"host with space", smtpEnv(map[string]string{"SMTP_HOST": "mail example.org"}), "local-dev", "SMTP_HOST"},
		{"host with colon and port", smtpEnv(map[string]string{"SMTP_HOST": "mail.example.org:25"}), "local-dev", "SMTP_HOST"},
		{"host leading dash", smtpEnv(map[string]string{"SMTP_HOST": "-mail.example.org"}), "local-dev", "SMTP_HOST"},
		{"host empty label", smtpEnv(map[string]string{"SMTP_HOST": "mail..example.org"}), "local-dev", "SMTP_HOST"},
		{"host too long", smtpEnv(map[string]string{"SMTP_HOST": strings.Repeat("a", 254)}), "local-dev", "SMTP_HOST"},
		{"host label too long", smtpEnv(map[string]string{"SMTP_HOST": strings.Repeat("a", 64) + ".org"}), "local-dev", "SMTP_HOST"},
		{"host underscore", smtpEnv(map[string]string{"SMTP_HOST": "mail_1.example.org"}), "local-dev", "SMTP_HOST"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := config.LoadSMTP(env(tc.env), tc.appEnv)
			if err == nil || got != nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("got %v, %v; want error containing %q", got, err, tc.wantErr)
			}
			for _, s := range []string{smtpUser, smtpPass} {
				if strings.Contains(err.Error(), s) {
					t.Fatalf("error %q leaks a credential", err)
				}
			}
		})
	}
}

func TestSMTPIsRedacted(t *testing.T) {
	s, err := config.LoadSMTP(env(smtpEnv(nil)), "production-live")
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{AppEnv: "production-live", Addr: ":1", SMTP: s}
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("x", "smtp", s, "smtpval", *s, "config", cfg)
	outs := map[string]string{
		"slog": buf.String(),
		"%v":   fmt.Sprintf("%v %v", s, *s), "%+v": fmt.Sprintf("%+v %+v", s, *s), "%#v": fmt.Sprintf("%#v %#v", s, *s),
		"cfg %v": fmt.Sprintf("%v", cfg), "cfg %+v": fmt.Sprintf("%+v", cfg), "cfg %#v": fmt.Sprintf("%#v", cfg),
		"String": s.String() + cfg.String() + s.GoString(),
	}
	for name, out := range outs {
		for _, secret := range []string{smtpUser, smtpPass} {
			if strings.Contains(out, secret) {
				t.Fatalf("%s leaks %q: %s", name, secret, out)
			}
		}
		if name != "cfg %v" && name != "cfg %+v" && name != "cfg %#v" && !strings.Contains(out, "smtp.example.org") {
			t.Fatalf("%s lost the host: %s", name, out)
		}
	}
	if !strings.Contains(outs["cfg %v"], "smtp.example.org") {
		t.Fatalf("Config String omits the SMTP host: %s", outs["cfg %v"])
	}
}

func TestConfigStringWithoutSMTP(t *testing.T) {
	if out := (config.Config{AppEnv: "local-dev"}).String(); strings.Contains(out, "SMTP:{") {
		t.Fatalf("unexpected SMTP section: %s", out)
	}
}

func TestLoadIncludesSMTP(t *testing.T) {
	got, err := config.Load(env(base(smtpEnv(nil))))
	if err != nil || got.SMTP == nil || got.SMTP.Host != "smtp.example.org" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if got, err = config.Load(env(base(nil))); err != nil || got.SMTP != nil {
		t.Fatalf("got %+v, %v; want nil SMTP", got, err)
	}
	if _, err = config.Load(env(base(map[string]string{"SMTP_HOST": "h"}))); err == nil || !strings.Contains(err.Error(), "SMTP_PORT") {
		t.Fatalf("err = %v", err)
	}
}
