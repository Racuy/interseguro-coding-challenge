import jwt from 'jsonwebtoken'
import type { RequestHandler } from 'express'

const BEARER_PREFIX = 'Bearer '

function parseBearer(header: string | undefined): string | null {
  if (!header || !header.startsWith(BEARER_PREFIX)) return null
  return header.slice(BEARER_PREFIX.length)
}

// rejects requests without a valid HS256 bearer token
export function jwtProtected(secret: string): RequestHandler {
  return function (req, res, next) {
    const token = parseBearer(req.get('Authorization'))
    if (!token) {
      res.status(401).json({ error: 'missing bearer token' })
      return
    }

    jwt.verify(token, secret, { algorithms: ['HS256'] }, (err) => {
      if (err) {
        res.status(401).json({ error: 'invalid or expired token' })
        return
      }
      next()
    })
  }
}
