package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/huy125/financial-data-web/api/dto"
	"github.com/huy125/financial-data-web/api/mapper"
	"github.com/huy125/financial-data-web/api/store"
)

type UserHandler struct {
	store UserStore
}

// CreateUserHandler creates a new user with hashed password
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var userDto dto.UserDto
	if err := json.NewDecoder(r.Body).Decode(&userDto); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := userDto.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	user := mapper.ToStoreUser(&userDto)

	err := h.store.Create(context.Background(), user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapper.ToAPIUser(user))
}

// UpdateUserHandler updates the existing user
func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	userId, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	var userUpdate model.User
	err = json.Unmarshal(body, &userUpdate)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	existUser, err := h.store.Find(context.Background(), userId)
	if existUser == nil {
		h.handleStoreError(w, err)
		return
	}

	err = h.store.Update(context.Background(), userId, userUpdate)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}

	response, err := json.Marshal("User updated successfully")
	if err != nil {
		http.Error(w, "Failed to marshal to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func (h *UserHandler) handleStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
