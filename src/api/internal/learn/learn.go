// Package learn is the Learn module (S4, US-LEARN-001): articles, tutorials and the free training series.
package learn

import (
	"log/slog"

	"github.com/Y4NN777/226os/src/api/internal/learn/repository"
	"github.com/Y4NN777/226os/src/api/internal/learn/service"
	"github.com/Y4NN777/226os/src/api/internal/learn/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
)

// NewHandler wires the Learn reads and article creation. limiter counts article creation on the
// shared write:user budget (60 per user per minute). A nil logger discards logs.
func NewHandler(db repository.Beginner, codec *cursor.Codec, limiter service.Limiter, logger *slog.Logger) *transport.Handler {
	return transport.New(service.New(repository.New(db), codec, limiter), logger)
}
