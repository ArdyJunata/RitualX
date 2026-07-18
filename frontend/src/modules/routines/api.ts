// frontend/src/modules/routines/api.ts

import { apiClient } from '@/shared/api-client'
import { CreateRoutineRequest, Routine, RoutineLog } from './types'

function authHeaders() {
  const token = localStorage.getItem('ritualx_access_token') ?? ''
  return { Authorization: `Bearer ${token}` }
}

export const routinesApi = {
  create(body: CreateRoutineRequest): Promise<Routine> {
    return apiClient.post<Routine>('/routines', body, {
      credentials: 'include',
      headers: authHeaders(),
    })
  },

  list(): Promise<Routine[]> {
    return apiClient.get<Routine[]>('/routines', {
      credentials: 'include',
      headers: authHeaders(),
    })
  },

  getTodayLog(routineId: string): Promise<RoutineLog | null> {
    return apiClient.get<RoutineLog>(`/routines/${routineId}/log`, {
      credentials: 'include',
      headers: authHeaders(),
    }).catch((err: unknown) => {
      // 404 = no log today, not an error
      if (err && typeof err === 'object' && 'statusCode' in err && (err as { statusCode: number }).statusCode === 404) {
        return null
      }
      throw err
    })
  },

  logRoutine(routineId: string): Promise<RoutineLog> {
    return apiClient.post<RoutineLog>(`/routines/${routineId}/log`, {}, {
      credentials: 'include',
      headers: authHeaders(),
    })
  },

  deleteLog(routineId: string, logId: string): Promise<void> {
    return apiClient.delete<void>(`/routines/${routineId}/log/${logId}`, {
      credentials: 'include',
      headers: authHeaders(),
    })
  },
}
