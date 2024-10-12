package api

import (
	"net/http"
)

// Server is the API server.
type Server struct {
	h http.Handler

	apiKey string
}

// New creates a new API server.
func New(apiKey string) *Server {
	s := &Server{apiKey: apiKey}

	s.h = s.routes()

	return s
}

// ServeHTTP serves the API server.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.h.ServeHTTP(w, r)
}

// routes returns the routes for the API server.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.HelloServerHandler)
	mux.HandleFunc("/stocks", s.GetStockBySymbolHandler)
	mux.HandleFunc("/users", s.CreateUserHandler)

	return mux
}
