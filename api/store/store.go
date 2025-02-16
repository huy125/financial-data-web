package store

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	db *DB

	users	*userService
	stocks	*stockService
	metrics	*metricService
}

func New(db *DB) *Store {
	store := &Store{
		db: db,
	}

	store.users = &userService{db: db}
	store.stocks = &stockService{db: db}
	store.metrics = &metricService{db: db}

	return store
}

func (s *Store) Create(ctx context.Context, user *User) (*User, error) {
	return s.users.Create(ctx, user)
}

func (s *Store) List(ctx context.Context, limit, offset int) ([]User, error) {
	return s.users.List(ctx, limit, offset)
}

func (s *Store) Find(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.users.Find(ctx, id)
}

func (s *Store) Update(ctx context.Context, user *User) (*User, error) {
	return s.users.Update(ctx, user)
}

func (s *Store) FindStockBySymbol(ctx context.Context, symbol string) (*Stock, error) {
	return s.stocks.FindStockBySymbol(ctx, symbol)
}

func (s *Store) CreateStockMetric(ctx context.Context, stockID, metricID uuid.UUID, value float64) (*StockMetric, error) {
	now := time.Now()
	stockMetric := &StockMetric{
		StockID:    stockID,
		MetricID:   metricID,
		Value:      value,
		RecordedAt: &now,
	}
	return s.stocks.CreateStockMetric(ctx, *stockMetric)
}

func (s *Store) ListMetrics(ctx context.Context, limit, offset int) ([]Metric, error) {
	return s.metrics.ListMetrics(ctx, limit, offset)
}
