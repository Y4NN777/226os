// Package solve is the Solve module (S3, US-SOLVE-001): questions and answers.
package solve

import (
	"log/slog"

	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/solve/repository"
	"github.com/Y4NN777/226os/src/api/internal/solve/service"
	"github.com/Y4NN777/226os/src/api/internal/solve/transport"
)

// NewHandler wires the question and answer operations. limiter counts writes on the shared write:user
// budget (60 per user per minute). A nil logger discards logs.
func NewHandler(db repository.Beginner, codec *cursor.Codec, limiter service.Limiter, logger *slog.Logger) *transport.Handler {
	return transport.New(service.New(repository.New(db), codec, limiter), logger)
}
