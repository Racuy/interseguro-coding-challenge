package application

import (
	"context"
	"errors"
	"testing"
)

type stubTokenIssuer struct {
	token string
	err   error
}

func (s stubTokenIssuer) IssueToken(ctx context.Context, subject string) (string, error) {
	return s.token, s.err
}

func TestAuthUsecase_Login(t *testing.T) {
	t.Run("correct credentials issue a token", func(t *testing.T) {
		usecase := NewAuthUsecase(stubTokenIssuer{token: "signed-token"})
		token, err := usecase.Login(context.Background(), "admin", "qwerty")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token != "signed-token" {
			t.Errorf("got %q, want %q", token, "signed-token")
		}
	})

	t.Run("wrong password is rejected", func(t *testing.T) {
		usecase := NewAuthUsecase(stubTokenIssuer{})
		_, err := usecase.Login(context.Background(), "admin", "wrong")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("got error %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("wrong username is rejected", func(t *testing.T) {
		usecase := NewAuthUsecase(stubTokenIssuer{})
		_, err := usecase.Login(context.Background(), "someone", "qwerty")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("got error %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("token issuer error is propagated", func(t *testing.T) {
		wantErr := errors.New("signing failed")
		usecase := NewAuthUsecase(stubTokenIssuer{err: wantErr})
		_, err := usecase.Login(context.Background(), "admin", "qwerty")
		if !errors.Is(err, wantErr) {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
	})
}
