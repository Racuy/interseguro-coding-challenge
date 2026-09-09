package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"go-api/internal/application"
)

type stubAuthUsecase struct {
	token string
	err   error
}

func (s stubAuthUsecase) Login(ctx context.Context, username, password string) (string, error) {
	return s.token, s.err
}

func newAuthTestApp(handler *AuthHandler) *fiber.App {
	app := fiber.New()
	RegisterAuthRoutes(app.Group("/api/v1"), handler)
	return app
}

func TestAuthHandler_Login(t *testing.T) {
	t.Run("correct credentials return 200 with a token", func(t *testing.T) {
		app := newAuthTestApp(NewAuthHandler(stubAuthUsecase{token: "signed-token"}))

		body, _ := json.Marshal(loginRequest{Username: "admin", Password: "qwerty"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		var got map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if got["token"] != "signed-token" {
			t.Errorf("token = %q, want %q", got["token"], "signed-token")
		}
	})

	t.Run("wrong credentials return 401", func(t *testing.T) {
		app := newAuthTestApp(NewAuthHandler(stubAuthUsecase{err: application.ErrInvalidCredentials}))

		body, _ := json.Marshal(loginRequest{Username: "admin", Password: "wrong"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusUnauthorized)
		}
	})

	t.Run("malformed JSON body returns 400", func(t *testing.T) {
		app := newAuthTestApp(NewAuthHandler(stubAuthUsecase{}))

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader([]byte("{not json")))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("unexpected error returns 500", func(t *testing.T) {
		app := newAuthTestApp(NewAuthHandler(stubAuthUsecase{err: errors.New("boom")}))

		body, _ := json.Marshal(loginRequest{Username: "admin", Password: "qwerty"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusInternalServerError)
		}
	})
}
