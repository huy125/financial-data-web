package authenticator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"net/http"
)

type tokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

// CallbackHandler handles the authenticator provider login callback
func (a *Authenticator) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" || !a.verifyState(state) {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, fmt.Sprintf("Missing authorization code"), http.StatusBadRequest)
		return
	}

	token, err := a.Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange token: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify the ID token
	_, err = a.verifyToken(r.Context(), token)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify token: %v", err), http.StatusInternalServerError)
		return
	}

	tokenResp := tokenResp{
		AccessToken: token.AccessToken,
		ExpiresIn:   token.ExpiresIn,
		TokenType:   "Bearer",
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tokenResp)
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}

// verifyToken verifies the ID token and returns the parsed token
func (a *Authenticator) verifyToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	// Configure the ID token verifier
	oidcConfig := &oidc.Config{
		ClientID: a.Config.ClientID,
	}
	verifier := a.Provider.Verifier(oidcConfig)

	// Verify the ID token
	return verifier.Verify(ctx, rawIDToken)
}
