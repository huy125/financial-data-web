package model

import "github.com/google/uuid"

// User represents user information
type User struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Hash     string    `json:"hash"`
}

type UserUpdate struct {
	Username string    `json:"username"`
}
