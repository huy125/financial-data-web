package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	model "github.com/huy125/financial-data-web/api/models"
	"github.com/jackc/pgx/v5/pgxpool"
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
}

func (p *Postgres) Create(ctx context.Context, user model.User) error {
	sql := "INSERT INTO users (username, hash) VALUES ($1, $2)"
	_, err := p.pool.Exec(ctx, sql, user.Username, user.Hash)

	return err
}

func (p *Postgres) List(ctx context.Context, limit, offset int) ([]model.User, error) {
	sql := "SELECT id, username, hash FROM users LIMIT $1 OFFSET $2"
	rows, err := p.pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Hash); err != nil {
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
	sql := "SELECT id, username, hash FROM users WHERE id = $1"
	var user model.User
	err := p.pool.QueryRow(ctx, sql, id).Scan(&user.ID, &user.Username, &user.Hash)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *Postgres) Update(ctx context.Context, id uuid.UUID, userUpdate model.UserUpdate) error {
	sql := "UPDATE users SET username = $1 WHERE id = $2"
	_, err := p.pool.Exec(ctx, sql, userUpdate.Username, id)

	return err
}
