package service

import (
	"bytes"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

const linkBase = "http://127.0.0.1:8088"

func linkSvc(t *testing.T, opts ...Option) (*Service, *fakeStore, *mailer.Fake) {
	t.Helper()
	st, mail := newFake(), mailer.NewFake()
	// The fake store is not goroutine safe: run background work inline.
	opts = append([]Option{WithBackground(func(f func()) { f() })}, opts...)
	return New(st, mail, []byte("0123456789abcdef0123456789abcdef"), opts...), st, mail
}

var tokenInText = regexp.MustCompile(`(?m)^[A-Za-z0-9_-]{43}$`)

// bareCode returns the code that stands alone on its line, or fails.
func bareCode(t *testing.T, text string) string {
	t.Helper()
	got := tokenInText.FindAllString(text, -1)
	if len(got) != 1 {
		t.Fatalf("want exactly one bare code on its own line, got %d in %q", len(got), text)
	}
	return got[0]
}

func lastMail(t *testing.T, mail *mailer.Fake) mailer.Message {
	t.Helper()
	sent := mail.Sent()
	if len(sent) == 0 {
		t.Fatal("no mail sent")
	}
	return sent[len(sent)-1]
}

func TestVerificationMailCarriesALinkWhenABaseURLIsSet(t *testing.T) {
	svc, _, mail := linkSvc(t, WithBaseURL(linkBase))
	if err := svc.SignUp(t.Context(), SignUpInput{Email: "new@example.com", Password: testPassword, Username: "newbie", UIlang: "en", IP: "203.0.113.1"}); err != nil {
		t.Fatal(err)
	}
	text := lastMail(t, mail).Text
	code := bareCode(t, text)
	want := "Confirm your e-mail address:\n\n" + linkBase + "/verifier-email#token=" + code +
		"\n\nOr paste this code:\n" + code + "\n\nThis link expires in 24 hours."
	if text != want {
		t.Fatalf("body:\n%q\nwant:\n%q", text, want)
	}
}

func TestVerificationMailIsTodaysTextWithoutABaseURL(t *testing.T) {
	svc, _, mail := linkSvc(t)
	if err := svc.SignUp(t.Context(), SignUpInput{Email: "new@example.com", Password: testPassword, Username: "newbie", UIlang: "en", IP: "203.0.113.1"}); err != nil {
		t.Fatal(err)
	}
	text := lastMail(t, mail).Text
	code := bareCode(t, text)
	if want := "Confirm your e-mail address:\n\n" + code + "\n\nThis link expires in 24 hours."; text != want {
		t.Fatalf("body:\n%q\nwant:\n%q", text, want)
	}
}

func TestResetMailCarriesALinkWhenABaseURLIsSet(t *testing.T) {
	svc, st, mail := linkSvc(t, WithBaseURL(linkBase))
	acct := seedAccount(t, st, "reset@example.com")
	if err := svc.RequestPasswordReset(t.Context(), acct.Email, "203.0.113.2"); err != nil {
		t.Fatal(err)
	}
	text := lastMail(t, mail).Text
	code := bareCode(t, text)
	want := "Reset your password:\n\n" + linkBase + "/reinitialiser#token=" + code +
		"\n\nOr paste this code:\n" + code + "\n\nThis link expires in 30 minutes."
	if text != want {
		t.Fatalf("body:\n%q\nwant:\n%q", text, want)
	}
}

func TestResetMailIsTodaysTextWithoutABaseURL(t *testing.T) {
	svc, st, mail := linkSvc(t)
	acct := seedAccount(t, st, "reset@example.com")
	if err := svc.RequestPasswordReset(t.Context(), acct.Email, "203.0.113.2"); err != nil {
		t.Fatal(err)
	}
	text := lastMail(t, mail).Text
	code := bareCode(t, text)
	if want := "Reset your password:\n\n" + code + "\n\nThis link expires in 30 minutes."; text != want {
		t.Fatalf("body:\n%q\nwant:\n%q", text, want)
	}
}

// The code in a link mail is the one the store holds: the link works.
func TestLinkTokenIsTheStoredOne(t *testing.T) {
	svc, st, mail := linkSvc(t, WithBaseURL(linkBase))
	if err := svc.SignUp(t.Context(), SignUpInput{Email: "new@example.com", Password: testPassword, Username: "newbie", UIlang: "en", IP: "203.0.113.1"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.VerifyEmail(t.Context(), bareCode(t, lastMail(t, mail).Text)); err != nil {
		t.Fatalf("verify with the mailed code: %v", err)
	}
	acct := seedAccount(t, st, "reset@example.com")
	if err := svc.RequestPasswordReset(t.Context(), acct.Email, "203.0.113.2"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ConfirmPasswordReset(t.Context(), bareCode(t, lastMail(t, mail).Text), "another-long-password", "203.0.113.2"); err != nil {
		t.Fatalf("reset with the mailed code: %v", err)
	}
}

// Neither the token nor the link may reach the logs, on success or on a failed token write or send.
// slog.SetDefault is process-wide, so this test is not parallel.
func TestMailLinksNeverReachTheLogs(t *testing.T) {
	var logs bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	svc, st, mail := linkSvc(t, WithBaseURL(linkBase))
	if err := svc.SignUp(t.Context(), SignUpInput{Email: "new@example.com", Password: testPassword, Username: "newbie", UIlang: "en", IP: "203.0.113.1"}); err != nil {
		t.Fatal(err)
	}
	acct := seedAccount(t, st, "reset@example.com")
	if err := svc.RequestPasswordReset(t.Context(), acct.Email, "203.0.113.2"); err != nil {
		t.Fatal(err)
	}
	var codes []string
	for _, m := range mail.Sent() {
		codes = append(codes, bareCode(t, m.Text))
	}
	// A failing token write logs a warning: it must stay free of the secrets too.
	st.errs = map[string]error{"CreatePasswordReset": errors.New("store down")}
	if err := svc.RequestPasswordReset(t.Context(), acct.Email, "203.0.113.5"); err != nil {
		t.Fatal(err)
	}
	// A failing send is swallowed on purpose.
	mail.FailWith(errors.New("smtp down"))
	st.errs = nil
	if err := svc.RequestPasswordReset(t.Context(), acct.Email, "203.0.113.6"); err != nil {
		t.Fatal(err)
	}
	if len(codes) != 2 {
		t.Fatalf("want two mails, got %d", len(codes))
	}
	out := logs.String()
	for _, secret := range append(codes, "token=", "verifier-email", "reinitialiser", linkBase) {
		if strings.Contains(out, secret) {
			t.Fatalf("logs contain %q:\n%s", secret, out)
		}
	}
	if !strings.Contains(out, "password reset token write failed") {
		t.Fatalf("the failed write should still be logged:\n%s", out)
	}
}
