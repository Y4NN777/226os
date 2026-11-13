package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

var errBoom = errors.New("boom")

func withErrs(st *fakeStore, errs map[string]error) {
	st.errs = errs
}

func token32() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(i)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// ---------- ipKey / hashToken / normalizeEmail ----------

func TestIPKeyBuckets(t *testing.T) {
	cases := map[string]string{
		"garbage":               "unknown",
		"":                      "unknown",
		"  203.0.113.77  ":      "203.0.113.0/24",
		"198.51.100.9":          "198.51.100.0/24",
		"::1":                   "::/48",
		"2001:db8:dead:beef::1": "2001:db8:dead::/48",
	}
	for in, want := range cases {
		if got := ipKey(in); got != want {
			t.Fatalf("ipKey(%q)=%q want %q", in, got, want)
		}
	}
}

func TestHashTokenRejectsMalformed(t *testing.T) {
	if _, err := hashToken("!!!not-base64!!!"); err == nil {
		t.Fatal("bad base64 accepted")
	}
	short := base64.RawURLEncoding.EncodeToString([]byte("short"))
	if _, err := hashToken(short); err == nil {
		t.Fatal("wrong-length token accepted")
	}
}

func TestNormalizeEmail(t *testing.T) {
	if normalizeEmail("  MiXeD@Example.COM ") != "mixed@example.com" {
		t.Fatal("normalization wrong")
	}
}

// ---------- SignUp error paths ----------

func TestSignUpStoreAndValidationErrors(t *testing.T) {
	ctx := context.Background()

	// Account lookup internal error -> 500, never a silent 202.
	svc, st, _ := testSvc(t)
	withErrs(st, map[string]error{"AccountByEmail": errBoom})
	if err := svc.SignUp(ctx, SignUpInput{Email: "a@b.co", Password: testPassword, Username: "aa", IP: "10.0.0.1"}); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("lookup error: %v", err)
	}

	// Weak password -> 422 validation, before any write.
	svc, _, _ = testSvc(t)
	if err := svc.SignUp(ctx, SignUpInput{Email: "a@b.co", Password: "short", Username: "aa", IP: "10.0.0.2"}); apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("weak password: %v", err)
	}

	// CreateAccount failures.
	for name, err := range map[string]error{
		"taken":    domain.ErrEmailTaken,
		"username": domain.ErrUsernameTaken,
		"boom":     errBoom,
	} {
		svc, st, _ = testSvc(t)
		st.errs = map[string]error{"CreateAccount": err}
		got := svc.SignUp(ctx, SignUpInput{Email: "c@b.co", Password: testPassword, Username: "cc", IP: "10.0.0.3"})
		if name == "boom" {
			if apperrCode(t, got) != apperr.CodeInternal {
				t.Fatalf("create error: %v", got)
			}
		} else if got != nil {
			t.Fatalf("%s should stay generic 202, got %v", name, got)
		}
	}
}

func TestSignUpNilMailerAndSendFailures(t *testing.T) {
	ctx := context.Background()
	st := newFake()
	svc := New(st, nil, []byte("0123456789abcdef0123456789abcdef"))
	if err := svc.SignUp(ctx, SignUpInput{Email: "nomail@b.co", Password: testPassword, Username: "nm", IP: "10.1.0.1"}); err != nil {
		t.Fatal(err)
	}

	// Mailer error and token-store error are swallowed: response stays 202.
	svc, _, mail := testSvc(t)
	mail.err = errBoom
	if err := svc.SignUp(ctx, SignUpInput{Email: "mailerr@b.co", Password: testPassword, Username: "me", IP: "10.1.0.2"}); err != nil {
		t.Fatal(err)
	}
	svc, st, _ = testSvc(t)
	st.errs = map[string]error{"CreateEmailVerification": errBoom}
	if err := svc.SignUp(ctx, SignUpInput{Email: "tokerr@b.co", Password: testPassword, Username: "te", IP: "10.1.0.3"}); err != nil {
		t.Fatal(err)
	}
	// randToken failure inside sendVerification is swallowed too.
	svc, _, _ = testSvc(t)
	orig := randRead
	randRead = func([]byte) (int, error) { return 0, errBoom }
	t.Cleanup(func() { randRead = orig })
	if err := svc.SignUp(ctx, SignUpInput{Email: "randerr@b.co", Password: testPassword, Username: "re", IP: "10.1.0.4"}); err != nil {
		t.Fatal(err)
	}
}

// ---------- SignIn error paths ----------

func TestSignInInternalErrors(t *testing.T) {
	ctx := context.Background()

	// Lookup internal error (not ErrNotFound).
	svc, st, _ := testSvc(t)
	withErrs(st, map[string]error{"AccountByEmail": errBoom})
	if _, err := svc.SignIn(ctx, SignInInput{Email: "x@y.co", Password: testPassword, IP: "10.2.0.1"}); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("lookup error: %v", err)
	}

	// Suspended account -> same generic invalid credentials.
	svc, st, _ = testSvc(t)
	acct := seedAccount(t, st, "sus@example.com")
	a := st.accounts[acct.ID]
	a.Status = "suspended"
	st.accounts[acct.ID] = a
	if _, err := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword, IP: "10.2.0.2"}); apperrCode(t, err) != apperr.CodeInvalidCredentials {
		t.Fatalf("suspended: %v", err)
	}

	// Corrupt stored hash -> Verify error -> invalid credentials.
	svc, st, _ = testSvc(t)
	acct = seedAccount(t, st, "cor@example.com")
	a = st.accounts[acct.ID]
	a.PasswordHash = "garbage"
	st.accounts[acct.ID] = a
	if _, err := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword, IP: "10.2.0.3"}); apperrCode(t, err) != apperr.CodeInvalidCredentials {
		t.Fatalf("corrupt hash: %v", err)
	}

	// Session write failures -> internal.
	for m, email := range map[string]string{"CreateSession": "csess@example.com", "RevokeSession": "rsess@example.com"} {
		svc, st, _ = testSvc(t)
		acct = seedAccount(t, st, email)
		st.errs = map[string]error{m: errBoom}
		in := SignInInput{Email: acct.Email, Password: testPassword, IP: "10.2.0.4"}
		if m == "RevokeSession" {
			// Need a current session id so rotation path runs.
			res, err := svc.SignIn(ctx, in)
			if err != nil {
				t.Fatal(err)
			}
			st.errs = map[string]error{m: errBoom}
			_, err = svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword,
				IP: "10.2.0.5", CurrentSessionID: viewID(t, svc, ctx, res.Cookie.Value)})
			if apperrCode(t, err) != apperr.CodeInternal {
				t.Fatalf("%s error: %v", m, err)
			}
			continue
		}
		if _, err := svc.SignIn(ctx, in); apperrCode(t, err) != apperr.CodeInternal {
			t.Fatalf("%s error: %v", m, err)
		}
	}

	// Entropy failure in sessionSecrets -> internal.
	svc, st, _ = testSvc(t)
	acct = seedAccount(t, st, "ent@example.com")
	orig := randRead
	randRead = func([]byte) (int, error) { return 0, errBoom }
	t.Cleanup(func() { randRead = orig })
	if _, err := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword, IP: "10.2.0.6"}); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("entropy: %v", err)
	}
}

func viewID(t *testing.T, svc *Service, ctx context.Context, plain string) string {
	t.Helper()
	v, err := svc.ResolveSession(ctx, plain)
	if err != nil {
		t.Fatal(err)
	}
	return v.SessionID
}

// ---------- ResolveSession / SignOut branches ----------

func TestResolveSessionBranches(t *testing.T) {
	ctx := context.Background()
	svc, st, _ := testSvc(t)

	if _, err := svc.ResolveSession(ctx, ""); apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("empty token: %v", err)
	}
	// Well-formed but unknown token.
	if _, err := svc.ResolveSession(ctx, token32()); apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("unknown token: %v", err)
	}
	// Store internal error.
	st.errs = map[string]error{"SessionByTokenHash": errBoom}
	svc2 := New(st, nil, []byte("0123456789abcdef0123456789abcdef"))
	if _, err := svc2.ResolveSession(ctx, token32()); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("store error: %v", err)
	}
	st.errs = nil

	// Session whose account is gone or suspended -> unauthenticated.
	acct := seedAccount(t, st, "gone@example.com")
	res, err := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword, IP: "10.3.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	delete(st.accounts, acct.ID)
	if _, err := svc.ResolveSession(ctx, res.Cookie.Value); apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("deleted account session: %v", err)
	}

	acct2 := seedAccount(t, st, "sus2@example.com")
	res2, _ := svc.SignIn(ctx, SignInInput{Email: acct2.Email, Password: testPassword, IP: "10.3.0.2"})
	a := st.accounts[acct2.ID]
	a.Status = "suspended"
	st.accounts[acct2.ID] = a
	if _, err := svc.ResolveSession(ctx, res2.Cookie.Value); apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("suspended account session: %v", err)
	}

	// Absolute-expired session.
	acct3 := seedAccount(t, st, "exp@example.com")
	res3, _ := svc.SignIn(ctx, SignInInput{Email: acct3.Email, Password: testPassword, IP: "10.3.0.3"})
	h, _ := hashToken(res3.Cookie.Value)
	sess := st.sessions[st.tokenHash[h]]
	sess.AbsoluteExpiresAt = time.Now().Add(-time.Minute)
	st.sessions[sess.ID] = sess
	if _, err := svc.ResolveSession(ctx, res3.Cookie.Value); apperrCode(t, err) != apperr.CodeUnauthenticated {
		t.Fatalf("absolute-expired: %v", err)
	}
}

func TestResolveSessionTouchesAfterFiveMinutes(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	acct := seedAccount(t, st, "touch@example.com")
	res, err := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword, IP: "10.4.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	h, _ := hashToken(res.Cookie.Value)
	sess := st.sessions[st.tokenHash[h]]
	sess.LastSeenAt = time.Now().Add(-10 * time.Minute)
	st.sessions[sess.ID] = sess
	if _, err := svc.ResolveSession(ctx, res.Cookie.Value); err != nil {
		t.Fatal(err)
	}
	if !st.sessions[sess.ID].LastSeenAt.After(sess.LastSeenAt) {
		t.Fatal("last_seen_at not refreshed")
	}
	// TouchSession failure is tolerated (best effort).
	st.errs = map[string]error{"TouchSession": errBoom}
	if _, err := svc.ResolveSession(ctx, res.Cookie.Value); err != nil {
		t.Fatalf("touch failure should be tolerated: %v", err)
	}
	st.errs = nil
	// Fresh session again -> touch skipped, no error.
	if _, err := svc.ResolveSession(ctx, res.Cookie.Value); err != nil {
		t.Fatal(err)
	}
}

func TestSignOutBranches(t *testing.T) {
	ctx := context.Background()
	svc, st, _ := testSvc(t)
	if err := svc.SignOut(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if err := svc.SignOut(ctx, "bad-token"); err != nil {
		t.Fatal(err)
	}
	// Unknown session -> nil.
	if err := svc.SignOut(ctx, token32()); err != nil {
		t.Fatal(err)
	}
	// Revoke failure propagates.
	acct := seedAccount(t, st, "out@example.com")
	res, _ := svc.SignIn(ctx, SignInInput{Email: acct.Email, Password: testPassword, IP: "10.5.0.1"})
	st.errs = map[string]error{"RevokeSession": errBoom}
	if err := svc.SignOut(ctx, res.Cookie.Value); err == nil {
		t.Fatal("revoke error swallowed")
	}
}

// ---------- CSRF service surface ----------

func TestAnonymousCSRFAndPreSession(t *testing.T) {
	svc, _, _ := testSvc(t)
	secret, token, err := svc.AnonymousCSRF()
	if err != nil || secret == "" || token == "" {
		t.Fatalf("AnonymousCSRF %q %q %v", secret, token, err)
	}
	got, ok := svc.PreSessionCSRF(secret)
	if !ok || got != token {
		t.Fatal("pre-session token mismatch")
	}
	if _, ok := svc.PreSessionCSRF("!!!bad!!!"); ok {
		t.Fatal("malformed secret accepted")
	}
	if _, ok := svc.PreSessionCSRF("short"); ok {
		t.Fatal("short secret accepted")
	}

	// Entropy failure.
	orig := randRead
	randRead = func([]byte) (int, error) { return 0, errBoom }
	t.Cleanup(func() { randRead = orig })
	if _, _, err := svc.AnonymousCSRF(); err == nil {
		t.Fatal("entropy failure not reported")
	}
}

func TestCheckCSRFBranches(t *testing.T) {
	svc, _, _ := testSvc(t)
	h := sha256Of([]byte("session"))

	if svc.CheckCSRF(&h, "", "") {
		t.Fatal("empty header accepted")
	}
	if svc.CheckCSRF(nil, "", "sometoken") {
		t.Fatal("no material accepted")
	}
	if svc.CheckCSRF(nil, "!!!bad!!!", "sometoken") {
		t.Fatal("malformed pre-session secret accepted")
	}
	secret, token, err := svc.AnonymousCSRF()
	if err != nil {
		t.Fatal(err)
	}
	if !svc.CheckCSRF(nil, secret, token) {
		t.Fatal("valid pre-session CSRF rejected")
	}
	if svc.CheckCSRF(nil, secret, token+"x") {
		t.Fatal("wrong pre-session CSRF accepted")
	}
	// Session material takes precedence over the pre-session secret.
	if !svc.CheckCSRF(&h, secret, svc.csrfToken(h)) {
		t.Fatal("session CSRF rejected")
	}
	if svc.CheckCSRF(&h, secret, token) {
		t.Fatal("pre-session token accepted for a session")
	}
}

func sha256Of(b []byte) [32]byte { return sha256.Sum256(b) }

// ---------- VerifyEmail / reset error paths ----------

func TestVerifyEmailErrors(t *testing.T) {
	ctx := context.Background()
	svc, st, _ := testSvc(t)
	if err := svc.VerifyEmail(ctx, "bad-token"); apperrCode(t, err) != apperr.CodeInvalidToken {
		t.Fatalf("bad token: %v", err)
	}
	st.errs = map[string]error{"ConsumeEmailVerification": errBoom}
	if err := svc.VerifyEmail(ctx, token32()); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("consume error: %v", err)
	}
	st.errs = nil
	acct := seedAccount(t, st, "ve@example.com")
	plain, hash, _ := randToken()
	_ = st.CreateEmailVerification(ctx, acct.ID, acct.Email, hash, time.Now().Add(time.Hour))
	st.errs = map[string]error{"MarkEmailVerified": errBoom}
	if err := svc.VerifyEmail(ctx, plain); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("mark error: %v", err)
	}
}

func TestRequestPasswordResetRateLimitAndErrors(t *testing.T) {
	ctx := context.Background()
	svc, st, _ := testSvc(t)
	acct := seedAccount(t, st, "rl2@example.com")

	// Exhaust the per-account bucket (3/hour); distinct IPs stay under 10.
	var last error
	for i := 0; i < resetAccountLimit+1; i++ {
		last = svc.RequestPasswordReset(ctx, acct.Email, fmt.Sprintf("10.6.%d.1", i))
	}
	if apperrCode(t, last) != apperr.CodeRateLimited {
		t.Fatalf("expected RATE_LIMITED, got %v", last)
	}

	// Internal store error on reset create -> internal (visible only as 500).
	svc, st, _ = testSvc(t)
	acct = seedAccount(t, st, "rl3@example.com")
	st.errs = map[string]error{"CreatePasswordReset": errBoom}
	if err := svc.RequestPasswordReset(ctx, acct.Email, "10.6.9.1"); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("create reset error: %v", err)
	}
	// randToken failure -> internal.
	svc, st, _ = testSvc(t)
	acct = seedAccount(t, st, "rl4@example.com")
	orig := randRead
	randRead = func([]byte) (int, error) { return 0, errBoom }
	t.Cleanup(func() { randRead = orig })
	if err := svc.RequestPasswordReset(ctx, acct.Email, "10.6.9.2"); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("rand error: %v", err)
	}
}

func TestConfirmPasswordResetErrors(t *testing.T) {
	ctx := context.Background()
	svc, st, _ := testSvc(t)
	acct := seedAccount(t, st, "cp@example.com")
	newPw := "a-brand-new-password-9"

	if err := svc.ConfirmPasswordReset(ctx, "bad-token", newPw); apperrCode(t, err) != apperr.CodeInvalidToken {
		t.Fatalf("bad token: %v", err)
	}
	// Unknown token -> invalid token.
	if err := svc.ConfirmPasswordReset(ctx, token32(), newPw); apperrCode(t, err) != apperr.CodeInvalidToken {
		t.Fatalf("unknown token: %v", err)
	}
	// Internal consume error.
	st.errs = map[string]error{"ConsumePasswordReset": errBoom}
	if err := svc.ConfirmPasswordReset(ctx, token32(), newPw); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("consume error: %v", err)
	}
	st.errs = nil
	// Valid token, weak new password -> validation.
	plain, hash, _ := randToken()
	_ = st.CreatePasswordReset(ctx, acct.ID, hash, time.Now().Add(time.Hour))
	if err := svc.ConfirmPasswordReset(ctx, plain, "short"); apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("weak password: %v", err)
	}
	// Store write failures.
	plain2, hash2, _ := randToken()
	_ = st.CreatePasswordReset(ctx, acct.ID, hash2, time.Now().Add(time.Hour))
	st.errs = map[string]error{"SetPasswordHash": errBoom}
	if err := svc.ConfirmPasswordReset(ctx, plain2, newPw); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("set hash error: %v", err)
	}
	st.errs = map[string]error{"RevokeUserSessions": errBoom}
	plain3, hash3, _ := randToken()
	_ = st.CreatePasswordReset(ctx, acct.ID, hash3, time.Now().Add(time.Hour))
	if err := svc.ConfirmPasswordReset(ctx, plain3, newPw); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("revoke error: %v", err)
	}
}

func TestSignUpRateLimitStoreError(t *testing.T) {
	ctx := context.Background()
	svc, st, _ := testSvc(t)
	st.errs = map[string]error{"Hit": errBoom}
	if err := svc.SignUp(ctx, SignUpInput{Email: "rls@example.com", Password: testPassword,
		Username: "rls", IP: "10.7.0.1"}); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("hit error: %v", err)
	}
}

func TestRequestPasswordResetRateLimitStoreError(t *testing.T) {
	ctx := context.Background()
	svc, st, _ := testSvc(t)
	acct := seedAccount(t, st, "rlip@example.com")
	st.errs = map[string]error{"Hit": errBoom}
	if err := svc.RequestPasswordReset(ctx, acct.Email, "10.8.0.1"); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("hit error: %v", err)
	}
}

func TestRequestPasswordResetInactiveAndNoMailer(t *testing.T) {
	ctx := context.Background()

	// Suspended account: silent 202-equivalent nil, no mail, no token row.
	svc, st, _ := testSvc(t)
	acct := seedAccount(t, st, "susrst@example.com")
	a := st.accounts[acct.ID]
	a.Status = "suspended"
	st.accounts[acct.ID] = a
	if err := svc.RequestPasswordReset(ctx, acct.Email, "10.9.0.1"); err != nil {
		t.Fatal(err)
	}
	if len(st.resets) != 0 {
		t.Fatal("reset token created for suspended account")
	}

	// Nil mailer: same silent nil for an active account.
	st2 := newFake()
	svc2 := New(st2, nil, []byte("0123456789abcdef0123456789abcdef"))
	acct2 := seedAccount(t, st2, "nomailrst@example.com")
	if err := svc2.RequestPasswordReset(ctx, acct2.Email, "10.9.0.2"); err != nil {
		t.Fatal(err)
	}
	if len(st2.resets) != 0 {
		t.Fatal("reset token created without a mailer")
	}
}
