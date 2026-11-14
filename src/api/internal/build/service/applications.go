package service

import (
	"context"
	"strings"

	"github.com/Y4NN777/226os/src/api/internal/build/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// CreateApplicationRequest is the createApplication body.
type CreateApplicationRequest struct {
	Message string
}

// ApplicationListRequest is listProjectApplications. Nil means unset; nil Limit means 20.
type ApplicationListRequest struct {
	Status *string
	Cursor string
	Limit  *int
}

// ApplicationPage is one application page. NextCursor is empty on the last page.
type ApplicationPage struct {
	Items      []domain.Application
	NextCursor string
}

// CreateApplication records the subject's interest in a contribution need. The need must be open and
// visible, and not on the subject's own project; a second application is refused unless the previous
// one was withdrawn (the row is then reopened).
func (s *Service) CreateApplication(ctx context.Context, sub subject.Subject, needID string, req CreateApplicationRequest) (domain.Application, error) {
	if err := publisher(sub); err != nil {
		return domain.Application{}, err
	}
	// The contract allows an empty message, but applications_message_ck requires 1 to 2000 characters.
	if strings.TrimSpace(req.Message) == "" {
		return domain.Application{}, invalid("/message", "too_short", "message must not be empty.")
	}
	if !validText(req.Message) {
		return domain.Application{}, invalid("/message", "invalid_value", "message must be valid text.")
	}
	if err := s.hit(ctx, writeScope, sub.UserID, writeLimit, writeWindow); err != nil {
		return domain.Application{}, err
	}
	id, err := s.store.CreateApplication(ctx, domain.NewApplication{NeedID: needID, UserID: sub.UserID, Message: req.Message},
		func(a domain.NeedAccess) error {
			switch {
			case a.Visibility != domain.VisibilityPublic && a.OwnerID != sub.UserID:
				return notFound()
			case a.OwnerID == sub.UserID:
				return stateConflict("You cannot apply to your own project.")
			case a.NeedStatus != domain.NeedOpen:
				return stateConflict("This need is not open.")
			}
			return nil
		})
	if err != nil {
		return domain.Application{}, storeErr(err)
	}
	return s.application(ctx, id)
}

// DecideApplication moves a pending application: the applicant may withdraw it; only the project
// owner accepts or declines. Anyone else gets 404 (cross-project IDOR), the wrong party 403, and an
// application that is no longer pending 409. Decided on the locked row.
func (s *Service) DecideApplication(ctx context.Context, sub subject.Subject, id, status string) (domain.Application, error) {
	if sub.Anonymous() {
		return domain.Application{}, unauthenticated()
	}
	if status != domain.ApplicationAccepted && status != domain.ApplicationDeclined && status != domain.ApplicationWithdrawn {
		return domain.Application{}, invalid("/status", "invalid_value", "status must be accepted, declined or withdrawn.")
	}
	if err := s.hit(ctx, writeScope, sub.UserID, writeLimit, writeWindow); err != nil {
		return domain.Application{}, err
	}
	err := s.store.DecideApplication(ctx, id, sub.UserID, status, func(a domain.ApplicationAccess) error {
		applicant, owner := a.ApplicantID == sub.UserID, a.OwnerID == sub.UserID
		switch {
		case !applicant && !owner:
			return notFound()
		case status == domain.ApplicationWithdrawn && !applicant:
			return apperr.New(apperr.CodeForbidden, "Only the applicant can withdraw an application.")
		case status != domain.ApplicationWithdrawn && !owner:
			return apperr.New(apperr.CodeForbidden, "Only the project owner can accept or decline.")
		case a.Status != domain.ApplicationPending:
			return stateConflict("The application is no longer pending.")
		}
		return nil
	})
	if err != nil {
		return domain.Application{}, storeErr(err)
	}
	return s.application(ctx, id)
}

// ListApplications returns one page of a project's applications to its owner only: 403 to anyone
// else on a public project, 404 on a private one, and never a count.
func (s *Service) ListApplications(ctx context.Context, sub subject.Subject, projectID string, req ApplicationListRequest) (ApplicationPage, error) {
	if sub.Anonymous() {
		return ApplicationPage{}, unauthenticated()
	}
	size, err := pageSize(req.Limit)
	if err != nil {
		return ApplicationPage{}, err
	}
	access, err := s.store.ProjectAccess(ctx, projectID)
	if err != nil {
		return ApplicationPage{}, storeErr(err)
	}
	if err := ownerOnly(sub)(access); err != nil {
		return ApplicationPage{}, err
	}
	query := domain.ApplicationQuery{ProjectID: projectID, OwnerID: sub.UserID, Status: deref(req.Status), Limit: size + 1}
	hash := lengthPrefixed("application", projectID, query.Status)
	if req.Cursor != "" {
		at, id, err := s.decode(req.Cursor, hash)
		if err != nil {
			return ApplicationPage{}, err
		}
		query.AfterCreated, query.AfterID = &at, id
	}
	rows, err := s.store.ListApplications(ctx, query)
	if err != nil {
		return ApplicationPage{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	page := ApplicationPage{Items: rows}
	if len(rows) > size {
		page.Items = rows[:size]
		last := page.Items[size-1]
		page.NextCursor = s.encode(last.CreatedAt, last.ID, hash)
	}
	if page.Items == nil {
		page.Items = []domain.Application{}
	}
	return page, nil
}

func (s *Service) application(ctx context.Context, id string) (domain.Application, error) {
	a, err := s.store.GetApplication(ctx, id)
	if err != nil {
		return domain.Application{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	return a, nil
}
