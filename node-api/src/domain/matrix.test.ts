import { test } from 'node:test'
import assert from 'node:assert/strict'
import { Matrix, QRFactorization, InvalidMatrixError, MatrixTooLargeError, MAX_MATRIX_DIMENSION } from './matrix.js'

// seeded PRNG, so failures are reproducible
function mulberry32(seed: number): () => number {
  return function () {
    seed |= 0
    seed = (seed + 0x6d2b79f5) | 0
    let t = Math.imul(seed ^ (seed >>> 15), 1 | seed)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

function randomMatrix(rand: () => number, rows: number, cols: number, scale = 1000): number[][] {
  const data: number[][] = []
  for (let i = 0; i < rows; i++) {
    const row: number[] = []
    for (let j = 0; j < cols; j++) row.push((rand() * 2 - 1) * scale)
    data.push(row)
  }
  return data
}

interface ReferenceStats {
  max: number
  min: number
  sum: number
  average: number
}

// independent reference implementation, reduce based, never spread
function referenceStats(q: number[][], r: number[][]): ReferenceStats {
  const values = [...q.flat(), ...r.flat()]
  const max = values.reduce((a, b) => Math.max(a, b))
  const min = values.reduce((a, b) => Math.min(a, b))
  const sum = values.reduce((a, b) => a + b, 0)
  return { max, min, sum, average: sum / values.length }
}

test('Matrix.validate rejects empty, no rows, empty row, jagged', () => {
  assert.throws(() => new Matrix(null as unknown as number[][]).validate(), InvalidMatrixError)
  assert.throws(() => new Matrix([]).validate(), InvalidMatrixError)
  assert.throws(() => new Matrix([[]]).validate(), InvalidMatrixError)
  assert.throws(() => new Matrix([[1, 2], [3]]).validate(), InvalidMatrixError)
})

test('Matrix.validate rejects non-finite values', () => {
  assert.throws(() => new Matrix([[1, NaN]]).validate(), InvalidMatrixError)
  assert.throws(() => new Matrix([[1, Infinity]]).validate(), InvalidMatrixError)
  assert.throws(() => new Matrix([[1, -Infinity]]).validate(), InvalidMatrixError)
})

test('Matrix.validate accepts square and rectangular matrices', () => {
  assert.doesNotThrow(() => new Matrix([[1, 2], [3, 4]]).validate())
  assert.doesNotThrow(() => new Matrix([[1, 2, 3], [4, 5, 6]]).validate())
})

test('Matrix.validate rejects a matrix over the size limit', () => {
  const tooLarge = Array.from({ length: MAX_MATRIX_DIMENSION + 1 }, () => [0])
  assert.throws(() => new Matrix(tooLarge).validate(), MatrixTooLargeError)
})

test('Matrix.isDiagonal true for identity and diagonal matrices', () => {
  assert.equal(new Matrix([[1, 0], [0, 1]]).isDiagonal(), true)
  assert.equal(new Matrix([[5]]).isDiagonal(), true)
  assert.equal(new Matrix([[2, 0, 0], [0, 3, 0], [0, 0, 4]]).isDiagonal(), true)
})

test('Matrix.isDiagonal false for non-square or non-diagonal matrices', () => {
  assert.equal(new Matrix([[1, 2], [3, 4]]).isDiagonal(), false)
  assert.equal(new Matrix([[1, 2, 3], [4, 5, 6]]).isDiagonal(), false)
})

test('Matrix.isDiagonal tolerance is exactly 1e-10', () => {
  // at or under 1e-10: tolerated, counts as diagonal
  assert.equal(new Matrix([[1, 1e-10], [0, 1]]).isDiagonal(), true)
  assert.equal(new Matrix([[1, 0], [1e-11, 1]]).isDiagonal(), true)
  // just over 1e-10: not tolerated, not diagonal
  assert.equal(new Matrix([[1, 1.1e-10], [0, 1]]).isDiagonal(), false)
  assert.equal(new Matrix([[1, 1e-8], [0, 1]]).isDiagonal(), false)
})

test('QRFactorization.stats matches hand-computed values, no diagonal', () => {
  const qr = new QRFactorization([[1, 2], [3, 4]], [[5, 6], [7, 8]])
  assert.deepEqual(qr.stats(), { max: 8, min: 1, average: 4.5, sum: 36, isDiagonal: false, diagonalMatrices: [] })
})

test('QRFactorization.stats matches hand-computed values, both diagonal', () => {
  const qr = new QRFactorization([[1, 0], [0, 1]], [[2, 0], [0, 3]])
  assert.deepEqual(qr.stats(), { max: 3, min: 0, average: 0.875, sum: 7, isDiagonal: true, diagonalMatrices: ['q', 'r'] })
})

test('QRFactorization.stats does not round, full float64 precision passes through', () => {
  const qr = new QRFactorization([[1.234567891234, 0]], [[0]])
  const stats = qr.stats()
  assert.equal(stats.max, 1.234567891234)
  assert.equal(stats.sum, 1.234567891234)
})

test('QRFactorization.diagonalMatrices reports which one, or both, or neither', () => {
  assert.deepEqual(new QRFactorization([[1, 2], [3, 4]], [[5, 6], [7, 8]]).diagonalMatrices(), [])
  assert.deepEqual(new QRFactorization([[1, 2], [3, 4]], [[5, 0], [0, 6]]).diagonalMatrices(), ['r'])
  assert.deepEqual(new QRFactorization([[1, 0], [0, 1]], [[5, 6], [7, 8]]).diagonalMatrices(), ['q'])
  assert.deepEqual(new QRFactorization([[1, 0], [0, 1]], [[5, 0], [0, 6]]).diagonalMatrices(), ['q', 'r'])
})

test('QRFactorization.isDiagonal true if only one of Q or R is diagonal', () => {
  const qr = new QRFactorization([[1, 2], [3, 4]], [[5, 0], [0, 6]])
  assert.equal(qr.isDiagonal(), true)
})

test('QRFactorization.validate propagates matrix errors', () => {
  assert.throws(() => new QRFactorization([], [[1]]).validate(), InvalidMatrixError)
})

test('stats matches an independent reference implementation, 300 random matrices', () => {
  const rand = mulberry32(42)
  for (let iter = 0; iter < 300; iter++) {
    const rows = 1 + Math.floor(rand() * 10)
    const cols = 1 + Math.floor(rand() * 10)
    const q = randomMatrix(rand, rows, cols)
    const r = randomMatrix(rand, rows, cols)

    const got = new QRFactorization(q, r).stats()
    const want = referenceStats(q, r)

    assert.equal(got.max, want.max, `iter ${iter}: max`)
    assert.equal(got.min, want.min, `iter ${iter}: min`)
    assert.ok(Math.abs(got.sum - want.sum) < 1e-6, `iter ${iter}: sum`)
    assert.ok(Math.abs(got.average - want.average) < 1e-6, `iter ${iter}: average`)
  }
})

test('isDiagonal true for 100 randomly generated diagonal matrices', () => {
  const rand = mulberry32(7)
  for (let iter = 0; iter < 100; iter++) {
    const n = 1 + Math.floor(rand() * 10)
    const data: number[][] = Array.from({ length: n }, () => new Array(n).fill(0))
    for (let i = 0; i < n; i++) data[i][i] = rand() * 100 - 50
    assert.equal(new Matrix(data).isDiagonal(), true, `iter ${iter}`)
  }
})

test('regression: 500x500 matrices do not overflow the call stack', () => {
  const big = (value: number): number[][] =>
    Array.from({ length: MAX_MATRIX_DIMENSION }, () => new Array(MAX_MATRIX_DIMENSION).fill(value))
  const qr = new QRFactorization(big(1), big(2))
  const stats = qr.stats()
  assert.equal(stats.max, 2)
  assert.equal(stats.min, 1)
  assert.equal(Number.isFinite(stats.sum), true)
  assert.deepEqual(stats.diagonalMatrices, [])
})
