// frontend/src/modules/stats/hooks/useRoutineHeatmap.ts

'use client'

import { useState, useEffect, useCallback } from 'react'
import { statsApi } from '../api'
import { HeatmapData } from '../types'

export interface UseRoutineHeatmapReturn {
  data: HeatmapData
  isLoading: boolean
  error: string | null
  refresh: () => void
}

export function useRoutineHeatmap(routineId: string, year: number): UseRoutineHeatmapReturn {
  const [data, setData] = useState<HeatmapData>({})
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const result = await statsApi.getHeatmap(routineId, year)
      setData(result ?? {})
    } catch {
      setError('Failed to load heatmap. Please try again.')
      setData({})
    } finally {
      setIsLoading(false)
    }
  }, [routineId, year])

  useEffect(() => {
    load()
  }, [load])

  return { data, isLoading, error, refresh: load }
}
