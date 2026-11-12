package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strconv"
	"strings"

	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

// SMTP is the outgoing mail configuration. Username and Password are secrets:
// they never appear in errors, String, GoString or LogValue.
type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TLS      string // none, starttls or tls
}

const (
	tlsNone     = "none"
	tlsStartTLS = "starttls"
	tlsImplicit = "tls"
	maxHostLen  = 253
	maxLabelLen = 63
)

// cleartextHosts are the only SMTP hosts that may use SMTP_TLS=none (local-dev only): loopback and
// the Compose service name of the mail catcher. A real provider must never receive cleartext.
var cleartextHosts = map[string]bool{"localhost": true, "127.0.0.1": true, "::1": true, "mailpit": true}

var smtpVars = []string{"SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_FROM", "SMTP_TLS"}

// LoadSMTP reads the SMTP_* variables. It returns nil when none of them is set.
// Error messages name the variable, never its value.
func LoadSMTP(getenv func(string) string, appEnv string) (*SMTP, error) {
	set := false
	for _, v := range smtpVars {
		if getenv(v) != "" {
			set = true
		}
	}
	if !set {
		return nil, nil //nolint:nilnil // nil means "mail not configured", not an error
	}
	host := getenv("SMTP_HOST")
	if !plausibleHost(host) {
		return nil, errors.New("SMTP_HOST must be a host name or IP address (no scheme, port or path)")
	}
	port, err := strconv.Atoi(getenv("SMTP_PORT"))
	if err != nil || port < 1 || port > 65535 {
		return nil, errors.New("SMTP_PORT must be a number between 1 and 65535")
	}
	from := getenv("SMTP_FROM")
	if err := mailer.ValidateAddress(from); err != nil {
		return nil, errors.New("SMTP_FROM must be a bare ASCII email address (no display name, quotes or brackets)")
	}
	user, pass := getenv("SMTP_USERNAME"), getenv("SMTP_PASSWORD")
	if (user == "") != (pass == "") {
		return nil, errors.New("SMTP_USERNAME and SMTP_PASSWORD must be set together")
	}
	mode := getenv("SMTP_TLS")
	if mode == "" {
		mode = tlsStartTLS
	}
	switch mode {
	case tlsStartTLS, tlsImplicit:
	case tlsNone:
		if appEnv != "local-dev" {
			return nil, errors.New("SMTP_TLS none is only allowed when APP_ENV is local-dev")
		}
		if !cleartextHosts[host] {
			return nil, errors.New("SMTP_TLS none is only allowed for localhost, 127.0.0.1, ::1 and mailpit")
		}
		if user != "" {
			return nil, errors.New("SMTP_TLS none cannot be combined with SMTP credentials")
		}
	default:
		return nil, errors.New("SMTP_TLS must be none, starttls or tls")
	}
	return &SMTP{Host: host, Port: port, Username: user, Password: pass, From: from, TLS: mode}, nil
}

// plausibleHost accepts an IP address or an RFC 1123 host name.
func plausibleHost(h string) bool {
	if net.ParseIP(h) != nil {
		return true
	}
	if h == "" || len(h) > maxHostLen {
		return false
	}
	for label := range strings.SplitSeq(h, ".") {
		if label == "" || len(label) > maxLabelLen || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if !isHostChar(c) {
				return false
			}
		}
	}
	return true
}

func isHostChar(c rune) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-'
}

// LogValue keeps slog from printing the credentials.
func (s SMTP) LogValue() slog.Value {
	return slog.GroupValue(slog.String("Host", s.Host), slog.Int("Port", s.Port), slog.String("TLS", s.TLS),
		slog.String("Username", "[redacted]"), slog.String("Password", "[redacted]"), slog.String("From", "[redacted]"))
}

// GoString is the redacted form used by %#v.
func (s SMTP) GoString() string { return s.String() }

// String is the redacted form used by %v / %+v and logs.
func (s SMTP) String() string {
	return fmt.Sprintf("{Host:%s Port:%d Username:[redacted] Password:[redacted] From:[redacted] TLS:%s}", s.Host, s.Port, s.TLS)
}
