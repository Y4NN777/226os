package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer/smtptest"
)

var (
	mailUser = fmt.Sprintf("sender-%d@example.org", 77)
	mailPass = fmt.Sprintf("pw-%d-%s", 8123, "ocelot")
)

func logger(buf *bytes.Buffer) *slog.Logger { return slog.New(slog.NewJSONHandler(buf, nil)) }

var testMsg = mailer.Message{To: []string{"bob@example.net"}, Subject: "s", Text: "t"}

func TestNewMailerIsDisabledWithoutSMTPConfig(t *testing.T) {
	var buf bytes.Buffer
	m := newMailer(config.Config{AppEnv: "local-dev"}, logger(&buf))
	if _, ok := m.(mailer.Disabled); !ok {
		t.Fatalf("mailer = %T, want Disabled", m)
	}
	if err := m.Send(t.Context(), testMsg); !errors.Is(err, mailer.ErrNotConfigured) {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(buf.String(), "mail not configured") {
		t.Fatalf("log = %s", buf.String())
	}
}

func TestNewMailerSendsThroughSMTPWithoutLoggingSecrets(t *testing.T) {
	s := smtptest.NewServer(t)
	var buf bytes.Buffer
	cfg := config.Config{AppEnv: "local-dev", SMTP: &config.SMTP{Host: s.Host(), Port: s.Port(), From: "no-reply@example.org", TLS: "none"}}
	m := newMailer(cfg, logger(&buf))
	if err := m.Send(t.Context(), testMsg); err != nil {
		t.Fatal(err)
	}
	if got := s.Messages(); len(got) != 1 || got[0].From != "no-reply@example.org" || got[0].To[0] != "bob@example.net" {
		t.Fatalf("messages = %+v", got)
	}
	log := buf.String()
	if !strings.Contains(log, "mail configured") || !strings.Contains(log, s.Host()) || !strings.Contains(log, `"none"`) || strings.Contains(log, "not configured") {
		t.Fatalf("log = %s", log)
	}
	if strings.Count(log, "\n") != 1 {
		t.Fatalf("want exactly one log line, got %s", log)
	}
	if strings.Contains(log, "no-reply") {
		t.Fatalf("log leaks the sender: %s", log)
	}
}

// The mapping must carry the TLS mode and the credentials: a STARTTLS adapter against a server
// that does not offer it fails closed and never sends AUTH in clear.
func TestNewMailerMapsModeAndCredentials(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithAuth(mailUser, mailPass))
	cfg := config.Config{SMTP: &config.SMTP{Host: s.Host(), Port: s.Port(), Username: mailUser, Password: mailPass, From: "no-reply@example.org", TLS: "starttls"}}
	err := newMailer(cfg, logger(&bytes.Buffer{})).Send(t.Context(), testMsg)
	if !errors.Is(err, mailer.ErrPermanent) || len(s.Auths()) != 0 || len(s.Messages()) != 0 {
		t.Fatalf("err = %v, auths = %v", err, s.Auths())
	}
}

func TestRunLogsMailStateWithoutSecrets(t *testing.T) {
	envWith := func(extra map[string]string) func(string) string {
		base := validEnv(t)
		return func(k string) string {
			if v, ok := extra[k]; ok {
				return v
			}
			return base(k)
		}
	}
	serve := func(context.Context, net.Listener, http.Handler, time.Duration) error { return nil }
	t.Run("configured", func(t *testing.T) {
		var out bytes.Buffer
		env := envWith(map[string]string{
			"SMTP_HOST": "smtp.example.org", "SMTP_PORT": "587", "SMTP_USERNAME": mailUser, "SMTP_PASSWORD": mailPass,
			"SMTP_FROM": "no-reply@example.org", "SMTP_TLS": "starttls",
		})
		if code := run(context.Background(), env, &out, serve, fakeOpener(&fakePool{})); code != 0 {
			t.Fatalf("code = %d, log = %s", code, out.String())
		}
		log := out.String()
		if !strings.Contains(log, "mail configured") || !strings.Contains(log, "smtp.example.org") || !strings.Contains(log, "starttls") {
			t.Fatalf("log = %s", log)
		}
		for _, secret := range []string{mailUser, mailPass, "no-reply@example.org", "226os <"} {
			if strings.Contains(log, secret) {
				t.Fatalf("log leaks %q: %s", secret, log)
			}
		}
	})
	t.Run("not configured", func(t *testing.T) {
		var out bytes.Buffer
		if code := run(context.Background(), validEnv(t), &out, serve, fakeOpener(&fakePool{})); code != 0 {
			t.Fatalf("code = %d", code)
		}
		if !strings.Contains(out.String(), "mail not configured") {
			t.Fatalf("log = %s", out.String())
		}
	})
	t.Run("invalid SMTP config fails fast without echoing it", func(t *testing.T) {
		var out bytes.Buffer
		env := envWith(map[string]string{"SMTP_HOST": "smtp.example.org", "SMTP_PASSWORD": mailPass})
		if code := run(context.Background(), env, &out, serve, fakeOpener(&fakePool{})); code != 1 || strings.Contains(out.String(), mailPass) {
			t.Fatalf("code = %d, log = %s", code, out.String())
		}
	})
}

func TestServerCarriesTheMailer(t *testing.T) {
	if _, ok := NewServer().Mailer().(mailer.Disabled); !ok {
		t.Fatal("default mailer is not Disabled")
	}
	f := mailer.NewFake()
	if NewServer(WithMailer(f)).Mailer() != mailer.Mailer(f) {
		t.Fatal("WithMailer ignored")
	}
	deps := newDeps(&fakePool{}, WithMailer(f))
	if srv, ok := deps.Server.(*Server); !ok || srv.Mailer() != mailer.Mailer(f) {
		t.Fatalf("deps server = %T", deps.Server)
	}
}

func TestSMTPConfigMapping(t *testing.T) {
	got := smtpConfig(&config.SMTP{Host: "h", Port: 25, Username: "u", Password: "p", From: "f@example.org", TLS: "tls"})
	want := mailer.SMTPConfig{Host: "h", Port: 25, Username: "u", Password: "p", From: "f@example.org", TLS: mailer.ModeTLS}
	if got != want {
		t.Fatalf("got %+v", got)
	}
}
