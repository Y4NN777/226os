package config

import (
	"errors"
	"net/url"
	"strings"
)

// LoadBaseURL reads APP_BASE_URL, the public origin that verification and reset mails link to. It returns
// "" when the variable is empty (the mails then carry the bare code only). A set value must be an absolute
// http or https origin: no user info, path, query or fragment, one trailing slash tolerated and dropped.
// staging-test and production-live require https. Errors name the variable and never its value.
func LoadBaseURL(getenv func(string) string, appEnv string) (string, error) {
	raw := getenv("APP_BASE_URL")
	if raw == "" {
		return "", nil
	}
	bad := errors.New("APP_BASE_URL must be an absolute http or https origin (scheme and host, optional port) without user info, path, query or fragment")
	// A bare "?" or "#" parses to an empty query or fragment, and surrounding whitespace parses silently:
	// refuse them on the raw text. url.Parse errors embed the input, so their text is discarded.
	if strings.ContainsAny(raw, "?# \t\r\n") {
		return "", bad
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" ||
		u.User != nil || (u.Path != "" && u.Path != "/") {
		return "", bad
	}
	if u.Scheme != "https" && appEnv != "local-dev" {
		return "", errors.New("APP_BASE_URL must use https in staging-test and production-live")
	}
	return u.Scheme + "://" + u.Host, nil
}
