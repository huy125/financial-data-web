package mapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/huy125/financial-data-web/api/dto"
	"github.com/huy125/financial-data-web/store"
)

// ToAPIUser converts a store user to an API user.
func ToAPIUser(u *store.User) *dto.UserDto {
	return &dto.UserDto{
		ID:        u.ID.String(),
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
	}
}

// ToStoreUser converts an API user to a store user.
func ToStoreUser(u *dto.UserDto) (*store.User, error) {
	now := time.Now()
	user := &store.User{
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
	}

	if u.ID == "" {
		user.ID = uuid.New()
		user.CreatedAt = now
		user.UpdatedAt = now
	} else {
		id, err := uuid.Parse(u.ID)
		if err != nil {
			return nil, err
		}
		user.ID = id
		user.UpdatedAt = now
	}

	return user, nil
}
