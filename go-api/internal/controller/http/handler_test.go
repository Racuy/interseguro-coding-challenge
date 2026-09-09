package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"

	"go-api/internal/application"
	"go-api/internal/domain"
)

// fake usecase, records the context it was called with
type stubUsecase struct {
	result  domain.MatrixProcessResult
	err     error
	lastCtx context.Context
}

func (s *stubUsecase) Process(ctx context.Context, req domain.MatrixRequest) (domain.MatrixProcessResult, error) {
	s.lastCtx = ctx
	return s.result, s.err
}

func newTestApp(handler *MatrixHandler) *fiber.App {
	app := fiber.New()
	RegisterRoutes(app.Group("/api/v1"), handler)
	return app
}

func TestMatrixHandler_Process(t *testing.T) {
	t.Run("valid request returns 200 with the result", func(t *testing.T) {
		want := domain.MatrixProcessResult{
			Original: domain.Matrix{{1, 2}, {3, 4}},
			Rotated:  domain.Matrix{{3, 1}, {4, 2}},
			Q:        domain.Matrix{{-1, 0}, {0, -1}},
			R:        domain.Matrix{{-3, -1}, {0, -2}},
			Stats:    domain.MatrixStats{Max: 4, Min: 1, Average: 2.5, Sum: 10, IsDiagonal: false, DiagonalMatrices: []string{}},
		}
		app := newTestApp(NewMatrixHandler(&stubUsecase{result: want}))

		body, _ := json.Marshal(domain.MatrixRequest{Matrix: [][]float64{{1, 2}, {3, 4}}})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/process", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
		}

		var got domain.MatrixProcessResult
		if err := json.NewDecoder(resp.Body).Decode(&got); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %+v, want %+v", got, want)
		}
	})

	t.Run("Authorization header is propagated to the use case via context", func(t *testing.T) {
		usecase := &stubUsecase{}
		app := newTestApp(NewMatrixHandler(usecase))

		body, _ := json.Marshal(domain.MatrixRequest{Matrix: [][]float64{{1, 2}, {3, 4}}})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/process", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		if _, err := app.Test(req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		token, ok := application.TokenFromContext(usecase.lastCtx)
		if !ok || token != "test-token" {
			t.Errorf("token in context = %q, ok=%v, want %q, true", token, ok, "test-token")
		}
	})

	t.Run("context passed to the use case carries a deadline", func(t *testing.T) {
		usecase := &stubUsecase{}
		app := newTestApp(NewMatrixHandler(usecase))

		body, _ := json.Marshal(domain.MatrixRequest{Matrix: [][]float64{{1, 2}, {3, 4}}})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/process", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		if _, err := app.Test(req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// this deadline is what stops a hung node-api from blocking forever
		deadline, ok := usecase.lastCtx.Deadline()
		if !ok {
			t.Fatal("expected the use case's context to have a deadline, got none")
		}
		if remaining := time.Until(deadline); remaining <= 0 || remaining > matrixProcessTimeout {
			t.Errorf("time until deadline = %v, want (0, %v]", remaining, matrixProcessTimeout)
		}
	})

	t.Run("missing Authorization header means no token in context", func(t *testing.T) {
		usecase := &stubUsecase{}
		app := newTestApp(NewMatrixHandler(usecase))

		body, _ := json.Marshal(domain.MatrixRequest{Matrix: [][]float64{{1, 2}, {3, 4}}})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/process", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		if _, err := app.Test(req); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := application.TokenFromContext(usecase.lastCtx); ok {
			t.Error("expected no token in context when Authorization header is absent")
		}
	})

	t.Run("invalid matrix returns 400", func(t *testing.T) {
		app := newTestApp(NewMatrixHandler(&stubUsecase{err: domain.ErrInvalidMatrix}))

		body, _ := json.Marshal(domain.MatrixRequest{Matrix: [][]float64{}})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/process", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
		}
	})

	t.Run("malformed JSON body returns 400", func(t *testing.T) {
		app := newTestApp(NewMatrixHandler(&stubUsecase{}))

		req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/process", bytes.NewReader([]byte("{not json")))
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
		app := newTestApp(NewMatrixHandler(&stubUsecase{err: errors.New("boom")}))

		body, _ := json.Marshal(domain.MatrixRequest{Matrix: [][]float64{{1, 2}, {3, 4}}})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/matrix/process", bytes.NewReader(body))
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
