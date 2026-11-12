// Package config parses and validates process configuration from the environment.
package config

import "fmt"

// Config is the validated runtime configuration of the API process.
type Config struct {
	AppEnv string // local-dev, staging-test or production-live (ADR-0006)
	Addr   string // TCP listen address
}

const defaultAddr = ":8080"

// Load reads configuration through getenv and fails fast on invalid values.
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
	return Config{AppEnv: env, Addr: addr}, nil
}
