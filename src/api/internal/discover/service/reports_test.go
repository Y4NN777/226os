package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/discover/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	member   = "00000000-0000-7000-8000-0000000000d1"
	resource = "00000000-0000-7000-8000-0000000000a1"
	source   = "00000000-0000-7000-8000-0000000000b1"
)

type fakeReports struct {
	got []domain.NewReport
	out domain.ErrorReport
	err error
}

func (f *fakeReports) CreateReport(_ context.Context, in domain.NewReport) (domain.ErrorReport, error) {
	f.got = append(f.got, in)
	return f.out, f.err
}

type fakeLimiter struct {
	allowed bool
	retry   time.Duration
	err     error
	calls   []string
	limit   int
	window  time.Duration
}

func (f *fakeLimiter) Hit(_ context.Context, scope, value string, limit int, window time.Duration) (bool, time.Duration, error) {
	f.calls = append(f.calls, scope+"|"+value)
	f.limit, f.window = limit, window
	return f.allowed, f.retry, f.err
}

func verified() subject.Subject { return subject.Subject{UserID: member, EmailVerified: true} }

func okReq() ReportRequest {
	return ReportRequest{ResourceID: resource, Kind: "outdated", Body: "The pricing page moved."}
}

func code(t *testing.T, err error, want apperr.Code) *apperr.Error {
	t.Helper()
	var ae *apperr.Error
	if !errors.As(err, &ae) || ae.Code != want {
		t.Fatalf("err = %v, want code %s", err, want)
	}
	return ae
}

func TestCreateErrorReportStoresTheReportOfTheSessionSubject(t *testing.T) {
	want := domain.ErrorReport{ID: "r1", ResourceID: resource, Kind: "outdated", Status: "received"}
	store, lim := &fakeReports{out: want}, &fakeLimiter{allowed: true}
	req := okReq()
	req.SourceID = source
	got, err := NewReports(store, lim).Create(t.Context(), verified(), req)
	if err != nil || got != want {
		t.Fatalf("Create = %+v, %v", got, err)
	}
	wantIn := domain.NewReport{ResourceID: resource, ReporterID: member, Kind: "outdated", Body: "The pricing page moved.", SourceID: source}
	if len(store.got) != 1 || store.got[0] != wantIn {
		t.Fatalf("stored %+v, want %+v", store.got, wantIn)
	}
	if len(lim.calls) != 1 || lim.calls[0] != "write:user|"+member || lim.limit != 60 || lim.window != time.Minute {
		t.Fatalf("limiter %v %d %v; want write:user, 60 per minute, keyed by the session user", lim.calls, lim.limit, lim.window)
	}
}

func TestCreateErrorReportAuthorization(t *testing.T) {
	cases := []struct {
		name string
		sub  subject.Subject
		want apperr.Code
	}{
		{"anonymous", subject.Subject{}, apperr.CodeUnauthenticated},
		{"member without a verified email", subject.Subject{UserID: member}, apperr.CodeEmailNotVerified},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, lim := &fakeReports{}, &fakeLimiter{allowed: true}
			_, err := NewReports(store, lim).Create(t.Context(), tc.sub, okReq())
			_ = code(t, err, tc.want)
			if len(store.got) != 0 || len(lim.calls) != 0 {
				t.Fatal("a denied caller reached the limiter or the store")
			}
		})
	}
}

func TestCreateErrorReportValidation(t *testing.T) {
	cases := []struct {
		name  string
		edit  func(*ReportRequest)
		field string
	}{
		{"unknown kind", func(r *ReportRequest) { r.Kind = "spam" }, "kind"},
		{"empty body", func(r *ReportRequest) { r.Body = "" }, "body"},
		{"blank body", func(r *ReportRequest) { r.Body = " \n\t " }, "body"},
		{"body of 2001 characters", func(r *ReportRequest) { r.Body = strings.Repeat("é", 2001) }, "body"},
		{"body with NUL", func(r *ReportRequest) { r.Body = "a\x00b" }, "body"},
		{"body that is not UTF-8", func(r *ReportRequest) { r.Body = "a\xffb" }, "body"},
		{"malformed source id", func(r *ReportRequest) { r.SourceID = "nope" }, "source_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store, lim := &fakeReports{}, &fakeLimiter{allowed: true}
			req := okReq()
			tc.edit(&req)
			_, err := NewReports(store, lim).Create(t.Context(), verified(), req)
			ae := code(t, err, apperr.CodeValidation)
			if len(ae.Violations) != 1 || ae.Violations[0].Field != tc.field {
				t.Fatalf("violations %+v, want one on %q", ae.Violations, tc.field)
			}
			if len(store.got) != 0 || len(lim.calls) != 0 {
				t.Fatal("an invalid request reached the limiter or the store")
			}
		})
	}
}

func TestCreateErrorReportAcceptsTheBoundaries(t *testing.T) {
	for _, body := range []string{"x", strings.Repeat("é", 2000)} {
		store := &fakeReports{}
		req := okReq()
		req.Body = body
		if _, err := NewReports(store, &fakeLimiter{allowed: true}).Create(t.Context(), verified(), req); err != nil {
			t.Fatalf("body of %d characters rejected: %v", len([]rune(body)), err)
		}
	}
}

func TestCreateErrorReportRateLimit(t *testing.T) {
	store := &fakeReports{}
	_, err := NewReports(store, &fakeLimiter{retry: 17 * time.Second}).Create(t.Context(), verified(), okReq())
	ae := code(t, err, apperr.CodeRateLimited)
	if ae.RetryAfter != 17*time.Second || len(store.got) != 0 {
		t.Fatalf("retry-after %v, stored %d; want 17s and nothing stored", ae.RetryAfter, len(store.got))
	}
}

// The write limit is an authentication-adjacent control (ADR-0003): a limiter failure fails CLOSED.
func TestCreateErrorReportFailsClosedWhenTheLimiterFails(t *testing.T) {
	store := &fakeReports{}
	boom := errors.New("connection refused")
	_, err := NewReports(store, &fakeLimiter{err: boom}).Create(t.Context(), verified(), okReq())
	ae := code(t, err, apperr.CodeInternal)
	if !errors.Is(ae, boom) || len(store.got) != 0 {
		t.Fatalf("err = %v, stored %d; want INTERNAL_ERROR wrapping the cause and nothing stored", err, len(store.got))
	}
}

func TestCreateErrorReportPropagatesStoreErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want apperr.Code
	}{
		{"resource not visible", domain.ErrNotFound, apperr.CodeNotFound},
		{"source of another resource", domain.ErrSourceMismatch, apperr.CodeValidation},
		{"database failure", errors.New("db down"), apperr.CodeInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewReports(&fakeReports{err: tc.err}, &fakeLimiter{allowed: true}).Create(t.Context(), verified(), okReq())
			ae := code(t, err, tc.want)
			if tc.want == apperr.CodeValidation && (len(ae.Violations) != 1 || ae.Violations[0].Field != "source_id") {
				t.Fatalf("violations %+v, want one on source_id", ae.Violations)
			}
		})
	}
}
