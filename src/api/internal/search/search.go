// Package search is the cross-content search module: one ranked, bounded query per content group.
package search

import (
	"log/slog"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
	"github.com/Y4NN777/226os/src/api/internal/search/repository"
	"github.com/Y4NN777/226os/src/api/internal/search/service"
	"github.com/Y4NN777/226os/src/api/internal/search/transport"
)

// NewHandler wires GET /search. limiter counts anonymous requests (60 per client network per minute).
// A nil logger discards logs.
func NewHandler(db dbgen.DBTX, limiter service.Limiter, logger *slog.Logger) *transport.Handler {
	return transport.New(service.New(repository.New(db), limiter, logger), logger)
}
