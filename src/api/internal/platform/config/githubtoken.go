package config

import (
	"errors"
	"strings"
)

const maxGitHubTokenLen = 255

// LoadGitHubToken reads the optional GITHUB_TOKEN. Unset or empty means unauthenticated GitHub calls.
// A token is a secret: the error names the variable, never its value.
func LoadGitHubToken(getenv func(string) string) (string, error) {
	token := getenv("GITHUB_TOKEN")
	if len(token) > maxGitHubTokenLen || strings.ContainsFunc(token, isNotKeyChar) {
		return "", errors.New("GITHUB_TOKEN must be printable ASCII without spaces, at most 255 characters")
	}
	return token, nil
}
