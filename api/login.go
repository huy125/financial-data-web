package api

import (
	"encoding/json"
	"net/http"
	"net/url"

	lctx "github.com/hamba/logger/v2/ctx"
)

// LoginHandler handles the authenticator login process.
func (s *Server) LoginHandler(w http.ResponseWriter, _ *http.Request) {
	state, err := s.authenticator.GenerateState()
	if err != nil {
		s.log.Error("Failed to generate random state", lctx.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", s.authenticator.GetClientOrigin())

	err = json.NewEncoder(w).Encode(map[string]string{"state": state})
	if err != nil {
		s.log.Error("Failed to encode the response", lctx.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// CallbackHandler handles the authenticator provider login callback.
func (s *Server) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state, err := url.QueryUnescape(r.URL.Query().Get("state"))
	if err != nil {
		http.Error(w, "Invalid state format", http.StatusBadRequest)
		return
	}

	if state == "" {
		http.Error(w, "Missing state", http.StatusBadRequest)
		return
	}

	if err := s.authenticator.VerifyState(state); err != nil {
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
		s.log.Error("Failed to exchange token", lctx.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify the ID token
	_, err = s.authenticator.VerifyToken(r.Context(), token)
	if err != nil {
		s.log.Error("Failed to verify token", lctx.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     s.cookieCfg.Name,
		Path:     s.cookieCfg.Path,
		HttpOnly: s.cookieCfg.HttpOnly,
		Secure:   s.cookieCfg.Secure,
		Value:    token.AccessToken,
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, s.authenticator.GetClientOrigin(), http.StatusFound)
}
