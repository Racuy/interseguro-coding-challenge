// matrix types and pure math, no framework code here

export const MAX_MATRIX_DIMENSION = 500
const DIAGONAL_EPSILON = 1e-9

// empty or non-rectangular matrix
export class InvalidMatrixError extends Error {
  constructor() {
    super('matrix must be non-empty and rectangular')
    this.name = 'InvalidMatrixError'
  }
}

// matrix bigger than the size limit
export class MatrixTooLargeError extends Error {
  constructor() {
    super(`matrix dimensions must not exceed ${MAX_MATRIX_DIMENSION}x${MAX_MATRIX_DIMENSION}`)
    this.name = 'MatrixTooLargeError'
  }
}

export interface MatrixStats {
  max: number
  min: number
  average: number
  sum: number
  isDiagonal: boolean
}

// wraps a 2D array of numbers
export class Matrix {
  data: number[][]

  constructor(data: number[][]) {
    this.data = data
  }

  // checks non-empty, rectangular, and within the size limit
  validate(): void {
    if (!Array.isArray(this.data) || this.data.length === 0 || !Array.isArray(this.data[0]) || this.data[0].length === 0) {
      throw new InvalidMatrixError()
    }
    if (this.data.length > MAX_MATRIX_DIMENSION || this.data[0].length > MAX_MATRIX_DIMENSION) {
      throw new MatrixTooLargeError()
    }
    const cols = this.data[0].length
    for (const row of this.data) {
      if (!Array.isArray(row) || row.length !== cols) throw new InvalidMatrixError()
    }
  }

  // square with every off diagonal entry near zero
  isDiagonal(epsilon = DIAGONAL_EPSILON): boolean {
    const rows = this.data.length
    const cols = this.data[0].length
    if (rows !== cols) return false
    for (let i = 0; i < rows; i++) {
      for (let j = 0; j < cols; j++) {
        if (i !== j && Math.abs(this.data[i][j]) > epsilon) return false
      }
    }
    return true
  }
}

// Q and R sent by go-api after its QR decomposition
export class QRFactorization {
  q: Matrix
  r: Matrix

  constructor(q: number[][], r: number[][]) {
    this.q = new Matrix(q)
    this.r = new Matrix(r)
  }

  validate(): void {
    this.q.validate()
    this.r.validate()
  }

  // true if either Q or R is a diagonal matrix
  isDiagonal(): boolean {
    return this.q.isDiagonal() || this.r.isDiagonal()
  }

  // max, min, average, sum over every value in Q and R combined
  // loop based on purpose, spread on 500k values blows V8's call stack
  stats(): MatrixStats {
    let max = -Infinity
    let min = Infinity
    let sum = 0
    let count = 0

    for (const matrix of [this.q, this.r]) {
      for (const row of matrix.data) {
        for (const value of row) {
          if (value > max) max = value
          if (value < min) min = value
          sum += value
          count++
        }
      }
    }

    return { max, min, average: count === 0 ? 0 : sum / count, sum, isDiagonal: this.isDiagonal() }
  }
}
