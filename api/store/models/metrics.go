package model

import (
	"github.com/google/uuid"
)

// Stock represents stock schema in database.
type Metric struct {
	ID          uuid.UUID
	Name        string
	Description string
}
