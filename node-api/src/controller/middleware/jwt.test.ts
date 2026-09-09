import { test } from 'node:test'
import assert from 'node:assert/strict'
import express from 'express'
import jwt from 'jsonwebtoken'
import { jwtProtected } from './jwt.js'

const SECRET = 'test-secret'

function startTestApp() {
  const app = express()
  app.use(jwtProtected(SECRET))
  app.get('/protected', (_req, res) => res.status(200).send('ok'))
  const server = app.listen(0)
  const address = server.address()
  const port = typeof address === 'object' && address !== null ? address.port : 0
  return { server, url: `http://127.0.0.1:${port}/protected` }
}

async function get(url: string, headers: Record<string, string>): Promise<number> {
  const res = await fetch(url, { headers })
  return res.status
}

test('valid token is let through', async () => {
  const { server, url } = startTestApp()
  const token = jwt.sign({ sub: 'test-user' }, SECRET, { algorithm: 'HS256', expiresIn: '1h' })
  const status = await get(url, { Authorization: `Bearer ${token}` })
  server.close()
  assert.equal(status, 200)
})

test('missing Authorization header is rejected', async () => {
  const { server, url } = startTestApp()
  const status = await get(url, {})
  server.close()
  assert.equal(status, 401)
})

test('non-Bearer scheme is rejected', async () => {
  const { server, url } = startTestApp()
  const status = await get(url, { Authorization: 'Basic dXNlcjpwYXNz' })
  server.close()
  assert.equal(status, 401)
})

test('garbage token is rejected', async () => {
  const { server, url } = startTestApp()
  const status = await get(url, { Authorization: 'Bearer not.a.jwt' })
  server.close()
  assert.equal(status, 401)
})

test('token signed with the wrong secret is rejected', async () => {
  const { server, url } = startTestApp()
  const token = jwt.sign({ sub: 'test-user' }, 'wrong-secret', { algorithm: 'HS256', expiresIn: '1h' })
  const status = await get(url, { Authorization: `Bearer ${token}` })
  server.close()
  assert.equal(status, 401)
})

test('expired token is rejected', async () => {
  const { server, url } = startTestApp()
  const token = jwt.sign({ sub: 'test-user' }, SECRET, { algorithm: 'HS256', expiresIn: '-1h' })
  const status = await get(url, { Authorization: `Bearer ${token}` })
  server.close()
  assert.equal(status, 401)
})

test('alg:none token is rejected (algorithm confusion)', async () => {
  const { server, url } = startTestApp()
  const token = jwt.sign({ sub: 'test-user' }, null, { algorithm: 'none' })
  const status = await get(url, { Authorization: `Bearer ${token}` })
  server.close()
  assert.equal(status, 401)
})
