// Package discover is the catalogue read and error report module.
package discover

import (
	"log/slog"

	"github.com/Y4NN777/226os/src/api/internal/discover/repository"
	"github.com/Y4NN777/226os/src/api/internal/discover/service"
	"github.com/Y4NN777/226os/src/api/internal/discover/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// NewHandler wires the published-catalogue reads and the member error report. limiter counts error
// reports (60 per user per minute). A nil logger discards logs.
func NewHandler(db dbgen.DBTX, codec *cursor.Codec, limiter service.Limiter, logger *slog.Logger) *transport.Handler {
	repo := repository.New(db)
	return transport.New(service.New(repo, codec), service.NewReports(repo, limiter), logger)
}
