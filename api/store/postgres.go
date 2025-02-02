package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	model "github.com/huy125/financial-data-web/api/store/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConnLifeTime int8 = 5 // 5 minutes
	maxConnIdleTime int8 = 1 // 1 minute
)

type Postgres struct {
	pool *pgxpool.Pool
	dsn  string
}

type Option func(*Postgres)

func WithDSN(dsn string) Option {
	return func(p *Postgres) {
		p.dsn = dsn
	}
}

func NewPostgres(opts ...Option) (*Postgres, error) {
	p := Postgres{}
	for _, opt := range opts {
		opt(&p)
	}

	config, err := pgxpool.ParseConfig(p.dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing dsn: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Minute * time.Duration(maxConnLifeTime)
	config.MaxConnIdleTime = time.Minute * time.Duration(maxConnIdleTime)

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("configuring pool connection: %w", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("ping connection: %w", err)
	}

	return &Postgres{
		pool: pool,
	}, nil
}

func (p *Postgres) Create(ctx context.Context, user *model.User) (*model.User, error) {
	sql := `
		INSERT INTO users (email, firstname, lastname)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := p.pool.QueryRow(ctx, sql,
		user.Email,
		user.Firstname,
		user.Lastname,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (p *Postgres) List(ctx context.Context, limit, offset int) ([]model.User, error) {
	sql := "SELECT id, email, firstname, lastname FROM users LIMIT $1 OFFSET $2"
	rows, err := p.pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Firstname, &user.Lastname); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return users, nil
}

func (p *Postgres) Find(ctx context.Context, id uuid.UUID) (*model.User, error) {
	sql := "SELECT id, email, firstname, lastname, created_at, updated_at FROM users WHERE id = $1"
	var user model.User
	err := p.pool.QueryRow(ctx, sql, id).Scan(
		&user.ID,
		&user.Email,
		&user.Firstname,
		&user.Lastname,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return &user, nil
}

func (p *Postgres) Update(ctx context.Context, user *model.User) (*model.User, error) {
	sql := `
		UPDATE users
		SET email = $1,
			firstname = $2,
			lastname = $3,
			updated_at = CURRENT_TIMESTAMP
			WHERE id = $4
	`

	res, err := p.pool.Exec(ctx, sql, user.Email, user.Firstname, user.Lastname, user.ID)
	if err != nil {
		return nil, err
	}

	if res.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return user, nil
}

func (p *Postgres) ListMetrics(ctx context.Context, limit, offset int) ([]model.Metric, error) {
	sql := "SELECT id, name, description FROM metric LIMIT $1 OFFSET $2"
	rows, err := p.pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []model.Metric
	for rows.Next() {
		var metric model.Metric
		if err := rows.Scan(&metric.ID, &metric.Name, &metric.Description); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return metrics, nil
}

func (p * Postgres) CreateStockMetric(ctx context.Context, stockMetric *model.StockMetric) (*model.StockMetric, error) {
	sql := `
		INSERT INTO stock_metric (stock_id, metric_id, value)
		VALUES ($1, $2, $3)
		RETURNING id, recorded_at
	`

	err := p.pool.QueryRow(ctx, sql,
		stockMetric.StockID,
		stockMetric.MetricID,
		stockMetric.Value,
	).Scan(&stockMetric.ID, &stockMetric.RecordedAt)

	if err != nil {
		return nil, err
	}

	return stockMetric, nil
}
