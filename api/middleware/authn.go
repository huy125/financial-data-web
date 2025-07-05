package middleware

import (
	"context"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// ContextKey represents a type for keys in context values.
type ContextKey string

const (
	UserContextKey ContextKey = "user"
)

// Authenticator defines the interface for handling authorization middleware.
type Authenticator interface {
	ExtractTokenFromRequest(r *http.Request) string
	VerifyAccessToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error)
}

// Claims represents the user profile claims from the ID token.
type Claims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Subject string `json:"sub"`
}

// RequireAuth is a helper function to protect individual routes.
func RequireAuth(handlerFunc http.HandlerFunc, a Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := AuthorizationMiddleware(handlerFunc, a)
		handler.ServeHTTP(w, r)
	}
}

// AuthorizationMiddleware protects routes that require authentication.
func AuthorizationMiddleware(next http.Handler, a Authenticator) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from the Authorization header
		tokenString := a.ExtractTokenFromRequest(r)
		if tokenString == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}
		token := &oauth2.Token{
			AccessToken: tokenString,
		}

		// Verify the token
		idToken, err := a.VerifyAccessToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract user claims
		var claims Claims
		if err := idToken.Claims(&claims); err != nil {
			http.Error(w, "Failed to parse claims", http.StatusInternalServerError)
			return
		}

		// Add claims to the request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
