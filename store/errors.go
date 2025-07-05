package store

import (
	"errors"
)

const errorMsgSize = 5

// ErrNotFound represents a not found error in the store.
var ErrNotFound = errors.New("not found")

// ValidationError represents a validation error.
type ValidationError struct {
	Err string `json:"err"`
}

// Error implements the error interface for ValidationError
func (v ValidationError) Error() string {
	return v.Err
}
