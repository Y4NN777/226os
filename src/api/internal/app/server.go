package app

import (
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/mailer"
)

// Server is the ONLY production apigen.ServerInterface. Every operation answers
// 501 until its slice implements it by overriding the method on this type;
// no other constructor of a ServerInterface may exist in production code.
// Besides the operations it exposes one accessor, Mailer, for the session mail flows (S1).
type Server struct {
	apigen.Unimplemented
	mailer mailer.Mailer
}

// Option configures NewServer.
type Option func(*Server)

// WithMailer sets the mail port the operations use.
func WithMailer(m mailer.Mailer) Option { return func(s *Server) { s.mailer = m } }

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
