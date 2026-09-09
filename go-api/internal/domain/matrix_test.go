package domain

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"testing"
)

const epsilon = 1e-9

func TestMatrix_Validate(t *testing.T) {
	tests := []struct {
		name    string
		matrix  Matrix
		wantErr error
	}{
		{"nil matrix", nil, ErrInvalidMatrix},
		{"no rows", Matrix{}, ErrInvalidMatrix},
		{"empty row", Matrix{{}}, ErrInvalidMatrix},
		{"jagged rows", Matrix{{1, 2}, {3}}, ErrInvalidMatrix},
		{"valid square", Matrix{{1, 2}, {3, 4}}, nil},
		{"valid rectangular", Matrix{{1, 2, 3}, {4, 5, 6}}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.matrix.Validate()
			if err != tt.wantErr {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}

	t.Run("matrix over the size limit is rejected", func(t *testing.T) {
		tooLarge := make(Matrix, maxMatrixDimension+1)
		for i := range tooLarge {
			tooLarge[i] = []float64{0}
		}
		if err := tooLarge.Validate(); err != ErrMatrixTooLarge {
			t.Errorf("got error %v, want ErrMatrixTooLarge", err)
		}
	})
}

func TestMatrix_Rotate90(t *testing.T) {
	tests := []struct {
		name  string
		input Matrix
		want  Matrix
	}{
		{
			name:  "square 2x2",
			input: Matrix{{1, 2}, {3, 4}},
			want:  Matrix{{3, 1}, {4, 2}},
		},
		{
			name:  "square 3x3",
			input: Matrix{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}},
			want:  Matrix{{7, 4, 1}, {8, 5, 2}, {9, 6, 3}},
		},
		{
			name:  "tall (more rows than cols)",
			input: Matrix{{1, 2}, {3, 4}, {5, 6}},
			want:  Matrix{{5, 3, 1}, {6, 4, 2}},
		},
		{
			name:  "wide (more cols than rows)",
			input: Matrix{{1, 2, 3}, {4, 5, 6}},
			want:  Matrix{{4, 1}, {5, 2}, {6, 3}},
		},
		{
			name:  "single row",
			input: Matrix{{1, 2, 3}},
			want:  Matrix{{1}, {2}, {3}},
		},
		{
			name:  "single column",
			input: Matrix{{1}, {2}, {3}},
			want:  Matrix{{3, 2, 1}},
		},
		{
			name:  "1x1",
			input: Matrix{{5}},
			want:  Matrix{{5}},
		},
		{
			name:  "negative and zero values",
			input: Matrix{{-1, 0}, {0, -2}},
			want:  Matrix{{0, -1}, {-2, 0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Rotate90()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Rotate90(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// property test: rotating 4x always returns the original
func TestMatrix_Rotate90_FourTimesIsIdentity(t *testing.T) {
	rng := rand.New(rand.NewSource(7))

	const iterations = 100
	const maxDim = 12

	for iter := 0; iter < iterations; iter++ {
		rows := rng.Intn(maxDim) + 1
		cols := rng.Intn(maxDim) + 1

		original := make(Matrix, rows)
		for i := range original {
			original[i] = make([]float64, cols)
			for j := range original[i] {
				original[i][j] = rng.Float64()*20 - 10
			}
		}

		result := original
		for i := 0; i < 4; i++ {
			result = result.Rotate90()
		}

		name := fmt.Sprintf("iter_%d_%dx%d", iter, rows, cols)
		t.Run(name, func(t *testing.T) {
			if !reflect.DeepEqual(result, original) {
				t.Errorf("4x Rotate90 = %v, want original %v", result, original)
			}
		})
	}
}

// test-only helper
func matMul(a, b Matrix) Matrix {
	rows := len(a)
	inner := len(b)
	cols := len(b[0])

	result := make(Matrix, rows)
	for i := range result {
		result[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			sum := 0.0
			for k := 0; k < inner; k++ {
				sum += a[i][k] * b[k][j]
			}
			result[i][j] = sum
		}
	}
	return result
}

func transpose(a Matrix) Matrix {
	rows := len(a)
	cols := len(a[0])

	t := make(Matrix, cols)
	for i := range t {
		t[i] = make([]float64, rows)
		for j := 0; j < rows; j++ {
			t[i][j] = a[j][i]
		}
	}
	return t
}

func almostEqualMatrix(t *testing.T, got, want Matrix, msg string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %d rows, want %d", msg, len(got), len(want))
	}
	for i := range got {
		if len(got[i]) != len(want[i]) {
			t.Fatalf("%s: row %d has %d cols, want %d", msg, i, len(got[i]), len(want[i]))
		}
		for j := range got[i] {
			if math.Abs(got[i][j]-want[i][j]) > epsilon {
				t.Errorf("%s: [%d][%d] = %v, want %v", msg, i, j, got[i][j], want[i][j])
			}
		}
	}
}

func isUpperTriangular(t *testing.T, r Matrix) {
	t.Helper()
	for i := range r {
		for j := 0; j < len(r[i]) && j < i; j++ {
			if math.Abs(r[i][j]) > epsilon {
				t.Errorf("R[%d][%d] = %v, want ~0 (R must be upper triangular)", i, j, r[i][j])
			}
		}
	}
}

func isOrthogonal(t *testing.T, q Matrix) {
	t.Helper()
	product := matMul(transpose(q), q)
	almostEqualMatrix(t, product, identityMatrix(len(q)), "Q^T * Q should be the identity")
}

// checks Q orthogonal, R upper triangular, Q*R = original
func TestMatrix_QR(t *testing.T) {
	tests := []struct {
		name   string
		matrix Matrix
	}{
		{"1x1", Matrix{{5}}},
		{"square 2x2", Matrix{{1, 2}, {3, 4}}},
		{"square 3x3", Matrix{{12, -51, 4}, {6, 167, -68}, {-4, 24, -41}}},
		{"tall (more rows than cols)", Matrix{{1, 2}, {3, 4}, {5, 6}}},
		{"wide (more cols than rows)", Matrix{{1, 2, 3}, {4, 5, 6}}},
		{"identity 3x3", identityMatrix(3)},
		{"zero column", Matrix{{0, 1}, {0, 2}, {0, 3}}},
		{"all zeros", Matrix{{0, 0}, {0, 0}}},
		{"negative values", Matrix{{-1, -2}, {-3, -4}}},
		{"dependent columns (col2 = 2 * col1)", Matrix{{1, 2}, {2, 4}, {3, 6}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, r := tt.matrix.QR()

			isOrthogonal(t, q)
			isUpperTriangular(t, r)
			almostEqualMatrix(t, matMul(q, r), tt.matrix, "Q * R should equal the original matrix")
		})
	}
}

// same checks, but on random matrices instead of a fixed list
func TestMatrix_QR_RandomMatrices(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	const iterations = 300
	const maxDim = 10
	const maxValue = 1000.0

	for iter := 0; iter < iterations; iter++ {
		rows := rng.Intn(maxDim) + 1
		cols := rng.Intn(maxDim) + 1

		matrix := make(Matrix, rows)
		for i := range matrix {
			matrix[i] = make([]float64, cols)
			for j := range matrix[i] {
				matrix[i][j] = (rng.Float64()*2 - 1) * maxValue
			}
		}

		q, r := matrix.QR()

		// subtest so one failure doesn't stop the rest
		name := fmt.Sprintf("iter_%d_%dx%d", iter, rows, cols)
		t.Run(name, func(t *testing.T) {
			isOrthogonal(t, q)
			isUpperTriangular(t, r)
			almostEqualMatrix(t, matMul(q, r), matrix, name)
		})
	}
}
