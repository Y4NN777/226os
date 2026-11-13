package app

import (
	"net/http"

	"github.com/Y4NN777/226os/src/api/internal/discover/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
	searchtransport "github.com/Y4NN777/226os/src/api/internal/search/transport"
)

// Server is the ONLY production apigen.ServerInterface. Every operation answers
// 501 until its slice implements it by overriding the method on this type;
// no other constructor of a ServerInterface may exist in production code.
// Besides the operations it exposes one accessor, Mailer, for the session mail flows (S1).
type Server struct {
	apigen.Unimplemented
	mailer    mailer.Mailer
	catalogue *transport.Handler
	search    *searchtransport.Handler
}

// Option configures NewServer.
type Option func(*Server)

// WithMailer sets the mail port the operations use.
func WithMailer(m mailer.Mailer) Option { return func(s *Server) { s.mailer = m } }

// WithCatalogue sets the published-catalogue reads. Without it those routes stay 501.
func WithCatalogue(h *transport.Handler) Option { return func(s *Server) { s.catalogue = h } }

// WithSearch sets the search read. Without it GET /search stays unimplemented.
func WithSearch(h *searchtransport.Handler) Option { return func(s *Server) { s.search = h } }

// NewServer returns the production server with all operations unimplemented.
// Without WithMailer its mailer is mailer.Disabled.
func NewServer(opts ...Option) *Server {
	s := &Server{mailer: mailer.Disabled{}}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Mailer is the mail port: operations send mail through it and never import net/smtp.
func (s *Server) Mailer() mailer.Mailer { return s.mailer }

// ListResources serves the published catalogue. A server built without WithCatalogue stays unimplemented.
func (s *Server) ListResources(w http.ResponseWriter, r *http.Request, params apigen.ListResourcesParams) {
	if s.catalogue == nil {
		s.Unimplemented.ListResources(w, r, params)
		return
	}
	s.catalogue.ListResources(w, r, params)
}

// GetResource serves one published resource. A server built without WithCatalogue stays unimplemented.
func (s *Server) GetResource(w http.ResponseWriter, r *http.Request, resourceID apigen.ResourceId) {
	if s.catalogue == nil {
		s.Unimplemented.GetResource(w, r, resourceID)
		return
	}
	s.catalogue.GetResource(w, r, resourceID)
}

// ListResourceSources serves provenance for one published resource.
func (s *Server) ListResourceSources(w http.ResponseWriter, r *http.Request, resourceID apigen.ResourceId) {
	if s.catalogue == nil {
		s.Unimplemented.ListResourceSources(w, r, resourceID)
		return
	}
	s.catalogue.ListResourceSources(w, r, resourceID)
}

// Search serves the grouped search. A server built without WithSearch stays unimplemented.
func (s *Server) Search(w http.ResponseWriter, r *http.Request, params apigen.SearchParams) {
	if s.search == nil {
		s.Unimplemented.Search(w, r, params)
		return
	}
	s.search.Search(w, r, params)
}
