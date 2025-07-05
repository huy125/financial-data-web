package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// LoginHandler handles the authenticator login process.
func (s *Server) LoginHandler(w http.ResponseWriter, _ *http.Request) {
	state, err := s.authenticator.GenerateState()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate random state: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", s.authenticator.GetClientOrigin())

	err = json.NewEncoder(w).Encode(map[string]string{"state": state})
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}

// CallbackHandler handles the authenticator provider login callback.
func (s *Server) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state, err := url.QueryUnescape(r.URL.Query().Get("state"))
	if err != nil {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	if state == "" || !s.authenticator.VerifyState(state) {
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	token, err := s.authenticator.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", "Failed to exchange token", err), http.StatusInternalServerError)
		return
	}

	// Verify the ID token
	_, err = s.authenticator.VerifyToken(r.Context(), token)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", "Failed to verify token", err), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token.AccessToken,
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // true in production with HTTPS
	})

	http.Redirect(w, r, s.authenticator.GetClientOrigin(), http.StatusFound)
}
