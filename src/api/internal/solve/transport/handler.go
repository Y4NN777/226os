// Package transport is the HTTP edge of Solve: decode, call one service method, encode.
package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/etag"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
	"github.com/Y4NN777/226os/src/api/internal/solve/service"
)

// Handler serves the ten P0 Solve operations: questions, answers, votes and the accepted answer.
type Handler struct {
	svc    *service.Service
	logger *slog.Logger
}

// New returns a handler. logger may be nil.
func New(svc *service.Service, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	return &Handler{svc: svc, logger: logger}
}

// ListQuestions serves GET /questions.
func (h *Handler) ListQuestions(w http.ResponseWriter, r *http.Request, params apigen.ListQuestionsParams) {
	page, err := h.svc.ListQuestions(r.Context(), service.QuestionListRequest{
		Q: params.Q, Tag: params.Tag, State: str(params.State), Country: params.Country, Lang: str(params.Lang),
		Cursor: deref(params.Cursor), Limit: params.Limit,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	items := make([]apigen.Question, len(page.Items))
	for i, q := range page.Items {
		items[i] = questionJSON(q)
	}
	writeJSON(w, http.StatusOK, apigen.QuestionPage{Items: items, NextCursor: next(page.NextCursor)})
}

// CreateQuestion serves POST /questions. The author is the session subject; Idempotency-Key is handled
// by the router middleware.
func (h *Handler) CreateQuestion(w http.ResponseWriter, r *http.Request, _ apigen.CreateQuestionParams) {
	var body apigen.QuestionCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	q, err := h.svc.CreateQuestion(r.Context(), subject.From(r.Context()), service.CreateQuestionRequest{
		Title: body.Title, BodyMD: body.BodyMd, Lang: string(body.Lang), Country: body.Country, Tags: body.Tags,
		Snippets: snippets(body.Snippets), Links: links(body.Links),
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/questions/"+q.ID)
	writeJSON(w, http.StatusCreated, questionJSON(q))
}

// GetQuestion serves GET /questions/{questionId} with its ETag.
func (h *Handler) GetQuestion(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId) {
	q, err := h.svc.GetQuestion(r.Context(), questionID.String())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	etag.Set(w, service.ETag(q))
	writeJSON(w, http.StatusOK, questionJSON(q))
}

// updateBody is QuestionUpdate with the one thing the generated type cannot say: "country": null (clear
// it) is not the same as an absent country (keep it).
type updateBody struct {
	Title    *string                   `json:"title"`
	BodyMd   *string                   `json:"body_md"`
	Tags     *[]string                 `json:"tags"`
	Snippets *[]apigen.Snippet         `json:"snippets"`
	Country  nullableString            `json:"country"`
	Links    *[]apigen.QuestionLinkRef `json:"links"`
}

// nullableString records whether a JSON member was present, and its value (nil for null).
type nullableString struct {
	set   bool
	value *string
}

// UnmarshalJSON is called for a present member, null included.
func (n *nullableString) UnmarshalJSON(b []byte) error {
	n.set = true
	return json.Unmarshal(b, &n.value)
}

// UpdateQuestion serves PATCH /questions/{questionId} (JSON Merge Patch). The If-Match precondition is
// handed to the service as a callback so it runs after the 404 and 403 decisions, on the locked row.
func (h *Handler) UpdateQuestion(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId, _ apigen.UpdateQuestionParams) {
	var body updateBody
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	req := service.UpdateQuestionRequest{Title: body.Title, BodyMD: body.BodyMd, Tags: body.Tags}
	if body.Country.set {
		country := ""
		if body.Country.value != nil {
			country = *body.Country.value
		}
		req.Country = &country
	}
	if body.Snippets != nil {
		s := snippets(body.Snippets)
		req.Snippets = &s
	}
	if body.Links != nil {
		l := links(body.Links)
		req.Links = &l
	}
	q, err := h.svc.UpdateQuestion(r.Context(), subject.From(r.Context()), questionID.String(), req,
		func(tag string) error { return etag.Check(r, tag) })
	if err != nil {
		h.fail(w, r, err)
		return
	}
	etag.Set(w, service.ETag(q))
	writeJSON(w, http.StatusOK, questionJSON(q))
}

// ListAnswers serves GET /questions/{questionId}/answers.
func (h *Handler) ListAnswers(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId, params apigen.ListAnswersParams) {
	page, err := h.svc.ListAnswers(r.Context(), service.AnswerListRequest{
		QuestionID: questionID.String(), Cursor: deref(params.Cursor), Limit: params.Limit,
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	items := make([]apigen.Answer, len(page.Items))
	for i, a := range page.Items {
		items[i] = answerJSON(a)
	}
	writeJSON(w, http.StatusOK, apigen.AnswerPage{Items: items, NextCursor: next(page.NextCursor)})
}

// CreateAnswer serves POST /questions/{questionId}/answers; Idempotency-Key is handled by the router
// middleware.
func (h *Handler) CreateAnswer(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId, _ apigen.CreateAnswerParams) {
	var body apigen.AnswerCreate
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	a, err := h.svc.CreateAnswer(r.Context(), subject.From(r.Context()), service.CreateAnswerRequest{
		QuestionID: questionID.String(), BodyMD: body.BodyMd, Lang: string(body.Lang),
		Snippets: snippets(body.Snippets), Sources: slice(body.Sources),
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/questions/"+a.QuestionID+"/answers")
	writeJSON(w, http.StatusCreated, answerJSON(a))
}

// CastVote serves POST /votes. The voter is the session subject; the contract does not bind an
// Idempotency-Key, repeating a vote is idempotent by construction.
func (h *Handler) CastVote(w http.ResponseWriter, r *http.Request) {
	var body apigen.VoteRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	state, err := h.svc.CastVote(r.Context(), subject.From(r.Context()), service.VoteRequest{
		TargetType: string(body.TargetType), TargetID: body.TargetId.String(), Value: int(body.Value),
	})
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, voteStateJSON(state))
}

// RetractVote serves DELETE /votes.
func (h *Handler) RetractVote(w http.ResponseWriter, r *http.Request, params apigen.RetractVoteParams) {
	state, err := h.svc.RetractVote(r.Context(), subject.From(r.Context()), string(params.TargetType), params.TargetId.String())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, voteStateJSON(state))
}

// AcceptAnswer serves POST /questions/{questionId}/accept.
func (h *Handler) AcceptAnswer(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId) {
	var body apigen.AcceptRequest
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	q, err := h.svc.AcceptAnswer(r.Context(), subject.From(r.Context()), questionID.String(), body.AnswerId.String())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, questionJSON(q))
}

// UnacceptAnswer serves POST /questions/{questionId}/unaccept.
func (h *Handler) UnacceptAnswer(w http.ResponseWriter, r *http.Request, questionID apigen.QuestionId) {
	q, err := h.svc.UnacceptAnswer(r.Context(), subject.From(r.Context()), questionID.String())
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, questionJSON(q))
}

func voteStateJSON(v domain.VoteState) apigen.VoteState {
	return apigen.VoteState{
		TargetType: apigen.VoteTargetType(v.TargetType), TargetId: mustID(v.TargetID),
		Value: apigen.VoteStateValue(v.Value), Score: v.Score,
	}
}

func questionJSON(q domain.Question) apigen.Question {
	related := make([]apigen.RelatedLink, len(q.Related))
	for i, l := range q.Related {
		lang := apigen.Lang(l.Lang)
		related[i] = apigen.RelatedLink{Id: mustID(l.ID), Type: apigen.RelatedLinkType(l.Type), Title: l.Title, Lang: &lang}
	}
	id := mustID(q.ID)
	out := apigen.Question{
		Id: &id, Title: q.Title, BodyMd: q.BodyMD, Lang: apigen.Lang(q.Lang), Tags: q.Tags,
		Snippets: snippetsJSON(q.Snippets), State: apigen.QuestionState(q.State()), Author: authorJSON(q.Author),
		AnswerCount: q.AnswerCount, VoteScore: q.VoteScore, Related: related,
		CreatedAt: q.CreatedAt.UTC(), UpdatedAt: q.UpdatedAt.UTC(),
	}
	if q.Country != "" {
		c := q.Country
		out.Country = &c
	}
	if q.AcceptedAnswerID != "" {
		a := mustID(q.AcceptedAnswerID)
		out.AcceptedAnswerId = &a
	}
	return out
}

func answerJSON(a domain.Answer) apigen.Answer {
	id := mustID(a.ID)
	return apigen.Answer{
		Id: &id, QuestionId: mustID(a.QuestionID), BodyMd: a.BodyMD, Lang: apigen.Lang(a.Lang),
		Snippets: snippetsJSON(a.Snippets), Sources: a.Sources, Author: authorJSON(a.Author),
		VoteScore: a.VoteScore, IsAccepted: a.Accepted, CreatedAt: a.CreatedAt.UTC(), UpdatedAt: a.UpdatedAt.UTC(),
	}
}

func snippetsJSON(in []domain.Snippet) []apigen.Snippet {
	out := make([]apigen.Snippet, len(in))
	for i, s := range in {
		out[i] = apigen.Snippet{Language: s.Language, Code: s.Code}
	}
	return out
}

func authorJSON(a domain.Author) apigen.AuthorRef {
	out := apigen.AuthorRef{Id: mustID(a.ID), Username: a.Username, DisplayName: a.DisplayName}
	if a.Country != "" {
		c := a.Country
		out.Country = &c
	}
	return out
}

func snippets(in *[]apigen.Snippet) []domain.Snippet {
	if in == nil {
		return nil
	}
	out := make([]domain.Snippet, len(*in))
	for i, s := range *in {
		out[i] = domain.Snippet{Language: s.Language, Code: s.Code}
	}
	return out
}

func links(in *[]apigen.QuestionLinkRef) []domain.LinkRef {
	if in == nil {
		return nil
	}
	out := make([]domain.LinkRef, len(*in))
	for i, l := range *in {
		out[i] = domain.LinkRef{Type: string(l.Type), ID: l.Id.String()}
	}
	return out
}

func slice(in *[]string) []string {
	if in == nil {
		return nil
	}
	return *in
}

// mustID parses an id the repository formatted from a uuid column; it cannot fail.
func mustID(s string) apigen.Uuid { return uuid.MustParse(s) }

func next(c string) *string {
	if c == "" {
		return nil
	}
	return &c
}

func str[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// fail writes the problem+json of err; httpx.WriteError turns anything that is not an *apperr.Error
// into a generic 500 and logs the cause.
func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	httpx.WriteError(h.logger)(w, r, err)
}
