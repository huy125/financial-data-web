package api

import (
	"context"
	"log"
	"net/http"

	model "github.com/huy125/financial-data-web/api/models"
)

type InMemoryStore interface {
	Create(user model.User) model.User
	List() []model.User
}

// Server is the API server.
type Server struct {
	h 		http.Handler

	apiKey 		string
	store	InMemoryStore
}

// New creates a new API server.
func New(apiKey string, store InMemoryStore) *Server {
	s := &Server{
		apiKey: 	apiKey,
		store: 	store,
	}

	s.h = s.routes()
	s.srv = &http.Server{
		Addr: ":8080",
		Handler: s.h,
	}

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

	userHandler := &UserHandler{Store: s.store}
	mux.HandleFunc("/users", userHandler.CreateUserHandler)

	return mux
}

func (s *Server) Start() error {
	log.Println("Starting server on port :8080")
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}
