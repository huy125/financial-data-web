package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB holds the overall store configuration.
type DB struct {
	pool *pgxpool.Pool

	dsn             string
	maxConns        int32
	minConns        int32
	maxConnIdleTime time.Duration
	maxConnLifetime time.Duration
}

type Option func(*DB)

// WithDSN sets the Data Source Name.
func WithDSN(dsn string) Option {
	return func(p *DB) {
		p.dsn = dsn
	}
}

// WithMaxConns sets the maximum pool connections.
func WithMaxConns(n int32) Option {
	return func(p *DB) {
		p.maxConns = n
	}
}

// WithMinConns sets the minimum pool connections.
func WithMinConns(n int32) Option {
	return func(p *DB) {
		p.minConns = n
	}
}

// WithMaxConnIdleTime sets the maximum connection idle duration.
func WithMaxConnIdleTime(d time.Duration) Option {
	return func(p *DB) {
		p.maxConnIdleTime = d
	}
}

// WithMaxConnLifetime sets the maximum connection lifetime duration.
func WithMaxConnLifetime(d time.Duration) Option {
	return func(p *DB) {
		p.maxConnLifetime = d
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

	config.MaxConns = p.maxConns
	config.MinConns = p.minConns
	config.MaxConnLifetime = p.maxConnLifetime
	config.MaxConnIdleTime = p.maxConnIdleTime

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
