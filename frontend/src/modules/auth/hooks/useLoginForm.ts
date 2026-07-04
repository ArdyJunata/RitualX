// frontend/src/modules/auth/hooks/useLoginForm.ts

'use client'

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { ApiError } from '@/shared/api-client'
import { authApi } from '../api'
import { LoginRequest } from '../types'

export interface LoginFormFields {
  email: string
  password: string
}

export interface LoginFormErrors {
  email?: string
  password?: string
  general?: string
}

export interface UseLoginFormReturn {
  fields: LoginFormFields
  errors: LoginFormErrors
  isLoading: boolean
  handleChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  handleSubmit: (e: React.FormEvent<HTMLFormElement>) => Promise<void>
}

function validateLogin(fields: LoginFormFields): LoginFormErrors {
  const errors: LoginFormErrors = {}
  if (!fields.email.trim()) {
    errors.email = 'Email is required'
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(fields.email)) {
    errors.email = 'Enter a valid email address'
  }
  if (!fields.password) {
    errors.password = 'Password is required'
  }
  return errors
}

export function useLoginForm(): UseLoginFormReturn {
  const router = useRouter()
  const [fields, setFields] = useState<LoginFormFields>({ email: '', password: '' })
  const [errors, setErrors] = useState<LoginFormErrors>({})
  const [isLoading, setIsLoading] = useState(false)

  const handleChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = e.target
    setFields(prev => ({ ...prev, [name]: value }))
  }, [])

  const handleSubmit = useCallback(async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    setErrors({})

    const clientErrors = validateLogin(fields)
    if (Object.keys(clientErrors).length > 0) {
      setErrors(clientErrors)
      return
    }

    setIsLoading(true)
    try {
      const body: LoginRequest = { email: fields.email, password: fields.password }
      const res = await authApi.login(body)
      localStorage.setItem('ritualx_access_token', res.access_token)
      router.replace('/')
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.hasFieldErrors && err.fields) {
          const mapped: LoginFormErrors = {}
          if (err.fields.email) mapped.email = err.fields.email[0]
          if (err.fields.password) mapped.password = err.fields.password[0]
          setErrors(mapped)
        } else {
          setErrors({ general: err.message || 'Login failed. Please try again.' })
        }
      } else {
        setErrors({ general: 'An unexpected error occurred.' })
      }
    } finally {
      setIsLoading(false)
    }
  }, [fields, router])

  return { fields, errors, isLoading, handleChange, handleSubmit }
}
