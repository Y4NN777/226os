package committers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"regexp"
	"syscall"
	"time"
)

// Same limits and guard as the Build GitHub client (build/github), on a fixed host.
const (
	rawBase        = "https://raw.githubusercontent.com"
	locationsPath  = "/ashkulz/committers.top/gh-pages/_data/locations/"
	requestTimeout = 5 * time.Second
	dialTimeout    = 3 * time.Second
	maxBody        = 2 << 20
	userAgent      = "226os-connect"
)

var (
	errBlockedAddress = errors.New("committers: address not allowed")
	slugPattern       = regexp.MustCompile(`^[a-z_]{2,40}$`)
)

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

// Client reads committers.top location files from raw.githubusercontent.com only.
type Client struct {
	http *http.Client
	base string
}

// New returns the production client: fixed host, 5 s timeout, 2 MiB cap, no redirect, no proxy, and a dial
// check that lets only public unicast addresses on port 443 through.
func New() *Client {
	return &Client{http: newHTTPClient(control), base: rawBase}
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
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
}

// control runs after DNS resolution on the address actually dialled, so a name that resolves to an
// internal address is refused too.
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

// Fetch returns the location file of slug (for example burkina_faso). Anything but a 200 with a body of at
// most 2 MiB is an error.
func (c *Client) Fetch(ctx context.Context, slug string) ([]byte, error) {
	if !slugPattern.MatchString(slug) {
		return nil, fmt.Errorf("committers: bad slug %q", slug)
	}
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+locationsPath+slug+".yml", nil)
	if err != nil {
		return nil, fmt.Errorf("committers: request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("committers: fetch: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("committers: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, fmt.Errorf("committers: read: %w", err)
	}
	if len(body) > maxBody {
		return nil, errors.New("committers: file larger than 2 MiB")
	}
	return body, nil
}
