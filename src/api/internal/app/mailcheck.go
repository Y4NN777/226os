package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

// Exit codes of MailCheck.
const (
	exitSendFailed = 1
	exitUsage      = 2
)

// MailCheck sends one test mail to args[0] through the configured SMTP server (Mailpit in local-dev).
// It reads APP_ENV and the SMTP_* variables only, so it needs no database URL or cursor key.
// It returns 0 on success, 2 for a usage or configuration problem, 1 when the send fails. Messages
// on errOut carry no credential: configuration errors name variables, send errors a phase and a code.
func MailCheck(ctx context.Context, args []string, getenv func(string) string, out, errOut io.Writer) int {
	if len(args) != 1 {
		say(errOut, "usage: mailcheck <recipient>\n")
		return exitUsage
	}
	appEnv := getenv("APP_ENV")
	smtp, err := config.LoadSMTP(getenv, appEnv)
	if err != nil {
		say(errOut, "mail check: invalid configuration: %v\n", err)
		return exitUsage
	}
	if smtp == nil {
		say(errOut, "mail check: SMTP is not configured (set SMTP_HOST, SMTP_PORT, SMTP_FROM)\n")
		return exitUsage
	}
	msg := mailer.Message{
		To:      []string{args[0]},
		Subject: "226os mail check " + time.Now().UTC().Format(time.RFC3339),
		Text:    fmt.Sprintf("This is a test message from 226os (APP_ENV=%s) sent by the mailcheck command.\nNo action is needed.\n", appEnv),
	}
	if err := mailer.NewSMTP(smtpConfig(smtp)).Send(ctx, msg); err != nil {
		say(errOut, "mail check: send failed: %v\n", err)
		return exitSendFailed
	}
	say(out, "mail check: sent to %s\n", args[0])
	return 0
}

// MailCheckMain runs MailCheck and calls exit only on failure (exit(0) would be a no-op).
func MailCheckMain(ctx context.Context, args []string, getenv func(string) string, out, errOut io.Writer, exit func(int)) {
	if code := MailCheck(ctx, args, getenv, out, errOut); code != 0 {
		exit(code)
	}
}

// say writes a status line; a failing terminal leaves nothing better to do than the exit code.
func say(w io.Writer, format string, args ...any) { _, _ = fmt.Fprintf(w, format, args...) }
