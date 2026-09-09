package application

import (
	"context"
	"errors"
)

// wrong username or password
var ErrInvalidCredentials = errors.New("invalid username or password")

// hardcoded for this coding challenge only, there's no real user store
const (
	demoUsername = "admin"
	demoPassword = "qwerty"
)

// what the HTTP controller calls into
type AuthUsecase interface {
	Login(ctx context.Context, username, password string) (string, error)
}

// signs a token for a given subject, implemented in controller/auth
type TokenIssuer interface {
	IssueToken(ctx context.Context, subject string) (string, error)
}

type authUsecase struct {
	tokenIssuer TokenIssuer
}

// tokenIssuer is required
func NewAuthUsecase(tokenIssuer TokenIssuer) AuthUsecase {
	return &authUsecase{tokenIssuer: tokenIssuer}
}

// check credentials, issue a token if they match
func (u *authUsecase) Login(ctx context.Context, username, password string) (string, error) {
	if username != demoUsername || password != demoPassword {
		return "", ErrInvalidCredentials
	}
	return u.tokenIssuer.IssueToken(ctx, username)
}
