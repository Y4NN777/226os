package transport

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

// fakeStore is an in-memory service.Store for transport tests, mirroring the
// semantics the sqlc repository implements (ErrNotFound, ErrTokenInvalid,
// fixed-window buckets).
type fakeStore struct {
	accounts          map[string]domain.Account // by id
	byEmail           map[string]string         // email -> id
	sessions          map[string]domain.Session // by session id
	byToken           map[[32]byte]string       // token hash -> session id
	verifs            map[[32]byte]verifRow
	resets            map[[32]byte]verifRow
	buckets           map[string]int
	hits              []string
	failHits          bool
	failSessionLookup bool
	failRevoke        bool
}

type verifRow struct {
	userID    string
	expiresAt time.Time
	consumed  bool
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		accounts: map[string]domain.Account{}, byEmail: map[string]string{},
		sessions: map[string]domain.Session{}, byToken: map[[32]byte]string{},
		verifs: map[[32]byte]verifRow{}, resets: map[[32]byte]verifRow{},
		buckets: map[string]int{},
	}
}

func (f *fakeStore) CreateAccount(_ context.Context, in service.NewAccount) (domain.Account, error) {
	if _, ok := f.byEmail[in.Email]; ok {
		return domain.Account{}, domain.ErrEmailTaken
	}
	for _, a := range f.accounts {
		if a.Username == in.Username {
			return domain.Account{}, domain.ErrUsernameTaken
		}
	}
	a := domain.Account{
		ID: uuid.NewString(), Email: in.Email, Role: "member", Status: "active",
		Username: in.Username, DisplayName: in.DisplayName,
		PasswordHash: in.PasswordHash, CreatedAt: time.Now(),
	}
	f.accounts[a.ID] = a
	f.byEmail[a.Email] = a.ID
	return a, nil
}

func (f *fakeStore) AccountByEmail(_ context.Context, email string) (domain.Account, error) {
	id, ok := f.byEmail[email]
	if !ok {
		return domain.Account{}, domain.ErrNotFound
	}
	return f.accounts[id], nil
}

func (f *fakeStore) AccountByID(_ context.Context, id string) (domain.Account, error) {
	a, ok := f.accounts[id]
	if !ok {
		return domain.Account{}, domain.ErrNotFound
	}
	return a, nil
}

func (f *fakeStore) CreateSession(_ context.Context, s domain.Session) (string, error) {
	f.sessions[s.ID] = s
	f.byToken[s.TokenHash] = s.ID
	return s.ID, nil
}

func (f *fakeStore) SessionByTokenHash(_ context.Context, hash [32]byte) (domain.Session, error) {
	if f.failSessionLookup {
		return domain.Session{}, errors.New("session store down")
	}
	id, ok := f.byToken[hash]
	if !ok {
		return domain.Session{}, domain.ErrNotFound
	}
	return f.sessions[id], nil
}

func (f *fakeStore) RevokeSession(_ context.Context, id, reason string, now time.Time) error {
	if f.failRevoke {
		return errors.New("revoke store down")
	}
	if s, ok := f.sessions[id]; ok && s.RevokedAt == nil {
		s.RevokedAt = &now
		s.RevokeReason = reason
		f.sessions[id] = s
	}
	return nil
}

func (f *fakeStore) RevokeUserSessions(_ context.Context, userID, reason string, now time.Time, keepID string) error {
	for id, s := range f.sessions {
		if s.UserID == userID && id != keepID && s.RevokedAt == nil {
			s.RevokedAt = &now
			s.RevokeReason = reason
			f.sessions[id] = s
		}
	}
	return nil
}

func (f *fakeStore) TouchSession(_ context.Context, id string, lastSeen, idleExpires time.Time) error {
	if s, ok := f.sessions[id]; ok {
		s.LastSeenAt = lastSeen
		s.IdleExpiresAt = idleExpires
		f.sessions[id] = s
	}
	return nil
}

func (f *fakeStore) CreateEmailVerification(_ context.Context, userID, _ string, h [32]byte, exp time.Time) error {
	f.verifs[h] = verifRow{userID: userID, expiresAt: exp}
	return nil
}

func (f *fakeStore) ConsumeEmailVerification(_ context.Context, h [32]byte, now time.Time) (string, error) {
	r, ok := f.verifs[h]
	if !ok || r.consumed || now.After(r.expiresAt) {
		return "", domain.ErrTokenInvalid
	}
	r.consumed = true
	f.verifs[h] = r
	return r.userID, nil
}

func (f *fakeStore) MarkEmailVerified(_ context.Context, userID string, now time.Time) error {
	for id, a := range f.accounts {
		if id == userID {
			a.EmailVerified = true
			f.accounts[id] = a
		}
	}
	_ = now
	return nil
}

func (f *fakeStore) CreatePasswordReset(_ context.Context, userID string, h [32]byte, exp time.Time) error {
	f.resets[h] = verifRow{userID: userID, expiresAt: exp}
	return nil
}

func (f *fakeStore) ConsumePasswordReset(_ context.Context, h [32]byte, now time.Time) (string, error) {
	r, ok := f.resets[h]
	if !ok || r.consumed || now.After(r.expiresAt) {
		return "", domain.ErrTokenInvalid
	}
	r.consumed = true
	f.resets[h] = r
	return r.userID, nil
}

func (f *fakeStore) SetPasswordHash(_ context.Context, userID, phc string, _ time.Time) error {
	if a, ok := f.accounts[userID]; ok {
		a.PasswordHash = phc
		f.accounts[userID] = a
	}
	return nil
}

func (f *fakeStore) Hit(_ context.Context, scope string, key [32]byte, limit int, _ time.Duration, _ time.Time) (bool, time.Duration, error) {
	if f.failHits {
		return false, 0, domain.ErrRateLimited
	}
	k := scope + ":" + base64.RawURLEncoding.EncodeToString(key[:])
	f.hits = append(f.hits, k)
	if f.buckets[k] >= limit {
		return false, time.Minute, nil
	}
	f.buckets[k]++
	return true, 0, nil
}

// newAccount builds a service.NewAccount with a syntactically valid PHC hash
// (the fake store does not verify it; the service under test does not either).
func newAccount(email string) service.NewAccount {
	return service.NewAccount{
		Email: email, Username: "u" + email[:3], DisplayName: "U",
		PasswordHash: "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA==$aGFzaA==",
	}
}

// testStack wires a real service over the fake store plus the handler under
// test. mail is optional (nil skips verification/reset mails).
func testStack(t *testing.T, secure bool) (*Handler, *service.Service, *fakeStore, *mailer.Fake) {
	t.Helper()
	st := newFakeStore()
	mail := mailer.NewFake()
	svc := service.New(st, mail, []byte("0123456789abcdef0123456789abcdef"))
	return New(svc, secure, nil), svc, st, mail
}

// signupAndSignIn runs the two calls every authenticated test needs and
// returns the session cookie plaintext and CSRF token.
func signupAndSignIn(t *testing.T, h *Handler, email string) (sessionPlain, csrf string) {
	t.Helper()
	rec := doJSON(t, h.SignUp, "/api/v1/auth/sign-up", `{"email":"`+email+`","password":"correct-horse-battery","username":"user`+email[:4]+`","display_name":"User"}`)
	if rec.Code != 202 {
		t.Fatalf("sign-up %d %s", rec.Code, rec.Body.String())
	}
	rec = doJSON(t, h.SignIn, "/api/v1/auth/sign-in", `{"email":"`+email+`","password":"correct-horse-battery"}`)
	if rec.Code != 200 {
		t.Fatalf("sign-in %d %s", rec.Code, rec.Body.String())
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName(false) || strings.HasPrefix(c.Name, "__Host-226os_session") {
			sessionPlain = c.Value
		}
	}
	var state struct {
		CsrfToken string `json:"csrf_token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	return sessionPlain, state.CsrfToken
}
