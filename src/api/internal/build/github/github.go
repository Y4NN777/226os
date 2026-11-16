// Package github is the read-only GitHub API adapter of Build (K10): one GET per project
// registration (authenticated only when a token is configured), with timeouts, a body cap, no redirects, no proxy, and a dial check that only
// lets public unicast addresses on port 443 through.
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
)

const (
	apiBase        = "https://api.github.com"
	requestTimeout = 5 * time.Second
	dialTimeout    = 3 * time.Second
	maxBody        = 1 << 20
	maxName        = 100 // projects_name_ck
	userAgent      = "226os-build"
)

var errBlockedAddress = errors.New("github: address not allowed")

// blocked are special-purpose ranges that netip still calls global unicast (RFC 6890 and successors).
var blocked = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("64:ff9b::/96"),
	netip.MustParsePrefix("2001::/32"),
	netip.MustParsePrefix("2001:db8::/32"),
	netip.MustParsePrefix("2002::/16"),
}

// Client reads public repositories.
type Client struct {
	http   *http.Client
	base   string
	token  string // optional secret; sent as a bearer token to base only
	logger *slog.Logger
}

// New returns the production client for https://api.github.com. A non-empty token is sent as a bearer
// token on the repository GET; it never goes to another host because base is fixed here.
// A nil logger discards logs.
func New(token string, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Client{http: newHTTPClient(control), base: apiBase, token: token, logger: logger}
}

// String and GoString keep the token out of %v, %+v, %#v and logs.
func (c Client) String() string {
	return "github.Client{base:" + c.base + " token:" + c.tokenState() + "}"
}

// GoString is the redacted form used by %#v.
func (c Client) GoString() string { return c.String() }

// LogValue keeps slog from printing the token.
func (c Client) LogValue() slog.Value {
	return slog.GroupValue(slog.String("base", c.base), slog.String("token", c.tokenState()))
}

func (c Client) tokenState() string {
	if c.token == "" {
		return "none"
	}
	return "[redacted]"
}

func newHTTPClient(check func(network, address string, c syscall.RawConn) error) *http.Client {
	dialer := &net.Dialer{Timeout: dialTimeout, Control: check}
	return &http.Client{
		Timeout: requestTimeout,
		Transport: &http.Transport{
			Proxy:                  nil,
			DialContext:            dialer.DialContext,
			ForceAttemptHTTP2:      true,
			TLSHandshakeTimeout:    dialTimeout,
			ResponseHeaderTimeout:  requestTimeout,
			MaxResponseHeaderBytes: 16 << 10,
			MaxIdleConns:           2,
			IdleConnTimeout:        30 * time.Second,
		},
		// A redirect is answered as is (a renamed repository is not followed to its new home).
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// control runs after DNS resolution on the address actually dialled, so a name that resolves (or
// re-resolves) to an internal address is refused too.
func control(_, address string, _ syscall.RawConn) error {
	host, port, err := net.SplitHostPort(address)
	if err != nil || port != "443" {
		return errBlockedAddress
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return errBlockedAddress
	}
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.Zone() != "" {
		return errBlockedAddress
	}
	for _, p := range blocked {
		if p.Contains(ip) {
			return errBlockedAddress
		}
	}
	return nil
}

type repoBody struct {
	Name        string  `json:"name"`
	FullName    string  `json:"full_name"`
	Description *string `json:"description"`
	Private     bool    `json:"private"`
}

// Repo returns the public repository owner/name. A missing, private, renamed or legally blocked
// repository is domain.ErrRepoNotFound; anything GitHub cannot answer now is domain.ErrUpstream.
func (c *Client) Repo(ctx context.Context, owner, name string) (domain.Repo, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	target := c.base + "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(name)
	resp, err := c.get(ctx, target, c.token)
	if err != nil {
		return domain.Repo{}, err
	}
	if resp.StatusCode == http.StatusUnauthorized && c.token != "" {
		// An expired or revoked token degrades to the unauthenticated quota instead of failing registration.
		_ = resp.Body.Close()
		c.logger.WarnContext(ctx, "github token rejected", "class", "github_token_rejected")
		if resp, err = c.get(ctx, target, ""); err != nil {
			return domain.Repo{}, err
		}
	}
	defer func() { _ = resp.Body.Close() }()
	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusMovedPermanently, http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect,
		http.StatusNotFound, http.StatusGone, http.StatusUnavailableForLegalReasons:
		return domain.Repo{}, domain.ErrRepoNotFound
	default:
		// 401 (after the unauthenticated retry), 403 and 429 (rate limit) and 5xx: upstream, never not-found.
		return domain.Repo{}, fmt.Errorf("%w: status %d", domain.ErrUpstream, resp.StatusCode)
	}
	var body repoBody
	dec := json.NewDecoder(io.LimitReader(resp.Body, maxBody))
	if err := dec.Decode(&body); err != nil {
		return domain.Repo{}, fmt.Errorf("%w: %w", domain.ErrUpstream, err)
	}
	if body.Private || body.Name == "" || utf8.RuneCountInString(body.Name) > maxName ||
		!strings.EqualFold(body.FullName, owner+"/"+name) {
		return domain.Repo{}, domain.ErrRepoNotFound
	}
	repo := domain.Repo{Name: clean(body.Name)}
	if body.Description != nil {
		repo.Description = clean(*body.Description)
	}
	return repo, nil
}

// get sends one GET, with a bearer token only when token is not empty.
func (c *Client) get(ctx context.Context, target, token string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrUpstream, err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", userAgent)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", domain.ErrUpstream, err)
	}
	return resp, nil
}

// clean drops what PostgreSQL text cannot hold: NUL and invalid UTF-8.
func clean(s string) string {
	return strings.ReplaceAll(strings.ToValidUTF8(s, "�"), "\x00", "")
}
