package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/identity/password"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

// fakeStore is an in-memory Store for unit tests. It is deliberately strict:
// unknown lookups return domain.ErrNotFound like the sqlc repository does.
type fakeStore struct {
	accounts   map[string]domain.Account
	sessions   map[string]domain.Session // by session id
	tokenHash  map[[32]byte]string       // session token hash -> session id
	verif      map[[32]byte]verifRow
	resets     map[[32]byte]resetRow
	verifiedAt map[string]time.Time
	buckets    map[string]int
	hits       []string
	errs       map[string]error // method name -> forced error
}

type verifRow struct {
	userID, email string
	expiresAt     time.Time
	consumed      bool
}
type resetRow struct {
	userID    string
	expiresAt time.Time
	consumed  bool
}

func newFake() *fakeStore {
	return &fakeStore{
		accounts: map[string]domain.Account{}, sessions: map[string]domain.Session{},
		tokenHash: map[[32]byte]string{}, verif: map[[32]byte]verifRow{},
		resets: map[[32]byte]resetRow{}, verifiedAt: map[string]time.Time{},
		buckets: map[string]int{},
	}
}

func (f *fakeStore) CreateAccount(_ context.Context, in NewAccount) (domain.Account, error) {
	if err := f.errs["CreateAccount"]; err != nil {
		return domain.Account{}, err
	}
	for _, a := range f.accounts {
		if a.Email == in.Email {
			return domain.Account{}, domain.ErrEmailTaken
		}
	}
	a := domain.Account{ID: "u-" + in.Username, Email: in.Email, Username: in.Username,
		DisplayName: in.DisplayName, PasswordHash: in.PasswordHash, Role: "member",
		Status: "active", CreatedAt: time.Now()}
	f.accounts[a.ID] = a
	return a, nil
}
func (f *fakeStore) AccountByEmail(_ context.Context, email string) (domain.Account, error) {
	if err := f.errs["AccountByEmail"]; err != nil {
		return domain.Account{}, err
	}
	for _, a := range f.accounts {
		if a.Email == email {
			return a, nil
		}
	}
	return domain.Account{}, domain.ErrNotFound
}
func (f *fakeStore) AccountByID(_ context.Context, id string) (domain.Account, error) {
	if err := f.errs["AccountByID"]; err != nil {
		return domain.Account{}, err
	}
	if a, ok := f.accounts[id]; ok {
		return a, nil
	}
	return domain.Account{}, domain.ErrNotFound
}
func (f *fakeStore) CreateSession(_ context.Context, s domain.Session) (string, error) {
	if err := f.errs["CreateSession"]; err != nil {
		return "", err
	}
	f.sessions[s.ID] = s
	f.tokenHash[s.TokenHash] = s.ID
	return s.ID, nil
}
func (f *fakeStore) SessionByTokenHash(_ context.Context, h [32]byte) (domain.Session, error) {
	if err := f.errs["SessionByTokenHash"]; err != nil {
		return domain.Session{}, err
	}
	if id, ok := f.tokenHash[h]; ok {
		return f.sessions[id], nil
	}
	return domain.Session{}, domain.ErrNotFound
}
func (f *fakeStore) RevokeSession(_ context.Context, id, reason string, now time.Time) error {
	if err := f.errs["RevokeSession"]; err != nil {
		return err
	}
	s := f.sessions[id]
	s.RevokedAt = &now
	s.RevokeReason = reason
	f.sessions[id] = s
	return nil
}
func (f *fakeStore) RevokeUserSessions(_ context.Context, userID, reason string, now time.Time, keepID string) error {
	if err := f.errs["RevokeUserSessions"]; err != nil {
		return err
	}
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
	if err := f.errs["TouchSession"]; err != nil {
		return err
	}
	s := f.sessions[id]
	s.LastSeenAt = lastSeen
	s.IdleExpiresAt = idleExpires
	f.sessions[id] = s
	return nil
}
func (f *fakeStore) CreateEmailVerification(_ context.Context, userID, email string, h [32]byte, exp time.Time) error {
	if err := f.errs["CreateEmailVerification"]; err != nil {
		return err
	}
	f.verif[h] = verifRow{userID: userID, email: email, expiresAt: exp}
	return nil
}
func (f *fakeStore) ConsumeEmailVerification(_ context.Context, h [32]byte, now time.Time) (string, error) {
	if err := f.errs["ConsumeEmailVerification"]; err != nil {
		return "", err
	}
	r, ok := f.verif[h]
	if !ok || r.consumed || now.After(r.expiresAt) {
		return "", domain.ErrTokenInvalid
	}
	r.consumed = true
	f.verif[h] = r
	return r.userID, nil
}
func (f *fakeStore) MarkEmailVerified(_ context.Context, userID string, now time.Time) error {
	if err := f.errs["MarkEmailVerified"]; err != nil {
		return err
	}
	f.verifiedAt[userID] = now
	for id, a := range f.accounts {
		if id == userID {
			a.EmailVerified = true
			f.accounts[id] = a
		}
	}
	return nil
}
func (f *fakeStore) CreatePasswordReset(_ context.Context, userID string, h [32]byte, exp time.Time) error {
	if err := f.errs["CreatePasswordReset"]; err != nil {
		return err
	}
	f.resets[h] = resetRow{userID: userID, expiresAt: exp}
	return nil
}
func (f *fakeStore) ConsumePasswordReset(_ context.Context, h [32]byte, now time.Time) (string, error) {
	if err := f.errs["ConsumePasswordReset"]; err != nil {
		return "", err
	}
	r, ok := f.resets[h]
	if !ok || r.consumed || now.After(r.expiresAt) {
		return "", domain.ErrTokenInvalid
	}
	r.consumed = true
	f.resets[h] = r
	return r.userID, nil
}
func (f *fakeStore) SetPasswordHash(_ context.Context, userID, phc string, _ time.Time) error {
	if err := f.errs["SetPasswordHash"]; err != nil {
		return err
	}
	for id, a := range f.accounts {
		if id == userID {
			a.PasswordHash = phc
			f.accounts[id] = a
		}
	}
	return nil
}
func (f *fakeStore) Hit(_ context.Context, scope string, key [32]byte, limit int, _ time.Duration, _ time.Time) (bool, time.Duration, error) {
	if err := f.errs["Hit"]; err != nil {
		return false, 0, err
	}
	k := scope + ":" + base64.RawURLEncoding.EncodeToString(key[:8])
	f.hits = append(f.hits, k)
	if f.buckets[k] >= limit {
		return false, time.Minute, nil
	}
	f.buckets[k]++
	return true, 0, nil
}

type fakeMail struct {
	sent []string
	err  error
}

func (m *fakeMail) Send(_ context.Context, msg mailer.Message) error {
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, msg.To...)
	return nil
}

func testSvc(t *testing.T) (*Service, *fakeStore, *fakeMail) {
	t.Helper()
	st, mail := newFake(), &fakeMail{}
	return New(st, mail, []byte("0123456789abcdef0123456789abcdef")), st, mail
}

const testPassword = "correct-horse-battery"

func mustHash(t *testing.T, pw string) string {
	t.Helper()
	h, err := password.Hash(pw)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func seedAccount(t *testing.T, st *fakeStore, email string) domain.Account {
	t.Helper()
	a, err := st.CreateAccount(context.Background(), NewAccount{
		Email: email, Username: "u" + email[:4], DisplayName: "U",
		PasswordHash: mustHash(t, testPassword),
	})
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func apperrCode(t *testing.T, err error) apperr.Code {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) {
		t.Fatalf("not an apperr: %v", err)
	}
	return ae.Code
}

func TestSignUpIsGenericAndSendsVerification(t *testing.T) {
	svc, st, mail := testSvc(t)
	ctx := context.Background()
	if err := svc.SignUp(ctx, SignUpInput{Email: "New@Example.com", Password: testPassword,
		Username: "newbie", DisplayName: "New", IP: "203.0.113.7"}); err != nil {
		t.Fatal(err)
	}
	a, err := st.AccountByEmail(ctx, "new@example.com") // normalized
	if err != nil || a.Email != "new@example.com" {
		t.Fatalf("account %+v err %v", a, err)
	}
	if len(mail.sent) != 1 || mail.sent[0] != "new@example.com" {
		t.Fatalf("mail %v", mail.sent)
	}
	// Second sign-up with the same email: same nil result, no second mail.
	if err := svc.SignUp(ctx, SignUpInput{Email: "new@example.com", Password: testPassword,
		Username: "newbie2", IP: "203.0.113.8"}); err != nil {
		t.Fatal(err)
	}
	if len(mail.sent) != 1 {
		t.Fatal("duplicate sign-up sent another mail")
	}
}

func TestSignInRotatesSessionAndRejectsBadCredentials(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	acct := seedAccount(t, st, "sam@example.com")

	// Unknown email and wrong password share one error code.
	_, errUnknown := svc.SignIn(ctx, SignInInput{Email: "ghost@example.com", Password: "whatever-password", IP: "198.51.100.1"})
	_, errWrong := svc.SignIn(ctx, SignInInput{Email: "sam@example.com", Password: "wrong-password-xx", IP: "198.51.100.2"})
	if apperrCode(t, errUnknown) != apperr.CodeInvalidCredentials || apperrCode(t, errWrong) != apperr.CodeInvalidCredentials {
		t.Fatal("credentials errors differ")
	}

	res, err := svc.SignIn(ctx, SignInInput{Email: "sam@example.com", Password: testPassword, IP: "198.51.100.3"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Cookie.Value == "" || res.Cookie.CSRF == "" || res.Cookie.MaxAge != int(absoluteTTL.Seconds()) {
		t.Fatalf("cookie %+v", res.Cookie)
	}
	view, err := svc.ResolveSession(ctx, res.Cookie.Value)
	if err != nil || view.Account.ID != acct.ID || view.CSRF != res.Cookie.CSRF {
		t.Fatalf("view %+v err %v", view, err)
	}

	// Sign in again carrying the first session: it is rotated away.
	res2, err := svc.SignIn(ctx, SignInInput{Email: "sam@example.com", Password: testPassword,
		IP: "198.51.100.4", CurrentSessionID: view.SessionID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveSession(ctx, res.Cookie.Value); apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("old session still valid: %v", err)
	}
	if _, err := svc.ResolveSession(ctx, res2.Cookie.Value); err != nil {
		t.Fatalf("new session invalid: %v", err)
	}
	// A second device session is untouched by the first device's rotation.
	res3, err := svc.SignIn(ctx, SignInInput{Email: "sam@example.com", Password: testPassword, IP: "198.51.100.5"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveSession(ctx, res3.Cookie.Value); err != nil {
		t.Fatal(err)
	}
}

func TestResolveSessionRejectsRevokedExpiredAndUnknown(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	seedAccount(t, st, "rev@example.com")

	res, err := svc.SignIn(ctx, SignInInput{Email: "rev@example.com", Password: testPassword, IP: "192.0.2.1"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveSession(ctx, "not-a-token"); apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("garbage token: %v", err)
	}
	if err := svc.SignOut(ctx, res.Cookie.Value); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ResolveSession(ctx, res.Cookie.Value); apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("signed-out session still valid: %v", err)
	}

	// Expired idle session.
	res2, err := svc.SignIn(ctx, SignInInput{Email: "rev@example.com", Password: testPassword, IP: "192.0.2.2"})
	if err != nil {
		t.Fatal(err)
	}
	h, _ := hashToken(res2.Cookie.Value)
	sess := st.sessions[st.tokenHash[h]]
	sess.IdleExpiresAt = time.Now().Add(-time.Minute)
	st.sessions[sess.ID] = sess
	if _, err := svc.ResolveSession(ctx, res2.Cookie.Value); apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("expired session valid: %v", err)
	}
}

func TestVerifyEmailSingleUse(t *testing.T) {
	svc, st, mail := testSvc(t)
	ctx := context.Background()
	acct := seedAccount(t, st, "ver@example.com")
	// Simulate the verification mail content: the token is the mail body tail.
	// Instead, drive the store directly through SignUp's send path is complex;
	// create the token row the same way the service does.
	plain, hash, err := randToken()
	if err != nil {
		t.Fatal(err)
	}
	if err := st.CreateEmailVerification(ctx, acct.ID, acct.Email, hash, time.Now().Add(verifyTTL)); err != nil {
		t.Fatal(err)
	}
	_ = mail
	if err := svc.VerifyEmail(ctx, plain); err != nil {
		t.Fatal(err)
	}
	if err := svc.VerifyEmail(ctx, plain); apperrCode(t, err) != apperr.CodeInvalidToken {
		t.Fatalf("replay accepted: %v", err)
	}
	a, _ := st.AccountByID(ctx, acct.ID)
	if !a.EmailVerified {
		t.Fatal("email not verified")
	}
	// Expired token.
	plain2, hash2, _ := randToken()
	_ = st.CreateEmailVerification(ctx, acct.ID, acct.Email, hash2, time.Now().Add(-time.Minute))
	if err := svc.VerifyEmail(ctx, plain2); apperrCode(t, err) != apperr.CodeInvalidToken {
		t.Fatalf("expired token accepted: %v", err)
	}
}

func TestPasswordResetRevokesAllSessions(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	acct := seedAccount(t, st, "rst@example.com")
	s1, _ := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword, IP: "203.0.113.1"})
	s2, _ := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword, IP: "203.0.113.2"})

	// Request is generic; the token row is what matters.
	if err := svc.RequestPasswordReset(ctx, acct.Email, "203.0.113.3"); err != nil {
		t.Fatal(err)
	}
	// Unknown email: still nil.
	if err := svc.RequestPasswordReset(ctx, "ghost@example.com", "203.0.113.4"); err != nil {
		t.Fatal(err)
	}
	// Find the stored reset hash and mint the matching plaintext is impossible
	// (hash only); drive ConfirmPasswordReset through a fresh token row.
	plain, hash, _ := randToken()
	_ = st.CreatePasswordReset(ctx, acct.ID, hash, time.Now().Add(resetTTL))
	newPw := "a-brand-new-password-1"
	if err := svc.ConfirmPasswordReset(ctx, plain, newPw); err != nil {
		t.Fatal(err)
	}
	for _, v := range []string{s1.Cookie.Value, s2.Cookie.Value} {
		if _, err := svc.ResolveSession(ctx, v); apperrCode(t, err) != apperr.CodeUnauthenticated {
			t.Fatal("session survived password reset")
		}
	}
	// Old password dead, new password works.
	if _, err := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword, IP: "203.0.113.5"}); apperrCode(t, err) != apperr.CodeInvalidCredentials {
		t.Fatal("old password still works")
	}
	if _, err := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: newPw, IP: "203.0.113.6"}); err != nil {
		t.Fatalf("new password rejected: %v", err)
	}
}

func TestRateLimitFailsClosed(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	seedAccount(t, st, "rl@example.com")
	// Exhaust the per-IP sign-in bucket (20/15min) with wrong passwords.
	var last error
	for i := 0; i < signInIPLimit+2; i++ {
		_, last = svc.SignIn(ctx, SignInInput{Email: "rl@example.com", Password: "nope-nope-nope",
			IP: "198.51.100.99"})
	}
	if apperrCode(t, last) != apperr.CodeRateLimited {
		t.Fatalf("expected RATE_LIMITED, got %v", last)
	}
	// Store error inside Hit must surface as internal, never as allowed.
	st.buckets["x"] = -1 // force the map read path; simulate failure via a broken store instead
	broken := &brokenStore{fakeStore: st}
	svc2 := New(broken, nil, []byte("0123456789abcdef0123456789abcdef"))
	_, err := svc2.SignIn(ctx, SignInInput{Email: "rl@example.com", Password: "x", IP: "198.51.100.100"})
	if apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("store failure not fail-closed: %v", err)
	}
}

type brokenStore struct{ *fakeStore }

func (b *brokenStore) Hit(context.Context, string, [32]byte, int, time.Duration, time.Time) (bool, time.Duration, error) {
	return false, 0, errors.New("db down")
}

func TestCSRFDerivationIsSessionBound(t *testing.T) {
	svc, _, _ := testSvc(t)
	a := sha256.Sum256([]byte("session-a"))
	b := sha256.Sum256([]byte("session-b"))
	if svc.csrfToken(a) == svc.csrfToken(b) {
		t.Fatal("CSRF tokens collide across sessions")
	}
	if !svc.CheckCSRF(&a, "", svc.csrfToken(a)) {
		t.Fatal("valid session CSRF rejected")
	}
	if svc.CheckCSRF(&a, "", svc.csrfToken(b)) {
		t.Fatal("foreign session CSRF accepted")
	}
	if svc.CheckCSRF(nil, "", svc.csrfToken(a)) {
		t.Fatal("session token accepted without a session")
	}
}
