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
		Id:        u.Id.String(),
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
	}
}

// ToStoreUser converts an API user to a store user
func ToStoreUser(u *dto.UserDto) (*model.User, error) {
	now := time.Now()
	user := &model.User{
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
	}

	if u.Id == "" {
		user.Id = uuid.New()
		user.CreatedAt = now
		user.UpdatedAt = now
	} else {
		id, err := uuid.Parse(u.Id)
		if err != nil {
			return nil, err
		}
		user.Id = id
		user.UpdatedAt = now
	}

	return user, nil
}
