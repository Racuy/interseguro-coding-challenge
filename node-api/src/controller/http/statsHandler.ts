import { InvalidMatrixError, MatrixTooLargeError } from '../../domain/matrix.js'
import type { RequestHandler } from 'express'
import type { StatsUsecase } from '../../application/statsUsecase.js'

export function createStatsHandler(usecase: StatsUsecase): RequestHandler {
  return function statsHandler(req, res) {
    try {
      const stats = usecase.process(req.body ?? {})
      return res.status(200).json(stats)
    } catch (err) {
      if (err instanceof InvalidMatrixError || err instanceof MatrixTooLargeError) {
        return res.status(400).json({ error: err.message })
      }
      console.error(err)
      return res.status(500).json({ error: 'failed to process stats' })
    }
  }
}
