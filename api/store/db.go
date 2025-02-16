package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConnLifeTime int8 = 5 // 5 minutes
	maxConnIdleTime int8 = 1 // 1 minute
)

type DB struct {
	pool *pgxpool.Pool
	dsn  string
}

type Option func(*DB)

func WithDSN(dsn string) Option {
	return func(p *DB) {
		p.dsn = dsn
	}
}

func NewDB(opts ...Option) (*DB, error) {
	p := DB{}
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

	return &DB{
		pool: pool,
	}, nil
}
