package store

import (
	"context"

	model "github.com/huy125/financial-data-web/api/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
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