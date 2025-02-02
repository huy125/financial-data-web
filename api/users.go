package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/huy125/financial-data-web/api/dto"
	"github.com/huy125/financial-data-web/api/mapper"
	"github.com/huy125/financial-data-web/api/store"
)

const requestTimeout = 5

type UserHandler struct {
	store Store
}

// CreateUserHandler creates a new user.
func (h *UserHandler) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var userDto dto.UserDto
	if err := json.NewDecoder(r.Body).Decode(&userDto); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := userDto.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(err); encodeErr != nil {
			http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		}
		return
	}

	user, err := mapper.ToStoreUser(&userDto)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to map dto to model: %v", err), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout*time.Second)
	defer cancel()

	createdUser, err := h.store.Create(ctx, user)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(mapper.ToAPIUser(createdUser))
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}

// UpdateUserHandler updates the existing user.
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

	if id != userDto.ID {
		http.Error(w, "Mismatch Id between path and body", http.StatusBadRequest)
		return
	}

	if err := userDto.Validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		if encodeErr := json.NewEncoder(w).Encode(err); encodeErr != nil {
			http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		}
		return
	}

	user, err := mapper.ToStoreUser(&userDto)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to map dto to model: %v", err), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout*time.Second)
	defer cancel()

	updatedUser, err := h.store.Update(ctx, user)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(mapper.ToAPIUser(updatedUser))
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}

// GetUserHandler gets an existing user.
func (h *UserHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Id is required", http.StatusBadRequest)
		return
	}

	uuid, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Failed to parse Id", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout*time.Second)
	defer cancel()

	user, err := h.store.Find(ctx, uuid)
	if err != nil {
		h.handleStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(mapper.ToAPIUser(user))
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
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
