package store

import (
	"errors"
	"strings"
)

const errorMsgSize = 5

// ErrNotFound represents a not found error in the store.
var ErrNotFound = errors.New("not found")

// ValidationError represents a validation error.
type ValidationError struct {
	Error string `json:"error"`
}

// ValidationErrors represents a list of validation errors.
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	errMsgs := make([]string, 0, errorMsgSize)

	for _, err := range ve {
		errMsgs = append(errMsgs, err.Error)
	}

	return strings.Join(errMsgs, "; ")
}
