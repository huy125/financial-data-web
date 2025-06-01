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
	token := extractTokenFromRequest(r)

	if token != "" {
		err := a.revokeToken(r.Context(), token)
		if err != nil && a.log != nil {
			a.log.Error("Failed to revoke token", lctx.Err(err))
		}
	}

	domain := a.getDomain()
	logoutURL, err := url.Parse(fmt.Sprintf("https://%s/v2/logout", domain))
	if err != nil {
		http.Error(w, "Failed to construct logout URL", http.StatusInternalServerError)
		return
	}

	parameters := url.Values{}

	// Get the return URL from query parameter or use default
	returnTo := r.URL.Query().Get("returnTo")
	if returnTo == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		returnTo = fmt.Sprintf("%s://%s/", scheme, r.Host)
	}

	parameters.Add("returnTo", returnTo)
	parameters.Add("client_id", a.Config.ClientID)
	logoutURL.RawQuery = parameters.Encode()

	a.log.Info("User logout initiated", lctx.Str("return_to", returnTo))

	// Redirect to Auth0 logout URL
	http.Redirect(w, r, logoutURL.String(), http.StatusTemporaryRedirect)
}
