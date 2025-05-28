package authenticator

import (
	"context"
	"errors"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// Authenticator encapsulates OAuth2 and OpenID Connect (OIDC) authentication functionality.
// It provides methods for user login, callback handling, token verification,
// and middleware for protecting routes that require authentication.

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
	hmacSecret []byte
}

var ErrInvalidToken = errors.New("invalid token")

func New(provider *oidc.Provider, config oauth2.Config, hmacSecret []byte) (*Authenticator, error) {
	return &Authenticator{
		Provider:   provider,
		Config:     config,
		hmacSecret: hmacSecret,
	}, nil
}

func (a *Authenticator) VerifyToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	oidcConfig := &oidc.Config{ClientID: a.Config.ClientID}

	return a.Provider.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}
