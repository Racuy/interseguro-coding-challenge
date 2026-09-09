package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"go-api/internal/application"
	"go-api/internal/domain"
)

func TestNodeStatsGateway_SendForStats(t *testing.T) {
	t.Run("sends the QR result and decodes the stats", func(t *testing.T) {
		want := domain.MatrixStats{Max: 9, Min: 1, Average: 5, Sum: 45, IsDiagonal: true, DiagonalMatrices: []string{"q"}}
		input := domain.QRFactorization{
			Q: [][]float64{{1, 0}, {0, 1}},
			R: [][]float64{{1, 2}, {0, 3}},
		}

		var gotBody domain.QRFactorization
		var gotAuth, gotContentType string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth = r.Header.Get("Authorization")
			gotContentType = r.Header.Get("Content-Type")
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatalf("server: decode request body: %v", err)
			}
			_ = json.NewEncoder(w).Encode(want)
		}))
		defer server.Close()

		gw := NewNodeStatsGateway(server.URL)
		ctx := application.ContextWithToken(context.Background(), "test-token")

		got, err := gw.SendForStats(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("stats = %+v, want %+v", got, want)
		}
		if gotAuth != "Bearer test-token" {
			t.Errorf("Authorization header = %q, want %q", gotAuth, "Bearer test-token")
		}
		if gotContentType != "application/json" {
			t.Errorf("Content-Type header = %q, want application/json", gotContentType)
		}
		if !reflect.DeepEqual(gotBody, input) {
			t.Errorf("request body = %+v, want %+v", gotBody, input)
		}
	})

	t.Run("no token in context means no Authorization header", func(t *testing.T) {
		var gotAuth string
		var sawHeader bool

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotAuth, sawHeader = r.Header.Get("Authorization"), r.Header.Get("Authorization") != ""
			_ = json.NewEncoder(w).Encode(domain.MatrixStats{})
		}))
		defer server.Close()

		gw := NewNodeStatsGateway(server.URL)
		if _, err := gw.SendForStats(context.Background(), domain.QRFactorization{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sawHeader {
			t.Errorf("Authorization header = %q, want none", gotAuth)
		}
	})

	t.Run("non-2xx response is an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("boom"))
		}))
		defer server.Close()

		gw := NewNodeStatsGateway(server.URL)
		if _, err := gw.SendForStats(context.Background(), domain.QRFactorization{}); err == nil {
			t.Error("expected an error for a 500 response, got nil")
		}
	})

	t.Run("malformed response body is an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("not json"))
		}))
		defer server.Close()

		gw := NewNodeStatsGateway(server.URL)
		if _, err := gw.SendForStats(context.Background(), domain.QRFactorization{}); err == nil {
			t.Error("expected an error for a malformed response body, got nil")
		}
	})

	t.Run("unreachable server is an error", func(t *testing.T) {
		gw := NewNodeStatsGateway("http://127.0.0.1:1") // nothing listens here
		if _, err := gw.SendForStats(context.Background(), domain.QRFactorization{}); err == nil {
			t.Error("expected an error for an unreachable server, got nil")
		}
	})

	t.Run("context deadline actually cuts off a slow server", func(t *testing.T) {
		// server takes 500ms, context only allows 50ms
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(500 * time.Millisecond)
			_ = json.NewEncoder(w).Encode(domain.MatrixStats{})
		}))
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		gw := NewNodeStatsGateway(server.URL)

		start := time.Now()
		_, err := gw.SendForStats(ctx, domain.QRFactorization{})
		elapsed := time.Since(start)

		if err == nil {
			t.Fatal("expected a deadline-exceeded error, got nil")
		}
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("error = %v, want it to wrap context.DeadlineExceeded", err)
		}
		if elapsed >= 500*time.Millisecond {
			t.Errorf("call took %v, want it cut off well before the server's 500ms", elapsed)
		}
	})
}
