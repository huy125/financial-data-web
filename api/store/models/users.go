package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents user schema in database
type User struct {
	ID        uuid.UUID
	Email     string
	Firstname string
	Lastname  string
	CreatedAt time.Time
	UpdatedAt time.Time
}
