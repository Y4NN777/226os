// Package smtptest is a minimal in-process ESMTP server for tests: it records every accepted
// message, can offer STARTTLS or implicit TLS with a certificate generated at test time, can
// require AUTH PLAIN, and can be scripted to fail, stall or drop the connection at a given command.
// It is a test helper: it binds 127.0.0.1 only and keeps credentials out of what it records.
package smtptest

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/big"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// Received is one message the server accepted (answered 250 after the final dot).
type Received struct {
	From string   // envelope sender
	To   []string // envelope recipients
	Data string   // DATA payload after dot-unstuffing, CRLF line endings
}

// AuthAttempt records an AUTH PLAIN exchange without its credentials.
type AuthAttempt struct {
	OverTLS bool // the connection was encrypted when AUTH arrived
	OK      bool // the credentials matched
}

type faultKind int

const (
	failKind faultKind = iota
	stallKind
	dropKind
	endlessKind
)

type fault struct {
	kind faultKind
	code int
}

type settings struct {
	starttls, implicit bool
	user, pass         string
	auth               bool
	faults             map[string]fault
}

// Option configures NewServer.
type Option func(*settings)

// WithSTARTTLS advertises STARTTLS on the plain connection.
func WithSTARTTLS() Option { return func(s *settings) { s.starttls = true } }

// WithImplicitTLS wraps every connection in TLS from the first byte.
func WithImplicitTLS() Option { return func(s *settings) { s.implicit = true } }

// WithAuth makes AUTH PLAIN available and accepts only this user and password.
func WithAuth(user, pass string) Option {
	return func(s *settings) { s.auth, s.user, s.pass = true, user, pass }
}

// WithFail answers the command cmd with the SMTP reply code instead of handling it.
// cmd is a verb (EHLO, STARTTLS, AUTH, MAIL, RCPT, DATA, RSET), GREETING, or END (after the final dot).
func WithFail(cmd string, code int) Option {
	return func(s *settings) { s.faults[cmd] = fault{kind: failKind, code: code} }
}

// WithStall makes the server stop answering at cmd until it is closed.
func WithStall(cmd string) Option {
	return func(s *settings) { s.faults[cmd] = fault{kind: stallKind} }
}

// WithDrop makes the server close the connection at cmd without a reply.
func WithDrop(cmd string) Option {
	return func(s *settings) { s.faults[cmd] = fault{kind: dropKind} }
}

// WithEndlessReply makes the server answer cmd with a never-ending multi-line reply
// ("220-xxxx..." lines, no final line) until the client or the server closes the connection.
func WithEndlessReply(cmd string) Option {
	return func(s *settings) { s.faults[cmd] = fault{kind: endlessKind} }
}

// Server is a running test SMTP server.
type Server struct {
	cfg     settings
	ln      net.Listener
	srvTLS  *tls.Config
	pool    *x509.CertPool
	done    chan struct{}
	wg      sync.WaitGroup
	mu      sync.Mutex
	conns   []net.Conn
	msgs    []Received
	auths   []AuthAttempt
	closing bool
	tb      testing.TB
	accepts int
}

// check fails the test on an error; it keeps test-helper plumbing free of unreachable branches.
func check(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Fatalf("smtptest: %v", err)
	}
}

// NewServer starts the server on 127.0.0.1:0 and registers its Close with t.Cleanup.
func NewServer(tb testing.TB, opts ...Option) *Server {
	tb.Helper()
	s := &Server{cfg: settings{faults: map[string]fault{}}, done: make(chan struct{}), tb: tb}
	for _, o := range opts {
		o(&s.cfg)
	}
	s.srvTLS, s.pool = newTLS(tb)
	var err error
	s.ln, err = (&net.ListenConfig{}).Listen(tb.Context(), "tcp", "127.0.0.1:0")
	check(tb, err)
	s.wg.Go(s.acceptLoop)
	tb.Cleanup(s.Close)
	return s
}

func newTLS(tb testing.TB) (*tls.Config, *x509.CertPool) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	check(tb, err)
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "smtptest"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	check(tb, err)
	leaf, err := x509.ParseCertificate(der)
	check(tb, err)
	pool := x509.NewCertPool()
	pool.AddCert(leaf)
	return &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: key}},
		MinVersion:   tls.VersionTLS12,
	}, pool
}

// Addr is the listen address, host:port.
func (s *Server) Addr() string { return s.ln.Addr().String() }

// Host is the listen IP.
func (s *Server) Host() string { return "127.0.0.1" }

// Port is the listen port.
func (s *Server) Port() int { return s.ln.Addr().(*net.TCPAddr).Port }

// CertPool trusts the server certificate.
func (s *Server) CertPool() *x509.CertPool { return s.pool }

// TLSConfig is a client configuration that trusts the server certificate.
func (s *Server) TLSConfig() *tls.Config {
	return &tls.Config{RootCAs: s.pool, MinVersion: tls.VersionTLS12}
}

// Messages returns a copy of the accepted messages, in arrival order.
func (s *Server) Messages() []Received {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Received, len(s.msgs))
	for i, m := range s.msgs {
		out[i] = Received{From: m.From, To: append([]string(nil), m.To...), Data: m.Data}
	}
	return out
}

// Accepted is the number of TCP connections the listener has accepted so far.
func (s *Server) Accepted() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.accepts
}

// Auths returns the AUTH PLAIN attempts, in arrival order.
func (s *Server) Auths() []AuthAttempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]AuthAttempt(nil), s.auths...)
}

// Close stops the listener, ends every session and waits for them. It is safe to call twice.
func (s *Server) Close() {
	s.mu.Lock()
	if !s.closing {
		s.closing = true
		close(s.done)
		_ = s.ln.Close()
		for _, c := range s.conns {
			_ = c.Close()
		}
	}
	s.mu.Unlock()
	s.wg.Wait()
}

func (s *Server) acceptLoop() {
	for {
		c, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		s.accepts++
		s.conns = append(s.conns, c)
		s.mu.Unlock()
		s.wg.Go(func() { s.session(c) })
	}
}

type session struct {
	srv  *Server
	conn net.Conn
	r    *bufio.Reader
	tls  bool
	from string
	to   []string
}

func (s *Server) session(c net.Conn) {
	defer func() { _ = c.Close() }()
	ss := &session{srv: s, conn: c}
	if s.cfg.implicit {
		ss.useTLS()
	} else {
		ss.r = bufio.NewReader(c)
	}
	if consumed, end := ss.intercept("GREETING"); consumed {
		if end {
			return
		}
	} else {
		ss.reply("220 smtptest ready")
	}
	for {
		line, err := ss.r.ReadString('\n')
		if err != nil {
			return
		}
		if ss.command(strings.TrimRight(line, "\r\n")) {
			return
		}
	}
}

func (ss *session) useTLS() {
	ss.conn = tls.Server(ss.conn, ss.srv.srvTLS)
	ss.r = bufio.NewReader(ss.conn)
	ss.tls = true
}

func (ss *session) reply(format string, args ...any) {
	_, _ = fmt.Fprintf(ss.conn, format+"\r\n", args...)
}

// intercept applies the scripted fault for cmd. consumed: the command must not be handled normally;
// end: the session is over.
func (ss *session) intercept(cmd string) (consumed, end bool) {
	f, ok := ss.srv.cfg.faults[cmd]
	if !ok {
		return false, false
	}
	switch f.kind {
	case failKind:
		ss.reply("%d scripted failure", f.code)
		return true, false
	case stallKind:
		<-ss.srv.done
		return true, true
	case endlessKind:
		line := []byte("220-" + strings.Repeat("x", 1020) + "\r\n")
		for {
			if _, err := ss.conn.Write(line); err != nil {
				return true, true
			}
		}
	default:
		return true, true
	}
}

// command handles one command line and reports whether the session ends.
func (ss *session) command(line string) bool {
	verb, arg, _ := strings.Cut(line, " ")
	verb = strings.ToUpper(verb)
	if consumed, end := ss.intercept(verb); consumed {
		return end
	}
	switch verb {
	case "EHLO", "HELO":
		ss.ehlo()
	case "STARTTLS":
		ss.reply("220 go ahead")
		ss.useTLS()
	case "AUTH":
		ss.auth(arg)
	case "MAIL":
		ss.from, ss.to = ss.path(verb, arg, "FROM:", true), nil
		ss.reply("250 ok")
	case "RCPT":
		ss.to = append(ss.to, ss.path(verb, arg, "TO:", false))
		ss.reply("250 ok")
	case "DATA":
		return ss.data()
	case "RSET":
		ss.from, ss.to = "", nil
		ss.reply("250 ok")
	case "QUIT":
		ss.reply("221 bye")
		return true
	default:
		ss.reply("502 command not implemented")
	}
	return false
}

func (ss *session) ehlo() {
	ss.reply("250-smtptest")
	if ss.srv.cfg.starttls && !ss.tls {
		ss.reply("250-STARTTLS")
	}
	if ss.srv.cfg.auth && (ss.tls || !ss.srv.cfg.starttls) {
		ss.reply("250-AUTH PLAIN")
	}
	ss.reply("250 8BITMIME")
}

// path extracts the address from "FROM:<a@b>" / "TO:<a@b>" and fails the test when the line is
// not exactly that: a missing bracket, a space or comma or extra '>' inside the path, or trailing
// text (MAIL may carry ESMTP parameters after one space; RCPT may not). A mismatch means the client
// wrote something other than the address it was given, so it must never be hidden.
func (ss *session) path(verb, arg, prefix string, params bool) string {
	inner, rest, ok := strings.Cut(strings.TrimPrefix(arg, prefix), ">")
	valid := ok && strings.HasPrefix(arg, prefix+"<") && !strings.ContainsAny(inner[1:], "< ,")
	valid = valid && (rest == "" || (params && strings.HasPrefix(rest, " ")))
	if !valid {
		ss.srv.tb.Errorf("smtptest: %s line %q is not exactly %s%s<address>", verb, arg, verb+" ", prefix)
		return arg
	}
	return inner[1:]
}

func (ss *session) auth(arg string) {
	ok := false
	if fields := strings.Fields(arg); len(fields) == 2 && strings.EqualFold(fields[0], "PLAIN") {
		if raw, err := base64.StdEncoding.DecodeString(fields[1]); err == nil {
			parts := strings.Split(string(raw), "\x00")
			ok = len(parts) == 3 && parts[1] == ss.srv.cfg.user && parts[2] == ss.srv.cfg.pass
		}
	}
	ss.srv.mu.Lock()
	ss.srv.auths = append(ss.srv.auths, AuthAttempt{OverTLS: ss.tls, OK: ok})
	ss.srv.mu.Unlock()
	if ok {
		ss.reply("235 authenticated")
		return
	}
	ss.reply("535 authentication failed")
}

// data reads the payload up to the lone dot, undoing dot-stuffing, and reports whether the session ends.
func (ss *session) data() bool {
	ss.reply("354 end with <CRLF>.<CRLF>")
	var b strings.Builder
	for {
		line, err := ss.r.ReadString('\n')
		if err != nil {
			return true
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "." {
			break
		}
		b.WriteString(strings.TrimPrefix(line, ".") + "\r\n")
	}
	if consumed, end := ss.intercept("END"); consumed {
		return end
	}
	ss.srv.mu.Lock()
	ss.srv.msgs = append(ss.srv.msgs, Received{From: ss.from, To: ss.to, Data: b.String()})
	ss.srv.mu.Unlock()
	ss.from, ss.to = "", nil
	ss.reply("250 queued")
	return false
}
