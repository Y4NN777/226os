package mailer_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer/smtptest"
)

var (
	user = fmt.Sprintf("sender-%d@example.org", 5)
	pass = fmt.Sprintf("pw-%d-%s", 9021, "lynx")
)

func cfgFor(s *smtptest.Server, mode mailer.Mode, creds bool) mailer.SMTPConfig {
	c := mailer.SMTPConfig{Host: s.Host(), Port: s.Port(), From: "no-reply@example.org", TLS: mode}
	if creds {
		c.Username, c.Password = user, pass
	}
	return c
}

func send(t *testing.T, s *smtptest.Server, mode mailer.Mode, creds bool, m mailer.Message, opts ...mailer.Option) error {
	t.Helper()
	opts = append([]mailer.Option{mailer.WithTLSConfig(s.TLSConfig())}, opts...)
	return mailer.NewSMTP(cfgFor(s, mode, creds), opts...).Send(t.Context(), m)
}

var okMsg = mailer.Message{To: []string{"bob@example.net", "carol@example.net"}, Subject: "Hello", Text: "Hi there\n"}

func TestSendPlainBuildsAndDeliversTheMessage(t *testing.T) {
	s := smtptest.NewServer(t)
	now := time.Date(2026, 11, 12, 10, 30, 0, 0, time.UTC)
	text := "Line one\n.\n.dotted\nlast é line without newline"
	m := mailer.Message{To: okMsg.To, Subject: "Café 226os ☕", Text: text}
	if err := send(t, s, mailer.ModeNone, false, m,
		mailer.WithClock(func() time.Time { return now }), mailer.WithHostname("api.example.org")); err != nil {
		t.Fatal(err)
	}
	got := s.Messages()
	if len(got) != 1 || got[0].From != "no-reply@example.org" || strings.Join(got[0].To, ",") != "bob@example.net,carol@example.net" {
		t.Fatalf("envelope = %+v", got)
	}
	msg, err := mail.ReadMessage(strings.NewReader(got[0].Data))
	if err != nil {
		t.Fatal(err)
	}
	h := msg.Header
	if from, _ := h.AddressList("From"); len(from) != 1 || from[0].Address != "no-reply@example.org" || from[0].Name != "" {
		t.Fatalf("From = %v", h.Get("From"))
	}
	if to, _ := h.AddressList("To"); len(to) != 2 || to[0].Address != "bob@example.net" {
		t.Fatalf("To = %v", h.Get("To"))
	}
	if subj, err := new(mime.WordDecoder).DecodeHeader(h.Get("Subject")); err != nil || subj != m.Subject {
		t.Fatalf("Subject = %q (%v)", h.Get("Subject"), err)
	}
	for k, want := range map[string]string{
		"Mime-Version":              "1.0",
		"Content-Type":              "text/plain; charset=utf-8",
		"Content-Transfer-Encoding": "quoted-printable",
	} {
		if h.Get(k) != want {
			t.Fatalf("%s = %q, want %q", k, h.Get(k), want)
		}
	}
	if d, err := h.Date(); err != nil || !d.Equal(now) {
		t.Fatalf("Date = %q (%v)", h.Get("Date"), err)
	}
	if !regexp.MustCompile(`^<[0-9a-f]{32}@api\.example\.org>$`).MatchString(h.Get("Message-Id")) {
		t.Fatalf("Message-ID = %q", h.Get("Message-Id"))
	}
	body, err := io.ReadAll(quotedprintable.NewReader(msg.Body))
	if err != nil {
		t.Fatal(err)
	}
	// net/smtp's Data writer terminates the last line, so one trailing CRLF is added.
	if want := strings.ReplaceAll(text, "\n", "\r\n") + "\r\n"; string(body) != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
	if strings.Contains(got[0].Data, "\n") && strings.Contains(strings.ReplaceAll(got[0].Data, "\r\n", ""), "\n") {
		t.Fatal("bare LF in the message")
	}
}

func TestMessageIDsAreUniqueAndDefaultToTheFromDomain(t *testing.T) {
	s := smtptest.NewServer(t)
	for range 2 {
		if err := send(t, s, mailer.ModeNone, false, okMsg); err != nil {
			t.Fatal(err)
		}
	}
	ids := map[string]bool{}
	for _, r := range s.Messages() {
		msg, err := mail.ReadMessage(strings.NewReader(r.Data))
		if err != nil {
			t.Fatal(err)
		}
		id := msg.Header.Get("Message-Id")
		if !strings.HasSuffix(id, "@example.org>") {
			t.Fatalf("Message-ID %q does not use the From domain", id)
		}
		ids[id] = true
	}
	if len(ids) != 2 {
		t.Fatalf("Message-IDs repeat: %v", ids)
	}
}

func TestLongNonASCIISubjectStaysWithinLineLimits(t *testing.T) {
	s := smtptest.NewServer(t)
	subject := strings.Repeat("é", 200)
	if err := send(t, s, mailer.ModeNone, false, mailer.Message{To: okMsg.To, Subject: subject, Text: "x"}); err != nil {
		t.Fatal(err)
	}
	data := s.Messages()[0].Data
	header, _, _ := strings.Cut(data, "\r\n\r\n")
	for _, l := range strings.Split(header, "\r\n") {
		if len(l) > 78 {
			t.Fatalf("header line of %d bytes: %q", len(l), l)
		}
	}
	msg, err := mail.ReadMessage(strings.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if got, err := new(mime.WordDecoder).DecodeHeader(msg.Header.Get("Subject")); err != nil || got != subject {
		t.Fatalf("subject round trip failed: %v", err)
	}
}

func TestSendRandFailure(t *testing.T) {
	s := smtptest.NewServer(t)
	err := send(t, s, mailer.ModeNone, false, okMsg, mailer.WithRand(failingReader{}))
	if err == nil || errors.Is(err, mailer.ErrInvalidMessage) || len(s.Messages()) != 0 {
		t.Fatalf("err = %v, messages = %d", err, len(s.Messages()))
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("entropy exhausted") }

func TestSendTransportModes(t *testing.T) {
	tests := []struct {
		name    string
		mode    mailer.Mode
		server  []smtptest.Option
		creds   bool
		tlsAuth bool
	}{
		{"none without credentials", mailer.ModeNone, nil, false, false},
		{"starttls without credentials", mailer.ModeSTARTTLS, []smtptest.Option{smtptest.WithSTARTTLS()}, false, false},
		{"starttls with credentials", mailer.ModeSTARTTLS, []smtptest.Option{smtptest.WithSTARTTLS(), smtptest.WithAuth(user, pass)}, true, true},
		{"implicit tls with credentials", mailer.ModeTLS, []smtptest.Option{smtptest.WithImplicitTLS(), smtptest.WithAuth(user, pass)}, true, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := smtptest.NewServer(t, tc.server...)
			if err := send(t, s, tc.mode, tc.creds, okMsg); err != nil {
				t.Fatal(err)
			}
			if len(s.Messages()) != 1 {
				t.Fatalf("messages = %d", len(s.Messages()))
			}
			auths := s.Auths()
			if tc.creds != (len(auths) == 1) || (tc.creds && (!auths[0].OverTLS || !auths[0].OK)) {
				t.Fatalf("auths = %+v", auths)
			}
		})
	}
}

func TestStartTLSIsRequiredNoSilentDowngrade(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithAuth(user, pass)) // does not advertise STARTTLS
	err := send(t, s, mailer.ModeSTARTTLS, true, okMsg)
	if !errors.Is(err, mailer.ErrPermanent) {
		t.Fatalf("err = %v, want ErrPermanent", err)
	}
	if len(s.Auths()) != 0 || len(s.Messages()) != 0 {
		t.Fatalf("data went out in clear: auths %v messages %d", s.Auths(), len(s.Messages()))
	}
}

func TestCredentialsRefusedWithoutTLS(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithAuth(user, pass))
	var dials atomic.Int32
	err := send(t, s, mailer.ModeNone, true, okMsg, countingDialer(&dials))
	if !errors.Is(err, mailer.ErrNotConfigured) || dials.Load() != 0 || len(s.Auths()) != 0 {
		t.Fatalf("err = %v, dials = %d, auths = %v", err, dials.Load(), s.Auths())
	}
}

func countingDialer(n *atomic.Int32) mailer.Option {
	return mailer.WithDialer(func(ctx context.Context, network, addr string) (net.Conn, error) {
		n.Add(1)
		return (&net.Dialer{}).DialContext(ctx, network, addr)
	})
}

func TestInvalidConfigIsRefusedBeforeAnyNetworkIO(t *testing.T) {
	good := mailer.SMTPConfig{Host: "127.0.0.1", Port: 25, From: "a@example.org", TLS: mailer.ModeNone}
	tests := map[string]func(*mailer.SMTPConfig){
		"no host":      func(c *mailer.SMTPConfig) { c.Host = "" },
		"bad port":     func(c *mailer.SMTPConfig) { c.Port = 0 },
		"port too big": func(c *mailer.SMTPConfig) { c.Port = 70000 },
		"bad from":     func(c *mailer.SMTPConfig) { c.From = "nope" },
		"unknown mode": func(c *mailer.SMTPConfig) { c.TLS = "ssl" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			var dials atomic.Int32
			c := good
			mutate(&c)
			err := mailer.NewSMTP(c, countingDialer(&dials)).Send(t.Context(), okMsg)
			if !errors.Is(err, mailer.ErrNotConfigured) || dials.Load() != 0 {
				t.Fatalf("err = %v, dials = %d", err, dials.Load())
			}
		})
	}
}

func TestInvalidMessagesNeverOpenAConnection(t *testing.T) {
	many := make([]string, 51)
	for i := range many {
		many[i] = fmt.Sprintf("u%d@example.net", i)
	}
	tests := map[string]mailer.Message{
		"no recipient":         {Subject: "s", Text: "t"},
		"51 recipients":        {To: many, Subject: "s", Text: "t"},
		"bad recipient":        {To: []string{"not an address"}, Subject: "s", Text: "t"},
		"recipient list":       {To: []string{"a@example.net, b@example.net"}, Subject: "s", Text: "t"},
		"recipient LF":         {To: []string{"a@example.net\nBcc: x@example.net"}, Subject: "s", Text: "t"},
		"recipient CR":         {To: []string{"a@example.net\r"}, Subject: "s", Text: "t"},
		"recipient NUL":        {To: []string{"a@example.net\x00"}, Subject: "s", Text: "t"},
		"quoted recipient LF":  {To: []string{"\"x\ny\" <a@example.net>"}, Subject: "s", Text: "t"},
		"display name":         {To: []string{"Bob <bob@example.net>"}, Subject: "s", Text: "t"},
		"empty subject":        {To: okMsg.To, Text: "t"},
		"subject too long":     {To: okMsg.To, Subject: strings.Repeat("s", 201), Text: "t"},
		"subject LF injection": {To: okMsg.To, Subject: "hi\nBcc: evil@example.net", Text: "t"},
		"subject CRLF":         {To: okMsg.To, Subject: "hi\r\nBcc: evil@example.net", Text: "t"},
		"subject NUL":          {To: okMsg.To, Subject: "hi\x00", Text: "t"},
		"subject BEL":          {To: okMsg.To, Subject: "hi\a", Text: "t"},
		"subject ESC":          {To: okMsg.To, Subject: "hi\x1b[2J", Text: "t"},
		"subject US":           {To: okMsg.To, Subject: "hi\x1f", Text: "t"},
		"subject DEL":          {To: okMsg.To, Subject: "hi\x7f", Text: "t"},
		"subject bad utf8":     {To: okMsg.To, Subject: "hi\xff", Text: "t"},
		"empty text":           {To: okMsg.To, Subject: "s"},
		"text too large":       {To: okMsg.To, Subject: "s", Text: strings.Repeat("x", 262145)},
		"text bad utf8":        {To: okMsg.To, Subject: "s", Text: "x\xff"},
	}
	for name, m := range tests {
		t.Run(name, func(t *testing.T) {
			s := smtptest.NewServer(t)
			var dials atomic.Int32
			err := send(t, s, mailer.ModeNone, false, m, countingDialer(&dials))
			if !errors.Is(err, mailer.ErrInvalidMessage) || dials.Load() != 0 || len(s.Messages()) != 0 {
				t.Fatalf("err = %v, dials = %d", err, dials.Load())
			}
			if wrapped := fmt.Errorf("outer: %w", err); !errors.Is(wrapped, mailer.ErrInvalidMessage) {
				t.Fatal("errors.Is does not see through wrapping")
			}
			if strings.Contains(err.Error(), "example.net") {
				t.Fatalf("error echoes recipient data: %v", err)
			}
		})
	}
}

func TestBoundaryMessagesAreAccepted(t *testing.T) {
	s := smtptest.NewServer(t)
	many := make([]string, 50)
	for i := range many {
		many[i] = fmt.Sprintf("u%d@example.net", i)
	}
	m := mailer.Message{To: many, Subject: strings.Repeat("s", 200), Text: strings.Repeat("x", 262144)}
	if err := send(t, s, mailer.ModeNone, false, m); err != nil {
		t.Fatal(err)
	}
	if got := s.Messages(); len(got) != 1 || len(got[0].To) != 50 {
		t.Fatalf("messages = %d", len(got))
	}
}

func TestSendClassifiesFailures(t *testing.T) {
	tests := []struct {
		name   string
		mode   mailer.Mode
		creds  bool
		server []smtptest.Option
		want   error
		code   string
	}{
		{"greeting 421", mailer.ModeNone, false, []smtptest.Option{smtptest.WithFail("GREETING", 421)}, mailer.ErrTemporary, "421"},
		{"ehlo and helo 502", mailer.ModeNone, false, []smtptest.Option{smtptest.WithFail("EHLO", 502), smtptest.WithFail("HELO", 502)}, mailer.ErrPermanent, "502"},
		{"starttls 454", mailer.ModeSTARTTLS, false, []smtptest.Option{smtptest.WithSTARTTLS(), smtptest.WithFail("STARTTLS", 454)}, mailer.ErrTemporary, "454"},
		{"mail 550", mailer.ModeNone, false, []smtptest.Option{smtptest.WithFail("MAIL", 550)}, mailer.ErrPermanent, "550"},
		{"rcpt 550", mailer.ModeNone, false, []smtptest.Option{smtptest.WithFail("RCPT", 550)}, mailer.ErrPermanent, "550"},
		{"rcpt 450", mailer.ModeNone, false, []smtptest.Option{smtptest.WithFail("RCPT", 450)}, mailer.ErrTemporary, "450"},
		{"data 451", mailer.ModeNone, false, []smtptest.Option{smtptest.WithFail("DATA", 451)}, mailer.ErrTemporary, "451"},
		{"end 452", mailer.ModeNone, false, []smtptest.Option{smtptest.WithFail("END", 452)}, mailer.ErrTemporary, "452"},
		{"end 554", mailer.ModeNone, false, []smtptest.Option{smtptest.WithFail("END", 554)}, mailer.ErrPermanent, "554"},
		{"auth rejected", mailer.ModeTLS, true, []smtptest.Option{smtptest.WithImplicitTLS(), smtptest.WithAuth(user, "other")}, mailer.ErrPermanent, "535"},
		{"auth not offered", mailer.ModeTLS, true, []smtptest.Option{smtptest.WithImplicitTLS()}, mailer.ErrPermanent, "auth"},
		{"connection dropped", mailer.ModeNone, false, []smtptest.Option{smtptest.WithDrop("RCPT")}, mailer.ErrTemporary, "rcpt"},
		{"dropped at greeting", mailer.ModeNone, false, []smtptest.Option{smtptest.WithDrop("GREETING")}, mailer.ErrTemporary, "greeting"},
		{"plain server for implicit tls", mailer.ModeTLS, false, nil, mailer.ErrTemporary, "tls"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := smtptest.NewServer(t, tc.server...)
			err := send(t, s, tc.mode, tc.creds, okMsg)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
			other := mailer.ErrPermanent
			if errors.Is(tc.want, mailer.ErrPermanent) {
				other = mailer.ErrTemporary
			}
			if errors.Is(err, other) {
				t.Fatalf("err %v matches both classes", err)
			}
			if !strings.Contains(strings.ToLower(err.Error()), tc.code) {
				t.Fatalf("err %q lacks %q", err, tc.code)
			}
			for _, secret := range []string{pass, user, "other"} {
				if strings.Contains(err.Error(), secret) {
					t.Fatalf("error leaks a credential: %v", err)
				}
			}
			if len(s.Messages()) != 0 {
				t.Fatal("a failed send delivered a message")
			}
		})
	}
}

func TestBodyWriteFailureIsTemporary(t *testing.T) {
	s := smtptest.NewServer(t)
	dial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		c, err := (&net.Dialer{}).DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		return &partialConn{Conn: c}, nil
	}
	m := mailer.Message{To: okMsg.To, Subject: "s", Text: strings.Repeat("x", 20000)}
	err := send(t, s, mailer.ModeNone, false, m, mailer.WithDialer(dial))
	if !errors.Is(err, mailer.ErrTemporary) || len(s.Messages()) != 0 {
		t.Fatalf("err = %v", err)
	}
}

// partialConn passes small writes through and fails large ones.
type partialConn struct{ net.Conn }

func (p *partialConn) Write(b []byte) (int, error) {
	if len(b) > 1000 {
		return 0, errors.New("broken pipe")
	}
	return p.Conn.Write(b)
}

func TestUntrustedCertificateIsPermanent(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithSTARTTLS())
	err := mailer.NewSMTP(cfgFor(s, mailer.ModeSTARTTLS, false)).Send(t.Context(), okMsg) // no trusted pool
	if !errors.Is(err, mailer.ErrPermanent) || len(s.Messages()) != 0 {
		t.Fatalf("err = %v", err)
	}
}

func TestDialFailureAndCustomDialer(t *testing.T) {
	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	err = mailer.NewSMTP(mailer.SMTPConfig{Host: "127.0.0.1", Port: port, From: "a@example.org", TLS: mailer.ModeNone}).Send(t.Context(), okMsg)
	if !errors.Is(err, mailer.ErrTemporary) {
		t.Fatalf("err = %v", err)
	}
	dial := func(context.Context, string, string) (net.Conn, error) { return nil, errors.New("no route") }
	err = mailer.NewSMTP(mailer.SMTPConfig{Host: "mail.example.org", Port: 25, From: "a@example.org", TLS: mailer.ModeNone}, mailer.WithDialer(dial)).Send(t.Context(), okMsg)
	if !errors.Is(err, mailer.ErrTemporary) {
		t.Fatalf("err = %v", err)
	}
}

func TestSendTimesOutOnAStalledServer(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithStall("MAIL"))
	start := time.Now()
	err := send(t, s, mailer.ModeNone, false, okMsg, mailer.WithTimeout(150*time.Millisecond))
	if !errors.Is(err, mailer.ErrTemporary) || time.Since(start) > 5*time.Second {
		t.Fatalf("err = %v after %v", err, time.Since(start))
	}
}

func TestSendStallAtGreetingIsBoundedToo(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithStall("GREETING"))
	err := send(t, s, mailer.ModeNone, false, okMsg, mailer.WithTimeout(150*time.Millisecond))
	if !errors.Is(err, mailer.ErrTemporary) {
		t.Fatalf("err = %v", err)
	}
}

func TestSendAbortsWhenTheContextIsCancelled(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithStall("MAIL"))
	ctx, cancel := context.WithCancel(t.Context())
	time.AfterFunc(100*time.Millisecond, cancel)
	start := time.Now()
	err := mailer.NewSMTP(cfgFor(s, mailer.ModeNone, false)).Send(ctx, okMsg)
	if !errors.Is(err, mailer.ErrTemporary) || !errors.Is(err, context.Canceled) || time.Since(start) > 5*time.Second {
		t.Fatalf("err = %v after %v", err, time.Since(start))
	}
}

func TestDisabled(t *testing.T) {
	var m mailer.Mailer = mailer.Disabled{}
	if err := m.Send(t.Context(), okMsg); !errors.Is(err, mailer.ErrNotConfigured) {
		t.Fatalf("err = %v", err)
	}
}

func TestFakeRecordsCopiesAndValidates(t *testing.T) {
	f := mailer.NewFake()
	var m mailer.Mailer = f
	msg := mailer.Message{To: []string{"a@example.net"}, Subject: "s", Text: "t"}
	if err := m.Send(t.Context(), msg); err != nil {
		t.Fatal(err)
	}
	msg.To[0] = "mutated@example.net" // the caller reuses its slice
	got := f.Sent()
	if len(got) != 1 || got[0].To[0] != "a@example.net" {
		t.Fatalf("sent = %+v", got)
	}
	got[0].To[0] = "changed"
	if f.Sent()[0].To[0] != "a@example.net" {
		t.Fatal("Sent returns shared state")
	}
	if err := m.Send(t.Context(), mailer.Message{To: []string{"a@example.net"}, Subject: "x\ny", Text: "t"}); !errors.Is(err, mailer.ErrInvalidMessage) || len(f.Sent()) != 1 {
		t.Fatalf("invalid message: err = %v, sent = %d", err, len(f.Sent()))
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if err := m.Send(ctx, msg); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context: err = %v", err)
	}
}

func TestFakeFailWith(t *testing.T) {
	f := mailer.NewFake()
	msg := mailer.Message{To: []string{"a@example.net"}, Subject: "s", Text: "t"}
	f.FailWith(mailer.ErrTemporary)
	for range 2 {
		if err := f.Send(t.Context(), msg); !errors.Is(err, mailer.ErrTemporary) {
			t.Fatalf("err = %v", err)
		}
	}
	if len(f.Sent()) != 0 {
		t.Fatal("a failed send was recorded")
	}
	f.FailWith(nil)
	if err := f.Send(t.Context(), msg); err != nil || len(f.Sent()) != 1 {
		t.Fatalf("err = %v, sent = %d", err, len(f.Sent()))
	}
}

func TestFakeIsGoroutineSafe(t *testing.T) {
	f := mailer.NewFake()
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			_ = f.Send(context.Background(), mailer.Message{To: []string{"a@example.net"}, Subject: "s", Text: "t"})
			_ = f.Sent()
			f.FailWith(nil)
		})
	}
	wg.Wait()
	if n := len(f.Sent()); n != 20 {
		t.Fatalf("sent = %d", n)
	}
}

// rawRCPT returns the verbatim argument of each RCPT line the server received, via a recording dialer.
type lineRecorder struct {
	mu    sync.Mutex
	lines []string
}

func (l *lineRecorder) dialer() mailer.Option {
	return mailer.WithDialer(func(ctx context.Context, network, addr string) (net.Conn, error) {
		c, err := (&net.Dialer{}).DialContext(ctx, network, addr)
		if err != nil {
			return nil, err
		}
		return &recordingConn{Conn: c, rec: l}, nil
	})
}

type recordingConn struct {
	net.Conn
	rec *lineRecorder
}

func (c *recordingConn) Write(b []byte) (int, error) {
	c.rec.mu.Lock()
	c.rec.lines = append(c.rec.lines, string(b))
	c.rec.mu.Unlock()
	return c.Conn.Write(b)
}

func (l *lineRecorder) commands(prefix string) []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []string
	for _, w := range l.lines {
		if strings.HasPrefix(w, prefix) {
			out = append(out, strings.TrimRight(w, "\r\n"))
		}
	}
	return out
}

func TestAddressesAreBareASCIIDotAtomOnly(t *testing.T) {
	long := strings.Repeat("a", 65) + "@example.org"
	bad := map[string]string{
		"quoted injection 1":    `"victim@good.com>,<attacker@evil.com"@x.com`,
		"quoted injection 2":    `"a> BODY=8BITMIME"@x.com`,
		"quoted injection 3":    `"x@ok.com> NOTIFY=NEVER ORCPT=rfc822;z@y.com"@x.com`,
		"space in local":        "a b@x.com",
		"NEL after domain":      "a@b.com\u0085",
		"LS after domain":       "a@b.com ",
		"embedded LF":           "a\nb@x.com",
		"non-ASCII local":       "é@x.com",
		"non-ASCII domain":      "a@é.com",
		"angle brackets":        "<a@b.com>",
		"display name":          "Name <a@b.com>",
		"IP literal":            "a@[127.0.0.1]",
		"double dot":            "a..b@x.com",
		"leading dot":           ".a@x.com",
		"trailing dot":          "a.@x.com",
		"no domain":             "a@",
		"no local":              "@x.com",
		"no at":                 "ax.com",
		"two ats":               "a@b@x.com",
		"surrounding space":     " a@x.com",
		"trailing space":        "a@x.com ",
		"comment":               "a@x.com (hi)",
		"comma list":            "a@x.com,b@x.com",
		"domain leading hyphen": "a@-x.com",
		"domain trailing dash":  "a@x-.com",
		"domain empty label":    "a@x..com",
		"domain underscore":     "a@x_y.com",
		"local too long":        long,
		"label too long":        "a@" + strings.Repeat("b", 64) + ".com",
		"domain too long":       "a@" + strings.Repeat(strings.Repeat("b", 60)+".", 5) + "com",
		"address too long":      strings.Repeat("a", 64) + "@" + strings.Repeat(strings.Repeat("b", 60)+".", 4) + "com",
		"DEL":                   "a\x7f@x.com",
		"tab":                   "a\t@x.com",
	}
	for name, addr := range bad {
		t.Run(name, func(t *testing.T) {
			s := smtptest.NewServer(t)
			err := send(t, s, mailer.ModeNone, false, mailer.Message{To: []string{addr}, Subject: "s", Text: "t"})
			if !errors.Is(err, mailer.ErrInvalidMessage) || s.Accepted() != 0 {
				t.Fatalf("err = %v, accepted connections = %d", err, s.Accepted())
			}
			if vErr := mailer.ValidateAddress(addr); vErr == nil {
				t.Fatal("ValidateAddress accepts it")
			}
		})
	}
}

func TestBareAddressesReachTheWireVerbatim(t *testing.T) {
	good := []string{
		"first.last+tag@example.org", "a@b", "x_y-z@sub.example.org", "o'brien@example.org", "{a}|b~c@example.org",
		"a@123.example.org", strings.Repeat("a", 64) + "@example.org", "UPPER@Example.ORG",
	}
	for _, addr := range good {
		t.Run(addr, func(t *testing.T) {
			s := smtptest.NewServer(t)
			var rec lineRecorder
			if err := send(t, s, mailer.ModeNone, false, mailer.Message{To: []string{addr}, Subject: "s", Text: "t"}, rec.dialer()); err != nil {
				t.Fatal(err)
			}
			if got := rec.commands("RCPT"); len(got) != 1 || got[0] != "RCPT TO:<"+addr+">" {
				t.Fatalf("RCPT lines = %q", got)
			}
			if got := rec.commands("MAIL"); len(got) != 1 || !strings.HasPrefix(got[0], "MAIL FROM:<no-reply@example.org>") {
				t.Fatalf("MAIL lines = %q", got)
			}
			if m := s.Messages(); len(m) != 1 || m[0].To[0] != addr {
				t.Fatalf("messages = %+v", m)
			}
		})
	}
}

func TestSenderMustBeABareAddress(t *testing.T) {
	for _, from := range []string{"226os <no-reply@example.org>", `"a> BODY=8BITMIME"@example.org`, "<a@example.org>", "é@example.org"} {
		var dials atomic.Int32
		c := mailer.SMTPConfig{Host: "127.0.0.1", Port: 25, From: from, TLS: mailer.ModeNone}
		err := mailer.NewSMTP(c, countingDialer(&dials)).Send(t.Context(), okMsg)
		if !errors.Is(err, mailer.ErrNotConfigured) || dials.Load() != 0 {
			t.Fatalf("From %q: err = %v, dials = %d", from, err, dials.Load())
		}
	}
}

// With implicit TLS deliver swaps its connection for the TLS one while the context's AfterFunc may
// fire on another goroutine; this must be race-free whatever the timing. Run under -race.
func TestImplicitTLSTimeoutRaceStress(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithImplicitTLS())
	var wg sync.WaitGroup
	for w := range 4 {
		wg.Go(func() {
			for i := range 300 {
				timeout := time.Duration((i+w*7)%80+1) * 100 * time.Microsecond // 0.1 to 8 ms: spans the handshake and the rest of the exchange
				_ = send(t, s, mailer.ModeTLS, false, okMsg, mailer.WithTimeout(timeout))
			}
		})
	}
	wg.Wait()
}

func TestEndlessReplyIsCappedAndMemoryStaysSmall(t *testing.T) {
	for _, cmd := range []string{"GREETING", "EHLO"} {
		t.Run(cmd, func(t *testing.T) {
			s := smtptest.NewServer(t, smtptest.WithEndlessReply(cmd))
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			err := send(t, s, mailer.ModeNone, false, okMsg, mailer.WithTimeout(time.Second))
			runtime.ReadMemStats(&after)
			if !errors.Is(err, mailer.ErrTemporary) || errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("err = %v, want the temporary class from the cap, not a timeout", err)
			}
			if !strings.Contains(err.Error(), "64 KiB") || strings.Contains(err.Error(), "xxxx") {
				t.Fatalf("error must name the cap and not the content: %v", err)
			}
			if grown := after.TotalAlloc - before.TotalAlloc; grown > 8<<20 {
				t.Fatalf("allocated %d bytes while reading an endless reply", grown)
			}
		})
	}
}

func TestCredentialsAreNotPrintable(t *testing.T) {
	cfg := mailer.SMTPConfig{Host: "smtp.example.org", Port: 587, Username: user, Password: pass, From: "no-reply@example.org", TLS: mailer.ModeSTARTTLS}
	m := mailer.NewSMTP(cfg)
	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("x", "cfg", cfg, "cfgptr", &cfg, "mailer", m)
	outs := map[string]string{
		"cfg %v": fmt.Sprintf("%v", cfg), "cfg %+v": fmt.Sprintf("%+v", cfg), "cfg %#v": fmt.Sprintf("%#v", cfg),
		"cfgptr %v": fmt.Sprintf("%v", &cfg), "cfgptr %+v": fmt.Sprintf("%+v", &cfg), "cfgptr %#v": fmt.Sprintf("%#v", &cfg),
		"mailer %v": fmt.Sprintf("%v", m), "mailer %+v": fmt.Sprintf("%+v", m), "mailer %#v": fmt.Sprintf("%#v", m),
		"mailer %s": fmt.Sprintf("%s", m), "slog": buf.String(),
	}
	for name, out := range outs {
		for _, secret := range []string{user, pass, "no-reply@example.org"} {
			if strings.Contains(out, secret) {
				t.Fatalf("%s prints %q: %s", name, secret, out)
			}
		}
	}
	for _, name := range []string{"cfg %v", "cfg %+v", "cfg %#v", "cfgptr %v", "mailer %v", "mailer %#v"} {
		if out := outs[name]; !strings.Contains(out, "smtp.example.org") || !strings.Contains(out, "587") || !strings.Contains(out, "starttls") {
			t.Fatalf("%s lost the host, port or mode: %s", name, out)
		}
	}
	if !strings.Contains(outs["slog"], `"Host":"smtp.example.org"`) || !strings.Contains(outs["slog"], "[redacted]") {
		t.Fatalf("slog output = %s", outs["slog"])
	}
}

func TestSubjectTabIsAllowed(t *testing.T) {
	s := smtptest.NewServer(t)
	if err := send(t, s, mailer.ModeNone, false, mailer.Message{To: okMsg.To, Subject: "a\tb", Text: "t"}); err != nil {
		t.Fatal(err)
	}
}

func TestASCIISubjectLookingLikeAnEncodedWordIsQEncoded(t *testing.T) {
	for _, subject := range []string{"=?utf-8?q?evil?=", "pay =?UTF-8?B?aGk=?= now", "a=?b"} {
		s := smtptest.NewServer(t)
		if err := send(t, s, mailer.ModeNone, false, mailer.Message{To: okMsg.To, Subject: subject, Text: "t"}); err != nil {
			t.Fatal(err)
		}
		msg, err := mail.ReadMessage(strings.NewReader(s.Messages()[0].Data))
		if err != nil {
			t.Fatal(err)
		}
		raw := msg.Header.Get("Subject")
		if !strings.HasPrefix(raw, "=?utf-8?q?") || strings.Contains(raw, "?q?evil") || strings.Contains(raw, "?B?") {
			t.Fatalf("raw subject %q is not safely encoded", raw)
		}
		if got, err := new(mime.WordDecoder).DecodeHeader(raw); err != nil || got != subject {
			t.Fatalf("decoded %q (%v), want %q", got, err, subject)
		}
	}
}

func TestPlainASCIISubjectIsNotEncoded(t *testing.T) {
	s := smtptest.NewServer(t)
	if err := send(t, s, mailer.ModeNone, false, mailer.Message{To: okMsg.To, Subject: "Reset = your password?", Text: "t"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(s.Messages()[0].Data, "\r\nSubject: Reset = your password?\r\n") {
		t.Fatalf("data = %q", s.Messages()[0].Data)
	}
}

func TestNilTLSConfigMeansDefaults(t *testing.T) {
	s := smtptest.NewServer(t, smtptest.WithSTARTTLS())
	// Defaults do not trust the test certificate: a certificate error, and no panic.
	err := mailer.NewSMTP(cfgFor(s, mailer.ModeSTARTTLS, false), mailer.WithTLSConfig(nil)).Send(t.Context(), okMsg)
	if !errors.Is(err, mailer.ErrPermanent) {
		t.Fatalf("err = %v", err)
	}
}

func TestCallerCannotDisableCertificateVerification(t *testing.T) {
	for name, tc := range map[string]struct {
		mode   mailer.Mode
		server smtptest.Option
	}{
		"starttls": {mailer.ModeSTARTTLS, smtptest.WithSTARTTLS()},
		"tls":      {mailer.ModeTLS, smtptest.WithImplicitTLS()},
	} {
		t.Run(name, func(t *testing.T) {
			s := smtptest.NewServer(t, tc.server)
			skip := &tls.Config{InsecureSkipVerify: true} //nolint:gosec // the test proves the adapter overrides this
			err := mailer.NewSMTP(cfgFor(s, tc.mode, false), mailer.WithTLSConfig(skip)).Send(t.Context(), okMsg)
			if !errors.Is(err, mailer.ErrPermanent) || len(s.Messages()) != 0 {
				t.Fatalf("err = %v, messages = %d", err, len(s.Messages()))
			}
			if !skip.InsecureSkipVerify {
				t.Fatal("the caller's config was modified")
			}
		})
	}
}
