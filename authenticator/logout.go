package authenticator

import (
	"fmt"
	"net/http"
	"net/url"

	lctx "github.com/hamba/logger/v2/ctx"
)

// LogoutHandler handles the Auth0 logout process.
// This should be called when the user wants to log out.
func (a *Authenticator) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	token := a.ExtractTokenFromRequest(r)
	if token != "" {
		err := a.revokeToken(r.Context(), token)
		if err != nil && a.Log != nil {
			a.Log.Error("Failed to revoke token", lctx.Err(err))
		}
	}

	u, err := a.getBaseURL()
	if err != nil {
		a.Log.Error("Failed to get domain", lctx.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logoutURL, err := url.Parse(u + "/v2/logout")
	if err != nil {
		a.Log.Error("Failed to parse logout URL", lctx.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get the return URL from query parameter or use default
	returnTo := r.URL.Query().Get("returnTo")
	if returnTo == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		returnTo = fmt.Sprintf("%s://%s/", scheme, r.Host)
	}

	parameters := url.Values{}
	parameters.Add("returnTo", returnTo)
	parameters.Add("client_id", a.Config.ClientID)
	logoutURL.RawQuery = parameters.Encode()

	a.Log.Info("User logout initiated", lctx.Str("return_to", returnTo))

	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
		Secure:   false, // true in production with HTTPS
	})
	// Redirect to Auth0 logout URL
	http.Redirect(w, r, logoutURL.String(), http.StatusTemporaryRedirect)
}
