// Package service is the grounded assistant use case: authorize, limit, retrieve, ask the model
// once, validate its citations, and answer or refuse. It never stores anything and never logs the
// question, the passages or the answer.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Y4NN777/226os/src/api/internal/assistant/domain"
	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

const (
	// The per-member budget is: 10 questions per hour (scope assistant:user).
	userScope  = "assistant:user"
	userLimit  = 10
	userWindow = time.Hour
	// The global guard counts only calls that reach the provider. ADR-0012: the shared limiter keeps
	// buckets for one hour at most, so the daily provider budget (300 calls a day) is held as 12 calls
	// per hour, 288 a day.
	globalScope  = "assistant:global"
	globalKey    = "provider"
	globalLimit  = 12
	globalWindow = time.Hour

	// DefaultTimeout is the deadline of the provider call, inside the 30 s route budget (B3).
	DefaultTimeout = 20 * time.Second
	// providerRetryAfter is the Retry-After of a 503 caused by the provider (API_CONTRACT).
	providerRetryAfter = 30 * time.Second
	// PassageLimit is how many passages are retrieved and sent.
	PassageLimit = 5

	minQuestion = 3
	maxQuestion = 1000
	maxTokens   = 768
)

// Retriever finds the published, visible, public passages for a question. The service owns this port.
type Retriever interface {
	Retrieve(ctx context.Context, question string, limit int) ([]domain.Passage, error)
}

// Limiter is the fixed-window rate limiter. The service owns this port.
type Limiter interface {
	Hit(ctx context.Context, scope, value string, limit int, window time.Duration) (allowed bool, retryAfter time.Duration, err error)
}

// Config is the per-environment setting. Enabled false answers 503 before anything is counted or sent.
type Config struct {
	Enabled bool
	Timeout time.Duration
}

// Request is the validated shape of the request body. Lang is empty when the client gave none.
type Request struct {
	Question string
	Lang     string
}

// Service answers questions from published content.
type Service struct {
	cfg       Config
	retriever Retriever
	model     llm.Client
	limiter   Limiter
	logger    *slog.Logger
	now       func() time.Time
}

// New returns the service. A zero Timeout means DefaultTimeout; a nil logger discards logs.
func New(cfg Config, retriever Retriever, model llm.Client, limiter Limiter, logger *slog.Logger) *Service {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	return &Service{cfg: cfg, retriever: retriever, model: model, limiter: limiter, logger: logger, now: time.Now}
}

// Ask answers one question for the session subject. Order: authentication, verified email, input,
// enable switch, per-member limit, retrieval, global provider guard, one provider call, citation check.
func (s *Service) Ask(ctx context.Context, sub subject.Subject, req Request) (domain.Answer, error) {
	if sub.Anonymous() {
		return domain.Answer{}, apperr.New(apperr.CodeUnauthenticated, "Sign in to use the assistant.")
	}
	if !sub.EmailVerified {
		return domain.Answer{}, apperr.New(apperr.CodeEmailNotVerified, "Verify your email address to use the assistant.")
	}
	question, lang, err := validate(req)
	if err != nil {
		return domain.Answer{}, err
	}
	if !s.cfg.Enabled {
		return domain.Answer{}, unavailable(nil, providerRetryAfter)
	}
	if err := s.allow(ctx, userScope, sub.UserID, userLimit, userWindow, apperr.CodeRateLimited); err != nil {
		return domain.Answer{}, err
	}
	passages, err := s.retriever.Retrieve(ctx, question, PassageLimit)
	if err != nil {
		return domain.Answer{}, apperr.Wrap(apperr.CodeInternal, err)
	}
	if len(passages) == 0 { // AI-1: nothing to ground on, the provider is not called
		return domain.NoSourceAnswer(lang, question), nil
	}
	if err := s.allow(ctx, globalScope, globalKey, globalLimit, globalWindow, apperr.CodeServiceUnavailable); err != nil {
		return domain.Answer{}, err
	}
	return s.generate(ctx, question, lang, passages)
}

// allow takes one hit. A limiter failure fails closed with 503 and the provider Retry-After; a spent budget is 429 for the member
// bucket and 503 for the global guard, both with Retry-After.
func (s *Service) allow(ctx context.Context, scope, key string, limit int, window time.Duration, spent apperr.Code) error {
	ok, retry, err := s.limiter.Hit(ctx, scope, key, limit, window)
	switch {
	case err != nil:
		return unavailable(err, providerRetryAfter)
	case !ok && spent == apperr.CodeRateLimited:
		return apperr.New(spent, "Too many requests.").WithRetryAfter(retry)
	case !ok:
		return unavailable(nil, retry)
	}
	return nil
}

func (s *Service) generate(ctx context.Context, question, lang string, passages []domain.Passage) (domain.Answer, error) {
	callCtx, cancel := context.WithTimeout(ctx, s.cfg.Timeout)
	defer cancel()
	start := s.now()
	resp, err := s.model.Generate(callCtx, llm.Request{
		System: domain.SystemPrompt(lang), User: domain.UserPrompt(question, passages), MaxOutputTokens: maxTokens,
	})
	if err != nil { // never retried; the Tracker already logged the class, status and latency
		if errors.Is(err, llm.ErrInvalidRequest) {
			return domain.Answer{}, apperr.Wrap(apperr.CodeInternal, err)
		}
		return domain.Answer{}, unavailable(err, providerRetryAfter)
	}
	text, removed := domain.RemoveGuarantees(llm.CleanAnswer(resp.Text)) // AI-3
	if removed {
		s.logger.Info("assistant sentence removed", "class", "guarantee_removed")
	}
	answer, err := domain.Parse(text, passages, lang)
	switch {
	case errors.Is(err, domain.ErrNoCitation), errors.Is(err, domain.ErrInvalidCitation): // AI-2: refuse
		s.logger.Info("assistant refused", "class", refusalClass(err), "duration_ms", s.now().Sub(start).Milliseconds())
		return domain.NoSourceAnswer(lang, question), nil
	case err != nil:
		return domain.Answer{}, unavailable(err, providerRetryAfter)
	}
	s.logger.Info("assistant answered", "sources", len(answer.Sources), "duration_ms", s.now().Sub(start).Milliseconds())
	return answer, nil
}

func refusalClass(err error) string {
	if errors.Is(err, domain.ErrInvalidCitation) {
		return "invalid_citation"
	}
	return "no_citation"
}

func unavailable(cause error, retry time.Duration) error {
	e := apperr.New(apperr.CodeServiceUnavailable, "The assistant is unavailable. Search still works.")
	if cause != nil {
		e = apperr.Wrap(apperr.CodeServiceUnavailable, cause)
		e.Detail = "The assistant is unavailable. Search still works."
	}
	return e.WithRetryAfter(retry)
}

// validate trims the question and checks the contract bounds (3 to 1000 characters, valid text).
func validate(req Request) (question, lang string, err error) {
	question = strings.TrimSpace(req.Question)
	n := utf8.RuneCountInString(question)
	if n < minQuestion || n > maxQuestion || !utf8.ValidString(question) || strings.ContainsRune(question, 0) {
		return "", "", apperr.New(apperr.CodeValidation, fmt.Sprintf("question must be between %d and %d characters.", minQuestion, maxQuestion)).
			WithViolations(apperr.Violation{Field: "question", Code: "out_of_range"})
	}
	lang = req.Lang
	if lang == "" {
		lang = domain.DetectLang(question)
	}
	return question, lang, nil
}
