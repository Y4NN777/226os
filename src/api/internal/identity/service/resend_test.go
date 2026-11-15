package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

// A member waiting for verification gets a new code by mail, and the new code works.
func TestResendEmailVerificationMailsANewWorkingCode(t *testing.T) {
	svc, st, mail := linkSvc(t)
	acct := seedAccount(t, st, "wait@example.com")
	if err := svc.ResendEmailVerification(t.Context(), " Wait@Example.com ", "10.20.0.1"); err != nil {
		t.Fatal(err)
	}
	msg := lastMail(t, mail)
	if len(msg.To) != 1 || msg.To[0] != acct.Email {
		t.Fatalf("mail to %v, want %s", msg.To, acct.Email)
	}
	if err := svc.VerifyEmail(t.Context(), bareCode(t, msg.Text)); err != nil {
		t.Fatalf("the resent code does not verify: %v", err)
	}
	if a, _ := st.AccountByID(t.Context(), acct.ID); !a.EmailVerified {
		t.Fatal("account not verified")
	}
}

// The new code replaces the earlier one: an older mail can no longer verify the account.
func TestResendEmailVerificationInvalidatesTheEarlierCode(t *testing.T) {
	svc, st, mail := linkSvc(t)
	seedAccount(t, st, "twice@example.com")
	if err := svc.ResendEmailVerification(t.Context(), "twice@example.com", "10.20.1.1"); err != nil {
		t.Fatal(err)
	}
	first := bareCode(t, lastMail(t, mail).Text)
	if err := svc.ResendEmailVerification(t.Context(), "twice@example.com", "10.20.1.1"); err != nil {
		t.Fatal(err)
	}
	second := bareCode(t, lastMail(t, mail).Text)
	if err := svc.VerifyEmail(t.Context(), first); apperrCode(t, err) != apperr.CodeInvalidToken {
		t.Fatalf("the earlier code still works: %v", err)
	}
	if err := svc.VerifyEmail(t.Context(), second); err != nil {
		t.Fatalf("the new code does not verify: %v", err)
	}
}

// Unknown, already verified and inactive accounts get the same silent success and no mail (no enumeration).
func TestResendEmailVerificationIsSilentWhenNothingIsDue(t *testing.T) {
	svc, st, mail := linkSvc(t)
	verified := seedAccount(t, st, "done@example.com")
	a := st.accounts[verified.ID]
	a.EmailVerified = true
	st.accounts[verified.ID] = a
	suspended := seedAccount(t, st, "gone@example.com")
	a = st.accounts[suspended.ID]
	a.Status = "suspended"
	st.accounts[suspended.ID] = a

	for i, email := range []string{"ghost@example.com", verified.Email, suspended.Email} {
		if err := svc.ResendEmailVerification(t.Context(), email, fmt.Sprintf("10.20.2.%d", i)); err != nil {
			t.Fatalf("%s: %v", email, err)
		}
	}
	if n := len(mail.Sent()); n != 0 {
		t.Fatalf("%d mails sent, want none", n)
	}
	if len(st.verif) != 0 {
		t.Fatalf("%d verification rows written, want none", len(st.verif))
	}
}

// No mailer configured: silent success, no token row.
func TestResendEmailVerificationWithoutAMailerWritesNothing(t *testing.T) {
	st := newFake()
	svc := New(st, nil, []byte("0123456789abcdef0123456789abcdef"), WithBackground(func(f func()) { f() }))
	seedAccount(t, st, "nomail@example.com")
	if err := svc.ResendEmailVerification(t.Context(), "nomail@example.com", "10.20.3.1"); err != nil {
		t.Fatal(err)
	}
	if len(st.verif) != 0 {
		t.Fatal("verification row written without a mailer")
	}
}

// 3 per account per hour, counted before the lookup: an unknown address hits the same 429 as a known one.
func TestResendEmailVerificationAccountLimitAppliesToUnknownEmails(t *testing.T) {
	for _, email := range []string{"known@example.com", "nobody@example.com"} {
		svc, st, _ := linkSvc(t)
		seedAccount(t, st, "known@example.com")
		var last error
		for i := 0; i < resendAccountLimit+1; i++ {
			last = svc.ResendEmailVerification(t.Context(), email, fmt.Sprintf("10.21.%d.1", i))
		}
		if apperrCode(t, last) != apperr.CodeRateLimited {
			t.Fatalf("%s: call %d = %v, want RATE_LIMITED", email, resendAccountLimit+1, last)
		}
	}
}

// 10 per IP network per hour, whatever the addresses asked for.
func TestResendEmailVerificationIPLimit(t *testing.T) {
	svc, _, _ := linkSvc(t)
	for i := 0; i < resendIPLimit; i++ {
		if err := svc.ResendEmailVerification(t.Context(), fmt.Sprintf("p%d@example.com", i), "198.51.100.7"); err != nil {
			t.Fatalf("call %d: %v", i+1, err)
		}
	}
	if err := svc.ResendEmailVerification(t.Context(), "last@example.com", "198.51.100.9"); apperrCode(t, err) != apperr.CodeRateLimited {
		t.Fatalf("11th same-network call: %v", err)
	}
}

// A limiter error fails closed.
func TestResendEmailVerificationFailsClosed(t *testing.T) {
	svc, st, _ := linkSvc(t)
	seedAccount(t, st, "closed@example.com")
	for _, scope := range []string{"resend:ip", "resend:account"} {
		st.errs = map[string]error{"Hit:" + scope: errBoom}
		if err := svc.ResendEmailVerification(t.Context(), "closed@example.com", "10.22.0.1"); apperrCode(t, err) != apperr.CodeInternal {
			t.Fatalf("%s error: %v", scope, err)
		}
	}
}

// A failed token write or a lookup error stays invisible (the response would otherwise tell accounts apart) and
// sends no mail; a token generation failure is internal for every address alike, before the lookup.
func TestResendEmailVerificationErrorsStaySilent(t *testing.T) {
	svc, st, mail := linkSvc(t)
	seedAccount(t, st, "write@example.com")
	st.errs = map[string]error{"CreateEmailVerification": errBoom}
	if err := svc.ResendEmailVerification(t.Context(), "write@example.com", "10.23.0.1"); err != nil {
		t.Fatalf("token write error must stay silent: %v", err)
	}
	st.errs = map[string]error{"AccountByEmail": errBoom}
	if err := svc.ResendEmailVerification(t.Context(), "write@example.com", "10.23.0.2"); err != nil {
		t.Fatalf("lookup error must stay silent: %v", err)
	}
	if n := len(mail.Sent()); n != 0 {
		t.Fatalf("%d mails sent, want none", n)
	}

	st.errs = nil
	orig := randRead
	randRead = func([]byte) (int, error) { return 0, errBoom }
	t.Cleanup(func() { randRead = orig })
	for _, email := range []string{"write@example.com", "ghost@example.com"} {
		if err := svc.ResendEmailVerification(context.Background(), email, "10.23.0.3"); apperrCode(t, err) != apperr.CodeInternal {
			t.Fatalf("%s: rand error = %v, want INTERNAL", email, err)
		}
	}
}
