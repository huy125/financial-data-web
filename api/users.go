package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

	user, err := mapper.ToStoreUser(&userDto)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to map dto to model: %v", err), http.StatusInternalServerError)
		return
	}

	createdUser, err := h.store.Create(context.Background(), user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapper.ToAPIUser(createdUser))
}

// UpdateUserHandler updates the existing user
func (h *UserHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Id is required", http.StatusBadRequest)
		return
	}

	var userDto dto.UserDto
	if err := json.NewDecoder(r.Body).Decode(&userDto); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if id != userDto.Id {
		http.Error(w, "Mismatch Id between path and body", http.StatusBadRequest)
		return
	}

	if err := userDto.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	user, err := mapper.ToStoreUser(&userDto)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to map dto to model: %v", err), http.StatusInternalServerError)
		return
	}

	updatedUser, err := h.store.Update(context.Background(), user)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapper.ToAPIUser(updatedUser))
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
