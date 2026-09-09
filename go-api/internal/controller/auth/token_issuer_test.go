package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTIssuer_IssueToken(t *testing.T) {
	issuer := NewJWTIssuer("test-secret")

	token, err := issuer.IssueToken(context.Background(), "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	})
	if err != nil || !parsed.Valid {
		t.Fatalf("issued token does not verify: %v", err)
	}

	if claims.Subject != "admin" {
		t.Errorf("subject = %q, want %q", claims.Subject, "admin")
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining <= 0 || remaining > tokenTTL {
		t.Errorf("time until expiry = %v, want (0, %v]", remaining, tokenTTL)
	}
}

func TestJWTIssuer_WrongSecretFailsVerification(t *testing.T) {
	issuer := NewJWTIssuer("test-secret")
	token, err := issuer.IssueToken(context.Background(), "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})
	if err == nil {
		t.Error("expected verification to fail with the wrong secret")
	}
}
