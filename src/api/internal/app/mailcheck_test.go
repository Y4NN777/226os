package app_test

import (
	"bytes"
	"context"
	"fmt"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/app"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer/smtptest"
)

func mailEnv(s *smtptest.Server, extra map[string]string) func(string) string {
	m := map[string]string{
		"APP_ENV": "local-dev", "SMTP_HOST": s.Host(), "SMTP_PORT": strconv.Itoa(s.Port()),
		"SMTP_TLS": "none", "SMTP_FROM": "no-reply@226os.local",
	}
	for k, v := range extra {
		m[k] = v
	}
	return func(k string) string { return m[k] }
}

func runCheck(env func(string) string, args ...string) (code int, out, errOut string) {
	var o, e bytes.Buffer
	code = app.MailCheck(context.Background(), args, env, &o, &e)
	return code, o.String(), e.String()
}

func TestMailCheckSendsOneMessage(t *testing.T) {
	s := smtptest.NewServer(t)
	before := time.Now().UTC().Add(-time.Second)
	code, out, errOut := runCheck(mailEnv(s, nil), "dev@example.net")
	if code != 0 || out != "mail check: sent to dev@example.net\n" || errOut != "" {
		t.Fatalf("code = %d out = %q err = %q", code, out, errOut)
	}
	got := s.Messages()
	if len(got) != 1 || got[0].From != "no-reply@226os.local" || len(got[0].To) != 1 || got[0].To[0] != "dev@example.net" {
		t.Fatalf("messages = %+v", got)
	}
	msg, err := mail.ReadMessage(strings.NewReader(got[0].Data))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`^226os mail check (\S+)$`).FindStringSubmatch(msg.Header.Get("Subject"))
	if m == nil {
		t.Fatalf("Subject = %q", msg.Header.Get("Subject"))
	}
	if ts, err := time.Parse(time.RFC3339, m[1]); err != nil || ts.Before(before) || ts.After(time.Now().Add(time.Second)) {
		t.Fatalf("subject time %q: %v", m[1], err)
	}
	var body bytes.Buffer
	if _, err := body.ReadFrom(quotedprintable.NewReader(msg.Body)); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body.String(), "local-dev") || !strings.Contains(strings.ToLower(body.String()), "test") {
		t.Fatalf("body = %q", body.String())
	}
}

func TestMailCheckFailures(t *testing.T) {
	pw := fmt.Sprintf("pw-%d-%s", 3301, "marten")
	tests := []struct {
		name     string
		server   []smtptest.Option
		extra    map[string]string
		args     []string
		wantCode int
		wantErr  string
	}{
		{"no argument", nil, nil, nil, 2, "usage"},
		{"two arguments", nil, nil, []string{"a@example.net", "b@example.net"}, 2, "usage"},
		{"smtp not configured", nil, map[string]string{"SMTP_HOST": "", "SMTP_PORT": "", "SMTP_TLS": "", "SMTP_FROM": ""}, []string{"a@example.net"}, 2, "not configured"},
		{"invalid smtp configuration", nil, map[string]string{"SMTP_PORT": "abc"}, []string{"a@example.net"}, 2, "SMTP_PORT"},
		{"none refused outside local-dev", nil, map[string]string{"APP_ENV": "staging-test"}, []string{"a@example.net"}, 2, "SMTP_TLS"},
		{"none with credentials refused", nil, map[string]string{"SMTP_USERNAME": "u", "SMTP_PASSWORD": pw}, []string{"a@example.net"}, 2, "SMTP_TLS"},
		{"invalid recipient", nil, nil, []string{"not an address"}, 1, "invalid message"},
		{"server rejects the recipient", []smtptest.Option{smtptest.WithFail("RCPT", 550)}, nil, []string{"a@example.net"}, 1, "550"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := smtptest.NewServer(t, tc.server...)
			code, out, errOut := runCheck(mailEnv(s, tc.extra), tc.args...)
			if code != tc.wantCode || out != "" || !strings.Contains(errOut, tc.wantErr) {
				t.Fatalf("code = %d out = %q err = %q", code, out, errOut)
			}
			if strings.Contains(errOut, pw) || len(s.Messages()) != 0 {
				t.Fatalf("leak or delivery: %q", errOut)
			}
		})
	}
}

func TestMailCheckMainExitsOnlyOnFailure(t *testing.T) {
	s := smtptest.NewServer(t)
	var out, errOut bytes.Buffer
	exited := -1
	app.MailCheckMain(context.Background(), []string{"dev@example.net"}, mailEnv(s, nil), &out, &errOut, func(c int) { exited = c })
	if exited != -1 || !strings.Contains(out.String(), "sent to") {
		t.Fatalf("exit = %d out = %q", exited, out.String())
	}
	app.MailCheckMain(context.Background(), nil, mailEnv(s, nil), &out, &errOut, func(c int) { exited = c })
	if exited != 2 {
		t.Fatalf("exit = %d", exited)
	}
}
