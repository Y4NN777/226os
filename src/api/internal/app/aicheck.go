package app

import (
	"context"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/config"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
)

const (
	aiUsage       = "usage: aicheck models | ping [fr|en] | limits | help\n"
	aiHelp        = aiUsage + "  models        list the models whose name contains \"gemma\" (check V1)\n  ping [fr|en] one fixed harmless prompt per language (check V4)\n  limits        three tiny probes, never a loop (check V2)\nIt reads AI_ENABLED, AI_API_KEY, AI_MODEL and AI_TIMEOUT only.\n"
	aiSnippetLen  = 160
	aiPingTokens  = 256
	aiProbeTokens = 8
	aiProbeCount  = 3
	aiProbePrompt = "Reply with the single word: ok"
)

// aiPingPrompts are fixed, harmless questions about HTTP status codes: never user data.
var aiPingPrompts = map[string]string{
	"en": "In one short sentence, what does HTTP status code 404 mean?",
	"fr": "En une courte phrase, que signifie le code d'état HTTP 404 ?",
}

// AICheck runs one of the provider checks owned by the S9 slice: models (V1), ping (V4) and
// limits (V2). It reads the AI_* variables only, through config.LoadAI. It returns 0 on success,
// 2 for a usage or configuration problem, 1 when the provider cannot be reached or refuses. Output
// never contains the API key: provider errors are classified and carry a status and a fixed phrase.
func AICheck(ctx context.Context, args []string, getenv func(string) string, out, errOut io.Writer) int {
	return aiCheck(ctx, args, getenv, out, errOut)
}

// AICheckMain runs AICheck and calls exit only on failure (exit(0) would be a no-op).
func AICheckMain(ctx context.Context, args []string, getenv func(string) string, out, errOut io.Writer, exit func(int)) {
	if code := AICheck(ctx, args, getenv, out, errOut); code != 0 {
		exit(code)
	}
}

// aiCheck is AICheck with adapter options, so tests can point it at an in-process server. The
// production entry point passes none: the endpoint is never taken from the environment.
func aiCheck(ctx context.Context, args []string, getenv func(string) string, out, errOut io.Writer, opts ...llm.Option) int {
	if len(args) == 1 && args[0] == "help" {
		say(out, "%s", aiHelp)
		return 0
	}
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}
	langs := []string{"en", "fr"}
	valid := false
	switch {
	case (cmd == "models" || cmd == "limits" || cmd == "ping") && len(args) == 1:
		valid = true
	case cmd == "ping" && len(args) == 2 && aiPingPrompts[args[1]] != "":
		langs, valid = args[1:], true
	}
	if !valid {
		say(errOut, "%s", aiUsage)
		return exitUsage
	}
	cfg, err := config.LoadAI(getenv)
	if err != nil {
		say(errOut, "ai check: invalid configuration: %v\n", err)
		return exitUsage
	}
	if cfg == nil {
		say(errOut, "ai check: AI is not enabled (set AI_ENABLED=true, AI_API_KEY and AI_MODEL)\n")
		return exitUsage
	}
	client := llm.NewAIStudio(llm.AIStudioConfig{APIKey: cfg.APIKey, Model: cfg.Model, Timeout: cfg.Timeout}, opts...)
	switch cmd {
	case "models":
		return aiModels(ctx, client, out, errOut)
	case "ping":
		return aiPing(ctx, client, langs, out, errOut)
	default:
		return aiLimits(ctx, client, out)
	}
}

// aiModels lists the models whose name contains "gemma" (V1).
func aiModels(ctx context.Context, c *llm.AIStudio, out, errOut io.Writer) int {
	models, err := c.ListModels(ctx)
	if err != nil {
		say(errOut, "ai check: models failed: %v\n", err)
		return exitSendFailed
	}
	found := 0
	for _, m := range models {
		if !strings.Contains(strings.ToLower(m.Name), "gemma") {
			continue
		}
		found++
		say(out, "%s | %s | %s | %d/%d\n", m.Name, m.DisplayName, strings.Join(m.SupportedMethods, ","), m.InputTokenLimit, m.OutputTokenLimit)
	}
	if found == 0 {
		say(out, "no model name contains \"gemma\" (%d models seen)\n", len(models))
	}
	return 0
}

// aiPing sends the fixed prompt once per language, in order, and stops at the first failure (V4).
func aiPing(ctx context.Context, c *llm.AIStudio, langs []string, out, errOut io.Writer) int {
	for _, lang := range langs {
		start := time.Now()
		resp, err := c.Generate(ctx, llm.Request{User: aiPingPrompts[lang], MaxOutputTokens: aiPingTokens})
		if err != nil {
			say(errOut, "ai check: ping %s failed: %v\n", lang, err)
			return exitSendFailed
		}
		say(out, "%s | %d ms | %d | %d | %s\n", lang, time.Since(start).Milliseconds(), resp.PromptTokens, resp.OutputTokens, snippet(resp.Text))
	}
	return 0
}

// aiLimits sends exactly three tiny requests (V2). It never loops or retries to find a limit.
func aiLimits(ctx context.Context, c *llm.AIStudio, out io.Writer) int {
	code := 0
	for i := 1; i <= aiProbeCount; i++ {
		start := time.Now()
		_, err := c.Generate(ctx, llm.Request{User: aiProbePrompt, MaxOutputTokens: aiProbeTokens})
		if err != nil {
			code = exitSendFailed
		}
		say(out, "request %d | %s | %d ms\n", i, probeClass(err), time.Since(start).Milliseconds())
	}
	say(out, "These %d probes do not measure the project limits: read the real limits in the AI Studio console.\n", aiProbeCount)
	return code
}

// probeClass names the HTTP-level class of one probe result.
func probeClass(err error) string {
	if err == nil {
		return "ok"
	}
	label := "error"
	for _, c := range []struct {
		target error
		label  string
	}{
		{llm.ErrQuotaExhausted, "quota exhausted"}, {llm.ErrUnavailable, "unavailable"}, {llm.ErrRejected, "rejected"},
	} {
		if errors.Is(err, c.target) {
			label = c.label
			break
		}
	}
	return label
}

// snippet flattens whitespace and keeps the first 160 characters.
func snippet(text string) string {
	r := []rune(strings.Join(strings.Fields(text), " "))
	return string(r[:min(len(r), aiSnippetLen)])
}
