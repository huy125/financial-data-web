package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type User struct {
	Model

	Email     string
	Firstname string
	Lastname  string
}

type userService struct {
	db *DB
}

func (s *userService) Create(ctx context.Context, user *User) (*User, error) {
	sql := `
		INSERT INTO users (email, firstname, lastname)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := s.db.pool.QueryRow(ctx, sql,
		user.Email,
		user.Firstname,
		user.Lastname,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) List(ctx context.Context, limit, offset int) ([]User, error) {
	sql := "SELECT id, email, firstname, lastname FROM users LIMIT $1 OFFSET $2"
	rows, err := s.db.pool.Query(ctx, sql, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
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

func (s *userService) Find(ctx context.Context, id uuid.UUID) (*User, error) {
	sql := "SELECT id, email, firstname, lastname, created_at, updated_at FROM users WHERE id = $1"
	var user User
	err := s.db.pool.QueryRow(ctx, sql, id).Scan(
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

func (s *userService) Update(ctx context.Context, user *User) (*User, error) {
	sql := `
		UPDATE users
		SET email = $1,
			firstname = $2,
			lastname = $3,
			updated_at = CURRENT_TIMESTAMP
			WHERE id = $4
	`

	res, err := s.db.pool.Exec(ctx, sql, user.Email, user.Firstname, user.Lastname, user.ID)
	if err != nil {
		return nil, err
	}

	if res.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	return user, nil
}
