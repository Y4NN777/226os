package smtptest_test

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/mailer/smtptest"
)

var (
	user = fmt.Sprintf("user-%d", 11)
	pass = fmt.Sprintf("pw-%d-%s", 4711, "quokka")
)

type client struct {
	t    *testing.T
	conn net.Conn
	r    *bufio.Reader
}

func dial(t *testing.T, s *smtptest.Server) *client {
	t.Helper()
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(t.Context(), "tcp", s.Addr())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return &client{t: t, conn: conn, r: bufio.NewReader(conn)}
}

func (c *client) send(format string, args ...any) {
	c.t.Helper()
	_ = c.conn.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := fmt.Fprintf(c.conn, format+"\r\n", args...); err != nil {
		c.t.Fatal(err)
	}
}

// line reads one reply line.
func (c *client) line() string {
	c.t.Helper()
	_ = c.conn.SetDeadline(time.Now().Add(5 * time.Second))
	l, err := c.r.ReadString('\n')
	if err != nil {
		c.t.Fatal(err)
	}
	return strings.TrimRight(l, "\r\n")
}

// expect reads one line and checks its prefix.
func (c *client) expect(prefix string) string {
	c.t.Helper()
	l := c.line()
	if !strings.HasPrefix(l, prefix) {
		c.t.Fatalf("got %q, want prefix %q", l, prefix)
	}
	return l
}

// ehlo sends EHLO and returns the capability lines.
func (c *client) ehlo() []string {
	c.t.Helper()
	c.send("EHLO test")
	var caps []string
	for {
		l := c.line()
		caps = append(caps, l)
		if strings.HasPrefix(l, "250 ") {
			return caps
		}
	}
}

func has(lines []string, sub string) bool {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}

func plain(u, p string) string {
	return base64.StdEncoding.EncodeToString([]byte("\x00" + u + "\x00" + p))
}

func (c *client) upgrade(s *smtptest.Server) *client {
	c.t.Helper()
	cfg := s.TLSConfig()
	cfg.ServerName = "localhost"
	tc := tls.Client(c.conn, cfg)
	return &client{t: c.t, conn: tc, r: bufio.NewReader(tc)}
}

func TestPlainSessionRecordsMessagesAndUnstuffsDots(t *testing.T) {
	s := smtptest.NewServer(t)
	if s.Host() != "127.0.0.1" || s.Port() == 0 || !strings.HasSuffix(s.Addr(), fmt.Sprint(s.Port())) {
		t.Fatalf("host %s port %d addr %s", s.Host(), s.Port(), s.Addr())
	}
	c := dial(t, s)
	c.expect("220")
	if caps := c.ehlo(); has(caps, "STARTTLS") || has(caps, "AUTH") {
		t.Fatalf("plain server advertises %v", caps)
	}
	c.send("HELO x")
	c.expect("250-") // the same capability block
	c.expect("250 ")
	c.send("MAIL FROM:<a@example.org>")
	c.expect("250")
	c.send("RCPT TO:<b@example.org>")
	c.expect("250")
	c.send("RCPT TO:<c@example.org>")
	c.expect("250")
	c.send("DATA")
	c.expect("354")
	c.send("line one")
	c.send("..leading dot")
	c.send(".")
	c.expect("250")
	c.send("NOOP")
	c.expect("502")
	c.send("MAIL FROM:<x@example.org>")
	c.expect("250")
	c.send("RSET")
	c.expect("250")
	c.send("QUIT")
	c.expect("221")

	got := s.Messages()
	if len(got) != 1 || got[0].From != "a@example.org" || strings.Join(got[0].To, ",") != "b@example.org,c@example.org" ||
		got[0].Data != "line one\r\n.leading dot\r\n" {
		t.Fatalf("messages = %+v", got)
	}
	got[0].To[0] = "mutated"
	if s.Messages()[0].To[0] != "b@example.org" {
		t.Fatal("Messages returns shared state")
	}
}

func TestSTARTTLSThenAuth(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithSTARTTLS(), smtptest.WithAuth(user, pass))
	c := dial(t, s)
	c.expect("220")
	caps := c.ehlo()
	if !has(caps, "STARTTLS") || has(caps, "AUTH") {
		t.Fatalf("plain EHLO = %v", caps)
	}
	c.send("STARTTLS")
	c.expect("220")
	c = c.upgrade(s)
	caps = c.ehlo()
	if has(caps, "STARTTLS") || !has(caps, "AUTH PLAIN") {
		t.Fatalf("TLS EHLO = %v", caps)
	}
	c.send("AUTH PLAIN %s", plain(user, "wrong"))
	c.expect("535")
	c.send("AUTH PLAIN %s", plain(user, pass))
	c.expect("235")
	want := []smtptest.AuthAttempt{{OverTLS: true, OK: false}, {OverTLS: true, OK: true}}
	if got := s.Auths(); len(got) != 2 || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("auths = %+v", got)
	}
}

func TestAuthMalformedAndBeforeTLS(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithAuth(user, pass))
	c := dial(t, s)
	c.expect("220")
	if caps := c.ehlo(); !has(caps, "AUTH PLAIN") {
		t.Fatalf("EHLO = %v", caps)
	}
	for _, arg := range []string{"", "PLAIN", "LOGIN abc", "PLAIN !!!notbase64", "PLAIN " + base64.StdEncoding.EncodeToString([]byte("only\x00two"))} {
		c.send("AUTH %s", arg)
		c.expect("535")
	}
	c.send("AUTH PLAIN %s", plain(user, pass))
	c.expect("235")
	got := s.Auths()
	if len(got) != 6 || got[5].OverTLS || !got[5].OK {
		t.Fatalf("auths = %+v", got)
	}
	for _, a := range got[:5] {
		if a.OK {
			t.Fatalf("malformed AUTH accepted: %+v", got)
		}
	}
}

func TestImplicitTLS(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithImplicitTLS())
	c := dial(t, s).upgrade(s)
	c.expect("220")
	c.ehlo()
	c.send("QUIT")
	c.expect("221")
}

func TestScriptedFailures(t *testing.T) {
	t.Run("fail RCPT", func(t *testing.T) {
		s := smtptest.NewServer(t, smtptest.WithFail("RCPT", 550))
		c := dial(t, s)
		c.expect("220")
		c.send("RCPT TO:<a@example.org>")
		c.expect("550")
	})
	t.Run("fail DATA", func(t *testing.T) {
		s := smtptest.NewServer(t, smtptest.WithFail("DATA", 451))
		c := dial(t, s)
		c.expect("220")
		c.send("DATA")
		c.expect("451")
	})
	t.Run("fail END keeps nothing", func(t *testing.T) {
		s := smtptest.NewServer(t, smtptest.WithFail("END", 452))
		c := dial(t, s)
		c.expect("220")
		c.send("DATA")
		c.expect("354")
		c.send("x")
		c.send(".")
		c.expect("452")
		if n := len(s.Messages()); n != 0 {
			t.Fatalf("%d messages recorded", n)
		}
	})
	t.Run("fail GREETING then serve", func(t *testing.T) {
		s := smtptest.NewServer(t, smtptest.WithFail("GREETING", 421))
		c := dial(t, s)
		c.expect("421")
		c.send("QUIT")
		c.expect("221")
	})
	t.Run("drop GREETING", func(t *testing.T) {
		s := smtptest.NewServer(t, smtptest.WithDrop("GREETING"))
		c := dial(t, s)
		if _, err := c.r.ReadString('\n'); err == nil {
			t.Fatal("expected a closed connection")
		}
	})
	t.Run("drop MAIL", func(t *testing.T) {
		s := smtptest.NewServer(t, smtptest.WithDrop("MAIL"))
		c := dial(t, s)
		c.expect("220")
		c.send("MAIL FROM:<a@example.org>")
		if _, err := c.r.ReadString('\n'); err == nil {
			t.Fatal("expected a closed connection")
		}
	})
	t.Run("stall MAIL until Close", func(t *testing.T) {
		s := smtptest.NewServer(t, smtptest.WithStall("MAIL"))
		c := dial(t, s)
		c.expect("220")
		c.send("MAIL FROM:<a@example.org>")
		_ = c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if _, err := c.r.ReadString('\n'); err == nil {
			t.Fatal("stalled server answered")
		}
		s.Close()
		s.Close() // idempotent
	})
	t.Run("stall GREETING until Close", func(t *testing.T) {
		s := smtptest.NewServer(t, smtptest.WithStall("GREETING"))
		c := dial(t, s)
		_ = c.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if _, err := c.r.ReadString('\n'); err == nil {
			t.Fatal("stalled server answered")
		}
	})
}

func TestClientHangUpMidDataRecordsNothing(t *testing.T) {
	s := smtptest.NewServer(t)
	c := dial(t, s)
	c.expect("220")
	c.send("DATA")
	c.expect("354")
	c.send("partial")
	_ = c.conn.Close()
	s.Close() // waits for the session to notice
	if n := len(s.Messages()); n != 0 {
		t.Fatalf("%d messages recorded", n)
	}
}

func TestCertPoolTrustsTheServer(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithImplicitTLS())
	if s.CertPool() == nil {
		t.Fatal("no pool")
	}
	// Without the pool the handshake must fail: the certificate is self-signed.
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(t.Context(), "tcp", s.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
	if err := tls.Client(conn, &tls.Config{ServerName: "localhost", MinVersion: tls.VersionTLS12}).HandshakeContext(t.Context()); err == nil {
		t.Fatal("untrusted handshake succeeded")
	}
}

func TestAcceptedCountsConnections(t *testing.T) {
	s := smtptest.NewServer(t)
	if s.Accepted() != 0 {
		t.Fatalf("Accepted = %d before any connection", s.Accepted())
	}
	c := dial(t, s)
	c.expect("220")
	if s.Accepted() != 1 {
		t.Fatalf("Accepted = %d, want 1", s.Accepted())
	}
}

// recordingTB captures Errorf so a test can assert that the server reported a mismatch.
type recordingTB struct {
	testing.TB
	mu   sync.Mutex
	errs []string
}

func (r *recordingTB) Errorf(format string, args ...any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.errs = append(r.errs, fmt.Sprintf(format, args...))
}

func (r *recordingTB) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.errs)
}

func TestPathsAreCheckedStrictly(t *testing.T) {
	tests := []struct {
		line string
		bad  bool
	}{
		{"MAIL FROM:<a@example.org>", false},
		{"MAIL FROM:<a@example.org> BODY=8BITMIME", false},
		{"MAIL FROM:<a@example.org>BODY=8BITMIME", true},
		{"MAIL FROM:a@example.org", true},
		{"RCPT TO:<a@example.org>", false},
		{"RCPT TO:<a@example.org> NOTIFY=NEVER", true},
		{"RCPT TO:<a@example.org> BODY=8BITMIME", true},
		{"RCPT TO:<a@example.org>,<b@example.org>", true},
		{"RCPT TO:<a b@example.org>", true},
		{"RCPT TO:<a,b@example.org>", true},
		{"RCPT TO:<a<b@example.org>", true},
		{"RCPT TO:a@example.org", true},
		{"RCPT TO:<a@example.org", true},
		{"RCPT FOR:<a@example.org>", true},
	}
	for _, tc := range tests {
		t.Run(tc.line, func(t *testing.T) {
			rec := &recordingTB{TB: t}
			s := smtptest.NewServer(rec)
			c := dial(t, s)
			c.expect("220")
			c.send("%s", tc.line)
			c.expect("250")
			c.send("QUIT")
			c.expect("221")
			s.Close()
			if got := rec.count(); (got > 0) != tc.bad {
				t.Fatalf("%q: %d errors reported, want bad=%v", tc.line, got, tc.bad)
			}
		})
	}
}

func TestEndlessReplyNeverEnds(t *testing.T) {
	for _, cmd := range []string{"GREETING", "EHLO"} {
		t.Run(cmd, func(t *testing.T) {
			s := smtptest.NewServer(t, smtptest.WithEndlessReply(cmd))
			c := dial(t, s)
			if cmd != "GREETING" {
				c.expect("220")
				c.send("EHLO x")
			}
			for range 200 {
				if l := c.expect("220-"); len(l) < 1000 {
					t.Fatalf("short line %d", len(l))
				}
			}
			_ = c.conn.Close()
			s.Close() // the writer notices the closed connection and the session ends
		})
	}
}
