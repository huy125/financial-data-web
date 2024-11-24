package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	model "github.com/huy125/financial-data-web/api/store/models"
)

type UserStore interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]model.User, error)
	Find(ctx context.Context, id uuid.UUID) (*model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
}

// Server is the API server.
type Server struct {
	h http.Handler

	apiKey string
	store  UserStore
}

// New creates a new API server.
func New(apiKey string, store UserStore) *Server {
	s := &Server{
		apiKey: apiKey,
		store:  store,
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

	userHandler := &UserHandler{store: s.store}
	mux.HandleFunc("/users", userHandler.CreateUserHandler)
	mux.HandleFunc("/users/{id}", userHandler.UpdateUserHandler)

	return mux
}
