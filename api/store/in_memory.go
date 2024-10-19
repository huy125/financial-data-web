package store

import (
	model "github.com/huy125/financial-data-web/api/models"
)

type InMemory struct {
	users map[int]model.User
}

func NewInMemory() *InMemory {
	return &InMemory{
		users: make(map[int]model.User),
	}
}

func (s *InMemory) Create(user model.User) {
	s.users[user.ID] = user
}

func (s *InMemory) List() map[int]model.User {
	return s.users
}
