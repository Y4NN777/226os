// Package discover is the catalogue read module.
package discover

import (
	"log/slog"

	"github.com/Y4NN777/226os/src/api/internal/discover/repository"
	"github.com/Y4NN777/226os/src/api/internal/discover/service"
	"github.com/Y4NN777/226os/src/api/internal/discover/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// NewHandler wires the published-catalogue reads. A nil logger discards logs.
func NewHandler(db dbgen.DBTX, codec *cursor.Codec, logger *slog.Logger) *transport.Handler {
	return transport.New(service.New(repository.New(db), codec), logger)
}
