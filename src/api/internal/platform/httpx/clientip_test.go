package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
)

func TestClientIPKey(t *testing.T) {
	cases := []struct {
		name       string
		xff        string
		remoteAddr string
		want       string
	}{
		{"no header uses the socket peer, reduced to /24", "", "203.0.113.77:51234", "203.0.113.0/24"},
		{"peer without a port is still parsed", "", "203.0.113.77", "203.0.113.0/24"},
		{"single forwarded entry", "198.51.100.9", "10.0.0.2:1", "198.51.100.0/24"},
		{"right-most entry wins: earlier ones are client supplied", "1.1.1.1, 2.2.2.2 ,198.51.100.9", "10.0.0.2:1", "198.51.100.0/24"},
		{"spoofed leading entry is ignored", "9.9.9.9, 198.51.100.200", "10.0.0.2:1", "198.51.100.0/24"},
		{"ipv6 reduced to /48", "2001:db8:abcd:1234:5678::1", "10.0.0.2:1", "2001:db8:abcd::/48"},
		{"ipv6 socket peer", "", "[2001:db8:abcd:1234::9]:443", "2001:db8:abcd::/48"},
		{"garbage collapses to one shared bucket", "not-an-ip", "10.0.0.2:1", "unknown"},
		{"empty right-most entry", "198.51.100.9,", "10.0.0.2:1", "unknown"},
		{"unparseable peer", "", "garbage", "unknown"},
		{"ipv4-mapped ipv6 is unmapped to its ipv4 /24", "::ffff:198.51.100.9", "10.0.0.2:1", "198.51.100.0/24"},
		{"ipv4-mapped ipv6 socket peer", "", "[::ffff:203.0.113.77]:443", "203.0.113.0/24"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
			r.RemoteAddr = tc.remoteAddr
			if tc.xff != "" {
				r.Header.Set("X-Forwarded-For", tc.xff)
			}
			if got := httpx.ClientIPKey(r); got != tc.want {
				t.Fatalf("ClientIPKey = %q, want %q", got, tc.want)
			}
		})
	}
}

// A second X-Forwarded-For header line is appended after the first; the right-most entry of the
// LAST line is what the proxy added, so a client-chosen first line must not win.
func TestClientIPKeyUsesTheLastHeaderLine(t *testing.T) {
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.2:1"
	r.Header.Add("X-Forwarded-For", "9.9.9.9")
	r.Header.Add("X-Forwarded-For", "198.51.100.9")
	if got := httpx.ClientIPKey(r); got != "198.51.100.0/24" {
		t.Fatalf("ClientIPKey = %q, want the last line's entry 198.51.100.0/24", got)
	}
}
