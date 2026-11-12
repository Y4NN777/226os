// Command api is the 226os HTTP API entrypoint.
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
	app.Main(ctx, os.Getenv, os.Stdout, os.Exit)
}
