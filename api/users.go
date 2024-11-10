package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/mail"

	model "github.com/huy125/financial-data-web/api/models"
)

// CreateUserValidator represents the required payload for user creation
type CreateUserValidator struct {
	Email string `json:"email"`
	Hash  string `json:"hash"`
}

type UserHandler struct {
	Store UserStore
}

// CreateUserHandler creates a new user with hashed password
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		http.Error(w, "Database connection is uninitialized", http.StatusInternalServerError)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
	}

	var validator CreateUserValidator
	err = json.Unmarshal(body, &validator)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := validator.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := model.User{
		Username: validator.Email,
		Hash:     validator.Hash,
	}

	err = h.Store.Create(context.Background(), user)

	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal("User created successfully")
	if err != nil {
		http.Error(w, "Failed to marshal to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

// Validate checks if the CreateUserValidator fields are valid
func (v *CreateUserValidator) Validate() error {
	if v.Email == "" {
		return errors.New("email is required")
	}

	if !isValidEmail(v.Email) {
		return errors.New("email is invalid")
	}

	if v.Hash == "" {
		return errors.New("password hash is required")
	}

	return nil
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)

	return err == nil
}
