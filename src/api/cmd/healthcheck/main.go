// Command healthcheck probes the API running in the same container (the distroless image has no shell or wget):
//
//	healthcheck            GET /healthz on the API address (API_ADDR, default :8080), exit 0 on 2xx
//	healthcheck /readyz    the readiness probe (checks the database)
//
// It is the image HEALTHCHECK, the Compose health check and the smoke test of docs/SHIP.md.
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Y4NN777/226os/src/api/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	app.HealthCheckMain(ctx, os.Args[1:], os.Getenv, http.DefaultClient, os.Stderr, os.Exit)
}
