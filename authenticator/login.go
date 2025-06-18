package authenticator

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	StateGenerationByteSize = 32
	StatePartCount          = 2
)

// LoginHandler handles the authenticator login process.
func (a *Authenticator) LoginHandler(w http.ResponseWriter, _ *http.Request) {
	state, err := a.generateState()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate random state: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", a.ClientOrigin)

	err = json.NewEncoder(w).Encode(map[string]string{"state": state})
	if err != nil {
		http.Error(w, "Failed to encode the response", http.StatusInternalServerError)
		return
	}
}

func (a *Authenticator) generateState() (string, error) {
	b := make([]byte, StateGenerationByteSize)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, a.HMACSecret)
	mac.Write([]byte(state))
	signature := mac.Sum(nil)

	return state + ":" + hex.EncodeToString(signature), nil
}

func (a *Authenticator) verifyState(s string) bool {
	parts := strings.SplitN(s, ":", StatePartCount)
	if len(parts) != StatePartCount {
		return false
	}

	state := parts[0]
	sig := parts[1]

	mac := hmac.New(sha256.New, a.HMACSecret)
	mac.Write([]byte(state))
	expectedSig := mac.Sum(nil)

	decodedSig, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}

	return hmac.Equal(decodedSig, expectedSig)
}

// CallbackHandler handles the authenticator provider login callback.
func (a *Authenticator) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state, err := url.QueryUnescape(r.URL.Query().Get("state"))
	if err != nil {
		http.Error(w, ErrInvalidState.Error(), http.StatusBadRequest)
		return
	}

	if state == "" || !a.verifyState(state) {
		http.Error(w, ErrInvalidState.Error(), http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, ErrMissingAuthCode.Error(), http.StatusBadRequest)
		return
	}

	token, err := a.Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", ErrTokenExchangeFail.Error(), err), http.StatusInternalServerError)
		return
	}

	// Verify the ID token
	_, err = a.verifyToken(r.Context(), token)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: %v", ErrTokenVerifyFail.Error(), err), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token.AccessToken,
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // true in production with HTTPS
	})

	http.Redirect(w, r, a.ClientOrigin, http.StatusFound)
}
