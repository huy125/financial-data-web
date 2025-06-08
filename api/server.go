package api

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/hamba/cmd/v2/observe"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/huy125/financial-data-web/store"
)

type Store interface {
	CreateUser(ctx context.Context, user *store.CreateUser) (*store.User, error)
	ListUsers(ctx context.Context, limit, offset int) ([]store.User, error)
	FindUser(ctx context.Context, id uuid.UUID) (*store.User, error)
	UpdateUser(ctx context.Context, user *store.UpdateUser) (*store.User, error)
	FindStockBySymbol(ctx context.Context, symbol string) (*store.Stock, error)
	ListMetrics(ctx context.Context, limit, offset int) ([]store.Metric, error)
	CreateStockMetric(ctx context.Context, stockID, metricID uuid.UUID, value float64) (*store.StockMetric, error)
	FindLatestStockMetrics(ctx context.Context, stockID uuid.UUID) ([]store.LatestStockMetric, error)
	CreateAnalysis(ctx context.Context, userID, stockID uuid.UUID, score float64) (*store.Analysis, error)
	CreateRecommendation(
		ctx context.Context,
		analysisID uuid.UUID,
		action store.Action,
		confidenceLevel float64,
		reason string,
	) (*store.Recommendation, error)
}

type Authenticator interface {
	LoginHandler(w http.ResponseWriter, r *http.Request)
	CallbackHandler(w http.ResponseWriter, r *http.Request)
	RequireAuth(handle http.HandlerFunc) http.HandlerFunc
	LogoutHandler(w http.ResponseWriter, r *http.Request)
}

// Server is the API server.
type Server struct {
	h http.Handler

	apiKey        string
	filePath      string
	store         Store
	authenticator Authenticator

	log *logger.Logger
}

// New creates a new API server.
func New(apiKey, filePath string, store Store, obsrv *observe.Observer, auth Authenticator) *Server {
	s := &Server{
		apiKey:        apiKey,
		filePath:      filePath,
		store:         store,
		authenticator: auth,

		log: obsrv.Log.With(lctx.Str("component", "api")),
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

	mux.HandleFunc("GET /auth/login", s.authenticator.LoginHandler)
	mux.HandleFunc("GET /auth/callback", s.authenticator.CallbackHandler)
	mux.HandleFunc("GET /auth/logout", s.authenticator.LogoutHandler)

	mux.HandleFunc("GET /", s.HelloServerHandler)

	mux.HandleFunc("GET /stocks", s.authenticator.RequireAuth(s.GetStockBySymbolHandler))
	mux.HandleFunc("GET /stocks/analysis", s.authenticator.RequireAuth(s.GetStockAnalysisBySymbolHandler))

	mux.HandleFunc("POST /users", s.authenticator.RequireAuth(s.CreateUserHandler))
	mux.HandleFunc("PUT /users/{id}", s.authenticator.RequireAuth(s.UpdateUserHandler))
	mux.HandleFunc("GET /users/{id}", s.authenticator.RequireAuth(s.GetUserHandler))
	mux.HandleFunc("GET /users/me", s.authenticator.RequireAuth(s.GetCurrentUserHandler))

	return corsMiddleware(mux)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
