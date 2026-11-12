package mailer

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

// Mode is the transport security of the SMTP connection.
type Mode string

// Transport modes.
const (
	// ModeNone is plain SMTP, for a local catcher such as Mailpit. Credentials are refused with it.
	ModeNone Mode = "none"
	// ModeSTARTTLS upgrades the connection with STARTTLS, which is required (no silent downgrade).
	ModeSTARTTLS Mode = "starttls"
	// ModeTLS is implicit TLS from the first byte.
	ModeTLS Mode = "tls"
)

const (
	dialTimeout    = 10 * time.Second
	defaultTimeout = 30 * time.Second
	defaultEHLO    = "localhost"
)

// SMTPConfig configures the SMTP adapter. It mirrors the application configuration without
// importing it: the composition root maps one onto the other.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string // with Password; AUTH PLAIN is only ever sent over TLS
	Password string
	From     string
	TLS      Mode
}

// String is the redacted form used by %v and %+v: host, port and TLS mode only.
func (c SMTPConfig) String() string {
	return fmt.Sprintf("{Host:%s Port:%d Username:[redacted] Password:[redacted] From:[redacted] TLS:%s}", c.Host, c.Port, c.TLS)
}

// GoString is the redacted form used by %#v.
func (c SMTPConfig) GoString() string { return "mailer.SMTPConfig" + c.String() }

// LogValue keeps slog from printing the credentials or the sender.
func (c SMTPConfig) LogValue() slog.Value {
	return slog.GroupValue(slog.String("Host", c.Host), slog.Int("Port", c.Port), slog.String("TLS", string(c.TLS)),
		slog.String("Username", "[redacted]"), slog.String("Password", "[redacted]"), slog.String("From", "[redacted]"))
}

// Option customises the SMTP adapter, mostly for tests.
type Option func(*smtpMailer)

// WithDialer replaces the TCP dialer.
func WithDialer(d func(ctx context.Context, network, addr string) (net.Conn, error)) Option {
	return func(s *smtpMailer) { s.dial = d }
}

// WithTLSConfig sets the base TLS configuration (for example RootCAs). ServerName is always the
// configured host and the minimum version is at least TLS 1.2. A nil c means the defaults. Certificate verification cannot be turned off.
func WithTLSConfig(c *tls.Config) Option {
	return func(s *smtpMailer) {
		if c != nil {
			s.tlsBase = c.Clone()
		}
	}
}

// WithClock replaces the Date header clock.
func WithClock(now func() time.Time) Option { return func(s *smtpMailer) { s.now = now } }

// WithHostname sets the EHLO name and the Message-ID domain.
func WithHostname(h string) Option { return func(s *smtpMailer) { s.hostname = h } }

// WithRand replaces the Message-ID entropy source (default crypto/rand).
func WithRand(r io.Reader) Option { return func(s *smtpMailer) { s.rng = r } }

// WithTimeout bounds a whole Send (default 30 s); the caller's context can only shorten it.
func WithTimeout(d time.Duration) Option { return func(s *smtpMailer) { s.timeout = d } }

type smtpMailer struct {
	cfg      SMTPConfig
	from     *mail.Address
	cfgErr   error
	dial     func(ctx context.Context, network, addr string) (net.Conn, error)
	tlsBase  *tls.Config
	now      func() time.Time
	hostname string
	rng      io.Reader
	timeout  time.Duration
}

// String is the redacted form of the adapter, so printing it never reveals its configuration.
func (s smtpMailer) String() string { return "smtp mailer " + s.cfg.String() }

// GoString is the redacted form used by %#v.
func (s smtpMailer) GoString() string { return "&mailer.smtpMailer " + s.cfg.String() }

// NewSMTP returns the SMTP adapter. A configuration it refuses is reported by Send as
// ErrNotConfigured, before any network I/O.
func NewSMTP(cfg SMTPConfig, opts ...Option) Mailer {
	s := &smtpMailer{
		cfg:     cfg,
		dial:    (&net.Dialer{Timeout: dialTimeout}).DialContext,
		tlsBase: &tls.Config{},
		now:     time.Now,
		rng:     rand.Reader,
		timeout: defaultTimeout,
	}
	for _, o := range opts {
		o(s)
	}
	s.from, s.cfgErr = checkConfig(cfg)
	return s
}

func checkConfig(cfg SMTPConfig) (*mail.Address, error) {
	bad := func(why string) error { return fmt.Errorf("%w: %s", ErrNotConfigured, why) }
	if cfg.Host == "" || cfg.Port < 1 || cfg.Port > 65535 {
		return nil, bad("host and port are required")
	}
	switch cfg.TLS {
	case ModeNone, ModeSTARTTLS, ModeTLS:
	default:
		return nil, bad("unknown TLS mode")
	}
	if cfg.TLS == ModeNone && (cfg.Username != "" || cfg.Password != "") {
		return nil, bad("credentials require TLS")
	}
	from, err := parseAddress(cfg.From)
	if err != nil {
		return nil, bad("invalid sender address")
	}
	return from, nil
}

// Send validates m, then delivers it. See the package errors for the failure classes.
func (s *smtpMailer) Send(ctx context.Context, m Message) error {
	to, err := validate(m)
	if err != nil {
		return err
	}
	if s.cfgErr != nil {
		return s.cfgErr
	}
	msgHost := s.hostname
	if msgHost == "" {
		msgHost = s.from.Address[strings.LastIndex(s.from.Address, "@")+1:]
	}
	data, err := build(s.from, to, m, s.now(), s.rng, msgHost)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()
	return s.deliver(ctx, to, data)
}

func (s *smtpMailer) tlsConfig() *tls.Config {
	c := s.tlsBase.Clone()
	c.ServerName = s.cfg.Host
	c.InsecureSkipVerify = false // never negotiable: the server certificate is always verified
	if c.MinVersion < tls.VersionTLS12 {
		c.MinVersion = tls.VersionTLS12
	}
	return c
}

// deliver runs one SMTP transaction. ctx carries the whole-Send deadline.
func (s *smtpMailer) deliver(ctx context.Context, to []*mail.Address, data []byte) error {
	var limited *cappedConn
	fail := func(phase string, err error) error {
		if limited != nil && limited.exceeded {
			return fmt.Errorf("smtp %s: server sent more than the %d KiB read cap: %w", phase, maxReadBytes>>10, ErrTemporary)
		}
		return classify(ctx, phase, err)
	}
	raw, err := s.dial(ctx, "tcp", net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port)))
	if err != nil {
		return fail("dial", err)
	}
	// raw is never reassigned: the AfterFunc goroutine and the deferred Close use it, while conn
	// (the TLS wrapper in implicit mode) is only touched by this goroutine.
	defer func() { _ = raw.Close() }()
	deadline, _ := ctx.Deadline()
	_ = raw.SetDeadline(deadline) // a conn that rejects deadlines still fails on the context below
	stop := context.AfterFunc(ctx, func() { _ = raw.Close() })
	defer stop()
	limited = &cappedConn{Conn: raw, remaining: maxReadBytes}
	conn := net.Conn(limited)

	if s.cfg.TLS == ModeTLS {
		tc := tls.Client(conn, s.tlsConfig())
		if err := tc.HandshakeContext(ctx); err != nil {
			return fail("tls", err)
		}
		conn = tc
	}
	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fail("greeting", err)
	}
	name := s.hostname
	if name == "" {
		name = defaultEHLO
	}
	if err := c.Hello(name); err != nil {
		return fail("ehlo", err)
	}
	if s.cfg.TLS == ModeSTARTTLS {
		if ok, _ := c.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp starttls: not offered by the server: %w", ErrPermanent)
		}
		if err := c.StartTLS(s.tlsConfig()); err != nil {
			return fail("starttls", err)
		}
	}
	if s.cfg.Username != "" {
		if ok, _ := c.Extension("AUTH"); !ok {
			return fmt.Errorf("smtp auth: not offered by the server: %w", ErrPermanent)
		}
		if err := c.Auth(smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)); err != nil {
			return fail("auth", err)
		}
	}
	if err := c.Mail(s.from.Address); err != nil {
		return fail("mail", err)
	}
	for _, a := range to {
		if err := c.Rcpt(a.Address); err != nil {
			return fail("rcpt", err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fail("data", err)
	}
	if _, err := w.Write(data); err != nil {
		return fail("data", err)
	}
	if err := w.Close(); err != nil {
		return fail("end", err)
	}
	_ = c.Quit() // the message is accepted; a failing goodbye is not a failed send
	return nil
}

// maxReadBytes caps the bytes read from the server during one Send. A whole transaction, TLS
// handshake included, needs a few KiB; the cap keeps a hostile never-ending reply from exhausting memory.
const maxReadBytes = 64 << 10

// cappedConn closes the connection once more than the allowed number of bytes has been read.
// It is only read by the goroutine running deliver, so exceeded needs no lock.
type cappedConn struct {
	net.Conn
	remaining int
	exceeded  bool
}

func (c *cappedConn) Read(p []byte) (int, error) {
	if c.remaining <= 0 {
		c.exceeded = true
		_ = c.Close()
		return 0, errors.New("read cap exceeded")
	}
	n, err := c.Conn.Read(p[:min(len(p), c.remaining)])
	c.remaining -= n
	return n, err
}

// classify turns an SMTP, network or TLS error into a sentinel with the phase and reply code only.
func classify(ctx context.Context, phase string, err error) error {
	var te *textproto.Error
	if errors.As(err, &te) {
		kind := ErrPermanent
		if te.Code < 500 {
			kind = ErrTemporary
		}
		return fmt.Errorf("smtp %s: reply %d: %w", phase, te.Code, kind)
	}
	var cert *tls.CertificateVerificationError
	if errors.As(err, &cert) {
		return fmt.Errorf("smtp %s: certificate rejected: %w", phase, ErrPermanent)
	}
	if cerr := ctx.Err(); cerr != nil {
		return fmt.Errorf("smtp %s: %w: %w", phase, ErrTemporary, cerr)
	}
	return fmt.Errorf("smtp %s: network or protocol failure: %w", phase, ErrTemporary)
}
