// signs JWTs for the login endpoint
package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// how long an issued token stays valid
const tokenTTL = time.Hour

type JWTIssuer struct {
	secret []byte
}

func NewJWTIssuer(secret string) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret)}
}

// signs an HS256 token with subject as the sub claim
func (i *JWTIssuer) IssueToken(ctx context.Context, subject string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(i.secret)
}
