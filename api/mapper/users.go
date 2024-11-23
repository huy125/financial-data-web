package mapper

import (
	"time"

	"github.com/google/uuid"
	"github.com/huy125/financial-data-web/api/dto"
	model "github.com/huy125/financial-data-web/api/store/models"
)

// ToAPIUser converts a store user to an API user
func ToAPIUser(u *model.User) *dto.UserDto {
	return &dto.UserDto{
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
	}
}

// ToStoreUser converts an API user to a store user
func ToStoreUser(u *dto.UserDto) *model.User {
	return &model.User{
		ID:        uuid.New(),
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
