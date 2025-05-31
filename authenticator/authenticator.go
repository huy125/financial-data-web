package authenticator

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/hamba/logger/v2"
	lctx "github.com/hamba/logger/v2/ctx"
	"golang.org/x/oauth2"
)

// Authenticator encapsulates OAuth2 and OpenID Connect (OIDC) authentication functionality.
// It provides methods for user login, callback handling, token verification,
// and middleware for protecting routes that require authentication.
type Authenticator struct {
	Provider    *oidc.Provider
	Config      oauth2.Config
	HmacSecret  []byte
	ApiAudience string

	log *logger.Logger
}

// Option defines a function type to apply options to Authenticator
type Option func(*Authenticator) error

// WithOAuthConfig sets the OAuth2 configuration
func WithOAuthConfig(clientID, clientSecret, redirectURL string) Option {
	return func(a *Authenticator) error {
		a.Config.ClientID = clientID
		a.Config.ClientSecret = clientSecret
		a.Config.RedirectURL = redirectURL
		return nil
	}
}

// WithHmacSecret sets the HMAC secret for state parameter verification
func WithHmacSecret(secret []byte) Option {
	return func(a *Authenticator) error {
		a.HmacSecret = secret
		return nil
	}
}

// WithApiAudience sets the API audience for access token verification
func WithApiAudience(aud string) Option {
	return func(a *Authenticator) error {
		a.ApiAudience = aud
		return nil
	}
}

// WithLogger sets the logger for the authenticator
func WithLogger(log *logger.Logger) Option {
	return func(a *Authenticator) error {
		a.log = log.With(lctx.Str("component", "authenticator"))
		return nil
	}
}

var ErrInvalidToken = errors.New("invalid token")

func New(ctx context.Context, domain string, opts ...Option) (*Authenticator, error) {
	// Create OIDC provider
	provider, err := oidc.NewProvider(ctx, "https://"+domain+"/")
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
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
		if err := opt(auth); err != nil {
			return nil, fmt.Errorf("failed to apply options: %w", err)
		}
	}

	return auth, nil
}
