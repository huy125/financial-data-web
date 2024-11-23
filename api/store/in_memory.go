package store

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	model "github.com/huy125/financial-data-web/api/store/models"
)

type InMemory struct {
	mu    sync.Mutex
	users []model.User
}

func NewInMemory() (*InMemory, error) {
	return &InMemory{}, nil
}

func (s *InMemory) Create(ctx context.Context, user model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user.ID = uuid.New()
	s.users = append(s.users, user)

	return nil
}

func (s *InMemory) List(ctx context.Context, limit, offset int) ([]model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if offset > len(s.users) {
		return nil, errors.New("offset is out of range")
	}

	start := offset
	end := start + limit
	if end > len(s.users) {
		end = len(s.users)
	}

	users := make([]model.User, end-start)
	copy(users, s.users[start:end])

	return users, nil
}

func (s *InMemory) Find(ctx context.Context, id uuid.UUID) (*model.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, user := range s.users {
		if user.ID == id {
			return &user, nil
		}
	}

	return nil, ErrNotFound
}

func (s *InMemory) Update(ctx context.Context, user model.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, u := range s.users {
		if u.ID == user.ID {
			s.users[i].Email = user.Email
			s.users[i].Firstname = user.Firstname
			s.users[i].Lastname = user.Lastname

			return nil
		}
	}

	return ErrNotFound
}
