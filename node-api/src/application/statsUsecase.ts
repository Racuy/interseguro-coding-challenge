import { QRFactorization } from '../domain/matrix.js'
import type { MatrixStats } from '../domain/matrix.js'

export interface StatsRequestBody {
  q?: unknown
  r?: unknown
}

export interface StatsUsecase {
  process(req: StatsRequestBody): MatrixStats
}

export function createStatsUsecase(): StatsUsecase {
  return {
    process(req: StatsRequestBody): MatrixStats {
      const qr = new QRFactorization(req.q as number[][], req.r as number[][])
      qr.validate()
      return qr.stats()
    },
  }
}
