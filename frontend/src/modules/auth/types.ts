// frontend/src/modules/auth/types.ts

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  display_name: string
  username: string
  email: string
  password: string
}

export interface AuthUser {
  id: string
  display_name: string
  username: string
  email: string
  created_at: string
}

export interface LoginResponse {
  access_token: string
  user: AuthUser
}

export interface RegisterResponse {
  access_token: string
  user: AuthUser
}
