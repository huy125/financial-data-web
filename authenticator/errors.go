package authenticator

import "errors"

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidState      = errors.New("invalid state")
	ErrMissingAuthCode   = errors.New("missing authorization code")
	ErrTokenExchangeFail = errors.New("failed to exchange token")
	ErrTokenVerifyFail   = errors.New("failed to verify token")
	ErrNoIDToken         = errors.New("no id_token field in oauth2 token")
	ErrMissingAuthHeader = errors.New("authorization header is required")
	ErrClaimsParseFail   = errors.New("failed to parse claims")
)
