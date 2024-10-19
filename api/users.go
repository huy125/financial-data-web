package api

import (
	"encoding/json"
	"io"
	"net/http"

	model "github.com/huy125/financial-data-web/api/models"
)

// CreateUserValidator represents the required payload for user creation
type CreateUserValidator struct {
	Email	string `json:"email"`
	Hash 	string `json:"hash"`
}

type UserHandler struct {
	Store InMemoryStore
}

// CreateUserHandler creates a new user with hashed password
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
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

	user := model.User{
		Username: validator.Email,
		Hash: validator.Hash,
	}

	user = h.Store.Create(user)

	response, err := json.Marshal(user)

	if err != nil {
		http.Error(w, "Failed to marshal to JSON", http.StatusInternalServerError)
	}

    w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
