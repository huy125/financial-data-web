package authenticator

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type tokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
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

	tokenResp := tokenResp{
		AccessToken: token.AccessToken,
		ExpiresIn:   token.ExpiresIn,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(tokenResp)
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}
