package transport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// waitMail polls until the fake mailer holds at least n messages: sends go
// out asynchronously so the request path never waits on SMTP.
func waitMail(t *testing.T, mail *mailer.Fake, n int) []mailer.Message {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if sent := mail.Sent(); len(sent) >= n {
			return sent
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected %d mails, got %d", n, len(mail.Sent()))
	return nil
}

// withSubject returns req with a member Subject in its context, like
// subject.Middleware installs after a successful session resolution.
func withSubject(req *http.Request, role string) context.Context {
	return subject.With(req.Context(), subject.Subject{
		UserID: "019a0000-0000-7000-8000-0000000000ff", Roles: []string{role},
	})
}

// withSubjectID is withSubject for a specific account id.
func withSubjectID(req *http.Request, userID, role string) context.Context {
	return subject.With(req.Context(), subject.Subject{
		UserID: userID, Roles: []string{role},
	})
}

// op is the handler shape every S1 auth operation shares.
type op func(http.ResponseWriter, *http.Request)

// doJSON posts a JSON body through op and returns the recorder.
func doJSON(t *testing.T, fn op, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	fn(rec, req)
	return rec
}

// doReq builds a request with optional headers and cookies.
func doReq(t *testing.T, method, path string, headers map[string]string, cookies ...*http.Cookie) *http.Request {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	return req
}

// call invokes op with req and returns the recorder.
func call(t *testing.T, fn op, req *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	fn(rec, req)
	return rec
}
