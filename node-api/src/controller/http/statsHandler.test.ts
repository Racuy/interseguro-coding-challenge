import { test } from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import jwt from 'jsonwebtoken'
import { createApp } from '../../app.js'
import { createStatsHandler } from './statsHandler.js'
import { MAX_MATRIX_DIMENSION } from '../../domain/matrix.js'
import type { StatsUsecase } from '../../application/statsUsecase.js'

const SECRET = 'test-secret'

function startTestApp() {
  const app = createApp({ jwtSecret: SECRET })
  const server = app.listen(0)
  const address = server.address()
  const port = typeof address === 'object' && address !== null ? address.port : 0
  return { server, url: `http://127.0.0.1:${port}` }
}

function token(): string {
  return jwt.sign({ sub: 'test-user' }, SECRET, { algorithm: 'HS256', expiresIn: '1h' })
}

async function post(url: string, body: string, headers: Record<string, string> = {}): Promise<{ status: number; body: any }> {
  const res = await fetch(`${url}/api/v1/stats`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...headers },
    body,
  })
  return { status: res.status, body: await res.json() }
}

test('valid request returns 200 with the stats', async () => {
  const { server, url } = startTestApp()
  const { status, body } = await post(
    url,
    JSON.stringify({ q: [[1, 0], [0, 1]], r: [[2, 0], [0, 3]] }),
    { Authorization: `Bearer ${token()}` },
  )
  server.close()
  assert.equal(status, 200)
  assert.deepEqual(body, { max: 3, min: 0, average: 0.875, sum: 7, isDiagonal: true })
})

test('non-diagonal matrices return isDiagonal false', async () => {
  const { server, url } = startTestApp()
  const { status, body } = await post(
    url,
    JSON.stringify({ q: [[1, 2], [3, 4]], r: [[5, 6], [7, 8]] }),
    { Authorization: `Bearer ${token()}` },
  )
  server.close()
  assert.equal(status, 200)
  assert.deepEqual(body, { max: 8, min: 1, average: 4.5, sum: 36, isDiagonal: false })
})

test('missing token returns 401', async () => {
  const { server, url } = startTestApp()
  const { status } = await post(url, JSON.stringify({ q: [[1]], r: [[1]] }))
  server.close()
  assert.equal(status, 401)
})

test('malformed JSON body returns 400', async () => {
  const { server, url } = startTestApp()
  const { status } = await post(url, '{not json', { Authorization: `Bearer ${token()}` })
  server.close()
  assert.equal(status, 400)
})

test('invalid matrix returns 400', async () => {
  const { server, url } = startTestApp()
  const { status } = await post(
    url,
    JSON.stringify({ q: [], r: [[1]] }),
    { Authorization: `Bearer ${token()}` },
  )
  server.close()
  assert.equal(status, 400)
})

test('unexpected error from the usecase returns 500', async () => {
  const brokenUsecase: StatsUsecase = {
    process: () => {
      throw new Error('boom')
    },
  }
  const app = express()
  app.use(express.json())
  app.post('/stats', createStatsHandler(brokenUsecase))
  const server = app.listen(0)
  const address = server.address()
  const port = typeof address === 'object' && address !== null ? address.port : 0

  const res = await fetch(`http://127.0.0.1:${port}/stats`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ q: [[1]], r: [[1]] }),
  })
  server.close()
  assert.equal(res.status, 500)
})

test('500x500 matrices are accepted over real HTTP', async () => {
  const { server, url } = startTestApp()
  const big = Array.from({ length: MAX_MATRIX_DIMENSION }, () => new Array(MAX_MATRIX_DIMENSION).fill(1))
  const { status, body } = await post(
    url,
    JSON.stringify({ q: big, r: big }),
    { Authorization: `Bearer ${token()}` },
  )
  server.close()
  assert.equal(status, 200)
  assert.equal(body.max, 1)
  assert.equal(body.min, 1)
})
