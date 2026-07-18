// frontend/src/modules/routines/api.ts

import { apiClient } from '@/shared/api-client'
import { CreateRoutineRequest, Routine } from './types'

export const routinesApi = {
  create(body: CreateRoutineRequest): Promise<Routine> {
    const token = localStorage.getItem('ritualx_access_token') ?? ''
    return apiClient.post<Routine>('/routines', body, {
      credentials: 'include',
      headers: { Authorization: `Bearer ${token}` },
    })
  },
}
