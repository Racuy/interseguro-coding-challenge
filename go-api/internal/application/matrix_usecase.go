// use cases and the ports they need, no framework code
package application

import (
	"context"

	"go-api/internal/domain"
)

// what the HTTP controller calls into
type MatrixUsecase interface {
	Process(ctx context.Context, req domain.MatrixRequest) (domain.MatrixStats, error)
}

// sends the QR result to node-api and gets stats back
type StatsGateway interface {
	SendForStats(ctx context.Context, result domain.QRFactorization) (domain.MatrixStats, error)
}

type matrixUsecase struct {
	statsGateway StatsGateway
}

// statsGateway is required
func NewMatrixUsecase(statsGateway StatsGateway) MatrixUsecase {
	return &matrixUsecase{statsGateway: statsGateway}
}

// validate, rotate, QR, send to node-api, return stats
func (u *matrixUsecase) Process(ctx context.Context, req domain.MatrixRequest) (domain.MatrixStats, error) {
	if err := req.Matrix.Validate(); err != nil {
		return domain.MatrixStats{}, err
	}

	rotated := req.Matrix.Rotate90()
	q, r := rotated.QR()

	stats, err := u.statsGateway.SendForStats(ctx, domain.QRFactorization{Q: q, R: r})
	if err != nil {
		return domain.MatrixStats{}, err
	}

	return stats, nil
}
