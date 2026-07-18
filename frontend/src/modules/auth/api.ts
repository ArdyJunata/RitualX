// frontend/src/modules/auth/api.ts

import { apiClient } from '@/shared/api-client'
import { LoginRequest, LoginResponse, RegisterRequest, RegisterResponse } from './types'

export const authApi = {
  login(body: LoginRequest): Promise<LoginResponse> {
    return apiClient.post<LoginResponse>('/auth/login', body, {
      credentials: 'include', // send/receive HttpOnly cookie
    })
  },

  register(body: RegisterRequest): Promise<RegisterResponse> {
    return apiClient.post<RegisterResponse>('/auth/register', body, {
      credentials: 'include',
    })
  },
}
