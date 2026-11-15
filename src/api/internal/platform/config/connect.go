package config

import "errors"

// LoadGitHubImport reads CONNECT_GITHUB_IMPORT, the switch of the daily committers.top import (ADR-0016).
// Unset means on in staging-test and production-live and off in local-dev; true or false overrides it.
func LoadGitHubImport(getenv func(string) string, appEnv string) (bool, error) {
	switch getenv("CONNECT_GITHUB_IMPORT") {
	case "":
		return appEnv != "local-dev", nil
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("CONNECT_GITHUB_IMPORT must be true or false")
	}
}
