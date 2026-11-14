// Package build is the Build module (S5, US-BUILD-001): projects registered from a public GitHub
// repository and their contribution needs.
package build

import (
	"log/slog"

	"github.com/Y4NN777/226os/src/api/internal/build/repository"
	"github.com/Y4NN777/226os/src/api/internal/build/service"
	"github.com/Y4NN777/226os/src/api/internal/build/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

// NewHandler wires projects and contribution needs. gh reads GitHub (github.New in production, a
// fake in tests); limiter counts the shared write:user budget and the GitHub import budget. A nil
// logger discards logs.
func NewHandler(db repository.Beginner, gh service.GitHub, codec *cursor.Codec, limiter service.Limiter, logger *slog.Logger) *transport.Handler {
	return transport.New(service.New(repository.New(db), gh, codec, limiter), logger)
}
