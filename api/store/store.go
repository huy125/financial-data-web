package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/huy125/financial-data-web/api"
	model "github.com/huy125/financial-data-web/api/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Option func(*options)

type options struct {
	dsn string
}

func WithDSN(dsn string) Option {
	return func(opts *options) {
		opts.dsn = dsn
	}
}

func New(opts ...Option) (api.UserStore, error) {
	c := &options{}
	for _, opt := range opts {
		opt(c)
	}

	switch {
	case c.dsn == "":
		return &InMemory{
			users: []model.User{},
		}, nil
	case strings.HasPrefix("postgres", c.dsn):
		config, err := pgxpool.ParseConfig(c.dsn)
		if err != nil {
			return nil, fmt.Errorf("parsing dsn: %w", err)
		}

		config.MaxConns = 25
		config.MinConns = 5
		config.MaxConnLifetime = time.Minute * 5
		config.MaxConnIdleTime = time.Minute * 1

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
	default:
		return nil, errors.New("unsupported store")
	}
}
