package main

import (
	"os"
	"strconv"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/mailer/smtptest"
)

// main exits only on failure, so the success path is the one a test can run in-process.
func TestMainSendsTheMailAndReturns(t *testing.T) {
	s := smtptest.NewServer(t)
	t.Setenv("APP_ENV", "local-dev")
	t.Setenv("SMTP_HOST", s.Host())
	t.Setenv("SMTP_PORT", strconv.Itoa(s.Port()))
	t.Setenv("SMTP_TLS", "none")
	t.Setenv("SMTP_FROM", "no-reply@226os.local")
	old := os.Args
	os.Args = []string{"mailcheck", "dev@example.net"}
	t.Cleanup(func() { os.Args = old })

	main()

	if got := s.Messages(); len(got) != 1 || got[0].To[0] != "dev@example.net" {
		t.Fatalf("messages = %+v", got)
	}
}
