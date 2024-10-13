package api

import (
	"encoding/json"
	"io"
	"net/http"

	model "github.com/huy125/financial-data-web/api/models"
	repository "github.com/huy125/financial-data-web/api/repositories/in-memory"
	"golang.org/x/crypto/bcrypt"
)

// CreateUserValidator represents the required payload for user creation
type CreateUserValidator struct {
	Email		string `json:"email"`
	Password 	string `json:"password"`
}

type UserHandler struct {
	Repo repository.UserRepository
}

var nextID = 1

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

	hashedPassword, err := hashPassword(validator.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	newUser := model.User{
		Username: validator.Email,
		Hash: hashedPassword,
	}

	newUser.ID = nextID
	h.Repo.Create(newUser)
	nextID++

	response, _ := json.Marshal(map[string]interface{}{
		"id": newUser.ID,
	})

    w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}
