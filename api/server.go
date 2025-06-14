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

// Server is the API server.
type Server struct {
	h http.Handler

	apiKey   string
	filePath string
	store    Store

	log *logger.Logger
}

// New creates a new API server.
func New(apiKey, filePath string, store Store, obsrv *observe.Observer) *Server {
	s := &Server{
		apiKey:   apiKey,
		filePath: filePath,
		store:    store,
		log:      obsrv.Log.With(lctx.Str("component", "api")),
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

	mux.HandleFunc("GET /", s.HelloServerHandler)
	mux.HandleFunc("GET /stocks", s.GetStockBySymbolHandler)
	mux.HandleFunc("GET /stocks/analysis", s.GetStockAnalysisBySymbolHandler)

	mux.HandleFunc("POST /users", s.CreateUserHandler)
	mux.HandleFunc("PUT /users/{id}", s.UpdateUserHandler)
	mux.HandleFunc("GET /users/{id}", s.GetUserHandler)

	return mux
}
