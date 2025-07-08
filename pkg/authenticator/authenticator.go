package authenticator

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
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

// Authorization headers constants.
const (
	AuthHeader       string = "Authorization"
	AuthHeaderPrefix string = "Bearer "
)

// OAuth state token generation constants.
const (
	StateGenerationByteSize = 32
	StatePartCount          = 2
)

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

// VerifyToken verifies the ID token and returns the parsed token.
func (a *Authenticator) VerifyToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.Config.ClientID,
	}
	verifier := a.Provider.Verifier(oidcConfig)
	return verifier.Verify(ctx, rawIDToken)
}

// VerifyAccessToken verifies an access token.
func (a *Authenticator) VerifyAccessToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	oidcConfig := &oidc.Config{
		ClientID: a.APIAudience,
	}
	verifier := a.Provider.Verifier(oidcConfig)

	return verifier.Verify(ctx, token.AccessToken)
}

// RevokeToken sends a request to Auth0 to revoke the token.
func (a *Authenticator) RevokeToken(ctx context.Context, token string) error {
	u, err := a.GetBaseURL()
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

// GetBaseURL returns the auth0 provider endpoint URL.
func (a *Authenticator) GetBaseURL() (string, error) {
	endpoint := a.Provider.Endpoint().AuthURL
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint URL: %w", err)
	}

	return strings.TrimSuffix(endpoint, u.RequestURI()), nil
}

// GenerateState computes the state based on HMACSecret.
func (a *Authenticator) GenerateState() (string, error) {
	b := make([]byte, StateGenerationByteSize)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("generating random string: %w", err)
	}

	state := base64.URLEncoding.EncodeToString(b)
	mac := hmac.New(sha256.New, a.HMACSecret)
	mac.Write([]byte(state))
	signature := mac.Sum(nil)

	return state + ":" + hex.EncodeToString(signature), nil
}

// VerifyState verifies if the state is matching with expected signature.
func (a *Authenticator) VerifyState(s string) error {
	parts := strings.SplitN(s, ":", StatePartCount)
	if len(parts) != StatePartCount {
		return fmt.Errorf("invalid state format: expecting %d parts", StatePartCount)
	}

	state := parts[0]
	sig := parts[1]

	decodedSig, err := hex.DecodeString(sig)
	if err != nil {
		return fmt.Errorf("decoding signature: %w", err)
	}

	mac := hmac.New(sha256.New, a.HMACSecret)
	mac.Write([]byte(state))
	expectedSig := mac.Sum(nil)

	if !hmac.Equal(decodedSig, expectedSig) {
		return fmt.Errorf("invalid signature: expected %x, got %x", expectedSig, decodedSig)
	}

	return nil
}

// Exchange allows to exchange token with provider.
func (a *Authenticator) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.Config.Exchange(ctx, code)
}

// GetClientID returns the OAuth's clientID.
func (a *Authenticator) GetClientID() string {
	return a.Config.ClientID
}

// GetClientOrigin returns the client origin that will redirect by authenticator.
func (a *Authenticator) GetClientOrigin() string {
	return a.ClientOrigin
}

// ExtractTokenFromRequest gets the bearer token from the Authorization header.
func (a *Authenticator) ExtractTokenFromRequest(r *http.Request) string {
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
