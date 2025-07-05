package middleware

import (
	"context"
	"net/http"

	lctx "github.com/hamba/logger/v2/ctx"
	"github.com/huy125/financial-data-web/authenticator"
	"golang.org/x/oauth2"
)

// ContextKey represents a type for keys in context values.
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// Claims represents the user profile claims from the ID token.
type Claims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Subject string `json:"sub"`
}

// RequireAuth is a helper function to protect individual routes.
func RequireAuth(handlerFunc http.HandlerFunc, a authenticator.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := AuthorizationMiddleware(handlerFunc, a)
		handler.ServeHTTP(w, r)
	}
}

// AuthorizationMiddleware protects routes that require authentication.
func AuthorizationMiddleware(next http.Handler, a authenticator.Authenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from the Authorization header
		tokenString := a.ExtractTokenFromRequest(r)
		if tokenString == "" {
			http.Error(w, authenticator.ErrMissingAuthHeader.Error(), http.StatusUnauthorized)
			return
		}
		token := &oauth2.Token{
			AccessToken: tokenString,
		}

		// Verify the token
		idToken, err := a.VerifyAccessToken(r.Context(), token)
		if err != nil {
			a.Log.Error("Invalid token", lctx.Error("err", err))
			http.Error(w, authenticator.ErrInvalidToken.Error(), http.StatusUnauthorized)
			return
		}

		// Extract user claims
		var claims Claims
		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, authenticator.ErrClaimsParseFail.Error(), http.StatusInternalServerError)
			return
		}

		// Add claims to the request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
