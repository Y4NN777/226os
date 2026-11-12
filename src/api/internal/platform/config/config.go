// Package config parses and validates process configuration from the environment.
package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
)

// Config is the validated runtime configuration of the API process.
// DatabaseURL and CursorKey are secrets: they are never echoed in errors and
// are redacted by String.
type Config struct {
	AppEnv      string // local-dev, staging-test or production-live (ADR-0006)
	Addr        string // TCP listen address
	DatabaseURL string // postgres:// or postgresql:// connection URL (secret)
	CursorKey   []byte // HMAC key for pagination cursors, at least 32 bytes (secret)
	SMTP        *SMTP  // nil when no SMTP_* variable is set; Username and Password are secrets
	AI          *AI    // nil unless AI_ENABLED is true; APIKey is a secret
}

const (
	defaultAddr  = ":8080"
	minCursorKey = 32
)

// Load reads configuration through getenv and fails fast on invalid values.
// Error messages name the variable, never its value.
func Load(getenv func(string) string) (Config, error) {
	env := getenv("APP_ENV")
	switch env {
	case "local-dev", "staging-test", "production-live":
	default:
		return Config{}, fmt.Errorf("APP_ENV must be local-dev, staging-test or production-live, got %q", env)
	}
	addr := getenv("API_ADDR")
	if addr == "" {
		addr = defaultAddr
	}
	dbURL := getenv("DATABASE_URL")
	if err := checkDatabaseURL(dbURL); err != nil {
		return Config{}, err
	}
	key, err := decodeCursorKey(getenv("CURSOR_KEY"))
	if err != nil {
		return Config{}, err
	}
	smtp, err := LoadSMTP(getenv, env)
	if err != nil {
		return Config{}, err
	}
	ai, err := LoadAI(getenv)
	if err != nil {
		return Config{}, err
	}
	return Config{AppEnv: env, Addr: addr, DatabaseURL: dbURL, CursorKey: key, SMTP: smtp, AI: ai}, nil
}

func checkDatabaseURL(raw string) error {
	if raw == "" {
		return errors.New("DATABASE_URL is required")
	}
	// url.Parse errors embed the input, so its error text is discarded.
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		return errors.New("DATABASE_URL must be a postgres:// or postgresql:// URL")
	}
	return nil
}

func decodeCursorKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, errors.New("CURSOR_KEY is required")
	}
	for _, enc := range []*base64.Encoding{base64.RawURLEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.StdEncoding} {
		if key, err := enc.DecodeString(raw); err == nil && len(key) >= minCursorKey {
			return key, nil
		}
	}
	return nil, errors.New("CURSOR_KEY must be base64 or base64url encoding of at least 32 bytes")
}

// LogValue keeps slog.Any("config", cfg) from printing the secrets.
func (c Config) LogValue() slog.Value {
	return slog.GroupValue(slog.String("AppEnv", c.AppEnv), slog.String("Addr", c.Addr),
		slog.String("DatabaseURL", "[redacted]"), slog.String("CursorKey", "[redacted]"), c.smtpAttr(), c.aiAttr())
}

func (c Config) smtpAttr() slog.Attr {
	if c.SMTP == nil {
		return slog.String("SMTP", "none")
	}
	return slog.Any("SMTP", *c.SMTP)
}

func (c Config) aiAttr() slog.Attr {
	if c.AI == nil {
		return slog.String("AI", "none")
	}
	return slog.Any("AI", *c.AI)
}

// GoString is the redacted form used by %#v.
func (c Config) GoString() string { return c.String() }

// String is the redacted form used by %v / %+v and logs.
func (c Config) String() string {
	out := fmt.Sprintf("{AppEnv:%s Addr:%s DatabaseURL:[redacted] CursorKey:[redacted]", c.AppEnv, c.Addr)
	if c.SMTP != nil {
		out += " SMTP:" + c.SMTP.String()
	}
	if c.AI != nil {
		out += " AI:" + c.AI.String()
	}
	return out + "}"
}
