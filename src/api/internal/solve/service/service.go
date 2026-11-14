// Package service is the Solve use case (US-SOLVE-001): public reads of published questions and their
// answers, member publishing with the original language kept, votes and the accepted answer.
package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/etag"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
)

const (
	defaultLimit = 20
	maxLimit     = 50
	maxQuery     = 200
	maxTags      = 5

	// The schema keeps one snippet per question (questions.snippet) and none per answer, and a
	// snippet language of at most 30 characters (migration 00005). The contract allows more; the
	// service answers 422 rather than drop what the author sent.
	maxQuestionSnippets = 1
	maxSnippetLanguage  = 30

	// Writes draw on the one authenticated-writes budget, 60 per user per minute (ADR-0003); Learn,
	// Build and error reports share the same write:user scope.
	writeScope  = "write:user"
	writeLimit  = 60
	writeWindow = time.Minute
)

// Store loads and saves Solve content. The service owns this port. A question that is missing or
// removed is domain.ErrNotFound everywhere.
//
// UpdateQuestion locks the question, calls guard with the locked row (ID, Author.ID, counters and
// UpdatedAt only) and stops with guard's error when it fails, so the ownership and ETag decisions are
// taken on the row the edit then changes. CreateQuestion and UpdateQuestion return domain.ErrLinkTarget
// and domain.ErrUnknownCountry when a link target or the country is not usable.
type Store interface {
	ListQuestions(ctx context.Context, q domain.QuestionQuery) ([]domain.Question, error)
	GetQuestion(ctx context.Context, id, viewerID string) (domain.Question, error)
	CreateQuestion(ctx context.Context, in domain.NewQuestion) (domain.Question, error)
	UpdateQuestion(ctx context.Context, id string, patch domain.QuestionPatch, guard func(domain.Question) error) (domain.Question, error)
	ListAnswers(ctx context.Context, q domain.AnswerQuery) ([]domain.Answer, error)
	CreateAnswer(ctx context.Context, in domain.NewAnswer) (domain.Answer, error)
	SetVote(ctx context.Context, v domain.Vote) (domain.VoteState, error)
	SetAcceptedAnswer(ctx context.Context, c domain.AcceptChange) (domain.Question, error)
}

// SetVote writes the vote (value 0 deletes it) and moves the target's vote_score by the difference, in one
// transaction under the target's row lock. It returns domain.ErrNotFound when the target is missing, not
// published or in a removed question, and domain.ErrSelfVote when the voter wrote the target.
//
// SetAcceptedAnswer locks the question, then returns domain.ErrNotFound (missing or removed question),
// domain.ErrForbidden (the actor is not its author) or domain.ErrAnswerNotInQuestion, and otherwise
// replaces or clears the accepted answer in that one transaction.

// Cursors signs page cursors and rejects a cursor used with different filters.
type Cursors interface {
	Encode(cursor.Payload) string
	Decode(token, filterHash string) (cursor.Payload, error)
}

// Limiter is the fixed-window rate limiter. The service owns this port.
type Limiter interface {
	Hit(ctx context.Context, scope, value string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)
}

// QuestionListRequest is listQuestions. Nil means unset; nil Limit means 20.
type QuestionListRequest struct {
	Q, Tag, State, Country, Lang *string
	Cursor                       string
	Limit                        *int
}

// AnswerListRequest is listAnswers. Nil Limit means 20.
type AnswerListRequest struct {
	QuestionID string
	Cursor     string
	Limit      *int
}

// CreateQuestionRequest is the createQuestion body.
type CreateQuestionRequest struct {
	Title, BodyMD, Lang string
	Country             *string
	Tags                []string
	Snippets            []domain.Snippet
	Links               []domain.LinkRef
}

// UpdateQuestionRequest is the updateQuestion body (JSON Merge Patch): nil means unchanged. A Country
// pointing to "" clears the country; a Snippets or Links slice that is set but empty clears them.
type UpdateQuestionRequest struct {
	Title, BodyMD, Country *string
	Tags                   *[]string
	Snippets               *[]domain.Snippet
	Links                  *[]domain.LinkRef
}

// CreateAnswerRequest is the createAnswer body.
type CreateAnswerRequest struct {
	QuestionID, BodyMD, Lang string
	Snippets                 []domain.Snippet
	Sources                  []string
}

// VoteRequest is the castVote body.
type VoteRequest struct {
	TargetType, TargetID string
	Value                int
}

// QuestionPage is one question page. NextCursor is empty on the last page.
type QuestionPage struct {
	Items      []domain.Question
	NextCursor string
}

// AnswerPage is one answer page. NextCursor is empty on the last page.
type AnswerPage struct {
	Items      []domain.Answer
	NextCursor string
}

// Service is the Solve use case.
type Service struct {
	store   Store
	cursors Cursors
	limiter Limiter
}

// New returns the Solve service.
func New(store Store, cursors Cursors, limiter Limiter) *Service {
	return &Service{store: store, cursors: cursors, limiter: limiter}
}

// ETag is the strong entity tag of a question as the viewer (the session subject's id, empty when
// anonymous) sees it: the question's id and updated_at, which only an edit of the question changes, and
// the viewer, because viewer_vote is part of the representation. Answers, votes and acceptance
// (answer_count, vote_score, accepted_answer_id, viewer_vote) do not touch updated_at, so they do not make
// an author's tag stale; the tag never depends on the votes table, so If-Match keeps validating the
// question version, recomputed on the locked row for the caller.
func ETag(q domain.Question, viewerID string) string {
	return etag.Tag(q.ID, q.UpdatedAt.UTC().Format(time.RFC3339Nano), viewerID)
}

// ListQuestions returns one page of published questions: newest first, or with q the most relevant
// first, then newest.
func (s *Service) ListQuestions(ctx context.Context, sub subject.Subject, req QuestionListRequest) (QuestionPage, error) {
	q, err := textFilter(req.Q)
	if err != nil {
		return QuestionPage{}, err
	}
	size, err := pageSize(req.Limit)
	if err != nil {
		return QuestionPage{}, err
	}
	query := domain.QuestionQuery{
		Q: q, Tag: deref(req.Tag), State: deref(req.State), Country: deref(req.Country), Lang: deref(req.Lang),
		ViewerID: sub.UserID, Limit: size + 1,
	}
	hash := lengthPrefixed("question", query.Q, query.Tag, query.State, query.Country, query.Lang)
	if req.Cursor != "" {
		p, err := s.cursors.Decode(req.Cursor, hash)
		if err != nil {
			return QuestionPage{}, err
		}
		rank, at, ok := decodeQuestionSort(p.Sort)
		if !ok || !validID(p.ID) {
			return QuestionPage{}, invalidCursor()
		}
		query.AfterRank, query.AfterCreated, query.AfterID = rank, &at, p.ID
	}
	rows, err := s.store.ListQuestions(ctx, query)
	if err != nil {
		return QuestionPage{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	page := QuestionPage{Items: rows}
	if len(rows) > size {
		page.Items = rows[:size]
		last := page.Items[size-1]
		page.NextCursor = s.cursors.Encode(cursor.Payload{Sort: encodeQuestionSort(last), ID: last.ID, FilterHash: hash})
	}
	if page.Items == nil {
		page.Items = []domain.Question{}
	}
	return page, nil
}

// GetQuestion returns a published question; unknown and removed are the same 404.
func (s *Service) GetQuestion(ctx context.Context, sub subject.Subject, id string) (domain.Question, error) {
	q, err := s.store.GetQuestion(ctx, id, sub.UserID)
	if err != nil {
		return domain.Question{}, readErr(err)
	}
	return q, nil
}

// CreateQuestion publishes a question for the session subject. It needs a session and a verified
// email. The author is always the subject, never the request. The text is stored as given: no
// trimming, no translation. Snippets are stored and never executed. The limiter fails CLOSED
// (ADR-0003: a write).
func (s *Service) CreateQuestion(ctx context.Context, sub subject.Subject, req CreateQuestionRequest) (domain.Question, error) {
	if err := requireVerified(sub, "publish a question"); err != nil {
		return domain.Question{}, err
	}
	in, err := validateNew(req)
	if err != nil {
		return domain.Question{}, err
	}
	in.AuthorID = sub.UserID
	if err := s.hit(ctx, sub); err != nil {
		return domain.Question{}, err
	}
	q, err := s.store.CreateQuestion(ctx, in)
	if err != nil {
		return domain.Question{}, writeErr(err)
	}
	return q, nil
}

// UpdateQuestion edits the caller's own question. Order: session (401), body (422), the write budget (429, fails closed), then in the store's
// transaction the unknown or removed question (404), a visible question that is not the caller's (403),
// and only then the precondition (check: 428 missing, 412 stale), so 403 and 404 never depend on the tag.
func (s *Service) UpdateQuestion(ctx context.Context, sub subject.Subject, id string, req UpdateQuestionRequest,
	check func(tag string) error,
) (domain.Question, error) {
	if sub.Anonymous() {
		return domain.Question{}, apperr.New(apperr.CodeUnauthenticated, "Sign in to edit a question.")
	}
	patch, err := validatePatch(req)
	if err != nil {
		return domain.Question{}, err
	}
	if err := s.hit(ctx, sub); err != nil {
		return domain.Question{}, err
	}
	guard := func(cur domain.Question) error {
		if cur.Author.ID != sub.UserID {
			return domain.ErrForbidden
		}
		return check(ETag(cur, sub.UserID))
	}
	q, err := s.store.UpdateQuestion(ctx, id, patch, guard)
	var precondition *apperr.Error
	switch {
	case errors.As(err, &precondition):
		return domain.Question{}, precondition
	case errors.Is(err, domain.ErrForbidden):
		return domain.Question{}, apperr.New(apperr.CodeForbidden, "Only the author can edit this question.")
	case err != nil:
		return domain.Question{}, writeErr(err)
	}
	return q, nil
}

// ListAnswers returns one page of a published question's answers: the accepted answer first, then by
// vote score, then oldest first, then id.
//
// The cursor is a keyset on that live order, so it holds the last row's position, not a snapshot. A vote
// or an acceptance between two pages never breaks the walk: the cursor stays valid and the next page
// continues after the cut position. It can move an answer across that position: one whose score fell
// below the cut shows again on a later page (a duplicate), one whose score rose above it is missed by
// this walk and shows first in the next. Every answer that did not change position is returned exactly
// once, in order. A client walking the pages de-duplicates by id and refreshes from the first page to
// see the current order.
func (s *Service) ListAnswers(ctx context.Context, sub subject.Subject, req AnswerListRequest) (AnswerPage, error) {
	size, err := pageSize(req.Limit)
	if err != nil {
		return AnswerPage{}, err
	}
	query := domain.AnswerQuery{QuestionID: req.QuestionID, ViewerID: sub.UserID, Limit: size + 1}
	hash := lengthPrefixed("answers", req.QuestionID)
	if req.Cursor != "" {
		p, err := s.cursors.Decode(req.Cursor, hash)
		if err != nil {
			return AnswerPage{}, err
		}
		key, ok := decodeAnswerKey(p)
		if !ok {
			return AnswerPage{}, invalidCursor()
		}
		query.After = &key
	}
	rows, err := s.store.ListAnswers(ctx, query)
	if err != nil {
		return AnswerPage{}, readErr(err)
	}
	page := AnswerPage{Items: rows}
	if len(rows) > size {
		page.Items = rows[:size]
		last := page.Items[size-1]
		page.NextCursor = s.cursors.Encode(encodeAnswerKey(last, hash))
	}
	if page.Items == nil {
		page.Items = []domain.Answer{}
	}
	return page, nil
}

// CreateAnswer saves an answer to a published question for the session subject: verified email, the
// question's answer_count raised in the same transaction. An unknown or removed question is a 404.
// Answers carry no code snippets at P0 (no column): a non-empty snippets is a 422.
func (s *Service) CreateAnswer(ctx context.Context, sub subject.Subject, req CreateAnswerRequest) (domain.Answer, error) {
	if err := requireVerified(sub, "answer"); err != nil {
		return domain.Answer{}, err
	}
	if len(req.Snippets) > 0 {
		return domain.Answer{}, invalid("/snippets", "too_many_items", "answers cannot carry code snippets yet; put the code in body_md.")
	}
	if !usableText(req.BodyMD) {
		return domain.Answer{}, invalid("/body_md", "invalid_value", "body_md must be non-blank valid text.")
	}
	sources, err := validateSources(req.Sources)
	if err != nil {
		return domain.Answer{}, err
	}
	if err := s.hit(ctx, sub); err != nil {
		return domain.Answer{}, err
	}
	a, err := s.store.CreateAnswer(ctx, domain.NewAnswer{
		QuestionID: req.QuestionID, AuthorID: sub.UserID, BodyMD: req.BodyMD, Lang: req.Lang, Sources: sources,
	})
	if err != nil {
		return domain.Answer{}, readErr(err)
	}
	return a, nil
}

// CastVote records the subject's vote on a published question or answer: verified email, one vote per
// target (the same value again changes nothing, another value replaces it), never on one's own content.
// The score moves by the difference in the store's transaction; it is never computed here.
func (s *Service) CastVote(ctx context.Context, sub subject.Subject, req VoteRequest) (domain.VoteState, error) {
	if err := requireVerified(sub, "vote"); err != nil {
		return domain.VoteState{}, err
	}
	if req.Value != 1 && req.Value != -1 {
		return domain.VoteState{}, invalid("/value", "invalid_value", "value must be 1 or -1.")
	}
	if err := validateTarget("/target_type", "/target_id", req.TargetType, req.TargetID); err != nil {
		return domain.VoteState{}, err
	}
	return s.vote(ctx, sub, domain.Vote{UserID: sub.UserID, TargetType: req.TargetType, TargetID: req.TargetID, Value: req.Value})
}

// RetractVote removes the subject's vote on a published question or answer; having none is a success
// (value 0).
func (s *Service) RetractVote(ctx context.Context, sub subject.Subject, targetType, targetID string) (domain.VoteState, error) {
	if err := requireVerified(sub, "vote"); err != nil {
		return domain.VoteState{}, err
	}
	if err := validateTarget("target_type", "target_id", targetType, targetID); err != nil {
		return domain.VoteState{}, err
	}
	return s.vote(ctx, sub, domain.Vote{UserID: sub.UserID, TargetType: targetType, TargetID: targetID})
}

// vote spends one write before the store looks the target up, on purpose: the budget bounds every attempt
// that reaches the database, so an unknown id costs the caller a write like any other (#155).
func (s *Service) vote(ctx context.Context, sub subject.Subject, v domain.Vote) (domain.VoteState, error) {
	if err := s.hit(ctx, sub); err != nil {
		return domain.VoteState{}, err
	}
	state, err := s.store.SetVote(ctx, v)
	if errors.Is(err, domain.ErrSelfVote) {
		return domain.VoteState{}, apperr.New(apperr.CodeSelfVoteForbidden, "You cannot vote on your own content.")
	}
	if err != nil {
		return domain.VoteState{}, readErr(err)
	}
	return state, nil
}

// AcceptAnswer makes one answer of the question the accepted answer, replacing the previous one. Only the
// question's author may; an answer of another question, unknown or removed is a 422.
func (s *Service) AcceptAnswer(ctx context.Context, sub subject.Subject, questionID, answerID string) (domain.Question, error) {
	if err := requireVerified(sub, "accept an answer"); err != nil {
		return domain.Question{}, err
	}
	if !validID(answerID) {
		return domain.Question{}, invalid("/answer_id", "invalid_value", "answer_id must be a UUID.")
	}
	return s.setAccepted(ctx, sub, domain.AcceptChange{QuestionID: questionID, ActorID: sub.UserID, AnswerID: answerID})
}

// UnacceptAnswer clears the accepted answer; nothing accepted is a success. Only the question's author may.
func (s *Service) UnacceptAnswer(ctx context.Context, sub subject.Subject, questionID string) (domain.Question, error) {
	if err := requireVerified(sub, "change the accepted answer"); err != nil {
		return domain.Question{}, err
	}
	return s.setAccepted(ctx, sub, domain.AcceptChange{QuestionID: questionID, ActorID: sub.UserID})
}

func (s *Service) setAccepted(ctx context.Context, sub subject.Subject, c domain.AcceptChange) (domain.Question, error) {
	if err := s.hit(ctx, sub); err != nil {
		return domain.Question{}, err
	}
	q, err := s.store.SetAcceptedAnswer(ctx, c)
	switch {
	case errors.Is(err, domain.ErrForbidden):
		return domain.Question{}, apperr.New(apperr.CodeForbidden, "Only the author can accept an answer.")
	case errors.Is(err, domain.ErrAnswerNotInQuestion):
		return domain.Question{}, invalid("/answer_id", "invalid_value", "answer_id must be a published answer of this question.")
	case err != nil:
		return domain.Question{}, readErr(err)
	}
	return q, nil
}

func validateTarget(typeField, idField, targetType, targetID string) error {
	if targetType != domain.TargetQuestion && targetType != domain.TargetAnswer {
		return invalid(typeField, "invalid_value", "target_type must be question or answer.")
	}
	if !validID(targetID) {
		return invalid(idField, "invalid_value", "target_id must be a UUID.")
	}
	return nil
}

func requireVerified(sub subject.Subject, action string) error {
	if sub.Anonymous() {
		return apperr.New(apperr.CodeUnauthenticated, "Sign in to "+action+".")
	}
	if !sub.EmailVerified {
		return apperr.New(apperr.CodeEmailNotVerified, "Verify your email address to "+action+".")
	}
	return nil
}

// hit counts one write on the shared budget and fails closed when the limiter errors.
func (s *Service) hit(ctx context.Context, sub subject.Subject) error {
	ok, retry, err := s.limiter.Hit(ctx, writeScope, sub.UserID, writeLimit, writeWindow)
	if err != nil {
		return apperr.Wrap(apperr.CodeInternal, err)
	}
	if !ok {
		return apperr.New(apperr.CodeRateLimited, "Too many requests.").WithRetryAfter(retry)
	}
	return nil
}

func validateNew(req CreateQuestionRequest) (domain.NewQuestion, error) {
	if !usableText(req.Title) {
		return domain.NewQuestion{}, invalid("/title", "invalid_value", "title must be non-blank valid text.")
	}
	if !usableText(req.BodyMD) {
		return domain.NewQuestion{}, invalid("/body_md", "invalid_value", "body_md must be non-blank valid text.")
	}
	tags, err := validateTags(req.Tags)
	if err != nil {
		return domain.NewQuestion{}, err
	}
	snippets, err := validateSnippets(req.Snippets)
	if err != nil {
		return domain.NewQuestion{}, err
	}
	links, err := validateLinks(req.Links)
	if err != nil {
		return domain.NewQuestion{}, err
	}
	return domain.NewQuestion{
		Title: req.Title, BodyMD: req.BodyMD, Lang: req.Lang, Country: deref(req.Country),
		Tags: tags, Snippets: snippets, Links: links,
	}, nil
}

func validatePatch(req UpdateQuestionRequest) (domain.QuestionPatch, error) {
	var p domain.QuestionPatch
	if req.Title != nil {
		if !usableText(*req.Title) {
			return p, invalid("/title", "invalid_value", "title must be non-blank valid text.")
		}
		p.Title = req.Title
	}
	if req.BodyMD != nil {
		if !usableText(*req.BodyMD) {
			return p, invalid("/body_md", "invalid_value", "body_md must be non-blank valid text.")
		}
		p.BodyMD = req.BodyMD
	}
	p.Country = req.Country
	if req.Tags != nil {
		tags, err := validateTags(*req.Tags)
		if err != nil {
			return p, err
		}
		p.Tags = &tags
	}
	if req.Snippets != nil {
		snippets, err := validateSnippets(*req.Snippets)
		if err != nil {
			return p, err
		}
		p.Snippets = &snippets
	}
	if req.Links != nil {
		links, err := validateLinks(*req.Links)
		if err != nil {
			return p, err
		}
		p.Links = &links
	}
	return p, nil
}

func validateTags(in []string) ([]string, error) {
	// Sorted, so two requests creating the same new tags lock the tags rows in the same order.
	tags := dedupe(in)
	slices.Sort(tags)
	if len(tags) < 1 || len(tags) > maxTags {
		return nil, invalid("/tags", "out_of_range", "tags must hold 1 to 5 distinct slugs.")
	}
	return tags, nil
}

func validateSnippets(in []domain.Snippet) ([]domain.Snippet, error) {
	if len(in) > maxQuestionSnippets {
		return nil, invalid("/snippets", "too_many_items", "a question holds one code snippet for now.")
	}
	for _, sn := range in {
		if !usableText(sn.Code) || !validText(sn.Language) || utf8.RuneCountInString(sn.Language) > maxSnippetLanguage {
			return nil, invalid("/snippets", "invalid_value", "a snippet needs valid code and a language of at most 30 characters.")
		}
	}
	return in, nil
}

func validateLinks(in []domain.LinkRef) ([]domain.LinkRef, error) {
	links := make([]domain.LinkRef, 0, len(in))
	for _, l := range in {
		if !slices.Contains(domain.LinkTypes, l.Type) {
			return nil, invalid("/links", "invalid_value", "a question can link to a resource, a project or an article.")
		}
		if !validID(l.ID) {
			return nil, invalid("/links", "invalid_value", "link id must be a UUID.")
		}
		if !slices.Contains(links, l) {
			links = append(links, l)
		}
	}
	return links, nil
}

func validateSources(in []string) ([]string, error) {
	for _, u := range in {
		if !validText(u) || (!strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://")) {
			return nil, invalid("/sources", "invalid_format", "each source must be an http or https URL.")
		}
	}
	return dedupe(in), nil
}

// usableText is valid text that is not only whitespace.
func usableText(s string) bool { return validText(s) && strings.TrimSpace(s) != "" }

// validText refuses what PostgreSQL text cannot hold (NUL, invalid UTF-8) rather than answer 500.
func validText(s string) bool { return utf8.ValidString(s) && !strings.ContainsRune(s, 0) }

func validID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func dedupe[T comparable](in []T) []T {
	out := make([]T, 0, len(in))
	for _, v := range in {
		if !slices.Contains(out, v) {
			out = append(out, v)
		}
	}
	return out
}

// encodeQuestionSort packs the relevance rank (a float32, printed exactly) and the creation time.
func encodeQuestionSort(q domain.Question) string {
	return strconv.FormatFloat(float64(q.Rank), 'g', -1, 32) + ":" + q.CreatedAt.UTC().Format(time.RFC3339Nano)
}

func decodeQuestionSort(sort string) (float32, time.Time, bool) {
	rank, rest, found := strings.Cut(sort, ":")
	r, errRank := strconv.ParseFloat(rank, 32)
	at, errAt := time.Parse(time.RFC3339Nano, rest)
	if !found || errRank != nil || errAt != nil || math.IsNaN(r) || math.IsInf(r, 0) {
		return 0, time.Time{}, false
	}
	return float32(r), at, true
}

func encodeAnswerKey(a domain.Answer, hash string) cursor.Payload {
	rank := 1
	if a.Accepted {
		rank = 0
	}
	return cursor.Payload{
		Sort:       fmt.Sprintf("%d:%d:%s", rank, -a.VoteScore, a.CreatedAt.UTC().Format(time.RFC3339Nano)),
		ID:         a.ID,
		FilterHash: hash,
	}
}

func decodeAnswerKey(p cursor.Payload) (domain.AnswerKey, bool) {
	parts := strings.SplitN(p.Sort, ":", 3)
	if len(parts) != 3 || !validID(p.ID) {
		return domain.AnswerKey{}, false
	}
	rank, errRank := strconv.Atoi(parts[0])
	neg, errNeg := strconv.Atoi(parts[1])
	at, errAt := time.Parse(time.RFC3339Nano, parts[2])
	if errRank != nil || errNeg != nil || errAt != nil || (rank != 0 && rank != 1) {
		return domain.AnswerKey{}, false
	}
	return domain.AnswerKey{Rank: rank, NegScore: neg, CreatedAt: at, ID: p.ID}, true
}

func textFilter(q *string) (string, error) {
	if q == nil {
		return "", nil
	}
	n := utf8.RuneCountInString(*q)
	if n < 1 || n > maxQuery {
		return "", invalid("q", "out_of_range", "q must be between 1 and 200 characters.")
	}
	if !validText(*q) {
		return "", invalid("q", "invalid_value", "q must be valid UTF-8 text.")
	}
	return *q, nil
}

func pageSize(v *int) (int, error) {
	if v == nil {
		return defaultLimit, nil
	}
	if *v < 1 || *v > maxLimit {
		return 0, invalid("limit", "out_of_range", "limit must be between 1 and 50.")
	}
	return *v, nil
}

// lengthPrefixed binds a cursor to its filters; every field is length-prefixed so no value can
// imitate a boundary.
func lengthPrefixed(fields ...string) string {
	var b strings.Builder
	for _, f := range fields {
		b.WriteString(strconv.Itoa(len(f)))
		b.WriteByte(':')
		b.WriteString(f)
	}
	return b.String()
}

// readErr makes a missing and a removed question one indistinguishable 404.
func readErr(err error) error {
	if errors.Is(err, domain.ErrNotFound) {
		return apperr.New(apperr.CodeNotFound, "Not found.")
	}
	return apperr.Wrap(apperr.CodeInternal, err)
}

// writeErr maps the store's write errors: a link target or country that cannot be used is a 422.
func writeErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrLinkTarget):
		return invalid("/links", "invalid_value", "each link must point to a published, public item.")
	case errors.Is(err, domain.ErrTagUnavailable):
		return invalid("/tags", "invalid_value", "each tag must be a tag you can use.")
	case errors.Is(err, domain.ErrUnknownCountry):
		return invalid("/country", "invalid_value", "country must be a known country code.")
	}
	return readErr(err)
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// invalid is a 422 with one violation: field is the JSON pointer of a body member or the name of a query
// parameter, and code is one of the Violation.code enum values of api/openapi.yaml.
func invalid(field, code, detail string) error {
	return apperr.New(apperr.CodeValidation, detail).WithViolations(apperr.Violation{Field: field, Code: code})
}

func invalidCursor() error {
	return apperr.New(apperr.CodeInvalidCursor, "The cursor is invalid or expired.")
}
