// Package assistant is the grounded AI assistant (S9, US-INTEL-001): it answers a member's question
// only from published, visible content, cites it, and refuses when there is none.
package assistant

import (
	"context"
	"log/slog"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/assistant/domain"
	"github.com/Y4NN777/226os/src/api/internal/assistant/service"
	"github.com/Y4NN777/226os/src/api/internal/assistant/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/llm"
	searchdomain "github.com/Y4NN777/226os/src/api/internal/search/domain"
	searchservice "github.com/Y4NN777/226os/src/api/internal/search/service"
)

// Searcher is the part of the search service the assistant reads through.
type Searcher interface {
	Search(ctx context.Context, req searchservice.Request) (searchservice.Results, error)
}

// Retriever adapts the platform search (published, visible, public content only) to the service port.
type Retriever struct{ search Searcher }

// NewRetriever returns a retriever over s.
func NewRetriever(s Searcher) Retriever { return Retriever{search: s} }

// kinds maps a search group to the passage kind and the in-app URL of an item of that group. Series
// are not retrieved, so they have no entry.
var kinds = map[searchdomain.Group]struct {
	kind string
	url  func(id string) string
}{
	searchdomain.Questions: {domain.KindQuestion, func(id string) string { return "/questions/" + id }},
	searchdomain.Resources: {domain.KindResource, func(id string) string { return "/discover/" + id }},
	searchdomain.Projects:  {domain.KindProject, func(id string) string { return "/build/" + id }},
	searchdomain.Learn:     {domain.KindArticle, func(id string) string { return "/apprendre/articles/" + id }},
}

// Retrieve runs the search once as a member (no anonymous limit) and keeps the best limit hits,
// taking one per group in turn so one content type cannot crowd out the others. Each passage is the
// search snippet, cut to MaxPassageRunes.
func (r Retriever) Retrieve(ctx context.Context, question string, limit int) ([]domain.Passage, error) {
	q := domain.SearchQuery(question)
	if q == "" {
		return nil, nil
	}
	per := limit
	res, err := r.search.Search(ctx, searchservice.Request{Q: q, PerGroup: &per})
	if err != nil {
		return nil, err
	}
	var out []domain.Passage
	for round := 0; len(out) < limit; round++ {
		added := false
		for _, g := range searchdomain.Groups() {
			items := res.Groups[g].Items
			if round >= len(items) || len(out) == limit {
				continue
			}
			h, k := items[round], kinds[g]
			out = append(out, domain.Passage{
				Type: k.kind, ID: h.ID, Title: h.Title, Text: domain.Cut(h.Snippet, domain.MaxPassageRunes), URL: k.url(h.ID),
				PublishedAt: h.PublishedAt,
			})
			added = true
		}
		if !added {
			break
		}
	}
	return out, nil
}

// NewHandler wires POST /assistant/answers. enabled false (AI off in this environment) keeps the
// route answering 503. A zero timeout means the 20 s default.
func NewHandler(enabled bool, timeout time.Duration, s Searcher, model llm.Client, limiter service.Limiter, logger *slog.Logger) *transport.Handler {
	svc := service.New(service.Config{Enabled: enabled, Timeout: timeout}, NewRetriever(s), model, limiter, logger)
	return transport.New(svc, logger)
}
