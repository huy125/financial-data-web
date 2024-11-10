package store

import (
	"context"

	model "github.com/huy125/financial-data-web/api/models"
)

type InMemory struct {
	users []model.User
}

func NewInMemory()(*InMemory, error) {
	return &InMemory{
		users: []model.User{},
	}, nil
}

func (s *InMemory) Create(ctx context.Context, user model.User) error {
	user.ID = len(s.users) + 1
	s.users = append(s.users, user)

	return nil
}

func (s *InMemory) List(ctx context.Context, limit, offset int) ([]model.User, error) {
	return s.users, nil
}
