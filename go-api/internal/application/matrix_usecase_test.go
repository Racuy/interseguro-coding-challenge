package application

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"go-api/internal/domain"
)

// fake gateway that records what it received
type stubGateway struct {
	received domain.QRFactorization
	stats    domain.MatrixStats
	err      error
}

func (s *stubGateway) SendForStats(ctx context.Context, result domain.QRFactorization) (domain.MatrixStats, error) {
	s.received = result
	return s.stats, s.err
}

func TestMatrixUsecase_Process(t *testing.T) {
	validCases := []struct {
		name   string
		matrix domain.Matrix
	}{
		{"square matrix", domain.Matrix{{1, 2}, {3, 4}}},
		{"rectangular matrix", domain.Matrix{{1, 2, 3}, {4, 5, 6}}},
		{"1x1 matrix", domain.Matrix{{7}}},
	}

	for _, tt := range validCases {
		t.Run(tt.name, func(t *testing.T) {
			gateway := &stubGateway{stats: domain.MatrixStats{Max: 42}}
			usecase := NewMatrixUsecase(gateway)

			req := domain.MatrixRequest{Matrix: tt.matrix}
			result, err := usecase.Process(context.Background(), req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Process shouldn't touch or reshape the stats
			if result != gateway.stats {
				t.Errorf("got %+v, want %+v", result, gateway.stats)
			}

			// gateway should get QR of the rotated matrix
			wantQ, wantR := tt.matrix.Rotate90().QR()
			if !reflect.DeepEqual(gateway.received.Q, wantQ) || !reflect.DeepEqual(gateway.received.R, wantR) {
				t.Errorf("gateway got Q=%v R=%v, want Q=%v R=%v", gateway.received.Q, gateway.received.R, wantQ, wantR)
			}
		})
	}

	t.Run("invalid matrix is rejected before calling the gateway", func(t *testing.T) {
		gateway := &stubGateway{}
		usecase := NewMatrixUsecase(gateway)

		_, err := usecase.Process(context.Background(), domain.MatrixRequest{Matrix: nil})
		if !errors.Is(err, domain.ErrInvalidMatrix) {
			t.Fatalf("got error %v, want ErrInvalidMatrix", err)
		}
		if gateway.received.Q != nil || gateway.received.R != nil {
			t.Error("gateway should not have been called for an invalid matrix")
		}
	})

	t.Run("gateway error is propagated", func(t *testing.T) {
		wantErr := errors.New("node-api unreachable")
		gateway := &stubGateway{err: wantErr}
		usecase := NewMatrixUsecase(gateway)

		req := domain.MatrixRequest{Matrix: domain.Matrix{{1, 2}, {3, 4}}}
		_, err := usecase.Process(context.Background(), req)
		if !errors.Is(err, wantErr) {
			t.Fatalf("got error %v, want %v", err, wantErr)
		}
	})
}

// safe for concurrent use, never mutated after construction
type fixedStatsGateway struct {
	stats domain.MatrixStats
}

func (g fixedStatsGateway) SendForStats(ctx context.Context, result domain.QRFactorization) (domain.MatrixStats, error) {
	return g.stats, nil
}

// hits Process from many goroutines at once, run with -race
func TestMatrixUsecase_Process_Concurrent(t *testing.T) {
	want := domain.MatrixStats{Max: 9, Min: 1, Average: 5, Sum: 45}
	usecase := NewMatrixUsecase(fixedStatsGateway{stats: want})
	matrix := domain.Matrix{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}}

	const goroutines = 50

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			req := domain.MatrixRequest{Matrix: matrix}
			result, err := usecase.Process(context.Background(), req)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result != want {
				t.Errorf("got %+v, want %+v", result, want)
			}
		}()
	}

	wg.Wait()
}
