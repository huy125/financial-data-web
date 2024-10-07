package api

import (
	"errors"
)

// ErrNotFound represents a not found error in the API.
var ErrNotFound = errors.New("not found")
