// Package transport is the HTTP edge of the assistant: decode, call one service method, encode.
package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/Y4NN777/226os/src/api/internal/assistant/domain"
	"github.com/Y4NN777/226os/src/api/internal/assistant/service"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/subject"
)

// Handler serves POST /assistant/answers.
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

// AskAssistant serves POST /assistant/answers. The subject comes from the session only.
func (h *Handler) AskAssistant(w http.ResponseWriter, r *http.Request) {
	var body apigen.AssistantQuestion
	if err := httpx.DecodeJSON(w, r, &body); err != nil {
		h.fail(w, r, err)
		return
	}
	req := service.Request{Question: body.Question}
	if body.Lang != nil {
		req.Lang = string(*body.Lang)
	}
	a, err := h.svc.Ask(r.Context(), subject.From(r.Context()), req)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(answerJSON(a))
}

// answerBody and sourceBody shadow the generated fields that cannot say null: an undated source
// is "published_at": null (the generated pointer would omit it).
type answerBody struct {
	apigen.AssistantAnswer
	Sources []sourceBody `json:"sources"`
}

type sourceBody struct {
	apigen.AssistantSource
	PublishedAt *time.Time `json:"published_at"`
}

func answerJSON(a domain.Answer) answerBody {
	out := apigen.AssistantAnswer{
		Outcome: apigen.AssistantAnswerOutcomeNoSource, GeneratedBy: apigen.Ai, Answer: a.Text,
		HowToVerify: a.HowToVerify, Lang: apigen.Lang(a.Lang),
	}
	sources := []sourceBody{}
	if a.Answered {
		out.Outcome = apigen.AssistantAnswerOutcomeAnswered
	} else {
		search := a.SearchURL
		out.SearchUrl = &search
	}
	for _, p := range a.Sources {
		src := sourceBody{AssistantSource: apigen.AssistantSource{
			Id: uuid.MustParse(p.ID), Type: apigen.AssistantSourceType(p.Type), Title: p.Title, Url: p.URL,
		}}
		if !p.PublishedAt.IsZero() {
			published := p.PublishedAt.UTC()
			src.PublishedAt = &published
		}
		sources = append(sources, src)
	}
	return answerBody{AssistantAnswer: out, Sources: sources}
}

// fail writes any error as problem+json; an error that is not an *apperr.Error becomes a 500.
func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	httpx.WriteError(h.logger)(w, r, err)
}
