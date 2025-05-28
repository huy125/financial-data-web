package authenticator

import (
	"context"
	"errors"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"os"
)

// Authenticator is used to authenticate users.
type Authenticator struct {
	*oidc.Provider
	oauth2.Config
}

var ErrInvalidToken = errors.New("invalid token")

func New(provider *oidc.Provider, config oauth2.Config) (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+os.Getenv("AUTH0_DOMAIN")+"/",
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	return &Authenticator{
		Provider: provider,
		Config:   config,
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
