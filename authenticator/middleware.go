package authenticator

import (
	"context"
	"github.com/coreos/go-oidc/v3/oidc"
	lctx "github.com/hamba/logger/v2/ctx"
	"golang.org/x/oauth2"
	"net/http"
	"strings"
)

// ContextKey is a type for keys in context values
type ContextKey string

const UserContextKey ContextKey = "user"

const AuthHeader = "Authorization"
const AuthHeaderPrefix = "Bearer "

// Claims represents the user profile claims from the ID token
type Claims struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Subject string `json:"sub"`
}

// RequireAuth is a helper function to protect individual routes
func (a *Authenticator) RequireAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler := a.AuthorizationMiddleware(handlerFunc)
		handler.ServeHTTP(w, r)
	}
}

// AuthorizationMiddleware protects routes that require authentication
func (a *Authenticator) AuthorizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get(AuthHeader)
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, AuthHeaderPrefix) {
			http.Error(w, "Authorization header must start with 'Bearer '", http.StatusUnauthorized)
			return
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, AuthHeaderPrefix)
		if tokenString == "" {
			http.Error(w, "Token is required", http.StatusUnauthorized)
			return
		}

		// Create a token object with the access token
		token := &oauth2.Token{
			AccessToken: tokenString,
		}

		// Verify the token
		idToken, err := a.verifyAccessToken(r.Context(), token)
		if err != nil {
			a.log.Error("Invalid token", lctx.Error("err", err))
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

		// Continue with the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// verifyAccessToken verifies an access token
func (a *Authenticator) verifyAccessToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	oidcConfig := &oidc.Config{
		ClientID: a.ApiAudience,
	}
	verifier := a.Provider.Verifier(oidcConfig)

	return verifier.Verify(ctx, token.AccessToken)
}
