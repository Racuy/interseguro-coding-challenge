import express from 'express'
import { jwtProtected } from './controller/middleware/jwt.js'
import { registerRoutes } from './controller/http/routes.js'
import { createStatsHandler } from './controller/http/statsHandler.js'
import { createStatsUsecase } from './application/statsUsecase.js'
import type { Express, ErrorRequestHandler } from 'express'

// Q and R together can carry up to 500k numbers as JSON text, default express limit is too small
const JSON_BODY_LIMIT = '20mb'

export interface AppOptions {
  jwtSecret: string
}

export function createApp({ jwtSecret }: AppOptions): Express {
  const app = express()
  app.disable('x-powered-by')
  app.use(express.json({ limit: JSON_BODY_LIMIT }))

  // malformed JSON body, mirrors go-api's BodyParser 400 behavior
  const handleParseError: ErrorRequestHandler = (err, _req, res, next) => {
    if (err.type === 'entity.parse.failed') {
      res.status(400).json({ error: 'invalid request body' })
      return
    }
    next(err)
  }
  app.use(handleParseError)

  app.get('/health', (_req, res) => res.status(200).send('ok'))

  const handler = createStatsHandler(createStatsUsecase())
  const api = express.Router()
  api.use(jwtProtected(jwtSecret))
  registerRoutes(api, handler)
  app.use('/api/v1', api)

  const handleUnexpectedError: ErrorRequestHandler = (err, _req, res, _next) => {
    console.error(err)
    res.status(500).json({ error: 'internal server error' })
  }
  app.use(handleUnexpectedError)

  return app
}
