// frontend/src/modules/stats/api.ts

import { apiClient } from '@/shared/api-client'
import { HeatmapData } from './types'

function authHeaders() {
  const token =
    typeof window !== 'undefined' ? localStorage.getItem('ritualx_access_token') ?? '' : ''
  return { Authorization: `Bearer ${token}` }
}

export const statsApi = {
  // GET /routines/:id/heatmap?year=YYYY -> { "YYYY-MM-DD": count }
  getHeatmap(routineId: string, year: number): Promise<HeatmapData> {
    return apiClient.get<HeatmapData>(`/routines/${routineId}/heatmap?year=${year}`, {
      credentials: 'include',
      headers: authHeaders(),
    })
  },
}
