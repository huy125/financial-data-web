package model

import (
	"github.com/google/uuid"
)

// Stock represents stock schema in database.
type Stock struct {
	ID      uuid.UUID
	Symbol  string
	Company string
}
