package api

import (
	"fmt"
	"net/http"
	"net/url"

	lctx "github.com/hamba/logger/v2/ctx"
)

// LogoutHandler handles the Auth0 logout process.
// This should be called when the user wants to log out.
func (s *Server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	token := s.authenticator.ExtractTokenFromRequest(r)
	if token != "" {
		err := s.authenticator.RevokeToken(r.Context(), token)
		if err != nil && s.log != nil {
			s.log.Error("Failed to revoke token", lctx.Err(err))
		}
	}

	u, err := s.authenticator.GetBaseURL()
	if err != nil {
		s.log.Error("Failed to get domain", lctx.Err(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	logoutURL, err := url.Parse(u + "/v2/logout")
	if err != nil {
		s.log.Error("Failed to parse logout URL", lctx.Err(err))
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
	parameters.Add("client_id", s.authenticator.GetClientID())
	logoutURL.RawQuery = parameters.Encode()

	s.log.Info("User logout initiated", lctx.Str("return_to", returnTo))

	cookie := &http.Cookie{
		Name:     s.cookieCfg.Name,
		Path:     s.cookieCfg.Path,
		HttpOnly: s.cookieCfg.HttpOnly,
		Secure:   s.cookieCfg.Secure,
		Value:    "",
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)

	// Redirect to Auth0 logout URL
	http.Redirect(w, r, logoutURL.String(), http.StatusTemporaryRedirect)
}
