import type { Router, RequestHandler } from 'express'

// wires up the stats route
export function registerRoutes(router: Router, handler: RequestHandler): void {
  router.post('/stats', handler)
}
