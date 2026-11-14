package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log/slog"
	"strings"
	"sync"
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
	lastWrite  domain.ProfileWrite // the last patch handed to UpdateProfile
	accounts   map[string]domain.Account
	sessions   map[string]domain.Session // by session id
	tokenHash  map[[32]byte]string       // session token hash -> session id
	verif      map[[32]byte]verifRow
	resets     map[[32]byte]resetRow
	verifiedAt map[string]time.Time
	buckets    map[string]int
	hits       []string
	errs       map[string]error                // method name -> forced error
	profiles   map[string]domain.ProfileRecord // by username
	tagIDs     map[string]string               // skill slug -> tag id (active only)
	directory  []domain.DirectoryEntry         // rows ListProfiles serves, in order
	lastDir    domain.DirectoryQuery           // the last ListProfiles query
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
		buckets: map[string]int{}, profiles: map[string]domain.ProfileRecord{},
		tagIDs: map[string]string{},
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
func (f *fakeStore) VerifyEmail(_ context.Context, h [32]byte, now time.Time) error {
	if err := f.errs["VerifyEmail"]; err != nil {
		return err
	}
	r, ok := f.verif[h]
	if !ok || r.consumed || now.After(r.expiresAt) {
		return domain.ErrTokenInvalid
	}
	r.consumed = true
	f.verif[h] = r
	f.verifiedAt[r.userID] = now
	for id, a := range f.accounts {
		if id == r.userID {
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
func (f *fakeStore) ResetPassword(_ context.Context, h [32]byte, phc string, now time.Time) error {
	if err := f.errs["ResetPassword"]; err != nil {
		return err
	}
	r, ok := f.resets[h]
	if !ok || r.consumed || now.After(r.expiresAt) {
		return domain.ErrTokenInvalid
	}
	r.consumed = true
	f.resets[h] = r
	for id, a := range f.accounts {
		if id == r.userID {
			a.PasswordHash = phc
			f.accounts[id] = a
		}
	}
	for id, s := range f.sessions {
		if s.UserID == r.userID && s.RevokedAt == nil {
			s.RevokedAt = &now
			s.RevokeReason = "password_reset"
			f.sessions[id] = s
		}
	}
	return nil
}
func (f *fakeStore) GiveBack(_ context.Context, scope string, key [32]byte, _ time.Duration, _ time.Time) error {
	if err := f.errs["GiveBack"]; err != nil {
		return err
	}
	k := scope + ":" + base64.RawURLEncoding.EncodeToString(key[:8])
	if f.buckets[k] > 0 {
		f.buckets[k]--
	}
	return nil
}

func (f *fakeStore) Hit(_ context.Context, scope string, key [32]byte, limit int, _ time.Duration, _ time.Time) (bool, time.Duration, error) {
	if err := f.errs["Hit"]; err != nil {
		return false, 0, err
	}
	if err := f.errs["Hit:"+scope]; err != nil {
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

// ---------- profiles ----------

func (f *fakeStore) ListProfiles(_ context.Context, q domain.DirectoryQuery) ([]domain.DirectoryEntry, error) {
	f.lastDir = q
	if err := f.errs["ListProfiles"]; err != nil {
		return nil, err
	}
	var out []domain.DirectoryEntry
	for _, e := range f.directory {
		if q.AfterUsername != "" && e.Username <= q.AfterUsername {
			continue
		}
		if len(out) == q.Limit {
			break
		}
		out = append(out, e)
	}
	return out, nil
}

func (f *fakeStore) ProfileByUserID(_ context.Context, userID string) (domain.Profile, error) {
	if err := f.errs["ProfileByUserID"]; err != nil {
		return domain.Profile{}, err
	}
	for _, rec := range f.profiles {
		if rec.UserID == userID {
			return rec.Profile, nil
		}
	}
	return domain.Profile{}, domain.ErrNotFound
}

func (f *fakeStore) ProfileByUsername(_ context.Context, username string) (domain.ProfileRecord, error) {
	if err := f.errs["ProfileByUsername"]; err != nil {
		return domain.ProfileRecord{}, err
	}
	rec, ok := f.profiles[username]
	if !ok {
		return domain.ProfileRecord{}, domain.ErrNotFound
	}
	return rec, nil
}

func (f *fakeStore) ResolveSkillSlugs(_ context.Context, slugs []string) ([]string, error) {
	if err := f.errs["ResolveSkillSlugs"]; err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(slugs))
	for _, s := range slugs {
		if id, ok := f.tagIDs[s]; ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// applyPatch mirrors the SQL CASE WHEN semantics on the fake's Profile copy.
func (f *fakeStore) UpdateProfile(_ context.Context, userID string, w domain.ProfileWrite) error {
	f.lastWrite = w
	if err := f.errs["UpdateProfile"]; err != nil {
		return err
	}
	for name, rec := range f.profiles {
		if rec.UserID != userID {
			continue
		}
		p := &rec.Profile
		setStr := func(dst **string, f2 domain.PatchField[string]) {
			if !f2.Present {
				return
			}
			*dst = f2.Value
		}
		if w.DisplayName.Present && w.DisplayName.Value != nil {
			p.DisplayName = *w.DisplayName.Value
		}
		setStr(&p.Intent, w.Intent)
		setStr(&p.CountryCode, w.CountryCode)
		if w.Languages.Present {
			if w.Languages.Value == nil {
				p.Languages = []string{}
			} else {
				p.Languages = *w.Languages.Value
			}
		}
		setStr(&p.GitHub, w.GitHub)
		setStr(&p.Bio, w.Bio)
		if w.OpenToCollab.Present && w.OpenToCollab.Value != nil {
			p.OpenToCollab = *w.OpenToCollab.Value
		}
		setVis := func(dst *domain.Visibility, f2 domain.PatchField[domain.Visibility]) {
			if f2.Present && f2.Value != nil {
				*dst = *f2.Value
			}
		}
		setVis(&p.VisCountry, w.VisCountry)
		setVis(&p.VisLanguages, w.VisLanguages)
		setVis(&p.VisSkills, w.VisSkills)
		setVis(&p.VisGitHub, w.VisGitHub)
		setVis(&p.VisBio, w.VisBio)
		setVis(&p.Visibility, w.Visibility)
		if w.SkillTagIDs.Present {
			if w.SkillTagIDs.Value == nil || *w.SkillTagIDs.Value == nil {
				p.Skills = []string{}
			} else {
				// The service hands resolved tag ids; the fake maps them back
				// to slugs because Profile.Skills exposes slugs.
				slugs := make([]string, 0, len(*w.SkillTagIDs.Value))
				for _, id := range *w.SkillTagIDs.Value {
					for slug, tid := range f.tagIDs {
						if tid == id {
							slugs = append(slugs, slug)
						}
					}
				}
				p.Skills = slugs
			}
		}
		rec.Profile = *p
		f.profiles[name] = rec
		return nil
	}
	return domain.ErrNotFound
}

type fakeMail struct {
	mu   sync.Mutex
	sent []string
	err  error
}

func (m *fakeMail) Send(_ context.Context, msg mailer.Message) error {
	if m.err != nil {
		return m.err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, msg.To...)
	return nil
}

// waitSent polls until at least n sends landed (mail goes out async).
func (m *fakeMail) waitSent(t *testing.T, n int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		m.mu.Lock()
		got := len(m.sent)
		m.mu.Unlock()
		if got >= n {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected %d mails, got none in time", n)
}

func testSvc(t *testing.T) (*Service, *fakeStore, *fakeMail) {
	t.Helper()
	st, mail := newFake(), &fakeMail{}
	// The fake store is not goroutine safe: run background work inline.
	return New(st, mail, []byte("0123456789abcdef0123456789abcdef"), WithBackground(func(f func()) { f() })), st, mail
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

// wantViolation asserts err carries exactly one violation on field with a contract enum code.
func wantViolation(t *testing.T, err error, field, code string) {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) || len(ae.Violations) != 1 {
		t.Fatalf("err %v, want one violation", err)
	}
	if ae.Violations[0].Field != field || ae.Violations[0].Code != code {
		t.Fatalf("violation %+v, want %s/%s", ae.Violations[0], field, code)
	}
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
	mail.waitSent(t, 1)
	mail.mu.Lock()
	if len(mail.sent) != 1 || mail.sent[0] != "new@example.com" {
		t.Fatalf("mail %v", mail.sent)
	}
	mail.mu.Unlock()
	// Second sign-up with the same email: same nil result, no second mail.
	if err := svc.SignUp(ctx, SignUpInput{Email: "new@example.com", Password: testPassword,
		Username: "newbie2", IP: "203.0.113.8"}); err != nil {
		t.Fatal(err)
	}
	// Give a hypothetical stray async send a moment to land before asserting.
	time.Sleep(50 * time.Millisecond)
	mail.mu.Lock()
	defer mail.mu.Unlock()
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

func TestSignInAccountBucketAppliesToUnknownEmails(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	// The per-account bucket is hit before the lookup, so an unknown email
	// exhausts it exactly like a known one: no 429-vs-401 enumeration oracle.
	var last error
	for i := 0; i < signInAccountLimit+1; i++ {
		_, last = svc.SignIn(ctx, SignInInput{Email: "ghost@example.com", Password: "whatever-password",
			IP: "198.51.100.50"})
	}
	if apperrCode(t, last) != apperr.CodeRateLimited {
		t.Fatalf("expected RATE_LIMITED, got %v", last)
	}
	found := false
	for _, h := range st.hits {
		if strings.HasPrefix(h, "signin:account:") {
			found = true
		}
	}
	if !found {
		t.Fatal("signin:account bucket never hit for an unknown email")
	}
}

// ADR-0003: "5 failures per account per 15 min". Successful sign-ins do not
// count, so a person (or a demo account) signing in from several devices is
// never throttled by the account bucket.
func TestSignInSuccessDoesNotCountAgainstTheAccountLimit(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	seedAccount(t, st, "sam@example.com")
	for i := 0; i < signInAccountLimit+3; i++ {
		if _, err := svc.SignIn(ctx, SignInInput{Email: "sam@example.com", Password: testPassword, IP: "198.51.100.5"}); err != nil {
			t.Fatalf("sign-in %d: %v", i+1, err)
		}
	}
}

func TestSignInFailuresBlockEvenTheRightPassword(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	seedAccount(t, st, "sam@example.com")
	for i := 0; i < signInAccountLimit; i++ {
		_, err := svc.SignIn(ctx, SignInInput{Email: "sam@example.com", Password: "wrong-password-12", IP: "198.51.100.6"})
		if apperrCode(t, err) != apperr.CodeInvalidCredentials {
			t.Fatalf("failure %d: %v", i+1, err)
		}
	}
	_, err := svc.SignIn(ctx, SignInInput{Email: "sam@example.com", Password: testPassword, IP: "198.51.100.6"})
	if apperrCode(t, err) != apperr.CodeRateLimited {
		t.Fatalf("expected RATE_LIMITED after %d failures, got %v", signInAccountLimit, err)
	}
}

func TestSignInRateLimitStoreErrorsFailClosed(t *testing.T) {
	for _, method := range []string{"Hit"} {
		svc, st, _ := testSvc(t)
		seedAccount(t, st, "sam@example.com")
		withErrs(st, map[string]error{method: errBoom})
		_, err := svc.SignIn(context.Background(), SignInInput{Email: "sam@example.com", Password: testPassword, IP: "198.51.100.7"})
		if apperrCode(t, err) != apperr.CodeInternal {
			t.Fatalf("%s error did not fail closed: %v", method, err)
		}
	}
	// The failure counter is written on the failure path too.
	svc, st, _ := testSvc(t)
	seedAccount(t, st, "sam@example.com")
	withErrs(st, map[string]error{"Hit:signin:account": errBoom})
	if _, err := svc.SignIn(context.Background(), SignInInput{Email: "ghost@example.com", Password: "x-wrong-password", IP: "198.51.100.8"}); apperrCode(t, err) != apperr.CodeInternal {
		t.Fatalf("unknown email: %v", err)
	}
}

func TestSignUpExistingEmailRunsDummyVerify(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	seedAccount(t, st, "dup@example.com")
	calls := 0
	orig := verifyDummy
	verifyDummy = func(string) { calls++ }
	t.Cleanup(func() { verifyDummy = orig })
	if err := svc.SignUp(ctx, SignUpInput{Email: "dup@example.com", Password: testPassword,
		Username: "other", IP: "203.0.113.20"}); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("dummy verify calls = %d, want 1 (timing equalization)", calls)
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
	if err := svc.ConfirmPasswordReset(ctx, plain, newPw, "10.20.0.1"); err != nil {
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

func TestConfirmPasswordResetWeakPasswordKeepsToken(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	acct := seedAccount(t, st, "weak@example.com")
	plain, hash, _ := randToken()
	_ = st.CreatePasswordReset(ctx, acct.ID, hash, time.Now().Add(resetTTL))
	// A weak password is rejected before the token is touched.
	err := svc.ConfirmPasswordReset(ctx, plain, "short", "10.20.0.1")
	if apperrCode(t, err) != apperr.CodeValidation {
		t.Fatalf("weak password: %v", err)
	}
	wantViolation(t, err, "/new_password", "invalid_value")
	// The same token must still work with a compliant password.
	if err := svc.ConfirmPasswordReset(ctx, plain, "a-brand-new-password-1", "10.20.0.1"); err != nil {
		t.Fatalf("token burned by the weak-password attempt: %v", err)
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

// A give-back failure must not turn an authenticated sign-in into an error: the
// hit simply stays counted (strict side) and a warning is logged without the
// e-mail address.
func TestSignInGiveBackFailureDoesNotFailTheSignIn(t *testing.T) {
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	svc, st, _ := testSvc(t)
	seedAccount(t, st, "sam@example.com")
	withErrs(st, map[string]error{"GiveBack": errBoom})
	if _, err := svc.SignIn(context.Background(), SignInInput{Email: "sam@example.com", Password: testPassword, IP: "198.51.100.9"}); err != nil {
		t.Fatalf("give-back failure failed the sign-in: %v", err)
	}
	out := logs.String()
	if !strings.Contains(out, "give back failed") {
		t.Fatalf("give-back failure not logged: %q", out)
	}
	if strings.Contains(out, "sam@example.com") {
		t.Fatalf("e-mail leaked into the log: %q", out)
	}
	// The hit stayed counted: only 4 more failures fit before the 5-hit limit.
	var last error
	for i := 0; i < signInAccountLimit; i++ {
		_, last = svc.SignIn(context.Background(), SignInInput{Email: "sam@example.com", Password: "wrong-password-12", IP: "198.51.100.9"})
	}
	if apperrCode(t, last) != apperr.CodeRateLimited {
		t.Fatalf("expected the retained hit to make the 5th failure RATE_LIMITED, got %v", last)
	}
}

// The give-back keeps ADR-0003's "5 failures" meaning: 4 failures, a success,
// then exactly one more failure is still allowed before the block.
func TestSignInSuccessGivesBackItsHit(t *testing.T) {
	svc, st, _ := testSvc(t)
	ctx := context.Background()
	seedAccount(t, st, "sam@example.com")
	in := func(pw string) error {
		_, err := svc.SignIn(ctx, SignInInput{Email: "sam@example.com", Password: pw, IP: "198.51.100.10"})
		return err
	}
	for i := 0; i < signInAccountLimit-1; i++ {
		if apperrCode(t, in("wrong-password-12")) != apperr.CodeInvalidCredentials {
			t.Fatalf("failure %d", i+1)
		}
	}
	if err := in(testPassword); err != nil {
		t.Fatal(err)
	}
	if apperrCode(t, in("wrong-password-12")) != apperr.CodeInvalidCredentials {
		t.Fatal("5th failure must still reach the password check")
	}
	if apperrCode(t, in("wrong-password-12")) != apperr.CodeRateLimited {
		t.Fatal("6th attempt must be rate limited")
	}
}
