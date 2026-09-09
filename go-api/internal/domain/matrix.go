// matrix type, data, and pure math, no framework code here
package domain

import (
	"errors"
	"fmt"
	"math"
)

// core matrix type, everything else operates on this
type Matrix [][]float64

// caps matrix size so QR can't burn unbounded CPU
const maxMatrixDimension = 500

// empty, non-rectangular, or containing a NaN/Infinity value
var ErrInvalidMatrix = errors.New("matrix must be non-empty, rectangular, and contain only finite values")

// matrix bigger than the size limit
var ErrMatrixTooLarge = fmt.Errorf("matrix dimensions must not exceed %dx%d", maxMatrixDimension, maxMatrixDimension)

// checks non-empty, rectangular, within the size limit, and every value finite
func (m Matrix) Validate() error {
	if len(m) == 0 || len(m[0]) == 0 {
		return ErrInvalidMatrix
	}
	if len(m) > maxMatrixDimension || len(m[0]) > maxMatrixDimension {
		return ErrMatrixTooLarge
	}
	cols := len(m[0])
	for _, row := range m {
		if len(row) != cols {
			return ErrInvalidMatrix
		}
		for _, v := range row {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return ErrInvalidMatrix
			}
		}
	}
	return nil
}

// rotates 90 degrees clockwise, e.g. [[1,2],[3,4]] -> [[3,1],[4,2]]
func (m Matrix) Rotate90() Matrix {
	rows := len(m)
	cols := len(m[0])

	rotated := make(Matrix, cols)
	for i := range rotated {
		rotated[i] = make([]float64, rows)
	}

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			rotated[c][rows-1-r] = m[r][c]
		}
	}

	return rotated
}

// splits m into Q and R (m=Q*R) using Householder, works with more cols than rows too
func (m Matrix) QR() (q, r Matrix) {
	rows := len(m)
	cols := len(m[0])

	r = m.clone()
	q = identityMatrix(rows)

	steps := min(rows, cols)

	for k := 0; k < steps; k++ {
		// length of the column below the diagonal
		norm := 0.0
		for i := k; i < rows; i++ {
			norm += r[i][k] * r[i][k]
		}
		norm = math.Sqrt(norm)
		if norm == 0 {
			continue
		}

		// avoid subtracting close numbers
		alpha := norm
		if r[k][k] > 0 {
			alpha = -alpha
		}

		// build the Householder vector
		v := make([]float64, rows-k)
		for i := k; i < rows; i++ {
			v[i-k] = r[i][k]
		}
		v[0] -= alpha

		vNorm := 0.0
		for _, x := range v {
			vNorm += x * x
		}
		if vNorm == 0 {
			continue
		}

		// R[k:, k:] -= 2 * v * (v^T * R[k:, k:]) / vNorm
		for j := k; j < cols; j++ {
			dot := 0.0
			for i := k; i < rows; i++ {
				dot += v[i-k] * r[i][j]
			}
			factor := 2 * dot / vNorm
			for i := k; i < rows; i++ {
				r[i][j] -= factor * v[i-k]
			}
		}

		// same reflection applied to Q
		for i := 0; i < rows; i++ {
			dot := 0.0
			for l := k; l < rows; l++ {
				dot += q[i][l] * v[l-k]
			}
			factor := 2 * dot / vNorm
			for l := k; l < rows; l++ {
				q[i][l] -= factor * v[l-k]
			}
		}
	}

	// clean up float noise below the diagonal
	for i := 0; i < rows; i++ {
		for j := 0; j < i && j < cols; j++ {
			r[i][j] = 0
		}
	}

	// sign convention for a unique decomposition: if R[i][i] < 0, flip Q's column i and R's row i together, Q*R and orthogonality both survive
	for i := 0; i < steps; i++ {
		if r[i][i] < 0 {
			for j := 0; j < cols; j++ {
				r[i][j] = -r[i][j]
			}
			for k := 0; k < rows; k++ {
				q[k][i] = -q[k][i]
			}
		}
	}

	return q, r
}

func (m Matrix) clone() Matrix {
	c := make(Matrix, len(m))
	for i, row := range m {
		c[i] = append([]float64(nil), row...)
	}
	return c
}

func identityMatrix(n int) Matrix {
	m := make(Matrix, n)
	for i := range m {
		m[i] = make([]float64, n)
		m[i][i] = 1
	}
	return m
}

// input matrix sent by the client
type MatrixRequest struct {
	Matrix Matrix `json:"matrix"`
}

// Q and R from the QR decomposition of the original (not rotated) matrix
type QRFactorization struct {
	Q Matrix `json:"q"`
	R Matrix `json:"r"`
}

// what node-api computes over Q and R combined
type MatrixStats struct {
	Max              float64  `json:"max"`
	Min              float64  `json:"min"`
	Average          float64  `json:"average"`
	Sum              float64  `json:"sum"`
	IsDiagonal       bool     `json:"isDiagonal"`
	DiagonalMatrices []string `json:"diagonalMatrices"`
}

// everything go-api's own caller gets back: original, rotation, QR factors of the original, and node-api's stats
type MatrixProcessResult struct {
	Original Matrix      `json:"original"`
	Rotated  Matrix      `json:"rotated"`
	Q        Matrix      `json:"q"`
	R        Matrix      `json:"r"`
	Stats    MatrixStats `json:"stats"`
}
