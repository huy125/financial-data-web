package store

import (
	"context"
	"errors"
	"time"

	model "github.com/huy125/financial-data-web/api/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config func(*configs)

type configs struct {
	dsn string
}

type Store interface {
	Create(ctx context.Context, user model.User) error
	List(ctx context.Context, limit, offset int) ([]model.User, error)
}

func WithDSN(dsn string) Config {
	return func(cfg *configs) {
		cfg.dsn = dsn
	}
}

func New(store Store, cfgs ...Config) (Store, error) {
	c := &configs{}
	for _, cfg := range cfgs {
		cfg(c)
	}

	switch store.(type) {
	case *InMemory:
		return &InMemory{
			users: []model.User{},
		}, nil
	case *Postgres:
		config, err := pgxpool.ParseConfig(c.dsn)
		if err != nil {
			return nil, err
		}

		config.MaxConns = 25
		config.MinConns = 5
		config.MaxConnLifetime = time.Minute * 5
		config.MaxConnIdleTime = time.Minute * 1

		pool, err := pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			return nil, err
		}

		err = pool.Ping(context.Background())
		if err != nil {
			return nil, err
		}

		return &Postgres{
			pool: pool,
		}, nil
	default:
		return nil, errors.New("unsupported store")
	}
}
