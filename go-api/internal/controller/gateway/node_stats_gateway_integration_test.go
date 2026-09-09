//go:build integration

// real HTTP against a live node-api, skipped unless NODE_API_URL and JWT_SECRET are set, run with -tags=integration
package gateway

import (
	"context"
	"os"
	"reflect"
	"testing"

	"go-api/internal/application"
	"go-api/internal/controller/auth"
	"go-api/internal/domain"
)

func TestNodeStatsGateway_Integration_RealNodeAPI(t *testing.T) {
	baseURL := os.Getenv("NODE_API_URL")
	secret := os.Getenv("JWT_SECRET")
	if baseURL == "" || secret == "" {
		t.Skip("NODE_API_URL and JWT_SECRET must both be set to run this integration test against a real node-api")
	}

	issuer := auth.NewJWTIssuer(secret)
	token, err := issuer.IssueToken(context.Background(), "integration-test")
	if err != nil {
		t.Fatalf("failed to mint a token: %v", err)
	}
	ctx := application.ContextWithToken(context.Background(), token)

	gw := NewNodeStatsGateway(baseURL)

	t.Run("hand-verified stats over a diagonal Q and R", func(t *testing.T) {
		// values: 1,0,0,1 (Q) and 2,0,0,3 (R) -> max 3, min 0, sum 7, average 0.875, both diagonal
		input := domain.QRFactorization{
			Q: domain.Matrix{{1, 0}, {0, 1}},
			R: domain.Matrix{{2, 0}, {0, 3}},
		}

		got, err := gw.SendForStats(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error calling real node-api: %v", err)
		}

		want := domain.MatrixStats{Max: 3, Min: 0, Average: 0.875, Sum: 7, IsDiagonal: true, DiagonalMatrices: []string{"q", "r"}}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("stats = %+v, want %+v", got, want)
		}
	})

	t.Run("full float64 precision survives the round trip, no rounding on either side", func(t *testing.T) {
		input := domain.QRFactorization{
			Q: domain.Matrix{{1.234567891234, 0}},
			R: domain.Matrix{{0}},
		}

		got, err := gw.SendForStats(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error calling real node-api: %v", err)
		}

		if got.Max != 1.234567891234 {
			t.Errorf("max = %v, want 1.234567891234 unrounded", got.Max)
		}
	})

	t.Run("a bad token is rejected independently by node-api", func(t *testing.T) {
		badCtx := application.ContextWithToken(context.Background(), token+"tampered")

		_, err := gw.SendForStats(badCtx, domain.QRFactorization{Q: domain.Matrix{{1}}, R: domain.Matrix{{1}}})
		if err == nil {
			t.Error("expected node-api to reject a tampered token, got no error")
		}
	})
}
