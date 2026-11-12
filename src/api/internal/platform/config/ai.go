package config

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
)

// AI is the assistant provider configuration. APIKey is a secret: it never appears in errors,
// String, GoString or LogValue.
type AI struct {
	Enabled bool
	APIKey  string // Google AI Studio key (secret)
	Model   string // provider model id; there is no default (the exact Gemma 4 id is unverified)
	Timeout time.Duration
}

const (
	defaultAITimeout = 30 * time.Second
	minAITimeout     = time.Second
	maxAITimeout     = 120 * time.Second
	maxAIKeyLen      = 512
)

// LoadAI reads the AI_* variables. It returns nil unless AI_ENABLED is true. When enabled,
// AI_API_KEY and AI_MODEL are required. Error messages name the variable, never its value.
func LoadAI(getenv func(string) string) (*AI, error) {
	switch getenv("AI_ENABLED") {
	case "", "false":
		return nil, nil //nolint:nilnil // nil means "assistant not enabled", not an error
	case "true":
	default:
		return nil, errors.New("AI_ENABLED must be true or false")
	}
	key := getenv("AI_API_KEY")
	if key == "" {
		return nil, errors.New("AI_API_KEY is required when AI_ENABLED is true")
	}
	if len(key) > maxAIKeyLen || strings.ContainsFunc(key, isNotKeyChar) {
		return nil, errors.New("AI_API_KEY must be printable ASCII without spaces, at most 512 characters")
	}
	model := getenv("AI_MODEL")
	if model == "" {
		return nil, errors.New("AI_MODEL is required when AI_ENABLED is true (there is no default model)")
	}
	if !llm.ValidModelName(model) {
		return nil, errors.New("AI_MODEL must be 1 to 64 characters among letters, digits, dot, underscore and hyphen")
	}
	timeout := defaultAITimeout
	if raw := getenv("AI_TIMEOUT"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d < minAITimeout || d > maxAITimeout {
			return nil, errors.New("AI_TIMEOUT must be a duration between 1s and 120s")
		}
		timeout = d
	}
	return &AI{Enabled: true, APIKey: key, Model: model, Timeout: timeout}, nil
}

func isNotKeyChar(r rune) bool { return r <= ' ' || r >= 0x7f }

// LogValue keeps slog from printing the key.
func (a AI) LogValue() slog.Value {
	return slog.GroupValue(slog.Bool("Enabled", a.Enabled), slog.String("APIKey", "[redacted]"),
		slog.String("Model", a.Model), slog.Duration("Timeout", a.Timeout))
}

// GoString is the redacted form used by %#v.
func (a AI) GoString() string { return a.String() }

// String is the redacted form used by %v / %+v and logs.
func (a AI) String() string {
	return fmt.Sprintf("{Enabled:%t APIKey:[redacted] Model:%s Timeout:%s}", a.Enabled, a.Model, a.Timeout)
}
