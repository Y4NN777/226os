package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Production endpoint and adapter limits. The base URL is a constant on purpose: it can only be
// replaced through WithBaseURL (tests), never from the environment, so a misconfiguration cannot
// redirect the API key to another host.
const (
	aiStudioBaseURL = "https://generativelanguage.googleapis.com"
	defaultTimeout  = 30 * time.Second
	dialTimeout     = 10 * time.Second
	maxBodyBytes    = 1 << 20
	listPageSize    = "200"
	maxListPages    = 5
	keyHeader       = "x-goog-api-key"
	redacted        = "[redacted]"
)

var errRedirect = errors.New("redirect refused")

// AIStudioConfig configures the Google AI Studio adapter. APIKey is a secret: it never appears in
// String, GoString, LogValue, %v, %+v or %#v, nor in any error.
type AIStudioConfig struct {
	APIKey  string
	Model   string        // validated against ValidModelName on use; there is no default
	Timeout time.Duration // per call, 30 s when not positive
}

// LogValue keeps slog from printing the key.
func (c AIStudioConfig) LogValue() slog.Value {
	return slog.GroupValue(slog.String("APIKey", redacted), slog.String("Model", c.Model), slog.Duration("Timeout", c.Timeout))
}

// GoString is the redacted form used by %#v.
func (c AIStudioConfig) GoString() string { return c.String() }

// String is the redacted form used by %v and %+v.
func (c AIStudioConfig) String() string {
	return fmt.Sprintf("{APIKey:%s Model:%s Timeout:%s}", redacted, c.Model, c.Timeout)
}

// ModelInfo describes one model returned by ListModels.
type ModelInfo struct {
	Name             string // without the models/ prefix
	DisplayName      string
	SupportedMethods []string
	InputTokenLimit  int
	OutputTokenLimit int
}

// Option customizes NewAIStudio. Options exist for tests only.
type Option func(*AIStudio)

// WithBaseURL replaces the production endpoint (for example with an httptest server on loopback).
func WithBaseURL(base string) Option {
	return func(a *AIStudio) { a.base = strings.TrimRight(base, "/") }
}

// AIStudio is the Client backed by the Google AI Studio (Gemini API) REST endpoint. It also lists
// the models available to the key.
type AIStudio struct {
	cfg  AIStudioConfig
	base string
	hc   *http.Client
}

var _ Client = (*AIStudio)(nil)

// NewAIStudio builds the adapter. It performs no I/O and never fails: a missing key or an invalid
// model name is reported by each call, before any request is sent.
func NewAIStudio(cfg AIStudioConfig, opts ...Option) *AIStudio {
	if cfg.Timeout <= 0 {
		cfg.Timeout = defaultTimeout
	}
	a := &AIStudio{cfg: cfg, base: aiStudioBaseURL, hc: newHTTPClient(cfg.Timeout)}
	for _, o := range opts {
		o(a)
	}
	return a
}

// newHTTPClient clones the default transport (TLS verification untouched) with dial and
// response-header timeouts, and refuses redirects: following one would forward the key.
func newHTTPClient(timeout time.Duration) *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.DialContext = (&net.Dialer{Timeout: dialTimeout, KeepAlive: 30 * time.Second}).DialContext
	tr.ResponseHeaderTimeout = timeout + time.Second
	return &http.Client{
		Transport:     tr,
		CheckRedirect: func(*http.Request, []*http.Request) error { return errRedirect },
	}
}

// LogValue keeps slog from printing the key.
func (a AIStudio) LogValue() slog.Value { return a.cfg.LogValue() }

// GoString is the redacted form used by %#v.
func (a AIStudio) GoString() string { return a.String() }

// String is the redacted form used by %v and %+v.
func (a AIStudio) String() string { return "AIStudio" + a.cfg.String() }

type part struct {
	Text    string `json:"text"`
	Thought bool   `json:"thought,omitempty"` // a reasoning summary the provider marks itself
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type generateRequest struct {
	SystemInstruction *content  `json:"systemInstruction,omitempty"`
	Contents          []content `json:"contents"`
	GenerationConfig  struct {
		MaxOutputTokens int `json:"maxOutputTokens"`
	} `json:"generationConfig"`
}

type generateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []part `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	PromptFeedback struct {
		BlockReason string `json:"blockReason"`
	} `json:"promptFeedback"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

// Generate sends one single-turn request. See the package documentation for the error contract.
func (a *AIStudio) Generate(ctx context.Context, r Request) (Response, error) {
	r, err := normalize(r)
	if err != nil {
		return Response{}, err
	}
	if err := a.ready(); err != nil {
		return Response{}, err
	}
	if !ValidModelName(a.cfg.Model) {
		return Response{}, fmt.Errorf("%w: invalid model name", ErrRejected)
	}
	payload := generateRequest{Contents: []content{{Role: "user", Parts: []part{{Text: r.User}}}}}
	if r.System != "" {
		payload.SystemInstruction = &content{Parts: []part{{Text: r.System}}}
	}
	payload.GenerationConfig.MaxOutputTokens = r.MaxOutputTokens
	body, _ := json.Marshal(payload) // cannot fail: strings and ints only

	ctx, cancel := context.WithTimeout(ctx, a.cfg.Timeout)
	defer cancel()
	raw, err := a.do(ctx, http.MethodPost, a.base+"/v1beta/models/"+a.cfg.Model+":generateContent", body)
	if err != nil {
		return Response{}, err
	}
	var g generateResponse
	if err := json.Unmarshal(raw, &g); err != nil {
		return Response{}, unavailable("invalid response")
	}
	text, err := answerText(g)
	if err != nil {
		return Response{}, err
	}
	return Response{
		Text: text, Model: a.cfg.Model,
		PromptTokens: g.UsageMetadata.PromptTokenCount, OutputTokens: g.UsageMetadata.CandidatesTokenCount,
	}, nil
}

// answerText extracts the first candidate text, or says why there is none.
func answerText(g generateResponse) (string, error) {
	switch {
	case g.PromptFeedback.BlockReason != "":
		return "", unavailable("prompt blocked")
	case len(g.Candidates) == 0:
		return "", unavailable("no candidate")
	case g.Candidates[0].FinishReason == "SAFETY":
		return "", unavailable("output blocked")
	}
	var sb strings.Builder
	for _, p := range g.Candidates[0].Content.Parts {
		if !p.Thought {
			sb.WriteString(p.Text)
		}
	}
	text := CleanAnswer(sb.String())
	if text == "" {
		return "", unavailable("empty output")
	}
	return text, nil
}

// ListModels returns the models visible to the key, following nextPageToken for at most five
// pages of 200.
func (a *AIStudio) ListModels(ctx context.Context) ([]ModelInfo, error) {
	if err := a.ready(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, a.cfg.Timeout)
	defer cancel()
	var out []ModelInfo
	token := ""
	for range maxListPages {
		q := url.Values{"pageSize": {listPageSize}}
		if token != "" {
			q.Set("pageToken", token)
		}
		raw, err := a.do(ctx, http.MethodGet, a.base+"/v1beta/models?"+q.Encode(), nil)
		if err != nil {
			return nil, err
		}
		var page struct {
			Models []struct {
				Name                       string   `json:"name"`
				DisplayName                string   `json:"displayName"`
				SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
				InputTokenLimit            int      `json:"inputTokenLimit"`
				OutputTokenLimit           int      `json:"outputTokenLimit"`
			} `json:"models"`
			NextPageToken string `json:"nextPageToken"`
		}
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, unavailable("invalid response")
		}
		for _, m := range page.Models {
			out = append(out, ModelInfo{
				Name: strings.TrimPrefix(m.Name, "models/"), DisplayName: m.DisplayName,
				SupportedMethods: m.SupportedGenerationMethods,
				InputTokenLimit:  m.InputTokenLimit, OutputTokenLimit: m.OutputTokenLimit,
			})
		}
		if token = page.NextPageToken; token == "" {
			break
		}
	}
	return out, nil
}

func (a *AIStudio) ready() error {
	if a.cfg.APIKey == "" {
		return ErrNotConfigured
	}
	return nil
}

// do sends one request and returns the body of an HTTP 200 answer, read through a size limit.
// Every error is classified and carries a status and a fixed phrase only.
func (a *AIStudio) do(ctx context.Context, method, target string, body []byte) ([]byte, error) {
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, rd)
	if err != nil {
		return nil, unavailable("bad endpoint") // the parse error would echo the URL
	}
	req.Header.Set(keyHeader, a.cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := a.hc.Do(req)
	if err != nil {
		return nil, transportError(ctx, "request failed") // *url.Error text is not forwarded
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusTooManyRequests:
		return nil, &StatusError{Kind: ErrQuotaExhausted, Status: resp.StatusCode}
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return nil, &StatusError{Kind: ErrRejected, Status: resp.StatusCode}
	default:
		return nil, &StatusError{Kind: ErrUnavailable, Status: resp.StatusCode}
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes+1))
	if err != nil {
		return nil, transportError(ctx, "reading the response failed")
	}
	if len(raw) > maxBodyBytes {
		return nil, unavailable("response too large")
	}
	return raw, nil
}

func unavailable(phrase string) error { return fmt.Errorf("%w: %s", ErrUnavailable, phrase) }

// transportError is ErrUnavailable, and also wraps the context error when the deadline or a
// cancellation caused the failure.
func transportError(ctx context.Context, phrase string) error {
	if cerr := ctx.Err(); cerr != nil {
		return fmt.Errorf("%w: %s: %w", ErrUnavailable, phrase, cerr)
	}
	return unavailable(phrase)
}
