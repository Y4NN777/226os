package committers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c := New()
	c.base = srv.URL
	c.http = newHTTPClient(nil) // the dial guard refuses loopback; tested on its own below
	return c
}

func TestNewTargetsTheFixedHost(t *testing.T) {
	c := New()
	if c.base != "https://raw.githubusercontent.com" || c.http.Timeout != requestTimeout {
		t.Fatalf("client %+v", c)
	}
	if c.http.Transport.(*http.Transport).Proxy != nil {
		t.Fatal("no proxy")
	}
}

func TestFetchReadsTheLocationFile(t *testing.T) {
	var path string
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = w.Write([]byte("body"))
	})
	got, err := c.Fetch(context.Background(), "burkina_faso")
	if err != nil || string(got) != "body" {
		t.Fatalf("%q %v", got, err)
	}
	if path != "/ashkulz/committers.top/gh-pages/_data/locations/burkina_faso.yml" {
		t.Fatalf("path %s", path)
	}
}

func TestFetchRefusesBadSlugsStatusesRedirectsAndLargeBodies(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ashkulz/committers.top/gh-pages/_data/locations/missing.yml":
			http.NotFound(w, r)
		case "/ashkulz/committers.top/gh-pages/_data/locations/moved.yml":
			http.Redirect(w, r, "http://169.254.169.254/", http.StatusFound)
		case "/ashkulz/committers.top/gh-pages/_data/locations/short.yml":
			w.Header().Set("Content-Length", "100")
			_, _ = w.Write([]byte("cut")) // the body ends before its length: a read error
		default:
			_, _ = w.Write([]byte(strings.Repeat("x", maxBody+1)))
		}
	})
	for _, slug := range []string{"missing", "moved", "short", "big", "../etc", "Upper", ""} {
		if _, err := c.Fetch(context.Background(), slug); err == nil {
			t.Errorf("%q: want an error", slug)
		}
	}
	c.base = "http://[::1]:0" // nothing listens: a transport error
	if _, err := c.Fetch(context.Background(), "ghana"); err == nil {
		t.Fatal("unreachable host: want an error")
	}
	c.base = "http://bad host"
	if _, err := c.Fetch(context.Background(), "ghana"); err == nil {
		t.Fatal("bad base: want an error")
	}
}

func TestDialGuardRefusesInternalAddresses(t *testing.T) {
	for _, addr := range []string{
		"127.0.0.1:443", "10.0.0.1:443", "169.254.169.254:443", "[::1]:443", "[fe80::1%eth0]:443",
		"100.64.0.1:443", "192.0.2.1:443", "8.8.8.8:80", "nonsense", "host:443",
	} {
		if err := control("tcp", addr, nil); !errors.Is(err, errBlockedAddress) {
			t.Errorf("%s: %v", addr, err)
		}
	}
	if err := control("tcp", "185.199.108.133:443", nil); err != nil {
		t.Fatalf("public address: %v", err)
	}
	// The production client dials through the guard: a loopback server is refused.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	defer srv.Close()
	c := New()
	c.base = srv.URL
	if _, err := c.Fetch(context.Background(), "ghana"); err == nil || !errors.Is(err, errBlockedAddress) {
		t.Fatalf("loopback through the guard: %v", err)
	}
}
