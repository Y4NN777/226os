package app

import (
	"log/slog"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/assistant"
	assistantservice "github.com/Y4NN777/226os/src/api/internal/assistant/service"
	assistanttransport "github.com/Y4NN777/226os/src/api/internal/assistant/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	searchrepository "github.com/Y4NN777/226os/src/api/internal/search/repository"
	searchservice "github.com/Y4NN777/226os/src/api/internal/search/service"
)

// assistantModel builds the provider client from the AI config. Without config (AI_ENABLED false)
// it is the Disabled client and on is false. The provider deadline is AI_TIMEOUT capped at 20 s, so
// it always ends inside the 30 s route budget (ADR-0012). Calls are counted by a Tracker.
func assistantModel(ai *config.AI, logger *slog.Logger) (model llm.Client, on bool, timeout time.Duration) {
	if ai == nil {
		return llm.Disabled{}, false, assistantservice.DefaultTimeout
	}
	timeout = min(ai.Timeout, assistantservice.DefaultTimeout)
	client := llm.NewAIStudio(llm.AIStudioConfig{APIKey: ai.APIKey, Model: ai.Model, Timeout: timeout})
	return llm.NewTracker(logger).Wrap(ai.Model, client), true, timeout
}

// newAssistantHandler wires the assistant over the platform search, the shared limiter and the
// provider client chosen by config.
func newAssistantHandler(ai *config.AI, db dbgen.DBTX, limiter assistantservice.Limiter, logger *slog.Logger) *assistanttransport.Handler {
	model, on, timeout := assistantModel(ai, logger)
	search := searchservice.New(searchrepository.New(db), limiter, logger)
	return assistant.NewHandler(on, timeout, search, model, limiter, logger)
}
