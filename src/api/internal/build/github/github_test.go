package github

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
)

func allowAll(string, string, syscall.RawConn) error { return nil }

// testClient talks to srv with the production transport settings except the dial check, so status,
// redirect and body handling are tested without a public address.
func testClient(srv *httptest.Server) *Client {
	return &Client{http: newHTTPClient(allowAll), base: srv.URL}
}

func TestControlRefusesNonPublicAddresses(t *testing.T) {
	refused := []string{
		"127.0.0.1:443", "10.1.2.3:443", "172.16.0.1:443", "192.168.1.1:443", "169.254.169.254:443",
		"100.64.0.1:443", "0.1.2.3:443", "192.0.0.1:443", "192.0.2.1:443", "198.18.0.1:443",
		"198.51.100.1:443", "203.0.113.1:443", "240.0.0.1:443", "255.255.255.255:443", "224.0.0.1:443",
		"0.0.0.0:443", "[::1]:443", "[fe80::1]:443", "[fc00::1]:443", "[::ffff:127.0.0.1]:443",
		"[::ffff:10.0.0.1]:443", "[64:ff9b::a00:1]:443", "[2001:db8::1]:443", "[2002:a00:1::]:443",
		"[2001::1]:443", "[ff02::1]:443", "[::]:443",
		"140.82.112.6:80",    // public, but not https
		"140.82.112.6:8443",  // public, other port
		"github.com:443",     // never a name: control runs after DNS
		"140.82.112.6",       // no port
		"[fe80::1%eth0]:443", // zoned link-local
	}
	for _, addr := range refused {
		if err := control("tcp", addr, nil); !errors.Is(err, errBlockedAddress) {
			t.Errorf("%s allowed (%v)", addr, err)
		}
	}
	for _, addr := range []string{"140.82.112.6:443", "[2606:50c0:8000::153]:443", "[::ffff:140.82.112.6]:443"} {
		if err := control("tcp", addr, nil); err != nil {
			t.Errorf("%s refused: %v", addr, err)
		}
	}
}

func TestNewUsesTheHardenedTransport(t *testing.T) {
	c := New()
	if c.base != "https://api.github.com" || c.http.Timeout != requestTimeout || c.http.CheckRedirect == nil {
		t.Fatalf("client %+v", c)
	}
	tr := c.http.Transport.(*http.Transport)
	if tr.Proxy != nil || tr.TLSHandshakeTimeout == 0 || tr.ResponseHeaderTimeout == 0 || tr.MaxResponseHeaderBytes == 0 {
		t.Fatalf("transport %+v", tr)
	}
}

// The production dialer never reaches a loopback server: the request fails before any byte is sent.
func TestRepoRefusesALoopbackTarget(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits.Add(1) }))
	defer srv.Close()
	c := &Client{http: newHTTPClient(control), base: srv.URL}
	if _, err := c.Repo(t.Context(), "org", "repo"); !errors.Is(err, domain.ErrUpstream) || !errors.Is(err, errBlockedAddress) {
		t.Fatalf("err %v", err)
	}
	if hits.Load() != 0 {
		t.Fatal("loopback server was reached")
	}
}

func TestRepoNeverFollowsARedirect(t *testing.T) {
	var followed atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/internal" {
			followed.Add(1)
			return
		}
		http.Redirect(w, r, "/internal", http.StatusMovedPermanently)
	}))
	defer srv.Close()
	if _, err := testClient(srv).Repo(t.Context(), "org", "repo"); !errors.Is(err, domain.ErrRepoNotFound) {
		t.Fatalf("err %v", err)
	}
	if followed.Load() != 0 {
		t.Fatal("redirect followed")
	}
}

func TestRepoMapsResponses(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
		want   error
		repo   domain.Repo
	}{
		{"public repository", 200, `{"name":"Repo","full_name":"Org/Repo","description":"A tool","private":false}`, nil, domain.Repo{Name: "Repo", Description: "A tool"}},
		{"no description", 200, `{"name":"repo","full_name":"org/repo","description":null}`, nil, domain.Repo{Name: "repo"}},
		{"text is cleaned", 200, `{"name":"repo","full_name":"org/repo","description":"a\u0000b\udc00c"}`, nil, domain.Repo{Name: "repo", Description: "ab�c"}},
		{"private repository", 200, `{"name":"repo","full_name":"org/repo","private":true}`, domain.ErrRepoNotFound, domain.Repo{}},
		{"renamed repository", 200, `{"name":"other","full_name":"org/other"}`, domain.ErrRepoNotFound, domain.Repo{}},
		{"empty name", 200, `{"name":"","full_name":"org/repo"}`, domain.ErrRepoNotFound, domain.Repo{}},
		{"name too long", 200, `{"name":"` + strings.Repeat("n", 101) + `","full_name":"org/repo"}`, domain.ErrRepoNotFound, domain.Repo{}},
		{"not json", 200, `<html>`, domain.ErrUpstream, domain.Repo{}},
		{"body over the cap", 200, `{"name":"repo","full_name":"org/repo","description":"` + strings.Repeat("x", maxBody) + `"}`, domain.ErrUpstream, domain.Repo{}},
		{"missing", 404, `{}`, domain.ErrRepoNotFound, domain.Repo{}},
		{"gone", 410, `{}`, domain.ErrRepoNotFound, domain.Repo{}},
		{"legal block", 451, `{}`, domain.ErrRepoNotFound, domain.Repo{}},
		{"rate limited", 403, `{}`, domain.ErrUpstream, domain.Repo{}},
		{"too many requests", 429, `{}`, domain.ErrUpstream, domain.Repo{}},
		{"server error", 502, `{}`, domain.ErrUpstream, domain.Repo{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/repos/org/repo" || r.Header.Get("Accept") != "application/vnd.github+json" ||
					r.Header.Get("User-Agent") == "" || r.Header.Get("Authorization") != "" {
					t.Errorf("request %s %v", r.URL.Path, r.Header)
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()
			got, err := testClient(srv).Repo(t.Context(), "org", "repo")
			if !errors.Is(err, tc.want) || (tc.want == nil && err != nil) || got != tc.repo {
				t.Fatalf("got %+v, %v; want %+v, %v", got, err, tc.repo, tc.want)
			}
		})
	}
}

func TestRepoTimesOut(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
	defer srv.Close()
	defer close(release)
	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()
	if _, err := testClient(srv).Repo(ctx, "org", "repo"); !errors.Is(err, domain.ErrUpstream) {
		t.Fatalf("err %v", err)
	}
}

func TestRepoBadBaseURL(t *testing.T) {
	c := &Client{http: newHTTPClient(allowAll), base: "http://bad\x7f"}
	if _, err := c.Repo(t.Context(), "org", "repo"); !errors.Is(err, domain.ErrUpstream) {
		t.Fatalf("err %v", err)
	}
}
