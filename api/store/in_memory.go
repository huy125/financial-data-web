package store

import (
	model "github.com/huy125/financial-data-web/api/models"
)

type InMemory struct {
	users []model.User
}

func NewInMemory() *InMemory {
	return &InMemory{
		users: []model.User{},
	}
}

func (s *InMemory) Create(user model.User) model.User {
	user.ID = len(s.users) + 1
	s.users = append(s.users, user)

	return user
}

func (s *InMemory) List() []model.User {
	return s.users
}
