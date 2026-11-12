// Package aistest is an in-process fake of the two Google AI Studio endpoints the llm adapter
// uses (generateContent and models.list), for tests of llm, app and cmd/aicheck. It binds to
// 127.0.0.1 only and makes no external network call.
package aistest

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// Token limits and display name prefix reported for every model by ModelsBody.
const (
	ModelInputLimit  = 131072
	ModelOutputLimit = 8192
	DisplayPrefix    = "Display "
)

// Behaviour scripts one response. The zero value is an empty 200 reply.
type Behaviour struct {
	Status   int           // HTTP status, 200 when zero
	Body     string        // JSON body
	Delay    time.Duration // wait before answering; ends early when the client disconnects
	Redirect string        // when set, answer 307 with this Location and no body
	Endless  bool          // stream filler bytes until the client disconnects
	Truncate bool          // declare a longer Content-Length than Body, so the client read fails
}

// Recorded is one request seen by the server.
type Recorded struct {
	Method string
	Path   string
	Query  string // raw query string
	Header http.Header
	Body   string
}

// Option configures NewServer.
type Option func(*Server)

// WithGenerate scripts the generateContent endpoint: behaviours are consumed in order and the last
// one repeats. Without it the endpoint answers 200 with a short text.
func WithGenerate(b ...Behaviour) Option { return func(s *Server) { s.generate.script = b } }

// WithModels scripts the models.list endpoint the same way. Without it the endpoint answers 200
// with a single page of one model.
func WithModels(b ...Behaviour) Option { return func(s *Server) { s.models.script = b } }

// Server is the fake. Create it with NewServer; it is closed through t.Cleanup.
type Server struct {
	ts       *httptest.Server
	generate endpoint
	models   endpoint

	mu   sync.Mutex
	reqs []Recorded
}

type endpoint struct {
	script []Behaviour
	next   int // guarded by Server.mu
}

// NewServer starts the fake and registers its shutdown on t.
func NewServer(t testing.TB, opts ...Option) *Server {
	t.Helper()
	s := &Server{}
	for _, o := range opts {
		o(s)
	}
	if len(s.generate.script) == 0 {
		s.generate.script = []Behaviour{{Body: AnswerBody("pong", 7, 3)}}
	}
	if len(s.models.script) == 0 {
		s.models.script = []Behaviour{{Body: ModelsBody("", "gemma-test")}}
	}
	s.ts = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.ts.Close)
	return s
}

// URL is the base URL (http://127.0.0.1:port) to pass to llm.WithBaseURL.
func (s *Server) URL() string { return s.ts.URL }

// Requests returns the requests received so far, in order.
func (s *Server) Requests() []Recorded {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Recorded, len(s.reqs))
	for i, r := range s.reqs {
		r.Header = r.Header.Clone()
		out[i] = r
	}
	return out
}

func (s *Server) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // a failed read only shortens the recorded body
	s.mu.Lock()
	s.reqs = append(s.reqs, Recorded{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery, Header: r.Header.Clone(), Body: string(body)})
	var ep *endpoint
	switch {
	case r.URL.Path == "/v1beta/models":
		ep = &s.models
	case strings.HasPrefix(r.URL.Path, "/v1beta/models/") && strings.HasSuffix(r.URL.Path, ":generateContent"):
		ep = &s.generate
	}
	var b Behaviour
	if ep != nil {
		b = ep.script[min(ep.next, len(ep.script)-1)]
		ep.next++
	} else {
		b = Behaviour{Status: http.StatusNotFound}
	}
	s.mu.Unlock()
	play(w, r, b)
}

func play(w http.ResponseWriter, r *http.Request, b Behaviour) {
	if b.Delay > 0 {
		select {
		case <-time.After(b.Delay):
		case <-r.Context().Done():
			return
		}
	}
	if b.Redirect != "" {
		w.Header().Set("Location", b.Redirect)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if b.Status == 0 {
		b.Status = http.StatusOK
	}
	w.Header().Set("Content-Type", "application/json")
	if b.Truncate {
		w.Header().Set("Content-Length", strconv.Itoa(len(b.Body)+100))
	}
	w.WriteHeader(b.Status)
	if b.Endless {
		filler := []byte(strings.Repeat(" ", 32<<10))
		for r.Context().Err() == nil {
			_, _ = w.Write(filler) // errors end with the client connection
		}
		return
	}
	_, _ = io.WriteString(w, b.Body)
}

// AnswerBody is a generateContent success body with one candidate.
func AnswerBody(text string, promptTokens, outputTokens int) string {
	return mustJSON(map[string]any{
		"candidates": []any{map[string]any{
			"content":      map[string]any{"role": "model", "parts": []any{map[string]any{"text": text}}},
			"finishReason": "STOP",
		}},
		"usageMetadata": map[string]any{"promptTokenCount": promptTokens, "candidatesTokenCount": outputTokens},
	})
}

// ModelsBody is a models.list page for the given model ids (without the models/ prefix). A
// non-empty next sets nextPageToken.
func ModelsBody(next string, names ...string) string {
	models := make([]any, 0, len(names))
	for _, n := range names {
		models = append(models, map[string]any{
			"name":                       "models/" + n,
			"displayName":                DisplayPrefix + n,
			"supportedGenerationMethods": []string{"generateContent", "countTokens"},
			"inputTokenLimit":            ModelInputLimit,
			"outputTokenLimit":           ModelOutputLimit,
		})
	}
	page := map[string]any{"models": models}
	if next != "" {
		page["nextPageToken"] = next
	}
	return mustJSON(page)
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v) // cannot fail: callers pass maps of strings, numbers and slices only
	return string(b)
}
