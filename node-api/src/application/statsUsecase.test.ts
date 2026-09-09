import { test } from 'node:test'
import assert from 'node:assert/strict'
import { createStatsUsecase } from './statsUsecase.js'
import { InvalidMatrixError, MatrixTooLargeError, MAX_MATRIX_DIMENSION } from '../domain/matrix.js'

test('process returns stats for a valid request', () => {
  const usecase = createStatsUsecase()
  const stats = usecase.process({ q: [[1, 0], [0, 1]], r: [[2, 0], [0, 3]] })
  assert.deepEqual(stats, { max: 3, min: 0, average: 0.875, sum: 7, isDiagonal: true, diagonalMatrices: ['q', 'r'] })
})

test('process does not round, full float64 precision passes through', () => {
  const usecase = createStatsUsecase()
  const stats = usecase.process({ q: [[1.234567891234, 0], [0, 1]], r: [[2, 0], [0, 3]] })
  assert.equal(stats.max, 3)
  assert.equal(stats.sum, 1.234567891234 + 1 + 2 + 3)
})

test('process rejects an invalid matrix before computing anything', () => {
  const usecase = createStatsUsecase()
  assert.throws(() => usecase.process({ q: [], r: [[1]] }), InvalidMatrixError)
})

test('process rejects a non-finite value before computing anything', () => {
  const usecase = createStatsUsecase()
  assert.throws(() => usecase.process({ q: [[NaN]], r: [[1]] }), InvalidMatrixError)
})

test('process rejects a matrix over the size limit', () => {
  const usecase = createStatsUsecase()
  const tooLarge = Array.from({ length: MAX_MATRIX_DIMENSION + 1 }, () => [0])
  assert.throws(() => usecase.process({ q: tooLarge, r: [[1]] }), MatrixTooLargeError)
})

test('process rejects a missing body', () => {
  const usecase = createStatsUsecase()
  assert.throws(() => usecase.process({}), InvalidMatrixError)
})
