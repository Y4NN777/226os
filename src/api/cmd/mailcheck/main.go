// Command mailcheck sends one test mail through the configured SMTP server (Mailpit in local-dev):
// mailcheck <recipient>. It reads APP_ENV and the SMTP_* variables.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Y4NN777/226os/src/api/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	app.MailCheckMain(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr, os.Exit)
}
