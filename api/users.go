package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/huy125/financial-data-web/authenticator"
	"github.com/huy125/financial-data-web/store"
)

const requestTimeout = 5

// Profile represents the structure of the user data sent to the client.
type Profile struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	ID      string `json:"id"`
}

type userResp struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

func toUserResp(u *store.User) userResp {
	return userResp{
		ID:        u.ID.String(),
		Email:     u.Email,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
	}
}

type userReq struct {
	Email     string `json:"email"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

// CreateUserHandler creates a new user.
func (s *Server) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var userReq userReq
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout*time.Second)
	defer cancel()

	createdUser, err := s.store.CreateUser(
		ctx,
		&store.CreateUser{
			Email:     userReq.Email,
			Lastname:  userReq.Lastname,
			Firstname: userReq.Firstname,
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(toUserResp(createdUser))
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}

// UpdateUserHandler updates the existing user.
func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	var userReq userReq
	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout*time.Second)
	defer cancel()

	userUUID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	updatedUser, err := s.store.UpdateUser(ctx, &store.UpdateUser{
		CreateUser: store.CreateUser{
			Email:     userReq.Email,
			Lastname:  userReq.Lastname,
			Firstname: userReq.Firstname,
		},
		ID: userUUID,
	})
	if err != nil {
		s.handleStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(toUserResp(updatedUser))
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}

// GetUserHandler gets an existing user.
func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := s.store.FindUser(ctx, uuid)
	if err != nil {
		s.handleStoreError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(toUserResp(user))
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) GetCurrentUserHandler(w http.ResponseWriter, r *http.Request) {
	userCtx := r.Context().Value(authenticator.UserContextKey)
	claims, ok := userCtx.(authenticator.Claims)
	if !ok {
		http.Error(w, "Failed to get user claims", http.StatusInternalServerError)
		return
	}

	// Create response object
	user := Profile{
		Email:   claims.Email,
		Name:    claims.Name,
		Picture: claims.Picture,
		ID:      claims.Subject,
	}

	// Set content type header
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
		return
	default:
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
