package dto

import (
	"net/mail"
	"strings"
)

// User represents user information
type UserDto struct {
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}


// ValidationError represents a validation error
type ValidationError struct {
	Error string `json:"error"`
}

// ValidationErrors represents a list of validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var errMsgs []string

	for _, err := range ve {
		errMsgs = append(errMsgs, err.Error)
	}

	return strings.Join(errMsgs, "; ")
}

func (u *UserDto) Validate() error {
	var errors ValidationErrors

	if u.Email == "" {
		errors = append(errors, ValidationError{Error: "email is required"})
	} else if !isValidEmail(u.Email) {
		errors = append(errors, ValidationError{Error: "email is invalid"})
	}

	if u.Firstname == "" {
		errors = append(errors, ValidationError{Error: "firstname is required"})
	}

	if u.Lastname == "" {
		errors = append(errors, ValidationError{Error: "lastname is required"})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil
}
