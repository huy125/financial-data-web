package store

import (
	"context"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// Store manages the database layer of applications.
type Store struct {
	db *DB

	users           *userService
	stocks          *stockService
	metrics         *metricService
	analyses        *analysisService
	recommendations *recommendationService
}

func New(db *DB) *Store {
	store := &Store{
		db: db,
	}

	store.users = &userService{db: db}
	store.stocks = &stockService{db: db}
	store.metrics = &metricService{db: db}
	store.analyses = &analysisService{db: db}
	store.recommendations = &recommendationService{db: db}

	return store
}

// CreateUser contains user creation information.
type CreateUser struct {
	Email     string
	Firstname string
	Lastname  string
}

// UpdateUser contains user updating information.
type UpdateUser struct {
	CreateUser

	ID	uuid.UUID
}

// Validate validates an CreateUser configuration.
func (c *CreateUser) Validate() error {
	var errors ValidationErrors

	if c.Email == "" {
		errors = append(errors, ValidationError{Error: "email is required"})
	} else if !isValidEmail(c.Email) {
		errors = append(errors, ValidationError{Error: "email is invalid"})
	}

	if c.Firstname == "" {
		errors = append(errors, ValidationError{Error: "firstname is required"})
	}

	if c.Lastname == "" {
		errors = append(errors, ValidationError{Error: "lastname is required"})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// Validate validates an UpdateUser configuration.
func (u *UpdateUser) Validate() error {
	if u.ID == uuid.Nil {
		return ValidationErrors{ValidationError{Error: "id is required"}}
	}

	return u.CreateUser.Validate()
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil
}

func (s *Store) CreateUser(ctx context.Context, u *CreateUser) (*User, error) {
	if errs := u.Validate(); errs != nil {
		return nil, errs
	}

	user := &User{
		Model: Model{
			ID: uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Email: u.Email,
		Firstname: u.Firstname,
		Lastname: u.Lastname,
	}

	return s.users.Create(ctx, user)
}

func (s *Store) ListUsers(ctx context.Context, limit, offset int) ([]User, error) {
	return s.users.List(ctx, limit, offset)
}

func (s *Store) FindUser(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.users.Find(ctx, id)
}

func (s *Store) UpdateUser(ctx context.Context, u *UpdateUser) (*User, error) {
	if errs := u.Validate(); errs != nil {
		return nil, errs
	}

	user := &User{
		Model: Model{
			ID: u.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Email: u.Email,
		Firstname: u.Firstname,
		Lastname: u.Lastname,
	}

	return s.users.Update(ctx, user)
}

func (s *Store) FindStockBySymbol(ctx context.Context, symbol string) (*Stock, error) {
	return s.stocks.Find(ctx, symbol)
}

func (s *Store) CreateStockMetric(
	ctx context.Context,
	stockID, metricID uuid.UUID,
	value float64,
) (*StockMetric, error) {
	stockMetric := &StockMetric{
		StockID:    stockID,
		MetricID:   metricID,
		Value:      value,
		RecordedAt: time.Now(),
	}
	return s.stocks.CreateStockMetric(ctx, *stockMetric)
}

func (s *Store) ListMetrics(ctx context.Context, limit, offset int) ([]Metric, error) {
	return s.metrics.ListMetrics(ctx, limit, offset)
}

func (s *Store) FindLastestStockMetrics(ctx context.Context, stockID uuid.UUID) ([]LatestStockMetric, error) {
	return s.stocks.FindLastestStockMetrics(ctx, stockID)
}

func (s *Store) CreateAnalysis(ctx context.Context, userID, stockID uuid.UUID, score float64) (*Analysis, error) {
	analysis := &Analysis{
		ID:        uuid.New(),
		UserID:    userID,
		StockID:   stockID,
		Score:     score,
		CreatedAt: time.Now(),
	}
	return s.analyses.CreateAnalysis(ctx, analysis)
}

func (s *Store) CreateRecommendation(
	ctx context.Context,
	analysisID uuid.UUID,
	action Action,
	confidenceLevel float64,
	reason string,
) (*Recommendation, error) {
	recommendation := &Recommendation{
		ID:              uuid.New(),
		AnalysisID:      analysisID,
		Action:          action,
		ConfidenceLevel: confidenceLevel,
		Reason:          reason,
		CreatedAt:       time.Now(),
	}

	return s.recommendations.Create(ctx, recommendation)
}
