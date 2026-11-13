package httpx

import (
	"net"
	"net/http"
	"net/netip"
	"strings"
)

// ClientIPKey returns the rate-limit identity of the caller: the address reduced to its /24 (IPv4)
// or /48 (IPv6) prefix, or "unknown" when no address can be parsed.
//
// Only the right-most X-Forwarded-For entry is used. The API port is not published (compose.yaml
// has no ports: on api), so Traefik is the only ingress; it overwrites an untrusted client's
// X-Forwarded-For with the peer it saw, so the right-most entry is that peer and earlier entries
// are client supplied and spoofable. Without the header the socket peer is used. If the API is
// ever exposed directly, or placed behind another proxy, a trusted-proxy setting is required
// before this value can be trusted.
func ClientIPKey(r *http.Request) string {
	return prefixKey(clientAddr(r))
}

func clientAddr(r *http.Request) string {
	// The proxy appends to the LAST header line; earlier lines are client supplied.
	if lines := r.Header.Values("X-Forwarded-For"); len(lines) > 0 && lines[len(lines)-1] != "" {
		xff := lines[len(lines)-1]
		if i := strings.LastIndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[i+1:])
		}
		return strings.TrimSpace(xff)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func prefixKey(ip string) string {
	a, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		return "unknown"
	}
	a = a.Unmap() // ::ffff:a.b.c.d is an IPv4 client, not one shared ::/48 bucket
	if a.Is4() {
		b := a.As4()
		return netip.AddrFrom4([4]byte{b[0], b[1], b[2], 0}).String() + "/24"
	}
	// Everything ParseAddr returns that is not Is4 is IPv6 (BitLen 128), so Prefix(48) cannot fail.
	p, _ := a.Prefix(48)
	return p.String()
}
