package authenticator

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
)

func (a *Authenticator) LoginHandler(w http.ResponseWriter, r *http.Request) {
	state, err := a.generateState()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate random state: %v", err), http.StatusInternalServerError)
		return
	}

	authCodeURL := a.Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("audience", a.ApiAudience),
		oauth2.SetAuthURLParam("scope", a.Config.Scopes[0]),
	)
	// Redirect to external provider login page
	http.Redirect(w, r, authCodeURL, http.StatusTemporaryRedirect)
}

func (a *Authenticator) generateState() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, a.HmacSecret)
	mac.Write([]byte(state))
	signature := mac.Sum(nil)

	return state + ":" + hex.EncodeToString(signature), nil
}

func (a *Authenticator) verifyState(s string) bool {
	parts := strings.SplitN(s, ":", 2)
	if len(parts) != 2 {
		return false
	}

	state := parts[0]
	sig := parts[1]

	mac := hmac.New(sha256.New, a.HmacSecret)
	mac.Write([]byte(state))
	expectedSig := mac.Sum(nil)

	decodedSig, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}

	return hmac.Equal(decodedSig, expectedSig)
}
