// Package service implements the S1 identity use cases: sign-up, sign-in,
// sign-out, session state, CSRF tokens, e-mail verification and password
// reset (ADR-0003). It owns every policy decision; transport only encodes
// results.
//
// Anti-enumeration: sign-up, reset request and sign-in return generic
// responses. Sign-in still pays one Argon2id on the unknown-email path
// (password.VerifyDummy) so timing does not reveal account existence.
package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/netip"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/password"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

const (
	idleTTL       = 2 * time.Hour
	absoluteTTL   = 7 * 24 * time.Hour
	verifyTTL     = 24 * time.Hour
	resetTTL      = 30 * time.Minute
	lastSeenEvery = 5 * time.Minute

	// Rate limits (ADR-0003), fixed windows.
	signInAccountLimit = 5
	signInIPLimit      = 20
	signInWindow       = 15 * time.Minute
	signUpIPLimit      = 5
	signUpWindow       = time.Hour
	resetAccountLimit  = 3
	resetIPLimit       = 10
	resetWindow        = time.Hour
	confirmIPLimit     = 10
	confirmWindow      = time.Hour
)

// Store is the persistence port. The repository implements it over sqlc;
// tests use fakes. Every method takes the request context.
type Store interface {
	CreateAccount(ctx context.Context, in NewAccount) (domain.Account, error)
	AccountByEmail(ctx context.Context, email string) (domain.Account, error)
	AccountByID(ctx context.Context, id string) (domain.Account, error)

	CreateSession(ctx context.Context, s domain.Session) (string, error)
	SessionByTokenHash(ctx context.Context, hash [32]byte) (domain.Session, error)
	RevokeSession(ctx context.Context, id, reason string, now time.Time) error
	TouchSession(ctx context.Context, id string, lastSeenAt, idleExpiresAt time.Time) error

	CreateEmailVerification(ctx context.Context, userID, email string, hash [32]byte, expiresAt time.Time) error
	VerifyEmail(ctx context.Context, hash [32]byte, now time.Time) error

	CreatePasswordReset(ctx context.Context, userID string, hash [32]byte, expiresAt time.Time) error
	// ResetPassword consumes the token, stores the new PHC hash and revokes
	// every session of the account in one atomic write.
	ResetPassword(ctx context.Context, hash [32]byte, phc string, now time.Time) error

	// Hit records one attempt in a fixed-window bucket and reports whether it
	// is allowed plus the retry delay when it is not. key is already hashed.
	Hit(ctx context.Context, scope string, key [32]byte, limit int, window time.Duration, now time.Time) (bool, time.Duration, error)

	// GiveBack atomically releases one hit from the bucket window that contains
	// now (never below zero). now must be the instant the matching Hit used.
	GiveBack(ctx context.Context, scope string, key [32]byte, window time.Duration, now time.Time) error

	// ---------- profiles ----------

	// ProfileByUserID returns the caller's own profile row plus skill slugs.
	ProfileByUserID(ctx context.Context, userID string) (domain.Profile, error)
	// ProfileByUsername returns the profile joined with the account status and
	// join date the public visibility decision needs.
	ProfileByUsername(ctx context.Context, username string) (domain.ProfileRecord, error)
	// ResolveSkillSlugs maps skill slugs to active tag ids; the count may be
	// smaller than the input when a slug is unknown or the tag is hidden.
	ResolveSkillSlugs(ctx context.Context, slugs []string) ([]string, error)
	// UpdateProfile applies a resolved merge patch atomically: profile columns
	// and the skill set commit or roll back together in one transaction.
	UpdateProfile(ctx context.Context, userID string, w domain.ProfileWrite) error
}

// NewAccount is the sign-up write model. Password is the PHC hash.
type NewAccount struct {
	Email, Username, DisplayName, PasswordHash, UIlang string
}

// SignUpInput is the validated sign-up request.
type SignUpInput struct {
	Email, Password, Username, DisplayName, UIlang, IP string
}

// SignInInput is the validated sign-in request. IP is the trusted client
// address; CurrentSessionID is the session the request's cookie resolved to,
// if any — sign-in rotates exactly that session away (ADR-0003), never the
// account's other sessions.
type SignInInput struct {
	Email, Password, IP, CurrentSessionID string
}

// Service is the identity use-case boundary.
type Service struct {
	store Store
	mail  mailer.Mailer
	// Purpose keys derived from the shared secret (the same secret signs
	// pagination cursors), so no purpose ever uses the secret itself.
	bucketKeyKey []byte // HMAC key for rate-limit bucket keys
	csrfKey      []byte // HMAC key for CSRF tokens
	now          func() time.Time
	baseURL      string // public origin for mail links; empty means bare codes only
	// spawn runs work off the request path (a goroutine in production; tests
	// run it inline to stay deterministic).
	spawn func(func())
}

// New returns an identity service. hmacKey is the shared secret, at least 32
// bytes (config.Load enforces it); the purpose keys are derived from it. mail may be nil: verification and reset
// mails are then skipped (local-dev without SMTP relies on this).
func New(store Store, mail mailer.Mailer, hmacKey []byte, opts ...Option) *Service {
	svc := &Service{
		store: store, mail: mail, now: time.Now, spawn: func(f func()) { go f() },
		bucketKeyKey: subkey(hmacKey, "rate-limit-bucket"),
		csrfKey:      subkey(hmacKey, "csrf"),
	}
	for _, o := range opts {
		o(svc)
	}
	return svc
}

// Option customises a Service.
type Option func(*Service)

// WithBackground sets how work that must leave the request path (mails, reset
// tokens) is run. The default starts a goroutine; a test over a single
// database connection passes an inline runner.
func WithBackground(run func(func())) Option { return func(s *Service) { s.spawn = run } }

// WithBaseURL sets the public origin (no trailing slash) that verification and reset mails link to. Empty
// keeps the mails to the bare code. config.Load validates the value; nothing here logs it.
func WithBaseURL(origin string) Option { return func(s *Service) { s.baseURL = origin } }

// subkey is HKDF-Expand (RFC 5869, SHA-256, one 32-byte block) of the shared
// secret for one purpose: HMAC(secret, info || 0x01). The secret is already
// high-entropy (at least 32 bytes), so the extract step is not needed.
func subkey(secret []byte, purpose string) []byte {
	m := hmac.New(sha256.New, secret)
	m.Write([]byte("226os/identity/" + purpose))
	m.Write([]byte{1})
	return m.Sum(nil)
}

// SessionCookie is what transport needs to set the session cookie.
type SessionCookie struct {
	Name    string
	Value   string // plaintext token, returned once, never stored
	CSRF    string // synchronizer token derived from the session
	MaxAge  int
	Expires time.Time
}

// SessionView is the authenticated caller plus the material transport needs:
// the session id (rotation, sign-out), the stored token hash (CSRF binding)
// and the absolute expiry for the cookie Max-Age.
type SessionView struct {
	Account   domain.Account
	SessionID string
	TokenHash [32]byte
	CSRF      string
	Expires   time.Time
}

// ---------- helpers ----------

func normalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

// randRead is crypto/rand.Read behind a variable so tests can exercise the
// entropy-failure branches.
var randRead = rand.Read

// verifyDummy is password.VerifyDummy behind a variable so tests can observe
// the timing-equalization work on enumeration paths.
var verifyDummy = password.VerifyDummy

// hashPassword is password.Hash behind a variable so tests can inject the
// entropy-failure branch.
var hashPassword = password.Hash

func randToken() (string, [32]byte, error) {
	b := make([]byte, 32)
	if _, err := randRead(b); err != nil {
		return "", [32]byte{}, err
	}
	return base64.RawURLEncoding.EncodeToString(b), sha256.Sum256(b), nil
}

func hashToken(plain string) ([32]byte, error) {
	b, err := base64.RawURLEncoding.DecodeString(plain)
	if err != nil || len(b) != 32 {
		return [32]byte{}, errors.New("bad token encoding")
	}
	return sha256.Sum256(b), nil
}

// bucketKey is the HMAC the rate-limit table stores instead of the raw IP or
// e-mail (ADR-0003: "keyed by an HMAC of the IP or the normalized email").
func (s *Service) bucketKey(scope, value string) [32]byte {
	m := hmac.New(sha256.New, s.bucketKeyKey)
	m.Write([]byte(scope + "|" + value))
	var out [32]byte
	copy(out[:], m.Sum(nil))
	return out
}

// allow records one hit and maps an exhausted bucket to 429 RATE_LIMITED.
func (s *Service) allow(ctx context.Context, scope, value string, limit int, window time.Duration) error {
	ok, retry, err := s.store.Hit(ctx, scope, s.bucketKey(scope, value), limit, window, s.now())
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err) // fail closed per ADR-0003
	}
	if !ok {
		return apperr.New(apperr.CodeRateLimited, "Too many requests.").WithRetryAfter(retry)
	}
	return nil
}

// clientIP reduces the address to the rate-limit identity. Only the
// Traefik-trusted X-Forwarded-For entry may be used; transport passes the
// resolved value, this keeps the /24-or-/48 prefix rule local.
func ipKey(ip string) string {
	a, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		return "unknown"
	}
	if a.Is4() {
		b := a.As4()
		return netip.AddrFrom4([4]byte{b[0], b[1], b[2], 0}).String() + "/24"
	}
	// Everything ParseAddr returns that is not Is4 is IPv6 (BitLen 128), so
	// Prefix(48) cannot fail.
	p, _ := a.Prefix(48)
	return p.String()
}

// csrfToken derives the synchronizer token from stored secret material, so
// the plaintext is never persisted: for a session it is HMAC(token_hash) and
// for the anonymous pre-session flow it is HMAC(pre-session secret). The
// derived value rotates with the session, as ADR-0003 requires.
func (s *Service) csrfToken(secret [32]byte) string {
	m := hmac.New(sha256.New, s.csrfKey)
	m.Write(secret[:])
	m.Write([]byte("|226os-csrf"))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// sessionSecrets builds the cookie + CSRF material for a fresh session.
func (s *Service) sessionSecrets() (SessionCookie, domain.Session, error) {
	plain, tokenHash, err := randToken()
	if err != nil {
		return SessionCookie{}, domain.Session{}, err
	}
	csrf := s.csrfToken(tokenHash)
	csrfHash := sha256.Sum256([]byte(csrf))
	now := s.now()
	sess := domain.Session{
		TokenHash: tokenHash, CSRFHash: csrfHash,
		CreatedAt: now, LastSeenAt: now, AuthenticatedAt: now,
		IdleExpiresAt: now.Add(idleTTL), AbsoluteExpiresAt: now.Add(absoluteTTL),
	}
	return SessionCookie{
		Value: plain, CSRF: csrf, MaxAge: int(absoluteTTL.Seconds()), Expires: now.Add(absoluteTTL),
	}, sess, nil
}

// ---------- sign-up ----------

// SignUp creates the account and queues the verification mail. It returns nil
// on every non-error path, including "email already registered": the caller
// always answers 202 Accepted (no enumeration).
func (s *Service) SignUp(ctx context.Context, in SignUpInput) error {
	if err := s.allow(ctx, "signup:ip", ipKey(in.IP), signUpIPLimit, signUpWindow); err != nil {
		return err
	}
	email := normalizeEmail(in.Email)
	// Policy check before the lookup: an invalid password must get the same
	// 422 whether or not the e-mail is registered, or the response leaks
	// account existence.
	if !password.Valid(in.Password) {
		return apperr.New(apperr.CodeValidation, "password does not meet the policy.").
			WithViolations(apperr.Violation{Field: "/password", Code: "invalid_value"})
	}
	// PostgreSQL text holds neither NUL nor invalid UTF-8: refuse them here (422, same answer whether or not the e-mail
	// is registered, before any hash or write) instead of a 22021 constraint failure mapped to 500.
	if !validText(in.DisplayName) {
		return apperr.New(apperr.CodeValidation, "display_name must be valid text.").
			WithViolations(apperr.Violation{Field: "/display_name", Code: "invalid_value"})
	}
	if _, err := s.store.AccountByEmail(ctx, email); err == nil {
		// Pay the same Argon2id cost as a new account so the response time
		// does not reveal that the e-mail is already registered.
		verifyDummy(in.Password)
		return nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	phc, err := hashPassword(in.Password)
	if err != nil {
		// Valid() already passed, so this is an entropy failure, not policy.
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	display := strings.TrimSpace(in.DisplayName)
	if display == "" {
		display = in.Username
	}
	acct, err := s.store.CreateAccount(ctx, NewAccount{
		Email: email, Username: in.Username, DisplayName: display,
		PasswordHash: phc, UIlang: in.UIlang,
	})
	if err != nil {
		if errors.Is(err, domain.ErrEmailTaken) || errors.Is(err, domain.ErrUsernameTaken) {
			return nil // race with an existing account: still generic 202
		}
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	s.sendVerification(ctx, acct)
	return nil
}

// sendVerification issues a 24 h single-use token and mails the link. Mail
// failures are logged by the caller's mailer; they never change the response.
func (s *Service) sendVerification(ctx context.Context, acct domain.Account) {
	if s.mail == nil {
		return
	}
	plain, hash, err := randToken()
	if err != nil {
		return
	}
	if err := s.store.CreateEmailVerification(ctx, acct.ID, acct.Email, hash, s.now().Add(verifyTTL)); err != nil {
		return
	}
	s.sendMail(ctx, mailer.Message{
		To: []string{acct.Email}, Subject: "Verify your 226os e-mail",
		Text: s.mailBody("Confirm your e-mail address:", "/verifier-email", plain, "24 hours"),
	})
}

// mailBody is the text of a verification or reset mail. Without a base URL it is the intro, the bare code and
// the lifetime. With one, the code travels in the URL fragment (never sent to a server, so not in an access
// log or a Referer) and stays on its own line as a fallback. The body is never logged.
func (s *Service) mailBody(intro, page, code, lifetime string) string {
	if s.baseURL == "" {
		return intro + "\n\n" + code + "\n\nThis link expires in " + lifetime + "."
	}
	return intro + "\n\n" + s.baseURL + page + "#token=" + code +
		"\n\nOr paste this code:\n" + code + "\n\nThis link expires in " + lifetime + "."
}

// mailTimeout bounds a detached send: the request path must not wait on SMTP.
const mailTimeout = 30 * time.Second

// sendMail delivers off the request path so silent and known-account
// responses stay in the same time envelope (no SMTP-timing oracle).
func (s *Service) sendMail(ctx context.Context, m mailer.Message) {
	s.spawn(func() {
		c, cancel := context.WithTimeout(context.WithoutCancel(ctx), mailTimeout)
		defer cancel()
		_ = s.mail.Send(c, m)
	})
}

// ---------- sign-in ----------

// SignInResult is the outcome transport encodes: cookie material plus the
// account. On failure it is the generic invalid-credentials error.
type SignInResult struct {
	Cookie  SessionCookie
	Account domain.Account
}

// SignIn authenticates and rotates the session. Every failure path returns
// the same ErrInvalidCredentials problem; rate limiting is checked first.
func (s *Service) SignIn(ctx context.Context, in SignInInput) (SignInResult, error) {
	ip := ipKey(in.IP)
	if err := s.allow(ctx, "signin:ip", ip, signInIPLimit, signInWindow); err != nil {
		return SignInResult{}, err
	}
	email := normalizeEmail(in.Email)
	// ADR-0003: 5 failures per account per window. The attempt is counted UP
	// FRONT by one atomic Hit, before the lookup and the password check, so N
	// concurrent attempts cannot all read "below the limit" (issue #97). It is
	// counted the same way for a known and an unknown e-mail (no 429-vs-401
	// oracle) and a store error fails closed. A successful sign-in gives its hit
	// back below, so only failures stay counted.
	hitAt := s.now()
	accountKey := s.bucketKey("signin:account", email)
	ok, retry, err := s.store.Hit(ctx, "signin:account", accountKey, signInAccountLimit, signInWindow, hitAt)
	if err != nil {
		return SignInResult{}, apperr.Wrap(apperr.CodeInternal, err) // fail closed per ADR-0003
	}
	if !ok {
		return SignInResult{}, apperr.New(apperr.CodeRateLimited, "Too many requests.").WithRetryAfter(retry)
	}
	acct, err := s.store.AccountByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			verifyDummy(in.Password)
			return SignInResult{}, invalidCredentials()
		}
		return SignInResult{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	if acct.Status != "active" {
		verifyDummy(in.Password)
		return SignInResult{}, invalidCredentials()
	}
	verified, err := password.Verify(in.Password, acct.PasswordHash)
	if err != nil || !verified {
		return SignInResult{}, invalidCredentials()
	}
	cookie, sess, err := s.sessionSecrets()
	if err != nil {
		return SignInResult{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	sess.UserID = acct.ID
	sess.ID = uuid.NewString()
	if _, err := s.store.CreateSession(ctx, sess); err != nil {
		return SignInResult{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	// Rotation: only the session this request carried is revoked.
	if in.CurrentSessionID != "" {
		if err := s.store.RevokeSession(ctx, in.CurrentSessionID, "rotated", s.now()); err != nil {
			return SignInResult{}, apperr.Wrap(apperr.CodeInternal, err)
		}
	}
	// The sign-in succeeded: release the up-front hit so only failures count.
	// A give-back error must not fail an authenticated sign-in; the hit then
	// stays counted (the strict side) until the window ends. The log carries
	// neither the e-mail nor the bucket key.
	if err := s.store.GiveBack(ctx, "signin:account", accountKey, signInWindow, hitAt); err != nil {
		slog.WarnContext(ctx, "identity sign-in rate limit give back failed", slog.String("error", err.Error()))
	}
	cookie.Name = "" // transport picks the profile-specific name
	return SignInResult{Cookie: cookie, Account: acct}, nil
}

// ---------- session ----------

// ResolveSession turns a cookie plaintext into the live session view. It
// returns ErrUnauthenticated for missing, unknown, expired or revoked tokens
// alike, and refreshes last_seen_at at most every five minutes.
func (s *Service) ResolveSession(ctx context.Context, plain string) (SessionView, error) {
	if plain == "" {
		return SessionView{}, unauthenticated()
	}
	hash, err := hashToken(plain)
	if err != nil {
		return SessionView{}, unauthenticated()
	}
	sess, err := s.store.SessionByTokenHash(ctx, hash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return SessionView{}, unauthenticated()
		}
		return SessionView{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	now := s.now()
	if !sess.Active(now) {
		return SessionView{}, unauthenticated()
	}
	if now.Sub(sess.LastSeenAt) >= lastSeenEvery {
		// Cap the refreshed idle expiry at the absolute expiry: beyond the
		// session's last two hours an uncapped value would violate
		// sessions_expiry_ck and the touch would silently fail forever.
		idle := now.Add(idleTTL)
		if idle.After(sess.AbsoluteExpiresAt) {
			idle = sess.AbsoluteExpiresAt
		}
		_ = s.store.TouchSession(ctx, sess.ID, now, idle)
	}
	acct, err := s.store.AccountByID(ctx, sess.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return SessionView{}, unauthenticated()
		}
		return SessionView{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	if acct.Status != "active" {
		return SessionView{}, unauthenticated()
	}
	return SessionView{
		Account: acct, SessionID: sess.ID, TokenHash: sess.TokenHash,
		CSRF: s.csrfToken(sess.TokenHash), Expires: sess.AbsoluteExpiresAt,
	}, nil
}

// SignOut revokes the caller's session. Anonymous or invalid sessions are a
// no-op success so sign-out never leaks session state.
func (s *Service) SignOut(ctx context.Context, plain string) error {
	if plain == "" {
		return nil
	}
	hash, err := hashToken(plain)
	if err != nil {
		return nil
	}
	sess, err := s.store.SessionByTokenHash(ctx, hash)
	if err != nil {
		// Only a missing session is the silent no-op; a store failure must
		// surface so the client retries instead of keeping a live session.
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	if err := s.store.RevokeSession(ctx, sess.ID, "logout", s.now()); err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	return nil
}

// ---------- CSRF ----------

// AnonymousCSRF issues the pre-session CSRF flow: a random secret that
// transport stores in the host-only pre-session cookie; the synchronizer
// token returned to the client is derived from it.
func (s *Service) AnonymousCSRF() (secret, token string, err error) {
	b := make([]byte, 32)
	if _, err := randRead(b); err != nil {
		return "", "", err
	}
	secret = base64.RawURLEncoding.EncodeToString(b)
	return secret, s.csrfToken(sha256.Sum256(b)), nil
}

// PreSessionCSRF derives the synchronizer token for an existing pre-session
// secret cookie, so GET /auth/csrf and GET /auth/session can answer without
// rotating the secret on every call. ok is false for a malformed secret.
func (s *Service) PreSessionCSRF(secret string) (token string, ok bool) {
	b, err := base64.RawURLEncoding.DecodeString(secret)
	if err != nil || len(b) != 32 {
		return "", false
	}
	return s.csrfToken(sha256.Sum256(b)), true
}

// CheckCSRF verifies the X-CSRF-Token header against session-bound or
// pre-session material. sessionHash is the stored session token hash (nil for
// anonymous calls, which use the pre-session secret instead).
func (s *Service) CheckCSRF(sessionHash *[32]byte, preSessionSecret, header string) bool {
	if header == "" {
		return false
	}
	var want string
	if sessionHash != nil {
		want = s.csrfToken(*sessionHash)
	} else if preSessionSecret != "" {
		b, err := base64.RawURLEncoding.DecodeString(preSessionSecret)
		if err != nil || len(b) != 32 {
			return false
		}
		want = s.csrfToken(sha256.Sum256(b))
	} else {
		return false
	}
	return hmac.Equal([]byte(want), []byte(header))
}

// ---------- e-mail verification ----------

// VerifyEmail consumes a 24 h single-use token and marks the account
// verified. Unknown, expired or replayed tokens all return INVALID_TOKEN.
func (s *Service) VerifyEmail(ctx context.Context, plain string) error {
	hash, err := hashToken(plain)
	if err != nil {
		return invalidToken()
	}
	if err := s.store.VerifyEmail(ctx, hash, s.now()); err != nil {
		if errors.Is(err, domain.ErrTokenInvalid) {
			return invalidToken()
		}
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	return nil
}

// ---------- password reset ----------

// RequestPasswordReset always succeeds silently: it mails a 30 min single-use
// token when the account exists and is active, and does nothing otherwise.
func (s *Service) RequestPasswordReset(ctx context.Context, email, ip string) error {
	if err := s.allow(ctx, "reset:ip", ipKey(ip), resetIPLimit, resetWindow); err != nil {
		return err
	}
	normalized := normalizeEmail(email)
	if err := s.allow(ctx, "reset:account", normalized, resetAccountLimit, resetWindow); err != nil {
		return err // hit before the lookup so 429 does not reveal account existence
	}
	acct, err := s.store.AccountByEmail(ctx, normalized)
	if err != nil || acct.Status != "active" || s.mail == nil {
		return nil // unknown, inactive, lookup failure or no mailer: same silent 202
	}
	plain, hash, err := randToken()
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	// The token write and the mail leave the request path together: the known
	// and the unknown account then do the same synchronous work (two bucket
	// hits and one lookup), so neither the response time nor a store error
	// tells them apart. A failed write simply sends no mail; the person asks again.
	expires, msg := s.now().Add(resetTTL), mailer.Message{
		To: []string{acct.Email}, Subject: "Reset your 226os password",
		Text: s.mailBody("Reset your password:", "/reinitialiser", plain, "30 minutes"),
	}
	s.spawn(func() {
		c, cancel := context.WithTimeout(context.WithoutCancel(ctx), mailTimeout)
		defer cancel()
		if err := s.store.CreatePasswordReset(c, acct.ID, hash, expires); err != nil {
			// Invisible to the caller by design; visible to operators (no email, no token).
			slog.WarnContext(c, "identity password reset token write failed", slog.String("error", err.Error()))
			return
		}
		_ = s.mail.Send(c, msg)
	})
	return nil
}

// ConfirmPasswordReset validates and hashes the new password first — a weak
// password must not burn the single-use token — then consumes the token and
// applies hash + session revocation in one atomic store call.
func (s *Service) ConfirmPasswordReset(ctx context.Context, plain, newPassword, ip string) error {
	// Anonymous endpoint that pays one Argon2id per call: bound it per IP
	// before any token or password work.
	if err := s.allow(ctx, "confirm:ip", ipKey(ip), confirmIPLimit, confirmWindow); err != nil {
		return err
	}
	hash, err := hashToken(plain)
	if err != nil {
		return invalidToken()
	}
	if !password.Valid(newPassword) {
		return apperr.New(apperr.CodeValidation, "password does not meet the policy.").
			WithViolations(apperr.Violation{Field: "/new_password", Code: "invalid_value"})
	}
	phc, err := hashPassword(newPassword)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	if err := s.store.ResetPassword(ctx, hash, phc, s.now()); err != nil {
		if errors.Is(err, domain.ErrTokenInvalid) {
			return invalidToken()
		}
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	return nil
}

// ---------- error mapping ----------

func unauthenticated() error {
	return apperr.New(apperr.CodeUnauthenticated, "Authentication required.")
}

func invalidCredentials() error {
	return apperr.New(apperr.CodeInvalidCredentials, "Invalid e-mail or password.")
}

func invalidToken() error {
	return apperr.New(apperr.CodeInvalidToken, "The token is invalid or expired.")
}

func notFound() error {
	return apperr.New(apperr.CodeNotFound, "Not found.")
}
