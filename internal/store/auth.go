package store

import "context"

type AuthData struct {
}

type AuthStore interface {
	SetState(ctx context.Context, state string) error
	SetCodeVerifier(ctx context.Context, codeVerifier, state string) error
	GetCodeVerifier(
		ctx context.Context,
		state string,
	) (string, error)
	GetState(
		ctx context.Context,
		state string,
	) (string, error)
}