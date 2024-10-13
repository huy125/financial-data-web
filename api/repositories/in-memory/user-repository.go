package repository

import (
	model "github.com/huy125/financial-data-web/api/models"
)

type UserRepository interface {
	Create(user model.User)
	List() map[int]model.User
}

type InMemoryUserRepository struct {
	users map[int]model.User
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[int]model.User),
	}
}

func (repo *InMemoryUserRepository) Create(user model.User) {
	repo.users[user.ID] = user
}

func (repo *InMemoryUserRepository) List() map[int]model.User {
	return repo.users
}
