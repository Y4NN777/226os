package service

import (
	"context"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const appID = "00000000-0000-7000-8000-0000000000f1"

// appFake is the application part of fakeStore: it runs authorize on the scripted access.
type appFake struct {
	need      domain.NeedAccess
	createErr error
	created   []domain.NewApplication
	app       domain.ApplicationAccess
	decideErr error
	decided   []string
	getErr    error
	list      []domain.Application
	listErr   error
	query     domain.ApplicationQuery
}

func (f *fakeStore) CreateApplication(_ context.Context, in domain.NewApplication, authorize func(domain.NeedAccess) error) (string, error) {
	if f.apps.createErr != nil {
		return "", f.apps.createErr
	}
	if err := authorize(f.apps.need); err != nil {
		return "", err
	}
	f.apps.created = append(f.apps.created, in)
	return appID, nil
}

func (f *fakeStore) DecideApplication(_ context.Context, _, _, status string, authorize func(domain.ApplicationAccess) error) error {
	if f.apps.decideErr != nil {
		return f.apps.decideErr
	}
	if err := authorize(f.apps.app); err != nil {
		return err
	}
	f.apps.decided = append(f.apps.decided, status)
	return nil
}

func (f *fakeStore) GetApplication(context.Context, string) (domain.Application, error) {
	return domain.Application{ID: appID, Status: "pending"}, f.apps.getErr
}

func (f *fakeStore) ListApplications(_ context.Context, q domain.ApplicationQuery) ([]domain.Application, error) {
	f.apps.query = q
	return f.apps.list, f.apps.listErr
}

var openPublic = domain.NeedAccess{NeedStatus: "open", OwnerID: aliceID, Visibility: "public"}

func TestCreateApplication(t *testing.T) {
	st := &fakeStore{apps: appFake{need: openPublic}}
	a, err := newSvc(t, st, nil, nil).CreateApplication(t.Context(), bob, "need", CreateApplicationRequest{Message: "Je peux aider"})
	if err != nil || a.ID != appID || len(st.apps.created) != 1 || st.apps.created[0].UserID != bobID || st.apps.created[0].NeedID != "need" {
		t.Fatalf("apply %+v %v %+v", a, err, st.apps.created)
	}
	tests := []struct {
		name string
		sub  subject.Subject
		msg  string
		apps appFake
		lim  *fakeLimiter
		code apperr.Code
	}{
		{name: "anonymous", sub: anon, msg: "m", code: apperr.CodeUnauthenticated},
		{name: "unverified", sub: unverif, msg: "m", code: apperr.CodeEmailNotVerified},
		{name: "empty message", sub: bob, msg: "  ", code: apperr.CodeValidation},
		{name: "nul message", sub: bob, msg: "a\x00", code: apperr.CodeValidation},
		{name: "rate limited", sub: bob, msg: "m", lim: &fakeLimiter{scope: writeScope, deny: true}, code: apperr.CodeRateLimited},
		{name: "private project of someone else", sub: bob, msg: "m", apps: appFake{need: domain.NeedAccess{NeedStatus: "open", OwnerID: aliceID, Visibility: "private"}}, code: apperr.CodeNotFound},
		{name: "own project", sub: alice, msg: "m", apps: appFake{need: openPublic}, code: apperr.CodeStateConflict},
		{name: "own private project", sub: alice, msg: "m", apps: appFake{need: domain.NeedAccess{NeedStatus: "open", OwnerID: aliceID, Visibility: "private"}}, code: apperr.CodeStateConflict},
		{name: "closed need", sub: bob, msg: "m", apps: appFake{need: domain.NeedAccess{NeedStatus: "closed", OwnerID: aliceID, Visibility: "public"}}, code: apperr.CodeStateConflict},
		{name: "duplicate", sub: bob, msg: "m", apps: appFake{createErr: domain.ErrDuplicateApplication}, code: apperr.CodeDuplicateApplication},
		{name: "missing need", sub: bob, msg: "m", apps: appFake{createErr: domain.ErrNotFound}, code: apperr.CodeNotFound},
		{name: "store", sub: bob, msg: "m", apps: appFake{createErr: errDB}, code: apperr.CodeInternal},
		{name: "reread", sub: bob, msg: "m", apps: appFake{need: openPublic, getErr: errDB}, code: apperr.CodeInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{apps: tc.apps}
			_, err := newSvc(t, st, nil, tc.lim).CreateApplication(t.Context(), tc.sub, "need", CreateApplicationRequest{Message: tc.msg})
			wantCode(t, err, tc.code)
			if tc.name != "reread" && len(st.apps.created) != 0 {
				t.Fatal("application created")
			}
		})
	}
}

func TestDecideApplication(t *testing.T) {
	pending := domain.ApplicationAccess{ApplicantID: bobID, OwnerID: aliceID, Status: "pending"}
	charlie := subject.Subject{UserID: "00000000-0000-7000-8000-0000000000b3", EmailVerified: true}
	for _, tc := range []struct {
		sub    subject.Subject
		status string
	}{{alice, "accepted"}, {alice, "declined"}, {bob, "withdrawn"}} {
		st := &fakeStore{apps: appFake{app: pending}}
		if _, err := newSvc(t, st, nil, nil).DecideApplication(t.Context(), tc.sub, appID, tc.status); err != nil || len(st.apps.decided) != 1 {
			t.Fatalf("%s by %s: %v", tc.status, tc.sub.UserID, err)
		}
	}
	tests := []struct {
		name, status string
		sub          subject.Subject
		apps         appFake
		lim          *fakeLimiter
		code         apperr.Code
	}{
		{name: "anonymous", status: "accepted", sub: anon, code: apperr.CodeUnauthenticated},
		{name: "unknown status", status: "pending", sub: alice, code: apperr.CodeValidation},
		{name: "rate limited", status: "accepted", sub: alice, lim: &fakeLimiter{scope: writeScope, deny: true}, code: apperr.CodeRateLimited},
		{name: "third party (IDOR)", status: "accepted", sub: charlie, apps: appFake{app: pending}, code: apperr.CodeNotFound},
		{name: "third party withdraws", status: "withdrawn", sub: charlie, apps: appFake{app: pending}, code: apperr.CodeNotFound},
		{name: "applicant accepts own", status: "accepted", sub: bob, apps: appFake{app: pending}, code: apperr.CodeForbidden},
		{name: "owner withdraws", status: "withdrawn", sub: alice, apps: appFake{app: pending}, code: apperr.CodeForbidden},
		{name: "already decided", status: "declined", sub: alice, apps: appFake{app: domain.ApplicationAccess{ApplicantID: bobID, OwnerID: aliceID, Status: "accepted"}}, code: apperr.CodeStateConflict},
		{name: "withdraw after decline", status: "withdrawn", sub: bob, apps: appFake{app: domain.ApplicationAccess{ApplicantID: bobID, OwnerID: aliceID, Status: "declined"}}, code: apperr.CodeStateConflict},
		{name: "race lost", status: "accepted", sub: alice, apps: appFake{decideErr: domain.ErrStateConflict}, code: apperr.CodeStateConflict},
		{name: "missing", status: "accepted", sub: alice, apps: appFake{decideErr: domain.ErrNotFound}, code: apperr.CodeNotFound},
		{name: "reread", status: "accepted", sub: alice, apps: appFake{app: pending, getErr: errDB}, code: apperr.CodeInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			st := &fakeStore{apps: tc.apps}
			_, err := newSvc(t, st, nil, tc.lim).DecideApplication(t.Context(), tc.sub, appID, tc.status)
			wantCode(t, err, tc.code)
			if tc.code != apperr.CodeInternal && len(st.apps.decided) != 0 {
				t.Fatal("application changed")
			}
		})
	}
}

func applications(n int) []domain.Application {
	out := make([]domain.Application, n)
	for i := range out {
		out[i] = domain.Application{ID: appID, CreatedAt: at}
	}
	return out
}

func TestListApplicationsOwnerOnly(t *testing.T) {
	st := &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "private"}, apps: appFake{list: applications(3)}}
	svc := newSvc(t, st, nil, nil)
	page, err := svc.ListApplications(t.Context(), alice, projID, ApplicationListRequest{Status: ptr("pending"), Limit: ptr(2)})
	if err != nil || len(page.Items) != 2 || page.NextCursor == "" ||
		st.apps.query != (domain.ApplicationQuery{ProjectID: projID, OwnerID: aliceID, Status: "pending", Limit: 3}) {
		t.Fatalf("page %+v %v %+v", page, err, st.apps.query)
	}
	st.apps.list = nil
	page, err = svc.ListApplications(t.Context(), alice, projID, ApplicationListRequest{Status: ptr("pending"), Limit: ptr(2), Cursor: page.NextCursor})
	if err != nil || page.Items == nil || st.apps.query.AfterID != appID || !st.apps.query.AfterCreated.Equal(at) {
		t.Fatalf("page2 %+v %v", page, err)
	}
	tests := []struct {
		name string
		sub  subject.Subject
		st   *fakeStore
		req  ApplicationListRequest
		code apperr.Code
	}{
		{name: "anonymous", sub: anon, st: &fakeStore{}, code: apperr.CodeUnauthenticated},
		{name: "limit", sub: alice, st: &fakeStore{}, req: ApplicationListRequest{Limit: ptr(51)}, code: apperr.CodeValidation},
		{name: "missing project", sub: alice, st: &fakeStore{accessErr: domain.ErrNotFound}, code: apperr.CodeNotFound},
		{name: "third party public", sub: bob, st: &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "public"}}, code: apperr.CodeForbidden},
		{name: "third party private", sub: bob, st: &fakeStore{access: domain.Access{OwnerID: aliceID, Visibility: "private"}}, code: apperr.CodeNotFound},
		{name: "cursor", sub: alice, st: &fakeStore{access: domain.Access{OwnerID: aliceID}}, req: ApplicationListRequest{Cursor: "x"}, code: apperr.CodeInvalidCursor},
		{name: "store", sub: alice, st: &fakeStore{access: domain.Access{OwnerID: aliceID}, apps: appFake{listErr: errDB}}, code: apperr.CodeInternal},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := newSvc(t, tc.st, nil, nil).ListApplications(t.Context(), tc.sub, projID, tc.req)
			wantCode(t, err, tc.code)
		})
	}
}
