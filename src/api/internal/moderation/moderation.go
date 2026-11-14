// Package moderation wires the moderator-only S6 operations.
package moderation

import (
	"log/slog"

	"github.com/Y4NN777/226os/src/api/internal/moderation/repository"
	"github.com/Y4NN777/226os/src/api/internal/moderation/service"
	"github.com/Y4NN777/226os/src/api/internal/moderation/transport"
)

func NewHandler(db repository.Beginner, limiter service.Limiter, logger *slog.Logger) *transport.Handler {
	return transport.New(service.New(repository.New(db), limiter), logger)
}
