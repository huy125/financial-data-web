package api

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/hamba/cmd/v2/observe"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/huy125/financial-data-web/api/middleware"
	"github.com/huy125/financial-data-web/store"
	"golang.org/x/oauth2"
)

// Store defines the interface for interacting with the application's persistent storage.
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

// Authenticator defines the interface for handling authentication flows within the application.
type Authenticator interface {
	middleware.Authenticator

	GenerateState() (string, error)
	GetBaseURL() (string, error)
	VerifyState(s string) bool
	VerifyToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error)
	RevokeToken(ctx context.Context, token string) error
	Exchange(ctx context.Context, code string) (*oauth2.Token, error)
	GetClientID() string
	GetClientOrigin() string
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
func New(apiKey, filePath string, store Store, auth Authenticator, obsrv *observe.Observer) *Server {
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

	mux.HandleFunc("GET /auth/login", s.LoginHandler)
	mux.HandleFunc("GET /auth/callback", s.CallbackHandler)
	mux.HandleFunc("GET /auth/logout", s.LogoutHandler)

	mux.HandleFunc("GET /", s.HelloServerHandler)

	mux.HandleFunc("GET /stocks", middleware.RequireAuth(s.GetStockBySymbolHandler, s.authenticator))
	mux.HandleFunc("GET /stocks/analysis", middleware.RequireAuth(s.GetStockAnalysisBySymbolHandler, s.authenticator))

	mux.HandleFunc("POST /users", middleware.RequireAuth(s.CreateUserHandler, s.authenticator))
	mux.HandleFunc("PUT /users/{id}", middleware.RequireAuth(s.UpdateUserHandler, s.authenticator))
	mux.HandleFunc("GET /users/{id}", middleware.RequireAuth(s.GetUserHandler, s.authenticator))
	mux.HandleFunc("GET /users/me", middleware.RequireAuth(s.GetCurrentUserHandler, s.authenticator))

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
