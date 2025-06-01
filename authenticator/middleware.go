package authenticator

import (
	"context"
	"net/http"

	lctx "github.com/hamba/logger/v2/ctx"
	"golang.org/x/oauth2"
)

// ContextKey represents a type for keys in context values.
type ContextKey string

const (
	UserContextKey ContextKey = "user"

	AuthHeader       string = "Authorization"
	AuthHeaderPrefix string = "Bearer "
)

// Claims represents the user profile claims from the ID token.
type Claims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Subject string `json:"sub"`
}

// RequireAuth is a helper function to protect individual routes.
func (a *Authenticator) RequireAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := a.AuthorizationMiddleware(handlerFunc)
		handler.ServeHTTP(w, r)
	}
}

// AuthorizationMiddleware protects routes that require authentication.
func (a *Authenticator) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from the Authorization header
		tokenString := extractTokenFromRequest(r)
		if tokenString == "" {
			http.Error(w, ErrMissingAuthHeader.Error(), http.StatusUnauthorized)
			return
		}
		token := &oauth2.Token{
			AccessToken: tokenString,
		}

		// Verify the token
		idToken, err := a.verifyAccessToken(r.Context(), token)
		if err != nil {
			a.log.Error("Invalid token", lctx.Error("err", err))
			http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
			return
		}

		// Extract user claims
		var claims Claims
		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, ErrClaimsParseFail.Error(), http.StatusInternalServerError)
			return
		}

		// Add claims to the request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
