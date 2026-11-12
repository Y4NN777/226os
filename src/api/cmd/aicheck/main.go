// Command aicheck verifies the AI provider configuration with real calls (owner checks V1, V2, V4):
//
//	aicheck models         list the models whose name contains "gemma"
//	aicheck ping [fr|en]   one fixed harmless prompt per language
//	aicheck limits         three tiny probes, never a loop
//
// It reads only the AI_* variables.
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
	app.AICheckMain(ctx, os.Args[1:], os.Getenv, os.Stdout, os.Stderr, os.Exit)
}
