import { createApp } from './app.js'
import { requireEnv } from './config/env.js'

const PORT = 3000
const SHUTDOWN_GRACE_PERIOD_MS = 10000

const jwtSecret = requireEnv('JWT_SECRET')
const app = createApp({ jwtSecret })
const server = app.listen(PORT, () => console.log(`node-api listening on ${PORT}`))

function shutdown(signal: string): void {
  console.log(`received ${signal}, shutting down...`)
  server.close(() => process.exit(0))
  setTimeout(() => {
    console.error(`forced shutdown after ${SHUTDOWN_GRACE_PERIOD_MS}ms`)
    process.exit(1)
  }, SHUTDOWN_GRACE_PERIOD_MS).unref()
}

process.on('SIGTERM', () => shutdown('SIGTERM'))
process.on('SIGINT', () => shutdown('SIGINT'))
