package app

import "github.com/Y4NN777/226os/src/api/internal/platform/apigen"

// Server is the ONLY production apigen.ServerInterface. Every operation answers
// 501 until its slice implements it by overriding the method on this type;
// no other constructor of a ServerInterface may exist in production code.
type Server struct {
	apigen.Unimplemented
}

// NewServer returns the production server with all operations unimplemented.
func NewServer() *Server { return &Server{} }
