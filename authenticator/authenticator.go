package authenticator

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"golang.org/x/oauth2"
)

// Authenticator encapsulates OAuth2 and OpenID Connect (OIDC) authentication functionality.
// It provides methods for user login, callback handling, token verification,
// and middleware for protecting routes that require authentication.
type Authenticator struct {
	Provider     *oidc.Provider
	Config       oauth2.Config
	HMACSecret   []byte
	APIAudience  string
	ClientOrigin string

	log *logger.Logger
}

// Option defines a function type to apply options to Authenticator.
type Option func(*Authenticator)

// WithOAuthConfig sets the OAuth2 configuration.
func WithOAuthConfig(clientID, clientSecret, redirectURL string) Option {
	return func(a *Authenticator) {
		a.Config.ClientID = clientID
		a.Config.ClientSecret = clientSecret
		a.Config.RedirectURL = redirectURL
	}
}

// WithHMACSecret sets the HMAC secret for state parameter verification.
func WithHMACSecret(secret []byte) Option {
	return func(a *Authenticator) {
		a.HMACSecret = secret
	}
}

// WithAPIAudience sets the API audience for access token verification.
func WithAPIAudience(aud string) Option {
	return func(a *Authenticator) {
		a.APIAudience = aud
	}
}

// WithClientOrigin sets the client origin for CORS configuration and redirection after the verification.
func WithClientOrigin(origin string) Option {
	return func(a *Authenticator) {
		a.ClientOrigin = origin
	}
}

// WithLogger sets the logger for the authenticator.
func WithLogger(log *logger.Logger) Option {
	return func(a *Authenticator) {
		a.log = log.With(lctx.Str("component", "authenticator"))
	}
}

func New(ctx context.Context, domain string, opts ...Option) (*Authenticator, error) {
	// Create OIDC provider
	u, err := url.Parse(domain)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	provider, err := oidc.NewProvider(ctx, u.String())
	if err != nil {
		return nil, fmt.Errorf("creating OIDC provider: %w", err)
	}

	// Configure OAuth2
	auth := &Authenticator{
		Provider: provider,
		Config: oauth2.Config{
			Endpoint: provider.Endpoint(),
			Scopes:   []string{oidc.ScopeOpenID, "profile", "email"},
		},
	}

	for _, opt := range opts {
		opt(auth)
	}

	return auth, nil
}

// verifyToken verifies the ID token and returns the parsed token.
func (a *Authenticator) verifyToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, ErrNoIDToken
	}

	oidcConfig := &oidc.Config{
		ClientID: a.Config.ClientID,
	}
	verifier := a.Provider.Verifier(oidcConfig)
	return verifier.Verify(ctx, rawIDToken)
}

// verifyAccessToken verifies an access token.
func (a *Authenticator) verifyAccessToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	oidcConfig := &oidc.Config{
		ClientID: a.APIAudience,
	}
	verifier := a.Provider.Verifier(oidcConfig)

	return verifier.Verify(ctx, token.AccessToken)
}

// revokeToken sends a request to Auth0 to revoke the token.
func (a *Authenticator) revokeToken(ctx context.Context, token string) error {
	u, err := a.getBaseURL()
	if err != nil {
		return fmt.Errorf("getting domain: %w", err)
	}

	revokeURL := u + "/oauth/revoke"

	// Prepare the request body
	form := url.Values{}
	form.Add("client_id", a.Config.ClientID)
	form.Add("client_secret", a.Config.ClientSecret)
	form.Add("token", token)

	// Create and send the request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("creating revocation request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending revocation request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading revocation response: %w", err)
		}
		return fmt.Errorf("revoking token with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (a *Authenticator) getBaseURL() (string, error) {
	endpoint := a.Provider.Endpoint().AuthURL
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint URL: %w", err)
	}

	return strings.TrimSuffix(endpoint, u.RequestURI()), nil
}

// extractTokenFromRequest gets the bearer token from the Authorization header.
func extractTokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get(AuthHeader)
	if authHeader == "" {
		return ""
	}

	if !strings.HasPrefix(authHeader, AuthHeaderPrefix) {
		return ""
	}

	tokenString := strings.TrimPrefix(authHeader, AuthHeaderPrefix)
	if tokenString == "" {
		return ""
	}

	return tokenString
}
