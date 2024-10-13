package api

import (
	"net/http"

	repository "github.com/huy125/financial-data-web/api/repositories/in-memory"
)

// Server is the API server.
type Server struct {
	h 		http.Handler

	apiKey 		string
	userRepo	repository.UserRepository
}

// New creates a new API server.
func New(apiKey string, userRepo repository.UserRepository) *Server {
	s := &Server{
		apiKey: 	apiKey,
		userRepo: 	userRepo,
	}

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

	userHandler := &UserHandler{Repo: s.userRepo}
	mux.HandleFunc("/users", userHandler.CreateUserHandler)

	return mux
}
